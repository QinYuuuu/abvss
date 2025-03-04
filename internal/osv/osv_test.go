package osv

import (
	"fmt"
	"sync"
	"testing"
)

func TestOSV_Init(t *testing.T) {
	n := 4
	tnum := 1
	msgchans := make([]chan Message, n)
	osvs := make([]*OSV, n)
	for i := 0; i < n; i++ {
		osvs[i] = NewOSV(n, tnum, i)
		msgchans[i] = make(chan Message)
		go func(i int) {
			msgs := osvs[i].Init()
			fmt.Println(msgs)
			for _, msg := range msgs {
				msgchans[msg.Dest()] <- msg
			}
		}(i)
	}
	var wait sync.WaitGroup
	wait.Add(n * n)
	for i := 0; i < n; i++ {
		go func(i int) {
			for msg := range msgchans[i] {
				fmt.Printf("message from %v to %v %v\n", msg.From(), msg.Dest(), msg.Type())
				wait.Done()
			}
		}(i)
	}
	wait.Wait()
}

func TestOSV_Recv(t *testing.T) {
	n := 4
	tnum := 1
	msgchans := make([]chan Message, n)
	ch := make(chan Message)
	osvs := make([]*OSV, n)
	var wait sync.WaitGroup
	wait.Add(n)
	for i := 0; i < n; i++ {
		osvs[i] = NewOSV(n, tnum, i)
		msgchans[i] = make(chan Message, 10)
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
