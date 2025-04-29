package vaba

import (
	"fmt"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestIndexCoverGatherImpl(t *testing.T) {
	// Test parameters
	n := int64(4)
	threshold := int64(1) // t = 1, so we need n-t = 3 parties
	sessionID := "test-session"

	// Create message delivery system
	msgQueues := make([]chan *protobuf.ICGMessage, n)
	IGMsgQueues := make([]chan *protobuf.IGMessage, n)
	RAMsgQueues := make(map[string][]chan *protobuf.RAMessage, n)
	for i := int64(0); i < n; i++ {
		msgQueues[i] = make(chan *protobuf.ICGMessage, 100)
		IGMsgQueues[i] = make(chan *protobuf.IGMessage, 100)
		RAMsgQueues["ICG"+strconv.FormatInt(i, 10)] = make([]chan *protobuf.RAMessage, n)
		for j := int64(0); j < n; j++ {
			RAMsgQueues["ICG"+strconv.FormatInt(i, 10)][j] = make(chan *protobuf.RAMessage, 100)
		}
	}

	// Create nodes
	nodes := make([]*IndexCoverGatherImpl, n)
	sendFunc := func(message *protobuf.ICGMessage) {
		destID := message.DestID
		msgQueues[destID] <- message
	}
	IGSendFunc := func(message *protobuf.IGMessage) {
		IGMsgQueues[message.DestID] <- message
	}
	RASendFunc := func(message *protobuf.RAMessage) {
		RAMsgQueues[message.InstanceID][message.DestID] <- message
	}
	for i := int64(0); i < n; i++ {
		// Create receive function for each node
		receiveFunc := func(nodeID int64) func() chan *protobuf.ICGMessage {
			return func() chan *protobuf.ICGMessage {
				return msgQueues[nodeID]
			}
		}(i)
		IGReceiveFunc := func(nodeID int64) func() chan *protobuf.IGMessage {
			return func() chan *protobuf.IGMessage {
				return IGMsgQueues[nodeID]
			}
		}(i)
		indexI := i
		// Create node
		nodes[i] = NewIndexCoverGatherImpl(i, n, threshold, sessionID, sendFunc, receiveFunc)
		nodes[i].output = make(chan []int64, 1)
		nodes[i].indexGatherInstance = NewIGImpl(i, n, threshold, "IG_for_ICG", IGSendFunc, IGReceiveFunc)
		for j := int64(0); j < n; j++ {
			indexJ := j
			nodes[i].reliableAgreementInstances[j].send = RASendFunc
			RAReceiveFunc := func() chan *protobuf.RAMessage {
				return RAMsgQueues["ICG"+strconv.FormatInt(indexJ, 10)][indexI]
			}
			nodes[i].reliableAgreementInstances[j].receive = RAReceiveFunc
		}
		// Start protocol
		nodes[i].Run()
	}

	// Create test validations
	// All nodes validate node 0 and 1
	// For successful termination, at least n-t nodes should be validated
	for i := int64(0); i < n; i++ {
		nodes[i].ValidateParty(0)
		nodes[i].ValidateParty(1)
		nodes[i].ValidateParty(2)
	}

	// Wait for output or timeout
	timeout := time.After(10 * time.Second)

	// Keep track of how many nodes terminated
	terminated := int64(0)

	// Expected validation set should contain at least parties 0, 1, 2
	expected := make(map[int64]bool)
	expected[0] = true
	expected[1] = true
	expected[2] = true

	var wg sync.WaitGroup
	wg.Add(1)
	outputChan := make(chan []int64, 4)
	for i := int64(0); i < n; i++ {
		go func(index int64) {
			xi := <-nodes[index].Output()
			outputChan <- xi
		}(i)
	}
	finishChan := make(chan bool)
	go func() {
		for {
			select {
			case <-outputChan:
				terminated++
				if terminated >= n-threshold {
					finishChan <- true
				}
			}
		}
	}()

	select {
	case <-timeout:
		t.Fatalf("Test timed out - only %d of %d nodes terminated", terminated, n)
		return
	case <-finishChan:
		fmt.Printf("finish only %d of %d nodes terminated", terminated, n)
	}
}
