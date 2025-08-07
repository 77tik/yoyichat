package db

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"path/filepath"
	"sync"
	"time"
)

var dbMap = map[string]*gorm.DB{}
var syncLock sync.Mutex

func init() {
	initDB("yoyichat")
}

func initDB(dbName string) {
	var e error
	realPath, _ := filepath.Abs("./")
	configFilePath := realPath + "/db/yoyichat.db"
	syncLock.Lock()
	logConfig := logger.Config{
		LogLevel: logger.Info,
		Colorful: true,
	}
	dbMap[dbName], e = gorm.Open(sqlite.Open(configFilePath), &gorm.Config{
		Logger: logger.New(log.Default(), logConfig),
	})
	db, _ := dbMap[dbName].DB()
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Second * 8)
	syncLock.Unlock()
	if e != nil {
		logrus.Error("connect db fail:%s", e.Error())
	}
}

func GetDB(dbName string) *gorm.DB {
	if db, ok := dbMap[dbName]; ok {
		return db
	} else {
		return nil
	}
}

type DbYoyiChat struct {
}

func (*DbYoyiChat) GetDbName() string { return "yoyichat" }
