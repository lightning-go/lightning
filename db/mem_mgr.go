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
	"database/sql"
	"github.com/json-iterator/go"
)

const memLogPath = "./logs/redis.log"

const (
	MEM_STATE_ORI    = iota
	MEM_STATE_NEW
	MEM_STATE_UPDATE
	MEM_STATE_DEL
)

type MemMode struct {
	State int                    `json:"state"`
	Data  map[string]interface{} `json:"data"`
}

func NewMemMode() *MemMode {
	return &MemMode{
		State: MEM_STATE_ORI,
		Data:  make(map[string]interface{}),
	}
}

type MemMgr struct {
	rc        *RedisClient
	dbName    string
	tableName string
	key       string
	pKey      string
	pk        string
	queryPk   string
	pks       []string
	iks       []string
	isPkIncr  bool
	log       *logger.Logger
	dbMgr     IDBMgr
}

func NewMemMgr(rc *RedisClient, initPK bool, dbName, tableName string, pks []string, iks ...string) *MemMgr {
	if rc == nil {
		logger.Error("redis client is nil")
		return nil
	}

	pksLen := len(pks)
	if pksLen == 0 {
		logger.Error("pk is nil")
		return nil
	}

	mm := &MemMgr{
		rc:        rc,
		dbName:    dbName,
		tableName: tableName,
		key:       fmt.Sprintf("%s:%s", dbName, tableName),
		pKey:      fmt.Sprintf("%s:%s:pk", dbName, tableName),
		isPkIncr:  false,
		log:       logger.NewLogger(logger.TRACE),
		dbMgr:     GetDB(dbName),
	}

	if mm.dbMgr == nil {
		return nil
	}
	if mm.log == nil {
		return nil
	}

	n := pksLen - 1
	var str strings.Builder
	var str2 strings.Builder
	for idx, v := range pks {
		str.WriteString(v)
		str2.WriteString(v)
		if idx < n {
			str.WriteString(":")
			str2.WriteString(",")
		}
	}
	mm.pks = pks
	mm.pk = str.String()
	mm.queryPk = str2.String()

	for _, v := range iks {
		mm.iks = append(mm.iks, v)
	}

	if initPK && pksLen == 1 {
		mm.isPkIncr = true
		mm.initPKValue()
	}

	return mm
}

func (mm *MemMgr) SetLogRotation(maxAge, rotationTime time.Duration, pathFile string) {
	if mm.log == nil {
		mm.log = logger.NewLogger(logger.TRACE)
	}
	mm.log.SetRotation(maxAge, rotationTime, pathFile)
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

func (mm *MemMgr) existIK(ik string) (int, bool) {
	for index, data := range mm.iks {
		if ik == data {
			return index, true
		}
	}
	return -1, false
}

func (mm *MemMgr) getPKValue(memMode *MemMode) string {
	if memMode == nil {
		return ""
	}
	n := len(mm.pks) - 1
	var s strings.Builder
	for idx, key := range mm.pks {
		d, ok := memMode.Data[key]
		if !ok {
			continue
		}
		s.WriteString(d.(string))
		if idx < n {
			s.WriteString(":")
		}
	}
	return s.String()
}

func (mm *MemMgr) UpdateIKField(key string, srcField, desField interface{}) {
	_, ok := mm.existIK(key)
	if !ok {
		return
	}

	srvVal, ok := srcField.(string)
	if !ok {
		return
	}
	desVal, ok := desField.(string)
	if !ok {
		return
	}

	v := mm.HGetIK(key, srvVal)
	if v == nil {
		mm.log.Error("HGET ik field failed", logger.Fields{
			"key":   key,
			"field": srvVal,
		})
		return
	}

	conn := mm.rc.GetConn()
	if conn == nil {
		return
	}

	keyName := mm.produceFieldKey(key, srvVal)
	mm.rc.PipeHDel(conn, mm.key, keyName)

	keyName = mm.produceFieldKey(key, desVal)
	mm.rc.PipeHSet(conn, mm.key, keyName, v)

	mm.rc.PipeEnd(conn)
	mm.rc.CloseConn(conn)
}

func (mm *MemMgr) AddAllIK(memMode *MemMode) {
	if memMode == nil {
		return
	}
	conn := mm.rc.GetConn()
	if conn == nil {
		return
	}
	pkValue := mm.getPKValue(memMode)
	if len(pkValue) == 0 {
		return
	}

	for _, ik := range mm.iks {
		for k, v := range memMode.Data {
			if k != ik {
				continue
			}
			field := mm.produceFieldKey(ik, v.(string))
			mm.rc.PipeHSet(conn, mm.key, field, pkValue)
		}
	}

	mm.rc.PipeEnd(conn)
	mm.rc.CloseConn(conn)
}

func (mm *MemMgr) query(key string, isAll bool) *MemMode {
	v := mm.HGet(key)
	if v == nil {
		return nil
	}

	d, ok := v.([]byte)
	if !ok {
		mm.log.Error("convert type error", logger.Fields{
			"key":   key,
			"value": v,
		})
		return nil
	}

	memMode := NewMemMode()
	err := jsoniter.Unmarshal(d, memMode)
	if err != nil {
		mm.log.Error(err)
		return nil
	}

	if !isAll && memMode.State == MEM_STATE_DEL {
		return nil
	}

	return memMode
}

func (mm *MemMgr) queryIK(ik, key string, isAll bool) *MemMode {
	v := mm.HGetIK(ik, key)
	if v == nil {
		mm.log.Trace("mem query failed", logger.Fields{
			"ikey": ik,
			"key":  key,
		})
		return nil
	}
	k, ok := v.([]byte)
	if !ok {
		mm.log.Error("query key error", logger.Fields{
			"key": k,
		})
	}
	return mm.query(string(k), isAll)
}

func (mm *MemMgr) QueryByPK(key string) *MemMode {
	return mm.query(key, false)
}

func (mm *MemMgr) QueryByIK(ik, key string) *MemMode {
	return mm.queryIK(ik, key, false)
}

func (mm *MemMgr) QueryCheckDBByPK(key string) *MemMode {
	memMode := mm.QueryByPK(key)
	if memMode != nil {
		return memMode
	}
	return mm.QueryRow(mm.pk, key)
}

func (mm *MemMgr) QueryCheckDBByIK(ik, key string) *MemMode {
	memMode := mm.QueryByIK(ik, key)
	if memMode != nil {
		return memMode
	}
	return mm.QueryRow(ik, key)
}

func (mm *MemMgr) QueryRow(field, value string) (memMode *MemMode) {
	where := fmt.Sprintf("%s = '%s'", field, value)
	mm.dbMgr.QueryCond(mm.tableName, where, func(rows *sql.Rows) {
		imem := mm.LoadRows(rows, true)
		if imem == nil {
			return
		}
		mem, ok := imem.(*MemMode)
		if !ok {
			return
		}
		memMode = mem
	})
	return
}

func (mm *MemMgr) LoadDB() {
	mm.dbMgr.Query(mm.tableName, func(rows *sql.Rows) {
		mm.LoadRows(rows)
	})
}

func (mm *MemMgr) LoadRows(rows *sql.Rows, justOne ...bool) interface{} {
	if rows == nil {
		return nil
	}

	conn := mm.rc.GetConn()
	if conn == nil {
		return nil
	}

	one := false
	if len(justOne) > 0 {
		one = justOne[0]
	}

	columns, err := rows.Columns()
	if err != nil {
		mm.log.Error(err)
		return nil
	}

	columnsLen := len(columns)
	scanArgs := make([]interface{}, columnsLen)
	values := make([]interface{}, columnsLen)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var memModes []*MemMode = nil

	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			mm.log.Error(err)
			return nil
		}

		memMode := NewMemMode()
		for k, v := range values {
			if k >= columnsLen {
				break
			}
			if v == nil {
				continue
			}
			byteValue, ok := v.([]byte)
			if !ok {
				continue
			}
			key := columns[k]
			memMode.Data[key] = string(byteValue)
		}

		keyValue := mm.getPKValue(memMode)

		row, err := jsoniter.Marshal(memMode)
		if err != nil {
			mm.log.Error(err)
			continue
		}

		field := mm.produceFieldKey(mm.pk, keyValue)
		mm.rc.PipeHSet(conn, mm.key, field, row)

		for _, ik := range mm.iks {
			idx, ok := mm.existIK(ik)
			if !ok || idx >= columnsLen {
				continue
			}
			byteValue, ok := values[idx].([]byte)
			if !ok {
				continue
			}
			field := mm.produceFieldKey(ik, string(byteValue))
			mm.rc.PipeHSet(conn, mm.key, field, keyValue)
		}

		if one {
			mm.rc.PipeEnd(conn)
			mm.rc.CloseConn(conn)
			return memMode
		}

		memModes = append(memModes, memMode)
	}

	mm.rc.PipeEnd(conn)
	mm.rc.CloseConn(conn)
	return memModes
}

func (mm *MemMgr) updateData(memMode *MemMode) {
	if memMode == nil {
		return
	}
	v, err := jsoniter.Marshal(memMode)
	if err != nil {
		fmt.Println(err)
		return
	}

	pk := mm.getPKValue(memMode)
	mm.HSet(pk, v)
}

func (mm *MemMgr) Update(memMode *MemMode) {
	memMode.State = MEM_STATE_UPDATE
	mm.updateData(memMode)
}

func (mm *MemMgr) UpdateByPk(key, field string, v interface{}) {
	memMode := mm.QueryByPK(key)
	if memMode == nil {
		mm.log.Error("key Error\n", logger.Fields{
			"key": key,
		})
		return
	}

	value, ok := memMode.Data[field]
	if !ok {
		mm.log.Error("field Error\n", logger.Fields{
			"field": field,
		})
		return
	}

	memMode.Data[field] = v
	memMode.State = MEM_STATE_UPDATE
	mm.updateData(memMode)
	mm.UpdateIKField(field, value, v)
}

func (mm *MemMgr) UpdateByIk(ik, key string, field string, v interface{}) {
	memMode := mm.QueryByIK(ik, key)
	if memMode == nil {
		mm.log.Error("key Error\n", logger.Fields{
			"key": key,
		})
		return
	}

	value, ok := memMode.Data[field]
	if !ok {
		mm.log.Error("field Error\n", logger.Fields{
			"field": field,
		})
		return
	}

	memMode.Data[field] = v
	memMode.State = MEM_STATE_UPDATE
	mm.updateData(memMode)
	mm.UpdateIKField(field, value, v)
}

func (mm *MemMgr) AddData(memMode *MemMode, isPKIncrs ...bool) {
	if memMode == nil {
		return
	}

	if mm.isPkIncr {
		Id := mm.PKIncr()
		if Id == nil {
			mm.log.Error("Get new Id failed")
			return
		}
		newId := fmt.Sprintf("%d", Id.(int64))
		memMode.Data[mm.pk] = newId
	}

	memMode.State = MEM_STATE_NEW

	v, err := jsoniter.Marshal(memMode)
	if err != nil {
		fmt.Println(err)
		return
	}

	pkValue := mm.getPKValue(memMode)
	mm.HSet(pkValue, v)
	mm.AddAllIK(memMode)
}

func (mm *MemMgr) DelDataByPK(key string) {
	memMode := mm.query(key, true)
	if memMode == nil {
		mm.log.Error("key Error\n", logger.Fields{
			"key": key,
		})
		return
	}

	conn := mm.rc.GetConn()
	if conn == nil {
		return
	}

	for _, k := range mm.iks {
		v, ok := memMode.Data[k]
		if ok {
			keyName := mm.produceFieldKey(k, v.(string))
			mm.rc.PipeHDel(conn, mm.key, keyName)
		}
	}

	keyName := mm.produceFieldKey(mm.pk, key)
	mm.rc.PipeHDel(conn, mm.key, keyName)

	mm.rc.PipeEnd(conn)
	mm.rc.CloseConn(conn)
}

func (mm *MemMgr) DelDataByIK(ik, key string) {
	//todo
}

