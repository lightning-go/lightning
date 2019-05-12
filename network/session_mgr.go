/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package network

import (
	"sync"
	"github.com/lightning-go/lightning/defs"
)

var defaultSessionMgr = NewSessionMgr()

func GetSessionMgr() *SessionMgr {
	return defaultSessionMgr
}

func AddSession(s defs.ISession) {
	defaultSessionMgr.Add(s)
}

func GetSession(sessionId string) defs.ISession {
	return defaultSessionMgr.Get(sessionId)
}

func DelSession(sessionId string) {
	defaultSessionMgr.Del(sessionId)
}

func RangeSession(f func(sId string, s defs.ISession)) {
	defaultSessionMgr.Range(f)
}

type SessionMgr struct {
	mux      sync.RWMutex
	sessions map[string]defs.ISession
}

func NewSessionMgr() *SessionMgr {
	return &SessionMgr{
		sessions: make(map[string]defs.ISession),
	}
}

func (sm *SessionMgr) Count() int {
	sm.mux.RLock()
	count := len(sm.sessions)
	sm.mux.RUnlock()
	return count
}

func (sm *SessionMgr) Add(s defs.ISession) {
	if s == nil {
		return
	}
	sm.mux.Lock()
	sm.sessions[s.GetSessionId()] = s
	sm.mux.Unlock()
}

func (sm *SessionMgr) Del(sessionId string) {
	sm.mux.Lock()
	_, ok := sm.sessions[sessionId]
	if ok {
		delete(sm.sessions, sessionId)
	}
	sm.mux.Unlock()
}

func (sm *SessionMgr) Get(sessionId string) defs.ISession {
	sm.mux.RLock()
	session, ok := sm.sessions[sessionId]
	if !ok {
		sm.mux.RUnlock()
		return nil
	}
	sm.mux.RUnlock()
	return session
}

func (sm *SessionMgr) Range(f func(sId string, s defs.ISession)) {
	sm.mux.RLock()
	tmpSessions := sm.sessions
	sm.mux.RUnlock()
	for sessionId, session := range tmpSessions {
		f(sessionId, session)
	}
}
