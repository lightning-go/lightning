/**
 * Created: 2019/4/19 0019
 * @author: Jason
 */

package network

import (
	"net"
	"github.com/lightning-go/lightning/defs"
	"github.com/satori/go.uuid"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/module"
	"sync/atomic"
	"github.com/lightning-go/lightning/utils"
	"context"
)

type Connection struct {
	connId        string
	conn          net.Conn
	codec         defs.ICodec
	ioModule      defs.IIOModule
	connCallback  defs.ConnCallback
	msgCallback   defs.MsgCallback
	closeCallback defs.CloseCallback
	writeComplete defs.WriteCompleteCallback
	authCallback  defs.AuthorizedCallback
	isClosed      int32
	isAuthorized  bool
	ctx           context.Context
}

func NewConnection(conn net.Conn) *Connection {
	id, err := uuid.NewV4()
	if err != nil {
		logger.Error(err)
		return nil
	}

	c := &Connection{
		connId:       id.String(),
		conn:         conn,
		isAuthorized: false,
		ctx:          utils.NewContextMap(context.Background()),
	}
	return c
}

func (c *Connection) SetCodec(codec defs.ICodec) {
	c.codec = codec
}

func (c *Connection) SetIOModule(ioModule defs.IIOModule) {
	c.ioModule = ioModule
}

func (c *Connection) SetConnCallback(cb defs.ConnCallback) {
	c.connCallback = cb
}

func (c *Connection) SetCloseCallback(cb defs.CloseCallback) {
	c.closeCallback = cb
}

func (c *Connection) SetMsgCallback(cb defs.MsgCallback) {
	c.msgCallback = cb
}

func (c *Connection) SetAuthorizedCallback(cb defs.AuthorizedCallback) {
	c.authCallback = cb
}

func (c *Connection) SetContext(key, value interface{}) {
	utils.SetMapContext(c.ctx, key, value)
}

func (c *Connection) GetContext(key interface{}) interface{} {
	return utils.GetMapContext(c.ctx, key)
}

func (c *Connection) DelContext(key interface{}) {
	utils.DelMapContext(c.ctx, key)
}

func (c *Connection) GetId() string {
	return c.connId
}

func (c *Connection) GetConn() net.Conn {
	return c.conn
}

func (c *Connection) LocalAddr() string {
	return c.conn.LocalAddr().String()
}

func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *Connection) IsClosed() bool {
	isClosed := atomic.LoadInt32(&c.isClosed)
	if isClosed > 0 {
		return true
	}
	return false
}

func (c *Connection) OnConnection() {
	if c.connCallback != nil {
		c.connCallback(c)
	}
}

func (c *Connection) Start() bool {
	if c.ioModule == nil {
		c.ioModule = module.NewIOModule(c)
		if c.ioModule == nil {
			logger.Error("connection create io module failed")
			return false
		}
	}
	ok := c.ioModule.Codec(c.codec)
	if !ok {
		logger.Error("io module codec error")
		return false
	}
	c.OnConnection()
	return true
}

func (c *Connection) Close() bool {
	isClosed := atomic.LoadInt32(&c.isClosed)
	if isClosed > 0 {
		logger.Trace("connection was closed")
		return false
	}
	atomic.AddInt32(&c.isClosed, 1)

	if c.closeCallback != nil {
		c.closeCallback(c)
	}

	if c.ioModule != nil {
		c.ioModule.OnConnectionLost()
	}

	err := c.conn.Close()
	if err != nil {
		logger.Warn(err)
	}
	return true
}

func (c *Connection) WriteComplete() {
	if c.writeComplete != nil {
		c.writeComplete(c)
	}
}

func (c *Connection) WriteData(data []byte) {
	if data == nil || len(data) == 0 {
		return
	}
	if c.IsClosed() {
		return
	}
	if c.ioModule == nil {
		return
	}
	p := &defs.Packet{}
	p.SetData(data)
	c.ioModule.Write(p)
}

func (c *Connection) WritePacket(packet defs.IPacket) {
	if c.IsClosed() {
		return
	}
	if c.ioModule == nil {
		return
	}
	c.ioModule.Write(packet)
}

func (c *Connection) ReadPacket(packet defs.IPacket) {
	c.onMsg(packet)
}

func (c *Connection) onMsg(packet defs.IPacket) {
	if c.authCallback != nil && !c.isAuthorized {
		c.isAuthorized = c.authCallback(c, packet)
		return
	}

	if c.msgCallback != nil {
		c.msgCallback(c, packet)
	}
}
