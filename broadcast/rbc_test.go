package broadcast

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRBC(t *testing.T) {
	var n int64 = 4
	var f int64 = 1
	rbc := make([]*RBC, n)
	msgchannel := make([]chan RBCMessage, n)
	var i int64
	for i = 0; i < n; i++ {
		rbc[i] = NewRBC(i, n, f, 0)
		msgchannel[i] = make(chan RBCMessage, 2048)
	}
	rbc[0].SetLeader()
	m := []byte("test")
	msgs, err := rbc[0].Send(m)
	assert.Nil(t, err, "err in send")
	for i = 0; i < n; i++ {
		msgchannel[i] <- msgs[i]
	}
	var wg sync.WaitGroup
	wg.Add(int(n))
	for i = 0; i < n; i++ {
		go func(i int64) {
			for msg := range msgchannel[i] {
				rmsgs, err := rbc[i].Recv(msg)
				assert.Nil(t, err, "err in recv")
				for _, rmsg := range rmsgs {
					msgchannel[rmsg.DestID] <- rmsg
				}
				if rbc[i].Output {
					wg.Done()
					break
				}
			}

		}(i)
	}
	wg.Wait()
	fmt.Printf("s")
}
