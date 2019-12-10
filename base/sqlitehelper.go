package base

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

type SqliteDB struct {
	db *gorm.DB
}

func NewSqliteDB(dbPath string, debug bool) *SqliteDB {

	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("new gorm err", err)
	}

	db.LogMode(debug)
	// 加表名前缀
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "t_" + defaultTableName
	}

	return &SqliteDB{db: db}
}
