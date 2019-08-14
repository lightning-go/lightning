/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package network

import (
	"sync"
)

var defaultSessionMgr = NewSessionMgr()

func GetSessionMgr() *SessionMgr {
	return defaultSessionMgr
}

//func AddSession(s defs.ISession) {
func AddSession(s *Session) {
	defaultSessionMgr.Add(s)
}

//func GetSession(sessionId string) defs.ISession {
func GetSession(sessionId string) *Session {
	return defaultSessionMgr.Get(sessionId)
}

func DelSession(sessionId string) {
	defaultSessionMgr.Del(sessionId)
}

//func RangeSession(f func(sId string, s defs.ISession)) {
func RangeSession(f func(sId string, s *Session)) {
	defaultSessionMgr.Range(f)
}

type SessionMgr struct {
	mux      sync.RWMutex
	//sessions map[string]defs.ISession
	sessions map[string]*Session
}

func NewSessionMgr() *SessionMgr {
	return &SessionMgr{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionMgr) Count() int {
	sm.mux.RLock()
	count := len(sm.sessions)
	sm.mux.RUnlock()
	return count
}

//func (sm *SessionMgr) Add(s defs.ISession) {
func (sm *SessionMgr) Add(s *Session) {
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

//func (sm *SessionMgr) Get(sessionId string) defs.ISession {
func (sm *SessionMgr) Get(sessionId string) *Session {
	sm.mux.RLock()
	session, ok := sm.sessions[sessionId]
	if !ok {
		sm.mux.RUnlock()
		return nil
	}
	sm.mux.RUnlock()
	return session
}

//func (sm *SessionMgr) Range(f func(sId string, s defs.ISession)) {
func (sm *SessionMgr) Range(f func(sId string, s *Session)) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	for sessionId, session := range sm.sessions {
		f(sessionId, session)
	}
}
