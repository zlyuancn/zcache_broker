/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/12/20
   Description :
-------------------------------------------------
*/

package wredis

import (
    "time"

    "github.com/go-redis/redis"

    "github.com/zlyuancn/zcache_broker"
)

var _ zcache_broker.CacheDB = (*RedisWrap)(nil)

type RedisWrap struct {
    c redis.UniversalClient
}

func Wrap(c redis.UniversalClient) *RedisWrap {
    return &RedisWrap{c}
}

func (m *RedisWrap) Get(key string) ([]byte, error) {
    bs, err := m.c.Get(key).Bytes()
    if err == redis.Nil {
        return nil, zcache_broker.ErrNoEntry
    }
    return bs, err
}

func (m *RedisWrap) Del(key string) error {
    err := m.c.Del(key).Err()
    if err == redis.Nil {
        return nil
    }
    return err
}

func (m *RedisWrap) Set(key string, value interface{}, ex time.Duration) error {
    return m.c.Set(key, value, ex).Err()
}

func (m *RedisWrap) SetTTL(key string, ex time.Duration) error {
    return m.c.PExpire(key, ex).Err()
}
