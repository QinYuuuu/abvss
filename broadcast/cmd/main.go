package main

import (
	"log/slog"
	"sync"

	"github.com/QinYuuuu/abvss/broadcast"
)

func main() {
	var Nnodes int64 = 4
	var f int64 = 1
	messageMap := map[int64]chan broadcast.Message{
		0: make(chan broadcast.Message, 100),
		1: make(chan broadcast.Message, 100),
		2: make(chan broadcast.Message, 100),
		3: make(chan broadcast.Message, 100),
	}
	outputMap := map[int64][]chan []byte{
		0: {make(chan []byte)},
		1: {make(chan []byte)},
		2: {make(chan []byte)},
		3: {make(chan []byte)},
	}
	send := func(i int64, m broadcast.Message) {
		messageMap[i] <- m
	}
	nodes := []*broadcast.Node{
		broadcast.NewNode(0, int64(Nnodes), int64(f), send, outputMap[0], func() (broadcast.Message, bool) {
			msg, ok := <-messageMap[0]
			return msg, ok
		}),
		broadcast.NewNode(1, int64(Nnodes), int64(f), send, outputMap[1], func() (broadcast.Message, bool) {
			msg, ok := <-messageMap[1]
			return msg, ok
		}),
		broadcast.NewNode(2, int64(Nnodes), int64(f), send, outputMap[2], func() (broadcast.Message, bool) {
			msg, ok := <-messageMap[2]
			return msg, ok
		}),
		broadcast.NewNode(3, int64(Nnodes), int64(f), send, outputMap[2], func() (broadcast.Message, bool) {
			msg, ok := <-messageMap[3]
			return msg, ok
		}),
	}
	nodes[0].Run()
	nodes[1].Run()
	nodes[2].Run()
	nodes[3].Run()
	nodes[0].CreateNewSession(0, 0)
	nodes[1].CreateNewSession(0, 0)
	nodes[2].CreateNewSession(0, 0)
	nodes[3].CreateNewSession(0, 0)
	payload := []byte("test")
	nodes[0].StartNewBroadcast(payload, 0, 4, 1, 0)
	var wg sync.WaitGroup
	wg.Add(4)
	for i := 0; i < 4; i++ {
		go func(i int) {
			output := nodes[i].Output(0)
			slog.Info("output", slog.String("data", string(output)))
			wg.Done()
		}(i)
	}
	wg.Wait()
}
