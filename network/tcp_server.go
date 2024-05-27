package network

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"

	"google.golang.org/protobuf/proto"
)

type TCPServer struct {
	Address string
}

func NewTCPServer(address string) *TCPServer {
	return &TCPServer{Address: address}
}

// Listening the messages
func (server *TCPServer) Listen(handleMessage func(JsonMessage)) {
	listener, err := net.Listen("tcp", server.Address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", server.Address, err)
	}
	defer listener.Close()
	log.Printf("Server listening on %s", server.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go func() {
			defer conn.Close()
			reader := bufio.NewReader(conn)

			message, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Failed to read from connection: %v", err)
				return
			}

			var msg JsonMessage
			if err := json.Unmarshal([]byte(message), &msg); err != nil {
				log.Printf("Failed to unmarshal JsonMessage: %v", err)
				return
			}

			handleMessage(msg)
		}()
	}
}

// Listening the rpc_messages
func (server *TCPServer) ListenRpc(handleMessage func(RpcMessage)) {
	listener, err := net.Listen("tcp", server.Address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", server.Address, err)
	}
	defer listener.Close()
	log.Printf("Server listening on %s", server.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go func() {
			defer conn.Close()
			for {

				lenBytes := make([]byte, 4)
				if _, err := io.ReadFull(conn, lenBytes); err != nil {
					if err != io.EOF {
						log.Printf("Failed to read length prefix: %v", err)
					}
					return
				}
				messageLength := binary.BigEndian.Uint32(lenBytes)

				messageBytes := make([]byte, messageLength)
				if _, err := io.ReadFull(conn, messageBytes); err != nil {
					log.Printf("Failed to read message: %v", err)
					return
				}

				var msg RpcMessage
				if err := proto.Unmarshal(messageBytes, &msg); err != nil {
					log.Printf("Failed to unmarshal RpcMessage: %v", err)
					return
				}

				handleMessage(msg)
			}
		}()
	}
}
