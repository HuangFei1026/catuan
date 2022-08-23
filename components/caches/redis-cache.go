package caches

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"time"
)

type RedisCache struct {
	client      *redis.Client
	expiresTime time.Duration //过期时间
}

func NewRedisCache(client *redis.Client, expiresTime time.Duration) *RedisCache {
	return &RedisCache{
		client:      client,
		expiresTime: expiresTime,
	}
}

func (r *RedisCache) Get(key string) (string, bool) {
	val, err := r.client.Get(context.TODO(), key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

func (r *RedisCache) Set(key string, value string) error {
	res := r.client.Set(context.TODO(), key, value, r.expiresTime)
	if res.Err() != nil {
		logrus.WithFields(logrus.Fields{
			"tip": "设置缓存异常",
		}).Error(res.Err().Error())
		return res.Err()
	}
	return nil
}

func (r *RedisCache) Del(key string) {
	res := r.client.Del(context.TODO(), key)
	if res.Err() != nil {
		logrus.WithFields(logrus.Fields{
			"tip": "删除缓存异常",
		}).Error(res.Err().Error())
	}
}

func (r *RedisCache) Clean() error {
	r.client.FlushDB(context.TODO())
	return nil
}
