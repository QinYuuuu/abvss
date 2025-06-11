package harts

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/QinYuuuu/abvss/internal/smvba"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"
	"go.dedis.ch/kyber/v4/pairing"
	"go.dedis.ch/kyber/v4/sign/tbls"
	"google.golang.org/protobuf/proto"
)

func (p *Party) mvbaRun() {
	tc := p.tc
	_ = <-p.mvbaReady
	slog.Info(fmt.Sprintf("[node %v] [DKG] start mvba", p.id))
	//input := p.mvbaInput

	ID := utils.IntToBytes(0)
	sigshare := [][]byte{}
	var buf bytes.Buffer
	buf.Write([]byte("Echo"))
	buf.Write(ID)
	buf.Write(utils.Uint32ToBytes(0))
	h := []byte("TEST")
	// h := input
	buf.Write(h)
	sm := buf.Bytes()
	for i := int64(0); i < 2*tc+1; i++ {
		sigShare, _ := tbls.Sign(pairing.NewSuiteBn256(), p.mvbaSigSK[i], sm)
		sigshare = append(sigshare, sigShare)
	}
	signature, err := tbls.Recover(pairing.NewSuiteBn256(), p.mvbaSigPK, sm, sigshare, int(2*tc+1), int(p.n))
	if err != nil {
		slog.Error(fmt.Sprintf("[node %v] [DKG] mvba recover failed", p.id))
	}
	pids := make([]uint32, 2*tc+1)
	hashes := make([][]byte, 2*tc+1)
	sigs := make([][]byte, 2*tc+1)
	for i := int64(0); i < 2*tc+1; i++ {
		pids[i] = 0
		hashes[i] = []byte("TEST")
		sigs[i] = signature
	}
	value, _ := proto.Marshal(&protobuf.BLockSetValue{
		Pid:  pids,
		Hash: hashes,
	})
	validation, _ := proto.Marshal(&protobuf.BLockSetValidation{
		Sig: sigs,
	})

	ans := smvba.MainProcess(p.mvbaParty, ID, value, validation, smvba.Q)
	p.mvbaOutput <- ans
}
