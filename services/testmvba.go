package main

import (
	"bytes"
	"github.com/QinYuuuu/abvss/internal/party"
	"github.com/QinYuuuu/abvss/internal/smvba"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bls"
	"google.golang.org/protobuf/proto"
	"log"
	"sync"
	"time"
)

func TestMVBA(id, n, f int, pk *share.PubPoly, sk *share.PriShare, epk kyber.Point, evk []*share.PubShare, esk *share.PriShare, testNum int, signature [][]byte) {
	_, ipList, portList := GenerateIplist(n)
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
			ans := smvba.MainProcess(p, ID, value, validation, Q)
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

func Q(p *party.HonestParty, ID []byte, value []byte, validation []byte, hashVerifyMap *sync.Map, sigVerifyMap *sync.Map) error {
	var L protobuf.BLockSetValue //L={(j,h)}
	proto.Unmarshal(value, &L)

	var S protobuf.BLockSetValidation
	proto.Unmarshal(validation, &S)

	if len(L.Hash) != 2*int(p.F)+1 || len(L.Pid) != 2*int(p.F)+1 || len(S.Sig) != 2*int(p.F)+1 {
		return errors.New("Q check failed")
	}

	for i := uint32(0); i < 2*p.F+1; i++ {
		h, ok1 := hashVerifyMap.Load(L.Pid[i])
		s, ok2 := sigVerifyMap.Load(L.Pid[i])
		if ok1 && ok2 {
			if bytes.Equal(L.Hash[i], h.([]byte)) && bytes.Equal(S.Sig[i], s.([]byte)) {
				continue
			} else {
				return nil
			}
		}
		var buf bytes.Buffer
		buf.Write([]byte("Echo"))
		buf.Write(ID[:4])
		buf.Write(utils.Uint32ToBytes(L.Pid[i]))
		buf.Write(L.Hash[i])
		sm := buf.Bytes()
		err := bls.Verify(pairing.NewSuiteBn256(), p.SigPK.Commit(), sm, S.Sig[i]) //verify("Echo"||e||j||h)
		if err != nil {
			return err
		}
		hashVerifyMap.Store(L.Pid[i], L.Hash[i])
		sigVerifyMap.Store(L.Pid[i], S.Sig[i])
	}

	return nil
}
