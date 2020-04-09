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
)

type LogicServer struct {
	*network.Server
	etcdMgr       *etcd.Etcd
	centerService *utils.ServiceFactory
}

func NewLogicServer(name, confPath string) *LogicServer {
	gs := &LogicServer{
		Server:        network.NewServer(name, confPath),
		centerService: utils.NewServiceFactory(),
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

	ls.OnSessionService(conn, packet)
}

func (ls *LogicServer) checkMsgStatus(conn defs.IConnection, packet defs.IPacket) bool {
	if conn == nil || packet == nil {
		return true
	}
	sessionId := packet.GetSessionId()
	status := packet.GetStatus()

	if status == msg.RESULT_DISCONN {
		ls.SendToCenter(packet)
		ls.DelSession(sessionId)
		return true
	}

	return false
}
