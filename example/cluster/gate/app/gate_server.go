/**
 * Created: 2020/3/25
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/etcd"
	"runtime/debug"
	"time"
	"github.com/lightning-go/lightning/example/cluster/gate/service"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/example/cluster/msg"
	"github.com/lightning-go/lightning/conf"
)

const (
	LAST_MSG_TIMESTAMP_KEY = "LAST_MSG_TIMESTAMP_KEY"
	MSG_COUNT_KEY          = "MSG_COUNT_KEY"
)

type GateServer struct {
	*network.Server
	etcdMgr       *etcd.Etcd
	serveSelector *ServeSelector
}

func NewGateServer(name, confPath string) *GateServer {
	gs := &GateServer{
		Server:        network.NewServer(name, confPath),
		serveSelector: NewSelector(),
	}
	gs.initLog()
	gs.init()
	return gs
}

func (gs *GateServer) init() {
	gs.initLog()

	gs.SetCodec(&module.HeadCodec{})
	gs.SetMsgCallback(gs.onMsg)
	gs.SetDisConnCallback(gs.onDisConn)

	gs.initEtcd()
	gs.RegisterService(&service.GateService{})

	gs.serveSelector.SetCleanSessionCallback(func(sessionId string) {
		session := gs.GetConn(sessionId)
		if session != nil {
			session.Close()
		}
	})
}

func (gs *GateServer) initLog() {
	logConf := conf.GetLogConf("gate")
	if logConf != nil {
		logger.InitLog(logConf.LogLevel, logConf.MaxAge, logConf.RotationTime, logConf.LogPath)
	}
}

func (gs *GateServer) onDisConn(conn defs.IConnection) {
	gs.disconnection(conn.GetId())
}

func (gs *GateServer) onMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onMsg %v, %s", conn.GetId(), packet.GetData())
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Error(string(debug.Stack()))
		}
	}()

	sessionId := conn.GetId()
	session := gs.GetConn(sessionId)
	if session == nil {
		return
	}
	if !gs.isMsgValid(session) {
		session.Close()
		return
	}

	packet.SetSessionId(sessionId)
	gs.onClientMsg(session, packet)
}

func (gs *GateServer) isMsgValid(session defs.ISession) bool {
	if session == nil {
		return false
	}

	now := time.Now().Unix()
	var lastTimestamp int64 = 0
	var msgCount int64 = 0

	iLastTimestamp := session.GetContext(LAST_MSG_TIMESTAMP_KEY)
	if iLastTimestamp != nil {
		lastTimestamp = iLastTimestamp.(int64)
	}

	var maxMsgCount int64 = 100 //todo constant
	val := now - lastTimestamp
	if val <= 1 {
		iCount := session.GetContext(MSG_COUNT_KEY)
		if iCount != nil {
			msgCount = iCount.(int64)
		}
		if msgCount > maxMsgCount {
			logger.Warnf("recv msg count > %v in 1 sec, msgCount %v, addr: %v",
				maxMsgCount, msgCount, session.(*network.Session).GetConn().RemoteAddr())
			return false
		}
		msgCount++
		session.SetContext(MSG_COUNT_KEY, msgCount)
		logger.Debugf("recv msg count %v in 1 sec", msgCount)

	} else {
		msgCount = 0
		session.SetContext(MSG_COUNT_KEY, msgCount)
		session.SetContext(LAST_MSG_TIMESTAMP_KEY, now)
	}

	return true
}

func (gs *GateServer) onClientMsg(session defs.ISession, packet defs.IPacket) {
	if gs.onGateService(session, packet) {
		return
	}

	sessionId := session.GetSessionId()
	remote := gs.serveSelector.GetRemoteSession(sessionId)
	if remote != nil {
		remote.SendPacket(packet)
		return
	}

	sessionData := gs.serveSelector.GetRemoteData()
	if sessionData == nil {
		logger.Warn("get remote session data failed")
		return
	}

	remote = gs.serveSelector.GetRemoteClient(sessionData.Name)
	if remote == nil {
		logger.Warnf("remote: %v nil", sessionData.Name)
		return
	}
	remote.SendPacket(packet)

	gs.serveSelector.AddRemoteSession(sessionId, remote)
}

func (gs *GateServer) onGateService(session defs.ISession, packet defs.IPacket) bool {
	return gs.OnServiceHandle(session, packet)
}

func (gs *GateServer) disconnection(sessionId string) {
	remote := gs.serveSelector.GetRemoteSession(sessionId)
	if remote == nil {
		return
	}
	p := &defs.Packet{}
	p.SetSessionId(sessionId)
	p.SetData(utils.NullData)
	p.SetStatus(msg.RESULT_DISCONN)
	remote.SendPacket(p)

	gs.serveSelector.DelRemoteSession(sessionId)
	gs.serveSelector.delSessionIdMap(sessionId, remote)
}
