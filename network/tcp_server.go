/**
 * Created: 2019/4/18 0018
 * @author: Jason
 */

package network

import (
	"net"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"time"
)

const ReAcceptDelay = 5

type TcpServer struct {
	listener              net.Listener
	name                  string
	maxConn               int
	connMgr               *ConnectionMgr
	codec                 defs.ICodec
	ioModule              defs.IIOModule
	connCallback          defs.ConnCallback
	msgCallback           defs.MsgCallback
	exitCallback          defs.ExitCallback
	authCallback          defs.AuthorizedCallback
	writeCompleteCallback defs.WriteCompleteCallback
}

func NewTcpServer(addr, name string, maxConn int) *TcpServer {
	return &TcpServer{
		listener: ListenTcp(addr),
		name:     name,
		maxConn:  maxConn,
		connMgr:  NewConnMgr(),
	}
}

func ListenTcp(addr string) net.Listener {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	return listener
}

func (tcpServer *TcpServer) SetCodec(codec defs.ICodec) {
	tcpServer.codec = codec
}

func (tcpServer *TcpServer) SetIOModule(ioModule defs.IIOModule) {
	tcpServer.ioModule = ioModule
}

func (tcpServer *TcpServer) SetConnCallback(cb defs.ConnCallback) {
	tcpServer.connCallback = cb
}

func (tcpServer *TcpServer) SetMsgCallback(cb defs.MsgCallback) {
	tcpServer.msgCallback = cb
}

func (tcpServer *TcpServer) SetExitCallback(cb defs.ExitCallback) {
	tcpServer.exitCallback = cb
}

func (tcpServer *TcpServer) SetAuthorizedCallback(cb defs.AuthorizedCallback) {
	tcpServer.authCallback = cb
}

func (tcpServer *TcpServer) SetWriteCompleteCallback(cb defs.WriteCompleteCallback) {
	tcpServer.writeCompleteCallback = cb
}

func (tcpServer *TcpServer) Name() string {
	return tcpServer.name
}

func (tcpServer *TcpServer) Serve() {
	tcpServer.start()
}

func (tcpServer *TcpServer) start() {
	go tcpServer.serveTcp()
	GetSrvMgr().AddServer(tcpServer)
}

func (tcpServer *TcpServer) serveTcp() {
	logger.Infof("%v start, listen %v", tcpServer.name, tcpServer.listener.Addr().String())
	var tmpDelay time.Duration
	maxDelay := 1 * time.Second

	for {
		conn, err := tcpServer.listener.Accept()
		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Temporary() {
				if tmpDelay == 0 {
					tmpDelay = time.Millisecond * ReAcceptDelay
				} else {
					tmpDelay *= 2
				}
				if tmpDelay > maxDelay {
					tmpDelay = maxDelay
				}
				logger.Warnf("%v accept error: %v, retrying in %v millisecond",
					tcpServer.name, err, tmpDelay)
				time.Sleep(tmpDelay)
				continue
			}
		}
		tmpDelay = 0

		onlineCount := tcpServer.connMgr.ConnCount()
		if onlineCount >= 3000 { //todo
			//todo conn limit
		} else {
			go tcpServer.connectionHandle(conn)
		}
	}
}

func (tcpServer *TcpServer) connectionHandle(conn net.Conn) {
	if conn == nil {
		return
	}

	newConn := tcpServer.newConnection(conn)
	if newConn == nil {
		logger.Error("alloc new connection failed")
		conn.Close()
		return
	}

	ok := newConn.Start()
	if !ok {
		logger.Error("new connection start failed")
		conn.Close()
		return
	}

	tcpServer.connMgr.AddConn(newConn)
}

func (tcpServer *TcpServer) newConnection(conn net.Conn) *Connection {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil
	}
	tcpConn.SetNoDelay(true)

	newConn := NewConnection(conn)
	if newConn == nil {
		return nil
	}

	newConn.SetCodec(tcpServer.codec)
	newConn.SetIOModule(tcpServer.ioModule)
	newConn.SetCloseCallback(tcpServer.CloseConnection)
	newConn.SetConnCallback(tcpServer.connCallback)
	newConn.SetMsgCallback(tcpServer.msgCallback)
	newConn.SetAuthorizedCallback(tcpServer.authCallback)
	newConn.SetWriteCompleteCallback(tcpServer.writeCompleteCallback)
	return newConn
}

func (tcpServer *TcpServer) CloseConnection(conn defs.IConnection) {
	if conn == nil {
		return
	}
	logger.Tracef("close connection: %v", conn.GetId())
	tcpServer.connMgr.DelConn(conn.GetId())
	conn.OnConnection()
}

func (tcpServer *TcpServer) Stop() {
	if tcpServer.exitCallback != nil {
		tcpServer.exitCallback()
	}

	logger.Warnf("stop %v server", tcpServer.name)
}
