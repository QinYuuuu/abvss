package main

import (
	"github.com/QinYuuuu/abvss/internal/party"
	"github.com/QinYuuuu/abvss/internal/smvba"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"
	"go.dedis.ch/kyber/v3/share"
	"google.golang.org/protobuf/proto"
	"log"
	"sync"
	"time"
)

func TestMVBA(id, n, f int, pk *share.PubPoly, sk *share.PriShare, epk kyber.Point, evk []*share.PubShare, esk *share.PriShare, testNum int, signature [][]byte) {
	_, ipList, portList := main.GenerateIplist(n)
	/*
		N := uint32(4)
		F := uint32(1)
		sk, pk := party.SigKeyGen(N, 2*F+1)
		epk, evk, esks := party.EncKeyGen(N, F+1)

		var p = make([]*party.HonestParty, N)*/
	p := party.NewHonestParty(uint32(n), uint32(f), uint32(id), ipList, portList, pk, sk, epk, evk, esk)

	p.InitReceiveChannel()
	p.InitSendChannel()

	defer p.Close()
	var wg sync.WaitGroup
	var mu sync.Mutex
	result := make([][][]byte, testNum)
	start := time.Now()
	for k := 0; k < testNum; k++ {
		ID := utils.IntToBytes(0)
		pids := make([]uint32, 2*f+1)
		hashes := make([][]byte, 2*f+1)
		sigs := make([][]byte, 2*f+1)
		for i := 0; i < 2*f+1; i++ {
			pids[i] = 0
			hashes[i] = []byte("TEST")
			sigs[i] = signature[k]
		}
		value, _ := proto.Marshal(&protobuf.BLockSetValue{
			Pid:  pids,
			Hash: hashes,
		})
		validation, _ := proto.Marshal(&protobuf.BLockSetValidation{
			Sig: sigs,
		})

		wg.Add(1)

		go func(k int) {
			ans := smvba.MainProcess(p, ID, value, validation, main.Q)
			mu.Lock()
			result[k] = append(result[k], ans)
			mu.Unlock()
			wg.Done()

		}(k)

	}
	wg.Wait()
	log.Println("MVBA success")
	end := time.Now()
	log.Printf("Test MVBA took %s\n", end.Sub(start))
	/*
		for k := 0; k < testNum; k++ {
			for i := 1; i < n; i++ {
				if result[k][i][0] != result[k][i-1][0] {
					log.Printf("err")
				}
			}
		}*/
}
