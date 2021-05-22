/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package network

import (
	"github.com/lightning-go/lightning/defs"
	"sync"
	"sync/atomic"
)

type Map struct {
	dict    sync.Map
	counter int64
}

func (m *Map) Length() int64 {
	count := atomic.LoadInt64(&m.counter)
	return count
}

func (m *Map) Get(key interface{}) (interface{}, bool) {
	return m.dict.Load(key)
}

func (m *Map) Add(key, value interface{}) {
	m.dict.Store(key, value)
	atomic.AddInt64(&m.counter, 1)
}

func (m *Map) Del(key interface{}) {
	m.dict.Delete(key)
	atomic.AddInt64(&m.counter, -1)
}

func (m *Map) Range(f func(k, v interface{}) bool) {
	m.dict.Range(func(key, value interface{}) bool {
		return f(key, value)
	})
}

////////////////////////////////////////////////////////////////

var defaultSessionMgr = NewSessionMgr()

func GetSessionMgr() *SessionMgr {
	return defaultSessionMgr
}

func AddSession(s defs.ISession) {
	defaultSessionMgr.AddSession(s)
}

func GetSession(sessionId string) defs.ISession {
	return defaultSessionMgr.GetSession(sessionId)
}

func DelSession(sessionId string) {
	defaultSessionMgr.DelSession(sessionId)
}

func RangeSession(f func(sId string, s defs.ISession) bool) {
	defaultSessionMgr.RangeSession(f)
}

////////////////////////////////////////////////////////////////

type SessionMgr struct {
	sessions *Map
	connDict *Map
}

func NewSessionMgr() *SessionMgr {
	return &SessionMgr{
		sessions: &Map{},
		connDict: &Map{},
	}
}

func (sm *SessionMgr) SessionCount() int64 {
	return sm.sessions.Length()
}

func (sm *SessionMgr) AddSession(s defs.ISession) {
	if s == nil {
		return
	}
	sessionId := s.GetSessionId()
	sm.sessions.Add(sessionId, s)

	connId := s.GetConnId()
	if sessionId != connId {
		d := sm.getConnSession(connId)
		if d == nil {
			d := &sync.Map{}
			d.Store(sessionId, struct{}{})
			sm.connDict.Add(connId, d)
		} else {
			d.Store(sessionId, struct{}{})
		}
	}
}

func (sm *SessionMgr) getConnSession(connId string) *sync.Map {
	d, ok := sm.connDict.Get(connId)
	if !ok {
		return nil
	}
	d2, ok := d.(*sync.Map)
	if !ok {
		return nil
	}
	return d2
}

func (sm *SessionMgr) GetSession(sessionId string) defs.ISession {
	v, ok := sm.sessions.Get(sessionId)
	if !ok {
		return nil
	}
	s, ok := v.(defs.ISession)
	if !ok {
		return nil
	}
	return s
}

func (sm *SessionMgr) DelSession(sessionId string) defs.ISession {
	session := sm.GetSession(sessionId)
	if session == nil {
		return nil
	}
	session.CloseSession()

	connId := session.GetConnId()
	sm.sessions.Del(sessionId)

	d := sm.getConnSession(connId)
	if d != nil {
		d.Delete(sessionId)
	}
	return session
}

func (sm *SessionMgr) DelConnSession(connId string) []defs.ISession {
	d := sm.getConnSession(connId)
	if d == nil {
		return nil
	}

	delSessions := make([]defs.ISession, 0)
	d.Range(func(key, value interface{}) bool {
		sessionId, ok := key.(string)
		if !ok {
			return true
		}

		session := sm.GetSession(sessionId)
		if session != nil {
			session.CloseSession()
		}

		sm.sessions.Del(sessionId)
		delSessions = append(delSessions, session)
		return true
	})
	sm.connDict.Del(connId)
	return delSessions
}

func (sm *SessionMgr) RangeSession(f func(string, defs.ISession) bool) {
	if f == nil {
		return
	}
	sm.sessions.Range(func(k, v interface{}) bool {
		session, ok := v.(defs.ISession)
		if !ok {
			return true
		}
		return f(session.GetSessionId(), session)
	})
}
