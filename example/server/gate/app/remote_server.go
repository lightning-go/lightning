/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/example/server/global"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/network"
)

func (gs *GateServer) initRemote() {
	remote := gs.GetRemoteClient("logic")
	if remote == nil {
		logger.Error("get remote client failed")
		return
	}
	remote.SetCodec(&module.HeadCodec{})
	remote.SetConnCallback(gs.onRemoteConn)
	remote.SetMsgCallback(gs.onRemoteMsg)
}

func (gs *GateServer) onRemoteConn(conn defs.IConnection) {
	logger.Tracef("%s -> %s server %s is %s",
		conn.LocalAddr(), "remote", conn.RemoteAddr(),
		utils.IF(conn.IsClosed(), "down", "up"))
	if conn.IsClosed() {
		gs.onRemoteDisconn(conn)
	} else {
		gs.onRemoteNewConn(conn)
	}
}

func (gs *GateServer) onRemoteNewConn(conn defs.IConnection) {
	data := global.GetAuthorizedData(global.ST_GATE, gs.Name(), global.GateKey)
	conn.WriteData(data)
}

func (gs *GateServer) onRemoteDisconn(conn defs.IConnection) {
}

func (gs *GateServer) onRemoteMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onRemoteMsg: %v - %v - %v", packet.GetSessionId(), packet.GetId(), string(packet.GetData()))

	session := network.GetSession(packet.GetSessionId())
	if session == nil {
		return
	}
	session.WritePacket(packet)
}

func (gs *GateServer) onClientMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onClientMsg: %v - %v - %v", packet.GetSessionId(), packet.GetId(), string(packet.GetData()))

	remote := gs.GetRemoteClient("logic")
	if remote == nil {
		return
	}
	remote.SendPacket(packet)
}
