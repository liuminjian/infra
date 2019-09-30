package base

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func NewSSHClient(h *Host) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            h.User,
		HostKeyCallback: hostKeyCallBackFunc(h.Ip),
		Timeout:         time.Second * 5,
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSA,
			ssh.KeyAlgoDSA,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
		},
	}

	if h.AuthType == PasswordAuth {
		config.Auth = []ssh.AuthMethod{ssh.Password(h.Password)}
	} else {
		config.Auth = []ssh.AuthMethod{publicKeyAuthFunc(h.KeyFile)}
	}

	addr := fmt.Sprintf("%s:%d", h.Ip, h.Port)
	client, err := ssh.Dial("tcp", addr, config)
	return client, err
}

func hostKeyCallBackFunc(host string) ssh.HostKeyCallback {
	hostPath, _ := homedir.Expand("~/.ssh/known_hosts")
	file, err := os.Open(hostPath)
	if err != nil {
		log.Fatalf("cannot find known_hosts file :%v", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var hostKey ssh.PublicKey

	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if fields[0] == host {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				log.Fatalf("error parsing %s:%v", fields[2], err)
			}
			break
		}
	}
	if hostKey == nil {
		log.Fatalf("no hostkey for %s,%v", host, err)
	}
	return ssh.FixedHostKey(hostKey)
}

func publicKeyAuthFunc(keyFile string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatal("ssh key file read failed ", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("ssh key signer failed ", err)
	}
	return ssh.PublicKeys(signer)
}

type EnvMap map[string]string

func runCommand(client *ssh.Client, command string, envs ...EnvMap) (stdout string, err error) {
	session, err := client.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	var outputBuf bytes.Buffer
	var errorBuf bytes.Buffer
	session.Stdout = &outputBuf
	session.Stderr = &errorBuf
	// 设置环境变量
	for _, env := range envs {
		for k, v := range env {
			err = session.Setenv(k, v)
			if err != nil {
				return
			}
		}
	}

	err = session.Run(command)
	if err != nil {
		stderr := string(errorBuf.Bytes())
		err = fmt.Errorf(stderr)
		return
	}
	stdout = string(outputBuf.Bytes())
	return
}

func sudoCommand(client *ssh.Client, command string, user string, password string, envs ...EnvMap) (stdout string, err error) {
	session, err := client.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	var outputBuf bytes.Buffer
	var errorBuf bytes.Buffer
	session.Stdout = &outputBuf
	session.Stderr = &errorBuf
	// 设置环境变量
	for _, env := range envs {
		for k, v := range env {
			err = session.Setenv(k, v)
			if err != nil {
				return
			}
		}
	}

	in, err := session.StdinPipe()
	if err != nil {
		return
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		return
	}

	var (
		prefix    = fmt.Sprintf("[sudo] password for %s:", user)
		prefixLen = len(prefix)
	)

	abort := make(chan bool)

	go func(in io.Writer, output *bytes.Buffer, done <-chan bool) {

		for {
			select {
			case <-done:
				return
			default:
				content := string(output.Bytes())
				if len(content) < prefixLen {
					continue
				}
				if strings.HasPrefix(content, prefix) {
					_, err = in.Write([]byte(password + "\n"))
					if err == io.EOF {
						log.Debug(err)
						err = nil
					}
					break
				}
			}
		}
	}(in, &outputBuf, abort)

	defer func() {
		abort <- true
	}()

	err = session.Run("sudo " + command)
	if err != nil {
		stderr := string(errorBuf.Bytes())
		err = fmt.Errorf(stderr)
		return
	}

	stdout = strings.TrimPrefix(string(outputBuf.Bytes()), prefix)
	return
}
