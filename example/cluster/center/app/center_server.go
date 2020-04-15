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
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/example/cluster/core"
)

type CenterServer struct {
	*network.Server
}

func NewCenterServer(name, confPath string) *CenterServer {
	cs := &CenterServer{
		Server:    network.NewServer(name, confPath),
	}
	cs.init()
	return cs
}

func (cs *CenterServer) init() {
	cs.initLog()

	cs.SetCodec(&module.HeadCodec{})
	cs.SetAuthorizedCallback(cs.onAuthorized)
	cs.SetMsgCallback(cs.onMsg)
	cs.SetDisConnCallback(cs.onDisConn)

	cs.RegisterService(&service.LogicService{})
}

func (cs *CenterServer) initLog() {
	logConf := conf.GetLogConf("center")
	if logConf != nil {
		logger.InitLog(logConf.LogLevel, logConf.MaxAge, logConf.RotationTime, logConf.LogPath)
	}
}

func (cs *CenterServer) onDisConn(conn defs.IConnection) {
	core.DelClientByConnId(conn.GetId())
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

	cs.OnClientService(conn, packet)
}

func (cs *CenterServer) checkMsgStatus(conn defs.IConnection, packet defs.IPacket) bool {
	if conn == nil || packet == nil {
		return true
	}
	sessionId := packet.GetSessionId()
	status := packet.GetStatus()

	switch status {
	case msg.RESULT_DISCONN:
		core.DelClient(sessionId)
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
	if err != nil || sessionData == nil {
		return
	}
	for _, s := range sessionData {
		if s == nil {
			continue
		}
		core.CheckAddClient(conn, s.SessionId, cs, true)
	}
}

func (cs *CenterServer) OnClientService(conn defs.IConnection, packet defs.IPacket) {
	sessionId := packet.GetSessionId()
	client := core.CheckAddClient(conn, sessionId, cs, true)
	if client == nil {
		logger.Errorf("session nil %v", sessionId)
		return
	}
	client.OnService(client, packet)
}
