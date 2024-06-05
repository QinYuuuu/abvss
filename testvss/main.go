package main

import (
	"crypto/elliptic"
	"flag"
	"github.com/QinYuuuu/abvss/pkg/config"
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
	//id := flag.Int("id", 0, "id of this server")
	batchsize := flag.Int("s", 1, "num of secret")
	str := flag.String("path", "", "path of node information")

	flag.Parse()
	//go TestDKG(i, n, f, batchsize, vnum, p, pk1, sk1[i], pk, sk[i], epk, evk, esks[i], testNum, signature)
	for i := 0; i < *n; i++ {
		go test(*n, *batchsize, *f, i, *str)
	}
	time.Sleep(300 * time.Second)
}

func test(n, batchsize, f, id int, str string) {
	vnum := n
	c := elliptic.P224()
	param := c.Params()
	p := param.P
	pk1, sk1 := config.LoadElgamalCurve25519(n, str)
	TestVSS(id, n, f, batchsize, vnum, p, pk1, sk1[id], str)
}
