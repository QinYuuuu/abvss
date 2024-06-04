package abvss

import (
	"context"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v3"
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
	zixBytes := shares.GetZix()
	ziyBytes := shares.GetZiy()
	xixBytes := shares.GetXix()
	xiyBytes := shares.GetXiy()
	zix := make([]kyber.Point, len(zixBytes))
	ziy := make([]kyber.Point, len(ziyBytes))
	xix := make([]kyber.Point, len(xixBytes))
	xiy := make([]kyber.Point, len(xiyBytes))
	for i := range zixBytes {
		zix[i] = vss.Curve.Point()
		ziy[i] = vss.Curve.Point()
		_ = zix[i].UnmarshalBinary(zixBytes[i])
		_ = ziy[i].UnmarshalBinary(ziyBytes[i])
	}
	for i := range xixBytes {
		xix[i] = vss.Curve.Point()
		xiy[i] = vss.Curve.Point()
		_ = xix[i].UnmarshalBinary(xixBytes[i])
		_ = xiy[i].UnmarshalBinary(xiyBytes[i])
	}
	err := vss.ObtainShares(zix, ziy, xix, xiy, int(shares.Index))
	if err != nil {
		log.Printf("node %v receive shares from node %v error: %v", vss.GetNodeID(), shares.GetFromID(), err)
		return &protobuf.AckMsg{}, nil
	}
	log.Printf("node %v receive shares %v from node %v", vss.GetNodeID(), shares.GetIndex(), shares.GetFromID())
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

func (vss *ABVSSService) SecretSharing(pk []kyber.Point, s []*big.Int) {
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
			zix, ziy, xix, xiy, err := vss.GenerateShares(i)
			if err != nil {
				log.Printf("generate shares error: %v", err)
			}

			zixBytes := make([][]byte, len(zix))
			ziyBytes := make([][]byte, len(ziy))
			xixBytes := make([][]byte, len(xix))
			xiyBytes := make([][]byte, len(xiy))
			for j := range zix {
				zixBytes[j], _ = zix[j].MarshalBinary()
				ziyBytes[j], _ = ziy[j].MarshalBinary()
			}
			for j := range xix {
				xixBytes[j], _ = xix[j].MarshalBinary()
				xiyBytes[j], _ = xiy[j].MarshalBinary()
			}
			sharesmsg := &protobuf.SharesMsg{
				FromID: int64(vss.nodeid),
				Index:  int64(i),
				Zix:    zixBytes,
				Ziy:    ziyBytes,
				Xix:    xixBytes,
				Xiy:    xiyBytes,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			for j := 0; j < vss.nodenum; j++ {
				if j == vss.nodeid {
					err := vss.ObtainShares(zix, ziy, xix, xiy, i)
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
