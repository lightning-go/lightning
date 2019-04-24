/**
 * Created: 2019/4/19 0019
 * @author: Jason
 */

package network

import (
	"sync"
	"github.com/lightning-go/lightning/defs"
)

type ConnectionMgr struct {
	mux   sync.RWMutex
	conns map[string]defs.IConnection
}

func NewConnMgr() *ConnectionMgr {
	return &ConnectionMgr{
		conns: make(map[string]defs.IConnection),
	}
}

func (cm *ConnectionMgr) ConnCount() int {
	cm.mux.RLock()
	count := len(cm.conns)
	cm.mux.RUnlock()
	return count
}

func (cm *ConnectionMgr) AddConn(conn defs.IConnection) {
	cm.mux.Lock()
	cm.conns[conn.GetId()] = conn
	cm.mux.Unlock()
}

func (cm *ConnectionMgr) DelConn(connId string) {
	cm.mux.Lock()
	_, ok := cm.conns[connId]
	if ok {
		delete(cm.conns, connId)
	}
	cm.mux.Unlock()
}

func (cm *ConnectionMgr) GetConn(connId string) defs.IConnection {
	cm.mux.RLock()
	conn, ok := cm.conns[connId]
	if !ok {
		cm.mux.RUnlock()
		return nil
	}
	cm.mux.RUnlock()
	return conn
}

func (cm *ConnectionMgr) Clean() {
	cm.mux.Lock()
	for _, conn := range cm.conns {
		if conn == nil {
			continue
		}
		conn.Close()
		delete(cm.conns, conn.GetId())
	}
	cm.mux.Unlock()
}
