package osv

import (
	"fmt"
	"log/slog"
	"sync"
	"testing"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
)

func TestOSV_Init(t *testing.T) {
	var n int64 = 4
	var tnum int64 = 1
	msgChans := make([]chan *protobuf.OSVMessage, n)
	osvs := make([]*Instance, n)
	var i int64
	for i = 0; i < n; i++ {
		osvs[i] = NewInstance(n, tnum, i, "0", nil, nil)
		msgChans[i] = make(chan *protobuf.OSVMessage)
		go func(i int64) {
			msgs := osvs[i].init()
			fmt.Println(msgs)
			for _, msg := range msgs {
				msgChans[msg.GetDestID()] <- msg
			}
		}(i)
	}
	var wait sync.WaitGroup
	wait.Add(int((n - 1) * n))
	for i = 0; i < n; i++ {
		go func(i int64) {
			for msg := range msgChans[i] {
				fmt.Printf("message from %v to %v %v\n", msg.GetFromID(), msg.GetDestID(), msg.GetMsgType())
				wait.Done()
			}
		}(i)
	}
	wait.Wait()
}

func TestOSV_Recv(t *testing.T) {
	var n int64 = 4
	var tnum int64 = 1
	msgchans := make([]chan *protobuf.OSVMessage, n)
	ch := make(chan *protobuf.OSVMessage, 2048)
	osvs := make([]*Instance, n)
	var wait sync.WaitGroup
	wait.Add(int(n))
	var i int64
	for i = 0; i < n; i++ {
		osvs[i] = NewInstance(n, tnum, i, "0", nil, nil)
		msgchans[i] = make(chan *protobuf.OSVMessage, 10)
		go func(i int64) {
			msgs := osvs[i].init()
			slog.Debug("get", slog.Any("message", msgs))
			for _, msg := range msgs {
				msgchans[msg.GetDestID()] <- msg
			}
			slog.Debug("Init done")
			wait.Done()
		}(i)
	}
	//msgs := make([][]Message, n)
	wait.Wait()

	wait.Add(int(n))
	for i = 0; i < n; i++ {
		go func(i int64) {
			for msg := range msgchans[i] {
				//fmt.Printf("message from %v to %v %v\n", msg.From(), msg.Dest(), msg.Type())
				replyMsgs, err := osvs[i].recv(msg)
				if err != nil {
					slog.Debug(fmt.Sprintf("recv err: %v\n", err))
				}
				for _, replyMsg := range replyMsgs {
					ch <- replyMsg
				}

			}
		}(i)
	}
	go func() {
		for replyMsg := range ch {
			msgchans[replyMsg.GetDestID()] <- replyMsg
			slog.Debug(fmt.Sprintf("out message %v\n", replyMsg))
		}
	}()
	for i = 0; i < n; i++ {
		go func(i int64) {
			output := <-osvs[i].Output()
			if output {
				wait.Done()
			}
		}(i)
	}
	wait.Wait()
}

func TestOSV_Run(t *testing.T) {
	var n int64 = 4
	var tnum int64 = 1
	var i int64 = 0
	osvInstances := InitLocalMulti(n, tnum, "0")
	var wait sync.WaitGroup
	wait.Add(int(n))
	for i = 0; i < n; i++ {
		go osvInstances[i].Run()
	}
	for i = 0; i < n; i++ {
		go func(i int64) {
			output := <-osvInstances[i].Output()
			if output {
				wait.Done()
			}
		}(i)
	}
	wait.Wait()
}

func TestOSV_Run_2Instance(t *testing.T) {
	var n int64 = 4
	var tNum int64 = 1
	var instanceNum int64 = 2
	var i, j int64
	osvInstances := make([][]*Instance, instanceNum)
	for i = 0; i < instanceNum; i++ {
		osvInstances[i] = InitLocalMulti(n, tNum, fmt.Sprintf("%v", i))
	}
	var wait sync.WaitGroup
	wait.Add(int(n * instanceNum))
	for i = 0; i < instanceNum; i++ {
		for j = 0; j < n; j++ {
			go osvInstances[i][j].Run()
		}

	}
	for i = 0; i < instanceNum; i++ {
		for j = 0; j < n; j++ {
			go func(i, j int64) {
				output := <-osvInstances[i][j].Output()
				if output {
					wait.Done()
				}
			}(i, j)
		}
	}
	wait.Wait()
}
