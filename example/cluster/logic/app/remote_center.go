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
	"github.com/lightning-go/lightning/example/cluster/common"
	"github.com/lightning-go/lightning/example/cluster/msg"
	"github.com/lightning-go/lightning/example/cluster/core"
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
	center.SetConnCallback(ls.onCenterConn)
	center.SetMsgCallback(ls.onCenterMsg)
}

func (ls *LogicServer) onCenterConn(conn defs.IConnection) {
	logger.Tracef("%s -> %s server %s is %s",
		conn.LocalAddr(), "center", conn.RemoteAddr(),
		utils.IF(conn.IsClosed(), "down", "up"))
	if conn.IsClosed() {
		ls.onCenterDisconn(conn)
	} else {
		ls.onCenterNewConn(conn)
	}
}

func (ls *LogicServer) onCenterNewConn(conn defs.IConnection) {
	data := common.GetAuthorizedData(int32(common.ST_LOGIC), ls.Name(), common.LogicKey)
	conn.WriteData(data)

	sessionData := make([]*msg.SessionData, 0)
	core.RangeClient(func(sessionId string, s defs.ISession) bool {
		session := &msg.SessionData{
			SessionId: sessionId,
		}
		sessionData = append(sessionData, session)
		return true
	})
	if len(sessionData) > 0 {
		data = common.MarshalDataEx(sessionData)
		p := &defs.Packet{}
		p.SetStatus(msg.RESULT_SYNC_SESSION)
		p.SetData(data)
		conn.WritePacket(p)
	}

}

func (ls *LogicServer) onCenterDisconn(conn defs.IConnection) {
}


func (ls *LogicServer) onCenterMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onCenterMsg: %v - %v - %v", packet.GetSessionId(), packet.GetId(), string(packet.GetData()))

	client := core.GetClient(packet.GetSessionId())
	if client == nil {
		return
	}
	ls.centerService.OnServiceHandle(client, packet)
}

