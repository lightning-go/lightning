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
type NewIOModuleCallback func(IConnection) IIOModule
type ParseMethodNameCallback func(string) (string, error)
type ParseDataCallback func([]byte, interface{}) bool
type SerializeDataCallback func(interface{}, ...interface{}) []byte

type IServer interface {
	Host() string
	Name() string
	Serve()
	Stop()
	SetCodec(ICodec)
	SetNewIOModuleCallback(NewIOModuleCallback)
	SetConnCallback(ConnCallback)
	SetMsgCallback(MsgCallback)
	SetExitCallback(ExitCallback)
	SetAuthorizedCallback(AuthorizedCallback)
	SetWriteCompleteCallback(WriteCompleteCallback)
}

type IWSServer interface {
	IServer
	SetMsgType(int)
}

type IClient interface {
	Name() string
	Connect() IConnection
	Close() bool
	SetRetry(bool)
	SetCodec(ICodec)
	SetIOModule(IIOModule)
	SetConnCallback(ConnCallback)
	SetMsgCallback(MsgCallback)
	SendData([]byte)
	SendDataById(string, []byte)
	SendPacket(IPacket)
	SendDataAwait([]byte) (IPacket, error)
	SendDataByIdAwait(string, []byte) (IPacket, error)
	SendPacketAwait(IPacket) (IPacket, error)
	GetConn() IConnection
}

type IConnection interface {
	GetId() string
	LocalAddr() string
	RemoteAddr() string
	Close() bool
	IsClosed() bool
	OnConnection()
	ReadPacket(IPacket)
	WriteData([]byte)
	WriteDataById(string, []byte)
	WritePacket(IPacket)
	WriteDataAwait([]byte) (IPacket, error)
	WriteDataByIdAwait(string, []byte) (IPacket, error)
	WritePacketAwait(IPacket) (IPacket, error)
	WriteComplete()
	SetContext(interface{}, interface{})
	GetContext(interface{}) interface{}
	DelContext(interface{})
	UpdateCodec(ICodec)
}

type ITcpConnection interface {
	GetConn() net.Conn
}

type IWSConnection interface {
	GetConn() *websocket.Conn
	GetMsgType() int
}

type ServeObj interface {
	OnServiceHandle(session ISession, packet IPacket) bool
}

type ISession interface {
	//GetServeObj() ServeObj
	GetConnId() string
	GetSessionId() string
	Close() bool
	CloseSession() bool
	WritePacket(IPacket)
	WriteData([]byte)
	WriteDataById(string, []byte)
	WritePacketAwait(IPacket) (IPacket, error)
	WriteDataAwait([]byte) (IPacket, error)
	WriteDataByIdAwait(string, []byte) (IPacket, error)
	OnService(ISession, IPacket) bool
	SetContext(key, value interface{})
	GetContext(key interface{}) interface{}
	SetPacket(IPacket)
	GetPacket() IPacket
}
