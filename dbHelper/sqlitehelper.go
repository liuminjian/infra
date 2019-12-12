package dbHelper

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

type SqliteDB struct {
	DB *gorm.DB
}

func NewSqliteDB(dbPath string, debug bool) (*SqliteDB, error) {

	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		log.Error("new gorm err", err)
		return nil, err
	}

	db.LogMode(debug)
	// 加表名前缀
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "t_" + defaultTableName
	}

	return &SqliteDB{DB: db}, nil
}
