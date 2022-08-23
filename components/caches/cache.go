package caches

import "catuan/comm"

type CleanAble interface {
	Clean() error
}

type CacheInf[K comm.KeyAble, V any] interface {
	CleanAble
	Get(key K) (V, bool)
	Set(key K, value V)
	Del(key K)
}

type ExpiresData[T any] struct {
	Expires int64
	Data    T
}
