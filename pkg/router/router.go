package router

import (
	"abvss/pkg/protobuf"
	"log"
	"sync"
	"sync/atomic"
)

// dispatchLoop 消息分发主循环
func dispatchLoop(receiveChannel chan *protobuf.Message, baseSize int) {
	dispatcheChannels := new(sync.Map)

	for msg := range receiveChannel {
		typeMap, _ := dispatcheChannels.LoadOrStore(msg.GetType(), new(sync.Map))
		sessionMap, _ := typeMap.(*sync.Map).LoadOrStore(msg.GetSessionID(), new(sync.Map))
		// 获取或创建具体ID的通道
		msgChan, _ := sessionMap.(*sync.Map).LoadOrStore(string(msg.GetId()), make(chan *protobuf.Message, baseSize))

		// 非阻塞发送消息
		select {
		case msgChan <- msg:
			dc.recordTraffic(msg)
		default:
			log.Printf("Channel full, dropping message. Type:%s ID:%s", msg.Type, msg.Id)
		}
	}
}
