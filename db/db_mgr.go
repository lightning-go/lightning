/**
 * @author: Jason
 * Created: 19-5-4
 */

package db

import (
	"github.com/jinzhu/gorm"
	"fmt"
	"github.com/lightning-go/lightning/logger"
	"time"
)

const dbLogPath = "./logs/db.log"

type IDBMgr interface {
	Close()
	QueryPrimaryKey(pk, tableName string) uint64
}

type DBMgr struct {
	host   string
	dbName string
	user   string
	pwd    string
	dbConn *gorm.DB
	log    *logger.Logger
}

func NewDBMgr(dbType, dbName, user, pwd, host string) *DBMgr {
	dbMgr := &DBMgr{
		host:   host,
		dbName: dbName,
		user:   user,
		pwd:    pwd,
		log:    logger.NewLogger(logger.TRACE),
	}

	if dbMgr.log == nil {
		return nil
	}
	dbMgr.log.SetRotation(time.Hour*24*30, time.Hour*24, dbLogPath)

	sqlConn := ""
	switch dbType {
	case "mysql":
		sqlConn = dbMgr.getMySQLConn()
	case "postgres":
		sqlConn = dbMgr.getPostgreSQLConn()
	case "sqlite3":
		sqlConn = dbMgr.getSqlite3Conn()
	}
	if len(sqlConn) == 0 {
		return nil
	}

	var err error
	dbMgr.dbConn, err = gorm.Open(dbType, sqlConn)
	if err != nil {
		dbMgr.log.Error(err)
		return nil
	}

	dbMgr.dbConn.SingularTable(true)

	dbMgr.log.Info("database is connected", logger.Fields{"db": dbName})
	return dbMgr
}

func (dbMgr *DBMgr) getMySQLConn() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=local",
		dbMgr.user, dbMgr.pwd, dbMgr.host, dbMgr.dbName)
}

func (dbMgr *DBMgr) getPostgreSQLConn() string {
	return fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable",
		dbMgr.user, dbMgr.pwd, dbMgr.host, dbMgr.dbName)
}

func (dbMgr *DBMgr) getSqlite3Conn() string {
	return dbMgr.dbName
}

func (dbMgr *DBMgr) Close() {
	dbMgr.dbConn.Close()
}

func (dbMgr *DBMgr) DBConn() *gorm.DB {
	return dbMgr.dbConn
}
