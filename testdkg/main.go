package main

import (
	"bytes"
	"crypto/elliptic"
	"flag"
	"github.com/QinYuuuu/abvss/pkg/config"
	"github.com/QinYuuuu/abvss/pkg/utils"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/tbls"
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
	str := flag.String("path", "", "path of node information")

	flag.Parse()
	//go TestDKG(i, n, f, batchsize, vnum, p, pk1, sk1[i], pk, sk[i], epk, evk, esks[i], testNum, signature)
	test(*n, *batchsize, *f, *id, *str)
	//time.Sleep(300 * time.Second)
}

func test(n, batchsize, f, id int, str string) {
	vnum := n
	c := elliptic.P224()
	param := c.Params()
	p := param.P
	pk1, sk1 := config.LoadElgamalCurve25519(n, str)
	//sk, pk := party.SigKeyGen(uint32(n), uint32(2*f+1))
	sk, pk := config.LoadSigKey(4, str)
	//epk, evk, esks := party.EncKeyGen(uint32(n), uint32(f+1))
	epk, evk, esks := config.LoadEncKey(4, str)
	testNum := 1
	signature := make([][]byte, testNum)
	for k := 0; k < testNum; k++ {
		ID := utils.IntToBytes(k)
		var sigshare [][]byte
		var buf bytes.Buffer
		buf.Write([]byte("Echo"))
		buf.Write([]byte(ID))
		buf.Write(utils.Uint32ToBytes(0))
		h := []byte("TEST")
		buf.Write(h)
		sm := buf.Bytes()
		for i := 0; i < 2*f+1; i++ {
			sigShare, _ := tbls.Sign(pairing.NewSuiteBn256(), sk[i], sm)
			sigshare = append(sigshare, sigShare)
		}
		signature[k], _ = tbls.Recover(pairing.NewSuiteBn256(), pk, sm, sigshare, 2*f+1, n)
	}
	TestDKG(id, n, f, batchsize, vnum, p, pk1, sk1[id], pk, sk[id], epk, evk, esks[id], testNum, signature)
}
