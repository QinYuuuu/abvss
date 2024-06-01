package rbc

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestNewState(t *testing.T) {
	n := 4
	f := 1
	rbc := make([]*RBC, n)
	msgchannel := make([]chan Message, n)
	for i := 0; i < n; i++ {
		rbc[i] = NewRBC(i, n, f, 0)
		msgchannel[i] = make(chan Message, 2048)
	}
	rbc[0].SetLeader()
	m := []byte("test")
	msgs, err := rbc[0].Send(m)
	assert.Nil(t, err, "err in send")
	for i := 0; i < n; i++ {
		msgchannel[i] <- msgs[i]
	}
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			for msg := range msgchannel[i] {
				rmsgs, err := rbc[i].Recv(msg)
				assert.Nil(t, err, "err in recv")
				for _, rmsg := range rmsgs {
					msgchannel[rmsg.destID] <- rmsg
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
