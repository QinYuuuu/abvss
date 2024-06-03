package core

import (
	"fmt"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"sync"
	"testing"
	"time"
)

var wg1 = sync.WaitGroup{}
var wg2 = sync.WaitGroup{}

func TestMakeReceiveChannel(t *testing.T) {
	port := "8882"
	receiveChannel := MakeReceiveChannel(port)

	m := <-(receiveChannel)
	fmt.Println("The Message Received from channel is")
	fmt.Println("id==", m.Type)
	fmt.Println("sender==", m.Sender)
	fmt.Println("len==", len(m.Data))

}

func TestMakeSendChannel(t *testing.T) {
	serverAddress := "127.0.0.1:8882"

	sendChannel := MakeSendChannel(serverAddress)
	fmt.Println(sendChannel)

	for i := 0; i < 100; i++ {
		m := &protobuf.Message{
			Type:   "Alice",
			Sender: uint32(i),
			Data:   make([]byte, 10000000),
		}
		(sendChannel) <- m
		time.Sleep(time.Duration(1) * time.Second)
	}
	for {

	}
}
