/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package network

import (
	"context"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/utils"
	uuid "github.com/satori/go.uuid"
)

type WSConnection struct {
	connId        string
	conn          *websocket.Conn
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
	msgType       int
}

func NewWSConnection(conn *websocket.Conn) *WSConnection {
	id := uuid.NewV4()
	// if err != nil {
	// 	logger.Error(err)
	// 	return nil
	// }

	wsc := &WSConnection{
		connId:       id.String(),
		conn:         conn,
		isAuthorized: false,
		msgType:      websocket.TextMessage,
		ctx:          utils.NewContextMap(context.Background()),
	}
	return wsc
}

func (wsc *WSConnection) UpdateCodec(codec defs.ICodec) {
	wsc.ioModule.UpdateCodec(codec)
}

func (wsc *WSConnection) SetAuthorized(val bool) {
	wsc.isAuthorized = val
}

func (wsc *WSConnection) SetMsgType(msgType int) {
	wsc.msgType = msgType
}

func (wsc *WSConnection) GetMsgType() int {
	return wsc.msgType
}

func (wsc *WSConnection) SetCodec(codec defs.ICodec) {
	wsc.codec = codec
}

func (wsc *WSConnection) SetIOModule(ioModule defs.IIOModule) {
	wsc.ioModule = ioModule
}

func (wsc *WSConnection) SetCloseCallback(cb defs.CloseCallback) {
	wsc.closeCallback = cb
}

func (wsc *WSConnection) SetConnCallback(cb defs.ConnCallback) {
	wsc.connCallback = cb
}

func (wsc *WSConnection) SetMsgCallback(cb defs.MsgCallback) {
	wsc.msgCallback = cb
}

func (wsc *WSConnection) SetAuthorizedCallback(cb defs.AuthorizedCallback) {
	wsc.authCallback = cb
}

func (wsc *WSConnection) SetWriteCompleteCallback(cb defs.WriteCompleteCallback) {
	wsc.writeComplete = cb
}

func (wsc *WSConnection) SetContext(key, value interface{}) {
	utils.SetMapContext(wsc.ctx, key, value)
}

func (wsc *WSConnection) GetContext(key interface{}) interface{} {
	return utils.GetMapContext(wsc.ctx, key)
}

func (wsc *WSConnection) DelContext(key interface{}) {
	utils.DelMapContext(wsc.ctx, key)
}

func (wsc *WSConnection) GetId() string {
	return wsc.connId
}

func (wsc *WSConnection) GetConn() *websocket.Conn {
	return wsc.conn
}

func (wsc *WSConnection) LocalAddr() string {
	return wsc.conn.LocalAddr().String()
}

func (wsc *WSConnection) RemoteAddr() string {
	return wsc.conn.RemoteAddr().String()
}

func (wsc *WSConnection) IsClosed() bool {
	isClosed := atomic.LoadInt32(&wsc.isClosed)
	if isClosed > 0 {
		return true
	}
	return false
}

func (wsc *WSConnection) OnConnection() {
	if wsc.connCallback != nil {
		wsc.connCallback(wsc)
	}
}

func (wsc *WSConnection) Start() bool {
	if wsc.ioModule == nil {
		wsc.ioModule = module.NewIOModule(wsc)
		if wsc.ioModule == nil {
			logger.Error("connection create io module failed")
			return false
		}
	}
	ok := wsc.ioModule.Codec(wsc.codec)
	if !ok {
		logger.Error("io module codec error")
		return false
	}
	wsc.OnConnection()
	return true
}

func (wsc *WSConnection) Close() bool {
	isClosed := atomic.LoadInt32(&wsc.isClosed)
	if isClosed > 0 {
		logger.Warn("connection was closed")
		return false
	}
	atomic.AddInt32(&wsc.isClosed, 1)

	if wsc.closeCallback != nil {
		wsc.closeCallback(wsc)
	}

	if wsc.ioModule != nil {
		wsc.ioModule.OnConnectionLost()
	}

	//
	err := wsc.conn.Close()
	if err != nil {
		logger.Warn(err)
	}
	return true
}

func (wsc *WSConnection) WriteComplete() {
	if wsc.writeComplete != nil {
		wsc.writeComplete(wsc)
	}
}

func (wsc *WSConnection) write(packet defs.IPacket, await bool) (defs.IPacket, error) {
	if wsc.IsClosed() {
		return nil, nil
	}
	if wsc.ioModule == nil {
		return nil, nil
	}
	if await {
		return wsc.ioModule.WriteAwait(packet)
	}
	wsc.ioModule.Write(packet)
	return nil, nil
}

func (wsc *WSConnection) writeData(id string, data []byte, await bool) (defs.IPacket, error) {
	if data == nil || len(data) == 0 {
		return nil, nil
	}
	p := &defs.Packet{}
	p.SetId(id)
	p.SetData(data)
	return wsc.write(p, await)
}

func (wsc *WSConnection) WriteData(data []byte) {
	wsc.writeData("", data, false)
}

func (wsc *WSConnection) WriteDataById(id string, data []byte) {
	wsc.writeData(id, data, false)
}

func (wsc *WSConnection) WritePacket(packet defs.IPacket) {
	wsc.write(packet, false)
}

func (wsc *WSConnection) WriteDataAwait(data []byte) (defs.IPacket, error) {
	return wsc.writeData("", data, true)
}

func (wsc *WSConnection) WriteDataByIdAwait(id string, data []byte) (defs.IPacket, error) {
	return wsc.writeData(id, data, true)
}

func (wsc *WSConnection) WritePacketAwait(packet defs.IPacket) (defs.IPacket, error) {
	return wsc.write(packet, true)
}

func (wsc *WSConnection) ReadPacket(packet defs.IPacket) {
	wsc.onMsg(packet)
}

func (wsc *WSConnection) onMsg(packet defs.IPacket) {
	if wsc.authCallback != nil && !wsc.isAuthorized {
		wsc.isAuthorized = wsc.authCallback(wsc, packet)
		return
	}
	if wsc.msgCallback != nil {
		wsc.msgCallback(wsc, packet)
	}
}
