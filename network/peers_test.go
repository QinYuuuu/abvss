package network

import (
	"context"
	"fmt"
	"github.com/QinYuuuu/abvss/protobuf"
	"testing"
	"time"
)

func TestPeer_Serve(t *testing.T) {
	iplist := []string{"127.0.0.1:8000", "127.0.0.1:8001", "127.0.0.1:8002"}
	peers := make([]*Peer, 3)
	for i := 0; i < 3; i++ {
		peer, err := NewPeer(3, i, iplist)
		if err != nil {
			fmt.Println(err)
		}
		peers[i] = peer
		service := Service{Id: i}
		protobuf.RegisterConnServiceServer(peers[i].Server, service)
		go peer.Serve(false)
	}
	for i := 0; i < 3; i++ {
		peers[i].Connect()
	}
	for i := 0; i < 3; i++ {
		peers[i].Close()
	}
	time.Sleep(2 * time.Second)
}

func TestPeer_Connect(t *testing.T) {
	iplist := []string{"127.0.0.1:8000", "127.0.0.1:8001", "127.0.0.1:8002"}
	peers := make([]*Peer, 3)
	for i := 0; i < 3; i++ {
		peer, err := NewPeer(3, i, iplist)
		if err != nil {
			fmt.Println(err)
		}
		peers[i] = peer
		service := Service{Id: i}
		protobuf.RegisterConnServiceServer(peers[i].Server, service)
		go peer.Serve(false)
	}
	clients := make([][]protobuf.ConnServiceClient, 3)
	for i := 0; i < 3; i++ {
		peers[i].Connect()
		clients[i] = make([]protobuf.ConnServiceClient, 3)
		for j := 0; j < 3; j++ {
			if j == i {
				continue
			}
			clients[i][j] = protobuf.NewConnServiceClient(peers[i].Conns[j])
			rsp, err := clients[i][j].Receive(context.TODO(), &protobuf.TestMessage{
				FromID: int64(i),
				DestID: int64(j),
			})
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(rsp)
		}
	}

	for i := 0; i < 3; i++ {
		peers[i].Close()
	}
}
