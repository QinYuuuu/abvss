package core

import (
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"
	"log"
	"net"

	"google.golang.org/protobuf/proto"
)

// MAXMESSAGE is the size of channels
var MAXMESSAGE = 1024

// MakeSendChannel returns a channel to send messages to hostIP
func MakeSendChannel(serverAddress string) chan *protobuf.Message {
	var addr *net.TCPAddr
	var conn *net.TCPConn
	var err1, err2 error
	//Retry to connet to node
	retry := true
	for retry {
		addr, err1 = net.ResolveTCPAddr("tcp4", serverAddress)
		conn, err2 = net.DialTCP("tcp4", nil, addr)
		if err1 != nil || err2 != nil {
			log.Fatalln(err1)
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
			_, err3 := conn.Write(byt)
			if err2 != nil || err3 != nil {
				log.Fatalln("The send channel has break down!", err2)
			}
		}
	}(conn, sendChannel)

	return sendChannel
}
