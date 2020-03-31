/**
 * @author: Jason
 * Created: 19-5-3
 */

package utils

import "sync"

type SafeQueue struct {
	mux   sync.RWMutex
	cond  *sync.Cond
	stop  bool
	queue []interface{}
}

func NewSafeQueue() *SafeQueue {
	q := &SafeQueue{
		stop: false,
	}
	q.cond = sync.NewCond(&q.mux)
	return q
}

func (sq *SafeQueue) Put(v interface{}) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	if sq.stop {
		return
	}
	sq.queue = append(sq.queue, v)
	sq.cond.Signal()
}

func (sq *SafeQueue) Take(newList *[]interface{}) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	if sq.stop {
		return
	}

	for len(sq.queue) == 0 {
		sq.cond.Wait()
	}

	*newList = make([]interface{}, len(sq.queue))
	copy(*newList, sq.queue)

	sq.queue = sq.queue[0:0]
}

func (sq *SafeQueue) Stop() {
	sq.mux.Lock()
	sq.stop = true
	sq.mux.Unlock()
	sq.cond.Broadcast()
}

func (sq *SafeQueue) Size() int {
	sq.mux.RLock()
	queueLen := len(sq.queue)
	sq.mux.RUnlock()
	return queueLen
}

func (sq *SafeQueue) Empty() bool {
	sq.mux.RLock()
	empty := len(sq.queue) == 0
	sq.mux.RUnlock()
	return empty
}
