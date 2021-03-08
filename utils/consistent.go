package utils

import (
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

const defaultReplicas = 100


type SortUint32 []uint32

func (su SortUint32) Len() int {
	return len(su)
}

func (su SortUint32) Less(i, j int) bool {
	return su[i] < su[j]
}

func (su SortUint32) Swap(i, j int) {
	su[i], su[j] = su[j], su[i]
}


type ConsistentHash struct {
	sync.RWMutex
	nodes 		map[uint32]string
	hashKeys 	SortUint32
	replicasNum int
}

func NewConsistentHash(initReplicasNum ...int) *ConsistentHash {
	replicasNum := defaultReplicas
	if len(initReplicasNum) > 0 {
		replicas := initReplicasNum[0]
		if replicas > 0 {
			replicasNum = replicas
		}
	}
	return &ConsistentHash{
		nodes: make(map[uint32]string),
		hashKeys: SortUint32{},
		replicasNum: replicasNum,
	}
}

func (ch *ConsistentHash) getHash(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func (ch *ConsistentHash) getHashKey(key string, serial int) uint32 {
	return ch.getHash(fmt.Sprintf("%s%d", key, serial))
}

func (ch *ConsistentHash) AddNodes(nodes []string) {
	if nodes == nil {
		return
	}

	ch.Lock()
	defer ch.Unlock()

	for _, node := range nodes {
		ch.add(node, false)
	}
	ch.updateHashKeys()
}

func (ch *ConsistentHash) Add(node string) {
	ch.Lock()
	defer ch.Unlock()
	ch.add(node)
}

func (ch *ConsistentHash) add(node string, update ...bool) {
	if len(node) == 0 {
		return
	}
	for i := 0; i < ch.replicasNum; i++ {
		k := ch.getHashKey(node, i)
		ch.nodes[k] = node
	}
	updateKeys := true
	if len(update) > 0 {
		updateKeys = update[0]
	}
	if updateKeys {
		ch.updateHashKeys()
	}

}

func (ch *ConsistentHash) RemoveNodes(nodes []string) {
	if nodes == nil {
		return
	}

	ch.Lock()
	defer ch.Unlock()

	for _, node := range nodes {
		ch.remove(node, false)
	}
	ch.updateHashKeys()
}

func (ch *ConsistentHash) Remove(node string) {
	ch.Lock()
	defer ch.Unlock()
	ch.remove(node)
}

func (ch *ConsistentHash) remove(node string, update ...bool) {
	if len(node) == 0 {
		return
	}
	for i := 0; i < ch.replicasNum; i++ {
		k := ch.getHashKey(node, i)
		delete(ch.nodes, k)
	}
	updateKeys := true
	if len(update) > 0 {
		updateKeys = update[0]
	}
	if updateKeys {
		ch.updateHashKeys()
	}

}

func (ch *ConsistentHash) updateHashKeys() {
	keys := ch.hashKeys[:0]

	n := len(ch.hashKeys)
	c := cap(ch.hashKeys)
	half := c / 2
	if half > n {
		keys = make([]uint32, 0, half)
	}
	for k := range ch.nodes {
		keys = append(keys, k)
	}

	sort.Sort(keys)
	ch.hashKeys = keys
}


func (ch *ConsistentHash) GetNode(key string) string {
	hash := ch.getHash(key)

	ch.RLock()
	defer ch.RUnlock()

	idx := ch.GetNodeIdx(hash)
	v, ok := ch.nodes[ch.hashKeys[idx]]
	if !ok {
		return ""
	}
	return v
}

func (ch *ConsistentHash) GetNodeIdx(hash uint32) int {
	i := sort.Search(len(ch.hashKeys), func(i int) bool {
		return ch.hashKeys[i] >= hash
	})
	keyLen := len(ch.hashKeys)
	lastIdx := keyLen - 1
	if i < keyLen {
		if i == lastIdx {
			return 0
		} else {
			return i
		}
	} else {
		return lastIdx
	}
}






