package base

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"sync"
)

var masterInstance *gorm.DB
var dbLock sync.RWMutex

func NewDBMaster(user string, password string, host string, port int, database string, debug bool) *gorm.DB {
	sourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		user, password, host, port, database)

	db, err := gorm.Open("mysql", sourceName)
	log.Info(sourceName)
	if err != nil {
		log.Fatal("new gorm err", err)
	}

	db.LogMode(debug)
	// 加表名前缀
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "t_" + defaultTableName
	}

	masterInstance = db
	return masterInstance
}

type MysqlConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Debug    bool
}

func InitDBMaster(opts ...*MysqlConfig) *gorm.DB {
	if masterInstance != nil {
		return masterInstance
	}

	dbLock.Lock()
	defer dbLock.Unlock()

	if masterInstance != nil {
		return masterInstance
	}
	for _, opt := range opts {
		return NewDBMaster(opt.User, opt.Password, opt.Host, opt.Port, opt.Database, opt.Debug)
	}
	return nil
}

func CreateTx() *gorm.DB {
	return masterInstance.Begin()
}
