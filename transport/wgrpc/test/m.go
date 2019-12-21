/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2019/12/21
   Description :
-------------------------------------------------
*/

package main

import (
    "crypto/rand"
    "github.com/go-redis/redis"
    "github.com/zlyuancn/zcache_broker"
    "github.com/zlyuancn/zcache_broker/cachedb/wredis"
    "github.com/zlyuancn/zcache_broker/transport/wgrpc"
    "log"
    "math/big"
    "net"
)

func getTestClient() *zcache_broker.CacheBroker {
    c := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        PoolSize: 20,
    })

    cb, err := zcache_broker.New(wredis.Wrap(c))
    if err != nil {
        panic(err)
    }
    return cb
}

func main() {
    space := "test"

    // 空间配置
    sc := zcache_broker.NewSpaceConfig()
    sc.SetLoadDBFn(func(space, key string) ([]byte, error) {
        value := make([]byte, 512)
        sr := new(big.Int).SetInt64(256)
        for i := 0; i < len(value); i++ {
            n, _ := rand.Int(rand.Reader, sr)
            s := n.Int64()
            value[i] = byte(s)
        }
        return value, nil
    })

    // cb
    cb := getTestClient()
    cb.SetOptions(zcache_broker.WithSpaceConf(space, sc))

    l, err := net.Listen("tcp", ":2333")
    if err != nil {
        log.Fatal(err)
    }
    // server
    s := wgrpc.NewServer(cb)
    defer s.Close()
    if err := s.Start(l); err != nil {
        log.Fatal(err)
    }
}
