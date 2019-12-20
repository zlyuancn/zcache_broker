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
    "github.com/zlyuancn/zcache_broker/transport/wgrpc/pb"
    "google.golang.org/grpc"
)

type Client struct {
    conn *grpc.ClientConn
    c    pb.RCBServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
    c := pb.NewRCBServiceClient(conn)
    m := &Client{
        conn: conn,
        c:    c,
    }
    return m
}

func (m *Client) Close() {
    _ = m.conn.Close()
}

func (m *Client) Get(ctx context.Context, space string, key string, opts ...grpc.CallOption) ([]byte, error) {
    resp, err := m.c.Get(ctx, &pb.GetReq{Space: space, Key: key}, opts ...)
    if err != nil {
        return nil, err
    }
    return resp.Data, nil
}

func (m *Client) Del(ctx context.Context, space string, key string, opts ...grpc.CallOption) error {
    _, err := m.c.Del(ctx, &pb.DelReq{Space: space, Key: key}, opts ...)
    return err
}

func (m *Client) Refresh(ctx context.Context, space string, key string, opts ...grpc.CallOption) error {
    _, err := m.c.Refresh(ctx, &pb.RefreshReq{Space: space, Key: key}, opts ...)
    return err
}
