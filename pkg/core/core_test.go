package core

import (
	"fmt"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"log"
	"net"
	"testing"
	"time"
)

func TestMakeReceiveChannel(t *testing.T) {
	port := "8882"
	lis, receiveChannel := MakeReceiveChannel(port)
	defer func(lis *net.TCPListener) {
		err := lis.Close()
		if err != nil {
			log.Printf("close failed, err:%v\n", err)
		}
	}(lis)
	m := <-(receiveChannel)
	fmt.Println("The Message Received from channel is")
	fmt.Println("id==", m.Type)
	fmt.Println("sender==", m.Sender)
	fmt.Println("len==", len(m.Data))

}

func TestMakeSendChannel(t *testing.T) {
	hostIP := "127.0.0.1"
	hostPort := "8882"

	conn, sendChannel := MakeSendChannel(hostIP, hostPort)
	defer func(conn *net.TCPConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("close failed, err:%v\n", err)
		}
	}(conn)
	fmt.Println(sendChannel)

	for i := 0; i < 100; i++ {
		m := &protobuf.Message{
			Type:   "Alice",
			Sender: uint32(i),
			Data:   make([]byte, 10000000),
		}
		sendChannel <- m
		time.Sleep(time.Duration(1) * time.Second)
	}
}
