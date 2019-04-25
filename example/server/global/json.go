/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package global

import (
	"sync"
	"github.com/json-iterator/go"
	"github.com/lightning-go/lightning/utils"
)

var jsonMgrInstance *jsonMgr
var jsonMgrOnce sync.Once

func GetJSONMgr() *jsonMgr {
	jsonMgrOnce.Do(func() {
		jsonMgrInstance = &jsonMgr{}
	})
	return jsonMgrInstance
}

type jsonMgr struct{}

func (jsonmgr *jsonMgr) ExistInt(data []byte, path ...interface{}) (int, bool) {
	d := jsoniter.Get(data, path...)
	if d.LastError() != nil {
		return Invalid, false
	}
	return d.ToInt(), true
}

func (jsonmgr *jsonMgr) ExistString(data []byte, path ...interface{}) (string, bool) {
	d := jsoniter.Get(data, path...)
	if d.LastError() != nil {
		return "", false
	}
	return d.ToString(), true
}

func (jsonmgr *jsonMgr) ExistBytes(data []byte, path ...interface{}) ([]byte, bool) {
	d := jsoniter.Get(data, path...)
	if d.LastError() != nil {
		return nil, false
	}
	return []byte(d.ToString()), true
}

func (jsonmgr *jsonMgr) GetString(data []byte, path ...interface{}) string {
	return jsoniter.Get(data, path...).ToString()
}

func (jsonmgr *jsonMgr) GetBytes(data []byte, path ...interface{}) []byte {
	return []byte(jsoniter.Get(data, path...).ToString())
}

func (jsonmgr *jsonMgr) GetInt(data []byte, path ...interface{}) int {
	return jsoniter.Get(data, path...).ToInt()
}

func (jsonmgr *jsonMgr) GetInt64(data []byte, path ...interface{}) int64 {
	return jsoniter.Get(data, path...).ToInt64()
}

func (jsonmgr *jsonMgr) GetFloat32(data []byte, path ...interface{}) float32 {
	return jsoniter.Get(data, path...).ToFloat32()
}

func (jsonmgr *jsonMgr) GetFloat64(data []byte, path ...interface{}) float64 {
	return jsoniter.Get(data, path...).ToFloat64()
}

func (jsonmgr *jsonMgr) GetArray(data []byte, key string) []interface{} {
	var dataMap map[string]interface{}
	err := jsoniter.Unmarshal(data, &dataMap)
	if err != nil {
		return nil
	}
	d, ok := dataMap[key]
	if !ok {
		return nil
	}

	d2, ok := d.([]interface{})
	if !ok {
		return nil
	}
	return d2
}

func (jsonmgr *jsonMgr) Marshal(v interface{}) ([]byte, error) {
	return jsoniter.Marshal(v)
}

func (jsonmgr *jsonMgr) Unmarshal(data []byte, v interface{}) error {
	return jsoniter.Unmarshal(data, v)
}

func (jsonmgr *jsonMgr) MarshalData(v interface{}) []byte {
	d, err := jsonmgr.Marshal(v)
	if err != nil || d == nil {
		return utils.NullData
	}
	return d
}

