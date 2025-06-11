package smvba

import (
	"bytes"
	"errors"
	"strconv"
	"sync"

	"github.com/QinYuuuu/abvss/internal/party"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"
	"go.dedis.ch/kyber/v4/pairing"
	"go.dedis.ch/kyber/v4/sign/bls"
	"google.golang.org/protobuf/proto"
)

func InitLocalMultiMVBA(n, t uint32) []*party.HonestParty {
	ipList, portList := initLocalIPList(n)
	sk, pk := party.SigKeyGen(n, 2*t+1)
	epk, evk, esks := party.EncKeyGen(n, t+1)
	var p []*party.HonestParty = make([]*party.HonestParty, n)
	for i := uint32(0); i < n; i++ {
		p[i] = party.NewHonestParty(n, t, i, ipList, portList, pk, sk[i], epk, evk, esks[i])
	}

	for i := uint32(0); i < n; i++ {
		p[i].InitReceiveChannel()
	}

	for i := uint32(0); i < n; i++ {
		p[i].InitSendChannel()
	}
	return p
}

func initLocalIPList(n uint32) (ipList, portList []string) {
	ipList = make([]string, n)
	portList = make([]string, n)
	for i := uint32(0); i < n; i++ {
		ipList[i] = "127.0.0.1"
		port := 9000 + i
		portList[i] = strconv.Itoa(int(port))
	}
	return ipList, portList
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
