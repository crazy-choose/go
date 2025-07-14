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

// Store 存储值的副本的指针，即使传入值类型也会创建副本
func (cm *Map[K, V]) Store(key K, value V) {
	// 创建值的副本并存储其指针
	copier := value // 强制创建副本（即使 value 是值类型）
	cm.m.Store(key, &copier)
	atomic.AddInt64(&cm.count, 1)
}

// StorePtr 直接存储指针（谨慎使用，确保指针有效）
func (cm *Map[K, V]) StorePtr(key K, ptr *V) {
	if ptr == nil {
		return
	}
	cm.m.Store(key, ptr)
	atomic.AddInt64(&cm.count, 1)
}

// Qry 查询键对应的值的指针，如果键不存在则返回 nil 和 false
func (cm *Map[K, V]) Qry(key K) (*V, bool) {
	value, loaded := cm.m.Load(key)
	if !loaded {
		return nil, false
	}
	return value.(*V), true
}

// Delete 删除键值对，如果键存在则计数器减1
func (cm *Map[K, V]) Delete(key K) {
	if _, loaded := cm.m.Load(key); loaded {
		cm.m.Delete(key)
		atomic.AddInt64(&cm.count, -1)
	}
}

// Clean 清空所有键值对
func (cm *Map[K, V]) Clean() {
	cm.m.Range(func(key, value any) bool {
		cm.m.Delete(key)
		return true
	})
	atomic.StoreInt64(&cm.count, 0)
}

// Len 返回 map 中元素的数量
func (cm *Map[K, V]) Len() int {
	return int(atomic.LoadInt64(&cm.count))
}

// Empty 判断 map 是否为空
func (cm *Map[K, V]) Empty() bool {
	return atomic.LoadInt64(&cm.count) <= 0
}

// Map 将 map 转换为普通的 map[K]*V，返回指针的副本
func (cm *Map[K, V]) Map() (ret map[K]*V) {
	if atomic.LoadInt64(&cm.count) <= 0 {
		return nil
	}
	ret = make(map[K]*V)
	cm.m.Range(func(key, value any) bool {
		// 将 key 转换为 K 类型
		ret[key.(K)] = value.(*V)
		return true
	})
	return ret
}

func (cm *Map[K, V]) KeyList() (ret []K) {
	if atomic.LoadInt64(&cm.count) <= 0 {
		return nil
	}
	ret = make([]K, 0)
	cm.m.Range(func(key, value any) bool {
		// 将 key 转换为 K 类型
		ret = append(ret, key.(K))
		return true
	})
	return ret
}

func (cm *Map[K, V]) ValList() (ret []*V) {
	if atomic.LoadInt64(&cm.count) <= 0 {
		return nil
	}
	ret = make([]*V, 0)
	cm.m.Range(func(key, value any) bool {
		// 将 key 转换为 K 类型
		ret = append(ret, value.(*V))
		return true
	})
	return ret
}

// M 返回原始的 sync.Map（谨慎使用，可能破坏计数器一致性）
func (cm *Map[K, V]) M() sync.Map {
	return cm.m
}
