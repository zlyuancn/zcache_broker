/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/12/18
   Description :
-------------------------------------------------
*/

package zcache_broker

import (
    "errors"
    "fmt"
    "github.com/zlyuancn/zsingleflight"
    "time"
)

const (
    DefaultSpace = "zcb"
)

var ErrLoadDBFnNotExists = errors.New("db加载函数不存在或为空")

type LoadDBFn func() ([]byte, error)

type CacheDB interface {
    Del(key string) error
    Get(key string) ([]byte, error)
    Set(key string, value interface{}, ex time.Duration) error
    SetTTL(key string, ex time.Duration) error
}

type PublicInterface interface {
    // 获取
    Get(space, key string) ([]byte, error)
    // 从缓存中删除
    Del(space, key string) error
    // 让一个key失效并立即从db中重新加载
    Refresh(space, key string) error
}

type CacheBroker struct {
    sf     *zsingleflight.SingleFlight // 单飞
    c      CacheDB                     // 客户端
    spaces map[string]*SpaceConfig     // 命名空间配置
}

func New(c CacheDB, opts ...Option) (*CacheBroker, error) {
    m := &CacheBroker{
        sf:     zsingleflight.New(),
        c:      c,
        spaces: make(map[string]*SpaceConfig),
    }

    for _, o := range opts {
        o(m)
    }

    return m, nil
}

// 设置(正式运行后不应该再调用它)
func (m *CacheBroker) SetOptions(opts ...Option) {
    for _, o := range opts {
        o(m)
    }
}

// 从缓存加载
func (m *CacheBroker) get(space, key string) ([]byte, error) {
    rkey := MakeKey(space, key)
    bs, err := m.c.Get(rkey)
    if err != nil {
        return nil, err
    }

    // 刷新ttl
    if sc, ok := m.spaces[space]; ok && sc.auto_ref {
        _ = m.c.SetTTL(rkey, sc.expirat())
    }
    return bs, nil
}

// 从db加载
func (m *CacheBroker) loadDB(space, key string, fn LoadDBFn) ([]byte, error) {
    if fn == nil {
        if sc, ok := m.spaces[space]; ok {
            fn = sc.getLoadDBFn(key)
        }
        if fn == nil {
            return nil, ErrLoadDBFnNotExists
        }
    }

    rkey := MakeKey(space, key)
    // 从db加载
    bs, err := fn()
    if err != nil {
        _ = m.c.Del(rkey) // 从db加载失败时从缓存删除
        return nil, err
    }

    // 缓存
    ex := time.Duration(0)
    if sc, ok := m.spaces[space]; ok {
        ex = sc.expirat()
    }
    _ = m.c.Set(rkey, bs, ex) // 不管缓存是否成功

    return bs, nil
}

// 获取
func (m *CacheBroker) Get(space, key string) ([]byte, error) {
    return m.GetWithFn(space, key, nil)
}

// 获取
func (m *CacheBroker) GetWithFn(space, key string, fn LoadDBFn) ([]byte, error) {
    // 同时只能有一个goroutine在获取数据,其它goroutine直接等待结果
    v, err := m.sf.Do(MakeKey(space, key), func() (interface{}, error) {
        bs, err := m.get(space, key)
        if err != nil {
            return m.loadDB(space, key, fn)
        }
        return bs, nil
    })
    if err != nil {
        return nil, err
    }

    return v.([]byte), nil
}

// 从缓存中删除
func (m *CacheBroker) Del(space, key string) error {
    rkey := MakeKey(space, key)
    _, err := m.sf.Do(rkey, func() (interface{}, error) {
        return nil, m.c.Del(rkey)
    })
    return err
}

// 让一个key失效并立即从db中重新加载
func (m *CacheBroker) Refresh(space, key string) error {
    rkey := MakeKey(space, key)
    _, err := m.sf.Do(rkey, func() (interface{}, error) {
        return m.loadDB(space, key, nil)
    })
    return err
}

func MakeKey(space, key string) string {
    if space == "" {
        space = DefaultSpace
    }
    return fmt.Sprintf("%s:%s", space, key)
}
