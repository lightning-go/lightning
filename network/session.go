/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package network

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/conf"
	"github.com/lightning-go/lightning/logger"
)

type Session struct {
	id         string
	conn       defs.IConnection
	packetData chan defs.IPacket
	isAsync    bool
	serve      defs.ServeObj
	packet     defs.IPacket
}

func NewSession(conn defs.IConnection, sessionId string, serve defs.ServeObj, async ...bool) *Session {
	if conn == nil || serve == nil {
		return nil
	}
	isAsync := false
	if len(async) > 0 {
		isAsync = async[0]
	}

	s := &Session{
		id:      sessionId,
		conn:    conn,
		isAsync: isAsync,
		serve:   serve,
	}

	if isAsync {
		s.enableReadQueue()
	}

	return s
}

func (s *Session) GetPacket() defs.IPacket {
	return s.packet
}

func (s *Session) SetPacket(packet defs.IPacket) {
	s.packet = packet
}

func (s *Session) GetServeObj() defs.ServeObj {
	return s.serve
}

func (s *Session) Close() bool {
	s.CloseSession()
	return s.conn.Close()
}

func (s *Session) CloseSession() bool {
	if s.packetData != nil {
		close(s.packetData)
	}
	return true
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

func (s *Session) GetConnId() string {
	if s.conn == nil {
		return ""
	}
	return s.conn.GetId()
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

func (s *Session) WriteDataById(id string, data []byte) {
	s.conn.WriteDataById(id, data)
}

func (s *Session) enableReadQueue() {
	if s.serve == nil {
		logger.Warn("server is nil")
		return
	}
	s.packetData = make(chan defs.IPacket, conf.GetGlobalVal().MaxQueueSize)
	go func() {
		for packet := range s.packetData {
			if packet == nil {
				continue
			}
			s.serve.OnServiceHandle(s, packet)
		}
		logger.Tracef("session closed %v", s.id)
	}()
}

func (s *Session) OnService(packet defs.IPacket) bool {
	if s.serve == nil {
		logger.Warn("server is nil")
		return false
	}
	if packet == nil {
		return false
	}
	if s.isAsync {
		select {
		case s.packetData <- packet:
		}
		return true
	}
	return s.serve.OnServiceHandle(s, packet)
}
