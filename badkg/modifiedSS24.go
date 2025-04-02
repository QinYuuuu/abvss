package badkg

import (
	"encoding/json"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"math/big"
	"math/rand"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/elgamal"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v3"
)

const (
	Distribute = "SS24.distribute"
	Complaint  = "SS24.complaint"
	Challenge  = "SS24.challenge"
)

type Share struct {
	Index       int64    `json:"index"`
	FShareBytes [][]byte `json:"f_share_bytes"`
	GShareBytes [][]byte `json:"g_share_bytes"`
}

type elgamalEncShare struct {
	Index       int64    `json:"index"`
	FShareBytes [][]byte `json:"f_share_bytes"`
	GShareBytes [][]byte `json:"g_share_bytes"`
}

type ChallengePoly struct {
	HBytes [][]byte `json:"h_bytes"`
}

type ACSSImpl struct {
	id, degree, nodeNum int64
	batchSize, r        int64
	sessionID           int64

	getShares *Share
	group     kyber.Group
	pkList    []kyber.Point
	sk        kyber.Scalar
	// TODO: batchsize * r
	challenge []*big.Int
	p         *big.Int
	randState *rand.Rand
	rbc       *broadcast.OptRBC
	osvNode   *osv.Instance
	*dealer
	output chan *Share
}

func NewACSSImpl(id, degree, nodeNum, batchSize, r, sessionID int64, p *big.Int, send func(int64, *protobuf.NewMessage), getChan func(string) chan *protobuf.NewMessage) *ACSSImpl {
	return &ACSSImpl{
		id:        id,
		degree:    degree,
		nodeNum:   nodeNum,
		batchSize: batchSize,
		r:         r,
		p:         p,
		sessionID: sessionID,
		output:    make(chan *Share, 16),
	}
}

type dealer struct {
	fPoly []*pkg.Poly
	gPoly []*pkg.Poly
	enc   func([]byte) ([]byte, error)
}

func (vss *ACSSImpl) Run() {
	vss.rbc.CreateNewSession(0, vss.id)
	vss.rbc.CreateNewSession(1, vss.id)
	for {
		select {
		case data := <-vss.rbc.Output(0):
			// handel RBC output
			var encShares *protobuf.ElgamalEncMultiShare
			err := proto.Unmarshal(data, encShares)
			if err != nil {
				slog.Error("proto unmarshal err:", err)
			}
			for _, encShare := range encShares.GetEncShares() {
				if encShare.GetIndex() == vss.id {
					c1 := vss.group.Point()
					c2 := vss.group.Point()
					err = c1.UnmarshalBinary(encShare.GetC1())
					if err != nil {
						return
					}
					err = c2.UnmarshalBinary(encShare.GetC2())
					if err != nil {
						return
					}
					decShareByte, err := elgamal.Decrypt(vss.group, vss.sk, c1, c2)
					if err != nil {
						return
					}
					var decShare *protobuf.SS24Share
					err = proto.Unmarshal(decShareByte, decShare)
					if err != nil {
						return
					}
					share := &Share{
						Index:       decShare.GetIndex(),
						FShareBytes: decShare.GetFShare(),
						GShareBytes: decShare.GetGShare(),
					}
					vss.handleDistribute(share)
				} else {
					slog.Info("receive", slog.Any("msg", encShare))
				}
			}
		case data := <-vss.rbc.Output(1):

		case output := <-vss.osvNode.Output():
			if output {
				vss.output <- vss.getShares
			}
		}
	}
}

func (vss *ACSSImpl) Share() {
	if vss.dealer == nil {
		slog.Error("dealer is nil")
		return
	}
	var i int64
	encShares := make([]*protobuf.ElgamalEncShare, vss.nodeNum)
	for i = 0; i < vss.nodeNum; i++ {
		share := vss.generateShare(i + 1)
		sS24Share := &protobuf.SS24Share{
			Index:  share.Index,
			FShare: share.FShareBytes,
			GShare: share.GShareBytes,
		}
		shareByte, err := proto.Marshal(sS24Share)
		if err != nil {
			slog.Error("proto marshal err:", err)
		}
		c1, c2, _ := elgamal.Encrypt(vss.group, vss.pkList[i], shareByte)
		c1Bytes, err := c1.Data()
		if err != nil {
			slog.Error("encrypt cipher err:", err)
		}
		c2Bytes, err := c2.Data()
		encShares[i] = &protobuf.ElgamalEncShare{
			Index: share.Index,
			C1:    c1Bytes,
			C2:    c2Bytes,
		}
	}
	encMultiShare := &protobuf.ElgamalEncMultiShare{
		EncShares: encShares,
	}
	encMultiShareBytes, err := proto.Marshal(encMultiShare)
	if err != nil {
		slog.Error("proto marshal err:", err)
		return
	}
	vss.rbc.StartNewBroadcast(encMultiShareBytes, vss.id, 0)
	hPoly := make([][]byte, vss.r)
	for i = 0; i < vss.batchSize; i++ {
		hPoly[i] = new(big.Int).SetInt64(vss.randState.Int63()).Bytes()
	}
	challenge := ChallengePoly{
		HBytes: hPoly,
	}
	challengeByte, _ := json.Marshal(challenge)
	// RBC
	vss.rbc.StartNewBroadcast(challengeByte, vss.id, 0)
}

func (vss *ACSSImpl) handleEncDistribute(share *elgamalEncShare) {

	return
}

func (vss *ACSSImpl) handleDistribute(share *Share) {
	fshares := make([]kyber.Scalar, vss.nodeNum)
	gShares := make([]kyber.Scalar, vss.nodeNum)
	for i, fShareByte := range share.FShareBytes {
		fshares[i] = vss.group.Scalar().SetBytes(fShareByte)
	}
	for i, gShareByte := range share.GShareBytes {
		gShares[i] = vss.group.Scalar().SetBytes(gShareByte)
	}
	// Wait vss.Challenge != nil
	hPoly := pkg.FromVecBig(vss.challenge)

	thelta := make([]*big.Int, vss.batchSize)
	right := gShares[0]
	var i int64
	for i = 0; i < vss.batchSize; i++ {
		thelta[i] = new(big.Int).SetInt64(vss.randState.Int63())
		tmp := vss.group.Scalar().SetInt64(thelta[i].Int64())
		tmp.Mul(tmp, fshares[i])
		right = right.Add(right, tmp)
	}
	h_ej := hPoly.EvalMod(new(big.Int).SetInt64(vss.id+1), vss.p)
	// check happy
	if vss.group.Scalar().SetInt64(h_ej.Int64()).Equal(right) {
		vss.osvNode.Run()
	}

}

func (vss *ACSSImpl) Output() *Share {
	share := <-vss.output
	return share
}

func (vss *ACSSImpl) generateShare(index int64) *Share {
	if vss.dealer == nil {
		slog.Error("not dealer")
		return nil
	}
	//f := make([]*big.Int, vss.batchSize)
	fByte := make([][]byte, vss.batchSize)
	for i, poly := range vss.dealer.fPoly {
		f := poly.EvalMod(new(big.Int).SetInt64(index), vss.p)
		fByte[i] = f.Bytes()
	}
	//g := make([]*big.Int, vss.r)
	gByte := make([][]byte, vss.r)
	for i, poly := range vss.dealer.gPoly {
		g := poly.EvalMod(new(big.Int).SetInt64(index), vss.p)
		gByte[i] = g.Bytes()
	}
	return &Share{
		Index:       index,
		FShareBytes: fByte,
		GShareBytes: gByte,
	}
}
