/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package defs

import (
	"strconv"
	"sync"
	"reflect"
)

//
type Packet struct {
	sessionId string
	id        string
	data      []byte
	status    int
	sequence  uint64
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

func (p *Packet) SetId(id interface{}) {
	switch id.(type) {
	case string:
		p.id = id.(string)
	case int:
		p.id = strconv.Itoa(id.(int))
	case int32:
		p.id = strconv.Itoa(int(id.(int32)))
	case int64:
		p.id = strconv.FormatInt(id.(int64), 10)
	}
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

func (p *Packet) GetSequence() uint64 {
	return p.sequence
}

func (p *Packet) SetSequence(sequence uint64) {
	p.sequence = sequence
}

//
type MethodType struct {
	sync.Mutex
	Method    reflect.Method //调用方法
	ArgType   reflect.Type   //方法参数类型
	ReplyType reflect.Type   //方法的返回值类型
	numCalls  uint           //被调用次数
}
