/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2019/12/18
   Description :
-------------------------------------------------
*/

package wrpcx

import (
    "bytes"
    "context"
    "fmt"
    "github.com/go-redis/redis"
    "github.com/zlyuancn/zcache_broker"
    "github.com/zlyuancn/zcache_broker/cachedb/wredis"
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
    sc.SetLoadDBFn(func(space, key string) ([]byte, error) {
        return []byte(key), nil
    })

    // cb
    cb := getTestClient()
    cb.SetOptions(zcache_broker.WithSpaceConf(space, sc))

    server_name := "test"
    // server
    s := NewServer(server_name, cb)
    defer s.Close()
    go func() {
        if err := s.Start(":2333"); err != nil {
            t.Fatal(err)
        }
    }()

    // client
    copt := DefaultClientOption
    copt.Address = []string{"localhost:2333"}
    c := NewClient(server_name, copt)
    defer c.Close()

    // test
    for i := 0; i < 10; i++ {
        k := fmt.Sprintf("%d", i)
        v := []byte(fmt.Sprintf("%d", i))
        bs, err := c.Get(context.Background(), space, k)
        if err != nil {
            t.Fatal(err)
        }
        if !bytes.Equal(bs, v) {
            t.Fatalf("收到的值非预期 %s: %s", k, string(bs))
        }
    }
}
