package network

import (
	pb "github.com/QinYuuuu/abvss/protobuf"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Peer struct {
	id     int
	server *grpc.Server
	ipList []string // Node IP Address List
}

func (p *Peer) Serve(aws bool) {
	addr := p.ipList[p.id]
	if aws {
		addr = "0.0.0.0:12001"
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("node failed to listen %v", err)
	}
	s := grpc.NewServer()
	pb.
		reflection.Register(s)
	log.Printf("node %d serve on %s", p.id, addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("node failed to serve %v", err)
	}
}
