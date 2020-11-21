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
	"github.com/golang/protobuf/proto"
	"errors"
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
	ParseMethodNameCallback defs.ParseMethodNameCallback
	ParseDataCallback       defs.ParseDataCallback
	serializeDataCallback   defs.SerializeDataCallback
}

func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

func (sf *ServiceFactory) SetParseMethodNameCallback(cb defs.ParseMethodNameCallback) {
	sf.ParseMethodNameCallback = cb
}

func (sf *ServiceFactory) SetParseDataCallback(cb defs.ParseDataCallback) {
	sf.ParseDataCallback = cb
}

func (sf *ServiceFactory) SetSerializeDataCallback(cb defs.SerializeDataCallback) {
	sf.serializeDataCallback = cb
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
		sf.ParseMethodNameCallback = cb[0]
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
	if packet == nil {
		logger.Trace("packet is nil")
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

	data := packet.GetData()
	if data != nil && len(data) > 0 {
		if sf.ParseDataCallback == nil {
			if !ParseDataByJson(data, req.Interface()) {
				logger.Trace("parse request data failed")
				return false
			}
		} else {
			if !sf.ParseDataCallback(data, req.Interface()) {
				logger.Trace("parse request data failed")
				return false
			}
		}
	}

	session.SetPacket(packet)
	defer session.SetPacket(nil)
	function := typ.Method.Func

	if typ.ReplyType == nil {
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
				d := ack.Interface()
				d = d
				if sf.serializeDataCallback == nil {
					data = SerializeDataByJson(ack.Interface())
				} else {
					data = sf.serializeDataCallback(ack.Interface())
				}
				p := &defs.Packet{}
				p.SetSessionId(packet.GetSessionId())
				p.SetId(packet.GetId())
				p.SetStatus(errno)
				p.SetData(data)
				p.SetSequence(packet.GetSequence())
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
		if sf.ParseMethodNameCallback == nil {
			nameVal = methodName
		} else {
			var err error
			nameVal, err = sf.ParseMethodNameCallback(methodName)
			if err != nil {
				continue
			}
		}

		if method.PkgPath != "" {
			continue
		}

		paramNum := mtype.NumIn()
		if paramNum < 3 {
			logger.Debugf("function: %v, param num %v, least 2",
				method.Name, IF(paramNum > 0, paramNum-1, 0))
			continue
		}

		argType := mtype.In(2)
		var replyType reflect.Type = nil

		if paramNum == 4 {
			replyType = mtype.In(3)
			if replyType.Kind() != reflect.Ptr {
				logger.Debug("method reply type not a pointer")
				continue
			}
			if !sf.IsExportedOrBuiltinType(replyType) {
				continue
			}
		}

		returnType := mtype.Out(0)
		if returnType != TypeOfInt {
			logger.Debug("method returns: not int")
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

func RegisterService(rcvr interface{}) {
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

func ParseDataByJson(data []byte, req interface{}) bool {
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

func SerializeDataByJson(d interface{}, param ...interface{}) []byte {
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

func ParseDataByProtobuf(data []byte, req interface{}) bool {
	if data == nil || len(data) == 0 {
		return false
	}
	pb, ok := req.(proto.Message)
	if !ok {
		return false
	}
	err := proto.Unmarshal(data, pb)
	if err != nil {
		logger.Error(err)
		return false
	}
	return true
}

func SerializeDataByProtobuf(d interface{}, param ...interface{}) []byte {
	if d == nil {
		return NullDataEx
	}
	pb, ok := d.(proto.Message)
	if !ok {
		logger.Error(errors.New(fmt.Sprintf("%T does not implement proto.Message", d)))
		return NullDataEx
	}

	buffOK := false
	var buf *proto.Buffer = nil
	if len(param) > 0 {
		b := param[0]
		buf, buffOK = b.(*proto.Buffer)
	}
	if !buffOK {
		buf = proto.NewBuffer(make([]byte, 0, 128))
	}

	err := buf.Marshal(pb)
	if err != nil {
		logger.Error(err)
		return NullDataEx
	}

	data := buf.Bytes()
	buf.Reset()
	return data
}
