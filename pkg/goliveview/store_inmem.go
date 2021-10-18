package goliveview

import "sync"

type store struct {
	data map[string]interface{}
	sync.RWMutex
}

func (s store) Set(kvs ...KV) error {
	s.Lock()
	defer s.Unlock()
	for _, kv := range kvs {
		if kv.Temp {
			continue
		}
		s.data[kv.K] = kv.V
	}
	return nil
}

func (s store) Get(key string) (interface{}, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.data[key]
	return v, ok
}
