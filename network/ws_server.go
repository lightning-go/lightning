/**
 * Created: 2019/4/23 0023
 * @author: Jason
 */

package network

import (
	"net"
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/logger"
	"time"
)

type WSServer struct {
	listener         net.Listener
	addr             string
	path             string
	name             string
	maxConn          int
	msgType          int
	enablePong       bool
	connMgr          *ConnectionMgr
	upgrader         *websocket.Upgrader
	httpSrv          *http.Server
	codec            defs.ICodec
	ioModuleCallback defs.NewIOModuleCallback
	connCallback     defs.ConnCallback
	msgCallback      defs.MsgCallback
	exitCallback     defs.ExitCallback
	authCallback     defs.AuthorizedCallback
	writeComplete    defs.WriteCompleteCallback
}

func NewWSServer(name, addr string, maxConn int, path ...string) *WSServer {
	wss := &WSServer{
		listener:   ListenTcp(addr),
		name:       name,
		addr:       addr,
		path:       "/",
		maxConn:    maxConn,
		msgType:    websocket.TextMessage,
		enablePong: false,
		connMgr:    NewConnMgr(),
	}
	if len(path) > 0 && len(path[0]) > 0 {
		wss.path = path[0]
	}
	return wss
}

func (ws *WSServer) EnablePong(val bool) {
	ws.enablePong = val
}

func (ws *WSServer) SetMsgType(msgType int) {
	ws.msgType = msgType
}

func (ws *WSServer) SetCodec(codec defs.ICodec) {
	ws.codec = codec
}

func (ws *WSServer) SetNewIOModuleCallback(newCallback defs.NewIOModuleCallback) {
	ws.ioModuleCallback = newCallback
}

func (ws *WSServer) SetConnCallback(cb defs.ConnCallback) {
	ws.connCallback = cb
}

func (ws *WSServer) SetMsgCallback(cb defs.MsgCallback) {
	ws.msgCallback = cb
}

func (ws *WSServer) SetExitCallback(cb defs.ExitCallback) {
	ws.exitCallback = cb
}

func (ws *WSServer) SetAuthorizedCallback(cb defs.AuthorizedCallback) {
	ws.authCallback = cb
}

func (ws *WSServer) SetWriteCompleteCallback(cb defs.WriteCompleteCallback) {
	ws.writeComplete = cb
}

func (ws *WSServer) Host() string {
	return ws.listener.Addr().String()
}

func (ws *WSServer) Name() string {
	return ws.name
}

func (ws *WSServer) Serve() {
	timeout := conf.GetGlobalVal().HttpTimeout

	ws.upgrader = &websocket.Upgrader{
		HandshakeTimeout: timeout,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(ws.path, ws.ServeHTTP)

	ws.httpSrv = &http.Server{
		Addr:           ws.addr,
		Handler:        mux,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1024,
	}

	logger.Infof("%v server start, listen %v", ws.name, ws.listener.Addr().String())
	go ws.httpSrv.Serve(ws.listener)

	GetSrvMgr().AddServer(ws)
}

func (ws *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	onlineCount := ws.connMgr.ConnCount()
	if onlineCount >= ws.maxConn {
		http.Error(w, "Too many connections", http.StatusInternalServerError)
		return
	}

	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err)
		http.Error(w, "Method upgrade failed", http.StatusInternalServerError)
		return
	}

	go ws.connectionHandle(conn)

}

func (ws *WSServer) connectionHandle(conn *websocket.Conn) {
	if conn == nil {
		return
	}

	conn.SetReadLimit(int64(conf.GetGlobalVal().MaxPacketSize))

	if ws.enablePong {
		conn.SetReadDeadline(time.Now().Add(conf.GetGlobalVal().PongWait))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(conf.GetGlobalVal().PongWait))
			return nil
		})
	}

	wsConn := ws.newConnection(conn)
	if wsConn == nil {
		logger.Error("alloc new ws connection failed")
		conn.Close()
	}

	ok := wsConn.Start()
	if !ok {
		logger.Error("new ws connection start failed")
		conn.Close()
		return
	}

	ws.connMgr.AddConn(wsConn)
}

func (ws *WSServer) newConnection(conn *websocket.Conn) *WSConnection {
	wsConn := NewWSConnection(conn)
	if wsConn == nil {
		return nil
	}
	if ws.ioModuleCallback != nil {
		wsConn.SetIOModule(ws.ioModuleCallback(wsConn))
	}
	wsConn.SetMsgType(ws.msgType)
	wsConn.SetCodec(ws.codec)
	wsConn.SetCloseCallback(ws.closeConnection)
	wsConn.SetConnCallback(ws.connCallback)
	wsConn.SetMsgCallback(ws.msgCallback)
	wsConn.SetAuthorizedCallback(ws.authCallback)
	wsConn.SetWriteCompleteCallback(ws.writeComplete)
	return wsConn
}

func (ws *WSServer) closeConnection(conn defs.IConnection) {
	if conn == nil {
		return
	}
	logger.Tracef("CloseConnection: %v", conn.GetId())
	ws.connMgr.DelConn(conn.GetId())
	conn.OnConnection()
}

func (ws *WSServer) Stop() {
	if ws.exitCallback != nil {
		ws.exitCallback()
	}

	logger.Warnf("stop %v server", ws.name)
}
