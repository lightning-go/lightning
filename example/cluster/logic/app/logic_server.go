/**
 * Created: 2020/3/26
 * @author: Jason
 */

package app

import (
	"fmt"
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
}

func NewLogicServer(name, confPath string) *LogicServer {
	ls := &LogicServer{
		Server:        network.NewServer(name, confPath),
		centerService: utils.NewServiceFactory(),
	}
	ls.init()
	return ls
}

func (ls *LogicServer) init() {
	ls.initLog()

	ls.SetCodec(&module.HeadCodec{})
	ls.SetAuthorizedCallback(ls.onAuthorized)
	ls.SetMsgCallback(ls.onMsg)
	ls.SetDisConnCallback(ls.onDisConn)

	ls.RegisterService(service.NewLogicService(ls))
	ls.centerService.Register(&service.CenterService{})

	ls.initEtcd()
	ls.initRemoteCenter()
}

func (ls *LogicServer) initLog() {
	logConf := conf.GetLogCfg("logic")
	if logConf != nil {
		logLv := logger.GetLevel(logConf.LogLevel)
		pathFile := fmt.Sprintf("%s/%s", logConf.LogPath, logConf.LogFile)
		logger.InitLog(logLv, logConf.MaxAge, logConf.RotationTime, pathFile)
	}
}

func (ls *LogicServer) onDisConn(conn defs.IConnection) {
	core.DelClientByConnId(conn.GetId())
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
		core.DelClient(sessionId)
		return true
	}

	return false
}

func (ls *LogicServer) OnClientService(conn defs.IConnection, packet defs.IPacket) {
	sessionId := packet.GetSessionId()
	client := core.CheckAddClient(conn, sessionId, ls.OnServiceHandle, true)
	if client == nil {
		logger.Errorf("session nil %v", sessionId)
		return
	}
	client.OnService(client, packet)
}
