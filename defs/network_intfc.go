/**
 * Created: 2019/4/18 0018
 * @author: Jason
 */

package defs

import (
	"net"
	"github.com/gorilla/websocket"
)

type ExitCallback func()
type CloseCallback func(IConnection)
type WriteCompleteCallback func(IConnection)
type ConnCallback func(IConnection)
type MsgCallback func(IConnection, IPacket)
type AuthorizedCallback func(IConnection, IPacket) bool
type ClientConnCallback func(net.Conn)

type IServer interface {
	Name() string
	Serve()
	Stop()
	SetCodec(ICodec)
	SetIOModule(IIOModule)
	SetConnCallback(ConnCallback)
	SetMsgCallback(MsgCallback)
	SetExitCallback(ExitCallback)
	SetAuthorizedCallback(AuthorizedCallback)
}

type IClient interface {
	Name() string
	Connect() IConnection
	Close()
	SetRetry(bool)
	SetCodec(ICodec)
	SetIOModule(IIOModule)
	SetConnCallback(ConnCallback)
	SetMsgCallback(MsgCallback)
	SendData([]byte)
	SendPacket(IPacket)
}

type IConnection interface {
	GetId() string
	LocalAddr() string
	RemoteAddr() string
	Close() bool
	IsClosed() bool
	OnConnection()
	ReadPacket(IPacket)
	WritePacket(IPacket)
	WriteData([]byte)
	WriteComplete()
	SetContext(interface{}, interface{})
	GetContext(interface{}) interface{}
	DelContext(interface{})
}

type ITcpConnection interface {
	GetConn() net.Conn
}

type IWSConnection interface {
	GetConn() *websocket.Conn
}
