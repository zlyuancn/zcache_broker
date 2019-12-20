/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/12/20
   Description :
-------------------------------------------------
*/

package wredis

import (
    "github.com/go-redis/redis"
    "time"
)

type RedisWrap struct {
    c redis.UniversalClient
}

func Wrap(c redis.UniversalClient) *RedisWrap {
    return &RedisWrap{c}
}

func (m *RedisWrap) Get(key string) ([]byte, error) {
    return m.c.Get(key).Bytes()
}

func (m *RedisWrap) Del(key string) error {
    return m.c.Del(key).Err()
}

func (m *RedisWrap) Set(key string, value interface{}, ex time.Duration) error {
    return m.c.Set(key, value, ex).Err()
}

func (m *RedisWrap) SetTTL(key string, ex time.Duration) error {
    return m.c.PExpire(key, ex).Err()
}
