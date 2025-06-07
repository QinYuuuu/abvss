package vaba

import (
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
)

func TestIGImpl_GetValid(t *testing.T) {
	ig := &IGImpl{
		valid: sync.Map{},
	}

	// Add some valid parties
	ig.valid.Store(int64(1), true)
	ig.valid.Store(int64(3), true)
	ig.valid.Store(int64(5), true)

	valid := ig.GetValid()

	// Sort for deterministic comparison
	sort.Slice(valid, func(i, j int) bool {
		return valid[i] < valid[j]
	})

	expected := []int64{1, 3, 5}
	if !reflect.DeepEqual(valid, expected) {
		t.Errorf("GetValid() = %v, want %v", valid, expected)
	}
}

func TestIGImpl_AddValid(t *testing.T) {
	ig := &IGImpl{
		valid: sync.Map{},
	}

	// Add a party to Valid
	ig.AddValid(1)
	ig.AddValid(2)

	// Check if parties are in Valid
	_, has1 := ig.valid.Load(int64(1))
	_, has2 := ig.valid.Load(int64(2))

	if !has1 || !has2 {
		t.Errorf("AddValid() failed to add parties to Valid")
	}

	// Check if validSize is updated
	if ig.validSize != 2 {
		t.Errorf("validSize = %d, want 2", ig.validSize)
	}

	// Adding the same party again should not change validSize
	ig.AddValid(1)
	if ig.validSize != 2 {
		t.Errorf("validSize = %d, want 2 after adding duplicate", ig.validSize)
	}
}

func TestIGImpl_Input(t *testing.T) {
	// Create channels for testing
	msgChan := make(chan *protobuf.IGMessage, 100)
	send := func(msg *protobuf.IGMessage) {
		msgChan <- msg
	}
	receive := func() chan *protobuf.IGMessage {
		return msgChan
	}

	ig := NewIGImpl(0, 4, 1, "test-instance", send, receive)

	// Call Input with a subset of valid parties
	ig.Input([]int64{1, 2})

	// Check if parties are added to Valid
	if ig.validSize != 2 {
		t.Errorf("Input() failed to add parties to Valid")
	}

	// Call Input again to trigger inform messages (n-t = 3)
	ig.Input([]int64{3})

	// Check if INFORM messages were sent
	if !ig.informSent {
		t.Errorf("informSent = false, want true")
	}

	// Verify messages were sent to all parties
	sentMessages := 0
	for i := 0; i < 4; i++ {
		select {
		case msg := <-msgChan:
			if msg.Type != "INFORM" {
				t.Errorf("Expected INFORM message, got %s", msg.Type)
			}
			sentMessages++
		default:
			// No more messages
		}
	}

	if sentMessages != 4 {
		t.Errorf("Expected 4 INFORM messages, got %d", sentMessages)
	}
}

func TestIGImpl_ProcessMessage_INFORM(t *testing.T) {
	// Create channels for testing
	msgChan := make(chan *protobuf.IGMessage, 100)
	send := func(msg *protobuf.IGMessage) {
		msgChan <- msg
	}
	receive := func() chan *protobuf.IGMessage {
		return msgChan
	}

	ig := NewIGImpl(0, 4, 1, "test-instance", send, receive)

	// Add valid parties
	ig.AddValid(1)
	ig.AddValid(2)

	// Process an INFORM message where Sj is a subset of Valid_i
	msg := &protobuf.IGMessage{
		FromID:     1,
		DestID:     0,
		InstanceID: "test-instance",
		Type:       "INFORM",
		Set:        []int64{1},
	}

	ig.ProcessMessage(msg)

	// Check if ACK message was sent
	select {
	case response := <-msgChan:
		if response.Type != "ACK" || response.DestID != 1 {
			t.Errorf("Expected ACK message to party 1, got %s to %d", response.Type, response.DestID)
		}
	default:
		t.Errorf("No ACK message sent")
	}

	// Process an INFORM message where Sj is not a subset of Valid_i
	msg = &protobuf.IGMessage{
		FromID:     2,
		DestID:     0,
		InstanceID: "test-instance",
		Type:       "INFORM",
		Set:        []int64{3},
	}

	ig.ProcessMessage(msg)

	// Check that no additional message was sent
	select {
	case <-msgChan:
		t.Errorf("Unexpected message sent")
	default:
		// This is expected, no message should be sent
	}
}

func TestIGImpl_ProcessMessage_ACK(t *testing.T) {
	// Create channels for testing
	msgChan := make(chan *protobuf.IGMessage, 100)
	send := func(msg *protobuf.IGMessage) {
		msgChan <- msg
	}
	receive := func() chan *protobuf.IGMessage {
		return msgChan
	}

	ig := NewIGImpl(0, 4, 1, "test-instance", send, receive)

	// Add valid parties
	ig.AddValid(1)
	ig.AddValid(2)

	// Process ACK messages from n-t parties
	for i := int64(1); i <= 3; i++ {
		msg := &protobuf.IGMessage{
			FromID:     i,
			DestID:     0,
			InstanceID: "test-instance",
			Type:       "ACK",
		}
		ig.ProcessMessage(msg)
	}

	// Check if PREPARE messages were sent to all parties
	sentMessages := 0
	for i := 0; i < 4; i++ {
		select {
		case msg := <-msgChan:
			if msg.Type != "PREPARE" {
				t.Errorf("Expected PREPARE message, got %s", msg.Type)
			}
			sentMessages++
		default:
			// No more messages
		}
	}

	if sentMessages != 4 {
		t.Errorf("Expected 4 PREPARE messages, got %d", sentMessages)
	}
}

func TestIGImpl_ProcessMessage_PREPARE(t *testing.T) {
	// Create channels for testing
	msgChan := make(chan *protobuf.IGMessage, 100)
	send := func(msg *protobuf.IGMessage) {
		msgChan <- msg
	}
	receive := func() chan *protobuf.IGMessage {
		return msgChan
	}

	ig := NewIGImpl(0, 4, 1, "test-instance", send, receive)

	// Add valid parties
	ig.AddValid(1)
	ig.AddValid(2)
	ig.AddValid(3)

	// Process PREPARE messages from n-t parties
	for i := int64(1); i <= 3; i++ {
		msg := &protobuf.IGMessage{
			FromID:     i,
			DestID:     0,
			InstanceID: "test-instance",
			Type:       "PREPARE",
			Set:        []int64{i}, // Each party's set contains just itself
		}
		ig.ProcessMessage(msg)
	}

	// Check if output was produced
	select {
	case result := <-ig.Output():
		// Sort the result for deterministic comparison
		sort.Slice(result, func(i, j int) bool {
			return result[i] < result[j]
		})

		expected := []int64{1, 2, 3}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Output = %v, want %v", result, expected)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("No output produced")
	}

	// Check if the protocol terminated
	if !ig.terminated {
		t.Errorf("Protocol did not terminate")
	}
}

func TestIGImpl_FullProtocolFlow(t *testing.T) {
	// Create a test network of 4 nodes with t=1
	n := int64(4)
	threshold := int64(1)
	instanceID := "test-full"

	nodes := InitLocalMultiIG(n, threshold, instanceID)

	// Start all nodes
	for _, node := range nodes {
		node.Run()
	}

	// Initialize with valid sets
	for _, node := range nodes {
		// Each node considers itself and the next node (modulo n) as valid
		node.Input([]int64{0, 1, 2})
	}

	// Wait for termination and check outputs
	outputs := make([][]int64, n)
	var wg sync.WaitGroup
	wg.Add(int(n))
	for i, node := range nodes {
		go func(id int) {
			outputs[id] = <-node.Output()
			// Sort output for comparison
			sort.Slice(outputs[id], func(a, b int) bool {
				return outputs[id][a] < outputs[id][b]
			})
			wg.Done()
		}(i)
	}
	wg.Wait()
	// All nodes should output the same set
	for i := int64(1); i < n; i++ {
		if !reflect.DeepEqual(outputs[0], outputs[i]) {
			t.Errorf("Node %d output %v differs from node 0 output %v", i, outputs[i], outputs[0])
		}
	}

	// The output set should include all nodes (0-3)
	expectedOutput := []int64{0, 1, 2}
	if !reflect.DeepEqual(outputs[0], expectedOutput) {
		t.Errorf("Expected output %v, got %v", expectedOutput, outputs[0])
	}
}
