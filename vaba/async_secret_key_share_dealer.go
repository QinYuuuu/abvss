package vaba

import (
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/hasher"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"math/big"
	"strconv"
)

type dealer struct {
	secret *big.Int // Secret to share (only for dealer)
	poly   *pkg.Poly
}

// NewASKSDealer creates a dealer for the ASKS protocol
func NewASKSDealer(id, n, t int64, secret *big.Int, prime *big.Int) *ASKSImpl {
	party := NewASKS(id, n, t, id, prime)
	// Let p(·) be a random degree-t polynomial
	polynomial, err := pkg.NewRandPoly(int(t), prime)
	if err != nil {
		slog.Error("generate rand poly", slog.Any("error", err))
		return nil
	}
	err = polynomial.SetCoefficientBig(0, secret)
	if err != nil {
		slog.Error("set secret", slog.Any("error", err))
		return nil
	}
	party.dealer = &dealer{poly: polynomial}
	return party
}

// Share initiates the sharing phase
func (p *ASKSImpl) Share() {
	if p.dealer == nil {
		slog.Error(fmt.Sprintf("[node %v] not dealer", p.id))
	}
	// SHARING PHASE
	// If p_i is the dealer
	if p.dealerID == p.id {
		// Let hⱼ = H(j, p(j)) for each j ∈ [n]
		hashVector := make([][]byte, p.n)
		shares := make([]*big.Int, p.n)
		var j int64
		for j = 0; j < p.n; j++ {
			shares[j] = p.poly.EvalMod(new(big.Int).SetInt64(j+1), p.p) // 1-indexed
			hashVector[j] = hasher.MD5Hasher(append([]byte{byte(j)}, shares[j].Bytes()...))
		}
		hasVec := &protobuf.ASKSHashVector{
			HashByte: hashVector,
		}
		hashVecByte, err := proto.Marshal(hasVec)
		if err != nil {
			slog.Error("proto marshal hasVec", slog.Any("error", err))
		}
		// Broadcast h = [h₁, h₂, ..., hₙ] using a RBC
		p.rbc.StartNewBroadcast(hashVecByte, p.id, "asks"+strconv.FormatInt(p.id, 10)+"hashVec")

		// Send (SHARE, p(j)) to party j
		for j = 0; j < p.n; j++ {
			shareMsg := &protobuf.ASKSMessage{
				FromID:     p.id,
				DestID:     j,
				InstanceID: p.instanceID,
				Type:       SHARE,
				Value:      shares[j].Bytes(),
			}
			slog.Info(fmt.Sprintf("[node %v] send", p.id), slog.Any("msg", shareMsg))
			p.send(shareMsg)
		}
	}
	// Start a reliable agreement protocol instance RA
	p.ra.Run()
}
