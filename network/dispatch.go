package network

import (
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"google.golang.org/protobuf/proto"
	"sync"
	"sync/atomic"
)

// MakeDispatchChannels dispatch messages from receiveChannel
// and make a double layer Map : (messageType) --> (channel)
func (p *Peer) MakeDispatchChannels() {
	totalMap := p.dispatchChannels
	go func() { //dispatcher
		for {
			m := <-p.receiveChannel
			protocolMap, _ := totalMap.LoadOrStore(m.Type, new(sync.Map))
			idChan, _ := protocolMap.(*sync.Map).LoadOrStore(string(m.Id), make(chan *protobuf.Message, p.buffLen))
			idChan.(chan *protobuf.Message) <- m
			atomic.AddInt64(&p.Traffic, int64(proto.Size(m)))
		}
	}()
}
