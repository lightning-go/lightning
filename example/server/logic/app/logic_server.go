/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/network"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/example/server/logic/service"
	"github.com/lightning-go/lightning/module"
	"sync"
)

type LogicServer struct {
	*network.Server
	clientSessions *sync.Map //map[string]map[string]struct{}
}

func NewGameServer(name, path string) *LogicServer {
	gs := &LogicServer{
		Server:         network.NewServer(name, path),
		clientSessions: &sync.Map{}, //make(map[string]map[string]struct{}),
	}
	gs.init()
	gs.initService()
	return gs
}

func (gs *LogicServer) init() {
	gs.SetCodec(&module.HeadCodec{})
	gs.SetConnCallback(gs.onConn)
	gs.SetMsgCallback(gs.onMsg)
	gs.SetAuthorizedCallback(gs.onAuthorized)
}

func (gs *LogicServer) initService() {
	utils.RegisterService(&service.Service{})
}

func (gs *LogicServer) onConn(conn defs.IConnection) {
	isClosed := conn.IsClosed()
	logger.Tracef("%v server %v <- %v is %v",
		gs.Name(), conn.LocalAddr(), conn.RemoteAddr(),
		utils.IF(isClosed, "down", "up").(string))

	if isClosed {
		gs.onDisconn(conn)
	} else {
		gs.onNewConn(conn)
	}
}

func (gs *LogicServer) onNewConn(conn defs.IConnection) {
}

func (gs *LogicServer) onDisconn(conn defs.IConnection) {
	gs.delClientSession(conn.GetId())
}

func (gs *LogicServer) onAuthorized(conn defs.IConnection, packet defs.IPacket) bool {
	ok, _ := service.AuthorizedCallback(conn, packet)
	return ok
}

func (gs *LogicServer) onMsg(conn defs.IConnection, packet defs.IPacket) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("panic error : %v", err)
		}
	}()

	sId := conn.GetId()
	sessionId := packet.GetSessionId()
	session := network.GetSession(sessionId)

	if packet.GetStatus() == -1 {
		network.DelSession(sessionId)
		gs.delSession(sId, sessionId)
		return
	}

	if session == nil {
		session = network.NewSession(conn, sessionId, true)
		if session == nil {
			return
		}
		network.AddSession(session)
		gs.addClientSession(sId, sessionId)
	}

	session.OnService(packet)
}

func (gs *LogicServer) addClientSession(sId, sessionId string) {
	//sessions, ok := gs.clientSessions[sId]
	//if !ok || sessions == nil {
	//	sessions := make(map[string]struct{})
	//	sessions[sessionId] = struct{}{}
	//	gs.clientSessions[sId] = sessions
	//} else {
	//	sessions[sessionId] = struct{}{}
	//}

	f := func() {
		s := &sync.Map{}
		s.Store(sessionId, struct{}{})
		gs.clientSessions.Store(sId, s)
	}

	iSessions, ok := gs.clientSessions.Load(sId)
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

func (gs *LogicServer) delClientSession(sId string) {
	//sessions, ok := gs.clientSessions[sId]
	//if !ok || sessions == nil {
	//	return
	//}
	//delete(gs.clientSessions, sId)
	//for sessionId := range sessions {
	//	session := network.GetSession(sessionId)
	//	if session == nil {
	//		continue
	//	}
	//	session.Close()
	//	network.DelSession(sessionId)
	//}

	iSessions, ok := gs.clientSessions.Load(sId)
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

func (gs *LogicServer) delSession(sId, sessionId string) {
	iSessions, ok := gs.clientSessions.Load(sId)
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

