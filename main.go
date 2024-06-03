package main

import (
	"github.com/QinYuuuu/abvss/services"
)

type node struct {
	in  chan services.Message
	out chan services.Message
}

/*
	func main() {
		iplist := []string{"127.0.0.1:8000", "127.0.0.1:8001", "127.0.0.1:8002"}
		peers := make([]*network.Peer, 3)
		for i := 0; i < 3; i++ {
			peer, err := network.NewPeer(3, i, iplist)
			if err != nil {
				fmt.Println(err)
			}
			peers[i] = peer
			service := network.Service{Id: i}
			protobuf.RegisterConnServer(peers[i].Server, service)
			go peer.Serve(false)
		}
		clients := make([][]protobuf.ConnClient, 3)
		for i := 0; i < 3; i++ {
			peers[i].Connect()
			clients[i] = make([]protobuf.ConnClient, 3)
			for j := 0; j < 3; j++ {
				if j == i {
					continue
				}
				clients[i][j] = protobuf.NewConnClient(peers[i].Conns[j])
				rsp, err := clients[i][j].Receive(context.TODO(), &protobuf.TestHelloMessage{
					FromID:  int64(i),
					DestID:  int64(j),
					Content: "hello world",
				})
				if err != nil {
					fmt.Println(err)
				}
				log.Printf("node %v receive respond from node %v: %v", i, rsp.GetFromID(), rsp.GetContent())
			}
		}
		time.Sleep(3 * time.Second)
		for i := 0; i < 3; i++ {
			peers[i].Close()
		}
	}
*/
func main() {
	services.TestVSS()
}

/*
func main() {
	n := 4
	tnum := 1
	msgchans := make([]chan osv.Message, n)
	ch := make(chan osv.Message)
	osvs := make([]*osv.OSV, n)
	var wait sync.WaitGroup
	wait.Add(n)
	for i := 0; i < n; i++ {
		osvs[i] = osv.NewOSV(n, tnum, i)
		msgchans[i] = make(chan osv.Message, 10)
		go func(i int) {
			msgs := osvs[i].Init()
			fmt.Println(msgs)
			for _, msg := range msgs {
				msgchans[msg.Dest()] <- msg
			}
			fmt.Println("Init done")
			wait.Done()
		}(i)
	}
	//msgs := make([][]Message, n)
	wait.Wait()

	wait.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			for msg := range msgchans[i] {
				//fmt.Printf("message from %v to %v %v\n", msg.From(), msg.Dest(), msg.Type())
				replymsgs, err := osvs[i].Recv(msg)
				if err != nil {
					fmt.Printf("recv err: %v", err)
				}
				for _, replymsg := range replymsgs {
					ch <- replymsg
				}
				if osvs[i].Done() {
					wait.Done()
				}
			}
		}(i)
	}
	go func() {
		for replymsg := range ch {
			msgchans[replymsg.Dest()] <- replymsg
			fmt.Printf("out message %v\n", replymsg)
		}
	}()
	wait.Wait()
}
*/
