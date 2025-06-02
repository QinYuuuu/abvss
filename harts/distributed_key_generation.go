package harts

import (
	"context"
	"log/slog"
	"math/big"

	"github.com/QinYuuuu/abvss/crypto/signature/RSA"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	Proposal int64 = iota
	Signature
)

// Party represents a participant in the protocol
type Party struct {
	id, n, t int64
	secret   *big.Int
	dealerID int64

	verKeys [][]byte

	signatureChan chan []byte
	Nonces        []*big.Int
	proposeSet    []int64
	sigSet        []singleSignature
	sharingSet    map[int]sharing

	avssInstances []*HAVSSImpl
	send          func(message *protobuf.HartsMessage)
	receive       func() chan *protobuf.HartsMessage
}

func NewParty(id, n, t, dealerID int64) *Party {
	return &Party{
		id:       id,
		n:        n,
		t:        t,
		dealerID: dealerID,
	}
}

// sharing represents a reconstructed sharing
type sharing struct {
	index     int
	OriginalS *big.Int
	Shares    map[int]*big.Int
}

type singleSignature struct {
	index     int64
	signature []byte
}

// Run runs the Packed Asynchronous DKG algorithm for a party
func (p *Party) Run() {
	for _, avss := range p.avssInstances {
		avss.Run()
	}

	avssFinishCounter := int64(0)
	dealerSet := make([]int64, 0)
	avssOutputChan := p.getAVSSOutput()
	for {
		avssOutput := <-avssOutputChan
		avssFinishCounter++
		if avssFinishCounter == p.n-p.t {
			dealerSet = append(dealerSet, avssOutput)
		}
	}
}

func (p *Party) messageLoop(ctx context.Context) {
	// Process messages
	select {
	case msg := <-p.receive():
		switch msg.Type {
		case "proposal":
			// handle proposal
		case "signature":
			// handle signature message
			p.handleSignature(ctx, msg)
		}
	}
}

func (p *Party) handleSignature(ctx context.Context, msg *protobuf.HartsMessage) {
	if msg.Type != "signature" {
		return
	}
	verKey := p.verKeys[msg.FromID]
	proposeSetBytes := p.convertProposeSetToBytes()
	// in goroutine, wait until proposeSet not nil
	err := RSA.Verify(proposeSetBytes, msg.Value, verKey)
	if err != nil {
		slog.Error("Verify harts signature failed", slog.String("err", err.Error()))
	}
	p.sigSet = append(p.sigSet, singleSignature{
		index:     msg.FromID,
		signature: msg.Value,
	})
}

func (p *Party) getAVSSOutput() chan int64 {
	finishChan := make(chan int64, p.n-p.t)
	for i := int64(0); i < p.n; i++ {
		go func(sessionID int64) {
			_ = p.avssInstances[sessionID].Output()
			finishChan <- sessionID
		}(i)
	}
	return finishChan
}

func (p *Party) handleProposal(ctx context.Context, msg *protobuf.HartsMessage) {}

func (p *Party) convertProposeSetToBytes() []byte {
	proposeSetMsg := &protobuf.HartsProposeSet{
		Index: p.proposeSet,
	}
	proposeSetBytes, err := proto.Marshal(proposeSetMsg)
	if err != nil {
		slog.Error("Marshal harts propose set failed", slog.String("err", err.Error()))
	}
	return proposeSetBytes
}
