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
	"github.com/lightning-go/lightning/conf"
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

type IkCallback func(obj interface{})(ikField string, ikVal interface{})
type PkCallback func(obj interface{}) (pkVal interface{})

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
	isPkIncr     bool
	log          *logger.Logger
	dbMgr        *DBMgr
	queue        *utils.SafeQueue
	queueWorking int32
	queueWait    sync.WaitGroup
	expire		 int64
	ikCallbackList	[]IkCallback
	pkCallback		PkCallback
}

func NewMemMgr(rc *RedisClient, dbName, tableName, pk string,
	pkCallback PkCallback, logCfg ...*conf.LogConfig) *MemMgr {
	log := logger.NewLogger(logger.TRACE)
	if len(logCfg) > 0 {
		logConf := logCfg[0]
		if logConf != nil {
			logLv := logger.GetLevel(logConf.LogLevel)
			pathFile := logger.GetLogPathFile(logConf)
			log.SetLevel(logLv)
			log.SetRotation(time.Minute*time.Duration(logConf.MaxAge),
				time.Minute*time.Duration(logConf.RotationTime), pathFile)
		}
	}
	if rc == nil {
		log.Error("redis client nil")
		return nil
	}
	if len(pk) == 0 {
		log.Error("primary key error")
		return nil
	}
	if pkCallback == nil {
		log.Warn("primary key callback nil")
		return nil
	}

	mm := &MemMgr{
		rc:           rc,
		dbName:       dbName,
		tableName:    tableName,
		key:          fmt.Sprintf("%s:%s", dbName, tableName),
		pKey:         fmt.Sprintf("%s:%s:pk", dbName, tableName),
		pk:           pk,
		isPkIncr:	  true,
		log:          log,
		dbMgr:        GetDB(dbName),
		queue:        utils.NewSafeQueue(),
		queueWorking: 0,
		expire: 	  60 * 60 * 24,
		pkCallback:	  pkCallback,
	}

	mm.initPKValue()
	mm.log.Infof("table %v cache init ok", mm.tableName)
	return mm
}

func (mm *MemMgr) SetPkCallback(pkCallback PkCallback) {
	mm.pkCallback = pkCallback
}

func (mm *MemMgr) SetIKCallback(ikCallbackList []IkCallback) {
	if ikCallbackList == nil {
		return
	}
	mm.ikCallbackList = ikCallbackList
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
		id := mm.pkIncr()
		if id == nil {
			mm.log.Error("get new Id failed")
			return -1
		}
		Id, ok := id.(int64)
		if !ok {
			mm.log.Error("Id type error")
			return -1
		}
		return Id
	}
	return -1
}

func (mm *MemMgr) SetLogConf(logConf *conf.LogConfig) {
	if logConf == nil {
		return
	}
	logLv := logger.GetLevel(logConf.LogLevel)
	pathFile := logger.GetLogPathFile(logConf)
	mm.SetLogLevel(logLv)
	mm.SetLogRotation(logConf.MaxAge, logConf.RotationTime, pathFile)
}

func (mm *MemMgr) SetLogLevel(lv int) {
	if mm.log == nil {
		mm.log = logger.NewLogger(lv)
	} else {
		mm.log.SetLevel(lv)
	}
}

func (mm *MemMgr) SetLogRotation(maxAge, rotationTime int, pathFile string) {
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

func (mm *MemMgr) set(key string, v interface{}, expire ...int64) (err error) {
	keyName := mm.producePriKey(key)
	exp := mm.expire
	if len(expire) > 0 {
		exp = expire[0]
	}
	_, err = mm.rc.Set(keyName, v, exp)
	return err
}

func (mm *MemMgr) get(key string) (d interface{}, err error) {
	keyName := mm.producePriKey(key)
	d, err = mm.rc.Get(keyName)
	return
}

func (mm *MemMgr) mSet(kv map[string]interface{}) (err error) {
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

func (mm *MemMgr) mGet(keys ...string) (v []interface{}, err error) {
	kvs := make([]interface{}, 0)
	for _, v := range keys {
		keyName := mm.producePriKey(v)
		kvs = append(kvs, keyName)
	}
	v, err = mm.rc.MGet(kvs)
	return
}

func (mm *MemMgr) setIK(ikName, ikValue string, v interface{}, expire ...int64) (err error) {
	keyName := mm.produceIKey(ikName, ikValue)
	exp := mm.expire
	if len(expire) > 0 {
		exp = expire[0]
	}
	_, err = mm.rc.Set(keyName, v, exp)
	return
}

func (mm *MemMgr) getIK(ikName, ikValue string) (d interface{}, err error) {
	keyName := mm.produceIKey(ikName, ikValue)
	d, err = mm.rc.Get(keyName)
	return
}

func (mm *MemMgr) hSet(key string, v interface{}) (err error) {
	field := mm.produceFieldKey(mm.pk, key)
	_, err = mm.rc.HSet(mm.key, field, v)
	return
}

func (mm *MemMgr) hGet(key string, produce ...bool) (d interface{}, err error) {
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

func (mm *MemMgr) hSetIK(ik, key string, v interface{}) (err error) {
	field := mm.produceFieldKey(ik, key)
	_, err = mm.rc.HSet(mm.key, field, v)
	return
}

func (mm *MemMgr) hGetIK(ik, key string) (d interface{}, err error) {
	field := mm.produceFieldKey(ik, key)
	d, err = mm.rc.HGet(mm.key, field)
	return
}

func (mm *MemMgr) del(key string) (err error) {
	keyName := mm.producePriKey(key)
	_, err = mm.rc.Del(keyName)
	return err
}

func (mm *MemMgr) hDel(key string) (err error) {
	field := mm.produceFieldKey(mm.pk, key)
	_, err = mm.rc.HDel(mm.key, field)
	return err
}

func (mm *MemMgr) hDelIK(ik, key string) (err error) {
	field := mm.produceFieldKey(ik, key)
	_, err = mm.rc.HDel(mm.key, field)
	return
}

func (mm *MemMgr) delIK(ikName, ikValue string) (err error) {
	keyName := mm.produceIKey(ikName, ikValue)
	_, err = mm.rc.Del(keyName)
	return
}

func (mm *MemMgr) setDataIK(ikName string, ikValue, pkValue interface{}, expire ...int64) {
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
	err = mm.setIK(ikName, ikVal, pkVal, expire...)
	if err != nil {
		mm.log.Error(err)
	}
}

func (mm *MemMgr) setDataAllIk(dest interface{}, expire ...int64) {
	if dest == nil || mm.ikCallbackList == nil || mm.pkCallback == nil {
		return
	}
	pkVal := mm.pkCallback(dest)
	if pkVal == nil {
		return
	}
	for _, ikCallback := range mm.ikCallbackList {
		if ikCallback == nil {
			continue
		}
		ikField, ikVal := ikCallback(dest)
		if len(ikField) == 0 {
			continue
		}
		mm.setDataIK(ikField, ikVal, pkVal, expire...)
	}
}

func (mm *MemMgr) delDataAllIk(dest interface{}) {
	if dest == nil || mm.ikCallbackList == nil || mm.pkCallback == nil {
		return
	}
	for _, ikCallback := range mm.ikCallbackList {
		if ikCallback == nil {
			continue
		}
		ikField, ikVal := ikCallback(dest)
		if len(ikField) == 0 {
			continue
		}
		mm.delDataIK(ikField, ikVal)
	}
}

func (mm *MemMgr) CleanIkData(dest interface{}) {
	mm.delDataAllIk(dest)
}

func (mm *MemMgr) setData(state int, dest interface{}, saveDB bool, expire ...int64) bool {
	if dest == nil || mm.pkCallback == nil {
		return false
	}
	key := mm.pkCallback(dest)
	if key == nil {
		return false
	}
	k, err := mm.convertKey(key)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	v, err := jsoniter.Marshal(dest)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	if saveDB {
		mm.putQueue(state, dest)
	}

	err = mm.set(k, v, expire...)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	mm.setDataAllIk(dest, expire...)
	return true
}

func (mm *MemMgr) AddMem(d interface{}, expire ...int64) bool {
	return mm.setData(MEM_STATE_ORI, d, false, expire...)
}

func (mm *MemMgr) AddData(d interface{}, expire ...int64) bool {
	return mm.setData(MEM_STATE_NEW, d, true, expire...)
}

func (mm *MemMgr) UpdateData(d interface{}, expire ...int64) bool {
	return mm.setData(MEM_STATE_UPDATE, d, true, expire...)
}

func (mm *MemMgr) getData(key string) ([]byte, error) {
	v, err := mm.get(key)
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
					if err != gorm.ErrRecordNotFound {
						mm.log.Error(err)
					}
					return err
				}
				mm.setData(MEM_STATE_ORI, dest, false, mm.expire)
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

func (mm *MemMgr) GetDataByIK(ikName string, ikValue, dest interface{}, queryWhere...string) error {
	ikVal, err := mm.convertKey(ikValue)
	if err != nil {
		mm.log.Error(err)
		return err
	}
	d, err := mm.getIK(ikName, ikVal)
	if err != nil {
		mm.log.Error(err)
		return err
	}
	if d == nil {
		err = gorm.ErrRecordNotFound
		if len(queryWhere) > 0 {
			where := queryWhere[0]
			err = mm.dbMgr.QueryRecord(mm.tableName, where, dest)
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					mm.log.Error(err)
				}
				return err
			}
			mm.setData(MEM_STATE_ORI, dest, false, mm.expire)
		}
		return err
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

	return mm.GetData(v, dest)

}

func (mm *MemMgr) DelData(dest interface{}) bool {
	if dest == nil {
		return false
	}
	key := mm.pkCallback(dest)
	if key == nil {
		return false
	}
	k, err := mm.convertKey(key)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	cond := mm.GetCond(mm.pk, key)
	if len(cond) > 0 {
		mm.putQueue(MEM_STATE_DEL, cond)
	}

	err = mm.del(k)
	if err != nil {
		mm.log.Error(err)
		return false
	}

	mm.delDataAllIk(dest)
	return true
}

func (mm *MemMgr) delDataIK(ikName string, ikValue interface{}) bool {
	ikVal, err := mm.convertKey(ikValue)
	if err != nil {
		mm.log.Error(err)
		return false
	}
	err = mm.delIK(ikName, ikVal)
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
		mm.log.Tracef("enable memMode queue")
		atomic.StoreInt32(&mm.queueWorking, 1)
		mm.queueWait.Done()

		defer func() {
			mm.log.Tracef("disable memMode queue")
			atomic.StoreInt32(&mm.queueWorking, 0)
			err := recover()
			if err != nil {
				mm.log.Error(err)
				mm.log.Error(string(debug.Stack()))
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


