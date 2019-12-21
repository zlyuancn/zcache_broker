/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/12/18
   Description :
-------------------------------------------------
*/

package zcache_broker

import (
    "bytes"
    "fmt"
    "github.com/go-redis/redis"
    "github.com/zlyuancn/zcache_broker/cachedb/wredis"
    "testing"
    "time"
)

func getTestClient() *CacheBroker {
    c := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        PoolSize: 20,
    })
    cb, err := New(wredis.Wrap(c))
    if err != nil {
        panic(err)
    }
    return cb
}

func TestGetAndCache(t *testing.T) {
    space := "test"
    c := getTestClient()
    for i := 0; i < 10; i++ {
        k := fmt.Sprintf("k%d", i)
        v := []byte(fmt.Sprintf("v%d", i))
        bs, err := c.GetWithFn(space, k, func() (bytes []byte, e error) {
            return v, nil
        })
        if err != nil {
            t.Fatal(err)
        }
        if !bytes.Equal(bs, v) {
            t.Fatalf("收到的值非预期 %s: %s", k, string(bs))
        }

        _, err = c.Get(space, k)
        if err != nil {
            t.Fatal(err)
        }
    }
}

func TestWithSpaceExpiration(t *testing.T) {
    c := getTestClient()
    space := "test"
    c.SetOptions(WithSpaceConf(
        space,
        NewSpaceConfig().SetExpirat(time.Millisecond*100, true),
    ))

    k := "k0"
    v := []byte("v0")
    bs, err := c.GetWithFn(space, k, func() (bytes []byte, e error) {
        return v, nil
    })
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(bs, v) {
        t.Fatalf("收到的值非预期 %s: %s", k, string(bs))
    }

    time.Sleep(time.Millisecond * 200)
    v = []byte("vr")
    bs, err = c.GetWithFn(space, k, func() (bytes []byte, e error) {
        return v, nil
    })
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(bs, v) {
        t.Fatalf("收到的值非预期 %s: %s", k, string(bs))
    }
}
