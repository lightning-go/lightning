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
)

var (
	ErrConnClosed    = errors.New("conn closed")
	ErrReadBuffNil   = errors.New("read buff is nil")
	ErrCodecWriteNil = errors.New("codec write is nil")
	ErrCodecReadNil  = errors.New("codec read is nil")
)

type IOModule struct {
	conn       defs.IConnection
	codec      defs.ICodec
	writeQueue chan defs.IPacket
	readClose  chan bool
}

func NewIOModule(conn defs.IConnection) *IOModule {
	if conn == nil {
		return nil
	}
	return &IOModule{
		conn:       conn,
		codec:      nil,
		writeQueue: make(chan defs.IPacket, conf.GetGlobalVal().MaxQueueSize),
		readClose:  make(chan bool),
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
				err := ioModule.readHandle()
				if err != nil {
					if err != io.EOF && err != ErrConnClosed {
						logger.Error(err)
					}
					break QUIT
				}
			}
		}
		ioModule.Close()
	}()
}

func (ioModule *IOModule) writeHandle(packet defs.IPacket) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	return ioModule.codec.Write(packet)
}

func (ioModule *IOModule) readHandle() error {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()

	packet, err := ioModule.codec.Read()
	if err != nil {
		return err
	}

	ioModule.conn.ReadPacket(packet)
	return nil
}
