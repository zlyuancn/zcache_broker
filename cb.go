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
    "time"

    "github.com/zlyuancn/zerrors"
    "github.com/zlyuancn/zlog2"
    "github.com/zlyuancn/zsingleflight"
)

const (
    DefaultSpace = "zcb"
)

var ErrLoadDBFnNotExists = errors.New("db加载函数不存在或为空")

type LoadDBFn func(space, key string) ([]byte, error)

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
    Refresh(space, key string) ([]byte, error)
}

type CacheBroker struct {
    sf     *zsingleflight.SingleFlight // 单飞
    c      CacheDB                     // 客户端
    spaces map[string]*SpaceConfig     // 命名空间配置
    log    Loger                       // 日志组件
}

func New(c CacheDB, opts ...Option) (*CacheBroker, error) {
    m := &CacheBroker{
        sf:     zsingleflight.New(),
        c:      c,
        spaces: make(map[string]*SpaceConfig),
        log:    zlog2.DefaultLogger,
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
        return nil, zerrors.WithMessage(err, "缓存加载失败")
    }

    // 刷新ttl
    if sc, ok := m.spaces[space]; ok && sc.prepare_auto_ref(key) {
        if err := m.c.SetTTL(rkey, sc.expirat()); err != nil {
            m.log.Warn(zerrors.WithMessagef(err, "刷新ttl失败<%s>", rkey))
        }
    }
    return bs, nil
}

// 从db加载
func (m *CacheBroker) loadDB(space, key string, fn LoadDBFn) ([]byte, error) {
    if fn == nil {
        if sc, ok := m.spaces[space]; ok {
            fn = sc.getLoadDBFn()
        }
        if fn == nil {
            return nil, ErrLoadDBFnNotExists
        }
    }

    rkey := MakeKey(space, key)
    // 从db加载
    bs, err := fn(space, key)
    if err != nil {
        if err := m.c.Del(rkey); err != nil { // 从db加载失败时从缓存删除
            m.log.Warn(zerrors.WithMessagef(err, "db加载失败后删除缓存失败<%s>", rkey))
        }
        return nil, zerrors.WithMessage(err, "db加载失败")
    }

    // 缓存
    ex := time.Duration(0)
    if sc, ok := m.spaces[space]; ok {
        ex = sc.expirat()
    }
    if err := m.c.Set(rkey, bs, ex); err != nil { // 不管缓存是否成功
        m.log.Warn(zerrors.WithMessagef(err, "db加载后缓存失败<%s>", rkey))
    }
    return bs, nil
}

// 获取
func (m *CacheBroker) Get(space, key string) ([]byte, error) {
    return m.GetWithFn(space, key, nil)
}

// 获取
func (m *CacheBroker) GetWithFn(space, key string, fn LoadDBFn) ([]byte, error) {
    // 同时只能有一个goroutine在获取数据,其它goroutine直接等待结果
    rkey := MakeKey(space, key)
    v, err := m.sf.Do(rkey, func() (interface{}, error) {
        bs, gerr := m.get(space, key)
        if gerr != nil {
            bs, lerr := m.loadDB(space, key, fn)
            if lerr != nil {
                err := zerrors.WithMessagef(lerr, "%s", gerr)
                return nil, err
            }
            return bs, nil
        }
        return bs, nil
    })
    if err != nil {
        m.log.Warn(zerrors.WithMessagef(err, "加载失败<%s>", rkey))
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
    if err != nil {
        m.log.Warn(zerrors.WithMessagef(err, "删除失败<%s>", rkey))
        return err
    }
    return nil
}

// 让一个key失效并立即从db中重新加载
func (m *CacheBroker) Refresh(space, key string) ([]byte, error) {
    rkey := MakeKey(space, key)
    v, err := m.sf.Do(rkey, func() (interface{}, error) {
        return m.loadDB(space, key, nil)
    })
    if err != nil {
        m.log.Warn(zerrors.WithMessagef(err, "刷新失败<%s>", rkey))
        return nil, err
    }
    return v.([]byte), nil
}

func MakeKey(space, key string) string {
    if space == "" {
        space = DefaultSpace
    }
    return fmt.Sprintf("%s:%s", space, key)
}
