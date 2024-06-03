package abvss

import (
	"context"
	"github.com/QinYuuuu/abvss/crypto/paillier"
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
	ziBytes := shares.GetZi()
	xiBytes := shares.GetXi()
	zi := make([]*big.Int, len(ziBytes))
	xi := make([]*big.Int, len(xiBytes))
	for i := range ziBytes {
		zi[i] = new(big.Int).SetBytes(ziBytes[i])
	}
	for i := range xiBytes {
		xi[i] = new(big.Int).SetBytes(xiBytes[i])
	}
	err := vss.ObtainShares(zi, xi, int(shares.Index))
	if err != nil {
		log.Printf("node %v receive shares from node %v error: %v", vss.GetNodeID(), shares.GetFromID(), err)
		return &protobuf.AckMsg{}, nil
	}
	//log.Printf("node %v receive shares %v from node %v", vss.GetNodeID(), shares.GetIndex(), shares.GetFromID())
	return &protobuf.AckMsg{}, nil
}

func (vss *ABVSSService) ReceiveLCM(ctx context.Context, lcmmsg *protobuf.LCMMsg) (*protobuf.AckMsg, error) {
	lcmBytes := lcmmsg.GetLcmi()
	lcm := make([]*big.Int, len(lcmBytes))
	for i := range lcmBytes {
		lcm[i] = new(big.Int).SetBytes(lcmBytes[i])
	}
	err := vss.VerifyLCM(lcm, int(lcmmsg.GetFromID()))
	if err != nil {
		log.Printf("node %v receive lcm from node %v error: %v", vss.GetNodeID(), lcmmsg.FromID, err)
		return &protobuf.AckMsg{}, nil
	}
	//log.Printf("node %v receive lcm from node %v", vss.GetNodeID(), lcmmsg.FromID)
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

func (vss *ABVSSService) SecretSharing(pk []paillier.PublicKey, s []*big.Int) {
	err := vss.DistributorInit(pk, s)
	if err != nil {
		log.Printf("init error: %v", err)
	}

	err = vss.SamplePoly()
	if err != nil {
		log.Printf("sample poly error: %v", err)
	}
	for i := 0; i < vss.nodenum; i++ {
		go func(i int) {
			zi, xi, err := vss.GenerateShares(i)
			if err != nil {
				log.Printf("generate shares error: %v", err)
			}

			ziBytes := make([][]byte, len(zi))
			xiBytes := make([][]byte, len(xi))
			for j := range ziBytes {
				ziBytes[j] = zi[j].Bytes()
			}
			for j := range xiBytes {
				xiBytes[j] = xi[j].Bytes()
			}
			sharesmsg := &protobuf.SharesMsg{
				FromID: int64(vss.nodeid),
				Index:  int64(i),
				Zi:     ziBytes,
				Xi:     xiBytes,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			for j := 0; j < vss.nodenum; j++ {
				if j == vss.nodeid {
					err := vss.ObtainShares(zi, xi, i)
					if err != nil {
						log.Printf("obtain shares error: %v", err)
					}
					continue
				}
				_, err = vss.Clients[j].ReceiveShares(ctx, sharesmsg)
				if err != nil {
					log.Printf("send shares to node %v error: %v", i, err)
				}
			}
		}(i)

	}
}

func (vss *ABVSSService) BroadcastLCM() {
	lcm, err := vss.ConstructLCM()
	if err != nil {
		log.Printf("construct lcm error: %v", err)
	}
	lcmBytes := make([][]byte, len(lcm))
	for i := range lcm {
		lcmBytes[i] = lcm[i].Bytes()
	}
	lcmmsg := &protobuf.LCMMsg{
		FromID: int64(vss.nodeid),
		Lcmi:   lcmBytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < vss.nodenum; i++ {
		if i == vss.nodeid {
			err := vss.VerifyLCM(lcm, vss.nodeid)
			if err != nil {
				log.Printf("VerifyLCM error: %v", err)
			}
			continue
		}
		_, err = vss.Clients[i].ReceiveLCM(ctx, lcmmsg)
		if err != nil {
			log.Printf("send shares to node %v error: %v", i, err)
		}
	}
}
