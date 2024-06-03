package smvba

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"

	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"google.golang.org/protobuf/proto"
)

type Address struct {
	Id   int    `json:"Id"`
	Addr string `json:"Addr"`
}

func TestMainProcess(t *testing.T) {
	filePath := "../../iplist.json"
	data, _ := os.Open(filePath)
	decoder := json.NewDecoder(data)
	// 解析JSON数据
	var addresses []Address
	_ = decoder.Decode(&addresses)
	fmt.Println(addresses)
	// 提取地址到列表
	var addressList []string
	for _, addr := range addresses {
		addressList = append(addressList, addr.Addr)
	}

	N := uint32(16)
	F := uint32(5)
	sk, pk := party2.SigKeyGen(N, 2*F+1)
	epk, evk, esks := party2.EncKeyGen(N, F+1)

	var p []*party2.HonestParty = make([]*party2.HonestParty, N)
	for i := uint32(0); i < N; i++ {
		p[i] = party2.NewHonestParty(N, F, i, addressList, pk, sk[i], epk, evk, esks[i])
	}

	for i := uint32(0); i < N; i++ {
		p[i].InitReceiveChannel()
	}

	for i := uint32(0); i < N; i++ {
		p[i].InitSendChannel()
	}

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

func Q(p *party2.HonestParty, ID []byte, value []byte, validation []byte, hashVerifyMap *sync.Map, sigVerifyMap *sync.Map) error {
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
