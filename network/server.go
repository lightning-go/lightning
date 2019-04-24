/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package network

import (
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/logger"
	"fmt"
	"sync"
)

type Server struct {
	*TcpServer
	remotes *sync.Map
	cfg     *conf.ServerConfig
}

func NewServer(name, confPath string) *Server {
	conf.InitCfg(confPath)
	cfg := conf.Get(name)
	if cfg == nil {
		logger.Error("%v config load failed", name)
		return nil
	}

	addr := fmt.Sprintf(":%v", cfg.Port)
	s := &Server{
		TcpServer: NewTcpServer(addr, name, cfg.MaxConn),
		remotes:   &sync.Map{},
		cfg:       cfg,
	}

	for _, remoteName := range cfg.Remotes {
		rCfg := conf.Get(remoteName)
		if rCfg == nil {
			continue
		}
		s.AddRemoteClient(rCfg)
	}

	return s
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

	logger.Info("%v is running...", s.name)
}
