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
	"github.com/lightning-go/lightning/example/cluster/data"
)

type CenterServer struct {
	*network.Server
	clientMgr *network.SessionMgr
}

func NewCenterServer(name, confPath string) *CenterServer {
	cs := &CenterServer{
		Server:    network.NewServer(name, confPath),
		clientMgr: network.NewSessionMgr(),
	}
	cs.initLog()
	cs.init()
	return cs
}

func (cs *CenterServer) init() {
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
	cs.clientMgr.DelConnSession(conn.GetId())
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
		cs.DelClient(sessionId)
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
		cs.CheckAddClient(conn, s.SessionId, true)
	}
}

func (cs *CenterServer) GetClient(sessionId string) defs.ISession {
	return cs.clientMgr.GetSession(sessionId)
}

func (cs *CenterServer) DelClient(sessionId string) {
	cs.clientMgr.DelSession(sessionId)
}

func (cs *CenterServer) RangeClient(f func(string, defs.ISession)) {
	cs.clientMgr.RangeSession(f)
}

func (cs *CenterServer) GetClientCount() int64 {
	return cs.clientMgr.SessionCount()
}

func (cs *CenterServer) CheckAddClient(conn defs.IConnection, sessionId string, async ...bool) defs.ISession {
	client := cs.clientMgr.GetSession(sessionId)
	if client == nil {
		client = data.NewClient(conn, sessionId, cs, async...)
		if client == nil {
			return nil
		}
		cs.clientMgr.AddSession(client)
	}
	return client
}

func (cs *CenterServer) OnClientService(conn defs.IConnection, packet defs.IPacket) {
	sessionId := packet.GetSessionId()
	client := cs.CheckAddClient(conn, sessionId, true)
	if client == nil {
		logger.Errorf("session nil %v", sessionId)
		return
	}
	client.OnService(client, packet)
}
