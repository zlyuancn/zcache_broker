/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/12/21
   Description :
-------------------------------------------------
*/

package wrpcx

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/smallnest/rpcx/client"
    "github.com/smallnest/rpcx/protocol"
    "github.com/smallnest/rpcx/share"
    "github.com/zlyuancn/zsingleflight"

    "github.com/zlyuancn/zcache_broker/transport/wrpcx/pb"
)

type Unmarshaler func(data []byte) (interface{}, error)

type Client struct {
    c        *client.XClientPool
    sf       *zsingleflight.SingleFlight // 单飞
    secretFn func() string               // 秘钥生成
}

type ClientOption struct {
    PoolSize       int               // 客户端池大小
    Address        []string          // 连接地址
    FailMode       client.FailMode   // 失败模式
    SelectMode     client.SelectMode // 负载均衡模式
    Retries        int               // 重试次数
    ConnectTimeout time.Duration     // 连接超时
    BackupLatency  time.Duration     // 故障转移模式等待时间
    SecretFn       func() string     // 秘钥生成函数
}

var DefaultClientOption = ClientOption{
    PoolSize:       10,
    FailMode:       client.Failover,
    SelectMode:     client.RoundRobin,
    Retries:        3,
    ConnectTimeout: 5 * time.Second,
    BackupLatency:  100 * time.Millisecond,
}

func NewClient(server_name string, opt ClientOption) *Client {
    if len(opt.Address) == 0 {
        panic("address是空的")
    }
    addrs := make([]*client.KVPair, len(opt.Address))
    for i, a := range opt.Address {
        addrs[i] = &client.KVPair{Key: a}
    }
    d := client.NewMultipleServersDiscovery(addrs)

    if opt.PoolSize <= 0 {
        opt.PoolSize = 10
    }

    option := client.Option{
        Retries:        opt.Retries,
        RPCPath:        share.DefaultRPCPath,
        ConnectTimeout: opt.ConnectTimeout,
        SerializeType:  protocol.ProtoBuffer,
        CompressType:   protocol.None,
        BackupLatency:  opt.BackupLatency,
    }

    c := client.NewXClientPool(opt.PoolSize, server_name, opt.FailMode, opt.SelectMode, d, option)
    m := &Client{
        c:        c,
        sf:       zsingleflight.New(),
        secretFn: opt.SecretFn,
    }
    return m
}

func (m *Client) getClient() client.XClient {
    c := m.c.Get()
    if m.secretFn != nil {
        c.Auth(m.secretFn())
    }
    return c
}

func (m *Client) Get(ctx context.Context, space string, key string, params ...string) ([]byte, error) {
    v, err := m.sf.Do(makeSFKey(space, key, params...), func() (interface{}, error) {
        resp := new(pb.GetResp)
        err := m.getClient().Call(ctx, "Get", &pb.GetReq{Space: space, Key: key, Params: params}, resp)
        if err != nil {
            return nil, err
        }
        return resp.Data, nil
    })
    if err != nil {
        return nil, err
    }

    return v.([]byte), nil
}

func (m *Client) GetAndUnmarshal(ctx context.Context, space string, key string, unmarshaler Unmarshaler, params ...string) (interface{}, error) {
    v, err := m.sf.Do(makeSFKey(space, key, params...), func() (interface{}, error) {
        resp := new(pb.GetResp)
        err := m.getClient().Call(ctx, "Get", &pb.GetReq{Space: space, Key: key, Params: params}, resp)
        if err != nil {
            return nil, err
        }
        return unmarshaler(resp.Data)
    })
    return v, err
}

func (m *Client) Del(ctx context.Context, space string, key string, params ...string) error {
    _, err := m.sf.Do(makeSFKey(space, key, params...), func() (interface{}, error) {
        resp := new(pb.DelResp)
        err := m.getClient().Call(ctx, "Del", &pb.DelReq{Space: space, Key: key, Params: params}, resp)
        return nil, err
    })
    return err
}

func (m *Client) Refresh(ctx context.Context, space string, key string, params ...string) ([]byte, error) {
    v, err := m.sf.Do(makeSFKey(space, key, params...), func() (interface{}, error) {
        resp := new(pb.RefreshResp)
        err := m.getClient().Call(ctx, "Refresh", &pb.RefreshReq{Space: space, Key: key, Params: params}, resp)
        if err != nil {
            return nil, err
        }
        return resp.Data, nil
    })
    if err != nil {
        return nil, err
    }

    return v.([]byte), nil
}

func (m *Client) RefreshAndUnmarshal(ctx context.Context, space string, key string, unmarshaler Unmarshaler, params ...string) (interface{}, error) {
    v, err := m.sf.Do(makeSFKey(space, key, params...), func() (interface{}, error) {
        resp := new(pb.RefreshResp)
        err := m.getClient().Call(ctx, "Refresh", &pb.RefreshReq{Space: space, Key: key, Params: params}, resp)
        if err != nil {
            return nil, err
        }
        return unmarshaler(resp.Data)
    })
    return v, err
}

func (m *Client) Close() {
    m.c.Close()
}

func makeSFKey(space, key string, params ...string) string {
    if len(params) != 0 {
        return fmt.Sprintf("%s:%s?%s", space, key, strings.Join(params, "&"))
    }
    return fmt.Sprintf("%s:%s", space, key)
}
