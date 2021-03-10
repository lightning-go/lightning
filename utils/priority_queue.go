package utils

import (
	"container/heap"
)

type Item struct {
	Value    interface{}
	Priority int64
	Index    int
}

type PriorityQueue []*Item

func NewPriorityQueue(capacity int) PriorityQueue {
	return make(PriorityQueue, 0, capacity)
}

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	c := cap(*pq)
	if n+1 > c {
		tmp := make(PriorityQueue, n, c*2)
		copy(tmp, *pq)
		*pq = tmp
	}

	*pq = (*pq)[:n+1]
	item, ok := x.(*Item)
	if !ok {
		return
	}

	item.Index = n
	(*pq)[n] = item
}

func (pq *PriorityQueue) Pop() interface{} {
	n := len(*pq)
	c := cap(*pq)
	half := c / 2
	if n < half && c > 64 {
		tmp := make(PriorityQueue, n, half)
		copy(tmp, *pq)
		*pq = tmp
	}

	idx := n - 1
	item := (*pq)[idx]
	item.Index = -1
	*pq = (*pq)[:idx]
	return item
}

func (pq *PriorityQueue) Update(item *Item, v interface{}, priority int64) {
	if item == nil {
		return
	}
	item.Value = v
	item.Priority = priority
	heap.Fix(pq, item.Index)
}

func (pq *PriorityQueue) Peek() *Item {
	if pq.Len() == 0 {
		return nil
	}
	return (*pq)[0]
}

func (pq *PriorityQueue) Get(maxPriority int64) *Item {
	if pq.Len() == 0 {
		return nil
	}
	item := (*pq)[0]
	if item == nil {
		return nil
	}
	if item.Priority > maxPriority {
		return nil
	}
	heap.Remove(pq, 0)
	return item
}
