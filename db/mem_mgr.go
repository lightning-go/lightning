/**
 * @author: Jason
 * Created: 19-5-3
 */

package db

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/json-iterator/go"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/utils"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)


const (
	MEM_STATE_ORI    = iota
	MEM_STATE_NEW
	MEM_STATE_UPDATE
	MEM_STATE_DEL
)

var ConvertTypeError = errors.New("convert type error")

var defaultMemModeExPool = sync.Pool{
	New: func() interface{} {
		return &MemMode{
			State: MEM_STATE_ORI,
			Data:  nil,
		}
	},
}

func NewMemMode() *MemMode {
	return defaultMemModeExPool.Get().(*MemMode)
}

func FreeMemMode(v *MemMode) {
	if v != nil {
		defaultMemModeExPool.Put(v)
	}
}

type MemMode struct {
	State int
	Data  interface{}
}

type MemMgr struct {
	rc           *RedisClient
	dbName       string
	tableName    string
	key          string
	pKey         string
	pk           string
	pks          []string
	isPkIncr     bool
	log          *logger.Logger
	dbMgr        *DBMgr
	queue        *utils.SafeQueue
	queueWorking int32
	queueWait    sync.WaitGroup
	expire		 int64
}

func NewMemMgr(rc *RedisClient, initPK bool, dbName, tableName string, pks []string) *MemMgr {
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

	mm := &MemMgr{
		rc:           rc,
		dbName:       dbName,
		tableName:    tableName,
		key:          fmt.Sprintf("%s:%s", dbName, tableName),
		pKey:         fmt.Sprintf("%s:%s:pk", dbName, tableName),
		pks:          pks,
		log:          log,
		dbMgr:        GetDB(dbName),
		queue:        utils.NewSafeQueue(),
		queueWorking: 0,
		expire: 	  int64(time.Second) * 86400,
	}

	n := pksLen - 1
	var str strings.Builder
	for idx, v := range pks {
		str.Write([]byte(v))
		if idx < n {
			str.Write([]byte(":"))
		}
	}
	mm.pk = str.String()

	if initPK && pksLen == 1 {
		mm.isPkIncr = true
		mm.initPKValue()
	}

	return mm
}

func (mm *MemMgr) GetTableName() string {
	return mm.tableName
}

func (mm *MemMgr) GetDBMgr() *DBMgr {
	return mm.dbMgr
}

func (mm *MemMgr) initPKValue() {
	pkValue := mm.dbMgr.QueryPrimaryKey(mm.pk, mm.tableName)
	if pkValue == -1 {
		return
	}
	_, err := mm.rc.Set(mm.pKey, pkValue)
	if err != nil {
		mm.log.Error("init pk value failed")
	}
}

func (mm *MemMgr) pkIncr() interface{} {
	v, err := mm.rc.Incr(mm.pKey)
	if err != nil {
		mm.log.Error("pk value incr failed")
		return nil
	}
	return v
}

func (mm *MemMgr) GetPKIncr() int64 {
	if mm.isPkIncr {
		Id := mm.pkIncr()
		if Id == nil {
			mm.log.Error("get new Id failed")
			return -1
		}
		return Id.(int64)
	}
	return -1
}

func (mm *MemMgr) SetLogLevel(lv int) {
	if mm.log == nil {
		mm.log = logger.NewLogger(lv)
	} else {
		mm.log.SetLevel(lv)
	}
}

func (mm *MemMgr) SetLogRotation(lv, maxAge, rotationTime int, pathFile string) {
	mm.SetLogLevel(lv)
	mm.log.SetRotation(time.Minute*time.Duration(maxAge), time.Minute*time.Duration(rotationTime), pathFile)
}

func (mm *MemMgr) producePriKey(key string) string {
	var d strings.Builder
	d.Write([]byte(mm.key))
	d.Write([]byte(":"))
	d.Write([]byte(mm.pk))
	d.Write([]byte(":"))
	d.Write([]byte(key))
	return d.String()
}

func (mm *MemMgr) produceIKey(ikName, ikValue string) string {
	var d strings.Builder
	d.Write([]byte(mm.key))
	d.Write([]byte(":"))
	d.Write([]byte(ikName))
	d.Write([]byte(":"))
	d.Write([]byte(ikValue))
	return d.String()
}

func (mm *MemMgr) produceFieldKey(prefix, key string) string {
	var d strings.Builder
	d.Write([]byte(prefix))
	d.Write([]byte(":"))
	d.Write([]byte(key))
	return d.String()
}

func (mm *MemMgr) produceSpecialKey(keys ...interface{}) string {
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
		keyStr.Write([]byte(val))
		if idx < n {
			keyStr.Write([]byte(":"))
		}
	}

	return keyStr.String()
}

func (mm *MemMgr) convertKey(v interface{}) (val string, err error) {
	switch v.(type) {
	case []byte:
		val = string(v.([]byte))
	case string:
		val = v.(string)
	case float64:
		val = fmt.Sprintf("%v", v.(float64))
	case int:
		val = strconv.Itoa(v.(int))
	case int32:
		val = strconv.Itoa(int(v.(int32)))
	case int64:
		val = strconv.FormatInt(v.(int64), 10)
	default:
		err = ConvertTypeError
	}
	return
}

func (mm *MemMgr) GetMultiPKValue(keyValMap map[string]interface{}) (string, string) {
	if keyValMap == nil {
		return "", ""
	}
	n := len(keyValMap)
	if n == 0 {
		return "", ""
	}

	n--
	idx := 0
	var keyStr strings.Builder
	var condStr strings.Builder

	for _, k := range mm.pks {
		v, ok := keyValMap[k]
		if !ok {
			return "", ""
		}
		val, err := mm.convertKey(v)
		if err != nil {
			continue
		}
		keyStr.Write([]byte(val))

		cond := mm.GetCond(k, v)
		if len(cond) > 0 {
			condStr.Write([]byte(cond))
		}

		if idx < n {
			keyStr.Write([]byte(":"))
			condStr.Write([]byte(" and "))
		}
		idx++
	}

	return keyStr.String(), condStr.String()
}

func (mm *MemMgr) GetCond(key string, val interface{}) string {
	var whereStr strings.Builder

	switch val.(type) {
	case string:
		whereStr.Write([]byte(key))
		whereStr.Write([]byte("="))
		whereStr.Write([]byte("'"))
		whereStr.Write([]byte(val.(string)))
		whereStr.Write([]byte("'"))
	case float64:
		whereStr.Write([]byte(key))
		whereStr.Write([]byte("="))
		whereStr.Write([]byte(fmt.Sprintf("%v", val.(float64))))
	case int:
		whereStr.Write([]byte(key))
		whereStr.Write([]byte("="))
		whereStr.Write([]byte(strconv.Itoa(val.(int))))
	case int64:
		whereStr.Write([]byte(key))
		whereStr.Write([]byte("="))
		whereStr.Write([]byte(strconv.FormatInt(val.(int64), 10)))
	}

	return whereStr.String()
}

func (mm *MemMgr) SetExpire(second int64) {
	mm.expire = second
}

func (mm *MemMgr) Expire(key string, second int64) (err error) {
	keyName := mm.producePriKey(key)
	_, err = mm.rc.Expire(keyName, second)
	return err
}

func (mm *MemMgr) Set(key string, v interface{}, expire ...int64) (err error) {
	keyName := mm.producePriKey(key)
	exp := mm.expire
	if len(expire) > 0 {
		exp = expire[0]
	}
	_, err = mm.rc.Set(keyName, v, exp)
	return err
}

func (mm *MemMgr) Get(key string) (d interface{}, err error) {
	keyName := mm.producePriKey(key)
	d, err = mm.rc.Get(keyName)
	return
}

func (mm *MemMgr) MSet(kv map[string]interface{}) (err error) {
	if kv == nil {
		return
	}
	kvs := make([]interface{}, 0)
	for k, v := range kv {
		keyName := mm.producePriKey(k)
		kvs = append(kvs, keyName, v)
	}
	_, err = mm.rc.MSet(kvs)
	return
}

func (mm *MemMgr) MGet(keys ...string) (v []interface{}, err error) {
	kvs := make([]interface{}, 0)
	for _, v := range keys {
		keyName := mm.producePriKey(v)
		kvs = append(kvs, keyName)
	}
	v, err = mm.rc.MGet(kvs)
	return
}

func (mm *MemMgr) SetIK(ikName, ikValue string, v interface{}, expire ...int64) (err error) {
	keyName := mm.produceIKey(ikName, ikValue)
	exp := mm.expire
	if len(expire) > 0 {
		exp = expire[0]
	}
	_, err = mm.rc.Set(keyName, v, exp)
	return
}

func (mm *MemMgr) GetIK(ikName, ikValue string) (d interface{}, err error) {
	keyName := mm.produceIKey(ikName, ikValue)
	d, err = mm.rc.Get(keyName)
	return
}

func (mm *MemMgr) HSet(key string, v interface{}) (err error) {
	field := mm.produceFieldKey(mm.pk, key)
	_, err = mm.rc.HSet(mm.key, field, v)
	return
}

func (mm *MemMgr) HGet(key string, produce ...bool) (d interface{}, err error) {
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

func (mm *MemMgr) HSetIK(ik, key string, v interface{}) (err error) {
	field := mm.produceFieldKey(ik, key)
	_, err = mm.rc.HSet(mm.key, field, v)
	return
}

func (mm *MemMgr) HGetIK(ik, key string) (d interface{}, err error) {
	field := mm.produceFieldKey(ik, key)
	d, err = mm.rc.HGet(mm.key, field)
	return
}

func (mm *MemMgr) Del(key string) (err error) {
	keyName := mm.producePriKey(key)
	_, err = mm.rc.Del(keyName)
	return err
}

func (mm *MemMgr) HDel(key string) (err error) {
	field := mm.produceFieldKey(mm.pk, key)
	_, err = mm.rc.HDel(mm.key, field)
	return err
}

func (mm *MemMgr) HDelIK(ik, key string) (err error) {
	field := mm.produceFieldKey(ik, key)
	_, err = mm.rc.HDel(mm.key, field)
	return
}

func (mm *MemMgr) DelIK(ikName, ikValue string) (err error) {
	keyName := mm.produceIKey(ikName, ikValue)
	_, err = mm.rc.Del(keyName)
	return
}

func (mm *MemMgr) SetDataIK(ikName string, ikValue, pkValue interface{}, expire ...int64) {
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
	mm.SetIK(ikName, ikVal, pkVal, expire...)
}

func (mm *MemMgr) setData(state int, key interface{}, d interface{}, saveDB bool, expire ...int64) bool {
	if d == nil {
		return false
	}
	k, err := mm.convertKey(key)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	v, err := jsoniter.Marshal(d)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	if saveDB {
		mm.putQueue(state, d)
	}

	err = mm.Set(k, v, expire...)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	return true
}

func (mm *MemMgr) AddMem(key interface{}, d interface{}, expire ...int64) bool {
	return mm.setData(MEM_STATE_ORI, key, d, false, expire...)
}

func (mm *MemMgr) AddMemByMultiPK(keyValMap map[string]interface{}, d interface{}, expire ...int64) bool {
	keys, _ := mm.GetMultiPKValue(keyValMap)
	if len(keys) == 0 {
		return false
	}
	return mm.AddMem(keys, d, expire...)
}

func (mm *MemMgr) AddData(key interface{}, d interface{}, expire ...int64) bool {
	return mm.setData(MEM_STATE_NEW, key, d, true, expire...)
}

func (mm *MemMgr) AddDataByMultiPK(keyValMap map[string]interface{}, d interface{}, expire ...int64) bool {
	keys, _ := mm.GetMultiPKValue(keyValMap)
	if len(keys) == 0 {
		return false
	}
	return mm.AddData(keys, d, expire...)
}

func (mm *MemMgr) UpdateData(key interface{}, d interface{}, expire ...int64) bool {
	return mm.setData(MEM_STATE_UPDATE, key, d, true, expire...)
}

func (mm *MemMgr) UpdateDataByMultiPK(keyValMap map[string]interface{}, d interface{}, expire ...int64) bool {
	keys, _ := mm.GetMultiPKValue(keyValMap)
	if len(keys) == 0 {
		return false
	}
	return mm.UpdateData(keys, d, expire...)
}

func (mm *MemMgr) getData(key string) ([]byte, error) {
	v, err := mm.Get(key)
	if err != nil {
		mm.log.Error(err)
		return nil, err
	}
	if v == nil {
		return nil, gorm.ErrRecordNotFound
	}

	d, ok := v.([]byte)
	if !ok {
		mm.log.Error("convert type error", logger.Fields{
			"key":   key,
			"value": v,
		})
		return nil, ConvertTypeError
	}

	return d, nil
}

func (mm *MemMgr) GetData(key interface{}, dest interface{}, checkDB ...bool) error {
	k, err := mm.convertKey(key)
	if err != nil {
		mm.log.Error(err)
		return nil
	}

	d, err := mm.getData(k)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fromDB := true
			if len(checkDB) > 0 {
				fromDB = checkDB[0]
			}
			if fromDB {
				cond := mm.GetCond(mm.pk, key)
				err = mm.dbMgr.QueryRecord(mm.tableName, cond, dest)
				if err != nil {
					mm.log.Error(err)
					return err
				}
				mm.setData(MEM_STATE_ORI, k, dest, false, mm.expire)
			}
		}
		mm.log.Error(err)
		return err
	}

	err = jsoniter.Unmarshal(d, dest)
	if err != nil {
		mm.log.Error(err)
		return err
	}
	return nil
}

func (mm *MemMgr) GetDataByMultiPK(keyValMap map[string]interface{}, dest interface{}, checkDB ...bool) error {
	keys, cond := mm.GetMultiPKValue(keyValMap)
	if len(keys) == 0 {
		return nil
	}

	d, err := mm.getData(keys)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fromDB := true
			if len(checkDB) > 0 {
				fromDB = checkDB[0]
			}
			if fromDB {
				err = mm.dbMgr.QueryRecord(mm.tableName, cond, dest)
				if err != nil {
					mm.log.Error(err)
					return err
				}
				mm.setData(MEM_STATE_ORI, keys, dest, false, mm.expire)
			}
		}
		mm.log.Error(err)
		return err
	}

	err = jsoniter.Unmarshal(d, dest)
	if err != nil {
		mm.log.Error(err)
		return err
	}

	return nil
}

func (mm *MemMgr) GetDataByIK(ikName string, ikValue, dest interface{}) error {
	ikVal, err := mm.convertKey(ikValue)
	if err != nil {
		mm.log.Error(err)
		return err
	}
	d, err := mm.GetIK(ikName, ikVal)
	if err != nil {
		mm.log.Error(err)
		return err
	}
	if d == nil {
		return gorm.ErrRecordNotFound
	}

	v, ok := d.([]byte)
	if !ok {
		mm.log.Error("convert type error", logger.Fields{
			"ikName": ikName,
			"ikVal":  ikVal,
			"value":  d,
		})
		return ConvertTypeError
	}

	pk, err := mm.convertKey(v)
	if err != nil {
		mm.log.Error(err)
		return err
	}
	d, err = mm.Get(pk)
	if err != nil {
		mm.log.Error(err)
		return err
	}
	v, ok = d.([]byte)
	if !ok {
		mm.log.Error("convert type error", logger.Fields{
			"key":   pk,
			"value": v,
		})
		return ConvertTypeError
	}

	err = jsoniter.Unmarshal(v, dest)
	if err != nil {
		mm.log.Error(err)
		return err
	}

	return nil
}

func (mm *MemMgr) DelData(key interface{}) bool {
	k, err := mm.convertKey(key)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	cond := mm.GetCond(mm.pk, key)
	if len(cond) > 0 {
		mm.putQueue(MEM_STATE_DEL, cond)
	}

	err = mm.Del(k)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	return true
}

func (mm *MemMgr) DelDataByMultiPK(keyValMap map[string]interface{}) bool {
	keys, cond := mm.GetMultiPKValue(keyValMap)
	if len(keys) == 0 {
		return false
	}

	mm.putQueue(MEM_STATE_DEL, cond)

	err := mm.Del(keys)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	return true
}

func (mm *MemMgr) DelDataIK(ikName string, ikValue interface{}) bool {
	ikVal, err := mm.convertKey(ikValue)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	err = mm.DelIK(ikName, ikVal)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	return true
}

func (mm *MemMgr) putQueue(state int, d interface{}) {
	if d == nil {
		return
	}
	v := atomic.LoadInt32(&mm.queueWorking)
	if v == 0 {
		mm.queueWait.Add(1)
		mm.enableQueue()
		mm.queueWait.Wait()
	}
	memMode := NewMemMode()
	memMode.State = state
	memMode.Data = d
	mm.queue.Put(memMode)
}

func (mm *MemMgr) enableQueue() {
	go func() {
		logger.Tracef("enable memMode queue")
		atomic.StoreInt32(&mm.queueWorking, 1)
		mm.queueWait.Done()

		defer func() {
			logger.Tracef("disable memMode queue")
			atomic.StoreInt32(&mm.queueWorking, 0)
			err := recover()
			if err != nil {
				logger.Error(err)
				logger.Error(string(debug.Stack()))
			}
		}()

		for {
			var taskList []interface{} = nil
			mm.queue.Take(&taskList)

			if taskList == nil {
				continue
			}
			for _, t := range taskList {
				if t == nil {
					continue
				}
				d, ok := t.(*MemMode)
				if !ok || d == nil {
					continue
				}
				if d.Data == nil {
					continue
				}
				mm.syncMemMode(d.State, d.Data)
				FreeMemMode(d)
			}
		}
	}()
}

func (mm *MemMgr) syncMemMode(state int, d interface{}) {
	if d == nil {
		return
	}
	var err error
	switch state {
	case MEM_STATE_NEW:
		err = mm.dbMgr.NewRecord(d)
	case MEM_STATE_UPDATE:
		err = mm.dbMgr.SaveRecord(d)
	case MEM_STATE_DEL:
		cond, ok := d.(string)
		if ok {
			err = mm.dbMgr.Delete(mm.tableName, cond)
		}
	}
	if err != nil {
		mm.log.Error(err)
	}
}


