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

const (
	DB_type_mysql    = "mysql"
	DB_type_postgres = "postgres"
)

type IDBMgr interface {
	Close()
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

	sqlConn := ""
	switch dbType {
	case DB_type_mysql:
		sqlConn = dbMgr.getMySQLConn()
	case DB_type_postgres:
		sqlConn = dbMgr.getPostgreSQLConn()
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
	AddDB(dbName, dbMgr)

	dbMgr.log.Info("database is connected", logger.Fields{"db": dbName})
	return dbMgr
}

func (dbMgr *DBMgr) SetLogLevel(lv int) {
	if dbMgr.log == nil {
		dbMgr.log = logger.NewLogger(lv)
	} else {
		dbMgr.log.SetLevel(lv)
	}
}

func (dbMgr *DBMgr) SetLogRotation(lv, maxAge, rotationTime int, pathFile string) {
	dbMgr.SetLogLevel(lv)
	dbMgr.log.SetRotation(time.Minute*time.Duration(maxAge), time.Minute*time.Duration(rotationTime), pathFile)
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

func (dbMgr *DBMgr) Close() {
	dbMgr.dbConn.Close()
}

func (dbMgr *DBMgr) DBConn() *gorm.DB {
	return dbMgr.dbConn
}

func (dbMgr *DBMgr) QueryRecord(tableName, where string, dest interface{}) error {
	s := fmt.Sprintf("SELECT * FROM %s WHERE %s;", tableName, where)
	raw := dbMgr.dbConn.Raw(s)
	if raw.Error != nil {
		return raw.Error
	}
	return raw.Scan(dest).Error
}

func (dbMgr *DBMgr) NewRecord(dest interface{}) error {
	return dbMgr.dbConn.Create(dest).Error
}

func (dbMgr *DBMgr) SaveRecord(dest interface{}) error {
	return dbMgr.dbConn.Save(dest).Error
}

func (dbMgr *DBMgr) Insert(tableName, fields, value string) error {
	s := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", tableName, fields, value)
	return dbMgr.dbConn.Exec(s).Error
}

func (dbMgr *DBMgr) Update(tableName, fields, value string) error {
	s := fmt.Sprintf("UPDATE %s SET %s WHERE %s;", tableName, fields, value)
	return dbMgr.dbConn.Exec(s).Error
}

func (dbMgr *DBMgr) Delete(tableName, where string) error {
	s := fmt.Sprintf("DELETE FROM %s WHERE %s;", tableName, where)
	return dbMgr.dbConn.Exec(s).Error
}

func (dbMgr *DBMgr) QueryPrimaryKey(pk, tableName string) int64 {
	s := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s DESC LIMIT 1;", pk, tableName, pk)
	row := dbMgr.dbConn.Raw(s).Row()
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			dbMgr.log.Error(err)
			return -1
		}
		return 0
	}
	return id
}

func (dbMgr *DBMgr) QueryOneCond(tableName, where string, f func(*sql.Row)) {
	if f == nil {
		return
	}
	s := fmt.Sprintf("SELECT * FROM %s WHERE %s;", tableName, where)
	row := dbMgr.dbConn.Raw(s).Row()
	f(row)
}

func (dbMgr *DBMgr) QueryCond(tableName, where string, f func(*sql.Rows)) {
	if f == nil {
		return
	}
	s := fmt.Sprintf("SELECT * FROM %s WHERE %s;", tableName, where)
	rows, err := dbMgr.dbConn.Raw(s).Rows()
	if err != nil {
		if err != sql.ErrNoRows {
			dbMgr.log.Error(err)
		}
		return
	}
	f(rows)
	rows.Close()
}

func (dbMgr *DBMgr) QueryKeyCond(tableName, key, where string, f func(*sql.Rows)) {
	if f == nil {
		return
	}
	s := fmt.Sprintf("SELECT %s FROM %s WHERE %s;", key, tableName, where)
	rows, err := dbMgr.dbConn.Raw(s).Rows()
	if err != nil {
		if err != sql.ErrNoRows {
			dbMgr.log.Error(err)
		}
		return
	}
	f(rows)
	rows.Close()
}

func (dbMgr *DBMgr) Query(tableName string, f func(*sql.Rows)) {
	if f == nil {
		return
	}
	s := fmt.Sprintf("SELECT * FROM %s;", tableName)
	rows, err := dbMgr.dbConn.Raw(s).Rows()
	if err != nil {
		if err != sql.ErrNoRows {
			dbMgr.log.Error(err)
		}
		return
	}
	f(rows)
	rows.Close()
}

func (dbMgr *DBMgr) QueryCondEx(tableName, where string, objCallback func()interface{}, f func(interface{})) {
	if objCallback == nil || f == nil {
		return
	}
	s := fmt.Sprintf("SELECT * FROM %s WHERE %s;", tableName, where)
	rows, err := dbMgr.dbConn.Raw(s).Rows()
	if err != nil {
		if err != sql.ErrNoRows {
			dbMgr.log.Error(err)
		}
		return
	}
	for rows.Next() {
		result := objCallback()
		err := dbMgr.dbConn.ScanRows(rows, result)
		if err != nil {
			dbMgr.log.Error(err)
			continue
		}
		f(result)
	}
	rows.Close()
}

