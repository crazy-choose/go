package custom

import (
	"sync"
	"sync/atomic"
)

type Map struct {
	m     sync.Map
	count int64
}

func (cm *Map) M() sync.Map {
	return cm.m
}

func (cm *Map) Store(key, value interface{}) {
	_, loaded := cm.m.LoadOrStore(key, value)
	if !loaded { // 新元素
		atomic.AddInt64(&cm.count, 1)
	}
}

func (cm *Map) Delete(key interface{}) {
	if _, loaded := cm.m.Load(key); loaded {
		cm.m.Delete(key)
		atomic.AddInt64(&cm.count, -1)
	}
}

func (cm *Map) Len() int {
	return int(atomic.LoadInt64(&cm.count))
}

func (cm *Map) Empty() bool {
	return atomic.LoadInt64(&cm.count) <= 0
}

func (cm *Map) AnyMap() map[any]any {
	if atomic.LoadInt64(&cm.count) <= 0 {
		return nil
	}
	ret := make(map[any]any)
	cm.m.Range(func(key, value any) bool {
		ret[key] = value
		return true
	})

	return ret
}
