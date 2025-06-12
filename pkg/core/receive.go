package core

import (
	"context"
	"io"
	"log"
	"net"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"

	"google.golang.org/protobuf/proto"
)

// MakeReceiveChannel returns a channel receiving messages
func MakeReceiveChannel(ctx context.Context, port string, n int) (*net.TCPListener, []*net.TCPConn, chan *protobuf.Message) {
	var addr *net.TCPAddr
	var lis *net.TCPListener
	var err1, err2 error
	retry := true
	//Retry to make listener
	for retry {
		addr, err1 = net.ResolveTCPAddr("tcp4", ":"+port)
		if err1 != nil {
			log.Fatalln(err1)
			retry = true
		}
		lis, err2 = net.ListenTCP("tcp4", addr)
		if err2 != nil {
			log.Fatalln(err2)
			retry = true
		} else {
			retry = false
		}
	}
	//Make the receive channel and the handle func
	conns := make([]*net.TCPConn, n)
	receiveChannel := make(chan *protobuf.Message, MAXMESSAGE)
	for i := range n {
		//The handle func run forever
		go func(i int) {
			conn, err3 := lis.AcceptTCP()
			if err3 != nil {
				// log.Fatalln(err3)
				return
			}
			conns[i] = conn
			conn.SetKeepAlive(true)

			//Once connect to a node, make a sub-handle func to handle this connection
			go func(conn *net.TCPConn, channel chan *protobuf.Message) {
				for {
					//Receive bytes
					lengthBuf := make([]byte, 4)
					_, err1 := io.ReadFull(conn, lengthBuf)
					if err1 != nil {
						// log.Println("The receive channel has break down", err1)
						break
					}
					length := utils.BytesToInt(lengthBuf)
					buf := make([]byte, length)
					_, err2 := io.ReadFull(conn, buf)
					if err2 != nil {
						// log.Println("The receive channel has break down", err2)
						break
					}
					//Do Unmarshal
					var m protobuf.Message
					err3 := proto.Unmarshal(buf, &m)
					if err3 != nil {
						// log.Println("Unmarshal from receive channel", err3)
					}
					//Push protobuf.Message to receivechannel
					(channel) <- &m
				}

			}(conn, receiveChannel)
		}(i)
	}
	return lis, conns, receiveChannel
}
