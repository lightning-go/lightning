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
	"strconv"
	"sync"
	"github.com/json-iterator/go"
	"errors"
)

var ConvertTypeError = errors.New("convert type error")

var defaultMemModeExPool = sync.Pool{
	New: func() interface{} {
		return &MemModeEx{
			State: MEM_STATE_ORI,
			Data:  nil,
		}
	},
}

func NewMemModeEx() *MemModeEx {
	return defaultMemModeExPool.Get().(*MemModeEx)
}

func FreeMemModeEx(v *MemModeEx) {
	if v != nil {
		defaultMemModeExPool.Put(v)
	}
}

type MemModeEx struct {
	State int    `json:"state"`
	Data  []byte `json:"data"`
}

type MemMgrEx struct {
	rc                *RedisClient
	dbName            string
	tableName         string
	key               string
	pKey              string
	pk                string
	isPkIncr          bool
	log               *logger.Logger
	dbMgr             IDBMgr
	producePKCallback func() string
	produceIKCallback func() string
}

func NewMemMgrEx(rc *RedisClient, initPK bool, dbName, tableName string, pks []string) *MemMgrEx {
	if rc == nil {
		logger.Error("redis client is nil")
		return nil
	}

	pksLen := len(pks)
	if pksLen == 0 {
		logger.Error("pk is nil")
		return nil
	}

	log := logger.NewLogger(logger.TRACE)
	if log == nil {
		logger.Error("log create failed")
		return nil
	}

	mm := &MemMgrEx{
		rc:        rc,
		dbName:    dbName,
		tableName: tableName,
		key:       fmt.Sprintf("%s:%s", dbName, tableName),
		pKey:      fmt.Sprintf("%s:%s:pk", dbName, tableName),
		log:       log,
		dbMgr:     GetDB(dbName),
	}

	n := pksLen - 1
	var str strings.Builder
	for idx, v := range pks {
		str.WriteString(v)
		if idx < n {
			str.WriteString(":")
		}
	}
	mm.pk = str.String()

	if initPK && pksLen == 1 {
		mm.isPkIncr = true
		mm.initPKValue()
	}

	return mm
}

func (mm *MemMgrEx) initPKValue() {
	pkValue := mm.dbMgr.QueryPrimaryKey(mm.pk, mm.tableName)
	if pkValue == 0 {
		return
	}
	_, err := mm.rc.Set(mm.pKey, pkValue)
	if err != nil {
		mm.log.Error("init pk value failed")
	}
}

func (mm *MemMgrEx) PKIncr() interface{} {
	v, err := mm.rc.Incr(mm.pKey)
	if err != nil {
		mm.log.Error("pk value incr failed")
		return nil
	}
	return v
}

func (mm *MemMgrEx) GetPKIncr() int64 {
	if mm.isPkIncr {
		Id := mm.PKIncr()
		if Id == nil {
			mm.log.Error("Get new Id failed")
			return -1
		}
		return Id.(int64)
	}
	return -1
}

func (mm *MemMgrEx) SetLogRotation(maxAge, rotationTime time.Duration, pathFile string) {
	if mm.log == nil {
		mm.log = logger.NewLogger(logger.TRACE)
	}
	mm.log.SetRotation(maxAge, rotationTime, pathFile)
}

func (mm *MemMgrEx) Close() {
	mm.rc.Close()
}

func (mm *MemMgrEx) produceFieldKey(prefix, key string) string {
	var d strings.Builder
	d.WriteString(prefix)
	d.WriteString(":")
	d.WriteString(key)
	return d.String()
}

func (mm *MemMgrEx) produceSpecialKey(keys ...interface{}) string {
	n := len(keys)
	if n == 0 {
		return ""
	}
	n--
	var keyStr strings.Builder

	for idx, k := range keys {
		val, err := mm.convertKey(k)
		if err != nil {
			continue
		}
		keyStr.WriteString(val)
		if idx < n {
			keyStr.WriteString(":")
		}
	}

	return keyStr.String()
}

func (mm *MemMgrEx) convertKey(v interface{}) (val string, err error) {
	switch v.(type) {
	case []byte:
		val = string(v.([]byte))
	case string:
		val = v.(string)
	case float64:
		val = fmt.Sprintf("%v", v.(float64))
	case int:
		val = strconv.Itoa(v.(int))
	case int64:
		val = strconv.FormatInt(v.(int64), 10)
	default:
		err = ConvertTypeError
	}
	return
}

func (mm *MemMgrEx) convertData(d interface{}) *MemModeEx {
	v, ok := d.([]byte)
	if !ok {
		mm.log.Error("convert type error", logger.Fields{
			"value": d,
		})
		return nil
	}

	memMode := NewMemModeEx()
	err := jsoniter.Unmarshal(v, memMode)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	return memMode
}

func (mm *MemMgrEx) Set(key string, v interface{}, expire ...int64) (err error) {
	keyName := mm.produceFieldKey(mm.pk, key)
	_, err = mm.rc.Set(keyName, v, expire...)
	return err
}

func (mm *MemMgrEx) Get(key string) (d interface{}, err error) {
	keyName := mm.produceFieldKey(mm.pk, key)
	d, err = mm.rc.Get(keyName)
	return
}

func (mm *MemMgrEx) MSet(kv map[string]interface{}) (err error) {
	if kv == nil {
		return
	}
	kvs := make([]interface{}, 0)
	for k, v := range kv {
		keyName := mm.produceFieldKey(mm.pk, k)
		kvs = append(kvs, keyName, v)
	}
	_, err = mm.rc.MSet(kvs)
	return
}

func (mm *MemMgrEx) MGet(keys ...string) (v []string, err error) {
	kvs := make([]interface{}, 0)
	for _, v := range keys {
		keyName := mm.produceFieldKey(mm.pk, v)
		kvs = append(kvs, keyName)
	}
	v, err = mm.rc.MGet(kvs)
	return
}

func (mm *MemMgrEx) HSet(key string, v interface{}) (err error) {
	field := mm.produceFieldKey(mm.pk, key)
	_, err = mm.rc.HSet(mm.key, field, v)
	return
}

func (mm *MemMgrEx) HGet(key string, produce ...bool) (d interface{}, err error) {
	field := key
	produceKey := true
	if len(produce) > 0 {
		produceKey = produce[0]
	}
	if produceKey {
		field = mm.produceFieldKey(mm.pk, key)
	}
	d, err = mm.rc.HGet(mm.key, field)
	return
}

func (mm *MemMgrEx) HSetIK(ik, key string, v interface{}) (err error) {
	field := mm.produceFieldKey(ik, key)
	_, err = mm.rc.HSet(mm.key, field, v)
	return
}

func (mm *MemMgrEx) HGetIK(ik, key string) (d interface{}, err error) {
	field := mm.produceFieldKey(ik, key)
	d, err = mm.rc.HGet(mm.key, field)
	return
}

func (mm *MemMgrEx) Del(key string) (err error) {
	keyName := mm.produceFieldKey(mm.pk, key)
	_, err = mm.rc.Del(keyName)
	return err
}

func (mm *MemMgrEx) HDel(key string) (err error) {
	field := mm.produceFieldKey(mm.pk, key)
	_, err = mm.rc.HDel(mm.key, field)
	return err
}

func (mm *MemMgrEx) HDelIK(ik, key string) (err error) {
	field := mm.produceFieldKey(ik, key)
	_, err = mm.rc.HDel(mm.key, field)
	return
}

func (mm *MemMgrEx) SetData(key interface{}, d interface{}) bool {
	if d == nil {
		return false
	}
	k, err := mm.convertKey(key)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	d1, err := jsoniter.Marshal(d)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	memMode := NewMemModeEx()
	memMode.State = MEM_STATE_NEW
	memMode.Data = d1

	v, err := jsoniter.Marshal(memMode)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	err = mm.HSet(k, v)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	return true
}

func (mm *MemMgrEx) SetDataByMultiPK(d interface{}, key ...interface{}) bool {
	k := mm.produceSpecialKey(key...)
	return mm.SetData(k, d)
}

func (mm *MemMgrEx) SetDataByIK(ikName string, ikValue, pkValue interface{}) {
	ikVal, err := mm.convertKey(ikValue)
	if err != nil {
		mm.log.Error(err)
		return
	}
	pkVal, err := mm.convertKey(pkValue)
	if err != nil {
		mm.log.Error(err)
		return
	}
	value := mm.produceFieldKey(mm.pk, pkVal)
	mm.HSetIK(ikName, ikVal, value)
}

func (mm *MemMgrEx) GetData(key interface{}) []byte {
	k, err := mm.convertKey(key)
	if err != nil {
		mm.log.Error(err)
		return nil
	}

	v, err := mm.HGet(k)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	if v == nil {
		return nil
	}

	d, ok := v.([]byte)
	if !ok {
		mm.log.Error("convert type error", logger.Fields{
			"key":   k,
			"value": v,
		})
		return nil
	}

	memMode := NewMemModeEx()
	err = jsoniter.Unmarshal(d, memMode)
	if err != nil {
		mm.log.Error(err)
		return nil
	}

	return memMode.Data
}

func (mm *MemMgrEx) GetDataByMultiPK(key ...interface{}) []byte {
	k := mm.produceSpecialKey(key...)
	return mm.GetData(k)
}

func (mm *MemMgrEx) GetDataByIK(ikName string, ikValue interface{}) []byte {
	ikVal, err := mm.convertKey(ikValue)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	d, err := mm.HGetIK(ikName, ikVal)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	v, ok := d.([]byte)
	if !ok {
		mm.log.Error("convert type error", logger.Fields{
			"ikName": ikName,
			"ikVal":  ikVal,
			"value":  d,
		})
		return nil
	}

	pk, err := mm.convertKey(v)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	d, err = mm.HGet(pk, false)
	if err != nil {
		mm.log.Error(err)
		return nil
	}
	v, ok = d.([]byte)
	if !ok {
		mm.log.Error("convert type error", logger.Fields{
			"key":   pk,
			"value": v,
		})
		return nil
	}

	memMode := NewMemModeEx()
	err = jsoniter.Unmarshal(v, memMode)
	if err != nil {
		mm.log.Error(err)
		return nil
	}

	return memMode.Data
}

func (mm *MemMgrEx) DelData(key interface{}) bool {
	k, err := mm.convertKey(key)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	err = mm.HDel(k)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	return true
}

func (mm *MemMgrEx) DelDataByMultiPK(key ...interface{}) bool {
	k := mm.produceSpecialKey(key...)
	if len(k) == 0 {
		return false
	}
	return mm.DelData(k)
}

func (mm *MemMgrEx) DelDataByIK(ikName string, ikValue interface{}) bool {
	ikVal, err := mm.convertKey(ikValue)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	err = mm.HDelIK(ikName, ikVal)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	return true
}
