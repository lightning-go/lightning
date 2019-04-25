/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package defs

import (
	"sync"
	"reflect"
)

//
type Packet struct {
	sessionId string
	id        string
	data      []byte
	status    int
}

func (p *Packet) GetSessionId() string {
	return p.sessionId
}

func (p *Packet) SetSessionId(sessionId string) {
	p.sessionId = sessionId
}

func (p *Packet) GetId() string {
	return p.id
}

func (p *Packet) SetId(id string) {
	p.id = id
}

func (p *Packet) GetData() []byte {
	return p.data
}

func (p *Packet) SetData(data []byte) {
	p.data = data
}

func (p *Packet) GetStatus() int {
	return p.status
}

func (p *Packet) SetStatus(status int) {
	p.status = status
}

//
type MethodType struct {
	sync.Mutex
	Method    reflect.Method //调用方法
	ArgType   reflect.Type   //方法参数类型
	ReplyType reflect.Type   //方法的返回值类型
	numCalls  uint           //被调用次数
}
