/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package network

import (
	"github.com/lightning-go/lightning/defs"
	"github.com/lightning-go/lightning/logger"
	"sync"
	"sync/atomic"
	"runtime/debug"
	"github.com/lightning-go/lightning/conf"
)

type queueData struct {
	session defs.ISession
	packet  defs.IPacket
}

var defaultSessionData = sync.Pool{
	New: func() interface{} {
		return &queueData{}
	},
}

func newSessionQueueData() *queueData {
	return defaultSessionData.Get().(*queueData)
}

func freeSessionQueueData(v *queueData) {
	if v != nil {
		defaultSessionData.Put(v)
	}
}

type Session struct {
	id   string
	conn defs.IConnection
	//packetData   chan defs.IPacket
	isAsync      bool
	serve        defs.ServeObj
	packet       defs.IPacket
	queue        chan *queueData
	queueWorking int32
	queueWait    sync.WaitGroup
	closed       int32
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
		id:           sessionId,
		conn:         conn,
		isAsync:      isAsync,
		serve:        serve,
		queueWorking: 0,
	}
	atomic.StoreInt32(&s.closed, 0)

	return s
}

func (s *Session) isSessionClosed() bool {
	v := atomic.LoadInt32(&s.closed)
	return v > 0
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
	if s.queue != nil {
		close(s.queue)
	}
	atomic.StoreInt32(&s.closed, 1)
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

func (s *Session) WritePacketAwait(packet defs.IPacket) (defs.IPacket, error) {
	return s.conn.WritePacketAwait(packet)
}

func (s *Session) WriteDataAwait(data []byte) (defs.IPacket, error) {
	return s.conn.WriteDataAwait(data)
}

func (s *Session) WriteDataByIdAwait(id string, data []byte) (defs.IPacket, error) {
	return s.conn.WriteDataByIdAwait(id, data)
}

func (s *Session) enableReadQueue() {
	if s.serve == nil {
		logger.Warn("server is nil")
		return
	}

	s.queue = make(chan *queueData, conf.GetGlobalVal().MaxQueueSize)

	go func() {
		atomic.StoreInt32(&s.queueWorking, 1)
		s.queueWait.Done()

		defer func() {
			atomic.StoreInt32(&s.queueWorking, 0)
			err := recover()
			if err != nil {
				logger.Error(err)
				logger.Error(string(debug.Stack()))
			}
		}()

		for d := range s.queue {
			if s.isSessionClosed() {
				break
			}
			if d == nil {
				continue
			}
			s.serve.OnServiceHandle(d.session, d.packet)
			freeSessionQueueData(d)
		}
		logger.Tracef("session closed %v", s.id)
	}()
}

func (s *Session) OnService(session defs.ISession, packet defs.IPacket) bool {
	if s.isSessionClosed() {
		return false
	}
	if s.serve == nil {
		logger.Warn("server is nil")
		return false
	}
	if session == nil || packet == nil {
		logger.Warn("session or packet is nil")
		return false
	}
	if s.isAsync {
		v := atomic.LoadInt32(&s.queueWorking)
		if v == 0 {
			s.queueWait.Add(1)
			s.enableReadQueue()
			s.queueWait.Wait()
		}

		d := newSessionQueueData()
		d.session = session
		d.packet = packet

		select {
		case s.queue <- d:
		}
		return true
	}
	return s.serve.OnServiceHandle(session, packet)
}
