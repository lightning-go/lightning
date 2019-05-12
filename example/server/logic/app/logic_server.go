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
	"github.com/lightning-go/lightning/example/server/global"
)

type LogicServer struct {
	*network.Server
	clientSessions *sync.Map //map[string]map[string]struct{}
}

func NewGameServer(name, path string) *LogicServer {
	ls := &LogicServer{
		Server:         network.NewServer(name, path),
		clientSessions: &sync.Map{}, //make(map[string]map[string]struct{}),
	}
	ls.init()
	ls.initRemote()
	ls.initService()
	return ls
}

func (ls *LogicServer) init() {
	ls.SetCodec(&module.HeadCodec{})
	ls.SetConnCallback(ls.onConn)
	ls.SetMsgCallback(ls.onMsg)
	ls.SetAuthorizedCallback(ls.onAuthorized)
}

func (ls *LogicServer) initService() {
	utils.RegisterService(&service.Service{})
}

func (ls *LogicServer) onConn(conn defs.IConnection) {
	isClosed := conn.IsClosed()
	logger.Tracef("%v server %v <- %v is %v",
		ls.Name(), conn.LocalAddr(), conn.RemoteAddr(),
		utils.IF(isClosed, "down", "up").(string))

	if isClosed {
		ls.onDisconn(conn)
	} else {
		ls.onNewConn(conn)
	}
}

func (ls *LogicServer) onNewConn(conn defs.IConnection) {
}

func (ls *LogicServer) onDisconn(conn defs.IConnection) {
	ls.delClientSession(conn.GetId())
}

func (ls *LogicServer) onAuthorized(conn defs.IConnection, packet defs.IPacket) bool {
	ok, _ := service.AuthorizedCallback(conn, packet)
	return ok
}

func (ls *LogicServer) onMsg(conn defs.IConnection, packet defs.IPacket) {
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

	if status == global.RESULT_DISCONN {
		network.DelSession(sessionId)
		ls.delSession(sId, sessionId)
		return
	}

	if session == nil {
		session = network.NewSession(conn, sessionId, true)
		if session == nil {
			return
		}
		network.AddSession(session)
		ls.addClientSession(sId, sessionId)
	}

	session.OnService(packet)
}

func (ls *LogicServer) addClientSession(sId, sessionId string) {
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
		ls.clientSessions.Store(sId, s)
	}

	iSessions, ok := ls.clientSessions.Load(sId)
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

func (ls *LogicServer) delClientSession(sId string) {
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

	iSessions, ok := ls.clientSessions.Load(sId)
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

func (ls *LogicServer) delSession(sId, sessionId string) {
	iSessions, ok := ls.clientSessions.Load(sId)
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

