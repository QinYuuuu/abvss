package badkg

import (
	"math/big"
	"sync"
	"testing"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/internal/osv"
	"go.dedis.ch/kyber/v4/group/edwards25519"
)

func TestACSSShare(t *testing.T) {
	// 设置测试参数
	f := int64(1)
	degree := int64(1)
	nodeNum := int64(4)
	batchSize := int64(3)
	r := int64(1)
	sessionID := int64(0)
	p, _ := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)
	group := edwards25519.NewBlakeSHA256Ed25519()

	// make acss instance
	osvList := osv.InitLocalMulti(nodeNum, degree, "acss")
	rbcList := broadcast.InitLocalMultiOptRBC(nodeNum, f)
	acssNodes := InitLocalMultiACSS(nodeNum, degree, 0, r, sessionID, batchSize, p, group, osvList, rbcList)

	// set RBC
	for i := int64(0); i < nodeNum; i++ {
		acssNodes[i].Run()
	}
	// 生成并验证共享
	acssNodes[0].Share()
	// outputShare := make([]*protobuf.SS24Share, nodeNum)
	var wg sync.WaitGroup
	wg.Add(4)
	for i := range nodeNum {
		go func(i int64) {
			<-acssNodes[i].output
			wg.Done()
		}(i)
	}
	wg.Wait()
}
