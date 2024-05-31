package services

import (
	"github.com/QinYuuuu/abvss/protobuf"
	"log"

	"google.golang.org/protobuf/proto"
)

// Encapsulation encapsulates a message to a general type(*protobuf.Message)
func Encapsulation(messageType string, from, dest, index int, payloadMessage any) *protobuf.Message {
	var data []byte
	var err error
	switch messageType {
	case "osv":
		data, err = proto.Marshal((payloadMessage).(*protobuf.OSVMessage))
	}

	if err != nil {
		log.Fatalln(err)
	}
	return &protobuf.Message{
		Mtype:  messageType,
		FromID: int64(from),
		DestID: int64(dest),
		Data:   data,
	}
}
