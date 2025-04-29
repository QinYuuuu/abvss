package vaba

import (
	"bytes"
	"fmt"
	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/hasher"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"math/big"
	"strconv"
)

const (
	ECHO  string = "ECHO"
	READY string = "READY"
	SHARE string = "SHARE"
	RECON string = "RECON"
)

type sharePhaseOutput struct {
	hashVector [][]byte
	share      *big.Int
}

// ASKSImpl represents a participant in the protocol
type ASKSImpl struct {
	id, n, t   int64
	dealerID   int64 // Whether this party is the dealer
	instanceID string
	p          *big.Int // Prime number for the field
	// For RBC (Reliable Broadcast)
	rbc *broadcast.OptRBC
	// For Reliable Agreement
	ra *RAImpl

	// ASKS protocol state
	validShares        map[int64]*big.Int // Set T of valid shares for reconstruction
	hasRBCOutput       bool
	hasShareOutput     bool
	hashVecReady       chan bool
	shareReady         chan bool
	sharePhaseReceived *big.Int
	sharePhaseOutput   *sharePhaseOutput
	ReconPhaseOutput   *big.Int // Final output of the protocol
	terminated         bool
	*dealer
	output    chan *sharePhaseOutput
	recOutput chan *big.Int
	send      func(message *protobuf.ASKSMessage)
	receive   func() chan *protobuf.ASKSMessage
}

// NewASKS creates a new party for the ASKS protocol
func NewASKS(id, n, t, dealerID int64, instanceID string, prime *big.Int) *ASKSImpl {
	return &ASKSImpl{
		id:         id,
		n:          n,
		t:          t,
		instanceID: instanceID,
		dealerID:   dealerID,
		p:          prime,
		sharePhaseOutput: &sharePhaseOutput{
			hashVector: nil,
			share:      nil,
		},
		hashVecReady: make(chan bool, 1),
		shareReady:   make(chan bool, 1),
		validShares:  make(map[int64]*big.Int),
		output:       make(chan *sharePhaseOutput, 1),
	}
}

// sendToAll sends a message to all parties
func (p *ASKSImpl) sendToAll(msgType string, content []byte) {
	var i int64
	for i = 0; i < p.n; i++ {
		msg := &protobuf.ASKSMessage{
			Type:   msgType,
			Value:  content,
			FromID: p.id,
			DestID: p.dealerID,
		}
		p.send(msg)
	}
}

func (p *ASKSImpl) Run() {
	p.ra.Run()
	p.rbc.Run()
	p.rbc.CreateNewSession("asks"+strconv.FormatInt(p.dealerID, 10)+"hashVec", p.dealerID)
	go p.messageLoop()
}

func (p *ASKSImpl) Terminated() bool {
	return p.terminated
}

func (p *ASKSImpl) messageLoop() {
	for {
		select {
		case data := <-p.rbc.Output("asks" + strconv.FormatInt(p.dealerID, 10) + "hashVec"):
			slog.Info(fmt.Sprintf("[node %v] receive rbc output", p.id), slog.Any("msg", data))
			var hashVec protobuf.ASKSHashVector
			err := proto.Unmarshal(data, &hashVec)
			if err != nil {
				slog.Error("proto unmarshal", slog.Any("error", err))
			}
			p.sharePhaseOutput.hashVector = hashVec.HashByte
			p.hasRBCOutput = true
			p.hashVecReady <- true
			if p.hasRBCOutput && p.hasShareOutput {
				p.check()
			}
		case msg := <-p.receive():
			slog.Info(fmt.Sprintf("[node %v] receive", p.id), slog.Any("msg", msg))
			p.handleMessage(msg)
		case raOutput := <-p.ra.Output():
			if bytes.Equal(raOutput, []byte("1")) {
				go func() {
					if p.sharePhaseOutput.hashVector == nil {
						slog.Error(fmt.Sprintf("[node %v] output when sharePhase.hashVec not output", p.id))
						<-p.hashVecReady
						slog.Info(fmt.Sprintf("[node %v] sharePhase.hashVec ready", p.id))
					}
					if p.sharePhaseOutput.share == nil {
						slog.Error(fmt.Sprintf("[node %v] output when sharePhase.share not output", p.id))
						<-p.shareReady
						slog.Info(fmt.Sprintf("[node %v] sharePhase.share ready", p.id))
					}
					p.output <- p.sharePhaseOutput
					p.terminated = true
				}()

			}
		}
	}
}

func (p *ASKSImpl) check() {
	pi := p.sharePhaseReceived
	// check hash(i,p(i))
	hash := hasher.MD5Hasher(append([]byte{byte(p.id)}, pi.Bytes()...))
	if bytes.Equal(p.sharePhaseOutput.hashVector[p.id], hash) {
		p.ra.Input([]byte("1"))
		p.sharePhaseOutput.share = pi
	}
}

func (p *ASKSImpl) handleMessage(msg *protobuf.ASKSMessage) {
	switch msg.Type {
	case SHARE:
		p.sharePhaseReceived = new(big.Int).SetBytes(msg.Value)
		p.hasShareOutput = true
		p.shareReady <- true
		if p.hasRBCOutput && p.hasShareOutput {
			p.check()
		}
	case RECON:
		pi := new(big.Int).SetBytes(msg.Value)
		hash := hasher.MD5Hasher(append([]byte{byte(p.id)}, pi.Bytes()...))
		if bytes.Equal(p.sharePhaseOutput.hashVector[p.id], hash) {
			p.validShares[msg.FromID+1] = pi
		}
		if int64(len(p.validShares)) == p.t+1 {
			x := make([]*big.Int, 0)
			y := make([]*big.Int, 0)
			for key, value := range p.validShares {
				x = append(x, new(big.Int).SetInt64(key))
				y = append(y, value)
			}
			interpolation, err := pkg.LagrangeInterpolation(x, y, p.p)
			if err != nil {
				return
			}
			output := interpolation.EvalMod(new(big.Int).SetInt64(0), p.p)
			p.recOutput <- output
		}
	}
}

// Reconstruct initiates the reconstruction phase
func (p *ASKSImpl) Reconstruct() {
	// RECONSTRUCTION PHASE
	// Let (h, sᵢ) be its output from the Sharing phase
	if !p.terminated {
		slog.Error(fmt.Sprintf("[node %v] not terminated, can't reconstruct", p.id))
	}
	si := p.sharePhaseOutput.share
	// If si not null
	if si != nil {
		// Send (RECON, sᵢ) to all
		p.sendToAll(RECON, si.Bytes())
		// Add own share to valid shares
		p.validShares[p.id+1] = si
	}
}

func (p *ASKSImpl) Output() chan *sharePhaseOutput {
	return p.output
}
