package main

import (
<<<<<<< HEAD
<<<<<<< HEAD
	"fmt"
	"github.com/QinYuuuu/abvss/network"
=======
	"abvss/network"
	"fmt"
>>>>>>> 19b0d27 (Initial commit)
=======
	"abvss/network"
	"fmt"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
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
		go peer.Connect()
	}
	for {

	}
	for i := 0; i < 3; i++ {
		peers[i].Close()
	}
}
