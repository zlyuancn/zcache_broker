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
    "net"

    "github.com/smallnest/rpcx/protocol"
    "github.com/smallnest/rpcx/server"

    "github.com/zlyuancn/zcache_broker"
    "github.com/zlyuancn/zcache_broker/transport/wrpcx/pb"
)

type serverObj struct {
    obj zcache_broker.PublicInterface
}

func (m *serverObj) Get(_ context.Context, req *pb.GetReq, resp *pb.GetResp) error {
    bs, err := m.obj.Get(req.Space, req.Key, req.Params...)
    if err != nil {
        return err
    }
    resp.Data = bs
    return nil
}

func (m *serverObj) Del(_ context.Context, req *pb.DelReq, resp *pb.DelResp) error {
    err := m.obj.Del(req.Space, req.Key, req.Params...)
    if err != nil {
        return err
    }
    return nil
}

func (m *serverObj) Refresh(_ context.Context, req *pb.RefreshReq, resp *pb.RefreshResp) error {
    bs, err := m.obj.Refresh(req.Space, req.Key, req.Params...)
    if err != nil {
        return err
    }
    resp.Data = bs
    return nil
}

type Server struct {
    name     string
    authFunc func(ctx context.Context, req *protocol.Message, token string) error
    o        *serverObj
    s        *server.Server
}

func NewServer(server_name string, obj zcache_broker.PublicInterface) *Server {
    s := server.NewServer()
    o := &serverObj{obj: obj}
    m := &Server{
        name: server_name,
        s:    s,
        o:    o,
    }
    return m
}

func (m *Server) SetAuthFunc(authFunc func(ctx context.Context, req *protocol.Message, token string) error) {
    m.s.AuthFunc = authFunc
}

func (m *Server) Start(address string) error {
    return m.serve("tcp", address)
}

func (m *Server) StartWithListener(l net.Listener) error {
    server.RegisterMakeListener(m.name, func(s *server.Server, _ string) (net.Listener, error) {
        return l, nil
    })
    return m.serve(m.name, "")
}

func (m *Server) serve(network, address string) error {
    if err := m.s.RegisterName(m.name, m.o, ""); err != nil {
        return err
    }

    err := m.s.Serve(network, address)
    if err == server.ErrServerClosed {
        return nil
    }
    return err
}

func (m *Server) Close() error {
    return m.s.Close()
}
