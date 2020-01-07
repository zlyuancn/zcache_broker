/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/12/18
   Description :
-------------------------------------------------
*/

package zcache_broker

import (
    "crypto/rand"
    "math/big"
    "sync"
    "time"
)

const DefaultAutoRefInterval = time.Second * 3 // 自动刷新间隔, 它禁止高并发Get每次都触发自动刷新

type SpaceConfig struct {
    rand     bool
    auto_ref bool
    ex       time.Duration
    endex    time.Duration
    fn       LoadDBFn

    auto_ref_interval time.Duration            // 自动刷新间隔
    next_ref_time     map[string]time.Duration // 下次刷新时间
    mx                sync.RWMutex
}

func NewSpaceConfig() *SpaceConfig {
    return &SpaceConfig{
        next_ref_time:     make(map[string]time.Duration),
        auto_ref_interval: DefaultAutoRefInterval,
    }
}

// 设置过期时间
func (m *SpaceConfig) SetExpirat(ex time.Duration, auto_refresh bool) *SpaceConfig {
    if ex < 0 {
        ex = 0
    }
    m.ex = ex
    m.endex = ex
    m.rand = false
    m.auto_ref = auto_refresh
    return m
}

// 设置随机过期时间
func (m *SpaceConfig) SetRandExpirat(start_ex time.Duration, end_ex time.Duration, auto_refresh bool) *SpaceConfig {
    if start_ex < 0 {
        start_ex = 0
    }
    if end_ex < start_ex {
        end_ex = start_ex
    }
    m.ex = start_ex
    m.endex = end_ex
    m.rand = true
    m.auto_ref = auto_refresh
    return m
}

// 设置自动刷新, 它将在Get操作时自动刷新TTL
func (m *SpaceConfig) SetAutoRefresh(auto_refresh bool) {
    m.auto_ref = auto_refresh
}

// 设置自动刷新间隔, 它禁止高并发Get每次都触发自动刷新
func (m *SpaceConfig) SetAutoRefreshInterval(stamp time.Duration) {
    if stamp <= 0 {
        m.auto_ref_interval = DefaultAutoRefInterval
        return
    }
    m.auto_ref_interval = stamp
}

// 设置加载函数
func (m *SpaceConfig) SetLoadDBFn(fn LoadDBFn) *SpaceConfig {
    m.fn = fn
    return m
}

func (m *SpaceConfig) prepare_auto_ref(key string) bool {
    if !m.auto_ref {
        return false
    }

    m.mx.RLock()
    next, _ := m.next_ref_time[key]
    m.mx.RUnlock()

    now := time.Duration(time.Now().UnixNano())
    if now < next {
        return false
    }

    m.mx.Lock()
    m.next_ref_time[key] = now + m.auto_ref_interval
    m.mx.Unlock()
    return true
}

func (m *SpaceConfig) expirat() time.Duration {
    if m.rand && m.ex != m.endex {
        max := new(big.Int).SetInt64(int64(m.endex - m.ex))
        n, _ := rand.Int(rand.Reader, max)
        return time.Duration(n.Int64()) + m.ex
    }
    return m.ex
}

func (m *SpaceConfig) getLoadDBFn() LoadDBFn {
    return m.fn
}
