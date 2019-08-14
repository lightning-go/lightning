/**
 * @author: Jason
 * Created: 19-5-12
 */

package app

import (
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/example/server/global"
	"github.com/lightning-go/lightning/network"
)

func (ls *LogicServer) initRemote() {
	remote := ls.GetRemoteClient("center")
	if remote == nil {
		logger.Error("get remote client failed")
		return
	}
	remote.SetCodec(&module.HeadCodec{})
	remote.SetConnCallback(ls.onRemoteConn)
	remote.SetMsgCallback(ls.onRemoteMsg)
}

func (ls *LogicServer) onRemoteConn(conn defs.IConnection) {
	logger.Tracef("%s -> %s server %s is %s",
		conn.LocalAddr(), "remote", conn.RemoteAddr(),
		utils.IF(conn.IsClosed(), "down", "up"))
	if conn.IsClosed() {
		ls.onRemoteDisconn(conn)
	} else {
		ls.onRemoteNewConn(conn)
	}
}

func (ls *LogicServer) onRemoteNewConn(conn defs.IConnection) {
	data := global.GetAuthorizedData(global.ST_GAME, ls.Name(), global.GameKey)
	conn.WriteData(data)

	sessionData := make([]*global.SessionData, 0)
	network.RangeSession(func(sId string, s *network.Session) {
		session := &global.SessionData{
			SessionId: sId,
		}
		sessionData = append(sessionData, session)
	})
	if len(sessionData) > 0 {
		data = global.GetJSONMgr().MarshalData(sessionData)
		p := &defs.Packet{}
		p.SetStatus(global.RESULT_SYNC_SESSION)
		p.SetData(data)
		conn.WritePacket(p)
	}
}

func (ls *LogicServer) onRemoteDisconn(conn defs.IConnection) {
}

func (ls *LogicServer) onRemoteMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onRemoteMsg: %v - %v - %v", packet.GetSessionId(), packet.GetId(), string(packet.GetData()))


}

