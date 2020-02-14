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
    "testing"
    "time"

    "github.com/go-redis/redis"

    "github.com/zlyuancn/zcache_broker/cachedb/wredis"
)

func getTestClient() *CacheBroker {
    c := redis.NewClient(&redis.Options{
        Addr:        "localhost:6379",
        Password:    "",
        DB:          0,
        PoolSize:    1,
        DialTimeout: time.Second * 3,
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
    c.SetOptions(WithSpaceConf("test", NewSpaceConfig().SetLoadDBFn(func(space, key string, params ...string) (i []byte, err error) {
        return []byte(fmt.Sprintf("v%s", key)), nil
    })))

    for i := 0; i < 10; i++ {
        k := fmt.Sprintf("%d", i)
        v := []byte(fmt.Sprintf("v%d", i))
        bs, err := c.Get(space, k)
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
    sconfig := NewSpaceConfig().SetExpirat(time.Millisecond*100, true)
    c.SetOptions(WithSpaceConf(space, sconfig))

    k := "k0"
    v1 := []byte("v0")

    sconfig.SetLoadDBFn(func(space, key string, params ...string) (i []byte, err error) {
        return v1, nil
    })
    bs, err := c.Get(space, k)
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(bs, v1) {
        t.Fatalf("收到的值非预期 %s: %s", k, string(bs))
    }

    time.Sleep(time.Millisecond * 200)
    v2 := []byte("vr")
    sconfig.SetLoadDBFn(func(space, key string, params ...string) (i []byte, err error) {
        return v2, nil
    })
    bs, err = c.Get(space, k)
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(bs, v2) {
        t.Fatalf("收到的值非预期 %s: %s", k, string(bs))
    }
}

func TestGetWithParams(t *testing.T) {
    space := "test"
    c := getTestClient()
    c.SetOptions(WithSpaceConf("test", NewSpaceConfig().SetLoadDBFn(func(space, key string, params ...string) (i []byte, err error) {
        if params[0] == "a" {
            return []byte("a"), nil
        }
        return []byte("b"), nil
    })))

    bs, err := c.Get(space, "key", "a")
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(bs, []byte("a")) {
        t.Fatalf("收到的值非预期 %s", string(bs))
    }

    bs, err = c.Get(space, "key", "b")
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(bs, []byte("b")) {
        t.Fatalf("收到的值非预期 %s", string(bs))
    }

}
