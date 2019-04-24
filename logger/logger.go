/**
 * Created: 2019/4/18 0018
 * @author: Jason
 */

package logger

import (
	"runtime"
	"fmt"
	"strings"
	l4g "github.com/alecthomas/log4go"
)

const (
	FINEST   = iota
	FINE
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	CRITICAL

	skipNum = 2
)

var defaultLogger l4g.Logger

func InitLog(logLv ...int) {
	if defaultLogger != nil {
		return
	}

	lv := l4g.TRACE
	if len(logLv) > 0 && logLv[0] >= 0 {
		lv = l4g.Level(logLv[0])
	}

	defaultLogger = l4g.NewDefaultLogger(lv)
}

func InitLogByConfig(path string) {
	if defaultLogger == nil {
		defaultLogger = l4g.NewDefaultLogger(l4g.TRACE)
	}
	defaultLogger.LoadConfiguration(path)
}

func Exit() {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Close()
}

func Logf(handler *l4g.Logger, level l4g.Level, skip int, arg0 interface{}, args ...interface{}) {
	pc, _, lineno, ok := runtime.Caller(skip)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}

	var msg string
	switch arg0.(type) {
	case string:
		msg = arg0.(string)
	default:
		msg = fmt.Sprint(arg0) + strings.Repeat(" %v", len(args))
	}

	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	handler.Log(level, src, msg)
}

func Debug(arg0 interface{}, args ...interface{}) {
	InitLog()
	Logf(&defaultLogger, l4g.DEBUG, skipNum, arg0, args...)
}

func Trace(arg0 interface{}, args ...interface{}) {
	InitLog()
	Logf(&defaultLogger, l4g.TRACE, skipNum, arg0, args...)
}

func Info(arg0 interface{}, args ...interface{}) {
	InitLog()
	Logf(&defaultLogger, l4g.INFO, skipNum, arg0, args...)
}

func Warn(arg0 interface{}, args ...interface{}) {
	InitLog()
	Logf(&defaultLogger, l4g.WARNING, skipNum, arg0, args...)
}

func Error(arg0 interface{}, args ...interface{}) {
	InitLog()
	Logf(&defaultLogger, l4g.ERROR, skipNum, arg0, args...)
}

func Cri(arg0 interface{}, args ...interface{}) {
	InitLog()
	Logf(&defaultLogger, l4g.CRITICAL, skipNum, arg0, args...)
}
