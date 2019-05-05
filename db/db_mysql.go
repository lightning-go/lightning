/**
 * @author: Jason
 * Created: 19-5-4
 */

package db

import "database/sql"

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
	s := "SELECT ? FROM ? ORDER BY ? DESC LIMIT 1;"
	row := mm.dbConn.Raw(s, pk, tableName, pk).Row()
	var id uint64
	err := row.Scan(&id)
	if err != nil {
		mm.log.Error(err)
		return 0
	}
	return id
}

func (mm *MysqlMgr) QueryOneCond(tableName, where string) *sql.Row {
	s := "SELECT * FROM ? WHERE ?;"
	row := mm.dbConn.Raw(s, tableName, where).Row()
	return row
}

func (mm *MysqlMgr) QueryCond(tableName, where string) *sql.Rows {
	s := "SELECT * FROM ? WHERE ?;"
	rows, err := mm.dbConn.Raw(s, tableName, where).Rows()
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	return rows
}

func (mm *MysqlMgr) QueryKeyCond(tableName, key, where string) *sql.Rows {
	s := "SELECT ? FROM ? WHERE ?;"
	rows, err := mm.dbConn.Raw(s, key, tableName, where).Rows()
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	return rows
}

func (mm *MysqlMgr) Query(tableName string) *sql.Rows {
	s := "SELECT * FROM ?;"
	rows, err := mm.dbConn.Raw(s, tableName).Rows()
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	return rows
}

func (mm *MysqlMgr) Insert(tableName, fields, value string) error {
	s := "INSERT INTO ? (?) VALUES (?);"
	return mm.dbConn.Exec(s, tableName, fields, value).Error
}

func (mm *MysqlMgr) Update(tableName, fields, value string) error {
	s := "UPDATE ? SET ? WHERE ?;"
	return mm.dbConn.Exec(s, tableName, fields, value).Error
}

func (mm *MysqlMgr) Delete(tableName, where string) error {
	s := "DELETE FROM ? WHERE ?;"
	return mm.dbConn.Exec(s, tableName, where).Error
}
