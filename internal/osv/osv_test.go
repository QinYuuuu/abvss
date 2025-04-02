package osv

import (
	"fmt"
	"log/slog"
	"sync"
	"testing"
)

func TestOSV_Init(t *testing.T) {
	var n int64 = 4
	var tnum int64 = 1
	msgChans := make([]chan Message, n)
	osvs := make([]*Instance, n)
	var i int64
	for i = 0; i < n; i++ {
		osvs[i] = NewInstance(n, tnum, i, 0, nil, nil, nil)
		msgChans[i] = make(chan Message)
		go func(i int64) {
			msgs := osvs[i].init()
			fmt.Println(msgs)
			for _, msg := range msgs {
				msgChans[msg.Dest()] <- msg
			}
		}(i)
	}
	var wait sync.WaitGroup
	wait.Add(int((n - 1) * n))
	for i = 0; i < n; i++ {
		go func(i int64) {
			for msg := range msgChans[i] {
				fmt.Printf("message from %v to %v %v\n", msg.From(), msg.Dest(), msg.Type())
				wait.Done()
			}
		}(i)
	}
	wait.Wait()
}

func TestOSV_Recv(t *testing.T) {
	var n int64 = 4
	var tnum int64 = 1
	msgchans := make([]chan Message, n)
	ch := make(chan Message, 2048)
	osvs := make([]*Instance, n)
	var wait sync.WaitGroup
	wait.Add(int(n))
	var i int64
	for i = 0; i < n; i++ {
		osvs[i] = NewInstance(n, tnum, i, 0, nil, nil, nil)
		msgchans[i] = make(chan Message, 10)
		go func(i int64) {
			msgs := osvs[i].init()
			slog.Debug("get", slog.Any("message", msgs))
			for _, msg := range msgs {
				msgchans[msg.Dest()] <- msg
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
			msgchans[replyMsg.Dest()] <- replyMsg
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
	msgChans := make([]chan Message, n)
	osvInstances := make([]*Instance, n)
	var wait sync.WaitGroup
	wait.Add(int(n))
	var i int64
	for i = 0; i < n; i++ {
		msgChans[i] = make(chan Message, 10)
	}
	for i = 0; i < n; i++ {
		ID := i
		send := func(msg Message) {
			slog.Info("send", slog.Any("msg", msg))
			msgChans[i] <- msg
		}
		receive := func() chan Message {
			return msgChans[ID]
		}
		osvInstances[i] = NewInstance(n, tnum, i, 0, send, nil, receive)
	}
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
	msgChans := make([][]chan Message, n)
	osvInstances := make([][]*Instance, n)
	var wait sync.WaitGroup
	wait.Add(int(n * instanceNum))
	var i, j int64
	for i = 0; i < n; i++ {
		msgChans[i] = make([]chan Message, instanceNum)
		for j = 0; j < instanceNum; j++ {
			msgChans[i][j] = make(chan Message, 10)
		}
		osvInstances[i] = make([]*Instance, instanceNum)
	}
	for i = 0; i < n; i++ {
		ID := i
		for j = 0; j < instanceNum; j++ {
			instanceID := j
			send := func(msg Message) {
				slog.Info("send", slog.Any("msg", msg))
				msgChans[msg.DestID][msg.InstanceID] <- msg
			}
			receive := func() chan Message {
				return msgChans[ID][instanceID]
			}
			osvInstances[i][j] = NewInstance(n, tNum, i, instanceID, send, nil, receive)
		}

	}
	for i = 0; i < n; i++ {
		for j = 0; j < instanceNum; j++ {
			go osvInstances[i][j].Run()
		}

	}
	for i = 0; i < n; i++ {
		for j = 0; j < instanceNum; j++ {
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
