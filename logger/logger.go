/**
 * Created: 2019/4/18 0018
 * @author: Jason
 */

package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"fmt"
	"time"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
)

const (
	PANIC = iota
	FATAL
	ERROR
	WARN
	INFO
	DEBUG
	TRACE

	skipNum = 3
)

type Fields = map[string]interface{}

type Logger struct {
	logger  *log.Logger
	skipNum int
}

func NewLogger(lv int, isJsonFormat ...bool) *Logger {
	l := &Logger{
		logger:  log.New(),
		skipNum: skipNum,
	}
	if l.logger == nil {
		return nil
	}

	var format log.Formatter = &log.TextFormatter{TimestampFormat: "2006/01/02 15:04:05"}
	if len(isJsonFormat) > 0 && isJsonFormat[0] {
		format = &log.JSONFormatter{TimestampFormat: "2006/01/02 15:04:05"}
	}
	l.logger.SetFormatter(format)
	l.logger.SetLevel(log.Level(lv))
	l.logger.SetOutput(os.Stdout)
	return l
}

func (l *Logger) SetRotation(maxAge, rotationTime time.Duration, pathFile string) {
	if l.logger == nil {
		return
	}

	if len(pathFile) == 0 {
		return
	}
	w, err := rotatelogs.New(
		pathFile+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(pathFile),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(rotationTime),
	)
	if err != nil {
		return
	}

	ifHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: w,
		log.InfoLevel:  w,
		log.WarnLevel:  w,
		log.ErrorLevel: w,
	}, &log.JSONFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
	})

	l.logger.Hooks.Add(ifHook)
}

func (l *Logger) SetLevel(level int) {
	l.logger.SetLevel(log.Level(level))
}

func (l *Logger) SetSkipNum(val int) {
	l.skipNum = val
}

func (l *Logger) getCaller(skip int) string {
	pc, _, lineno, ok := runtime.Caller(skip)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}
	return src
}

func (l *Logger) logFields(level log.Level, fields map[string]interface{}, args interface{}) {
	src := l.getCaller(l.skipNum)

	if fields != nil && len(fields) > 0 {
		fields["file"] = src
		l.logger.WithFields(fields).Log(level, args)
		return
	}

	l.logger.WithFields(log.Fields{
		"file": src,
	}).Log(level, args)
}

func (l *Logger) logf(level log.Level, format string, args ...interface{}) {
	src := l.getCaller(l.skipNum)

	msg := ""
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	l.logger.WithFields(log.Fields{
		"file": src,
	}).Log(level, msg)
}

func (l *Logger) Panic(args interface{}, fields ...map[string]interface{}) {
	if len(fields) > 0 && fields[0] != nil {
		l.logFields(log.PanicLevel, fields[0], args)
	} else {
		l.logFields(log.PanicLevel, nil, args)
	}
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.logf(log.PanicLevel, format, args...)
}

func (l *Logger) Fatal(args interface{}, fields ...map[string]interface{}) {
	if len(fields) > 0 && fields[0] != nil {
		l.logFields(log.FatalLevel, fields[0], args)
	} else {
		l.logFields(log.PanicLevel, nil, args)
	}
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logf(log.FatalLevel, format, args...)
}

func (l *Logger) Error(args interface{}, fields ...map[string]interface{}) {
	if len(fields) > 0 && fields[0] != nil {
		l.logFields(log.ErrorLevel, fields[0], args)
	} else {
		l.logFields(log.ErrorLevel, nil, args)
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(log.ErrorLevel, format, args...)
}

func (l *Logger) Warn(args interface{}, fields ...map[string]interface{}) {
	if len(fields) > 0 && fields[0] != nil {
		l.logFields(log.ErrorLevel, fields[0], args)
	} else {
		l.logFields(log.ErrorLevel, nil, args)
	}
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf(log.WarnLevel, format, args...)
}

func (l *Logger) Info(args interface{}, fields ...map[string]interface{}) {
	if len(fields) > 0 && fields[0] != nil {
		l.logFields(log.InfoLevel, fields[0], args)
	} else {
		l.logFields(log.InfoLevel, nil, args)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(log.InfoLevel, format, args...)
}

func (l *Logger) Debug(args interface{}, fields ...map[string]interface{}) {
	if len(fields) > 0 && fields[0] != nil {
		l.logFields(log.DebugLevel, fields[0], args)
	} else {
		l.logFields(log.DebugLevel, nil, args)
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(log.DebugLevel, format, args...)
}

func (l *Logger) Trace(args interface{}, fields ...map[string]interface{}) {
	if len(fields) > 0 && fields[0] != nil {
		l.logFields(log.TraceLevel, fields[0], args)
	} else {
		l.logFields(log.TraceLevel, nil, args)
	}
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	l.logf(log.TraceLevel, format, args...)
}
