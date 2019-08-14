/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package network

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/utils"
	"github.com/lightning-go/lightning/conf"
)

type Session struct {
	id         string
	conn       defs.IConnection
	packetData chan defs.IPacket
	isAsync    bool
}

func NewSession(conn defs.IConnection, sessionId string, async ...bool) *Session {
	isAsync := false
	if len(async) > 0 {
		isAsync = async[0]
	}

	s := &Session{
		id:      sessionId,
		conn:    conn,
		isAsync: isAsync,
	}

	if isAsync {
		s.enableReadQueue()
	}

	return s
}

func (s *Session) Close() bool {
	if s.packetData != nil {
		close(s.packetData)
	}
	return s.conn.Close()
}

func (s *Session) SetContext(key, value interface{}) {
	s.conn.SetContext(key, value)
}

func (s *Session) GetContext(key interface{}) interface{} {
	return s.conn.GetContext(key)
}

func (s *Session) GetConn() defs.IConnection {
	return s.conn
}

func (s *Session) GetSessionId() string {
	return s.id
}

func (s *Session) WritePacket(packet defs.IPacket) {
	s.conn.WritePacket(packet)
}

func (s *Session) WriteData(data []byte) {
	s.conn.WriteData(data)
}

func (s *Session) enableReadQueue() {
	s.packetData = make(chan defs.IPacket, conf.GetGlobalVal().MaxQueueSize)
	go func() {
		for packet := range s.packetData {
			if packet == nil {
				continue
			}
			utils.OnServiceHandle(s, packet)
		}
	}()
}

func (s *Session) OnService(packet defs.IPacket) bool {
	if packet == nil {
		return false
	}
	if s.isAsync {
		select {
		case s.packetData <- packet:
		}
		return true
	}
	return utils.OnServiceHandle(s, packet)
}
