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
    "time"
)

type SpaceConfig struct {
    rand     bool
    auto_ref bool
    ex       time.Duration
    endex    time.Duration
    fn       LoadDBFn
}

func NewSpaceConfig() *SpaceConfig {
    return &SpaceConfig{}
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

// 设置加载函数
func (m *SpaceConfig) SetLoadDBFn(fn LoadDBFn) *SpaceConfig {
    m.fn = fn
    return m
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
