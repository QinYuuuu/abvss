package badkg

import (
	"fmt"
	"log/slog"
	"math/big"
	"math/rand"
	"strconv"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v4"
	"google.golang.org/protobuf/proto"
)

type shareList struct {
	x []*big.Int
	y []*big.Int
}

type rbcOutput struct {
	index int64
	qele  []byte
}

type acssOutput struct {
	index int64
	data  *protobuf.SS24Share
}

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
	vShares                        shareList

	setQss            []int64
	setQrbc           []int64
	acssOutputted     []*acssOutput
	acssOutputCounter int64
	rbcOutputted      []*rbcOutput
	rbcOutputCounter  int64

	mvba1       *MVBA
	mvba2       *MVBA
	rbc         *broadcast.OptRBC
	output      chan []*protobuf.SS24Share
	send        func(*protobuf.DKGMessage)
	receiveChan func() chan *protobuf.DKGMessage
}

func NewDKGImpl(id, degree, nodeNum, batchSize int64, p *big.Int, group kyber.Group, g kyber.Point, mvba1 *MVBA, mvba2 *MVBA, output chan []*protobuf.SS24Share, send func(*protobuf.DKGMessage), receiveChan func() chan *protobuf.DKGMessage) *DKGImpl {
	acssImpls := make([]*ACSSImpl, nodeNum)
	s := make([]*big.Int, batchSize+1)
	randSource := rand.New(rand.NewSource(id))
	for i := range s {
		s[i] = new(big.Int).Rand(randSource, p)
	}
	var i int64
	for i = 0; i < nodeNum; i++ {
		if i == id {
			acssImpls[i] = NewACSSImplDealer(id, degree, nodeNum, batchSize+1, 1, id, s, p, group)
			continue
		}
		acssImpls[i] = NewACSSImpl(id, degree, nodeNum, batchSize+1, 1, i, i, p, group)
	}
	return &DKGImpl{
		acssImpls:     acssImpls,
		id:            id,
		degree:        degree,
		nodeNum:       nodeNum,
		batchSize:     batchSize,
		p:             p,
		group:         group,
		g:             g,
		hasVote1:      false,
		hasVote2:      false,
		acssOutputted: make([]*acssOutput, nodeNum),
		rbcOutputted:  make([]*rbcOutput, nodeNum),
		mvba1:         mvba1,
		mvba2:         mvba2,
		output:        output,
		send:          send,
		receiveChan:   receiveChan,
	}
}

func (dkg *DKGImpl) Run() {
	for _, acss := range dkg.acssImpls {
		acss.Run()
	}
	dkg.mvba1.Run()
	dkg.mvba2.Run()

	Qss := dkg.phase1()
	QssMvba1Input := make([]int64, len(Qss))
	for i := 0; i < len(Qss); i++ {
		QssMvba1Input[i] = Qss[i].index
	}
	dkg.mvba1.Input(QssMvba1Input)

	dkg.setQss = <-dkg.mvba1.Output()

	Qrbc := dkg.phase2()
	QssMvba2Input := make([]int64, len(Qss))
	for i := 0; i < len(Qrbc); i++ {
		QssMvba2Input[i] = Qrbc[i].index
	}
	dkg.mvba2.Input(QssMvba2Input)

	dkg.setQrbc = <-dkg.mvba2.Output()

	dkg.phase3()

	dkg.phase4()
}

func (dkg *DKGImpl) getACSSOutput() chan *acssOutput {
	finishChan := make(chan *acssOutput, dkg.nodeNum-dkg.degree)
	var i int64
	for i = 0; i < dkg.nodeNum; i++ {
		go func(sessionId int64) {
			output := dkg.acssImpls[sessionId].Output()
			if output != nil {
				slog.Info(fmt.Sprintf("[node %v] output in ACSS %v", dkg.id, output))
				finishChan <- &acssOutput{
					index: i,
					data:  output,
				}
			}
		}(i)
	}
	return finishChan
}

func (dkg *DKGImpl) getRBCOutput() chan *rbcOutput {
	rbcOutputChans := make([]chan []byte, dkg.nodeNum)
	var i int64
	for i = 0; i < dkg.nodeNum; i++ {
		rbcOutputChans[i] = dkg.rbc.Output(strconv.FormatInt(i, 10) + "RBC_on_Qele")
	}
	finishChan := make(chan *rbcOutput, dkg.nodeNum)
	// goroutine for every rbc output
	for index, ch := range rbcOutputChans {
		go func(index int, c chan []byte) {
			val, ok := <-c
			if ok {
				slog.Info("rbc output", slog.Any("msg", val))
				finishChan <- &rbcOutput{
					index: int64(index),
					qele:  val,
				}
			}
		}(index, ch)
	}
	return finishChan
}

func (dkg *DKGImpl) phase1() []*acssOutput {
	dkg.acssImpls[dkg.id].Share()
	finishChan := dkg.getACSSOutput()
	for {
		newACSSOutput := <-finishChan
		dkg.acssOutputted[newACSSOutput.index] = newACSSOutput
		dkg.acssOutputCounter++
		if dkg.acssOutputCounter >= dkg.nodeNum-dkg.degree {
			result := make([]*acssOutput, 0)
			var j int64
			for j = 0; j < dkg.nodeNum; j++ {
				if dkg.acssOutputted[j] != nil {
					result = append(result, dkg.acssOutputted[j])
				}
			}
		}
	}
}

func (dkg *DKGImpl) phase2() []*rbcOutput {
	Qss := dkg.setQss
	Qele := make([][]kyber.Point, len(Qss))
	for i, index := range Qss {
		dkg.sk[i], _ = pkg.MatrixMulVector(dkg.hyperMatrix, dkg.fShares[index])
		Qele[i] = make([]kyber.Point, dkg.batchSize+1)
		var j int64
		for j = 0; j < dkg.batchSize; j++ {
			fShareScalar := dkg.group.Scalar().SetInt64(dkg.fShares[index][j].Int64())
			Qele[i][j] = dkg.group.Point().Mul(fShareScalar, dkg.g)
		}
		gShareScalar := dkg.group.Scalar().SetInt64(dkg.gShares[index].Int64())
		Qele[i][dkg.batchSize] = dkg.group.Point().Mul(gShareScalar, dkg.g)
	}
	dkg.rbc.StartNewBroadcast([]byte(""), dkg.id, strconv.FormatInt(dkg.id, 10)+"RBC_on_Qele")

	finishChan := dkg.getRBCOutput()
	for {
		newRbcOutput := <-finishChan
		dkg.rbcOutputted[newRbcOutput.index] = newRbcOutput
		dkg.rbcOutputCounter++
		if dkg.rbcOutputCounter >= dkg.nodeNum {
			result := make([]*rbcOutput, 0)
			var j int64
			for j = 0; j < dkg.nodeNum; j++ {
				if dkg.rbcOutputted[j] != nil {
					result = append(result, dkg.rbcOutputted[j])
				}

			}
			return result
		}
	}
}

func (dkg *DKGImpl) phase3() {
	for {
		select {
		case msg := <-dkg.receiveChan():
			var share protobuf.DKGShareMessage
			err := proto.Unmarshal(msg.Value, &share)
			if err != nil {
				slog.Error("proto unmarshal error")
			}
			dkg.vShares.x = append(dkg.vShares.x, new(big.Int).SetInt64(share.Index))
			dkg.vShares.y = append(dkg.vShares.y, new(big.Int).SetBytes(share.Vj))
			interpolation, err := pkg.LagrangeInterpolation(dkg.vShares.x, dkg.vShares.y, dkg.p)
			if err != nil {
				return
			}
			_, err = interpolation.GetCoefficient(0)
			if err != nil {
				return
			}
		}
	}
}

func (dkg *DKGImpl) phase4() {

}

func (dkg *DKGImpl) handle() {
	// |Qss| = n-t
	Qss := <-dkg.mvba1.Output()

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
		product, err := pkg.DotProduct(rj, dkg.fShares[index])
		if err != nil {
			slog.Error("calculate r_j * f_j(i)", slog.Any("Error", err))
		}
		vi = vi.Add(vi, product)
	}
	// broadcast vi
}
