package smvba

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"

	"go.dedis.ch/kyber/v4/pairing"
	"go.dedis.ch/kyber/v4/sign/tbls"
	"google.golang.org/protobuf/proto"
)

func TestMainProcess(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	N := uint32(4)
	F := uint32(1)
	p := InitLocalMultiMVBA(ctx, N, F)
	testNum := 1
	var wg sync.WaitGroup
	var mu sync.Mutex
	result := make([][][]byte, testNum)

	for k := 0; k < testNum; k++ {
		ID := utils.IntToBytes(k)
		sigshare := [][]byte{}
		var buf bytes.Buffer
		buf.Write([]byte("Echo"))
		buf.Write(ID)
		buf.Write(utils.Uint32ToBytes(0))
		h := []byte("TEST")
		buf.Write(h)
		sm := buf.Bytes()
		for i := uint32(0); i < 2*F+1; i++ {
			sigShare, _ := tbls.Sign(pairing.NewSuiteBn256(), p[i].SigSK, sm)
			sigshare = append(sigshare, sigShare)
		}
		signature, _ := tbls.Recover(pairing.NewSuiteBn256(), p[0].SigPK, sm, sigshare, int(2*F+1), int(N))
		pids := make([]uint32, 2*F+1)
		hashes := make([][]byte, 2*F+1)
		sigs := make([][]byte, 2*F+1)
		for i := uint32(0); i < 2*F+1; i++ {
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

		for i := uint32(0); i < N; i++ {
			wg.Add(1)

			go func(i uint32, k int) {
				ans := MainProcess(p[i], ID, value, validation, Q)
				mu.Lock()
				result[k] = append(result[k], ans)
				mu.Unlock()
				wg.Done()

			}(i, k)

		}

	}
	wg.Wait()
	for k := 0; k < testNum; k++ {
		for i := uint32(1); i < N; i++ {
			if result[k][i][0] != result[k][i-1][0] {
				t.Error()
			}
		}
	}
}
