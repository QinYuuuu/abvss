package pb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/QinYuuuu/abvss/internal/party"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bls"
	"golang.org/x/crypto/sha3"
	"os"
	"sync"
	"testing"
)

type Address struct {
	Id   int    `json:"Id"`
	Addr string `json:"Addr"`
}

func TestPb(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
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
	sk, pk := party.SigKeyGen(N, 2*F+1)
	epk, evk, esks := party.EncKeyGen(N, F+1)

	var p []*party.HonestParty = make([]*party.HonestParty, N)
	for i := uint32(0); i < N; i++ {
		p[i] = party.NewHonestParty(N, F, i, addressList, pk, sk[i], epk, evk, esks[i])
	}

	for i := uint32(0); i < N; i++ {
		p[i].InitReceiveChannel()
	}

	for i := uint32(0); i < N; i++ {
		p[i].InitSendChannel()
	}

	value := make([]byte, 10)
	validation := make([]byte, 1)
	ID := []byte{1, 2}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, sig, _ := Sender(ctx, p[0], ID, value, validation)

		h := sha3.Sum512(value)
		var buf bytes.Buffer
		buf.Write([]byte("Echo"))
		buf.Write(ID)
		buf.Write(h[:])
		sm := buf.Bytes()
		err := bls.Verify(pairing.NewSuiteBn256(), p[0].SigPK.Commit(), sm, sig)

		fmt.Println(err)
		wg.Done()
	}()

	for i := uint32(0); i < N; i++ {
		go Receiver(ctx, p[i], 0, ID, nil, nil, nil)
	}

	wg.Wait()
}
