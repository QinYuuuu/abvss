package harts

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"slices"

	"github.com/QinYuuuu/abvss/crypto/signature/rsa"
	"github.com/QinYuuuu/abvss/internal/party"
	"github.com/QinYuuuu/abvss/pkg"

	"github.com/QinYuuuu/abvss/crypto/commit/pedersen"
	"github.com/QinYuuuu/abvss/crypto/zk/nizk"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/share"
	"google.golang.org/protobuf/proto"
)

const (
	Proposal int64 = iota
	Signature
)

// Party represents a participant in the protocol
type Party struct {
	id, n, tc, tr int64
	secret        *big.Int

	verKeys       [][]byte
	signKey       []byte
	signatureChan chan []byte
	nonces        []*big.Int
	dealerSet     []int64
	proposeSet    []int64
	sigSet        []singleSignature
	sharingSet    map[int]sharing
	mvbaReady     chan bool
	mvbaInput     []byte
	output        chan []kyber.Scalar
	avssInstances []*HAVSSImpl
	superMatrix   [][]kyber.Scalar
	// wrapper of smvba
	mvbaParty  *party.HonestParty
	mvbaSig    []byte
	mvbaSigPK  *share.PubPoly
	mvbaSigSK  []*share.PriShare
	mvbaOutput chan []byte

	DKGNetwork
}

type DKGNetwork struct {
	bandwidthCounter int

	send     func(message *protobuf.HartsMessage)
	receive  func() chan *protobuf.HartsMessage
	havssMap map[string]HAVSSNetwork
}

func NewParty(id, n, tc, tr int64,
	group kyber.Group,
	nizkIPAParam *nizk.NizkIPAParam,
	pedersenParam *pedersen.VectorParam,
	dkgNetwork DKGNetwork,
) *Party {
	return &Party{
		id:         id,
		n:          n,
		tc:         tc,
		tr:         tr,
		mvbaReady:  make(chan bool, 1),
		mvbaOutput: make(chan []byte, 1),
		output:     make(chan []kyber.Scalar, 1),
		DKGNetwork: dkgNetwork,
	}
}

// sharing represents a reconstructed sharing
type sharing struct {
	index     int
	originalS *big.Int
	shares    map[int]*big.Int
}

type singleSignature struct {
	index     int64
	signature []byte
}

// Run runs the Packed Asynchronous DKG algorithm for a party
func (p *Party) Run() {
	ctx, _ := context.WithCancel(context.Background())
	for _, avss := range p.avssInstances {
		avss.Run(ctx)
	}
	p.avssInstances[p.id].CommitAndDistribute()
	go p.messageLoop()
	go p.handleHAVSSOutput()
	go p.mvbaRun()
}

func (p *Party) messageLoop() {
	// Process messages
	for {
		select {
		case msg := <-p.receive():
			if msg.Value != nil {
				p.bandwidthCounter += len(msg.Value)
			}
			switch msg.Type {
			case "proposal":
				// handle proposal
				p.handleProposal(msg)
			case "signature":
				// handle signature message
				p.handleSignature(msg)
			}
		case <-p.mvbaOutput:
			slog.Debug(fmt.Sprintf("[node %v] [DKG] start reconstruct for havss %v", p.id, p.mvbaInput))
			shares := make([]kyber.Scalar, 0)
			for _, index := range p.proposeSet {
				shares = append(shares, p.avssInstances[index].Rec())
			}
			nonces := p.appplySI(shares)
			p.output <- nonces
		}
	}
}

func (p *Party) handleProposal(msg *protobuf.HartsMessage) {
	proposeSetBytes := msg.Value
	var proposeSet protobuf.HartsProposeSet
	err := proto.Unmarshal(proposeSetBytes, &proposeSet)
	if err != nil {
		slog.Error("Unmarshal harts propose set failed", slog.String("err", err.Error()))
	}
	slog.Debug(fmt.Sprintf("[node %v] [DKG] receive proposal from %v, propose set: %v", p.id, msg.FromID, proposeSet.Index))
	sig, err := rsa.Sign(proposeSetBytes, p.signKey)
	if err != nil {
		slog.Error("Sign harts propose set failed", slog.String("err", err.Error()))
	}
	/*{
		verKey := p.verKeys[p.id]
		err = rsa.Verify(proposeSetBytes, sig, verKey)
		if err != nil {
			slog.Error("Verify harts signature failed", slog.String("err", err.Error()))
		} else {
			slog.Debug(fmt.Sprintf("[node %v] [DKG] generate signature success, use verkey %s", p.id, verKey))
		}
	}*/
	sigMsg := &protobuf.HartsSignatureMessage{
		Signature:  sig,
		ProposeSet: &proposeSet,
	}
	sigBytes, err := proto.Marshal(sigMsg)
	if err != nil {
		slog.Error("Marshal harts signature message failed", slog.String("err", err.Error()))
	}
	dkgMsg := &protobuf.HartsMessage{
		Type:   "signature",
		FromID: p.id,
		DestID: msg.FromID,
		Value:  sigBytes,
	}
	slog.Debug(fmt.Sprintf("[node %v] [DKG] send signature to %v", p.id, msg.FromID))
	p.send(dkgMsg)
}

func (p *Party) handleHAVSSOutput() {
	avssFinishCounter := int64(0)
	avssOutputChan := p.getAVSSOutput()
	for {
		avssOutput := <-avssOutputChan
		avssFinishCounter++
		p.dealerSet = append(p.dealerSet, avssOutput)
		if avssFinishCounter == p.n-p.tc {
			// send proposal
			proposeSet := make([]int64, p.n-p.tc)
			copy(proposeSet, p.dealerSet)
			p.proposeSet = proposeSet
			value := p.convertProposeSetToBytes(proposeSet)
			slog.Debug(fmt.Sprintf("[node %v] [DKG] broadcast proposal", p.id))
			for i := int64(0); i < p.n; i++ {
				proposalMsg := &protobuf.HartsMessage{
					Type:   "proposal",
					FromID: p.id,
					DestID: i,
					Value:  value,
				}
				if i == p.id {
					p.receive() <- proposalMsg
				} else {
					p.send(proposalMsg)
				}
			}
			break
		}
	}
	slog.Debug(fmt.Sprintf("[node %v] [DKG] finish handle havss output", p.id))
}

func (p *Party) handleSignature(msg *protobuf.HartsMessage) {
	if msg.Type != "signature" {
		return
	}
	verKey := p.verKeys[msg.FromID]
	sigMsgBytes := msg.Value
	var sigMsg protobuf.HartsSignatureMessage
	err := proto.Unmarshal(sigMsgBytes, &sigMsg)
	proposeSet := sigMsg.ProposeSet
	slog.Debug(fmt.Sprintf("[node %v] [DKG] receive signature from %v, propose set: %v", p.id, msg.FromID, proposeSet.Index))
	proposeSetBytes, err := proto.Marshal(proposeSet)
	if err != nil {
		slog.Error("Marshal harts propose set failed", slog.String("err", err.Error()))
	}
	// in goroutine, wait until proposeSet not nil
	err = rsa.Verify(proposeSetBytes, sigMsg.Signature, verKey)
	if err != nil {
		slog.Error(fmt.Sprintf("[node %v] [DKG] verify signature from %v failed", p.id, msg.FromID), slog.Any("verkey", verKey))
	}
	p.sigSet = append(p.sigSet, singleSignature{
		index: msg.FromID,
		// signature: msg.Value,
	})
	if len(p.sigSet) == int(p.tc+1) {
		input := make([]byte, 0)
		for _, sig := range p.sigSet {
			input = append(input, byte(sig.index))
		}
		slices.Sort(input)
		p.mvbaInput = input
		slog.Debug(fmt.Sprintf("[node %v] [DKG] receive %v signature, start mvba", p.id, len(p.sigSet)))
		p.mvbaReady <- true
	}
}

func (p *Party) getAVSSOutput() chan int64 {
	finishChan := make(chan int64, p.n-p.tc)
	for i := int64(0); i < p.n; i++ {
		go func(sessionID int64) {
			_ = p.avssInstances[sessionID].Output()
			slog.Debug(fmt.Sprintf("[node %v] [DKG] receive avss output from %v", p.id, sessionID))
			finishChan <- sessionID
		}(i)
	}
	return finishChan
}

func (p *Party) appplySI(shares []kyber.Scalar) []kyber.Scalar {
	// calculate nonces = superMatrix * shares
	nonces, err := pkg.MatrixMulVectorKyber(p.superMatrix, shares)
	if err != nil {
		slog.Error("MatrixMulVectorKyber failed", slog.String("err", err.Error()))
	}
	return nonces
}

func (p *Party) convertProposeSetToBytes(proposeSet []int64) []byte {
	// sort proposeSet
	for i := 0; i < len(proposeSet); i++ {
		for j := i + 1; j < len(proposeSet); j++ {
			if proposeSet[i] > proposeSet[j] {
				proposeSet[i], proposeSet[j] = proposeSet[j], proposeSet[i]
			}
		}
	}
	slog.Debug(fmt.Sprintf("[node %v] [DKG] propose set: %v", p.id, proposeSet))
	proposeSetMsg := &protobuf.HartsProposeSet{
		Index: proposeSet,
	}
	proposeSetBytes, err := proto.Marshal(proposeSetMsg)
	if err != nil {
		slog.Error("Marshal harts propose set failed", slog.String("err", err.Error()))
	}
	return proposeSetBytes
}

func (p *Party) GetBandwidth() int {
	bandwidth := 0
	for _, avss := range p.avssInstances {
		bandwidth += avss.bandwidthCounter
	}
	bandwidth += p.bandwidthCounter
	return bandwidth
}
