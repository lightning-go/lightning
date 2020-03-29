/**
 * Created: 2020/3/27
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/example/cluster/common"
	"github.com/lightning-go/lightning/example/cluster/msg"
)

func (ls *LogicServer) SendDataToCenter(session defs.ISession, d interface{}) {
	if session == nil || d == nil {
		return
	}
	packet := session.GetPacket()
	id := utils.IF(packet == nil, "", packet.GetId()).(string)
	sessionId := utils.IF(packet == nil, "", packet.GetSessionId()).(string)

	data := common.MarshalDataEx(d)
	p := &defs.Packet{}
	p.SetSessionId(sessionId)
	p.SetId(id)
	p.SetData(data)
	ls.SendToCenter(p)
}

func (ls *LogicServer) SendToCenter(packet defs.IPacket) {
	if packet == nil {
		return
	}
	center := ls.GetRemoteClient("center")
	if center == nil {
		logger.Warn("remote center client instance nil")
		return
	}
	center.SendPacket(packet)
}

func (ls *LogicServer) initRemoteCenter() {
	center := ls.GetRemoteClient("center")
	if center == nil {
		logger.Error("GetRemoteClient center failed")
		return
	}
	center.SetCodec(&module.HeadCodec{})
	center.SetConnCallback(ls.onRemoteConn)
	center.SetMsgCallback(ls.onRemoteMsg)
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
	data := common.GetAuthorizedData(int32(common.ST_LOGIC), ls.Name(), common.LogicKey)
	conn.WriteData(data)

	sessionData := make([]*msg.SessionData, 0)
	ls.RangeSession(func(sessionId string, s *network.Session) {
		session := &msg.SessionData{
			SessionId: sessionId,
		}
		sessionData = append(sessionData, session)
	})
	if len(sessionData) > 0 {
		data = common.MarshalDataEx(sessionData)
		p := &defs.Packet{}
		p.SetStatus(msg.RESULT_SYNC_SESSION)
		p.SetData(data)
		conn.WritePacket(p)
	}

}

func (ls *LogicServer) onRemoteDisconn(conn defs.IConnection) {
}


func (ls *LogicServer) onRemoteMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onRemoteMsg: %v - %v - %v", packet.GetSessionId(), packet.GetId(), string(packet.GetData()))

	session := ls.GetSession(packet.GetSessionId())
	if session == nil {
		return
	}
	ls.centerService.OnServiceHandle(session, packet)
}

