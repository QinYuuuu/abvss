package network

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

type RPCNode struct {
	ID           int
	Server       *TCPServer
	Clients      map[int]*TCPClient // Keyed by RPCNode ID
	ClientsMutex sync.Mutex
}

func NewRPCNode(id int, address string, peerAddresses map[int]string) *RPCNode {
	RPCNode := &RPCNode{
		ID:      id,
		Server:  NewTCPServer(address),
		Clients: make(map[int]*TCPClient),
	}

	for peerID, addr := range peerAddresses {
		if peerID == id {
			continue // Skip self
		}
		RPCNode.Clients[peerID] = NewTCPClient(addr)
	}

	return RPCNode
}

func (RPCNode *RPCNode) Start() {
	handleMessage := func(msg RpcMessage) {
		log.Printf("RPCNode %d received message: %s", RPCNode.ID, msg.Data)
	}

	go RPCNode.Server.ListenRpc(handleMessage)
}

func (RPCNode *RPCNode) SendRpcMessage(toID int, msg RpcMessage) {
	client, ok := RPCNode.Clients[toID]
	if !ok {
		log.Printf("No client found for RPCNode %d", toID)
		return
	}
	client.SendRpcMessage(&msg)
}

func (node *RPCNode) ConnectToNode(peerID int, address string) error {
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

func (node *RPCNode) DisconnectFromNode(peerID int) {
	node.ClientsMutex.Lock()
	if client, exists := node.Clients[peerID]; exists {
		client.Close()
		delete(node.Clients, peerID)
	}
	node.ClientsMutex.Unlock()
	log.Printf("node %d disconnect from node %d\n", node.ID, peerID)
}

func TestRPCNodeCommunication(T *testing.T) {
	nodeAddresses := map[int]string{
		1: "localhost:8081",
		2: "localhost:8082",
		3: "localhost:8083",
		4: "localhost:8084",
	}

	// create and init nodes
	nodes := make(map[int]*RPCNode)
	for id, addr := range nodeAddresses {
		node := NewRPCNode(id, addr, nodeAddresses)
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
				T.Errorf("Node %d failed to connect to node %d: %v", senderID, receiverID, err)
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
			go func(senderID, receiverID int, node *RPCNode) {
				defer wg.Done()
				msgContent := fmt.Sprintf("Message from node %d to node %d", senderID, receiverID)

				msg := RpcMessage{Type: 0, Dest: int32(receiverID), From: int32(senderID), Data: []byte(msgContent)}

				node.SendRpcMessage(receiverID, msg)
			}(senderID, receiverID, node)
		}
	}

	wg.Wait()

	fmt.Println("Disconnecting node 3 from node 4...")
	nodes[3].DisconnectFromNode(4)
	nodes[4].DisconnectFromNode(3)

	time.Sleep(1 * time.Second)

	if client, ok := nodes[3].Clients[4]; ok && client.IsConnected() {
		T.Errorf("Node 3 is still connected to node 4")
	} else {
		fmt.Println("Node 3 and node 4 have been disconnected")
	}
	if client, ok := nodes[4].Clients[3]; ok && client.IsConnected() {
		T.Errorf("Node 4 is still connected to node 3")
	} else {
		fmt.Println("Node 4 and node 3 have been disconnected")
	}
}
