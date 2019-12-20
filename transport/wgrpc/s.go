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
    "google.golang.org/grpc"
    "net"
)

type Server struct {
    obj zcache_broker.PublicInterface
    s   *grpc.Server
}

func (m *Server) Get(_ context.Context, req *pb.GetReq) (*pb.GetResp, error) {
    bs, err := m.obj.Get(req.Space, req.Key)
    if err != nil {
        return nil, err
    }
    return &pb.GetResp{Data: bs}, nil
}

func (m *Server) Del(_ context.Context, req *pb.DelReq) (*pb.DelResp, error) {
    err := m.obj.Del(req.Space, req.Key)
    if err != nil {
        return nil, err
    }
    return &pb.DelResp{}, nil
}

func (m *Server) Refresh(_ context.Context, req *pb.RefreshReq) (*pb.RefreshResp, error) {
    err := m.obj.Refresh(req.Space, req.Key)
    if err != nil {
        return nil, err
    }
    return &pb.RefreshResp{}, nil
}

func NewServer(obj zcache_broker.PublicInterface) *Server {
    s := grpc.NewServer()
    m := &Server{
        obj: obj,
        s:   s,
    }
    pb.RegisterRCBServiceServer(s, m)
    return m
}

func (m *Server) Start(l net.Listener) error {
    return m.s.Serve(l)
}

func (m *Server) Close() {
    m.s.Stop()
}
