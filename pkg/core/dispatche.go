package core

import (
<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/QinYuuuu/abvss/pkg/protobuf"
=======
	"abvss/pkg/protobuf"
>>>>>>> 19b0d27 (Initial commit)
=======
	"abvss/pkg/protobuf"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
	"sync"

	"google.golang.org/protobuf/proto"
)

var Mu = new(sync.Mutex)
var Traffic = 0

// MakeDispatcheChannels dispatche messages from receiveChannel
// and make a double layer Map : (messageType) --> (id) --> (channel)
func MakeDispatcheChannels(receiveChannel chan *protobuf.Message, N uint32) *sync.Map {
	dispatcheChannels := new(sync.Map)

	go func() { //dispatcher
		for {
			m := <-(receiveChannel)
			value1, _ := dispatcheChannels.LoadOrStore(m.Type, new(sync.Map))

			var value2 any
			if m.Type == "Dec" {
				value2, _ = value1.(*sync.Map).LoadOrStore(string(m.Id), make(chan *protobuf.Message, N*N))
			} else {
				value2, _ = value1.(*sync.Map).LoadOrStore(string(m.Id), make(chan *protobuf.Message, N))
			}

			value2.(chan *protobuf.Message) <- m

			Mu.Lock()
			Traffic += proto.Size(m)
			Mu.Unlock()
		}
	}()
	return dispatcheChannels
}
