package network

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

type TCPNode struct {
	ID           int
	Server       *TCPServer
	Clients      map[int]*TCPClient // Keyed by TCPNode ID
	ClientsMutex sync.Mutex
}

func NewTCPNode(id int, address string, peerAddresses map[int]string) *TCPNode {
	TCPNode := &TCPNode{
		ID:      id,
		Server:  NewTCPServer(address),
		Clients: make(map[int]*TCPClient),
	}

	for peerID, addr := range peerAddresses {
		if peerID == id {
			continue // Skip self
		}
		TCPNode.Clients[peerID] = NewTCPClient(addr)
	}

	return TCPNode
}

func (TCPNode *TCPNode) Start() {
	handleMessage := func(msg JsonMessage) {
		log.Printf("TCPNode %d received message: %s", TCPNode.ID, msg)
	}

	go TCPNode.Server.Listen(handleMessage)
}

func (TCPNode *TCPNode) SendMessage(toID int, msg JsonMessage) {
	client, ok := TCPNode.Clients[toID]
	if !ok {
		log.Printf("No client found for TCPNode %d", toID)
		return
	}
	client.SendMessage(msg)
}

func (node *TCPNode) ConnectToNode(peerID int, address string) error {
	// check connection
	if _, exists := node.Clients[peerID]; exists {
		log.Printf("Node %d is already connected to node %d", node.ID, peerID)
		return nil
	}
	// create new connection
	client := NewTCPClient(address)
	if err := client.Connect(); err != nil {
		return err
	}

	// save conn
	node.ClientsMutex.Lock()
	node.Clients[peerID] = client
	node.ClientsMutex.Unlock()

	log.Printf("node %d connect to node %d\n", node.ID, peerID)
	return nil
}

func (node *TCPNode) DisconnectFromNode(peerID int) {
	node.ClientsMutex.Lock()
	if client, exists := node.Clients[peerID]; exists {
		client.Close()
		delete(node.Clients, peerID)
	}
	node.ClientsMutex.Unlock()
	log.Printf("node %d disconnect from node %d\n", node.ID, peerID)
}

func TestTCPNodeCommunication(t *testing.T) {

	nodeAddresses := map[int]string{
		1: "localhost:8081",
		2: "localhost:8082",
		3: "localhost:8083",
		4: "localhost:8084",
	}

	// create and init nodes
	nodes := make(map[int]*TCPNode)
	for id, addr := range nodeAddresses {
		node := NewTCPNode(id, addr, nodeAddresses)
		nodes[id] = node
		go node.Start()
	}

	time.Sleep(2 * time.Second)

	// send message
	var wg sync.WaitGroup
	for senderID, node := range nodes {
		for receiverID := range nodeAddresses {
			if senderID == receiverID {
				continue
			}
			wg.Add(1)
			go func(senderID, receiverID int, node *TCPNode) {
				defer wg.Done()
				msgContent := fmt.Sprintf("Message from node %d to node %d", senderID, receiverID)
				msg := JsonMessage{DataType: "test", Content: []byte(msgContent)}
				node.SendMessage(receiverID, msg)

			}(senderID, receiverID, node)
		}
	}

	wg.Wait()

	time.Sleep(2 * time.Second)

}

// test concurrent connection management
func TestTCPNode_Concurrency(t *testing.T) {
	nodeAddresses := map[int]string{
		1: "localhost:8081",
		2: "localhost:8082",
		3: "localhost:8083",
		4: "localhost:8084",
	}

	// create and init nodes
	nodes := make(map[int]*TCPNode)
	for id, addr := range nodeAddresses {
		node := NewTCPNode(id, addr, nodeAddresses)
		nodes[id] = node
		go node.Start()
	}
	time.Sleep(2 * time.Second)

	// set connection
	for senderID, node := range nodes {
		for receiverID, addr := range nodeAddresses {
			if senderID == receiverID {
				continue
			}
			if err := node.ConnectToNode(receiverID, addr); err != nil {
				t.Errorf("Node %d failed to connect to node %d: %v", senderID, receiverID, err)
			}
		}
	}

	time.Sleep(2 * time.Second)

	// send message
	var wg sync.WaitGroup
	for senderID, node := range nodes {
		for receiverID := range nodeAddresses {
			if senderID == receiverID {
				continue
			}
			wg.Add(1)
			go func(senderID, receiverID int, node *TCPNode) {
				defer wg.Done()
				msgContent := fmt.Sprintf("Message from node %d to node %d", senderID, receiverID)
				msg := JsonMessage{DataType: "test", Content: []byte(msgContent)}
				node.SendMessage(receiverID, msg)
			}(senderID, receiverID, node)
		}
	}

	wg.Wait()

	fmt.Println("Disconnecting node 3 from node 4...")
	nodes[3].DisconnectFromNode(4)
	nodes[4].DisconnectFromNode(3)

	time.Sleep(1 * time.Second)

	if client, ok := nodes[3].Clients[4]; ok && client.IsConnected() {
		t.Errorf("Node 3 is still connected to node 4")
	} else {
		fmt.Println("Node 3 and node 4 have been disconnected")
	}
	if client, ok := nodes[4].Clients[3]; ok && client.IsConnected() {
		t.Errorf("Node 4 is still connected to node 3")
	} else {
		fmt.Println("Node 4 and node 3 have been disconnected")
	}

}
