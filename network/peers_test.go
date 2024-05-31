package network

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPeer_Serve(t *testing.T) {
	iplist := []string{"127.0.0.1:8000", "127.0.0.1:8001", "127.0.0.1:8002"}
	for i := 0; i < 3; i++ {
		peer, err := NewPeer(3, i, iplist)
		assert.Nil(t, err)
		peer.Serve(false)
	}
}

func TestPeer_Connect(t *testing.T) {
}
