/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/12/18
   Description :
-------------------------------------------------
*/

package zcache_broker

import (
    "crypto/md5"
    "encoding/hex"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/zlyuancn/zerrors"
    "github.com/zlyuancn/zlog2"
    "github.com/zlyuancn/zsingleflight"
)

var ErrNoEntry = errors.New("条目不存在")
var ErrLoadDBFnNotExists = errors.New("db加载函数不存在或为空")

type LoadDBFn func(space, key string, params ...string) ([]byte, error)

type CacheDB interface {
    Del(key string) error
    // 获取一个值, 如果这个key在缓存中不存在应该返回 ErrNoEntry 错误
    Get(key string) ([]byte, error)
    Set(key string, value interface{}, ex time.Duration) error
    SetTTL(key string, ex time.Duration) error
}

type PublicInterface interface {
    // 获取
    Get(space, key string, params ...string) ([]byte, error)
    // 从缓存中删除
    Del(space, key string, params ...string) error
    // 让一个key失效并立即从db中重新加载
    Refresh(space, key string, params ...string) ([]byte, error)
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
func (m *CacheBroker) get(space, key string, params ...string) ([]byte, error) {
    ckey := makeCacheKey(space, key, params...)
    bs, err := m.c.Get(ckey)
    if err != nil {
        if err == ErrNoEntry {
            return nil, err
        }
        return nil, zerrors.WithMessage(err, "缓存加载失败")
    }

    // 刷新ttl
    if sc, ok := m.spaces[space]; ok && sc.prepare_auto_ref(key) {
        if err := m.c.SetTTL(ckey, sc.expirat()); err != nil {
            sfkey := makeSFKey(space, key, params...)
            m.log.Warn(zerrors.WithMessagef(err, "刷新ttl失败<%s>", sfkey))
        }
    }
    return bs, nil
}

// 从db加载
func (m *CacheBroker) loadDB(space, key string, params ...string) ([]byte, error) {
    var fn LoadDBFn
    if sc, ok := m.spaces[space]; ok {
        fn = sc.getLoadDBFn()
    }
    if fn == nil {
        return nil, ErrLoadDBFnNotExists
    }

    ckey := makeCacheKey(space, key, params...)
    // 从db加载
    bs, err := fn(space, key, params...)
    if err != nil {
        if err := m.c.Del(ckey); err != nil { // 从db加载失败时从缓存删除
            sfkey := makeSFKey(space, key, params...)
            m.log.Warn(zerrors.WithMessagef(err, "db加载失败后删除缓存失败<%s>", sfkey))
        }
        return nil, zerrors.WithMessage(err, "db加载失败")
    }

    // 缓存
    ex := time.Duration(0)
    if sc, ok := m.spaces[space]; ok {
        ex = sc.expirat()
    }
    if err := m.c.Set(ckey, bs, ex); err != nil { // 不管缓存是否成功
        sfkey := makeSFKey(space, key, params...)
        m.log.Warn(zerrors.WithMessagef(err, "db加载后缓存失败<%s>", sfkey))
    }
    return bs, nil
}

// 获取
func (m *CacheBroker) Get(space, key string, params ...string) ([]byte, error) {
    // 同时只能有一个goroutine在获取数据,其它goroutine直接等待结果
    sfkey := makeSFKey(space, key, params...)
    v, err := m.sf.Do(sfkey, func() (interface{}, error) {
        bs, gerr := m.get(space, key, params...)
        if gerr == nil {
            return bs, nil
        }

        bs, lerr := m.loadDB(space, key, params...)
        if lerr == nil {
            return bs, nil
        }

        if gerr == ErrNoEntry {
            return nil, lerr
        }
        return nil, zerrors.WithMessage(gerr, lerr.Error())
    })
    if err != nil {
        m.log.Warn(zerrors.WithMessagef(err, "加载失败<%s>", sfkey))
        return nil, err
    }

    return v.([]byte), nil
}

// 从缓存中删除
func (m *CacheBroker) Del(space, key string, params ...string) error {
    sfkey := makeSFKey(space, key, params...)
    _, err := m.sf.Do(sfkey, func() (interface{}, error) {
        return nil, m.c.Del(makeCacheKey(space, key, params...))
    })
    if err != nil {
        m.log.Warn(zerrors.WithMessagef(err, "删除失败<%s>", sfkey))
        return err
    }
    return nil
}

// 让一个key失效并立即从db中重新加载
func (m *CacheBroker) Refresh(space, key string, params ...string) ([]byte, error) {
    sfkey := makeSFKey(space, key, params...)
    v, err := m.sf.Do(sfkey, func() (interface{}, error) {
        return m.loadDB(space, key, params...)
    })
    if err != nil {
        m.log.Warn(zerrors.WithMessagef(err, "刷新失败<%s>", sfkey))
        return nil, err
    }
    return v.([]byte), nil
}

func makeCacheKey(space, key string, params ...string) string {
    if len(params) != 0 {
        key = fmt.Sprintf("%s?%s", key, strings.Join(params, "&"))
    }
    return fmt.Sprintf("%s:%s", space, makeMd5(key))
}

func makeSFKey(space, key string, params ...string) string {
    if len(params) != 0 {
        return fmt.Sprintf("%s:%s?%s", space, key, strings.Join(params, "&"))
    }
    return fmt.Sprintf("%s:%s", space, key)
}

func makeMd5(text string) string {
    m := md5.New()
    m.Write([]byte(text))
    return hex.EncodeToString(m.Sum(nil))
}
