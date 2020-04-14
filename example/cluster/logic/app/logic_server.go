/**
 * Created: 2020/3/26
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/etcd"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"runtime/debug"
	"github.com/lightning-go/lightning/example/cluster/msg"
	"github.com/lightning-go/lightning/example/cluster/logic/service"
	"github.com/lightning-go/lightning/example/cluster/common"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/example/cluster/core"
)

type LogicServer struct {
	*network.Server
	etcdMgr       *etcd.Etcd
	centerService *utils.ServiceFactory
	clientMgr    *network.SessionMgr
}

func NewLogicServer(name, confPath string) *LogicServer {
	gs := &LogicServer{
		Server:        network.NewServer(name, confPath),
		centerService: utils.NewServiceFactory(),
		clientMgr:    network.NewSessionMgr(),
	}
	gs.initLog()
	gs.init()
	gs.initRemoteCenter()
	return gs
}

func (ls *LogicServer) init() {
	ls.SetCodec(&module.HeadCodec{})
	ls.SetAuthorizedCallback(ls.onAuthorized)
	ls.SetMsgCallback(ls.onMsg)
	ls.SetDisConnCallback(ls.onDisConn)
	ls.RegisterService(service.NewLogicService(ls))
	ls.centerService.Register(&service.CenterService{})
	ls.registerEtcd()
}

func (ls *LogicServer) initLog() {
	logConf := conf.GetLogConf("logic")
	if logConf != nil {
		logger.InitLog(logConf.LogLevel, logConf.MaxAge, logConf.RotationTime, logConf.LogPath)
	}
}

func (ls *LogicServer) onDisConn(conn defs.IConnection) {
	ls.clientMgr.DelConnSession(conn.GetId())
}

func (ls *LogicServer) onAuthorized(conn defs.IConnection, packet defs.IPacket) bool {
	ok, _ := common.AuthorizedCallback(conn, packet)
	return ok
}

func (ls *LogicServer) onMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onMsg %v, %s", packet.GetSessionId(), packet.GetData())
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Error(string(debug.Stack()))
		}
	}()

	if ls.checkMsgStatus(conn, packet) {
		return
	}

	ls.OnClientService(conn, packet)
}

func (ls *LogicServer) checkMsgStatus(conn defs.IConnection, packet defs.IPacket) bool {
	if conn == nil || packet == nil {
		return true
	}
	sessionId := packet.GetSessionId()
	status := packet.GetStatus()

	if status == msg.RESULT_DISCONN {
		ls.SendToCenter(packet)
		ls.DelClient(sessionId)
		return true
	}

	return false
}

func (ls *LogicServer) GetClient(sessionId string) defs.ISession {
	return ls.clientMgr.GetSession(sessionId)
}

func (ls *LogicServer) DelClient(sessionId string) {
	ls.clientMgr.DelSession(sessionId)
}

func (ls *LogicServer) RangeClient(f func(string, defs.ISession)) {
	ls.clientMgr.RangeSession(f)
}

func (ls *LogicServer) GetClientCount() int64 {
	return ls.clientMgr.SessionCount()
}

func (ls *LogicServer) CheckAddClient(conn defs.IConnection, sessionId string, async ...bool) defs.ISession {
	client := ls.clientMgr.GetSession(sessionId)
	if client == nil {
		client = core.NewClient(conn, sessionId, ls, async...)
		if client == nil {
			return nil
		}
		ls.clientMgr.AddSession(client)
	}
	return client
}

func (ls *LogicServer) OnClientService(conn defs.IConnection, packet defs.IPacket) {
	sessionId := packet.GetSessionId()
	client := ls.CheckAddClient(conn, sessionId, true)
	if client == nil {
		logger.Errorf("session nil %v", sessionId)
		return
	}
	client.OnService(client, packet)
}
