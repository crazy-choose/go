package custom

import (
	"sync"
	"sync/atomic"
)

// Map 是一个并发安全的泛型 map，支持任意类型的键和值
type Map[K comparable, V any] struct {
	m     sync.Map
	count int64
}

// Save 存储键值对，如果键不存在则计数器加1
func (cm *Map[K, V]) Save(key K, value V) {
	_, loaded := cm.m.LoadOrStore(key, value)
	if !loaded { // 新元素
		atomic.AddInt64(&cm.count, 1)
	}
}

// Qry 查询键对应的值，如果键不存在则返回零值和 false
func (cm *Map[K, V]) Qry(key K) (V, bool) {
	value, loaded := cm.m.Load(key)
	if !loaded {
		var zero V
		return zero, false
	}
	return value.(V), true
}

// Delete 删除键值对，如果键存在则计数器减1
func (cm *Map[K, V]) Delete(key K) {
	if _, loaded := cm.m.Load(key); loaded {
		cm.m.Delete(key)
		atomic.AddInt64(&cm.count, -1)
	}
}

// Len 返回 map 中元素的数量
func (cm *Map[K, V]) Len() int {
	return int(atomic.LoadInt64(&cm.count))
}

// Empty 判断 map 是否为空
func (cm *Map[K, V]) Empty() bool {
	return atomic.LoadInt64(&cm.count) <= 0
}

// AnyMap 将 map 转换为普通的 map[K]V，返回副本
func (cm *Map[K, V]) AnyMap() map[any]V {
	if atomic.LoadInt64(&cm.count) <= 0 {
		return nil
	}
	ret := make(map[any]V)
	cm.m.Range(func(key, value any) bool {
		ret[key.(K)] = value.(V)
		return true
	})
	return ret
}

// M 返回原始的 sync.Map（谨慎使用，可能破坏计数器一致性）
func (cm *Map[K, V]) M() sync.Map {
	return cm.m
}
