package vaba

import (
	"math/big"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
)

func InitLocalMultiASKS(n, t, dealerID int64, instanceID string, secret, prime *big.Int, rbcList []*broadcast.OptRBC, raList []*RAImpl) []*ASKSImpl {
	channels := make(map[int64]chan *protobuf.ASKSMessage)
	for i := int64(0); i < n; i++ {
		channels[i] = make(chan *protobuf.ASKSMessage, 100)
	}

	parties := make([]*ASKSImpl, n)
	for i := int64(0); i < n; i++ {
		id := i
		if id == dealerID {
			parties[i] = NewASKSDealer(i, n, t, instanceID, secret, prime)
		} else {
			parties[i] = NewASKS(i, n, t, dealerID, instanceID, prime)
		}

		// 设置通信函数
		parties[i].send = func(id int64) func(message *protobuf.ASKSMessage) {
			return func(message *protobuf.ASKSMessage) {
				channels[message.DestID] <- message
			}
		}(i)

		parties[i].receive = func(id int64) func() chan *protobuf.ASKSMessage {
			return func() chan *protobuf.ASKSMessage {
				return channels[id]
			}
		}(i)

		parties[i].rbc = rbcList[i]
		parties[i].ra = raList[i]
	}
	return parties
}

func InitLocalMultiICG(n, t int64, instanceID string, raList [][]*RAImpl, igList []*IGImpl) []*IndexCoverGatherImpl {
	msgChannels := make([]chan *protobuf.ICGMessage, n)
	for i := range msgChannels {
		msgChannels[i] = make(chan *protobuf.ICGMessage, 100)
	}
	send := func(msg *protobuf.ICGMessage) {
		msgChannels[msg.DestID] <- msg
	}
	icgList := make([]*IndexCoverGatherImpl, n)
	for i := int64(0); i < n; i++ {
		receive := func() chan *protobuf.ICGMessage {
			return msgChannels[i]
		}
		icgList[i] = NewIndexCoverGatherImpl(i, n, t, instanceID, send, receive)
		raInstanceList := make([]*RAImpl, n)
		for j := int64(0); j < n; j++ {
			raInstanceList[j] = raList[j][i]
		}
		icgList[i].reliableAgreementInstances = raInstanceList
		icgList[i].indexGatherInstance = igList[i]
	}
	return icgList
}

func InitLocalMultiIG(n, t int64, instanceID string) []*IGImpl {
	msgChannels := make([]chan *protobuf.IGMessage, n)
	for i := range msgChannels {
		msgChannels[i] = make(chan *protobuf.IGMessage, 100)
	}
	send := func(msg *protobuf.IGMessage) {
		msgChannels[msg.DestID] <- msg
	}
	igList := make([]*IGImpl, n)
	for i := int64(0); i < n; i++ {
		receive := func() chan *protobuf.IGMessage {
			return msgChannels[i]
		}
		igList[i] = NewIGImpl(i, n, t, instanceID, send, receive)
	}
	return igList
}

func InitLocalMultiRA(n, t int64, instanceID string) []*RAImpl {
	msgChannels := make([]chan *protobuf.RAMessage, n)
	for i := range msgChannels {
		msgChannels[i] = make(chan *protobuf.RAMessage, 100)
	}

	raList := make([]*RAImpl, n)
	for i := int64(0); i < n; i++ {
		raList[i] = NewRAImpl(i, n, t, instanceID)
		raList[i].send = func(msg *protobuf.RAMessage) {
			msgChannels[msg.DestID] <- msg
		}
		raList[i].receive = func() chan *protobuf.RAMessage {
			return msgChannels[i]
		}
	}
	return raList
}
