package abvss

import (
	"context"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/pkg/core"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v3"
	"log"
	"math/big"
)

type ABVSSService struct {
	*ABVSS
	*osv.OSV
	DealerBandwidthUsage int
	BandwidthUsage       int
	Receivechannel       chan *protobuf.Message
	Sendchannels         []chan *protobuf.Message
	//protobuf.UnimplementedABVSSServer
}

func NewABVSSService(n int, send []chan *protobuf.Message, receive chan *protobuf.Message) *ABVSSService {
	return &ABVSSService{
		DealerBandwidthUsage: 0,
		BandwidthUsage:       0,
		Receivechannel:       receive,
		Sendchannels:         send,
	}
}

func (vss *ABVSSService) Receive() {
	for {
		//log.Printf("node %v waiting", dkg.id)
		msg := <-vss.Receivechannel
		//log.Printf("node %v handle msg: %v from node %v", dkg.id, msg.GetType(), msg.Sender)
		go func(msg *protobuf.Message) {
			msgType := msg.GetType()
			if msgType == "Shares" {
				newmsg := core.Decapsulation(msgType, msg).(*protobuf.SharesMsg)
				zixBytes := newmsg.GetZix()
				ziyBytes := newmsg.GetZiy()
				xiyBytes := newmsg.GetXiy()
				xixBytes := newmsg.GetXix()
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
				err := vss.ObtainShares(zix, ziy, xix, xiy, int(newmsg.GetIndex()))
				if err != nil {
					log.Printf("node %v receive shares from node %v error: %v", vss.GetNodeID(), newmsg.GetFromID(), err)
				}
				log.Printf("node %v receive shares %v from node %v in instance %v", vss.GetNodeID(), newmsg.GetIndex(), newmsg.GetFromID(), newmsg.GetInstanceID())
			}
			if msgType == "LCM" {
				newmsg := core.Decapsulation(msgType, msg).(*protobuf.LCMMsg)
				vss.ReceiveLCM(newmsg)
			}
			if msgType == "SK" {

			}
			if msgType == "OSV" {
				newmsg := core.Decapsulation(msgType, msg).(*protobuf.OSVMsg)
				vss.ReceiveOSV(newmsg)
			}
		}(msg)

	}
}

/*
	func (vss *ABVSSService) ReceiveShares(shares *protobuf.SharesMsg) (*protobuf.AckMsg, error) {
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
		//log.Printf("node %v receive shares %v from node %v", vss.GetNodeID(), shares.GetIndex(), shares.GetFromID())
		return &protobuf.AckMsg{}, nil
	}
*/
func (vss *ABVSSService) ReceiveLCM(lcmmsg *protobuf.LCMMsg) {
	lcmBytes := lcmmsg.GetLcmi()
	lcm := make([]*big.Int, len(lcmBytes))
	for i := range lcmBytes {
		lcm[i] = new(big.Int).SetBytes(lcmBytes[i])
	}
	err := vss.VerifyLCM(lcm, int(lcmmsg.GetFromID()))
	if err != nil {
		log.Printf("node %v receive lcm from node %v error: %v", vss.GetNodeID(), lcmmsg.FromID, err)
		return
	}
	log.Printf("node %v receive lcm from node %v", vss.GetNodeID(), lcmmsg.FromID)
	return
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

			m := core.Encapsulation("Shares", nil, uint32(vss.nodeid), sharesmsg)

			//ctx, cancel := context.WithCancel(context.Background())
			//defer cancel()
			for j := 0; j < vss.nodenum; j++ {
				if j == vss.nodeid {
					err := vss.ObtainShares(zix, ziy, xix, xiy, i)
					if err != nil {
						log.Printf("obtain shares error: %v", err)
					}
					continue
				}
				log.Printf("node %v send shares to node %v", vss.nodeid, j)
				vss.Sendchannels[j] <- m
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

	for i := 0; i < vss.nodenum; i++ {
		lcmmsg := &protobuf.LCMMsg{
			FromID: int64(vss.nodeid),
			DestID: int64(i),
			Lcmi:   lcmBytes,
		}
		if i == vss.nodeid {
			err := vss.VerifyLCM(lcm, vss.nodeid)
			if err != nil {
				log.Printf("VerifyLCM error: %v", err)
			}
			continue
		}
		m := core.Encapsulation("LCM", nil, uint32(vss.nodeid), lcmmsg)
		vss.Sendchannels[i] <- m
	}
}

func (s *ABVSSService) OSVInit() {
	osv := s.OSV
	//log.Printf("node %v osv init in instance %v", s.id, i)
	msgs := osv.Init()
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	for _, msg := range msgs {
		protonewmsg := &protobuf.OSVMsg{
			FromID:     int64(msg.FromID),
			DestID:     int64(msg.DestID),
			InstanceID: 0,
			Mtype:      msg.Mtype,
		}
		m := core.Encapsulation("OSV", nil, uint32(s.nodeid), protonewmsg)
		s.Sendchannels[msg.DestID] <- m
	}
}

func (s *ABVSSService) ReceiveOSV(osvmsg *protobuf.OSVMsg) {
	msg := osv.Message{
		FromID: int(osvmsg.GetFromID()),
		DestID: int(osvmsg.GetDestID()),
		Mtype:  osvmsg.GetMtype(),
	}
	//log.Printf("node %v receive %v from node %v from in instance %v", s.id, osvmsg.FromID, osvmsg.Mtype, osvmsg.GetInstanceID())
	recvmsgs, err := s.OSV.Recv(msg)
	if err != nil {
		log.Printf("node %v receive msg err: %v", s.nodeid, err)
		return
	}
	for _, newmsg := range recvmsgs {
		protonewmsg := &protobuf.OSVMsg{
			FromID:     int64(newmsg.FromID),
			DestID:     int64(newmsg.DestID),
			InstanceID: osvmsg.InstanceID,
			Mtype:      newmsg.Mtype,
		}
		m := core.Encapsulation("OSV", nil, uint32(s.nodeid), protonewmsg)
		s.Sendchannels[newmsg.DestID] <- m
	}
}
