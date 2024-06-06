package main

import (
	"crypto/elliptic"
	"flag"
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/internal/abvss"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/network"
	"github.com/QinYuuuu/abvss/pkg/config"
	"go.dedis.ch/kyber/v3"
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

func main() {

	n := flag.Int("n", 4, "number of nodes in the cluster")
	f := flag.Int("f", 1, "number of faulty nodes to tolerate")
	id := flag.Int("id", 0, "id of this server")
	batchsize := flag.Int("s", 1, "num of secret")
	aws := flag.Int("aws", 0, "1 means run on aws")
	str := flag.String("path", "", "path of node information")

	flag.Parse()
	test(*n, *batchsize, *f, *id, *str, *aws)
	time.Sleep(3 * time.Second)
}

func test(n, batchsize, f, id int, str string, aws int) {
	vnum := n
	c := elliptic.P224()
	param := c.Params()
	p := param.P
	pk1, sk1 := config.LoadElgamalCurve25519(n, str)
	TestVSS(id, n, f, batchsize, vnum, p, pk1, sk1[id], str, aws)
}

func TestVSS(id, n, f, batchsize, vnum int, p *big.Int, pk []kyber.Point, sk kyber.Scalar, addr string, aws int) {
	var portList1, ipList []string
	if aws == 1 {
		portList1, ipList, _ = config.LoadIPList_aws(n, addr)
	} else {
		log.Printf("node %v running local", id)
		portList1, ipList, _ = config.LoadIPList_Local(n, addr)
	}
	node := new(ABVSSNode)
	var mutex sync.Mutex
	abvss_instance, err := abvss.NewVSS(0, id, n, f, batchsize, vnum, p, 1, &mutex)
	osv_instance := osv.NewOSV(n, f, id)
	if err != nil {
		log.Println("NewVSS err:", err)
	}
	abvss_instance.ReceiverInit(sk)
	abvss_instance.VerifyInit()
	peer, err := network.NewPeer(n, id, ipList, portList1)
	/*
		connservice := network.Service{Id: id}
		protobuf.RegisterConnServer(peer.Server, connservice)

		if err != nil {
			log.Println("NewPeer err:", err)
		}*/

	/*
		protobuf.RegisterOSVServer(peer.Server, osvservice)*/
	go peer.Serve()
	peer.Connect()
	/*
		for j := 0; j < n; j++ {
			if j == id {
				continue
			}
			abvssservice.Clients[j] = protobuf.NewABVSSClient(peer.Conns[j])
			osvservice.Clients[j] = protobuf.NewOSVClient(peer.Conns[j])
		}*/
	abvssservice := abvss.NewABVSSService(n, peer.SendChannels, peer.ReceiveChannel)
	abvssservice.ABVSS = abvss_instance
	/*
		protobuf.RegisterABVSSServer(peer.Server, abvssservice)*/
	osvservice := osv.NewOSVService(n)
	osvservice.OSV = osv_instance
	node = &ABVSSNode{ABVSSService: abvssservice, OSVService: osvservice, Peer: peer}

	var wg sync.WaitGroup
	s := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		s[i] = utils.RandomNum(p)
	}
	wg.Wait()
	//fmt.Println(nodes[0].Conns)
	//fmt.Println(nodes[0].Clients)
	start := time.Now()
	if id == 1 {
		node.SecretSharing(pk, s)
	}

	go func() {
		for {
			if node.Received {
				node.BroadcastLCM()
				return
			}
		}
	}()

	go func() {
		for {
			if node.Count == n-f {
				node.OSVService.Init()
				return
			}
		}
	}()

	var flag bool

	go func() {
		for {
			if node.Done() {
				flag = true
				break
			}
		}
	}()

	for flag == false {

	}
	end := time.Now()
	timeusage := end.Sub(start)
	bandwidth := 0
	//fmt.Println("SUCCESS")
	path := "/home/ubuntu/testvss"
	exist, err := config.PathExists(path)
	if err != nil {
		fmt.Printf("get dir error: %v \n", err)
	}
	if !exist {
		err = os.Mkdir("/home/ubuntu/testvss", 0777)
		if err != nil {
			fmt.Printf("make dir error: %v \n", err)
			return
		}
	}

	for i := 0; i < batchsize; i++ {
		bandwidth += int(peer.Bandwidth[i])
	}

	file3, _ := os.OpenFile(fmt.Sprintf("/home/ubuntu/testvss/%v_%v", n, batchsize), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	file3.WriteString(fmt.Sprintf("time\n%v\nband\n%v\n", timeusage, bandwidth))
	time.Sleep(10 * time.Second)
}
