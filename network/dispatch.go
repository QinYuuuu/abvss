package network

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"google.golang.org/protobuf/proto"
)

// MakeDispatchChannels dispatch messages from receiveChannel
// and make a double layer Map : (messageType) --> (channel)
func (p *Peer) MakeDispatchChannels() {
	totalMap := p.dispatchChannels
	go func() { //dispatcher
		for {
			m := <-p.receiveChannel
			protocolMap, _ := totalMap.LoadOrStore(m.Protocol, new(sync.Map))
			idChan, _ := protocolMap.(*sync.Map).LoadOrStore(m.InstanceID, make(chan *protobuf.Message, p.buffLen))
			idChan.(chan *protobuf.NewMessage) <- m
			atomic.AddInt64(&p.Traffic, int64(proto.Size(m)))
		}
	}()
}

func (p *Peer) GetMessageChan(protocol, session string) chan *protobuf.NewMessage {
	protocolMap, _ := p.dispatchChannels.LoadOrStore(protocol, new(sync.Map))
	sessionChan, _ := protocolMap.(*sync.Map).LoadOrStore(session, make(chan *protobuf.Message, p.buffLen))
	return sessionChan.(chan *protobuf.NewMessage)
}

func encapsulate(protocol string, payloadMsg any) *protobuf.NewMessage {
	var data []byte
	var err error
	var fromID, destID int64
	var instanceID string
	switch protocol {
	case string(protobuf.OSVProtocol):
		newMsg := payloadMsg.(*protobuf.OSVMessage)
		data, err = proto.Marshal(newMsg)
		fromID = newMsg.FromID
		destID = newMsg.DestID
		instanceID = newMsg.InstanceID
	case string(protobuf.OptRBCProtocol):
		newMsg := payloadMsg.(*protobuf.OptRBCMessage)
		data, err = proto.Marshal(newMsg)

		fromID = newMsg.FromID
		destID = newMsg.DestID
		instanceID = newMsg.SessionID
	default:
		slog.Error("unhandled", "protocol", protocol)
		return nil
	}
	if err != nil {
		slog.Error("encapsulate proto marshal fail", "protocol", protocol, "err", err)
		return nil
	}
	return &protobuf.NewMessage{
		Protocol:   protocol,
		InstanceID: instanceID,
		FromID:     fromID,
		DestID:     destID,
		Data:       data,
	}
}

// Decapsulate decapsulates a message to its original type
func Decapsulate(m *protobuf.NewMessage) any {
	switch m.Protocol {
	case string(protobuf.OSVProtocol):
		var payloadMessage protobuf.OSVMessage
		err := proto.Unmarshal(m.Data, &payloadMessage)
		if err != nil {
			slog.Error("OSVProtocol proto unmarshal", slog.Any("error", err))
			return nil
		}
		return &payloadMessage
	case string(protobuf.OptRBCProtocol):
		var payloadMessage protobuf.OptRBCMessage
		err := proto.Unmarshal(m.Data, &payloadMessage)
		if err != nil {
			slog.Error("OptRBCProtocol proto unmarshal", slog.Any("error", err))
			return nil
		}
		return &payloadMessage
	default:
		slog.Error("Unsupported message type", slog.Any("type", m.Protocol))
		return nil
	}
}
