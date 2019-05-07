/**
 * Created: 2019/5/7 0007
 * @author: Jason
 */

package selector

import (
	"github.com/lightning-go/lightning/defs"
	"sync"
)

type SessionData struct {
	conn   defs.IConnection `json:"-"`
	Host   string           `json:"host"`
	Name   string           `json:"name"`
	Type   int              `json:"type"`
	Weight int              `json:"weight"`
}

type WeightSelector struct {
	mux      sync.RWMutex
	sessions []*SessionData
	lastIdx  int
}

func NewWeightSelector() *WeightSelector {
	return &WeightSelector{
		sessions: make([]*SessionData, 0),
	}
}

func (selector *WeightSelector) Add(data *SessionData) {
	if data == nil || data.conn == nil {
		return
	}

	selector.mux.Lock()
	for _, sd := range selector.sessions {
		if sd == nil || sd.conn == nil {
			continue
		}
		if sd.conn.GetId() == data.conn.GetId() {
			sd.conn = data.conn
			sd.Host = data.Host
			sd.Name = data.Name
			sd.Type = data.Type
			sd.Weight = data.Weight
			selector.mux.Unlock()
			return
		}
	}

	selector.sessions = append(selector.sessions, data)
	selector.mux.Unlock()
}

func (selector *WeightSelector) Del(key string) {
	selector.mux.Lock()

	count := len(selector.sessions)
	if count == 1 {
		selector.sessions = selector.sessions[:0]
		selector.mux.Unlock()
		return
	}

	delIdx := -1
	for idx, session := range selector.sessions {
		if session == nil || session.conn == nil {
			continue
		}
		if session.conn.GetId() == key {
			delIdx = idx
			break
		}
	}
	if delIdx == -1 {
		selector.mux.Unlock()
		return
	}

	var s []*SessionData
	if delIdx == 0 {
		s = selector.sessions[1:]
	} else if delIdx == count-1 {
		s = selector.sessions[:delIdx]
	} else {
		s = selector.sessions[:delIdx]
		s = append(s, selector.sessions[delIdx+1:]...)
	}
	selector.sessions = s

	selector.mux.Unlock()
}

func (selector *WeightSelector) Update(key string, weight int) {
	selector.mux.Lock()
	for _, sd := range selector.sessions {
		if sd == nil || sd.conn == nil {
			continue
		}
		if sd.conn.GetId() == key {
			sd.Weight = weight
			selector.mux.Unlock()
			return
		}
	}
	selector.mux.Unlock()
}

func (selector *WeightSelector) SelectWeightLeast() *SessionData {
	selector.mux.Lock()
	defer selector.mux.Unlock()

	sessionList := selector.sessions
	sessionCount := len(sessionList)
	if sessionCount == 0 {
		return nil
	}

	start := (selector.lastIdx + 1) % sessionCount
	minLoadIdx := start
	minLoad := sessionList[start].Weight

	for i := 0; i < sessionCount; i++ {
		idx := (start + i) % sessionCount
		s := sessionList[idx]
		if s == nil || s.conn == nil {
			continue
		}

		load := sessionList[idx].Weight
		if load < minLoad {
			minLoad = load
			minLoadIdx = idx
			if minLoad == 0 {
				break
			}
		}
	}

	selector.lastIdx = minLoadIdx
	session := sessionList[minLoadIdx]
	return session
}
