/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package network

import (
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/defs"
	"fmt"
	"sync"
)

type Server struct {
	*TcpServer
	remotes         *sync.Map
	cfg             *conf.ServerConfig
	service         *utils.ServiceFactory
	connMgr         *SessionMgr
	sessionMgr      *SessionMgr
	newConnCallback defs.ConnCallback
	disConnCallback defs.ConnCallback
}

func NewServer(name, confPath string) *Server {
	conf.InitCfg(confPath)
	cfg := conf.GetServer(name)
	if cfg == nil {
		panic(fmt.Sprintf("%v config load failed", name))
		return nil
	}

	addr := fmt.Sprintf(":%v", cfg.Port)
	s := &Server{
		TcpServer:  NewTcpServer(addr, name, cfg.MaxConn),
		remotes:    &sync.Map{},
		cfg:        cfg,
		service:    utils.NewServiceFactory(),
		connMgr:    NewSessionMgr(),
		sessionMgr: NewSessionMgr(),
	}
	s.init()

	for _, remoteName := range cfg.Remotes {
		rCfg := conf.GetServer(remoteName)
		if rCfg == nil {
			continue
		}
		s.AddRemoteClient(rCfg)
	}

	return s
}

func (s *Server) init() {
	s.SetConnCallback(s.onConn)
}

func (s *Server) AddRemoteClient(cfg *conf.ServerConfig) *TcpClient {
	if cfg == nil {
		return nil
	}

	addr := fmt.Sprintf("%v:%v", cfg.Host, cfg.Port)
	c := NewTcpClient(cfg.Name, addr)
	if c == nil {
		return nil
	}
	s.remotes.Store(cfg.Name, c)
	return c
}

func (s *Server) GetRemoteClient(name string) *TcpClient {
	c, ok := s.remotes.Load(name)
	if !ok {
		return nil
	}
	client, ok := c.(*TcpClient)
	if !ok {
		return nil
	}
	return client
}

func (s *Server) Start() {
	s.Serve()

	s.remotes.Range(func(k, v interface{}) bool {
		client, ok := v.(*TcpClient)
		if !ok {
			return true
		}
		client.Connect()
		return true
	})

	logger.Infof("%v is running", s.name)
}

func (s *Server) GetCfg() *conf.ServerConfig {
	return s.cfg
}

func (s *Server) Host() string {
	return fmt.Sprintf("%v:%v", s.cfg.Host, s.cfg.Port)
}

func (s *Server) SetNewConnCallback(cb defs.ConnCallback) {
	s.newConnCallback = cb
}

func (s *Server) SetDisConnCallback(cb defs.ConnCallback) {
	s.disConnCallback = cb
}

func (s *Server) RegisterService(rcvr interface{}, cb ...defs.ParseMethodNameCallback) {
	s.service.Register(rcvr, cb...)
}

func (s *Server) OnServiceHandle(session defs.ISession, packet defs.IPacket) bool {
	return s.service.OnServiceHandle(session, packet)
}

func (s *Server) onConn(conn defs.IConnection) {
	isClosed := conn.IsClosed()
	logger.Tracef("%v server %v <- %v is %v",
		s.Name(), conn.LocalAddr(), conn.RemoteAddr(),
		utils.IF(isClosed, "down", "up").(string))

	if isClosed {
		s.OnDisConn(conn)
	} else {
		s.OnNewConn(conn)
	}
}

func (s *Server) OnNewConn(conn defs.IConnection) {
	session := NewSession(conn, conn.GetId(), s, false)
	s.connMgr.AddSession(session)
	if s.newConnCallback != nil {
		s.newConnCallback(conn)
	}
}

func (s *Server) OnDisConn(conn defs.IConnection) {
	if s.disConnCallback != nil {
		s.disConnCallback(conn)
	}
	connId := conn.GetId()
	s.connMgr.DelSession(connId)
	s.sessionMgr.DelConnSession(connId)
}

func (s *Server) GetConn(connId string) *Session {
	return s.connMgr.GetSession(connId)
}

func (s *Server) RangeConn(f func(string, *Session)) {
	s.connMgr.RangeSession(f)
}

func (s *Server) GetSession(sessionId string) *Session {
	return s.sessionMgr.GetSession(sessionId)
}

func (s *Server) DelSession(sessionId string) {
	s.sessionMgr.DelSession(sessionId)
}

func (s *Server) RangeSession(f func(string, *Session)) {
	s.sessionMgr.RangeSession(f)
}

func (s *Server) GetSessionCount() int64 {
	return s.sessionMgr.SessionCount()
}

func (s *Server) CheckAddSession(conn defs.IConnection, sessionId string, async ...bool) *Session {
	session := s.sessionMgr.GetSession(sessionId)
	if session == nil {
		session = NewSession(conn, sessionId, s, async...)
		if session == nil {
			return nil
		}
		s.sessionMgr.AddSession(session)
	}
	return session
}

func (s *Server) OnSessionService(conn defs.IConnection, packet defs.IPacket) {
	sessionId := packet.GetSessionId()
	session := s.CheckAddSession(conn, sessionId, true)
	if session == nil {
		logger.Errorf("session nil %v", sessionId)
		return
	}
	session.OnService(packet)
}
