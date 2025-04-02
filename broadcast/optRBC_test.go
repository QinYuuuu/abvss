package broadcast

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"
	"testing"
)

func TestOptRBC(t *testing.T) {
	// parameters
	n := int64(4)
	f := int64(1)
	sessionID := int64(0)
	testMessage := []byte("test")

	// make send in channel
	channels := make([]chan Message, n)
	for i := range channels {
		channels[i] = make(chan Message, 100)
	}

	// make send out channel
	outputs := make([][]chan []byte, n)
	for i := range outputs {
		outputs[i] = make([]chan []byte, 1)
		outputs[i][0] = make(chan []byte, 1)
	}

	// make rbc nodes
	nodes := make([]*OptRBC, n)
	for i := int64(0); i < n; i++ {
		pid := i
		send := func(dest int64, msg Message) {
			channels[dest] <- msg
		}
		receive := func() (Message, bool) {
			select {
			case msg := <-channels[pid]:
				return msg, true
			default:
				return Message{}, false
			}
		}

		nodes[i] = NewOptRBC(pid, n, f, send, outputs[i], receive)
		nodes[i].CreateNewSession(sessionID, 0) // 设置节点0为leader
	}

	// start all message
	for i := range nodes {
		nodes[i].Run()
	}

	// 开始广播
	nodes[0].StartNewBroadcast(testMessage, 0, sessionID)

	// 等待结果并验证
	results := make([][]byte, n)

	for i := range n {
		select {
		case results[i] = <-outputs[i][0]:
			slog.Info(fmt.Sprintf("[node %v] %s", i, results[i]))
			if !bytes.Equal(results[i], testMessage) {
				t.Errorf("node %d receive wrong message", i)
			}
		}
	}

	// 验证所有节点都收到了相同的消息
	for i := range n {
		if !bytes.Equal(results[i], testMessage) {
			t.Errorf("节点 %d 收到的消息与预期不符", i)
		}
	}
}

func TestOptRBC_TwoSessions(t *testing.T) {
	// 设置网络规模和容错参数
	n := int64(4)
	f := int64(1)

	// 创建消息通道
	msgChan := make(chan Message, 1000)

	// 创建节点
	nodes := make([]*OptRBC, n)
	outputs := make([][]chan []byte, n)

	// 为每个节点初始化输出通道 (两个会话)
	for i := int64(0); i < n; i++ {
		outputs[i] = make([]chan []byte, 2)
		for j := 0; j < 2; j++ {
			outputs[i][j] = make(chan []byte, 10)
		}
	}
	// 创建发送和接收函数
	for i := int64(0); i < n; i++ {
		pid := i
		sendFunc := func(dest int64, msg Message) {
			msgChan <- msg
		}
		recvFunc := func() (Message, bool) {
			select {
			case msg := <-msgChan:
				if msg.DestID == pid {
					return msg, true
				}
				// 将消息放回通道
				msgChan <- msg
			default:
				// 无消息可接收
			}
			return Message{}, false
		}
		nodes[i] = NewOptRBC(pid, n, f, sendFunc, outputs[i], recvFunc)
		// 创建两个会话
		nodes[i].CreateNewSession(0, 0) // 会话0，节点0为领导者
		nodes[i].CreateNewSession(1, 1) // 会话1，节点1为领导者
	}

	// 启动所有节点
	for i := int64(0); i < n; i++ {
		nodes[i].Run()
	}

	// 开始广播
	msg1 := []byte("Hello from session 0")
	msg2 := []byte("Hello from session 1")

	// 领导者发起广播
	nodes[0].StartNewBroadcast(msg1, 0, 0)
	nodes[1].StartNewBroadcast(msg2, 1, 1)

	// 验证所有节点都收到了两个会话的消息
	for i := range n {
		go func(i int64) {
			msg := <-outputs[i][0]
			//slog.Info(fmt.Sprintf("[node %v] %s", i, msg))
			if !bytes.Equal(msg, msg1) {
				t.Errorf("node %d receive wrong message", i)
			}
		}(i)
	}
	var wg sync.WaitGroup
	wg.Add(int(n))
	for i := range n {
		go func(i int64) {
			msg := <-outputs[i][1]
			slog.Info(fmt.Sprintf("[node %v] %s", i, msg))
			if !bytes.Equal(msg, msg2) {
				t.Errorf("node %d session 1 receive wrong message", i)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
