package dns

import (
	"sync"

	"github.com/masx200/http3-reverse-proxy-server-experiment/generic"
)

type MapImplementSynchronous[T comparable, Y any] struct {
	data       generic.MapInterface[T, Y]
	cacheMutex sync.Mutex
}

func (m *MapImplementSynchronous[T, Y]) Clear() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.data.Clear()
}

// Delete implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) Delete(key T) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.data.Delete(key)
}

// Entries implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) Entries() []generic.PairInterface[T, Y] {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.data.Entries()
}

// ForEach implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) ForEach(f func(Y, T, generic.MapInterface[T, Y])) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.data.ForEach(f)
}

// Get implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) Get(key T) (Y, bool) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.data.Get(key)
}

// Has implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) Has(key T) bool {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.data.Has(key)
}

// Iterator implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) Iterator() generic.Iterator[generic.PairInterface[T, Y]] {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.data.Iterator()
}

// Keys implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) Keys() []T {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.data.Keys()
}

// Set implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) Set(key T, value Y) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.data.Set(key, value)
}

// Size implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) Size() int64 {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.data.Size()
}

// Values implements generic.MapInterface.
func (m *MapImplementSynchronous[T, Y]) Values() []Y {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.data.Values()
}

func NewMapImplementSynchronous[T comparable, Y any](entries ...generic.PairInterface[T, Y]) generic.MapInterface[T, Y] {
	return &MapImplementSynchronous[T, Y]{
		data:       generic.NewMapImplement[T, Y](entries...),
		cacheMutex: sync.Mutex{},
	}
}
