package abdkg

import (
	"context"
	"github.com/QinYuuuu/abvss/internal/abvss"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v3"
	"log"
	"math/big"
)

type ABDKGService struct {
	id, nodenum int
	Vss         []*abvss.ABVSS
	Osv         []*osv.OSV
	Clients     []protobuf.ABDKGClient
	protobuf.UnimplementedABDKGServer
	Bandwidth int
}

func NewABDKGService(id, n int) *ABDKGService {
	return &ABDKGService{
		id:        id,
		nodenum:   n,
		Bandwidth: 0,
		Clients:   make([]protobuf.ABDKGClient, n),
	}
}
func (dkg *ABDKGService) ReceiveShares(ctx context.Context, shares *protobuf.SharesMsg) (*protobuf.AckMsg, error) {
	vss := dkg.Vss[shares.GetInstanceID()]
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
	//log.Printf("node %v receive shares %v from node %v in instance %v", vss.GetNodeID(), shares.GetIndex(), shares.GetFromID(), shares.GetInstanceID())
	return &protobuf.AckMsg{}, nil
}

func (dkg *ABDKGService) ReceiveLCM(ctx context.Context, lcmmsg *protobuf.LCMMsg) (*protobuf.AckMsg, error) {
	vss := dkg.Vss[lcmmsg.GetInstanceID()]
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
	//log.Printf("node %v receive lcm from node %v in instance %v", vss.GetNodeID(), lcmmsg.FromID, lcmmsg.InstanceID)
	return &protobuf.AckMsg{}, nil
}

func (dkg *ABDKGService) ReconstructLCM(ctx context.Context, sk *protobuf.SKMsg) (*protobuf.AckMsg, error) {
	vss := dkg.Vss[sk.GetInstanceID()]
	if int(sk.DestID) != vss.GetNodeID() {
		log.Printf("node %v receive shares wrong desID %v", vss.GetNodeID(), sk.GetDestID())
		return &protobuf.AckMsg{}, nil
	}
	return &protobuf.AckMsg{}, nil
}

func (dkg *ABDKGService) ReceiveRecShares(ctx context.Context, sk *protobuf.SKMsg) (*protobuf.AckMsg, error) {
	vss := dkg.Vss[sk.GetInstanceID()]
	if int(sk.DestID) != vss.GetNodeID() {
		log.Printf("node %v receive shares wrong desID %v", vss.GetNodeID(), sk.GetDestID())
		return &protobuf.AckMsg{}, nil
	}
	return &protobuf.AckMsg{}, nil
}

func (dkg *ABDKGService) SecretSharing(pk []kyber.Point, s []*big.Int) {
	vss := dkg.Vss[dkg.id]
	err := vss.DistributorInit(pk, s)
	if err != nil {
		log.Printf("init error: %v", err)
	}

	err = vss.SamplePoly()
	if err != nil {
		log.Printf("sample poly error: %v", err)
	}
	for i := 0; i < dkg.nodenum; i++ {
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
				FromID:     int64(dkg.id),
				InstanceID: int64(dkg.id),
				Index:      int64(i),
				Zix:        zixBytes,
				Ziy:        ziyBytes,
				Xix:        xixBytes,
				Xiy:        xiyBytes,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			for j := 0; j < dkg.nodenum; j++ {
				if j == dkg.id {
					err := vss.ObtainShares(zix, ziy, xix, xiy, i)
					if err != nil {
						log.Printf("obtain shares error: %v", err)
					}
					continue
				}
				_, err = dkg.Clients[j].ReceiveShares(ctx, sharesmsg)
				if err != nil {
					log.Printf("send shares to node %v error: %v", i, err)
				}
			}
		}(i)

	}
}

func (dkg *ABDKGService) BroadcastLCM(i int) {
	vss := dkg.Vss[i]
	lcm, err := vss.ConstructLCM()
	if err != nil {
		log.Printf("construct lcm error: %v", err)
	}
	lcmBytes := make([][]byte, len(lcm))
	for i := range lcm {
		lcmBytes[i] = lcm[i].Bytes()
	}
	lcmmsg := &protobuf.LCMMsg{
		FromID:     int64(dkg.id),
		InstanceID: int64(i),
		Lcmi:       lcmBytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < dkg.nodenum; i++ {
		if i == dkg.id {
			err := vss.VerifyLCM(lcm, dkg.id)
			if err != nil {
				log.Printf("VerifyLCM error: %v", err)
			}
			continue
		}
		_, err = dkg.Clients[i].ReceiveLCM(ctx, lcmmsg)
		if err != nil {
			log.Printf("send shares to node %v error: %v", i, err)
		}
	}
}

func (s *ABDKGService) Init(i int) {
	osv := s.Osv[i]
	//log.Printf("node %v osv init in instance %v", s.id, i)
	msgs := osv.Init()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, msg := range msgs {
		protonewmsg := &protobuf.OSVMsg{
			FromID:     int64(msg.FromID),
			DestID:     int64(msg.DestID),
			InstanceID: int64(i),
			Mtype:      msg.Mtype,
		}
		_, err := s.Clients[msg.DestID].ReceiveOSV(ctx, protonewmsg)
		if err != nil {
			log.Printf("node %v init err: %v", s.id, err)
		}
	}
}

func (s *ABDKGService) ReceiveOSV(ctx context.Context, osvmsg *protobuf.OSVMsg) (*protobuf.AckMsg, error) {
	msg := osv.Message{
		FromID: int(osvmsg.GetFromID()),
		DestID: int(osvmsg.GetDestID()),
		Mtype:  osvmsg.GetMtype(),
	}
	//log.Printf("node %v receive %v from node %v from in instance %v", s.id, osvmsg.FromID, osvmsg.Mtype, osvmsg.GetInstanceID())
	recvmsgs, err := s.Osv[osvmsg.InstanceID].Recv(msg)
	if err != nil {
		log.Printf("node %v receive msg err: %v", s.id, err)
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, newmsg := range recvmsgs {
		protonewmsg := &protobuf.OSVMsg{
			FromID:     int64(newmsg.FromID),
			DestID:     int64(newmsg.DestID),
			InstanceID: osvmsg.InstanceID,
			Mtype:      newmsg.Mtype,
		}
		_, err := s.Clients[newmsg.DestID].ReceiveOSV(ctx, protonewmsg)
		if err != nil {
			log.Printf("node %v receive msg err: %v", s.id, err)
		}
	}
	return &protobuf.AckMsg{}, nil
}
