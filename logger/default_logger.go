/**
 * @author: Jason
 * Created: 19-5-3
 */

package logger

import (
	"fmt"
	"github.com/lightning-go/lightning/conf"
	"sync"
	"time"
)

var defaultLogger *Logger
var defaultLoggerOnce sync.Once

func init() {
	defaultLoggerOnce.Do(func() {
		defaultLogger = NewLogger(TRACE)
		defaultLogger.SetSkipNum(4)
	})
}

func InitLog(level, maxAge, rotationTime int, pathFile string) {
	if defaultLogger == nil {
		defaultLogger = NewLogger(level)
	}
	defaultLogger.SetLevel(level)

	defaultLogger.SetRotation(time.Minute*time.Duration(maxAge),
		time.Minute*time.Duration(rotationTime), pathFile)
}

func GetLevel(lvKey string) (lv int) {
	switch lvKey {
	case "panic":
		lv = PANIC
	case "fatal":
		lv = FATAL
	case "error":
		lv = ERROR
	case "warn":
		lv = WARN
	case "info":
		lv = INFO
	case "debug":
		lv = DEBUG
	case "trace":
		lv = TRACE
	default:
		lv = TRACE
	}
	return
}

func SetLevel(level int) {
	defaultLogger.SetLevel(level)
}

func GetLogPathFile(logConf *conf.LogConfig) string {
	if logConf == nil {
		return ""
	}
	return fmt.Sprintf("%v/%v", logConf.LogPath, logConf.LogFile)
}

func Panic(args interface{}, fields ...map[string]interface{}) {
	defaultLogger.Panic(args, fields...)
}

func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}

func Fatal(args interface{}, fields ...map[string]interface{}) {
	defaultLogger.Fatal(args, fields...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func Error(args interface{}, fields ...map[string]interface{}) {
	defaultLogger.Error(args, fields...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Warn(args interface{}, fields ...map[string]interface{}) {
	defaultLogger.Warn(args, fields...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Info(args interface{}, fields ...map[string]interface{}) {
	defaultLogger.Info(args, fields...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Debug(args interface{}, fields ...map[string]interface{}) {
	defaultLogger.Debug(args, fields...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Trace(args interface{}, fields ...map[string]interface{}) {
	defaultLogger.Trace(args, fields...)
}

func Tracef(format string, args ...interface{}) {
	defaultLogger.Tracef(format, args...)
}
