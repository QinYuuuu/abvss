package main

import (
<<<<<<< HEAD
<<<<<<< HEAD
=======
=======
>>>>>>> 19b0d27 (Initial commit)
	"abvss/crypto/utils"
	"abvss/internal/abdkg"
	"abvss/internal/abvss"
	"abvss/internal/osv"
	"abvss/internal/party"
	"abvss/internal/smvba"
	"abvss/network"
	"abvss/pkg/config"
	"abvss/pkg/protobuf"
	utils2 "abvss/pkg/utils"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27 (Initial commit)
	"bytes"
	"crypto/elliptic"
	"errors"
	"flag"
	"fmt"
<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/internal/abdkg"
	"github.com/QinYuuuu/abvss/internal/abvss"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/internal/party"
	"github.com/QinYuuuu/abvss/internal/smvba"
	"github.com/QinYuuuu/abvss/network"
	"github.com/QinYuuuu/abvss/pkg/config"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	utils2 "github.com/QinYuuuu/abvss/pkg/utils"
=======
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27 (Initial commit)
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"google.golang.org/protobuf/proto"
	"log"
	"math/big"
	"os"
	"sync"
	"time"
)

/*
type node struct {
	in  chan osv.Message
	out chan osv.Message
}
*/

func GenerateIplist(n int) ([]string, []string, []string) {
	iplist := make([]string, n)
	for i := 0; i < n; i++ {
		iplist[i] = fmt.Sprintf("127.0.0.1:%d", 8000+i)
	}
	addlist := make([]string, n)
	portlist := make([]string, n)
	for i := 0; i < n; i++ {
		addlist[i] = "127.0.0.1"
		portlist[i] = fmt.Sprintf("%d", 9000+i)
	}
	return iplist, addlist, portlist
}

func main() {

	n := flag.Int("n", 4, "number of nodes in the cluster")
	f := flag.Int("f", 1, "number of faulty nodes to tolerate")
	id := flag.Int("id", 0, "id of this server")
	batchsize := flag.Int("s", 1, "num of secret")
	aws := flag.Int("aws", 0, "1 means run on aws")
	str := flag.String("path", "", "path of node information")

	flag.Parse()
	//go TestDKG(i, n, f, batchsize, vnum, p, pk1, sk1[i], pk, sk[i], epk, evk, esks[i], testNum, signature)
	test(*n, *batchsize, *f, *id, *str, *aws)
	//time.Sleep(300 * time.Second)
}

func test(n, batchsize, f, id int, str string, aws int) {
	vnum := n
	c := elliptic.P224()
	param := c.Params()
	p := param.P
	pk1, sk1 := config.LoadElgamalCurve25519(n, str)
	//sk, pk := party.SigKeyGen(uint32(n), uint32(2*f+1))
	sk, pk := config.LoadSigKey(n, str)
	//epk, evk, esks := party.EncKeyGen(uint32(n), uint32(f+1))
	epk, evk, esks := config.LoadEncKey(n, str)
	testNum := f + 1
	signature := make([][]byte, testNum)

	for k := 0; k < testNum; k++ {
		ID := utils2.IntToBytes(k)
		var sigshare [][]byte
		var buf bytes.Buffer
		buf.Write([]byte("Echo"))
		buf.Write([]byte(ID))
		buf.Write(utils2.Uint32ToBytes(0))
		h := []byte("TEST")
		buf.Write(h)
		sm := buf.Bytes()
		for i := 0; i < 2*f+1; i++ {
			sigShare, _ := tbls.Sign(pairing.NewSuiteBn256(), sk[i], sm)
			sigshare = append(sigshare, sigShare)
		}
		signature[k], _ = tbls.Recover(pairing.NewSuiteBn256(), pk, sm, sigshare, 2*f+1, n)
	}
	//fmt.Println(len(signature[0]))
	TestDKG(id, n, f, batchsize, vnum, p, pk1, sk1[id], pk, sk[id], epk, evk, esks[id], testNum, signature, str, aws)
}

type ABDKGNode struct {
	*abdkg.ABDKGService
	*network.Peer
}

func TestDKG(id, n, f, batchsize, vnum int, pint *big.Int, pk1 []kyber.Point, sk1 kyber.Scalar, pk *share.PubPoly, sk *share.PriShare, epk kyber.Point, evk []*share.PubShare, esk *share.PriShare, testNum int, signature [][]byte, addr string, aws int) {
	var portList1, ipList, portList2 []string
	if aws == 1 {
		portList1, ipList, portList2 = config.LoadIPList_aws(n, addr)
	} else {
		log.Printf("node %v running local", id)
		portList1, ipList, portList2 = config.LoadIPList_Local(n, addr)
	}

	p := party.NewHonestParty(uint32(n), uint32(f), uint32(id), ipList, portList2, pk, sk, epk, evk, esk)
<<<<<<< HEAD
<<<<<<< HEAD
	p.InitReceiveChannel()
	p.InitSendChannel()
=======
=======
>>>>>>> 19b0d27 (Initial commit)
	log.Printf("1")
	p.InitReceiveChannel()
	log.Printf("2")
	p.InitSendChannel()
	log.Printf("3")
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27 (Initial commit)
	node := new(ABDKGNode)
	var err error
	abvss_instance := make([]*abvss.ABVSS, n)
	osv_instance := make([]*osv.OSV, n)
	for i := 0; i < n; i++ {
		var mutex sync.Mutex
		abvss_instance[i], err = abvss.NewVSS(i, id, n, f, batchsize, vnum, pint, 1, &mutex)
		if err != nil {
			log.Println("NewVSS err:", err)
		}
		osv_instance[i] = osv.NewOSV(n, f, id)
	}
	for i := 0; i < n; i++ {
		abvss_instance[i].ReceiverInit(sk1)
		abvss_instance[i].VerifyInit()
	}
	peer, err := network.NewPeer(n, id, ipList, portList1)
<<<<<<< HEAD
<<<<<<< HEAD

	go peer.Serve()
	peer.Connect()
	abdkgservice := abdkg.NewABDKGService(id, n, peer.SendChannels, peer.ReceiveChannel)
=======
=======
>>>>>>> 19b0d27 (Initial commit)
	if err != nil {
		log.Println("NewPeer err:", err)
	}
	log.Printf("1")
	go peer.Serve()
	peer.Connect()
	abdkgservice := abdkg.NewABDKGService(id, n, peer.GetSendChannel(), peer.GetReceiveChannel())
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27 (Initial commit)
	abdkgservice.Vss = abvss_instance
	abdkgservice.Osv = osv_instance

	node = &ABDKGNode{ABDKGService: abdkgservice, Peer: peer}
	var wg sync.WaitGroup

	s := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		s[i] = utils.RandomNum(pint)
	}
	//log.Printf("node %v DKG running", id)
	go abdkgservice.Receive()
	start1 := time.Now()
	go node.SecretSharing(pk1, s)
	for i := 0; i < n; i++ {
		go func(i int) {
			receive := <-node.Vss[i].Received
			if receive {
				//fs, gs := node.Vss[i].GetShares()
				//log.Printf("node %v received shares %v\n%v", id, fs, gs)
				node.BroadcastLCM(i)
				return
			}
		}(i)
	}
	for i := 0; i < 2*f+1; i++ {
		go func(i int) {
			ready := <-node.Vss[i].Ready
			if ready {
				node.Init(i)
				return
			}
		}(i)
	}

	wg.Add(2*f + 1)
	for i := 0; i < 2*f+1; i++ {
		go func(i int) {
			output := <-node.Osv[i].OutPut
			if output {
				wg.Done()
				return
			}
		}(i)
	}
	for i := 0; i < 2*f+1; i++ {
		abdkg.IIPA_Prover1(batchsize)
	}

	wg.Wait()
	end1 := time.Now()
	time.Sleep(10 * time.Second)

	var mu sync.Mutex
	result := make([][][]byte, testNum)
	start2 := time.Now()
	for k := 0; k < testNum; k++ {
		ID := utils2.IntToBytes(k)
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
	end2 := time.Now()
	//fmt.Println("SUCCESS")
	timeusage := end2.Sub(start2) + end1.Sub(start1)
	band1 := 0
	for i := 0; i < n; i++ {
		band1 += int(peer.Bandwidth[i])
	}
	bandwidth := band1 + int(p.Bandwidth)
	fmt.Printf("node %v Time cost: %v %v %v\n", id, timeusage, end1.Sub(start1), end2.Sub(start2))
	fmt.Printf("node %v bandwidth cost: %v\n", id, bandwidth)
	path := "/home/ubuntu/testDKG"
	exist, err := config.PathExists(path)
	if err != nil {
		fmt.Printf("get dir error: %v \n", err)
	}
	if !exist {
		err = os.Mkdir("/home/ubuntu/testDKG", 0777)
		if err != nil {
			fmt.Printf("make dir error: %v \n", err)
			return
		}
	}
	file3, _ := os.OpenFile(fmt.Sprintf("/home/ubuntu/testDKG/node%v_%v_%v", id, n, batchsize), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	file3.WriteString(fmt.Sprintf("time\n%v\nband\n%v\n", timeusage, bandwidth))
	time.Sleep(30 * time.Second)
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
		buf.Write(utils2.Uint32ToBytes(L.Pid[i]))
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
