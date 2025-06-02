package harts

import (
	"fmt"
	"log/slog"
	"math/big"
	"strconv"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/commit/pedersen"
	"github.com/QinYuuuu/abvss/crypto/zk/nizk"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"google.golang.org/protobuf/proto"

	"go.dedis.ch/kyber/v4"
)

type HAVSSImpl struct {
	n, tc, tr    int64 // degree tc in Y, tr in X
	id, dealerID int64
	instanceID   string

	p     *big.Int
	group kyber.Group

	voteSenders []bool
	doneSenders []bool
	voteCounter int64
	doneCounter int64

	sentDone bool
	sentVote bool

	columnXList []int64
	columnYList []kyber.Scalar
	polyCi      chan *pkg.PolyKyberImpl
	si          chan []kyber.Point
	*dealer
	rbc           *broadcast.OptRBC
	pedersenParam *pedersen.VectorParam
	nizkIPAParam  *nizk.NizkIPAParam

	send    func(*protobuf.HartsHavssMessage)
	receive func() chan *protobuf.HartsHavssMessage
}

func NewHAVSSImpl(
	id, n, tc, tr, dealerID int64,
	instanceID string,
	group kyber.Group,
	nizkIPAParam *nizk.NizkIPAParam,
	pedersenParam *pedersen.VectorParam,
) *HAVSSImpl {
	return &HAVSSImpl{
		n:           n,
		tc:          tc,
		tr:          tr,
		id:          id,
		dealerID:    dealerID,
		instanceID:  instanceID,
		group:       group,
		voteSenders: make([]bool, n),
		doneSenders: make([]bool, n),
		sentDone:    false,
		sentVote:    false,
		voteCounter: 0,
		doneCounter: 0,
		polyCi:      make(chan *pkg.PolyKyberImpl, 1),
		si:          make(chan []kyber.Point, 1),

		pedersenParam: pedersenParam,
		nizkIPAParam:  nizkIPAParam,
	}
}

func (vss *HAVSSImpl) Run() {
	vss.rbc.CreateNewSession("HAVSS_COMMIT_"+strconv.FormatInt(vss.dealerID, 10), vss.dealerID)
	vss.rbc.Run()
	go vss.messageLoop()
}

func (vss *HAVSSImpl) messageLoop() {
	for {
		select {
		case msg := <-vss.receive():
			switch msg.Type {
			case Row:
				// handle Row
				vss.handleRow(msg)
			case Column:
				// handle ready
				vss.handleColumn(msg)
			case Vote:
				// handle vote
				vss.handleVote(msg)
			case Done:
				vss.handleDone(msg.FromID)
			default:
				slog.Error(fmt.Sprintf("[node %v] [HAVSS: %v] unhandled default case", vss.id, vss.instanceID))
			}
		case output := <-vss.rbc.Output("HAVSS_COMMIT_" + strconv.FormatInt(vss.dealerID, 10)):
			var msg protobuf.HartsCommitMessage
			err := proto.Unmarshal(output, &msg)
			if err != nil {
				slog.Error("proto harts msg unmarshal", slog.String("error", err.Error()))
			}
			vss.handleCommit(&msg)
		}
	}
}

func (vss *HAVSSImpl) handleRow(msg *protobuf.HartsHavssMessage) {
	slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] handle row", vss.id, vss.instanceID))
	var rowMsg protobuf.HartsRowVectorMessage
	err := proto.Unmarshal(msg.Value, &rowMsg)
	if err != nil {
		slog.Error("proto harts row msg unmarshal", slog.String("error", err.Error()))
	}
	for i := int64(0); i < vss.n; i++ {
		xIndex := vss.group.Scalar().SetInt64(i + 1)
		cijBytes := rowMsg.CI[i]
		cij := vss.group.Scalar()
		err := cij.UnmarshalBinary(cijBytes)
		if err != nil {
			slog.Error("proto harts row msg unmarshal", slog.String("error", err.Error()))
		}
		proofMsg := rowMsg.PiI[i]
		proof, err := nizk.UnmarshalNizkIPAProofFromProto(vss.group, proofMsg)
		if err != nil {
			slog.Error("proto harts row msg unmarshal", slog.String("error", err.Error()))
		}
		verifyResult, err := vss.nizkIPAParam.VerifyForPoly(proof, xIndex)
		if err != nil {
			slog.Error("verify nizkIPA proof", slog.String("error", err.Error()))
		}
		if verifyResult {
			columnMsg := &protobuf.HartsColumnMessage{
				CIj: rowMsg.CI[i],
				PIj: proofMsg,
			}
			columnMsgBytes, err := proto.Marshal(columnMsg)
			if err != nil {
				slog.Error("proto harts column msg marshal", slog.String("error", err.Error()))
			}
			sendMsg := &protobuf.HartsHavssMessage{
				FromID:     vss.id,
				DestID:     i,
				InstanceID: vss.instanceID,
				Type:       Column,
				Value:      columnMsgBytes,
			}
			vss.send(sendMsg)
			voteMsg := &protobuf.HartsHavssMessage{
				FromID:     vss.id,
				DestID:     i,
				InstanceID: vss.instanceID,
				Type:       Vote,
			}
			vss.send(voteMsg)
		} else {
			slog.Error(fmt.Sprintf("[node %v] [HAVSS: %v] verify nizkIPA proof failed", vss.id, vss.instanceID))
		}
	}
}

func (vss *HAVSSImpl) handleCommit(msg *protobuf.HartsCommitMessage) {
	slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] handle commit", vss.id, vss.instanceID))

	SiBytes := msg.GetSi()
	Si := make([]kyber.Point, len(SiBytes))
	xi := make([]int64, len(SiBytes))
	for i, siByte := range SiBytes {
		xi[i] = int64(i)
		Si[i] = vss.group.Point()
		err := Si[i].UnmarshalBinary(siByte)
		if err != nil {
			slog.Error("proto harts msg unmarshal", slog.String("error", err.Error()))
		}
	}
	// calculate S_j = g^C_i(j)
	siList := make([]kyber.Point, vss.n+1)
	for i := int64(0); i < vss.n+1; i++ {
		index := vss.group.Scalar().SetInt64(i)
		si, err := pkg.InterpolationExpAtIndexKyber(xi, Si, vss.group, index)
		if err != nil {
			slog.Error("interpolation at zero", slog.String("error", err.Error()))
		}
		slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] S_%d = %v", vss.id, vss.instanceID, i, si))
		siList[i] = si
	}
	vss.si <- siList
}

func (vss *HAVSSImpl) handleColumn(msg *protobuf.HartsHavssMessage) {
	slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] handle column from %v", vss.id, vss.instanceID, msg.FromID))
	var column protobuf.HartsColumnMessage
	err := proto.Unmarshal(msg.Value, &column)
	if err != nil {
		slog.Error("proto harts column msg unmarshal", slog.String("error", err.Error()))
	}
	proofMsg := column.GetPIj()
	proof, err := nizk.UnmarshalNizkIPAProofFromProto(vss.group, proofMsg)
	index := vss.group.Scalar().SetInt64(vss.id + 1)
	verifyResult, err := vss.nizkIPAParam.VerifyForPoly(proof, index)
	if err != nil {
		slog.Error("verify nizkIPA proof", slog.String("error", err.Error()))
	}
	if verifyResult {
		// slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] verify nizkIPA proof success", vss.id, vss.instanceID))
		c_ij := vss.group.Scalar()
		err := c_ij.UnmarshalBinary(column.GetCIj())
		if err != nil {
			slog.Error("proto harts column msg unmarshal", slog.String("error", err.Error()))
			return
		}
		vss.columnXList = append(vss.columnXList, msg.FromID)
		vss.columnYList = append(vss.columnYList, c_ij)
		if len(vss.columnXList) == int(vss.tc)+1 {
			c_i, err := pkg.InterpolationKyber(vss.columnXList, vss.columnYList, vss.group)
			if err != nil {
				slog.Error("interpolation", slog.String("error", err.Error()))
				return
			}
			vss.polyCi <- c_i
			slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] calculate c_i = %v", vss.id, vss.instanceID, c_i))
		}
	} else {
		slog.Error(fmt.Sprintf("[node %v] [HAVSS: %v] verify nizkIPA proof failed", vss.id, vss.instanceID))
	}
	// slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] handle column from %v success", vss.id, vss.instanceID, msg.FromID))
}

func (vss *HAVSSImpl) handleVote(msg *protobuf.HartsHavssMessage) {
	sender := msg.FromID
	if vss.voteSenders[sender] {
		slog.Error("have receive vote", slog.Any("fromID", sender))
	}
	slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] handle vote from %v", vss.id, vss.instanceID, sender))
	vss.voteCounter++
	if vss.voteCounter >= vss.n-vss.tc && !vss.sentDone {
		// send done
		vss.sentDone = true
		var i int64
		for i = 0; i < vss.n; i++ {
			msg := &protobuf.HartsHavssMessage{
				FromID:     vss.id,
				DestID:     i,
				InstanceID: vss.instanceID,
				Type:       Done,
			}
			vss.send(msg)
		}
	}
}

func (vss *HAVSSImpl) handleDone(sender int64) {
	if vss.doneSenders[sender] {
		slog.Error("have receive vote", slog.Any("fromID", sender))
	}
	slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] handle done from %v", vss.id, vss.instanceID, sender))
	vss.doneCounter++
	if vss.doneCounter >= vss.tc+1 && !vss.sentDone {
		// send done
		vss.sentDone = true
		var i int64
		for i = 0; i < vss.n; i++ {
			msg := &protobuf.HartsHavssMessage{
				FromID:     vss.id,
				DestID:     i,
				InstanceID: vss.instanceID,
				Type:       Done,
			}
			vss.send(msg)
		}
	}
	if vss.doneCounter >= vss.n-vss.tc {
		// terminate
		slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] terminate", vss.id, vss.instanceID))
	}
}

func (vss *HAVSSImpl) Output() *HavssOutput {
	output := &HavssOutput{}

	ci, ok := <-vss.polyCi
	if !ok {
		slog.Error("polyCi channel closed")
		return nil
	}
	output._Ci = ci

	s, ok := <-vss.si
	if !ok {
		slog.Error("si channel closed")
		return nil
	}
	output._S = s

	return output
}

func (vss *HAVSSImpl) Reconstruct() {
	return
}
