/**
 * Created: 2020/3/25
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/selector"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/utils"
	"time"
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/example/cluster/gate/service"
	"github.com/lightning-go/lightning/example/cluster/common"
)

type RemoteClient struct {
	*network.TcpClient
	service *utils.ServiceFactory
	sd      *selector.SessionData
	gate    *GateServer
}

func NewRemoteClient(name, addr string) *RemoteClient {
	rc := &RemoteClient{
		TcpClient: network.NewTcpClient(name, addr),
		service:   utils.NewServiceFactory(),
	}
	var timeout int64 = 3 //todo constant
	rc.SetTimeout(time.Duration(timeout) * time.Second)
	rc.SetRetry(false)
	rc.SetCodec(&module.HeadCodec{})
	rc.SetConnCallback(rc.onRemoteConn)
	rc.SetMsgCallback(rc.onRemoteMsg)
	rc.service.Register(&service.LogicService{})
	return rc
}

func (rc *RemoteClient) onRemoteConn(conn defs.IConnection) {
	closed := conn.IsClosed()
	logger.Tracef("%s -> %s server %s is %s",
		conn.LocalAddr(), "remote", conn.RemoteAddr(),
		utils.IF(closed, "down", "up"))
	if closed {
		rc.onRemoteDisconn(conn)
	} else {
		rc.onRemoteNewConn(conn)
	}
}

func (rc *RemoteClient) onRemoteNewConn(conn defs.IConnection) {
	d := common.GetAuthorizedData(int32(common.ST_GATE), rc.gate.Name(), common.GateKey)
	rc.SendData(d)
	rc.gate.serveSelector.AddRemoteClient(rc)
}

func (rc *RemoteClient) onRemoteDisconn(conn defs.IConnection) {
	rc.gate.serveSelector.DelRemoteClient(rc)
}

func (rc *RemoteClient) onRemoteMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onRemoteMsg: %v - %v - %v", packet.GetSessionId(), packet.GetId(), string(packet.GetData()))

	sessionId := packet.GetSessionId()
	session := rc.gate.GetConn(sessionId)
	if session == nil {
		return
	}

	if rc.service.OnServiceHandle(session, packet) {
		return
	}

	session.WritePacket(packet)
}

func (gs *GateServer) initRemote(sd *selector.SessionData) bool {
	if sd == nil {
		return false
	}
	remote := NewRemoteClient(gs.Name(), sd.Host)
	if remote == nil {
		return false
	}
	remote.gate = gs
	remote.sd = sd
	remote.Connect()
	return remote.IsWorking()
}
