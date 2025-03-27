package badkg

import (
	"encoding/json"
	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"log/slog"
	"math/big"
	"math/rand"
	"strconv"
)

const Distribute = "SS24.distribute"
const Complaint = "SS24.complaint"

type Share struct {
	Index       int64    `json:"index"`
	FShareBytes [][]byte `json:"f_share_bytes"`
	GShareBytes [][]byte `json:"g_share_bytes"`
}

type ACSSImpl struct {
	id, degree, nodeNum int64
	batchSize, r        int64
	sessionID           int64
	p                   *big.Int
	randState           *rand.Rand
	rbcSession          *broadcast.Session
	osvNode             *osv.OSV
	*dealer
	output  chan *Share
	send    func(int64, *protobuf.Message)
	getChan func(string) chan *protobuf.Message
}

type dealer struct {
	fPoly []*pkg.Poly
	gPoly []*pkg.Poly
	enc   func([]byte) ([]byte, error)
}

func (vss *ACSSImpl) Run() {
	var msg *protobuf.Message
	for {
		select {
		case msg = <-vss.getChan(Distribute):
			// handel share

		case msg = <-vss.getChan(Complaint):
			// handle complain
		}
	}
}

func (vss *ACSSImpl) Share() {
	//as dealer
	if vss.dealer != nil {
		var i int64
		for i = 0; i < vss.nodeNum; i++ {
			share := vss.generateShare(i + 1)
			shareByte, _ := json.Marshal(share)
			msg := &protobuf.Message{
				Type:   Distribute,
				Id:     []byte(strconv.FormatInt(i, 10)),
				Sender: uint32(vss.id),
				Data:   shareByte,
			}
			vss.send(i, msg)
		}
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
