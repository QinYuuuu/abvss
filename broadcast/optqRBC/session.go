package optqRBC

import (
	"abvss/pkg/protobuf"
	"log/slog"
	"sync/atomic"
)

type Session struct {
	pid       int64
	sessionID int64
	leader    int64
	n, f      int64

	// protocol state
	stripes      [][]byte
	echoCounter  map[string]*atomic.Int64
	readyCounter map[string]*atomic.Int64

	echoSenders           map[int64]bool
	readySenders          map[int64]bool
	terminateSenders      map[int]struct{}
	addTriggerSenders     map[int]struct{}
	addDisperseSenders    map[int]struct{}
	addReconstructSenders map[int]struct{}
	addDisperseCounter    map[string]int

	// message store
	leaderHash        []byte
	reconstructedHash []byte
	leaderMsg         []byte
	reconstructedMsg  []byte
	committedHash     []byte

	// 标志位
	readySent    bool
	addReadySent bool
	committed    bool

	// 外部接口
	predicate func([]byte) bool
	output    func([]byte)

	// network interface
	send    func(int64, any)
	receive chan *protobuf.OPTRBCMsg
}

func (s *Session) initiateBroadcast(msg []byte) {
	if s.leader != s.pid {
		slog.Error("only leader send propose")
		return
	}
	s.leaderMsg = msg
	proposeMsg := Message{
		SessionID: s.sessionID,
		MsgType:   Propose,
		Payload:   msg,
	}
	for i := range s.n {
		s.send(i, proposeMsg)
	}
}

func (s *Session) messageLoop(send func(int, any), output func([]byte)) {
	for {
		select {
		case msg := <-s.receive:
			switch msg.Mtype {
			case Propose:
				// handle propose
				s.handlePropose(msg.FromID, msg.Payload)
			case Echo:
				// handle echo
				s.handleEcho(msg.FromID, msg.Payload)
			default:
				panic("unhandled default case")
			}

		}
	}
}
