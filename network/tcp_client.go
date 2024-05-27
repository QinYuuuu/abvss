package network

import (
	"encoding/binary"

	"encoding/json"
	"fmt"
	"log"
	"net"
	"google.golang.org/protobuf/proto"
)

type TCPClient struct {
	ServerAddress string
	Conn          net.Conn

	isConnected   bool
}

// create a client instance
func NewTCPClient(serverAddress string) *TCPClient {
	return &TCPClient{ServerAddress: serverAddress}
}

func (c *TCPClient) IsConnected() bool {
	return c.isConnected
}

// create a connection
func (c *TCPClient) Connect() error {
	var err error
	c.Conn, err = net.Dial("tcp", c.ServerAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to server at %s: %v", c.ServerAddress, err)
	}
	c.isConnected = true
	return nil
}

// close the connection
func (c *TCPClient) Close() error {
	if c.Conn != nil {
		err := c.Conn.Close()
		if err != nil {
			return err
		}
		c.isConnected = false
		return nil
	}
	return fmt.Errorf("connection does not exist")
}

// sendMessage to server
func (c *TCPClient) SendMessage(msg JsonMessage) error {
	if c.Conn == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("Error while serializing JsonMessage: %v", err)
	}
	msgBytes = append(msgBytes, '\n')
	_, err = c.Conn.Write(msgBytes)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	return nil
}

// sendMessage to server
func (c *TCPClient) SendRpcMessage(msg *RpcMessage) error {
	if c.Conn == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Fatalf("Error while serializing RpcMessage: %v", err)
	}

	var lenPrefix = make([]byte, 4)
	binary.BigEndian.PutUint32(lenPrefix, uint32(len(msgBytes)))

	// 将长度前缀和消息本体一起写入连接
	fullMessage := append(lenPrefix, msgBytes...)
	_, err = c.Conn.Write(fullMessage)

	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	return nil
}
