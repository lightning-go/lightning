/**
 * @author: Jason
 * Created: 19-5-4
 */

package db

import "sync"

var dbFactory = newDBFactory()

func AddDB(dbName string, dbMgr IDBMgr) {
	dbFactory.Add(dbName, dbMgr)
}

func GetDB(dbName string) IDBMgr {
	return dbFactory.Get(dbName)
}

func Del(dbName string) {
	dbFactory.Del(dbName)
}

type DBFactory struct {
	dbList *sync.Map
}

func newDBFactory() *DBFactory {
	return &DBFactory{
		dbList: &sync.Map{},
	}
}

func (db *DBFactory) Add(dbName string, dbMgr IDBMgr) {
	db.dbList.Store(dbName, dbMgr)
}

func (db *DBFactory) Get(dbName string) IDBMgr {
	idb, ok := db.dbList.Load(dbName)
	if !ok {
		return nil
	}
	dbMgr, ok := idb.(IDBMgr)
	if !ok {
		return nil
	}
	return dbMgr
}

func (db *DBFactory) Del(dbName string) {
	idb, ok := db.dbList.Load(dbName)
	if !ok {
		return
	}
	dbMgr, ok := idb.(IDBMgr)
	if !ok {
		return
	}
	dbMgr.Close()
	db.dbList.Delete(dbName)
}

func (db *DBFactory) Clean() {
	db.dbList.Range(func(key, value interface{}) bool {
		dbMgr, ok := value.(IDBMgr)
		if !ok {
			return true
		}
		dbMgr.Close()
		return true
	})
	db.dbList = &sync.Map{}
}

