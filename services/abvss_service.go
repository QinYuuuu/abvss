package services

import (
	"context"
	"github.com/QinYuuuu/abvss/protobuf"
	"log"
	"math/big"
)

type ABVSSService struct {
	*ABVSS
	Clients []protobuf.ABVSSClient
	protobuf.UnimplementedABVSSServer
}

func NewABVSSService(n int) *ABVSSService {
	return &ABVSSService{
		Clients: make([]protobuf.ABVSSClient, n),
	}
}

func (vss *ABVSSService) ReceiveShares(ctx context.Context, shares *protobuf.SharesMsg) (*protobuf.AckMsg, error) {
	if int(shares.DestID) != vss.GetNodeID() {
		log.Printf("node %v receive shares wrong desID %v", vss.GetNodeID(), shares.GetDestID())
		return &protobuf.AckMsg{}, nil
	}
	ziBytes := shares.GetZi()
	xiBytes := shares.GetXi()
	zi := make([]Cipher, len(ziBytes))
	xi := make([]Cipher, len(xiBytes))
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

func (vss *ABVSSService) ReceiveLCM(ctx context.Context, lcm *protobuf.LCMMsg) (*protobuf.AckMsg, error) {
	if int(lcm.DestID) != vss.GetNodeID() {
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

func (vss *ABVSSService) ReconstructLCM(ctx context.Context, sk *protobuf.SKMsg) (*protobuf.AckMsg, error) {
	if int(sk.DestID) != vss.GetNodeID() {
		log.Printf("node %v receive shares wrong desID %v", vss.GetNodeID(), sk.GetDestID())
		return &protobuf.AckMsg{}, nil
	}
	return &protobuf.AckMsg{}, nil
}

func (vss *ABVSSService) ReceiveRecShares(ctx context.Context, sk *protobuf.SKMsg) (*protobuf.AckMsg, error) {
	if int(sk.DestID) != vss.GetNodeID() {
		log.Printf("node %v receive shares wrong desID %v", vss.GetNodeID(), sk.GetDestID())
		return &protobuf.AckMsg{}, nil
	}
	return &protobuf.AckMsg{}, nil
}

func (vss *ABVSSService) SecretSharing(pk []PublicKey, s []*big.Int) {
	err := vss.DistributorInit(pk, s)
	if err != nil {
		log.Printf("init error: %v", err)
	}

	err = vss.SamplePoly()
	if err != nil {
		log.Printf("sample poly error: %v", err)
	}
	for i := 0; i < vss.nodenum; i++ {
		zi, xi, err := vss.GenerateShares(i)
		if err != nil {
			log.Printf("generate shares error: %v", err)
		}
		ziBytes := make([][]byte, len(zi))
		xiBytes := make([][]byte, len(xi))
		for i := range ziBytes {
			ziBytes[i] = zi[i].(*big.Int).Bytes()
		}
		for i := range xiBytes {
			xiBytes[i] = xi[i].(*big.Int).Bytes()
		}
		_, err = vss.Clients[i].ReceiveShares(context.Background(), &protobuf.SharesMsg{
			FromID: int64(vss.nodeid),
			DestID: int64(i),
			Zi:     ziBytes,
			Xi:     xiBytes,
		})
		if err != nil {
			log.Printf("send shares to node %v error: %v", i, err)
		}
	}
}
