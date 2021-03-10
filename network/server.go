/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package network

import (
	"fmt"
	"sync"

	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"github.com/lightning-go/lightning/utils"
)

type Server struct {
	*TcpServer
	remotes         *sync.Map
	cfg             *conf.ServerConfig
	service         *utils.ServiceFactory
	connMgr         *SessionMgr
	newConnCallback defs.ConnCallback
	disConnCallback defs.ConnCallback
}

func NewServer(name, confPath string) *Server {
	if len(confPath) > 0 {
		conf.InitCfg(confPath)
	}
	cfg := conf.GetServer(name)
	if cfg == nil {
		panic(fmt.Sprintf("%v config load failed", name))
	}

	addr := fmt.Sprintf(":%v", cfg.Port)
	s := &Server{
		TcpServer: NewTcpServer(addr, name, cfg.MaxConn),
		remotes:   &sync.Map{},
		cfg:       cfg,
		service:   utils.NewServiceFactory(),
		connMgr:   NewSessionMgr(),
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
	session := NewSession(conn, conn.GetId(), s.OnServiceHandle, true)
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
}

func (s *Server) GetConn(connId string) defs.ISession {
	return s.connMgr.GetSession(connId)
}

func (s *Server) RangeConn(f func(string, defs.ISession) bool) {
	s.connMgr.RangeSession(f)
}

func (s *Server) GetConnNum() int64 {
	return s.connMgr.SessionCount()
}
