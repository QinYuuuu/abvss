package network

import (
	"errors"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

type Peer struct {
	n, id  int
	Server *grpc.Server
	nConn  []*grpc.ClientConn
	ipList []string // Node IP Address List
}

func NewPeer(n, id int, iplist []string) (*Peer, error) {
	if n != len(iplist) {
		return nil, errors.New("n does not match iplist ")
	}
	return &Peer{n: n, id: id, ipList: iplist}, nil
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
	p.Server = grpc.NewServer()
	log.Printf("node %d serve on %s", p.id, addr)
	if err := p.Server.Serve(lis); err != nil {
		log.Fatalf("node failed to serve %v", err)
	}
}

func (p *Peer) Connect() {
	var wg sync.WaitGroup
	wg.Add(p.n - 1)
	for i := 0; i < len(p.ipList); i++ {
		if i == p.id {
			continue
		}
		go func(i int) {
			flag := false
			for !flag {
				nConn, err := grpc.NewClient(p.ipList[i])
				if err != nil {
					log.Printf("node did not connect to node: %v", err)
					continue
				}
				flag = true
				p.nConn[i] = nConn
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
