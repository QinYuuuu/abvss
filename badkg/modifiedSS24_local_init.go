package badkg

import (
	"math/big"
	"math/rand"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/internal/osv"
	"go.dedis.ch/kyber/v4"
)

func InitLocalMultiACSS(
	nodeNum, degree, dealerID, r, sessionID, batchSize int64,
	p *big.Int,
	group kyber.Group,
	osvList []*osv.Instance,
	rbcList []*broadcast.OptRBC,
) []*ACSSImpl {
	// make acss instance
	s := make([]*big.Int, batchSize)
	randSource := rand.New(rand.NewSource(0))
	for i := range s {
		s[i] = new(big.Int).Rand(randSource, p)
	}
	acssNodes := make([]*ACSSImpl, nodeNum)

	// generate key pairs
	sk := group.Scalar().SetInt64(555)
	pk := group.Point().Mul(sk, nil)
	for i := range acssNodes {
		if i == int(dealerID) {
			acssNodes[0] = NewACSSImplDealer(0, degree, nodeNum, batchSize, r, sessionID, s, p, group)
		} else {
			acssNodes[i] = NewACSSImpl(int64(i), degree, nodeNum, batchSize, r, sessionID, dealerID, p, group)
		}
		acssNodes[i].pkList = make([]kyber.Point, nodeNum)
		for j := int64(0); j < nodeNum; j++ {
			// use same pk
			acssNodes[i].pkList[j] = pk
			acssNodes[i].sk = sk
		}
		acssNodes[i].osvNode = osvList[i]
		acssNodes[i].rbc = rbcList[i]
	}
	return acssNodes
}
