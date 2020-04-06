/**
 * @author: Jason
 * Created: 19-5-4
 */

package db

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm"
	"fmt"
	"github.com/lightning-go/lightning/logger"
	"time"
	"database/sql"
)

const dbLogPath = "./logs/db.log"

const (
	DB_type_mysql    = "mysql"
	DB_type_postgres = "postgres"
	DB_type_sqlite3  = "sqlite3"
)

type IDBMgr interface {
	Close()
	QueryPrimaryKey(pk, tableName string) int64
	QueryOneCond(tableName, where string, f func(*sql.Row))
	QueryCond(tableName, where string, f func(*sql.Rows))
	QueryKeyCond(tableName, key, where string, f func(*sql.Rows))
	Query(tableName string, f func(*sql.Rows))
	Insert(tableName, fields, value string) error
	Update(tableName, fields, value string) error
	Delete(tableName, where string) error
	NewRecord(v interface{}) error
	SaveRecord(v interface{}) error
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
	case DB_type_mysql:
		sqlConn = dbMgr.getMySQLConn()
	case DB_type_postgres:
		sqlConn = dbMgr.getPostgreSQLConn()
	case DB_type_sqlite3:
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
	//return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=local",
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True",
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
