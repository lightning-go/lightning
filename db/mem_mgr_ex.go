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
	"github.com/lightning-go/lightning/utils"
	"sync/atomic"
	"runtime/debug"
	"github.com/jinzhu/gorm"
)


const defaultLogPath = "./logs"

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
	State int
	Data  interface{}
}

type MemMgrEx struct {
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
	logPath := fmt.Sprintf("%v/mem_%v.log", defaultLogPath, tableName)
	log.SetRotation(-1, time.Hour*24, logPath)

	mm := &MemMgrEx{
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

func (mm *MemMgrEx) GetTableName() string {
	return mm.tableName
}

func (mm *MemMgrEx) GetDBMgr() *DBMgr {
	return mm.dbMgr
}

func (mm *MemMgrEx) initPKValue() {
	pkValue := mm.dbMgr.QueryPrimaryKey(mm.pk, mm.tableName)
	if pkValue == -1 {
		return
	}
	_, err := mm.rc.Set(mm.pKey, pkValue)
	if err != nil {
		mm.log.Error("init pk value failed")
	}
}

func (mm *MemMgrEx) pkIncr() interface{} {
	v, err := mm.rc.Incr(mm.pKey)
	if err != nil {
		mm.log.Error("pk value incr failed")
		return nil
	}
	return v
}

func (mm *MemMgrEx) GetPKIncr() int64 {
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

func (mm *MemMgrEx) setLogRotation(maxAge, rotationTime int, pathFile string) {
	if mm.log == nil {
		mm.log = logger.NewLogger(logger.TRACE)
	}
	mm.log.SetRotation(time.Minute*time.Duration(maxAge), time.Minute*time.Duration(rotationTime), pathFile)
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

func (mm *MemMgrEx) GetMultiPKValue(keyValMap map[string]interface{}) (string, string) {
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
		keyStr.WriteString(val)

		cond := mm.GetCond(k, v)
		if len(cond) > 0 {
			condStr.WriteString(cond)
		}

		if idx < n {
			keyStr.WriteString(":")
			condStr.WriteString(" and ")
		}
		idx++
	}

	return keyStr.String(), condStr.String()
}

func (mm *MemMgrEx) GetCond(key string, val interface{}) string {
	var whereStr strings.Builder

	switch val.(type) {
	case string:
		whereStr.WriteString(key)
		whereStr.WriteString("=")
		whereStr.WriteString("'")
		whereStr.WriteString(val.(string))
		whereStr.WriteString("'")
	case float64:
		whereStr.WriteString(key)
		whereStr.WriteString("=")
		whereStr.WriteString(fmt.Sprintf("%v", val.(float64)))
	case int:
		whereStr.WriteString(key)
		whereStr.WriteString("=")
		whereStr.WriteString(strconv.Itoa(val.(int)))
	case int64:
		whereStr.WriteString(key)
		whereStr.WriteString("=")
		whereStr.WriteString(strconv.FormatInt(val.(int64), 10))
	}

	return whereStr.String()
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

func (mm *MemMgrEx) SetDataIK(ikName string, ikValue, pkValue interface{}) {
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

func (mm *MemMgrEx) setData(state int, key interface{}, d interface{}, saveDB bool) bool {
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

	err = mm.HSet(k, v)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	return true
}

func (mm *MemMgrEx) AddMem(key interface{}, d interface{}) bool {
	return mm.setData(MEM_STATE_ORI, key, d, false)
}

func (mm *MemMgrEx) AddMemByMultiPK(keyValMap map[string]interface{}, d interface{}) bool {
	keys, _ := mm.GetMultiPKValue(keyValMap)
	if len(keys) == 0 {
		return false
	}
	return mm.AddMem(keys, d)
}

func (mm *MemMgrEx) AddData(key interface{}, d interface{}) bool {
	return mm.setData(MEM_STATE_NEW, key, d, true)
}

func (mm *MemMgrEx) AddDataByMultiPK(keyValMap map[string]interface{}, d interface{}) bool {
	keys, _ := mm.GetMultiPKValue(keyValMap)
	if len(keys) == 0 {
		return false
	}
	return mm.AddData(keys, d)
}

func (mm *MemMgrEx) UpdateData(key interface{}, d interface{}) bool {
	return mm.setData(MEM_STATE_UPDATE, key, d, true)
}

func (mm *MemMgrEx) UpdateDataByMultiPK(keyValMap map[string]interface{}, d interface{}) bool {
	keys, _ := mm.GetMultiPKValue(keyValMap)
	if len(keys) == 0 {
		return false
	}
	return mm.UpdateData(keys, d)
}

func (mm *MemMgrEx) getData(key string) ([]byte, error) {
	v, err := mm.HGet(key)
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

func (mm *MemMgrEx) GetData(key interface{}, dest interface{}, checkDB ...bool) error {
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
				mm.setData(MEM_STATE_ORI, k, dest, false)
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

func (mm *MemMgrEx) GetDataByMultiPK(keyValMap map[string]interface{}, dest interface{}, checkDB ...bool) error {
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
				mm.setData(MEM_STATE_ORI, keys, dest, false)
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

func (mm *MemMgrEx) GetDataByIK(ikName string, ikValue, dest interface{}) error {
	ikVal, err := mm.convertKey(ikValue)
	if err != nil {
		mm.log.Error(err)
		return err
	}
	d, err := mm.HGetIK(ikName, ikVal)
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
	d, err = mm.HGet(pk, false)
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

func (mm *MemMgrEx) DelData(key interface{}) bool {
	k, err := mm.convertKey(key)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	cond := mm.GetCond(mm.pk, key)
	if len(cond) > 0 {
		mm.putQueue(MEM_STATE_DEL, cond)
	}

	err = mm.HDel(k)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	return true
}

func (mm *MemMgrEx) DelDataByMultiPK(keyValMap map[string]interface{}) bool {
	keys, cond := mm.GetMultiPKValue(keyValMap)
	if len(keys) == 0 {
		return false
	}

	mm.putQueue(MEM_STATE_DEL, cond)

	err := mm.HDel(keys)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	return true
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

func (mm *MemMgrEx) putQueue(state int, d interface{}) {
	if d == nil {
		return
	}
	v := atomic.LoadInt32(&mm.queueWorking)
	if v == 0 {
		mm.queueWait.Add(1)
		mm.enableQueue()
		mm.queueWait.Wait()
	}
	memMode := NewMemModeEx()
	memMode.State = state
	memMode.Data = d
	mm.queue.Put(memMode)
}

func (mm *MemMgrEx) enableQueue() {
	go func() {
		logger.Tracef("enable memMode queue")
		atomic.StoreInt32(&mm.queueWorking, 1)
		mm.queueWait.Done()

		defer func() {
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
				d, ok := t.(*MemModeEx)
				if !ok || d == nil {
					continue
				}
				if d.Data == nil {
					continue
				}
				mm.syncMemMode(d.State, d.Data)
				FreeMemModeEx(d)
			}
		}
	}()
}

func (mm *MemMgrEx) syncMemMode(state int, d interface{}) {
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
