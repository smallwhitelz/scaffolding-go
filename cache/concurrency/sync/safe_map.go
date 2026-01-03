package sync

import "sync"

type SafeMap[K comparable, V any] struct {
	data  map[K]V
	mutex sync.RWMutex
}

func (s *SafeMap[K, V]) Put(key K, val V) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = val
}

func (s *SafeMap[K, V]) Get(key K) (any, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *SafeMap[K, V]) LoadOrStore(key K, newVal V) (val V, loaded bool) {
	s.mutex.RLock()
	res, ok := s.data[key]
	s.mutex.RUnlock()
	if ok {
		return res, true
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// double-check写法
	res, ok = s.data[key]
	if ok {
		return res, true
	}

	s.data[key] = newVal
	return newVal, false
}
