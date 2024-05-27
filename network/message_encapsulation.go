package network

import (
	"log"

	"github.com/QinYuuuu/avid-d/protobuf"

	"google.golang.org/protobuf/proto"
)

func Encapsulation(messageType string, id []byte, sender uint32, payloadMessage any) *protobuf.Message {
	var data []byte
	var err error
	switch messageType {
	case "Z":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Z))
	}
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}
	return &protobuf.Message{
		Type:   messageType,
		Id:     id,
		Sender: sender,
		Data:   data,
	}
}

func Decapsulation(messageType string, message *protobuf.Message) any {
	//var payloadMessage any
	switch messageType {
	case "Z":
		var payloadMessage protobuf.Z
		proto.Unmarshal(message.Data, &payloadMessage)
		return &payloadMessage
	default:
		var payloadMessage protobuf.Message
		proto.Unmarshal(message.Data, &payloadMessage)
		return &payloadMessage
	}
}
