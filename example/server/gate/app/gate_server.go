/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/example/server/gate/service"
)

type GateServer struct {
	*network.Server
}

func NewGateServer(name, path string) *GateServer {
	gs := &GateServer{
		Server: network.NewServer(name, path),
	}
	gs.init()
	gs.initRemote()
	return gs
}

func (gs *GateServer) init() {
	gs.SetCodec(&module.HeadCodec{})
	gs.SetConnCallback(gs.onConn)
	gs.SetMsgCallback(gs.onMsg)

	utils.RegisterService(&service.GateService{})
}

func (gs *GateServer) onConn(conn defs.IConnection) {
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

func (gs *GateServer) onNewConn(conn defs.IConnection) {
	session := network.NewSession(conn, conn.GetId(), false)
	network.AddSession(session)
}

func (gs *GateServer) onDisconn(conn defs.IConnection) {
	network.DelSession(conn.GetId())
}

func (gs *GateServer) onMsg(conn defs.IConnection, packet defs.IPacket) {
	packet.SetSessionId(conn.GetId())
	gs.onClientMsg(conn, packet)
}

