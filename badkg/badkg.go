package badkg

import (
	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v3"
	"log/slog"
	"math/big"
)

type DKGImpl struct {
	acssImpls                      []*ACSSImpl
	id, degree, nodeNum, batchSize int64
	p                              *big.Int
	fShares                        [][]*big.Int
	gShares                        []*big.Int
	hyperMatrix                    [][]*big.Int
	sk                             [][]*big.Int
	group                          kyber.Group
	g                              kyber.Point
	hasVote1                       bool
	hasVote2                       bool
	mvba1                          *MVBA
	mvba2                          *MVBA
	rbc                            *broadcast.OptRBC
	output                         chan *Share
	send                           func(int64, *protobuf.Message)
	getChan                        func(string) chan *protobuf.Message
}

func NewDKGImpl(id, degree, nodeNum, batchSize int64, p *big.Int, group kyber.Group, g kyber.Point, mvba1 *MVBA, mvba2 *MVBA, output chan *Share, send func(int64, *protobuf.Message), getChan func(string) chan *protobuf.Message) *DKGImpl {
	acssImpls := make([]*ACSSImpl, nodeNum)
	return &DKGImpl{
		acssImpls: acssImpls,
		id:        id,
		degree:    degree,
		nodeNum:   nodeNum,
		batchSize: batchSize,
		p:         p,
		group:     group,
		g:         g,
		hasVote1:  false,
		hasVote2:  false,
		mvba1:     mvba1,
		mvba2:     mvba2,
		output:    output,
		send:      send,
		getChan:   getChan,
	}
}

func (dkg *DKGImpl) Run() {
	for _, acss := range dkg.acssImpls {
		acss.Run()
	}
	for {
		select {}
	}
}

func (dkg *DKGImpl) handle() {
	// |Qss|=n-t
	Qss := <-dkg.mvba1.Output()
	Qele := make([][]kyber.Point, len(Qss))
	for i, index := range Qss {
		dkg.sk[i], _ = utils.MatrixMulVector(dkg.hyperMatrix, dkg.fShares[index])
		Qele[i] = make([]kyber.Point, dkg.batchSize+1)
		var j int64
		for j = 0; j < dkg.batchSize; j++ {
			fShareScalar := dkg.group.Scalar().SetInt64(dkg.fShares[index][j].Int64())
			Qele[i][j] = Qele[i][j].Mul(fShareScalar, dkg.g)
		}
	}
	// RBC on Qele
	// |Q_rbc| = n-t
	// var i int64
	// Qrbc := make([]int64, 0)
	// for i = 0;i<dkg.nodeNum;i++{
	//	 output := dkg.rbc.Output(0)
	// }
	vi := new(big.Int).SetInt64(0)
	for _, index := range Qss {
		// challenge value
		rj := make([]*big.Int, dkg.batchSize)
		var j int64
		for j = 0; j < dkg.batchSize; j++ {
			rj[j] = utils.RandomNum(dkg.p)
		}
		product, err := utils.DotProduct(rj, dkg.fShares[index])
		if err != nil {
			slog.Error("calculate r_j * f_j(i)", slog.Any("Error", err))
		}
		vi = vi.Add(vi, product)
	}
	// broadcast vi
}
