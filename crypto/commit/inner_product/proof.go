package inner_product

import (
	"log/slog"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v4"
	"google.golang.org/protobuf/proto"
)

type Proof struct {
	lVec       []kyber.Point
	rVec       []kyber.Point
	aVecToSend []kyber.Scalar
}

func MarshalToBinary(proof *Proof) ([]byte, error) {
	proofMsg, err := MarshalToProto(proof)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(proofMsg)
}

func UnmarshalFromBinary(group kyber.Group, data []byte) (*Proof, error) {
	proofMsg := &protobuf.InnerProductProof{}
	err := proto.Unmarshal(data, proofMsg)
	if err != nil {
		slog.Error("proof msg unmarshal", slog.String("error", err.Error()))
		return nil, err
	}
	return UnmarshalFromProto(group, proofMsg)
}

func MarshalToProto(proof *Proof) (*protobuf.InnerProductProof, error) {
	lVecBytes := make([][]byte, len(proof.lVec))
	rVecBytes := make([][]byte, len(proof.rVec))
	aVecToSendBytes := make([][]byte, len(proof.aVecToSend))
	for i, l := range proof.lVec {
		lVecByte, err := l.MarshalBinary()
		if err != nil {
			slog.Error("proof lVec marshal", slog.Any("index", i), slog.String("error", err.Error()))
			return nil, err
		}
		lVecBytes[i] = lVecByte
	}
	for i, r := range proof.rVec {
		rVecByte, err := r.MarshalBinary()
		if err != nil {
			slog.Error("proof rVec marshal", slog.Any("index", i), slog.String("error", err.Error()))
			return nil, err
		}
		rVecBytes[i] = rVecByte
	}
	for i, a := range proof.aVecToSend {
		aVecToSendByte, err := a.MarshalBinary()
		if err != nil {
			slog.Error("proof aVecToSend marshal", slog.Any("index", i), slog.String("error", err.Error()))
			return nil, err
		}
		aVecToSendBytes[i] = aVecToSendByte
	}
	proofMsg := &protobuf.InnerProductProof{
		LVec:       lVecBytes,
		RVec:       rVecBytes,
		AVecToSend: aVecToSendBytes,
	}
	return proofMsg, nil
}

func UnmarshalFromProto(group kyber.Group, proofMsg *protobuf.InnerProductProof) (*Proof, error) {
	lVec := make([]kyber.Point, len(proofMsg.GetLVec()))
	rVec := make([]kyber.Point, len(proofMsg.GetRVec()))
	aVecToSend := make([]kyber.Scalar, len(proofMsg.GetAVecToSend()))
	for i, l := range proofMsg.GetLVec() {
		lVec[i] = group.Point()
		err := lVec[i].UnmarshalBinary(l)
		if err != nil {
			slog.Error("proof lVec unmarshal", slog.Any("index", i), slog.String("error", err.Error()))
			return nil, err
		}
	}
	for i, r := range proofMsg.GetRVec() {
		rVec[i] = group.Point()
		err := rVec[i].UnmarshalBinary(r)
		if err != nil {
			slog.Error("proof rVec unmarshal", slog.Any("index", i), slog.String("error", err.Error()))
			return nil, err
		}
	}
	for i, a := range proofMsg.GetAVecToSend() {
		aVecToSend[i] = group.Scalar()
		err := aVecToSend[i].UnmarshalBinary(a)
		if err != nil {
			slog.Error("proof aVecToSend unmarshal", slog.Any("index", i), slog.String("error", err.Error()))
			return nil, err
		}
	}
	return &Proof{
		lVec:       lVec,
		rVec:       rVec,
		aVecToSend: aVecToSend,
	}, nil
}
