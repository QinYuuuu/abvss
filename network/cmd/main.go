package main

import (
	"fmt"
	"github.com/QinYuuuu/abvss/network"
)

func main() {
	iplist := []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"}
	portlist := []string{"8000", "8001", "8002"}
	peers := make([]*network.Peer, 3)
	for i := 0; i < 3; i++ {
		peer, err := network.NewPeer(3, i, iplist, portlist)
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
