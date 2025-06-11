package harts

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/commit/pedersen"
	"github.com/QinYuuuu/abvss/crypto/signature/rsa"
	"github.com/QinYuuuu/abvss/crypto/zk/nizk"
	"github.com/QinYuuuu/abvss/internal/smvba"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	"go.dedis.ch/kyber/v4/share"
)

func InitLocalDKG(n, t int64) {
	tc, tr := t, t
	group := edwards25519.NewBlakeSHA256Ed25519()
	nizkIPAParam := nizk.SetupNizkIPA(group, tc+1, group.RandomStream())
	pedersenParam := pedersen.NewVectorParamWithG(group, nizkIPAParam.GetCRS().GetG())
	siMatrix, err := pkg.GenerateVandermondeKyber(int(n-2*tc), int(n-tc), group)
	if err != nil {
		slog.Error("si matrix generate error", slog.Any("error", err))
	}
	// init 4 dkgMsgChans
	dkgMsgChans := make([]chan *protobuf.HartsMessage, n)
	sendDKGMsg := func(msg *protobuf.HartsMessage) {
		dkgMsgChans[msg.DestID] <- msg
	}
	// init n rbc
	havssList := make([][]*HAVSSImpl, n)
	rbcList := broadcast.InitLocalMultiOptRBC(n, tc)
	mvbaParty := smvba.InitLocalMultiMVBA(uint32(n), uint32(tc))
	mvbaSigSK := make([]*share.PriShare, n)
	for i := int64(0); i < n; i++ {
		dealerID := i
		instanceID := "HAVSS_" + strconv.FormatInt(i, 10)
		havssList[i] = InitLocalMultiHAVSS(n, tc, tr, dealerID, instanceID, group, nizkIPAParam, pedersenParam, rbcList)
		mvbaSigSK[i] = mvbaParty[i].SigSK
	}

	verKeys := make([][]byte, n)
	signKeys := make([][]byte, n)
	for i := int64(0); i < n; i++ {
		signKey, verKey, err := rsa.KeyGenBytes(1024)
		if err != nil {
			slog.Error("generate rsa key")
		}
		verKeys[i] = verKey
		signKeys[i] = signKey
	}
	// init 4 DKG implementation
	dkg := make([]*Party, n)
	for i := int64(0); i < n; i++ {
		dkgMsgChans[i] = make(chan *protobuf.HartsMessage, 10)
		dkgNetwork := DKGNetwork{
			send:    sendDKGMsg,
			receive: func() chan *protobuf.HartsMessage { return dkgMsgChans[i] },
		}
		dkg[i] = NewParty(i, n, tc, tr, group, nizkIPAParam, pedersenParam, dkgNetwork)
		havssImpls := make([]*HAVSSImpl, n)
		for j := int64(0); j < n; j++ {
			havssImpls[j] = havssList[j][i]
		}
		dkg[i].avssInstances = havssImpls
		dkg[i].verKeys = verKeys
		dkg[i].signKey = signKeys[i]
		dkg[i].mvbaParty = mvbaParty[i]
		// dkg[i].mvbaSig = signature
		dkg[i].mvbaSigPK = mvbaParty[0].SigPK
		dkg[i].mvbaSigSK = mvbaSigSK
		dkg[i].superMatrix = siMatrix
		dkg[i].Run()
	}
	var wg sync.WaitGroup
	wg.Add(4)
	for i := int64(0); i < n; i++ {
		go func(i int64) {
			defer wg.Done()
			_ = <-dkg[i].output
			slog.Info(fmt.Sprintf("[node %v] [DKG] finish", i))
		}(i)
	}
	wg.Wait()
}
