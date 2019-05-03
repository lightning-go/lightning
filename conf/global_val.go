/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package conf

import (
	"sync"
	"time"
)

type GlobalVal struct {
	MaxConnNum    int   //最大连接数
	MaxPacketSize int32 //单个消息包大小最大值
	//MaxMsgSerial      int32  //消息序列号最大值
	//SerialIncrValue   int32  //消息序列号增长值
	//CompressThreshold int32  //消息包压缩阈值,大于则压缩
	MaxQueueSize int32 //消息队列边界值
	//MaxWaitQueueSize  int    //连接等待队列大小
	HttpTimeout      time.Duration
	PongWait         time.Duration
	WriteWait        time.Duration
	RedisIdleTimeout time.Duration
}

func newGlobalVal() *GlobalVal {
	return &GlobalVal{
		MaxConnNum:    3000,
		MaxPacketSize: 1024 * 1024 * 100,
		//MaxMsgSerial:      4096,
		//SerialIncrValue:   2,
		//CompressThreshold: 100,
		MaxQueueSize: 8192,
		//MaxWaitQueueSize:  1000,
		HttpTimeout:      time.Second * 5,
		PongWait:         time.Second * 120,
		WriteWait:        time.Second * 60,
		RedisIdleTimeout: time.Second * 60,
	}
}

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
