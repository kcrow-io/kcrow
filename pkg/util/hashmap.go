package util

import (
	"fmt"
	"sync"
)

type HashMap[K comparable, V any] struct {
	m map[K]V
	sync.RWMutex
}

func New[K comparable, V any]() *HashMap[K, V] {
	return &HashMap[K, V]{m: make(map[K]V)}
}

func (m *HashMap[K, V]) Put(key K, value V) {
	m.Lock()
	defer m.Unlock()
	m.m[key] = value
}

func (m *HashMap[K, V]) Get(key K) (value V, found bool) {
	m.RLock()
	defer m.RUnlock()
	value, found = m.m[key]
	return
}

func (m *HashMap[K, V]) Iter(fn func(K, V) bool) {
	m.RLock()
	defer m.RUnlock()
	for k, v := range m.m {
		ok := fn(k, v)
		if !ok {
			return
		}
	}
}

func (m *HashMap[K, V]) Remove(key K) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, key)
}

func (m *HashMap[K, V]) Empty() bool {
	return m.Size() == 0
}

func (m *HashMap[K, V]) Size() int {
	return len(m.m)
}

func (m *HashMap[K, V]) Keys() []K {
	keys := make([]K, m.Size())
	count := 0
	m.RLock()
	defer m.RUnlock()
	for key := range m.m {
		keys[count] = key
		count++
	}
	return keys
}

func (m *HashMap[K, V]) Values() []V {
	values := make([]V, m.Size())
	count := 0
	m.RLock()
	defer m.RUnlock()
	for _, value := range m.m {
		values[count] = value
		count++
	}
	return values
}

func (m *HashMap[K, V]) Clear() {
	m.Lock()
	defer m.Unlock()
	clear(m.m)
}

func (m *HashMap[K, V]) String() string {
	str := "HashMap: "
	str += fmt.Sprintf("%v", m.m)
	return str
}
