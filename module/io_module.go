/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package module

import (
	"github.com/lightning-go/lightning/defs"
	"reflect"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/conf"
	"io"
	"errors"
	"runtime/debug"
	"sync"
	"github.com/lightning-go/lightning/utils"
)

var (
	ErrConnClosed    = errors.New("conn closed")
	ErrReadBuffNil   = errors.New("read buff is nil")
	ErrCodecWriteNil = errors.New("codec write is nil")
	ErrCodecReadNil  = errors.New("codec read is nil")
)

type RpcCall struct {
	request  defs.IPacket
	response defs.IPacket
	reply    interface{}
	Done     chan *RpcCall
}

func (rpcCall *RpcCall) done() {
	select {
	case rpcCall.Done <- rpcCall:
	default:
		logger.Warn("RpcClient: discarding call reply due to insufficient done chan capacity")
	}
}

type IOModule struct {
	conn       defs.IConnection
	codec      defs.ICodec
	writeQueue chan defs.IPacket
	readClose  chan bool
	rpcPool    sync.Pool
	idGen      *utils.IdGenerator
	pending    sync.Map
}

func NewIOModule(conn defs.IConnection) *IOModule {
	if conn == nil {
		return nil
	}
	m := &IOModule{
		conn:       conn,
		codec:      nil,
		writeQueue: make(chan defs.IPacket, conf.GetGlobalVal().MaxQueueSize),
		readClose:  make(chan bool),
		idGen:      utils.NewIdGenerator(),
	}
	m.rpcPool.New = func() interface{} {
		return &RpcCall{}
	}
	return m
}

func (ioModule *IOModule) newRpcCall() *RpcCall {
	return ioModule.rpcPool.Get().(*RpcCall)
}

func (ioModule *IOModule) freeRpcCall(rpcCall *RpcCall) {
	ioModule.rpcPool.Put(rpcCall)
}

func (ioModule *IOModule) UpdateCodec(codec defs.ICodec) {
	ioModule.codec = ioModule.newCodec(codec)
	if ioModule.codec == nil {
		logger.Error("new codec failed")
		return
	}
	if !ioModule.codec.Init(ioModule.conn) {
		logger.Error("codec init failed")
		return
	}
}

func (ioModule *IOModule) newCodec(codec defs.ICodec) defs.ICodec {
	mType := reflect.TypeOf(codec)
	obj := reflect.New(mType.Elem())
	v, ok := obj.Interface().(defs.ICodec)
	if !ok {
		return nil
	}
	return v
}

func (ioModule *IOModule) Close() {
	if ioModule.conn != nil {
		ioModule.conn.Close()
	}
}

func (ioModule *IOModule) OnConnectionLost() {
	if ioModule.writeQueue != nil {
		close(ioModule.writeQueue)
	}
	if ioModule.readClose != nil {
		close(ioModule.readClose)
	}
}

func (ioModule *IOModule) Codec(codec defs.ICodec) bool {
	if codec == nil {
		switch ioModule.conn.(type) {
		case defs.ITcpConnection:
			ioModule.codec = NewStreamCodec()
		case defs.IWSConnection:
			ioModule.codec = NewWSCodec()
		}
	} else {
		ioModule.codec = ioModule.newCodec(codec)
	}

	if ioModule.codec == nil {
		logger.Error("new codec failed")
		return false
	}

	if !ioModule.codec.Init(ioModule.conn) {
		logger.Error("codec init failed")
		return false
	}

	ioModule.enableRead()
	ioModule.enableWrite()
	return true
}

func (ioModule *IOModule) readPending(packet defs.IPacket) bool {
	if packet == nil {
		return false
	}
	seq := packet.GetSequence()
	iCall, ok := ioModule.pending.Load(seq)
	if !ok {
		return false
	}
	ioModule.pending.Delete(seq)

	if iCall == nil {
		logger.Errorf("seq:%v call interface nil", seq)
		return false
	}
	call, ok := iCall.(*RpcCall)
	if !ok {
		logger.Errorf("seq:%v call object nil", seq)
		return false
	}

	switch {
	default:
		call.response = packet
		call.done()
	}
	return true
}

func (ioModule *IOModule) pendDone() {
	ioModule.pending.Range(func(key, value interface{}) bool {
		call, ok := value.(*RpcCall)
		if ok && call != nil {
			call.done()
		}
		return true
	})
}

func (ioModule *IOModule) WriteAwait(packet defs.IPacket) (response defs.IPacket, err error) {
	seq := ioModule.idGen.Get()
	packet.SetSequence(seq)

	call := ioModule.newRpcCall()
	call.request = packet
	call.response = nil
	call.Done = make(chan *RpcCall, 1)

	ioModule.pending.Store(seq, call)
	err = ioModule.writeHandle(packet)
	if err != nil {
		logger.Error(err)
		iCall, ok := ioModule.pending.Load(seq)
		if ok {
			ioModule.pending.Delete(seq)
		}
		if iCall != nil {
			call := iCall.(*RpcCall)
			if call != nil {
				call.done()
			}
		}
	}

	call = <-call.Done
	if call != nil {
		response = call.response
	}
	ioModule.freeRpcCall(call)
	return
}

func (ioModule *IOModule) Write(packet defs.IPacket) {
	if packet == nil {
		return
	}
	if ioModule.conn.IsClosed() {
		return
	}
	select {
	case ioModule.writeQueue <- packet:
	}
}

func (ioModule *IOModule) enableWrite() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)
				trackBack := string(debug.Stack())
				logger.Error(trackBack)
			}
		}()

		for packet := range ioModule.writeQueue {
			if packet == nil {
				continue
			}
			err := ioModule.writeHandle(packet)
			if err != nil {
				logger.Error(err)
				break
			}
			if len(ioModule.writeQueue) == 0 {
				ioModule.conn.WriteComplete()
			}
		}

	}()
}

func (ioModule *IOModule) enableRead() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)
				trackBack := string(debug.Stack())
				logger.Error(trackBack)
			}
		}()

	QUIT:
		for {
			select {
			case <-ioModule.readClose:
				break QUIT
			default:
				packet, err := ioModule.readHandle()
				if err != nil {
					if err != io.EOF && err != ErrConnClosed {
						logger.Error(err)
					}
					break QUIT
				}
				if packet == nil {
					continue
				}
				if ioModule.readPending(packet) {
					continue
				}
				ioModule.conn.ReadPacket(packet)
			}
		}
		ioModule.pendDone()
		ioModule.Close()
	}()
}

func (ioModule *IOModule) writeHandle(packet defs.IPacket) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			trackBack := string(debug.Stack())
			logger.Errorf("%v", trackBack)
		}
	}()
	return ioModule.codec.Write(packet)
}

func (ioModule *IOModule) readHandle() (defs.IPacket, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			trackBack := string(debug.Stack())
			logger.Errorf("%v", trackBack)
		}
	}()
	return ioModule.codec.Read()
}
