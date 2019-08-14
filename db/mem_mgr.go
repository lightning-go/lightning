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
	"strconv"
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

func (mm *MemMgr) getSpecialPKValue(keys ...interface{}) string {
	n := len(keys)
	if n == 0 {
		return ""
	}

	n--
	var keyStr strings.Builder

	for idx, k := range keys {
		val := ""
		switch k.(type) {
		case string:
			val = k.(string)
		case float64:
			val = fmt.Sprintf("%v", k.(float64))
		case int:
			val = strconv.Itoa(k.(int))
		default:
			continue
		}

		keyStr.WriteString(val)
		if idx < n {
			keyStr.WriteString(":")
		}
	}

	return keyStr.String()
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

		val := ""
		switch d.(type) {
		case string:
			val = d.(string)
		case float64:
			val = fmt.Sprintf("%v", d.(float64))
		case int:
			val = strconv.Itoa(d.(int))
		default:
			continue
		}

		s.WriteString(val)
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
		//mm.log.Error("HGET ik field failed", logger.Fields{
		//	"key":   key,
		//	"field": srvVal,
		//})
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
		//mm.log.Trace("mem query failed", logger.Fields{
		//	"ikey": ik,
		//	"key":  key,
		//})
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

func (mm *MemMgr) QueryByPKs(keys ...interface{}) *MemMode {
	key := mm.getSpecialPKValue(keys...)
	return mm.query(key, false)
}

func (mm *MemMgr) QueryByPK(key string) *MemMode {
	return mm.query(key, false)
}

func (mm *MemMgr) QueryByIK(ik, key string) *MemMode {
	return mm.queryIK(ik, key, false)
}

func (mm *MemMgr) QueryCheckDBByPKs(keys ...interface{}) *MemMode {
	key := mm.getSpecialPKValue(keys...)
	memMode := mm.QueryByPK(key)
	if memMode != nil {
		return memMode
	}
	return mm.QueryRowByPKs(mm.pk, keys...)
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

func (mm *MemMgr) QueryRowByPKs(field string, values ...interface{}) (memMode *MemMode) {
	pkLen := len(mm.pks)
	valLen := len(values)
	if pkLen != valLen {
		return nil
	}

	n := pkLen - 1
	var valStr strings.Builder
	for idx, pk := range mm.pks {
		if idx >= valLen {
			return nil
		}
		valStr.WriteString(pk)
		valStr.WriteString("=")

		val := values[idx]
		switch val.(type) {
		case string:
			valStr.WriteString("'")
			valStr.WriteString(val.(string))
			valStr.WriteString("'")
		case float64:
			v := fmt.Sprintf("%v", val.(float64))
			valStr.WriteString(v)
		case int:
			valStr.WriteString(strconv.Itoa(val.(int)))
		default:
			continue
		}

		if idx < n {
			valStr.WriteString(" and ")
		}
	}
	where := valStr.String()

	mm.dbMgr.QueryCond(mm.tableName, where, func(rows *sql.Rows) {
		memList := mm.LoadRows(rows, true)
		if memList == nil || len(memList) == 0 {
			return
		}
		memMode = memList[0]
	})
	return
}

func (mm *MemMgr) QueryRow(field, value string) (memMode *MemMode) {
	where := fmt.Sprintf("%s = '%s'", field, value)
	mm.dbMgr.QueryCond(mm.tableName, where, func(rows *sql.Rows) {
		memList := mm.LoadRows(rows, true)
		if memList == nil  || len(memList) == 0 {
			return
		}
		memMode = memList[0]
	})
	return
}

func (mm *MemMgr) LoadDB() (d []*MemMode) {
	mm.dbMgr.Query(mm.tableName, func(rows *sql.Rows) {
		d = mm.LoadRows(rows)
	})
	return
}

func (mm *MemMgr) LoadDBCond(where string) (d []*MemMode) {
	mm.dbMgr.QueryCond(mm.tableName, where, func(rows *sql.Rows) {
		d = mm.LoadRows(rows)
	})
	return
}

func (mm *MemMgr) LoadRows(rows *sql.Rows, justOne ...bool) []*MemMode {
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
			return []*MemMode{memMode}
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

func (mm *MemMgr) Update(memMode *MemMode, updateState ...bool) {
	us := true
	if len(updateState) > 0 && updateState[0] {
		us = updateState[0]
	}
	if us {
		memMode.State = MEM_STATE_UPDATE
	}
	mm.updateData(memMode)
}

func (mm *MemMgr) UpdateCheckDBByPks(field string, v interface{}, keys ...interface{}) *MemMode {
	memMode := mm.UpdateByPks(field, v, keys...)
	if memMode != nil {
		return memMode
	}
	memMode = mm.QueryRowByPKs(mm.pk, keys...)
	if memMode != nil {
		key := mm.getSpecialPKValue(keys...)
		memMode = mm.UpdateByPk(key, field, v)
		return memMode
	}
	return nil
}

func (mm *MemMgr) UpdateByPks(field string, v interface{}, keys ...interface{}) *MemMode {
	key := mm.getSpecialPKValue(keys...)
	return mm.UpdateByPk(key, field, v)
}

func (mm *MemMgr) UpdateByPk(key, field string, v interface{}) *MemMode {
	memMode := mm.QueryByPK(key)
	if memMode == nil {
		//mm.log.Error("key Error\n", logger.Fields{
		//	"key": key,
		//})
		return nil
	}

	value, ok := memMode.Data[field]
	if !ok {
		mm.log.Error("field Error\n", logger.Fields{
			"field": field,
		})
		return nil
	}

	memMode.Data[field] = v
	memMode.State = MEM_STATE_UPDATE
	mm.updateData(memMode)
	mm.UpdateIKField(field, value, v)

	return memMode
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
		mm.log.Error(err)
		return
	}

	pkValue := mm.getPKValue(memMode)
	mm.HSet(pkValue, v)
	mm.AddAllIK(memMode)
}

func (mm *MemMgr) DelDataCheckDBByPKs(keys ...interface{}) *MemMode {
	m := mm.DelDataByPKs(keys...)
	if m == nil {
		return nil
	}
	m.State = MEM_STATE_DEL
	mm.SyncMemMode(m)
	return m
}

func (mm *MemMgr) DelDataByPKs(keys ...interface{}) *MemMode {
	key := mm.getSpecialPKValue(keys...)
	return mm.DelDataByPK(key)
}

func (mm *MemMgr) DelDataCheckDBByPK(key string) *MemMode {
	m := mm.DelDataByPK(key)
	if m == nil {
		return nil
	}
	m.State = MEM_STATE_DEL
	mm.SyncMemMode(m)
	return m
}

func (mm *MemMgr) DelDataByPK(key string) *MemMode {
	memMode := mm.query(key, true)
	if memMode == nil {
		//mm.log.Error("key Error\n", logger.Fields{
		//	"key": key,
		//})
		return nil
	}

	conn := mm.rc.GetConn()
	if conn == nil {
		return nil
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

	return memMode
}

func (mm *MemMgr) DelDataByIK(ik, key string) {
	memMode := mm.queryIK(ik, key, true)
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

func (mm *MemMgr) DelByIK(ik, key string) {
	memMode := mm.QueryByIK(ik, key)
	if memMode == nil {
		mm.log.Error("key Error\n", logger.Fields{
			"key": key,
		})
		return
	}
	memMode.State = MEM_STATE_DEL
	mm.updateData(memMode)
}

func (mm *MemMgr) GetAllKeys() []string {
	v, err := mm.rc.HGetAll(mm.key)
	if err != nil || v == nil {
		mm.log.Error(err)
		return nil
	}

	array, ok := v.([]interface{})
	if !ok {
		mm.log.Error("get keys type error")
		return nil
	}

	keys := make([]string, 0)
	for _, v := range array {
		record, ok := v.([]byte)
		if !ok {
			continue
		}
		pkLen := len(mm.pk) + 1
		if len(record) < pkLen {
			continue
		}
		recordHead := record[:pkLen]
		keyHead := fmt.Sprintf("%s:", mm.pk)
		if !strings.EqualFold(string(recordHead), keyHead) {
			continue
		}
		keys = append(keys, string(record))
	}

	return keys
}

func (mm *MemMgr) Sync() {
	keys := mm.GetAllKeys()
	if keys == nil {
		return
	}

	conn := mm.rc.GetConn()
	if conn == nil {
		return
	}

	for _, k := range keys {
		mm.rc.PipeHGet(conn, mm.key, k)
	}
	mm.rc.PipeEnd(conn)
	count := len(keys)

	result := make([]interface{}, 0)
	for i := 0; i < count; i++ {
		v, err := mm.rc.PipeRecv(conn)
		if err != nil {
			mm.log.Error(err)
			continue
		}
		result = append(result, v)
	}
	mm.rc.CloseConn(conn)

	for _, v := range result {
		memMode := NewMemMode()
		err := jsoniter.Unmarshal(v.([]byte), memMode)
		if err != nil {
			fmt.Println(err)
			return
		}
		mm.SyncMemMode(memMode)
	}

}

func (mm *MemMgr) SyncMemMode(memMode *MemMode) {
	if memMode == nil {
		return
	}

	var err error
	switch memMode.State {
	case MEM_STATE_NEW:
		fields, values := mm.GetInsertFieldAndValue(memMode)
		err = mm.dbMgr.Insert(mm.tableName, fields, values)
	case MEM_STATE_UPDATE:
		fields, where := mm.GetUpdateFieldAndValue(memMode)
		err = mm.dbMgr.Update(mm.tableName, fields, where)
	case MEM_STATE_DEL:
		//where := fmt.Sprintf("%s = %s", mm.pk, memMode.Data[mm.pk].(string))
		where := mm.GetDeleteCond(memMode)
		err = mm.dbMgr.Delete(mm.tableName, where)
	}

	if err == nil {
		if memMode.State != MEM_STATE_DEL {
			memMode.State = MEM_STATE_ORI
			mm.updateData(memMode) //fix
		} else {
			key := mm.getPKValue(memMode)
			mm.DelDataByPK(key)
		}
	} else {
		mm.log.Error(err)
	}

}

func (mm *MemMgr) GetInsertFieldAndValue(memMode *MemMode) (fields, values string) {
	if memMode == nil {
		return
	}

	dataLen := len(memMode.Data)
	index := 0
	var fieldStr strings.Builder
	var valueStr strings.Builder

	for k, v := range memMode.Data {
		fieldStr.WriteString(k)

		switch v.(type) {
		case string:
			valueStr.WriteString("'")
			valueStr.WriteString(v.(string))
			valueStr.WriteString("'")
		case int:
			valueStr.WriteString(strconv.Itoa(v.(int)))
		case float64:
			val := fmt.Sprintf("%v", v.(float64))
			valueStr.WriteString(val)
		case map[string]interface{}:
			d, err := jsoniter.Marshal(v)
			if err != nil {
				valueStr.WriteString("'{}'")
			} else {
				valueStr.WriteString("'")
				valueStr.WriteString(string(d))
				valueStr.WriteString("'")
			}
		default:
			valueStr.WriteString("'{}'")
		}

		if index+1 < dataLen {
			fieldStr.WriteString(", ")
			valueStr.WriteString(",")
		}
		index++
	}

	fields = fieldStr.String()
	values = valueStr.String()
	return
}

func (mm *MemMgr) GetUpdateFieldAndValue(memMode *MemMode) (fields, where string) {
	if memMode == nil {
		return
	}

	dataLen := len(memMode.Data)
	index := 0
	var fieldStr strings.Builder

	for k, v := range memMode.Data {
		switch v.(type) {
		case string:
			fieldStr.WriteString(k)
			fieldStr.WriteString("=")
			fieldStr.WriteString("'")
			fieldStr.WriteString(v.(string))
			fieldStr.WriteString("'")
		case float64:
			fieldStr.WriteString(k)
			fieldStr.WriteString("=")
			val := fmt.Sprintf("%v", v.(float64))
			fieldStr.WriteString(val)
		case int:
			fieldStr.WriteString(k)
			fieldStr.WriteString("=")
			fieldStr.WriteString(strconv.Itoa(v.(int)))
		case map[string]interface{}:
			d, err := jsoniter.Marshal(v)
			if err != nil {
				fieldStr.WriteString(k)
				fieldStr.WriteString("=")
				fieldStr.WriteString("'{}'")
			} else {
				fieldStr.WriteString(k)
				fieldStr.WriteString("=")
				fieldStr.WriteString("'")
				fieldStr.WriteString(string(d))
				fieldStr.WriteString("'")
			}
		default:
			fieldStr.WriteString(k)
			fieldStr.WriteString("=")
			fieldStr.WriteString("'{}'")
		}

		if index+1 < dataLen {
			fieldStr.WriteString(", ")
		}
		index++
	}

	fields = fieldStr.String()
	where = mm.GetPKCond(memMode)
	return
}

func (mm *MemMgr) GetPKCond(memMode *MemMode) string {
	if memMode == nil {
		return ""
	}

	var whereStr strings.Builder
	pkLen := len(mm.pks) - 1

	for idx, key := range mm.pks {
		d, ok := memMode.Data[key]
		if !ok {
			continue
		}

		switch d.(type) {
		case string:
			whereStr.WriteString(key)
			whereStr.WriteString("=")
			whereStr.WriteString("'")
			whereStr.WriteString(d.(string))
			whereStr.WriteString("'")
		case float64:
			whereStr.WriteString(key)
			whereStr.WriteString("=")
			whereStr.WriteString(fmt.Sprintf("%v", d.(float64)))
		case int:
			whereStr.WriteString(key)
			whereStr.WriteString("=")
			whereStr.WriteString(strconv.Itoa(d.(int)))
		}

		if idx < pkLen {
			whereStr.WriteString(" and ")
		}
	}

	return whereStr.String()
}

func (mm *MemMgr) GetDeleteCond(memMode *MemMode) (where string) {
	if memMode == nil {
		return
	}
	where = mm.GetPKCond(memMode)
	return
}