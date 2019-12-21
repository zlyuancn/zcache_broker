/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2019/12/18
   Description :
-------------------------------------------------
*/

package wgrpc

import (
    "bytes"
    "context"
    "fmt"
    "github.com/go-redis/redis"
    "github.com/zlyuancn/zcache_broker"
    "github.com/zlyuancn/zcache_broker/cachedb/wredis"
    "google.golang.org/grpc"
    "net"
    "testing"
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

func TestGet(t *testing.T) {
    space := "test"

    // 空间配置
    sc := zcache_broker.NewSpaceConfig()
    for i := 0; i < 10; i++ {
        k := fmt.Sprintf("k%d", i)
        v := []byte(fmt.Sprintf("v%d", i))
        sc.SetLoadDBFn(k, func() (i []byte, e error) {
            return v, nil
        })
    }

    // cb
    cb := getTestClient()
    cb.SetOptions(zcache_broker.WithSpaceConf(space, sc))

    // server
    l, err := net.Listen("tcp", ":2333")
    if err != nil {
        t.Fatal(err)
    }
    s := NewServer(cb)
    defer s.Close()
    go func() {
        if err := s.Start(l); err != nil {
            t.Fatal(err)
        }
    }()

    // client
    conn, err := grpc.Dial("localhost:2333", grpc.WithInsecure())
    if err != nil {
        t.Fatal(err)
    }
    c := NewClient(conn)
    defer c.Close()

    // test
    for i := 0; i < 10; i++ {
        k := fmt.Sprintf("k%d", i)
        v := []byte(fmt.Sprintf("v%d", i))
        bs, err := c.Get(context.Background(), space, k)
        if err != nil {
            t.Fatal(err)
        }
        if !bytes.Equal(bs, v) {
            t.Fatalf("收到的值非预期 %s: %s", k, string(bs))
        }
    }
}
