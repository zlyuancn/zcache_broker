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
    "github.com/smallnest/rpcx/server"
    "github.com/zlyuancn/zcache_broker"
    "github.com/zlyuancn/zcache_broker/transport/wrpcx/pb"
    "net"
)

type Server struct {
    name string
    o    *serverObj
    s    *server.Server
    done chan struct{}
}
type serverObj struct {
    obj zcache_broker.PublicInterface
}

func (m *serverObj) Get(_ context.Context, req *pb.GetReq, resp *pb.GetResp) error {
    bs, err := m.obj.Get(req.Space, req.Key)
    if err != nil {
        return err
    }
    resp.Data = bs
    return nil
}

func (m *serverObj) Del(_ context.Context, req *pb.DelReq, resp *pb.DelResp) error {
    err := m.obj.Del(req.Space, req.Key)
    if err != nil {
        return err
    }
    return nil
}

func (m *serverObj) Refresh(_ context.Context, req *pb.RefreshReq, resp *pb.RefreshResp) error {
    err := m.obj.Refresh(req.Space, req.Key)
    if err != nil {
        return err
    }
    return nil
}

func NewServer(server_name string, obj zcache_broker.PublicInterface) *Server {
    s := server.NewServer()
    o := &serverObj{obj: obj}
    m := &Server{
        name: server_name,
        s:    s,
        o:    o,
        done: make(chan struct{}, 1),
    }
    return m
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
    select {
    case <-m.done:
        return nil
    default:
        return err
    }
}

func (m *Server) Close() {
    if len(m.done) == 0 {
        m.done <- struct{}{}
        _ = m.s.Close()
    }
}
