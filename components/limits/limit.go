package limits

import (
	"catuan/comm"
	"sync"
)

// UniLimit 唯一限制器 单 goroutine 限制访问
type UniLimit[K comm.KeyAble] struct {
	mu   sync.Mutex
	dest map[K]bool
}

func NewUniLimit[K comm.KeyAble]() *UniLimit[K] {
	return &UniLimit[K]{
		dest: make(map[K]bool),
		mu:   sync.Mutex{},
	}
}

func (l *UniLimit[K]) Check(key K) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.dest[key]
	if !ok {
		l.dest[key] = true
	}
	return !ok
}

func (l *UniLimit[K]) Release(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.dest[key]; ok {
		delete(l.dest, key)
	}
}
