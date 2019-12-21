/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2019/12/21
   Description :
-------------------------------------------------
*/

package main

import (
    "context"
    "crypto/rand"
    "fmt"
    "github.com/zlyuancn/zcache_broker/transport/wrpcx"
    "math/big"
    "testing"
)

func Benchmark_A(b *testing.B) {
    space := "test"
    server_name := "test"

    // client
    copt := wrpcx.DefaultClientOption
    copt.Address = []string{"localhost:2333"}
    c := wrpcx.NewClient(server_name, copt)
    defer c.Close()

    b.ResetTimer()
    b.RunParallel(func(p *testing.PB) {
        i := 0
        for p.Next() {
            sr := new(big.Int).SetInt64(10000)
            n, _ := rand.Int(rand.Reader, sr)
            m := n.Int64()

            k := fmt.Sprintf("%d", m)
            bs, err := c.Get(context.Background(), space, k)
            if err != nil {
                b.Fatal(err)
            }
            if len(bs) != 512 {
                b.Fatalf("收到的值长度 %d[%d]:%d, 他应该是 512", i, m, len(bs))
            }
            i++
        }
    })
}
