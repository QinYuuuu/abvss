package network

import (
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
			protocolMap, _ := totalMap.LoadOrStore(m.Type, new(sync.Map))
			idChan, _ := protocolMap.(*sync.Map).LoadOrStore(string(m.InstanceID), make(chan *protobuf.Message, p.buffLen))
			idChan.(chan *protobuf.NewMessage) <- m
			atomic.AddInt64(&p.Traffic, int64(proto.Size(m)))
		}
	}()
}
