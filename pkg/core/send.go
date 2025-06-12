package core

import (
	"log"
	"net"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"

	"google.golang.org/protobuf/proto"
)

// MAXMESSAGE is the size of channels
var MAXMESSAGE = 1024

// MakeSendChannel returns a channel to send messages to hostIP
func MakeSendChannel(hostIP string, hostPort string) (*net.TCPConn, chan *protobuf.Message) {
	var addr *net.TCPAddr
	var conn *net.TCPConn
	var err1, err2 error
	//Retry to connet to node
	retry := true
	for retry {
		addr, err1 = net.ResolveTCPAddr("tcp4", hostIP+":"+hostPort)
		if err1 != nil {
			log.Fatalln(err1)
			retry = true
		}
		conn, err2 = net.DialTCP("tcp4", nil, addr)
		if err2 != nil {
			log.Fatalln(err2)
			retry = true
			continue
		} else {
			retry = false
		}
		conn.SetKeepAlive(true)
	}
	//Make the send channel and the handle func
	sendChannel := make(chan *protobuf.Message, MAXMESSAGE)
	go func(conn *net.TCPConn, channel chan *protobuf.Message) {
		for {
			//Pop protobuf.Message form sendchannel
			m := <-(channel)

			//Do Marshal
			byt, err1 := proto.Marshal(m)
			if err1 != nil {
				log.Fatalln(err1)
			}
			//Send bytes
			length := len(byt)
			_, err2 := conn.Write(utils.IntToBytes(length))
			if err2 != nil {
				// log.Println("The send channel has break down", err2)
				break
			}
			_, err3 := conn.Write(byt)
			if err3 != nil {
				// log.Println("The send channel has break down", err3)
				break
			}
		}
	}(conn, sendChannel)

	return conn, sendChannel
}
