/**
 * Created: 2019/4/22 0022
 * @author: Jason
 */

package network

import (
	"github.com/lightning-go/lightning/defs"
	"sync"
	"github.com/lightning-go/lightning/utils"
)

var srvMgr *ServerMgr
var srvMgrOnce sync.Once

func GetSrvMgr() *ServerMgr {
	srvMgrOnce.Do(func() {
		srvMgr = newServerMgr()
	})
	return srvMgr
}

func WaitExit() {
	utils.WaitSignal()
	GetSrvMgr().AllStop()
}

type ServerMgr struct {
	servers map[string]defs.IServer
}

func newServerMgr() *ServerMgr {
	return &ServerMgr{
		servers: make(map[string]defs.IServer),
	}
}

func (sm *ServerMgr) AddServer(server defs.IServer) {
	if server == nil {
		return
	}
	sm.servers[server.Name()] = server
}

func (sm *ServerMgr) DelServerByName(name string) {
	srv, ok := sm.servers[name]
	if ok {
		delete(sm.servers, name)
		if srv != nil {
			srv.Stop()
		}
	}
}

func (sm *ServerMgr) DelServer(server defs.IServer) {
	if server == nil {
		return
	}
	delete(sm.servers, server.Name())
	server.Stop()
}

func (sm *ServerMgr) AllDelete() {
	for _, srv := range sm.servers {
		sm.DelServer(srv)
	}
}

func (sm *ServerMgr) StopServerByName(name string) {
	srv, ok := sm.servers[name]
	if ok && srv != nil {
		srv.Stop()
	}
}

func (sm *ServerMgr) StopServer(server defs.IServer) {
	if server == nil {
		return
	}
	server.Stop()
}

func (sm *ServerMgr) AllStop() {
	for _, srv := range sm.servers {
		if srv == nil {
			continue
		}
		srv.Stop()
	}
}
