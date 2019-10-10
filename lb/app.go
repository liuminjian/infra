package lb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"io/ioutil"
	"strings"
	"time"
)

type Apps struct {
	client *clientv3.Client
}

type App struct {
	Name string
	lb   Balance
}

type Balance interface {
	Next(key string) *ServerInstance
	Add(element string)
	Remove(element string)
}

type Status string

const (
	Enable  Status = "enable"
	Disable Status = "disable"
)

type ServerInstance struct {
	AppName string
	Address string
	Status  Status
}

func NewApps(endpoints []string, tls *tls.Config) (*Apps, error) {

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 2 * time.Second,
		TLS:         tls,
	})

	if err != nil {
		return nil, err
	}

	return &Apps{client: client}, nil
}

func NewTLSConfig(ca string, cert string, key string) (*tls.Config, error) {
	etcdCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	caData, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)
	return &tls.Config{Certificates: []tls.Certificate{etcdCert}, RootCAs: pool}, nil
}

func (a *Apps) Get(key, appName string) *App {
	resp, err := a.client.KV.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		log.Error(err)
		return nil
	}

	app := &App{
		Name: appName,
		lb:   NewConsistent(),
	}

	for _, kvpair := range resp.Kvs {
		app.lb.Add(strings.TrimPrefix(string(kvpair.Key), key))
	}

	// watch 新增或删除的实例
	go func() {

		watchReversion := resp.Header.Revision

		watchChan := a.client.Watcher.Watch(context.TODO(), key, clientv3.WithRev(watchReversion),
			clientv3.WithPrefix())

		for watchResp := range watchChan {
			for _, watchEvent := range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					log.Info(strings.TrimPrefix(string(watchEvent.Kv.Key), key))
					app.lb.Add(strings.TrimPrefix(string(watchEvent.Kv.Key), key))
				case mvccpb.DELETE:
					log.Info(strings.TrimPrefix(string(watchEvent.Kv.Key), key))
					app.lb.Remove(strings.TrimPrefix(string(watchEvent.Kv.Key), key))
				}
			}
		}

	}()

	return app
}

func (a *App) Get(key string) *ServerInstance {
	return a.lb.Next(key)
}
