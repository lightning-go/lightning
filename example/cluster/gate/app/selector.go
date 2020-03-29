/**
 * Created: 2020/3/25
 * @author: Jason
 */

package app

import (
	"github.com/lightning-go/lightning/selector"
	"sync"
	"github.com/lightning-go/lightning/logger"
)

type ServeSelector struct {
	*selector.WeightSelector
	remoteClientMap      sync.Map //all remote
	remoteSession        sync.Map //remote mapping of sessionId
	sessionIdMap         sync.Map //sessionId list mapping of remote
	cleanSessionCallback func(string)
}

func NewSelector() *ServeSelector {
	return &ServeSelector{
		WeightSelector: selector.NewWeightSelector(),
	}
}

func (ss *ServeSelector) SetCleanSessionCallback(cb func(string)) {
	ss.cleanSessionCallback = cb
}

func (ss *ServeSelector) AddRemoteSession(sessionId string, session *RemoteClient) {
	if session == nil {
		return
	}
	ss.remoteSession.Store(sessionId, session)
	ss.addSessionIdMap(sessionId, session)
}

func (ss *ServeSelector) GetRemoteSession(sessionId string) *RemoteClient {
	d, ok := ss.remoteSession.Load(sessionId)
	if !ok {
		return nil
	}
	session, ok := d.(*RemoteClient)
	if !ok {
		return nil
	}
	return session
}

func (ss *ServeSelector) DelRemoteSession(sessionId string) {
	ss.remoteSession.Delete(sessionId)
}

func (ss *ServeSelector) AddRemoteData(sd *selector.SessionData, cb func(*selector.SessionData) bool) {
	if sd == nil {
		return
	}
	logger.Debugf("put - name: %v, host: %v, type: %v, weight: %v",
		sd.Name, sd.Host, sd.Type, sd.Weight)

	isNew, changed := ss.IsNew(sd)
	if isNew {
		if cb != nil {
			cb(sd)
		}
	} else if changed {
		ss.DelRemoteData(sd.Name)
	}
}

func (ss *ServeSelector) GetRemoteData() *selector.SessionData {
	return ss.SelectRoundWeightLeast()
}

func (ss *ServeSelector) DelRemoteData(key string) {
	rc := ss.GetRemoteClient(key)
	if rc == nil {
		return
	}
	ss.DelRemoteClient(rc)
}

func (ss *ServeSelector) AddRemoteClient(rc *RemoteClient) {
	if rc == nil {
		return
	}
	ss.remoteClientMap.Store(rc.sd.Name, rc)
	ss.Add(rc.sd)
}

func (ss *ServeSelector) GetRemoteClient(key string) *RemoteClient {
	d, ok := ss.remoteClientMap.Load(key)
	if !ok {
		return nil
	}
	conn, ok := d.(*RemoteClient)
	if !ok {
		return nil
	}
	return conn
}

func (ss *ServeSelector) DelRemoteClient(rc *RemoteClient) {
	if rc == nil {
		return
	}
	ss.Del(rc.sd.Name)
	ss.remoteClientMap.Delete(rc.sd.Name)
	ss.cleanSessionIdMap(rc)
}

func (ss *ServeSelector) addSessionIdMap(sessionId string, rc *RemoteClient) {
	if rc == nil {
		return
	}
	conn := rc.GetConn()
	if conn == nil {
		return
	}
	connId := conn.GetId()

	d, ok := ss.sessionIdMap.Load(connId)
	if !ok {
		sessionDict := &sync.Map{}
		sessionDict.Store(sessionId, struct {}{})
		ss.sessionIdMap.Store(connId, sessionDict)
	} else {
		sessionDict, ok := d.(*sync.Map)
		if !ok {
			sessionDict := &sync.Map{}
			sessionDict.Store(sessionId, struct {}{})
			ss.sessionIdMap.Store(connId, sessionDict)
		} else {
			sessionDict.Store(sessionId, struct{}{})
		}
	}
}

func (ss *ServeSelector) delSessionIdMap(sessionId string, rc *RemoteClient) {
	if rc == nil {
		return
	}
	conn := rc.GetConn()
	if conn == nil {
		return
	}
	connId := conn.GetId()
	d, ok := ss.sessionIdMap.Load(connId)
	if !ok {
		return
	}
	sessionDict, ok := d.(*sync.Map)
	if !ok {
		return
	}
	sessionDict.Delete(sessionId)
}

func (ss *ServeSelector) cleanSessionIdMap(rc *RemoteClient) {
	if rc == nil {
		return
	}
	conn := rc.GetConn()
	if conn == nil {
		return
	}

	connId := conn.GetId()
	d, ok := ss.sessionIdMap.Load(connId)
	if !ok {
		return
	}
	sessionDict, ok := d.(*sync.Map)
	if !ok {
		return
	}

	sessionDict.Range(func(key, value interface{}) bool {
		sessionId, ok := key.(string)
		if !ok {
			return true
		}
		if ss.cleanSessionCallback != nil {
			ss.cleanSessionCallback(sessionId)
		}
		ss.DelRemoteSession(sessionId)
		return true
	})

	ss.sessionIdMap.Delete(connId)
}
