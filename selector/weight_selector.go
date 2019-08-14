/**
 * Created: 2019/5/7 0007
 * @author: Jason
 */

package selector

import (
	"sync"
)

type SessionData struct {
	//Id     int              `json:"id"`
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

func (selector *WeightSelector) IsNew(data *SessionData) bool {
	if data == nil {
		return false
	}

	selector.mux.Lock()
	for _, sd := range selector.sessions {
		if sd == nil {
			continue
		}
		//if sd.Id == data.Id {
		if sd.Name == data.Name {
			//sd.Host = data.Host
			//sd.Name = data.Name
			//sd.Type = data.Type
			//sd.Weight = data.Weight
			selector.mux.Unlock()
			return false
		}
	}
	selector.mux.Unlock()

	return true
}

func (selector *WeightSelector) Add(data *SessionData) {
	if data == nil {
		return
	}
	selector.mux.Lock()
	selector.sessions = append(selector.sessions, data)
	selector.mux.Unlock()
	return
}

func (selector *WeightSelector) AddCheck(data *SessionData) (new bool) {
	if data == nil {
		return false
	}

	selector.mux.Lock()
	for _, sd := range selector.sessions {
		if sd == nil {
			continue
		}
		//if sd.Id == data.Id {
		if sd.Name == data.Name {
			sd.Host = data.Host
			sd.Name = data.Name
			sd.Type = data.Type
			sd.Weight = data.Weight
			selector.mux.Unlock()
			return false
		}
	}

	selector.sessions = append(selector.sessions, data)
	selector.mux.Unlock()
	return true
}

//func (selector *WeightSelector) Del(key int) {
func (selector *WeightSelector) Del(key string) {
	selector.mux.Lock()

	count := len(selector.sessions)
	if count == 1 {
		session := selector.sessions[0]
		if session != nil && session.Name == key {
			selector.sessions = selector.sessions[:0]
			selector.mux.Unlock()
			return
		}
	}

	delIdx := -1
	for idx, session := range selector.sessions {
		if session == nil {
			continue
		}
		//if session.Id == key {
		if session.Name == key {
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

//func (selector *WeightSelector) Update(key int, weight int) {
func (selector *WeightSelector) Update(key string, weight int) {
	selector.mux.Lock()
	for _, sd := range selector.sessions {
		if sd == nil {
			continue
		}
		if sd.Name == key {
			sd.Weight = weight
			selector.mux.Unlock()
			return
		}
	}
	selector.mux.Unlock()
}

func (selector *WeightSelector) AddWeight(key string, weight int) {
	selector.mux.Lock()
	for _, sd := range selector.sessions {
		if sd == nil {
			continue
		}
		if sd.Name == key {
			sd.Weight += weight
			selector.mux.Unlock()
			return
		}
	}
	selector.mux.Unlock()
}

func (selector *WeightSelector) SubWeight(key string, weight int) {
	selector.mux.Lock()
	for _, sd := range selector.sessions {
		if sd == nil {
			continue
		}
		if sd.Name == key {
			sd.Weight -= weight
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
	} else if sessionCount == 1 {
		return sessionList[0]
	}

	start := (selector.lastIdx + 1) % sessionCount
	minLoadIdx := start
	minLoad := sessionList[start].Weight

	for i := 0; i < sessionCount; i++ {
		idx := (start + i) % sessionCount
		s := sessionList[idx]
		if s == nil {
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

func (selector *WeightSelector) SelectRoundWeightLeast() *SessionData {
	selector.mux.Lock()
	defer selector.mux.Unlock()

	sessionList := selector.sessions
	sessionCount := len(sessionList)
	if sessionCount == 0 {
		return nil
	} else if sessionCount == 1 {
		return sessionList[0]
	}

	start := (selector.lastIdx + 1) % sessionCount
	minLoadIdx := start

	minLoad := sessionList[start].Weight

	for i := 0; i < sessionCount; i++ {
		s := sessionList[i]
		if s == nil {
			continue
		}

		if s.Weight < minLoad {
			minLoad = s.Weight
			minLoadIdx = i
			if minLoad == 0 {
				break
			}
		}
	}

	selector.lastIdx = minLoadIdx
	session := sessionList[minLoadIdx]
	return session
}

