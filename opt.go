/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/12/18
   Description :
-------------------------------------------------
*/

package zcache_broker

type Option func(m *CacheBroker)

// 设置命名空间配置
func WithSpaceConf(space string, conf *SpaceConfig) Option {
    return func(m *CacheBroker) {
        m.spaces[space] = conf
    }
}

// 设置一些命名空间配置
func WithSpaceConfs(mm map[string]*SpaceConfig) Option {
    return func(m *CacheBroker) {
        for space, conf := range mm {
            m.spaces[space] = conf
        }
    }
}
