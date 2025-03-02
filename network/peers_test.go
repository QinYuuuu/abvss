package network

import (
	"fmt"
	"testing"
	"time"
)

func TestPeer_Serve(t *testing.T) {
	iplist := []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"}
	portlist := []string{"8000", "8001", "8002"}
	peers := make([]*Peer, 3)
	for i := 0; i < 3; i++ {
		peer, err := NewPeer(3, i, iplist, portlist)
		if err != nil {
			fmt.Println(err)
		}
		peers[i] = peer
		peers[i] = peer
		go peer.Serve()
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
	iplist := []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"}
	portlist := []string{"8000", "8001", "8002"}
	peers := make([]*Peer, 3)
	for i := 0; i < 3; i++ {
		peer, err := NewPeer(3, i, iplist, portlist)
		if err != nil {
			fmt.Println(err)
		}
		peers[i] = peer
		go peer.Serve()
		peer.Connect()
	}
	for i := 0; i < 3; i++ {
		peers[i].Close()
	}
}
