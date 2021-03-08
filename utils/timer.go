/**
 * Created: 2019/4/24 0024
 * @author: Jason
 */

package utils

import (
	"container/heap"
	"github.com/lightning-go/lightning/logger"
	"runtime/debug"
	"sync"
	"time"
)

type Timer struct {
	tick    time.Duration
	quit    chan struct{}
	f       func()
	running bool
}

func NewTimer(tick time.Duration, f func()) *Timer {
	return &Timer{
		tick: tick,
		f:    f,
		quit: make(chan struct{}),
	}
}

func (t *Timer) Start(once ...bool) {
	if t.f == nil {
		return
	}
	t.running = true

	go func() {
		tick := time.NewTicker(t.tick)
		defer tick.Stop()
	QUIT:
		for {
			select {
			case <-tick.C:
				t.f()
				if len(once) > 0 && once[0] {
					break QUIT
				}
			case <-t.quit:
				break QUIT
			}
		}
	}()
}


func (t *Timer) Stop() {
	t.running = false
	close(t.quit)
}

func (t *Timer) IsRunning() bool {
	return t.running
}


///////////////////////////////////////////////////////////////////////


type timerEx struct {
	id 			uint64
	repeat 		bool
	interval 	time.Duration
	expireTime 	time.Duration
	callback 	func()
}

func (t *timerEx) Cancel() {
	t.callback = nil
}

func (t *timerEx) IsActive() bool {
	return t.callback != nil
}

type HeapTimer struct {
	pq				PriorityQueue
	initQueueSize 	int
	idGen 			*IdGenerator
	minMillInterval int64
	heapLock 		sync.Mutex
	quit 			chan struct{}
	running 		bool
}

func NewHeapTimer(millInterval int64) *HeapTimer {
	capacity := 64
	return &HeapTimer{
		pq: NewPriorityQueue(capacity),
		initQueueSize: capacity,
		idGen: NewIdGenerator(),
		minMillInterval: millInterval,
		quit: make(chan struct{}),
		running: false,
	}
}

func (ht *HeapTimer) AddTimer(f func(), millInterval int64, repeat ...bool) {
	if f == nil {
		return
	}

	isRepeat := true
	if len(repeat) > 0 {
		isRepeat = repeat[0]
	}
	if millInterval < ht.minMillInterval {
		millInterval = ht.minMillInterval
	}

	now := GetNowTimeMillisecond()
	t := &timerEx {
		id: ht.idGen.Get(),
		repeat: isRepeat,
		interval: time.Duration(millInterval),
		expireTime: time.Duration(now + millInterval),
		callback: f,
	}
	item := &Item{
		Value: t,
		Priority: int64(t.expireTime),
	}

	ht.heapLock.Lock()
	heap.Push(&ht.pq, item)
	ht.heapLock.Unlock()

}

func (ht *HeapTimer) run() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Error(string(debug.Stack()))
		}
	}()

	now := GetNowTimeMillisecond()
	for {
		ht.heapLock.Lock()
		item := ht.pq.Get(now)
		ht.heapLock.Unlock()
		if item == nil || item.Value == nil {
			break
		}

		t, ok := item.Value.(*timerEx)
		if !ok {
			continue
		}
		if !t.IsActive() {
			continue
		}
		t.callback()

		if t.repeat {
			t.id = ht.idGen.Get()
			t.expireTime += t.interval
			item.Value = t
			item.Priority = int64(t.expireTime)
			ht.heapLock.Lock()
			heap.Push(&ht.pq, item)
			ht.heapLock.Unlock()
		} else {
			t.Cancel()
		}
	}
}

func (ht *HeapTimer) IsRunning() bool {
	return ht.running
}

func (ht *HeapTimer) Stop() {
	ht.running = false
	close(ht.quit)
}

func (ht *HeapTimer) Run() {
	go func() {
		defer ht.Stop()
		ht.running = true
	QUIT:
		for {
			select {
			case <-ht.quit:
				break QUIT
			default:
				ht.run()
			}
			time.Sleep(time.Duration(ht.minMillInterval) * time.Millisecond)
		}
		logger.Trace("heap timer stopped")
	}()
}


