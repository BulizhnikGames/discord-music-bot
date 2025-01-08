package internal

import "sync"

type AsyncMap[K comparable, V any] struct {
	Data  map[K]V
	Mutex *sync.RWMutex
}

func (m *AsyncMap[K, V]) Put(k K, v V) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.Data[k] = v
}

func (m *AsyncMap[K, V]) Remove(k K) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	delete(m.Data, k)
}
