package network_test

import (
	"testing"

	"github.com/QinYuuuu/abvss/network"
)

func TestSR(t *testing.T) {
	network.TestTCPNodeCommunication(t)
}

func TestConn(t *testing.T) {
	network.TestTCPNode_Concurrency(t)
}

func TestRpc(t *testing.T) {
	network.TestRPCNodeCommunication(t)
}
