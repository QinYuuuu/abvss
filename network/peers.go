package network

import (
	"errors"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"sync"
)

type Peer struct {
	n, id  int
	Server *grpc.Server
	Conns  []*grpc.ClientConn
	ipList []string // Node IP Address List
	Ready  bool
}

func NewPeer(n, id int, iplist []string) (*Peer, error) {
	if n != len(iplist) {
		return nil, errors.New("n does not match iplist ")
	}
	return &Peer{
		n:      n,
		id:     id,
		Conns:  make([]*grpc.ClientConn, n),
		ipList: iplist,
		Server: grpc.NewServer(),
		Ready:  false,
	}, nil
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
	log.Printf("node %d serve on %s", p.id, addr)
	if err := p.Server.Serve(lis); err != nil {
		log.Fatalf("node failed to serve %v", err)
	}
	defer func() {
		p.Server.Stop()
		err := lis.Close()
		if err != nil {
			log.Printf("node failed to close listen %v", err)
		}
	}()
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
				nConn, err := grpc.NewClient(p.ipList[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Printf("node %v did not connect to node %v: %v", p.id, i, err)
					continue
				}
				flag = true
				p.Conns[i] = nConn
				log.Printf("node %v connect to node %v", p.id, i)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (p *Peer) Close() {
	for i, Conn := range p.Conns {
		if i == p.id {
			continue
		}
		err := Conn.Close()
		if err != nil {
			log.Printf("node %v close %v", p.id, err)
			continue
		}
		log.Printf("node %v close success", p.id)
	}
}

type Service struct {
	Id int
	protobuf.UnimplementedConnServer
}

func (n Service) Receive(ctx context.Context, req *protobuf.TestHelloMessage) (*protobuf.TestResMessage, error) {
	log.Printf("node %v receive request from node %v: %v", n.Id, req.GetFromID(), req.GetContent())
	return &protobuf.TestResMessage{Content: "have received Hello", FromID: int64(n.Id), DestID: req.GetFromID()}, nil
}
