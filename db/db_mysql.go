/**
 * @author: Jason
 * Created: 19-5-4
 */

package db

import (
	"database/sql"
	"fmt"
)

type MysqlMgr struct {
	*DBMgr
}

func NewMysqlMgr(dbName, user, pwd, host string) *MysqlMgr {
	mm := &MysqlMgr{
		DBMgr: NewDBMgr("mysql", dbName, user, pwd, host),
	}
	if mm.DBMgr == nil {
		return nil
	}
	AddDB(dbName, mm)
	return mm
}

func (mm *MysqlMgr) QueryPrimaryKey(pk, tableName string) uint64 {
	s := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s DESC LIMIT 1;", pk, tableName, pk)
	row := mm.dbConn.Raw(s).Row()
	var id uint64
	err := row.Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			mm.log.Error(err)
		}
		return 0
	}
	return id
}

func (mm *MysqlMgr) QueryOneCond(tableName, where string, f func(*sql.Row)) {
	if f == nil {
		return
	}
	s := fmt.Sprintf("SELECT * FROM %s WHERE %s;", tableName, where)
	row := mm.dbConn.Raw(s).Row()
	f(row)
}

func (mm *MysqlMgr) QueryCond(tableName, where string, f func(*sql.Rows)) {
	if f == nil {
		return
	}
	s := fmt.Sprintf("SELECT * FROM %s WHERE %s;", tableName, where)
	rows, err := mm.dbConn.Raw(s).Rows()
	if err != nil {
		if err != sql.ErrNoRows {
			mm.log.Error(err)
		}
		return
	}
	f(rows)
	rows.Close()
}

func (mm *MysqlMgr) QueryKeyCond(tableName, key, where string, f func(*sql.Rows)) {
	if f == nil {
		return
	}
	s := fmt.Sprintf("SELECT %s FROM %s WHERE %s;", key, tableName, where)
	rows, err := mm.dbConn.Raw(s).Rows()
	if err != nil {
		if err != sql.ErrNoRows {
			mm.log.Error(err)
		}
		return
	}
	f(rows)
	rows.Close()
}

func (mm *MysqlMgr) Query(tableName string, f func(*sql.Rows)) {
	if f == nil {
		return
	}
	s := fmt.Sprintf("SELECT * FROM %s;", tableName)
	rows, err := mm.dbConn.Raw(s).Rows()
	if err != nil {
		if err != sql.ErrNoRows {
			mm.log.Error(err)
		}
		return
	}
	f(rows)
	rows.Close()
}

func (mm *MysqlMgr) Insert(tableName, fields, value string) error {
	s := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", tableName, fields, value)
	return mm.dbConn.Exec(s).Error
}

func (mm *MysqlMgr) Update(tableName, fields, value string) error {
	s := fmt.Sprintf("UPDATE %s SET %s WHERE %s;", tableName, fields, value)
	return mm.dbConn.Exec(s).Error
}

func (mm *MysqlMgr) Delete(tableName, where string) error {
	s := fmt.Sprintf("DELETE FROM %s WHERE %s;", tableName, where)
	return mm.dbConn.Exec(s).Error
}
