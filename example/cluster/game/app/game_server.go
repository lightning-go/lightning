/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/conf"
	"lightning/common/util"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/defs"
)

type GameServer struct {
	*network.Server
}

func NewGameServer(name string) *GameServer {
	gs := &GameServer{
		Server: network.NewServer(name, conf.GetConfPath()),
	}
	gs.init()
	return gs
}

func (gs *GameServer) init() {
	gs.SetConnCallback(gs.onConn)
	gs.SetMsgCallback(gs.onMsg)
	gs.SetAuthorizedCallback(gs.onAuthorized)
}

func (gs *GameServer) onConn(conn defs.IConnection) {
	isClosed := conn.IsClosed()
	logger.Trace("%v server %v <- %v is %v",
		gs.Name(), conn.LocalAddr(), conn.RemoteAddr(),
		util.IF(isClosed, "down", "up").(string))

	if isClosed {
		gs.onDisconn(conn)
	} else {
		gs.onNewConn(conn)
	}
}

func (gs *GameServer) onNewConn(conn defs.IConnection) {
}

func (gs *GameServer) onDisconn(conn defs.IConnection) {
}

func (gs *GameServer) onAuthorized(conn defs.IConnection, packet defs.IPacket) bool {

	return true
}

func (gs *GameServer) onMsg(conn defs.IConnection, packet defs.IPacket) {

}
