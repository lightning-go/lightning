/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package conf

import (
	"time"
	"sync"
)

var globalVal *GlobalVal
var globalOnce sync.Once

func init() {
	globalOnce.Do(func() {
		globalVal = newGlobalVal()
	})
}

func GetGlobalVal() *GlobalVal {
	return globalVal
}

type GlobalVal struct {
	MaxConnNum       int
	MaxPacketSize    int32
	MaxQueueSize     int32
	HttpTimeout      time.Duration
	PongWait         time.Duration
	WriteWait        time.Duration
	RedisIdleTimeout time.Duration
}

func newGlobalVal() *GlobalVal {
	return &GlobalVal{
		MaxConnNum:       3000,
		MaxPacketSize:    1024 * 1024 * 100,
		MaxQueueSize:     8192,
		HttpTimeout:      time.Second * 5,
		PongWait:         time.Second * 120,
		WriteWait:        time.Second * 60,
		RedisIdleTimeout: time.Second * 60,
	}
}
