package harts

import (
	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"math/big"
	"strconv"

	"github.com/QinYuuuu/abvss/crypto/commit/pedersen"
	"go.dedis.ch/kyber/v3"
)

const (
	Commit string = "harts.Commit"
	Row    string = "harts.Row"
	Column string = "harts.Column"
	Vote   string = "harts.Vote"
	Done   string = "harts.Done"
)

type CommitPayload struct {
	CM      []kyber.Point
	row0_S  []kyber.Point
	row0_Pi []kyber.Point
}

type HAVSSImpl struct {
	n, tc, tr  int64
	id         int64
	dealerID   int64
	instanceID string

	p     *big.Int
	group kyber.Group

	voteSenders []bool
	doneSenders []bool
	voteCounter int64
	doneCounter int64

	sentDone bool
	sentVote bool

	*dealer
	rbc *broadcast.OptRBC

	send    func(*protobuf.HartsMessage)
	receive func() chan *protobuf.HartsMessage
}

type Share struct {
	XIndex      int
	PolyY       []*big.Int     // S(i,y)的单变量多项式
	Commitments *pedersen.Comm // 多项式承诺
	//Proofs      []NIZKProof    // NIZK证明
}

func (vss *HAVSSImpl) Run() {
	vss.rbc.CreateNewSession(strconv.FormatInt(vss.dealerID, 10), vss.dealerID)
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
				vss.handleColumn(msg.FromID)
			case Vote:
				// handle addTrigger
				vss.handleVote(msg.FromID)
			case Done:
				vss.handleDone(msg.FromID)
			default:
				panic("unhandled default case")
			}
		case output := <-vss.rbc.Output(strconv.FormatInt(vss.dealerID, 10)):
			var msg protobuf.HartsMessage
			err := proto.Unmarshal(output, &msg)
			if err != nil {
				slog.Error("proto unmarshal error", err)
			}
			if msg.Type == Commit {
				vss.handleCommit()
			}
		}
	}
}

func (vss *HAVSSImpl) handleRow(msg *protobuf.HartsMessage) {
	var row protobuf.HartsRow
	err := proto.Unmarshal(msg.Value, &row)
	if err != nil {
		slog.Error("proto unmarshal error", err)
	}
	row.GetRows()
}

func (vss *HAVSSImpl) handleCommit() {

}

func (vss *HAVSSImpl) handleColumn(sender int64) {

}

func (vss *HAVSSImpl) handleVote(sender int64) {
	if vss.voteSenders[sender] {
		slog.Error("have receive vote", slog.Any("fromID", sender))
	}
	vss.voteCounter++
	if vss.voteCounter >= vss.n-vss.tc && !vss.sentDone {
		// send done
		vss.sentDone = true
		var i int64
		for i = 0; i < vss.n; i++ {
			msg := &protobuf.HartsMessage{
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
	vss.doneCounter++
	if vss.doneCounter >= vss.tc+1 && !vss.sentDone {
		// send done
		vss.sentDone = true
		var i int64
		for i = 0; i < vss.n; i++ {
			msg := &protobuf.HartsMessage{
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
	}
}
