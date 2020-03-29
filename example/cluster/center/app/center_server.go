/**
 * Created: 2020/3/27
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/example/cluster/common"
	"github.com/lightning-go/lightning/example/cluster/center/service"
	"github.com/lightning-go/lightning/example/cluster/msg"
)

type CenterServer struct {
	*network.Server
}

func NewCenterServer(name, confPath string) *CenterServer {
	cs := &CenterServer{
		Server: network.NewServer(name, confPath),
	}
	cs.init()
	return cs
}

func (cs *CenterServer) init() {
	cs.SetCodec(&module.HeadCodec{})
	cs.SetAuthorizedCallback(cs.onAuthorized)
	cs.SetMsgCallback(cs.onMsg)
	cs.RegisterService(&service.LogicService{})
}

func (cs *CenterServer) onAuthorized(conn defs.IConnection, packet defs.IPacket) bool {
	ok, _ := common.AuthorizedCallback(conn, packet)
	return ok
}

func (cs *CenterServer) onMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onMsg: %v - %v - %v", packet.GetSessionId(), packet.GetId(), string(packet.GetData()))

	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("panic error : %v", err)
		}
	}()

	if cs.checkMsgStatus(conn, packet) {
		return
	}

	cs.OnSessionService(conn, packet)
}

func (cs *CenterServer) checkMsgStatus(conn defs.IConnection, packet defs.IPacket) bool {
	if conn == nil || packet == nil {
		return true
	}
	sessionId := packet.GetSessionId()
	status := packet.GetStatus()

	switch status {
	case msg.RESULT_DISCONN:
		cs.DelSession(sessionId)
	case msg.RESULT_SYNC_SESSION:
		cs.syncSessionData(conn, packet)
	default:
		return false
	}

	return true
}

func (cs *CenterServer) syncSessionData(conn defs.IConnection, packet defs.IPacket) {
	data := packet.GetData()
	if data == nil {
		return
	}
	var sessionData []*msg.SessionData
	err := common.Unmarshal(data, &sessionData)
	if err != nil || sessionData == nil{
		return
	}
	for _, s := range sessionData {
		if s == nil {
			continue
		}
		cs.CheckAddSession(conn, s.SessionId, true)
	}
}
