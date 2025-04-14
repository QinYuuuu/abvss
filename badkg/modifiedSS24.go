package badkg

import (
	"fmt"
	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/elgamal"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v3"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"math/big"
	"math/rand"
	"strconv"
)

const (
	Distribute = "SS24.distribute"
	Complaint  = "SS24.complaint"
	Challenge  = "SS24.challenge"
)

type elgamalEncShare struct {
	Index       int64
	FShareBytes [][]byte
	GShareBytes [][]byte
}

type ChallengePoly struct {
	HBytes [][]byte
}

type ACSSImpl struct {
	id, degree, nodeNum int64
	batchSize, r        int64
	sessionID           int64
	dealerID            int64

	shareReceived     bool
	challengeReceived bool
	verified          bool

	myShare       *protobuf.SS24Share
	encShares     *protobuf.ElgamalEncShare
	challengePoly []*pkg.Poly
	theta         [][]*big.Int

	group     kyber.Group
	p         *big.Int
	pkList    []kyber.Point
	sk        kyber.Scalar
	randState *rand.Rand

	rbc     *broadcast.OptRBC
	osvNode *osv.Instance
	*dealer
	output chan *protobuf.SS24Share
}

func (vss *ACSSImpl) GetTheta() [][]int64 {
	result := make([][]int64, len(vss.theta))
	for i := range result {
		result[i] = make([]int64, len(vss.theta[i]))
		for j := range result[i] {
			result[i][j] = vss.theta[i][j].Int64()
		}
	}
	return result
}

func NewACSSImpl(id, degree, nodeNum, batchSize, r, sessionID, dealerID int64, p *big.Int, group kyber.Group) *ACSSImpl {
	randState := rand.New(rand.NewSource(1))
	theta := make([][]*big.Int, r)
	for i := int64(0); i < r; i++ {
		theta[i] = make([]*big.Int, batchSize)
		for j := int64(0); j < batchSize; j++ {
			theta[i][j] = new(big.Int).SetInt64(randState.Int63())
		}
	}
	return &ACSSImpl{
		id:        id,
		degree:    degree,
		nodeNum:   nodeNum,
		dealerID:  dealerID,
		sessionID: sessionID,
		batchSize: batchSize,
		randState: randState,
		r:         r,
		p:         p,
		group:     group,
		theta:     theta,
		output:    make(chan *protobuf.SS24Share, 1),
	}
}

func (vss *ACSSImpl) Run() {
	vss.rbc.Run()
	vss.rbc.CreateNewSession(strconv.FormatInt(vss.dealerID, 10)+"0", vss.dealerID)
	vss.rbc.CreateNewSession(strconv.FormatInt(vss.dealerID, 10)+"1", vss.dealerID)
	go vss.messageLoop()
}

func (vss *ACSSImpl) messageLoop() {
	var err error
	for {
		select {
		case data := <-vss.rbc.Output(strconv.FormatInt(vss.dealerID, 10) + "0"):
			// handel RBC output
			slog.Debug(fmt.Sprintf("[node %v] receive encMultiShare", vss.id), slog.Any("data", data))
			var encShares protobuf.AESEncMultiShare
			err = proto.Unmarshal(data, &encShares)
			if err != nil {
				slog.Error("proto unmarshal", slog.Any("error", err))
			}
			for i, encShare := range encShares.GetElGamalEncShares() {
				if encShare.GetIndex() == vss.id+1 {
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
					aesKey, err := elgamal.Decrypt(vss.group, vss.sk, c1, c2)
					if err != nil {
						return
					}

					aesEncShare := encShares.GetAesEncShares()[i]
					decShareByte, err := aesDec(aesKey, aesEncShare.Cipher)
					if err != nil {
						return
					}
					var decShare protobuf.SS24Share
					err = proto.Unmarshal(decShareByte, &decShare)
					if err != nil {
						return
					}
					vss.myShare = &decShare
					slog.Debug(fmt.Sprintf("[node %v]", vss.id), slog.Any("share", decShare.FShare))
					vss.shareReceived = true
				} else {
					slog.Debug(fmt.Sprintf("[node %v] receive", vss.id), slog.Any("enc share", encShare))
				}
			}
			if vss.shareReceived && vss.challengeReceived && !vss.verified && vss.id != vss.dealerID {
				vss.verifyDistribute()
				vss.verified = true
			}
			if !vss.verified && vss.id == vss.dealerID {
				slog.Info(fmt.Sprintf("[node %v] dealer run osv", vss.id))
				vss.osvNode.Run()
				vss.verified = true
			}
		case data := <-vss.rbc.Output(strconv.FormatInt(vss.dealerID, 10) + "1"):
			slog.Info(fmt.Sprintf("[node %v] receive", vss.id), slog.Any("challenge poly", data))
			var challengePoly protobuf.ChallengePoly
			err = proto.Unmarshal(data, &challengePoly)
			if err != nil {
				slog.Error("proto unmarshal", slog.Any("error", err))
			}
			vss.challengePoly = make([]*pkg.Poly, len(challengePoly.GetPolys()))
			for i, poly := range challengePoly.GetPolys() {
				coeffsBytes := poly.GetCoefficient()
				coeffs := make([]*big.Int, len(coeffsBytes))
				for j := range coeffs {
					coeffs[j] = new(big.Int).SetBytes(coeffsBytes[j])
				}
				vss.challengePoly[i] = pkg.FromVecBig(coeffs)
			}
			vss.challengeReceived = true
			if vss.shareReceived && vss.challengeReceived && !vss.verified && vss.id != vss.dealerID {
				vss.verifyDistribute()
				vss.verified = true
			}
			if !vss.verified && vss.id == vss.dealerID {
				slog.Info(fmt.Sprintf("[node %v] dealer run osv", vss.id))
				vss.osvNode.Run()
				vss.verified = true
			}
		case output := <-vss.osvNode.Output():
			slog.Info(fmt.Sprintf("[node %v] osv output", vss.id))
			if output {
				vss.output <- vss.myShare
			}
		}
	}
}

func (vss *ACSSImpl) verifyDistribute() {
	fShares := make([]*big.Int, vss.batchSize)
	gShares := make([]*big.Int, vss.r)
	for i, fShareByte := range vss.myShare.FShare {
		fShares[i] = new(big.Int).SetBytes(fShareByte)
	}
	for i, gShareByte := range vss.myShare.GShare {
		gShares[i] = new(big.Int).SetBytes(gShareByte)
	}
	// Wait vss.Challenge != nil
	slog.Info(fmt.Sprintf("[node %v] try verify", vss.id))
	for i, hPoly := range vss.challengePoly {
		right := gShares[i]
		var j int64
		for j = 0; j < vss.batchSize; j++ {
			tmp := new(big.Int).Set(vss.theta[i][j])
			tmp.Mul(tmp, fShares[j])
			right.Add(right, tmp)
			right.Mod(right, vss.p)
		}
		hej := hPoly.EvalMod(new(big.Int).SetInt64(vss.id+1), vss.p)
		// check happy
		slog.Debug("verify", slog.Any("want", hej.Int64()), slog.Any("got", right.Int64()))
		if hej.Cmp(right) == 0 {
			slog.Debug(fmt.Sprintf("[node %v] run osv", vss.id))
			vss.osvNode.Run()
		}
	}
}

func (vss *ACSSImpl) Output() *protobuf.SS24Share {
	share := <-vss.output
	return share
}
