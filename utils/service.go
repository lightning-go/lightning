/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package utils

import (
	"strings"
	"fmt"
	"sync"
	"reflect"
	"unicode/utf8"
	"unicode"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"

	"github.com/json-iterator/go"
	"runtime/debug"
)

var theServiceFactory *ServiceFactory
var theServiceOnce sync.Once

func GetMsgFactory() *ServiceFactory {
	theServiceOnce.Do(func() {
		theServiceFactory = NewServiceFactory()
	})
	return theServiceFactory
}

type ServiceFactory struct {
	msgRCVR                 reflect.Value
	msgHandle               sync.Map
	parseMethodNameCallback defs.ParseMethodNameCallback
	parseReqDataCallback    defs.ParseReqDataCallback
	parseAckDataCallback    defs.ParseAckDataCallback
	autoReply               bool
}

func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{
		autoReply: true,
	}
}

func (sf *ServiceFactory) SetAutoReply(val bool) {
	sf.autoReply = val
}

func (sf *ServiceFactory) SetParseMethodNameCallback(cb defs.ParseMethodNameCallback) {
	sf.parseMethodNameCallback = cb
}

func (sf *ServiceFactory) SetParseReqDataCallback(cb defs.ParseReqDataCallback) {
	sf.parseReqDataCallback = cb
}

func (sf *ServiceFactory) SetParseAckDataCallback(cb defs.ParseAckDataCallback) {
	sf.parseAckDataCallback = cb
}

func (sf *ServiceFactory) get(key interface{}) *defs.MethodType {
	cb, ok := sf.msgHandle.Load(key)
	if ok {
		return cb.(*defs.MethodType)
	}
	return nil
}

func (sf *ServiceFactory) Register(rcvr interface{}, cb ...defs.ParseMethodNameCallback) {
	if len(cb) > 0 {
		sf.parseMethodNameCallback = cb[0]
	}
	sf.suitableMethods(rcvr, &sf.msgRCVR, &sf.msgHandle)
}

func (sf *ServiceFactory) OnServiceHandle(session defs.ISession, packet defs.IPacket) bool {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			trackBack := string(debug.Stack())
			logger.Errorf("%v", trackBack)
		}
	}()

	if session == nil {
		logger.Trace("session is nil")
		return false
	}

	key := packet.GetId()
	typ := sf.get(key)
	if typ == nil {
		logger.Trace("callback for service is nil ", logger.Fields{"type": key})
		return false
	}

	//
	if typ.ArgType.Kind() != reflect.Ptr {
		logger.Error("req type error")
		return false
	}
	req := reflect.New(typ.ArgType.Elem())

	if sf.parseReqDataCallback == nil {
		if !ParseReqDataByJson(packet.GetData(), req.Interface()) {
			logger.Trace("parse request data failed")
			return false
		}
	} else {
		if !sf.parseReqDataCallback(packet.GetData(), req.Interface()) {
			logger.Trace("parse request data failed")
			return false
		}
	}

	function := typ.Method.Func
	if !sf.autoReply {
		function.Call([]reflect.Value{sf.msgRCVR, reflect.ValueOf(session), req})
		return true
	}

	ack := reflect.New(typ.ReplyType.Elem())
	result := function.Call([]reflect.Value{sf.msgRCVR, reflect.ValueOf(session), req, ack})
	if result != nil && len(result) > 0 {
		iErrno := result[0].Interface()
		switch iErrno.(type) {
		case int:
			errno, ok := iErrno.(int)
			if ok {
				var data []byte
				if sf.parseAckDataCallback == nil {
					data = ParseAckDataByJson(ack.Interface())
				} else {
					data = sf.parseAckDataCallback(ack.Interface())
				}
				p := &defs.Packet{}
				p.SetSessionId(packet.GetSessionId())
				p.SetId(packet.GetId())
				p.SetStatus(errno)
				p.SetData(data)
				session.WritePacket(p)
			}
		default:
			logger.Warn("unexpected type")
		}
	}
	return true
}

func (sf *ServiceFactory) IsExported(name string) bool {
	runeName, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(runeName)
}

func (sf *ServiceFactory) IsExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return sf.IsExported(t.Name()) || t.PkgPath() == ""
}

func (sf *ServiceFactory) suitableMethods(rcvr interface{}, rcvr2 *reflect.Value, serviceMap *sync.Map) error {
	*rcvr2 = reflect.ValueOf(rcvr)
	typ := reflect.TypeOf(rcvr)

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		mtype := method.Type
		methodName := method.Name

		var nameVal string
		if sf.parseMethodNameCallback == nil {
			nameVal = methodName
		} else {
			var err error
			nameVal, err = sf.parseMethodNameCallback(methodName)
			if err != nil {
				continue
			}
		}

		if method.PkgPath != "" {
			continue
		}

		paramNum := mtype.NumIn()
		if paramNum != 4 && sf.autoReply {
			continue
		} else if paramNum != 3 && !sf.autoReply {
			continue
		}

		argType := mtype.In(2)
		var replyType reflect.Type

		if paramNum == 4 {
			replyType = mtype.In(3)
			if replyType.Kind() != reflect.Ptr {
				logger.Error("method reply type not a pointer")
				continue
			}
			if !sf.IsExportedOrBuiltinType(replyType) {
				continue
			}
		}

		returnType := mtype.Out(0)
		if returnType != TypeOfInt {
			logger.Error("method returns: not int")
			continue
		}

		serviceMap.Store(
			nameVal,
			&defs.MethodType{
				Method:    method,
				ArgType:   argType,
				ReplyType: replyType,
			})
	}

	return nil
}

func RegisterService(rcvr interface{}, autoReply ...bool) {
	if len(autoReply) > 0 {
		GetMsgFactory().SetAutoReply(autoReply[0])
	}
	GetMsgFactory().Register(rcvr)
}

func OnServiceHandle(session defs.ISession, packet defs.IPacket) bool {
	return GetMsgFactory().OnServiceHandle(session, packet)
}

func ParseMethodNameBySuffix(methodName string) (name string, err error) {
	delimiter := strings.Index(methodName, "_")
	if delimiter < 0 {
		err = fmt.Errorf("method name ill-formed: %v", methodName)
		return "", err
	}
	typeName := methodName[delimiter+1:]
	if len(typeName) > 0 {
		return typeName, nil
	}
	return "", fmt.Errorf("type name: %v is non-existent", typeName)
}

func ParseReqDataByJson(data []byte, req interface{}) bool {
	if data == nil || len(data) == 0 {
		return false
	}
	err := jsoniter.Unmarshal(data, req)
	if err != nil {
		logger.Error(err)
		return false
	}
	return true
}

func ParseAckDataByJson(d interface{}) []byte {
	if d == nil {
		return NullData
	}
	data, err := jsoniter.Marshal(d)
	if err != nil {
		logger.Error(err)
		return NullData
	}
	return data
}
