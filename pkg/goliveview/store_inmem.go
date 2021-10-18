package goliveview

import "sync"

type store struct {
	data map[string]interface{}
	sync.RWMutex
}

func (s store) Set(m map[string]interface{}) error {
	s.Lock()
	defer s.Unlock()
	for k, v := range m {
		s.data[k] = v
	}
	return nil
}

func (s store) Get(key string) (interface{}, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.data[key]
	return v, ok
}
