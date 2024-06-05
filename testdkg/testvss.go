package main

import (
	"github.com/QinYuuuu/abvss/internal/abvss"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/network"
)

type ABVSSNode struct {
	*abvss.ABVSSService
	*osv.OSVService
	*network.Peer
}
