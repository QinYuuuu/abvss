package core

import (
<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"
=======
	"abvss/pkg/protobuf"
	"abvss/pkg/utils"
>>>>>>> 19b0d27 (Initial commit)
=======
	"abvss/pkg/protobuf"
	"abvss/pkg/utils"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net"
	"time"
)

// MakeReceiveChannel returns a channel receiving messages
func MakeReceiveChannel(port string) (*net.TCPListener, chan *protobuf.Message) {
	var addr *net.TCPAddr
	var lis *net.TCPListener
	var err1, err2 error
	retry := true
	//Retry to make listener
	for retry {
		addr, err1 = net.ResolveTCPAddr("tcp4", ":"+port)
		lis, err2 = net.ListenTCP("tcp4", addr)
		if err1 != nil || err2 != nil {
			log.Println("In mvba make listener falied and retry", err1, err2)
			retry = true
			time.Sleep(3 * time.Second)
		} else {
			retry = false
		}
	}
	//log.Printf("listen on port: %v", port)
	//Make the receive channel and the handle func
	var conn *net.TCPConn
	var err3 error
	receiveChannel := make(chan *protobuf.Message, MAXMESSAGE)
	go func() {
		for {
			//The handle func run forever
			conn, err3 = lis.AcceptTCP()
			if err3 != nil {
				log.Fatalln(err3)
			}
			conn.SetKeepAlive(true)

			//Once connect to a node, make a sub-handle func to handle this connection
			go func(conn *net.TCPConn, channel chan *protobuf.Message) {
				for {
					//Receive bytes
					lengthBuf := make([]byte, 4)
					_, err1 := io.ReadFull(conn, lengthBuf)
					length := utils.BytesToInt(lengthBuf)
					buf := make([]byte, length)
					_, err2 := io.ReadFull(conn, buf)
					if err1 != nil || err2 != nil {
						//log.Println("The receive channel has break down", err1, err2)
						break
					}
					//Do Unmarshal
					var m protobuf.Message
					err3 := proto.Unmarshal(buf, &m)
					if err3 != nil {
						log.Fatalln(err3)
					}
					//Push protobuf.Message to receivechannel
					channel <- &m
				}

			}(conn, receiveChannel)
		}
	}()
	return lis, receiveChannel
}
