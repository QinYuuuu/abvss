package vaba

import (
	"bytes"
	"fmt"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"sync"
	"testing"
	"time"
)

// TestRAImpl_BasicFunctionality 测试 RAImpl 的基本功能
func TestRAImpl_BasicFunctionality(t *testing.T) {
	// 设置测试参数
	n := int64(4) // 总节点数
	f := int64(1) // 最大容错数

	// 创建消息通道
	msgChannels := make([]chan *protobuf.RAMessage, n)
	for i := range msgChannels {
		msgChannels[i] = make(chan *protobuf.RAMessage, 100)
	}

	// 创建节点
	nodes := make([]*RAImpl, n)
	for i := int64(0); i < n; i++ {
		nodes[i] = NewRAImpl(i, n, f, "test-instance-1")

		// 设置发送和接收函数
		nodeID := i
		nodes[i].send = func(msg *protobuf.RAMessage) {
			destID := msg.DestID
			// 将消息发送到目标节点的通道
			select {
			case msgChannels[destID] <- msg:
			default:
				t.Errorf("Channel full when sending from %d to %d", nodeID, destID)
			}
		}

		nodes[i].receive = func() chan *protobuf.RAMessage {
			return msgChannels[nodeID]
		}
	}

	// 启动所有节点的处理循环
	var wg sync.WaitGroup
	for i := int64(0); i < n; i++ {
		wg.Add(1)
		go func(idx int64) {
			defer wg.Done()
			nodes[idx].Run()
		}(i)
	}

	// 节点0输入消息
	testInput := []byte("test message")
	for i := int64(0); i < n; i++ {
		nodes[i].Input(testInput)
	}

	// 等待所有节点输出
	outputs := make([][]byte, n)
	var outputWg sync.WaitGroup

	for i := int64(0); i < n; i++ {
		outputWg.Add(1)
		go func(idx int64) {
			defer outputWg.Done()
			output := <-nodes[idx].Output()
			outputs[idx] = output
			t.Logf("Node %d received output: %s", idx, string(output))
		}(i)
	}

	// 等待输出完成或超时
	outputDone := make(chan bool)
	go func() {
		outputWg.Wait()
		outputDone <- true
	}()

	select {
	case <-outputDone:
		// 所有输出都已收到
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for all outputs")
	}

	// 验证所有节点的输出是否一致
	for i := int64(0); i < n; i++ {
		if !bytes.Equal(outputs[i], testInput) {
			t.Errorf("Node %d output mismatch: expected %s, got %s",
				i, string(testInput), string(outputs[i]))
		}
	}
}

// TestRAImpl_ByzantineNode 测试存在拜占庭节点的情况
func TestRAImpl_ByzantineNode(t *testing.T) {
	// 设置测试参数
	n := int64(4) // 总节点数
	f := int64(1) // 最大容错数

	// 创建消息通道
	msgChannels := make([]chan *protobuf.RAMessage, n)
	for i := range msgChannels {
		msgChannels[i] = make(chan *protobuf.RAMessage, 100)
	}

	// 创建节点
	nodes := make([]*RAImpl, n)
	for i := int64(0); i < n; i++ {
		nodes[i] = NewRAImpl(i, n, f, "test-instance-2")

		// 设置发送和接收函数
		nodeID := i
		nodes[i].send = func(msg *protobuf.RAMessage) {
			// 模拟节点3是拜占庭节点，它会发送错误消息
			if nodeID == 3 {
				// 拜占庭节点发送不同的消息给不同的节点
				for j := int64(0); j < n; j++ {
					if j == nodeID {
						continue // 不发给自己
					}
					byzantineMsg := &protobuf.RAMessage{
						FromID:     nodeID,
						DestID:     j,
						InstanceID: msg.InstanceID,
						Type:       msg.Type,
						Value:      []byte(fmt.Sprintf("byzantine message for %d", j)),
					}
					select {
					case msgChannels[j] <- byzantineMsg:
						// 消息成功发送
					default:
						t.Errorf("Channel full when sending from %d to %d", nodeID, j)
					}
				}
			} else {
				// 正常节点正常发送
				destID := msg.DestID
				select {
				case msgChannels[destID] <- msg:
					// 消息成功发送
				default:
					t.Errorf("Channel full when sending from %d to %d", nodeID, destID)
				}
			}
		}
		nodes[i].receive = func() chan *protobuf.RAMessage {
			return msgChannels[nodeID]
		}
	}

	// 启动所有节点的处理循环
	var wg sync.WaitGroup
	for i := int64(0); i < n; i++ {
		wg.Add(1)
		go func(idx int64) {
			defer wg.Done()
			nodes[idx].Run()
		}(i)
	}

	// 节点0输入消息
	testInput := []byte("test message")
	for i := int64(0); i < n; i++ {
		nodes[i].Input(testInput)
	}

	// 等待诚实节点输出
	outputs := make([][]byte, n-1) // 不包括拜占庭节点
	var outputWg sync.WaitGroup

	for i := int64(0); i < n-1; i++ { // 只等待诚实节点
		outputWg.Add(1)
		go func(idx int64) {
			defer outputWg.Done()
			select {
			case output := <-nodes[idx].Output():
				outputs[idx] = output
				t.Logf("Node %d received output: %s", idx, string(output))
			case <-time.After(3 * time.Second):
				t.Errorf("Timeout waiting for output from node %d", idx)
			}
		}(i)
	}

	// 等待输出完成或超时
	outputDone := make(chan bool)
	go func() {
		outputWg.Wait()
		outputDone <- true
	}()

	select {
	case <-outputDone:
		// 所有输出都已收到
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for all outputs")
	}

	// 验证所有诚实节点的输出是否一致
	for i := int64(0); i < n-1; i++ {
		if !bytes.Equal(outputs[i], testInput) {
			t.Errorf("Node %d output mismatch: expected %s, got %s",
				i, string(testInput), string(outputs[i]))
		}
	}
}

// TestRAImpl_ConcurrentInstances 测试多个实例并发运行
func TestRAImpl_ConcurrentInstances(t *testing.T) {
	// 设置测试参数
	n := int64(4)      // 总节点数
	f := int64(1)      // 最大容错数
	instanceCount := 3 // 并发实例数

	// 为每个实例创建节点和消息通道
	allNodes := make([][]*RAImpl, instanceCount)
	allChannels := make([][]chan *protobuf.RAMessage, instanceCount)

	for inst := 0; inst < instanceCount; inst++ {
		// 创建消息通道
		msgChannels := make([]chan *protobuf.RAMessage, n)
		for i := range msgChannels {
			msgChannels[i] = make(chan *protobuf.RAMessage, 100)
		}
		allChannels[inst] = msgChannels

		// 创建节点
		nodes := make([]*RAImpl, n)
		for i := int64(0); i < n; i++ {
			nodes[i] = NewRAImpl(i, n, f, fmt.Sprintf("test-instance-%d", inst+1))

			// 设置发送和接收函数
			nodeID := i
			instID := inst
			nodes[i].send = func(msg *protobuf.RAMessage) {
				destID := msg.DestID
				// 将消息发送到目标节点的通道
				select {
				case allChannels[instID][destID] <- msg:
					// 消息成功发送
				default:
					t.Errorf("Channel full when sending from %d to %d in instance %d",
						nodeID, destID, instID)
				}
			}

			nodes[i].receive = func() chan *protobuf.RAMessage {
				return allChannels[instID][nodeID]
			}
		}

		allNodes[inst] = nodes
	}

	// 启动所有节点的处理循环
	for inst := 0; inst < instanceCount; inst++ {
		for i := int64(0); i < n; i++ {
			allNodes[inst][i].Run()
		}
	}

	// 为每个实例设置不同的输入
	testInputs := make([][]byte, instanceCount)
	for i := 0; i < instanceCount; i++ {
		testInputs[i] = []byte(fmt.Sprintf("test message for instance %d", i+1))
		for j := int64(0); j < n; j++ {
			allNodes[i][j].Input(testInputs[i]) // 节点0输入消息
		}
	}

	// 等待所有节点在所有实例中的输出
	outputs := make([][][]byte, instanceCount)
	for i := range outputs {
		outputs[i] = make([][]byte, n)
	}

	var outputWg sync.WaitGroup

	for inst := 0; inst < instanceCount; inst++ {
		for i := int64(0); i < n; i++ {
			outputWg.Add(1)
			go func(instIdx int, nodeIdx int64) {
				defer outputWg.Done()
				select {
				case output := <-allNodes[instIdx][nodeIdx].Output():
					outputs[instIdx][nodeIdx] = output
					t.Logf("Instance %d, Node %d received output: %s",
						instIdx+1, nodeIdx, string(output))
				case <-time.After(3 * time.Second):
					t.Errorf("Timeout waiting for output from instance %d, node %d",
						instIdx+1, nodeIdx)
				}
			}(inst, i)
		}
	}

	// 等待输出完成或超时
	outputDone := make(chan bool)
	go func() {
		outputWg.Wait()
		outputDone <- true
	}()

	select {
	case <-outputDone:
		// 所有输出都已收到
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out waiting for all outputs")
	}

	// 验证所有节点在每个实例中的输出是否一致
	for inst := 0; inst < instanceCount; inst++ {
		for i := int64(0); i < n; i++ {
			if !bytes.Equal(outputs[inst][i], testInputs[inst]) {
				t.Errorf("Instance %d, Node %d output mismatch: expected %s, got %s",
					inst+1, i, string(testInputs[inst]), string(outputs[inst][i]))
			}
		}
	}

	// 清理资源
	for inst := 0; inst < instanceCount; inst++ {
		for i := int64(0); i < n; i++ {
			allNodes[inst][i].terminated = true
		}
	}
}
