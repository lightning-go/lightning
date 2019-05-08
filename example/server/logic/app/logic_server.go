/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/example/server/logic/service"
	"github.com/lightning-go/lightning/module"
)

type LogicServer struct {
	*network.Server
}

func NewGameServer(name, path string) *LogicServer {
	gs := &LogicServer{
		Server: network.NewServer(name, path),
	}
	gs.init()
	gs.initService()
	return gs
}

func (gs *LogicServer) init() {
	gs.SetCodec(&module.HeadCodec{})
	gs.SetConnCallback(gs.onConn)
	gs.SetMsgCallback(gs.onMsg)
	gs.SetAuthorizedCallback(gs.onAuthorized)
}

func (gs *LogicServer) initService() {
	utils.RegisterService(&service.Service{})
}

func (gs *LogicServer) onConn(conn defs.IConnection) {
	isClosed := conn.IsClosed()
	logger.Tracef("%v server %v <- %v is %v",
		gs.Name(), conn.LocalAddr(), conn.RemoteAddr(),
		utils.IF(isClosed, "down", "up").(string))

	if isClosed {
		gs.onDisconn(conn)
	} else {
		gs.onNewConn(conn)
	}
}

func (gs *LogicServer) onNewConn(conn defs.IConnection) {
}

func (gs *LogicServer) onDisconn(conn defs.IConnection) {
}

func (gs *LogicServer) onAuthorized(conn defs.IConnection, packet defs.IPacket) bool {
	ok, _ := service.AuthorizedCallback(conn, packet)
	return ok
}

func (gs *LogicServer) onMsg(conn defs.IConnection, packet defs.IPacket) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("panic error : %v", err)
		}
	}()

	sessionId := packet.GetSessionId()
	session := network.GetSession(sessionId)

	if packet.GetStatus() == -1 {
		network.DelSession(sessionId)
		return
	}

	if session == nil {
		session = network.NewSession(conn, sessionId, true)
		if session == nil {
			conn.Close()
			return
		}
		network.AddSession(session)
	}
	session.OnService(packet)
}
