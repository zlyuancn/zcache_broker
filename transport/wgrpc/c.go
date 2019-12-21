/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2019/12/18
   Description :
-------------------------------------------------
*/

package wgrpc

import (
    "context"
    "github.com/zlyuancn/zcache_broker"
    "github.com/zlyuancn/zcache_broker/transport/wgrpc/pb"
    "github.com/zlyuancn/zsingleflight"
    "google.golang.org/grpc"
)

type Client struct {
    conn *grpc.ClientConn
    c    pb.CBServiceClient
    sf   *zsingleflight.SingleFlight // 单飞
}

func NewClient(conn *grpc.ClientConn) *Client {
    c := pb.NewCBServiceClient(conn)
    m := &Client{
        conn: conn,
        c:    c,
        sf:   zsingleflight.New(),
    }
    return m
}

func (m *Client) Close() {
    _ = m.conn.Close()
}

func (m *Client) Get(ctx context.Context, space string, key string, opts ...grpc.CallOption) ([]byte, error) {
    v, err := m.sf.Do(zcache_broker.MakeKey(space, key), func() (interface{}, error) {
        resp, err := m.c.Get(ctx, &pb.GetReq{Space: space, Key: key}, opts ...)
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

func (m *Client) Del(ctx context.Context, space string, key string, opts ...grpc.CallOption) error {
    _, err := m.sf.Do(zcache_broker.MakeKey(space, key), func() (interface{}, error) {
        _, err := m.c.Del(ctx, &pb.DelReq{Space: space, Key: key}, opts ...)
        return nil, err
    })
    return err
}

func (m *Client) Refresh(ctx context.Context, space string, key string, opts ...grpc.CallOption) error {
    _, err := m.sf.Do(zcache_broker.MakeKey(space, key), func() (interface{}, error) {
        _, err := m.c.Refresh(ctx, &pb.RefreshReq{Space: space, Key: key}, opts ...)
        return nil, err
    })
    return err
}
