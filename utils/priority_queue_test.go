package utils

import (
	"container/heap"
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	max := 100
	pq := NewPriorityQueue(max)

	for i := 0; i < max; i++ {
		heap.Push(&pq, &Item{Value: i, Priority: int64(i)})
	}

	for i := 0; i < max; i++ {
		peek := pq.Peek()
		item := heap.Pop(&pq)
		fmt.Println(peek.Value.(int), item.(*Item).Value.(int))
	}


}

func TestPriorityQueue2(t *testing.T) {
	max := 100
	pq := NewPriorityQueue(max)
	d := make([]int, 0, max)

	for i := 0; i < max; i++ {
		v := rand.Int()
		d = append(d, v)
		heap.Push(&pq, &Item{Value: i, Priority: int64(v)})
	}
	sort.Ints(d)
	fmt.Println(d)

	for i := 0; i < max; i++ {
		peek := pq.Peek()
		item := heap.Pop(&pq)
		fmt.Println(d[i], peek.Value.(int), peek.Priority, " - ", item.(*Item).Value.(int), item.(*Item).Priority)
	}


}

