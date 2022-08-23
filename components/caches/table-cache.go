package caches

import (
	"catuan/comm"
	"sync"
)

type TableHandlerFunc[K comm.KeyAble, V any] func(key K) (V, error)

// TableCache 适用数据量少,配置类数据/定量数据，例如 系统配置/商家信息/商品信息
type TableCache[K comm.KeyAble, V any] struct {
	handler TableHandlerFunc[K, V]
	data    map[K]V
	mu      sync.RWMutex
}

func NewTableCache[K comm.KeyAble, V any](handler TableHandlerFunc[K, V]) *TableCache[K, V] {
	return &TableCache[K, V]{
		handler: handler,
		data:    make(map[K]V),
		mu:      sync.RWMutex{},
	}
}

// Get 获取缓存
func (t *TableCache[K, V]) Get(key K) (V, bool) {
	t.mu.RLock()
	val, ok := t.data[key]
	t.mu.RUnlock()
	if ok {
		return val, true
	}
	if t.handler != nil {
		val, err := t.handler(key)
		if err != nil {
			return val, false
		}
		t.Set(key, val)

		return val, true
	}
	return val, false
}

// Set 获取缓存
func (t *TableCache[K, V]) Set(key K, value V) {
	t.mu.Lock()
	t.data[key] = value
	t.mu.Unlock()
}

// Del 删除缓存
func (t *TableCache[K, V]) Del(key K) {
	t.mu.Lock()
	if _, ok := t.data[key]; ok {
		delete(t.data, key)
	}
	t.mu.Unlock()
}

// Clean 清空缓存
func (t *TableCache[K, V]) Clean() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.data = make(map[K]V)

	return nil
}
