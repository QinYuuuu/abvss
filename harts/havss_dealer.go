package harts

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/QinYuuuu/abvss/crypto/commit/pedersen"
	"github.com/QinYuuuu/abvss/crypto/zk/nizk"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v4"
	"google.golang.org/protobuf/proto"
)

type dealer struct {
	biPoly *pkg.BivariatePolyKyberImpl
}

func NewHAVSSDealerImpl(
	id, n, tc, tr int64,
	instanceID string,
	group kyber.Group,
	nizkIPAParam *nizk.NizkIPAParam,
	pedersenParam *pedersen.VectorParam,
	havssNetwork HAVSSNetwork,
) *HAVSSImpl {
	impl := NewHAVSSImpl(id, n, tc, tr, id, instanceID, group, nizkIPAParam, pedersenParam, havssNetwork)
	biPoly, err := pkg.NewRandBiPolyKyber(int(tr), int(tc), group)
	if err != nil {
		slog.Error("new rand bi poly", slog.String("error", err.Error()))
		return nil
	}
	secret, err := biPoly.GetCoefficient(0, 0)
	if err != nil {
		slog.Error("get coefficient", slog.String("error", err.Error()))
		return nil
	}
	slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] init secret: %v", id, instanceID, secret.String()))
	impl.dealer = &dealer{
		biPoly: biPoly,
	}
	return impl
}

func (vss *HAVSSImpl) SetSecret(s kyber.Scalar) {
	if vss.dealer == nil {
		slog.Error("not dealer, cannot set secret")
	}
	vss.dealer.biPoly.SetCoefficientScalar(0, 0, s)
	slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] set secret: %v", vss.id, vss.instanceID, s))
}

func (vss *HAVSSImpl) CommitAndDistribute() {
	if vss.dealer == nil {
		slog.Error("not dealer, cannot commit")
	}
	// calculate commit and row message
	CiPoly := make([]*pkg.PolyKyberImpl, vss.n)
	CiPolyComm := make([]kyber.Point, vss.n)

	for i := range vss.n {
		xIndex := vss.group.Scalar().SetInt64(i)
		CiPoly[i] = vss.biPoly.EvalAtXMod(xIndex)
		// slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] generate CiPoly[%v]: %v", vss.id, vss.instanceID, i, CiPoly[i].ToString()))
		aVec := CiPoly[i].GetAllCoefficient()
		CiPolyComm[i] = vss.pedersenParam.Commit(aVec)
		cijBytes := make([][]byte, vss.n)
		proofMsgs := make([]*protobuf.NizkIPAProof, vss.n)
		for j := range vss.n {
			yIndex := vss.group.Scalar().SetInt64(j + 1)
			cijByte, err := CiPoly[i].EvalMod(yIndex).MarshalBinary()
			if err != nil {
				slog.Error("marshal cij value", slog.String("error", err.Error()))
				return
			}
			cijBytes[j] = cijByte

			proof, err := vss.nizkIPAParam.ProveForPoly(aVec, yIndex)
			if err != nil {
				slog.Error("prove for poly", slog.String("error", err.Error()))
				return
			}
			// only for debug
			/*{
				result, err := vss.nizkIPAParam.VerifyForPoly(proof, yIndex)
				if err != nil {
					slog.Error("verify for poly", slog.String("error", err.Error()))
					return
				}
				slog.Info(fmt.Sprintf("[node %v] [HAVSS: %v] generate proof correctnes: %v", vss.id, vss.instanceID, result))
			}*/
			proofMsg, err := nizk.MarshalNizkIPAProofToProto(proof)
			if err != nil {
				slog.Error("marshal proof", slog.String("error", err.Error()))
				return
			}
			proofMsgs[j] = proofMsg
		}
		rowMsg := &protobuf.HartsRowVectorMessage{
			CI:  cijBytes,
			PiI: proofMsgs,
		}
		rowMsgBytes, err := proto.Marshal(rowMsg)
		if err != nil {
			slog.Error("marshal row msg", slog.String("error", err.Error()))
			return
		}
		vss.send(&protobuf.HartsHavssMessage{
			FromID:     vss.id,
			DestID:     int64(i),
			InstanceID: vss.instanceID,
			Type:       Row,
			Value:      rowMsgBytes,
		})

	}

	// calculate s_i = g^C_i(0)
	S := make([]kyber.Point, vss.n)
	for i := range vss.n {
		Ci0 := CiPoly[i].EvalMod(vss.group.Scalar().Zero())
		Ci0Scale := vss.group.Scalar().Set(Ci0)
		S[i] = vss.group.Point().Mul(Ci0Scale, nil)
	}
	// calculate pi
	commitMsgs := &protobuf.HartsCommitMessage{
		Si: make([][]byte, vss.tr+1),
	}
	for i := int64(0); i < vss.tr+1; i++ {
		SiBytes, err := S[i].MarshalBinary()
		if err != nil {
			slog.Error("marshal commitMsg.Si", slog.String("error", err.Error()))
			return
		}
		commitMsgs.Si[i] = SiBytes
	}
	coomitMsgBytes, err := proto.Marshal(commitMsgs)
	if err != nil {
		slog.Error("marshal commit msg error", slog.String("error", err.Error()))
		return
	}
	// broadcast commit message
	sessionID := vss.instanceID + "_COMMIT_" + strconv.FormatInt(vss.dealerID, 10)
	vss.rbc.StartNewBroadcast(coomitMsgBytes, vss.id, sessionID)
}
