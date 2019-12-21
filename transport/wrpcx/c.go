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
    "github.com/smallnest/rpcx/client"
    "github.com/smallnest/rpcx/protocol"
    "github.com/smallnest/rpcx/share"
    "github.com/zlyuancn/zcache_broker/transport/wrpcx/pb"
    "google.golang.org/grpc"
    "time"
)

type Client struct {
    conn *grpc.ClientConn
    c    *client.XClientPool
}

type ClientOption struct {
    PoolSize       int               // 客户端池大小
    Address        []string          // 连接地址
    FailMode       client.FailMode   // 失败模式
    SelectMode     client.SelectMode // 负载均衡模式
    Retries        int               // 重试次数
    ConnectTimeout time.Duration     // 连接超时
    BackupLatency  time.Duration     // 故障转移模式等待时间
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
        c: c,
    }
    return m
}

func (m *Client) Get(ctx context.Context, space string, key string) ([]byte, error) {
    resp := new(pb.GetResp)
    err := m.c.Get().Call(ctx, "Get", &pb.GetReq{Space: space, Key: key}, resp)
    if err != nil {
        return nil, err
    }
    return resp.Data, nil
}

func (m *Client) Del(ctx context.Context, space string, key string) error {
    resp := new(pb.DelResp)
    err := m.c.Get().Call(ctx, "Del", &pb.DelReq{Space: space, Key: key}, resp)
    return err
}

func (m *Client) Refresh(ctx context.Context, space string, key string) error {
    resp := new(pb.RefreshResp)
    err := m.c.Get().Call(ctx, "Refresh", &pb.RefreshReq{Space: space, Key: key}, resp)
    return err
}

func (m *Client) Close() {
    m.c.Close()
}
