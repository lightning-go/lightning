/**
 * Created: 2019/4/19 0019
 * @author: Jason
 */

package network

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/utils"
	uuid "github.com/satori/go.uuid"
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
	id := uuid.NewV4()
	//if id == nil {
	//	logger.Error("uuid nil")
	//	return nil
	//}

	c := &Connection{
		connId:       id.String(),
		conn:         conn,
		isAuthorized: false,
		ctx:          utils.NewContextMap(context.Background()),
	}
	return c
}

func (c *Connection) UpdateCodec(codec defs.ICodec) {
	c.ioModule.UpdateCodec(codec)
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

func (c *Connection) SetWriteCompleteCallback(cb defs.WriteCompleteCallback) {
	c.writeComplete = cb
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

func (c *Connection) write(packet defs.IPacket, await bool) (defs.IPacket, error) {
	if c.IsClosed() {
		return nil, nil
	}
	if c.ioModule == nil {
		return nil, nil
	}
	if await {
		return c.ioModule.WriteAwait(packet)
	}
	c.ioModule.Write(packet)
	return nil, nil
}

func (c *Connection) writeData(id string, data []byte, await bool) (defs.IPacket, error) {
	if data == nil || len(data) == 0 {
		return nil, nil
	}
	p := &defs.Packet{}
	p.SetId(id)
	p.SetData(data)
	return c.write(p, await)
}

func (c *Connection) WriteData(data []byte) {
	c.writeData("", data, false)
}

func (c *Connection) WriteDataById(id string, data []byte) {
	c.writeData(id, data, false)
}

func (c *Connection) WritePacket(packet defs.IPacket) {
	c.write(packet, false)
}

func (c *Connection) WriteDataAwait(data []byte) (defs.IPacket, error) {
	return c.writeData("", data, true)
}

func (c *Connection) WriteDataByIdAwait(id string, data []byte) (defs.IPacket, error) {
	return c.writeData(id, data, true)
}

func (c *Connection) WritePacketAwait(packet defs.IPacket) (defs.IPacket, error) {
	return c.write(packet, true)
}

func (c *Connection) ReadPacket(packet defs.IPacket) {
	if packet == nil {
		return
	}
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
