/**
 * Created: 2019/4/19 0019
 * @author: Jason
 */

package utils

import (
	"os"
	"os/signal"
	"github.com/lightning-go/lightning/logger"
	"reflect"
	"unsafe"
	"time"
	"errors"
	"context"
	"sync"
	"syscall"
)

var (
	NullData      = []byte("{}")
	NullDataEx    = []byte("")
	TypeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()
	TypeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	TypeOfInt     = reflect.TypeOf((*int)(nil)).Elem()
)

func IF(b bool, t1, t2 interface{}) interface{} {
	if b {
		return t1
	}
	return t2
}

func NowTimeFormat() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func WaitExit(fList ...func()) {
	WaitSignal()
	if len(fList) > 0 {
		for _, f := range fList {
			if f != nil {
				f()
			}
		}
	}
}

func WaitSignal() {
	c := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGPIPE)
	//signal.Notify(c, os.Interrupt, os.Kill)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-c
	logger.Warn("recv signal", logger.Fields{"signal": sig})
}

func String(b []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}

func Slice(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}

var SliceIsNilError = errors.New("slice is nil")

func DeleteSlice(slice interface{}, index int) (interface{}, error) {
	sliceValue := reflect.ValueOf(slice)
	length := sliceValue.Len()
	if slice == nil || length == 0 || (length-1) < index {
		return nil, SliceIsNilError
	}
	if length-1 == index {
		return sliceValue.Slice(0, index).Interface(), nil
	} else if (length - 1) >= index {
		return reflect.AppendSlice(sliceValue.Slice(0, index), sliceValue.Slice(index+1, length)).Interface(), nil
	}
	return nil, errors.New("error")
}

const ctxKey = "ctxKey"

func NewContextMap(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey, &sync.Map{})
}

func GetContextMap(ctx context.Context) (m *sync.Map, ok bool) {
	m, ok = ctx.Value(ctxKey).(*sync.Map)
	return
}

func SetMapContext(ctx context.Context, key, value interface{}) {
	contextMap, ok := GetContextMap(ctx)
	if ok {
		(*contextMap).Store(key, value)
	}
}

func GetMapContext(ctx context.Context, key interface{}) interface{} {
	contextMap, ok := GetContextMap(ctx)
	if ok {
		val, ok := (contextMap).Load(key)
		if ok {
			return val
		}
	}
	return nil
}

func DelMapContext(ctx context.Context, key interface{}) {
	contextMap, ok := GetContextMap(ctx)
	if ok {
		_, ok := (contextMap).Load(key)
		if ok {
			contextMap.Delete(key)
		}
	}
}
