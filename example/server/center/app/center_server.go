/**
 * @author: Jason
 * Created: 19-5-12
 */

package app

import (
	"github.com/lightning-go/lightning/network"
	"sync"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/module"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/example/server/center/service"
	"github.com/lightning-go/lightning/example/server/global"
)

type CenterServer struct {
	*network.Server
	clientSessions *sync.Map
}

func NewCenterServer(name, path string) *CenterServer {
	gs := &CenterServer{
		Server:         network.NewServer(name, path),
		clientSessions: &sync.Map{},
	}
	gs.init()
	gs.initService()
	return gs
}

func (cs *CenterServer) init() {
	cs.SetCodec(&module.HeadCodec{})
	cs.SetConnCallback(cs.onConn)
	cs.SetMsgCallback(cs.onMsg)
	cs.SetAuthorizedCallback(cs.onAuthorized)
}

func (cs *CenterServer) initService() {
	//utils.RegisterService()
}

func (cs *CenterServer) onConn(conn defs.IConnection) {
	isClosed := conn.IsClosed()
	logger.Tracef("%v server %v <- %v is %v",
		cs.Name(), conn.LocalAddr(), conn.RemoteAddr(),
		utils.IF(isClosed, "down", "up").(string))

	if isClosed {
		cs.onDisconn(conn)
	} else {
		cs.onNewConn(conn)
	}
}

func (cs *CenterServer) onNewConn(conn defs.IConnection) {
}

func (cs *CenterServer) onDisconn(conn defs.IConnection) {
	cs.delClientSession(conn.GetId())
}

func (cs *CenterServer) onAuthorized(conn defs.IConnection, packet defs.IPacket) bool {
	ok, _ := service.AuthorizedCallback(conn, packet)
	return ok
}

func (cs *CenterServer) onMsg(conn defs.IConnection, packet defs.IPacket) {
	logger.Tracef("onMsg: %v - %v - %v", packet.GetSessionId(), packet.GetId(), string(packet.GetData()))

	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("panic error : %v", err)
		}
	}()

	sId := conn.GetId()
	sessionId := packet.GetSessionId()
	session := network.GetSession(sessionId)
	status := packet.GetStatus()

	switch status {
	case global.RESULT_DISCONN:
		network.DelSession(sessionId)
		cs.delSession(sId, sessionId)
		return
	case global.RESULT_SYNC_SESSION:
		cs.syncSessionData(conn, packet)
		return
	}

	if session == nil {
		session = network.NewSession(conn, sessionId, true)
		if session == nil {
			return
		}
		network.AddSession(session)
		cs.addClientSession(sId, sessionId)
	}

	session.OnService(packet)

}

func (cs *CenterServer) syncSessionData(conn defs.IConnection, packet defs.IPacket) {
	sId := conn.GetId()
	data := packet.GetData()
	if data == nil {
		return
	}
	var sessionData []*global.SessionData
	err := global.GetJSONMgr().Unmarshal(data, &sessionData)
	if err != nil {
		return
	}
	if sessionData == nil {
		return
	}
	for _, s := range sessionData {
		if s == nil {
			continue
		}
		session := network.GetSession(s.SessionId)
		if session == nil {
			session = network.NewSession(conn, s.SessionId, true)
		}
		network.AddSession(session)
		cs.addClientSession(sId, s.SessionId)
	}

}

func (cs *CenterServer) addClientSession(sId, sessionId string) {
	f := func() {
		s := &sync.Map{}
		s.Store(sessionId, struct{}{})
		cs.clientSessions.Store(sId, s)
	}

	iSessions, ok := cs.clientSessions.Load(sId)
	if !ok {
		f()
		return
	}
	sessions, ok := iSessions.(*sync.Map)
	if !ok || sessions == nil {
		f()
		return
	}

	sessions.Store(sessionId, struct{}{})
}

func (cs *CenterServer) delClientSession(sId string) {
	iSessions, ok := cs.clientSessions.Load(sId)
	if !ok {
		return
	}
	sessions, ok := iSessions.(*sync.Map)
	if !ok || sessions == nil {
		return
	}
	sessions.Range(func(key, value interface{}) bool {
		sessionId, ok := key.(string)
		if !ok {
			return true
		}
		session := network.GetSession(sessionId)
		if session == nil {
			return true
		}
		session.Close()
		network.DelSession(sessionId)
		return true
	})

}

func (cs *CenterServer) delSession(sId, sessionId string) {
	iSessions, ok := cs.clientSessions.Load(sId)
	if !ok {
		return
	}
	sessions, ok := iSessions.(*sync.Map)
	if !ok || sessions == nil {
		return
	}
	_, ok = sessions.Load(sessionId)
	if !ok {
		return
	}
	sessions.Delete(sessionId)
}


