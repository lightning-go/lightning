/**
 * @author: Jason
 * Created: 19-5-3
 */

package utils

import (
	"github.com/pkg/errors"
	"math"
	"container/list"
	"sync"
)

var (
	GenerateIdError = errors.New("Generate Id failed")

	defaultIdGenerator = NewIdGenerator()

	minId = uint64(1)
	maxId = uint64(math.MaxUint64)
)

func GetId() uint64 {
	return defaultIdGenerator.Get()
}

func GetAvailableId() uint64 {
	return defaultIdGenerator.GetAvailableId()
}

func CollectId(id uint64) {
	defaultIdGenerator.Put(id)
}

type IdGenerator struct {
	mux        sync.RWMutex
	id         uint64
	cycleCount uint64
	idList     *list.List
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{
		id:     minId,
		idList: list.New(),
	}
}

func (idg *IdGenerator) Get() (id uint64) {
	defer func() {
		if recover() != nil {
			panic(GenerateIdError)
		}
	}()

	idg.mux.Lock()
	id = idg.id
	if id == maxId {
		id = minId
		idg.cycleCount++
	} else {
		idg.id++
	}
	idg.mux.Unlock()

	return
}

func (idg *IdGenerator) GetAvailableId() (id uint64) {
	defer func() {
		if recover() != nil {
			panic(GenerateIdError)
		}
	}()

	idg.mux.Lock()

	element := idg.idList.Front()
	if element != nil {
		id = element.Value.(uint64)
		idg.idList.Remove(element)
		idg.mux.Unlock()
		return
	}

	id = idg.id
	if id == maxId {
		id = minId
		idg.cycleCount++
	} else {
		idg.id++
	}

	idg.mux.Unlock()
	return
}

func (idg *IdGenerator) CycleCount() uint64 {
	idg.mux.RLock()
	c := idg.cycleCount
	idg.mux.RUnlock()
	return c
}

func (idg *IdGenerator) Put(id uint64) {
	idg.mux.Lock()
	idg.idList.PushBack(id)
	idg.mux.Unlock()
}
