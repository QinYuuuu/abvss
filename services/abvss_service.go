package services

import (
	"context"
	"github.com/QinYuuuu/abvss/protobuf"
	"github.com/QinYuuuu/abvss/protocol"
	"log"
	"math/big"
)

type ABVSSService struct {
	protocol.ABVSS
	protobuf.UnimplementedABVSSServiceServer
}

func (vss *ABVSSService) ReceiveShares(ctx context.Context, shares *protobuf.SharesMessage) (*protobuf.AckMsg, error) {
	if int(shares.FromID) != vss.GetNodeID() {
		log.Printf("node %v receive shares wrong desID %v", vss.GetNodeID(), shares.GetDestID())
		return &protobuf.AckMsg{}, nil
	}
	ziBytes := shares.GetZi()
	xiBytes := shares.GetXi()
	zi := make([]protocol.Cipher, len(ziBytes))
	xi := make([]protocol.Cipher, len(xiBytes))
	for i := range ziBytes {
		zi[i] = new(big.Int).SetBytes(ziBytes[i])
	}
	for i := range xiBytes {
		xi[i] = new(big.Int).SetBytes(xiBytes[i])
	}
	err := vss.ObtainShares(zi, xi)
	if err != nil {
		log.Printf("node %v receive shares from node %v error: %v", vss.GetNodeID(), shares.FromID, err)
		return &protobuf.AckMsg{}, nil
	}
	log.Printf("node %v receive shares from node %v", vss.GetNodeID(), shares.FromID)
	return &protobuf.AckMsg{}, nil
}

func (vss *ABVSSService) ReceiveLCM(ctx context.Context, lcm *protobuf.LCMMessage) (*protobuf.AckMsg, error) {
	if int(lcm.FromID) != vss.GetNodeID() {
		log.Printf("node %v receive shares wrong desID %v", vss.GetNodeID(), lcm.GetDestID())
		return &protobuf.AckMsg{}, nil
	}
	lcmBytes := lcm.GetLcmi()
	lcmi := make([]*big.Int, len(lcmBytes))
	for i := range lcmBytes {
		lcmi[i] = new(big.Int).SetBytes(lcmBytes[i])
	}
	err := vss.VerifyLCM(lcmi, vss.GetNodeID())
	if err != nil {
		log.Printf("node %v receive lcm from node %v error: %v", vss.GetNodeID(), lcm.FromID, err)
		return &protobuf.AckMsg{}, nil
	}
	log.Printf("node %v receive shares from node %v", vss.GetNodeID(), lcm.FromID)
	return &protobuf.AckMsg{}, nil
}
