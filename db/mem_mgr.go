/**
 * @author: Jason
 * Created: 19-5-3
 */

package db

import (
	"github.com/lightning-go/lightning/logger"
	"fmt"
	"strings"
	"time"
)

const memLogPath = "./logs/redis.log"

type MemMgr struct {
	rc        *RedisClient
	dbName    string
	tableName string
	key       string
	pKey      string
	pk        string
	pks       []string
	iks       []string
	log       *logger.Logger
	dbMgr     IDBMgr
}

func NewMemMgr(rc *RedisClient, initPK bool, dbName, tableName string, pks []string, iks ...string) *MemMgr {
	if rc == nil {
		logger.Info("redis client is nil")
		return nil
	}
	mm := &MemMgr{
		rc:        rc,
		dbName:    dbName,
		tableName: tableName,
		key:       fmt.Sprintf("%s:%s", dbName, tableName),
		pKey:      fmt.Sprintf("%s:%s:pk", dbName, tableName),
		log:       logger.NewLogger(logger.TRACE),
		dbMgr:     GetDB(dbName),
	}

	if mm.dbMgr == nil {
		return nil
	}

	if mm.log == nil {
		return nil
	}
	mm.log.SetRotation(time.Hour*24*30, time.Hour*24, memLogPath)

	n := len(pks) - 1
	var str strings.Builder
	for idx, v := range pks {
		mm.pks = append(mm.pks, v)
		str.WriteString(v)
		if idx < n {
			str.WriteString(":")
		}
	}
	mm.pk = str.String()

	for _, v := range iks {
		mm.iks = append(mm.iks, v)
	}

	if initPK {
		mm.initPKValue()
	}

	return mm
}

func (mm *MemMgr) initPKValue() {
	pkValue := mm.dbMgr.QueryPrimaryKey(mm.pk, mm.tableName)
	if pkValue == 0 {
		return
	}
	_, err := mm.rc.Set(mm.pKey, pkValue)
	if err != nil {
		mm.log.Error("init pk value failed")
	}
}

func (mm *MemMgr) Close() {
	mm.rc.Close()
}

func (mm *MemMgr) AddIK(iks ...string) {
	for _, v := range iks {
		mm.iks = append(mm.iks, v)
	}
}

func (mm *MemMgr) PKIncr() interface{} {
	v, err := mm.rc.Incr(mm.pKey)
	if err != nil {
		mm.log.Error("pk value incr failed")
		return nil
	}
	return v
}

func (mm *MemMgr) produceKey() string {
	var d strings.Builder
	d.WriteString(mm.dbName)
	d.WriteString(":")
	d.WriteString(mm.tableName)
	return d.String()
}

func (mm *MemMgr) produceFieldKey(prefix, key string) string {
	var d strings.Builder
	d.WriteString(prefix)
	d.WriteString(":")
	d.WriteString(key)
	return d.String()
}

func (mm *MemMgr) produceSuffixKey(suffix, key string) string {
	var d strings.Builder
	d.WriteString(key)
	d.WriteString(":")
	d.WriteString(suffix)
	return d.String()
}

func (mm *MemMgr) Get(key string) interface{} {
	keyName := mm.produceFieldKey(mm.pk, key)
	v, err := mm.rc.Get(keyName)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	return v
}

func (mm *MemMgr) MGet(keys ...string) (v []string) {
	kvLen := len(keys)
	kvs := make([]interface{}, kvLen)
	for idx, v := range keys {
		keyName := mm.produceFieldKey(mm.pk, v)
		kvs[idx] = keyName
	}

	v, err := mm.rc.MGet(kvs)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	return v
}

func (mm *MemMgr) HGet(key string) interface{} {
	keyName := mm.produceFieldKey(mm.pk, key)
	v, err := mm.rc.HGet(mm.key, keyName)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	return v
}

func (mm *MemMgr) HGetIK(ik, key string) interface{} {
	keyName := mm.produceFieldKey(ik, key)
	v, err := mm.rc.HGet(mm.key, keyName)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	return v
}

func (mm *MemMgr) Set(key string, v interface{}) {
	keyName := mm.produceFieldKey(mm.pk, key)
	_, err := mm.rc.Set(keyName, v)
	if err != nil {
		mm.log.Error(err)
	}
}

func (mm *MemMgr) MSet(kv map[string]interface{}) {
	if kv == nil {
		return
	}
	kvs := make([]interface{}, 0)
	for k, v := range kv {
		kvs = append(kvs, k, v)
	}
	mm.rc.MSet(kvs)
}

func (mm *MemMgr) HSet(key string, v interface{}) {
	field := mm.produceFieldKey(mm.pk, key)
	_, err := mm.rc.HSet(mm.key, field, v)
	if err != nil {
		mm.log.Error(err)
	}
}

func (mm *MemMgr) HSetIK(ik, key string, v interface{}) {
	field := mm.produceFieldKey(ik, key)
	_, err := mm.rc.HSet(mm.key, field, v)
	if err != nil {
		mm.log.Error(err)
	}
}

func (mm *MemMgr) Del(key string) {
	keyName := mm.produceFieldKey(mm.pk, key)
	_, err := mm.rc.Del(keyName)
	if err != nil {
		mm.log.Error(err)
	}
}

func (mm *MemMgr) HDel(key string) {
	keyName := mm.produceFieldKey(mm.pk, key)
	_, err := mm.rc.HDel(mm.key, keyName)
	if err != nil {
		mm.log.Error(err)
	}
}

func (mm *MemMgr) HDelIK(ik, key string) {
	keyName := mm.produceFieldKey(ik, key)
	_, err := mm.rc.HDel(mm.key, keyName)
	if err != nil {
		mm.log.Error(err)
	}
}

func (mm *MemMgr) LoadDB() {

}
