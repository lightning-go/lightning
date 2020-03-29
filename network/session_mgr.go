/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package network

import (
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

func AddSession(s *Session) {
	defaultSessionMgr.AddSession(s)
}

func GetSession(sessionId string) *Session {
	return defaultSessionMgr.GetSession(sessionId)
}

func DelSession(sessionId string) {
	defaultSessionMgr.DelSession(sessionId)
}

func RangeSession(f func(sId string, s *Session)) {
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

func (sm *SessionMgr) AddSession(s *Session) {
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

func (sm *SessionMgr) GetSession(sessionId string) *Session {
	v, ok := sm.sessions.Get(sessionId)
	if !ok {
		return nil
	}
	s, ok := v.(*Session)
	if !ok {
		return nil
	}
	return s
}

func (sm *SessionMgr) DelSession(sessionId string) {
	session := sm.GetSession(sessionId)
	if session == nil {
		return
	}
	session.CloseSession()

	connId := session.GetConnId()
	sm.sessions.Del(sessionId)

	d := sm.getConnSession(connId)
	if d != nil {
		d.Delete(sessionId)
	}
}

func (sm *SessionMgr) DelConnSession(connId string) {
	d := sm.getConnSession(connId)
	if d == nil {
		return
	}

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
		return true
	})
	sm.connDict.Del(connId)
}

func (sm *SessionMgr) RangeSession(f func(string, *Session)) {
	if f == nil {
		return
	}
	sm.sessions.Range(func(k, v interface{}) bool {
		session, ok := v.(*Session)
		if !ok {
			return true
		}
		f(session.GetSessionId(), session)
		return true
	})
}
