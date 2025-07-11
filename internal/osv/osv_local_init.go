package osv

import (
	"github.com/QinYuuuu/abvss/pkg/protobuf"
)

func InitLocalMulti(nodeNum, tnum int64, instanceID string) []*Instance {
	msgChans := make([]chan *protobuf.OSVMessage, nodeNum)
	osvInstances := make([]*Instance, nodeNum)
	for i := 0; i < int(nodeNum); i++ {
		msgChans[i] = make(chan *protobuf.OSVMessage, nodeNum*10)
	}
	for i := int64(0); i < nodeNum; i++ {
		ID := i
		send := func(msg *protobuf.OSVMessage) {
			msgChans[msg.DestID] <- msg
		}
		receive := func() chan *protobuf.OSVMessage {
			return msgChans[ID]
		}
		osvInstances[i] = NewInstance(nodeNum, tnum, ID, instanceID, send, receive)
	}
	return osvInstances
}
