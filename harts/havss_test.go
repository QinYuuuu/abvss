package harts

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"testing"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/commit/pedersen"
	"github.com/QinYuuuu/abvss/crypto/zk/nizk"
	"github.com/QinYuuuu/abvss/pkg"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
)

func Test_HAVSS_Share(t *testing.T) {
	dealerID := int64(0)
	n := int64(4)
	tc := int64(1)
	tr := int64(1)
	// p, _ := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)
	group := edwards25519.NewBlakeSHA256Ed25519()
	nizkIPAParam := nizk.SetupNizkIPA(group, tc+1, group.RandomStream())
	pedersenParam := pedersen.NewVectorParamWithG(group, nizkIPAParam.GetCRS().GetG())

	rbcList := broadcast.InitLocalMultiOptRBC(n, tc)
	havssList := InitLocalMultiHAVSS(n, tc, tr, dealerID, "HAVSS", group, nizkIPAParam, pedersenParam, rbcList)

	// start 4 havss
	for i := int64(0); i < n; i++ {
		havssList[i].Run()
	}
	havssList[0].CommitAndDistribute()
	xlist := make([]int64, 4)
	clist := make([]kyber.Scalar, 4)
	var wg sync.WaitGroup
	wg.Add(int(n))
	for i := int64(0); i < n; i++ {
		xlist[i] = i + 1
		go func(i int64) {
			output := havssList[i].Output()
			if output != nil {
				clist[i] = output._Ci.EvalMod(group.Scalar().Zero())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	originPoly, err := pkg.InterpolationKyber(xlist, clist, group)
	if err != nil {
		t.Errorf("interpolation failed: %v", err)
	}
	secret, err := originPoly.GetCoefficient(0)
	if err != nil {
		t.Errorf("get coefficient failed: %v", err)
	}
	slog.Info(fmt.Sprintf("secret: %v", secret.String()))
}

func Test_HAVSS_Multi_Session(t *testing.T) {
	sessionNum := 2
	dealerList := []int64{0, 1}
	n := int64(4)
	tc := int64(1)
	tr := int64(1)
	group := edwards25519.NewBlakeSHA256Ed25519()
	nizkIPAParam := nizk.SetupNizkIPA(group, tc+1, group.RandomStream())
	pedersenParam := pedersen.NewVectorParamWithG(group, nizkIPAParam.GetCRS().GetG())

	// init 4 HAVSS implementation
	havss := make([][]*HAVSSImpl, sessionNum)
	rbcList := broadcast.InitLocalMultiOptRBC(n, tc)
	for j := 0; j < sessionNum; j++ {
		dealerID := dealerList[j]
		instanceID := "HAVSS_" + strconv.Itoa(j)
		havss[j] = InitLocalMultiHAVSS(n, tc, tr, dealerID, instanceID, group, nizkIPAParam, pedersenParam, rbcList)
	}

	// start 4 havss
	for j := 0; j < sessionNum; j++ {
		for i := int64(0); i < n; i++ {
			havss[j][i].Run()
		}
		havss[j][dealerList[j]].CommitAndDistribute()
	}
	xlist := make([][]int64, sessionNum)
	clist := make([][]kyber.Scalar, sessionNum)
	var wg sync.WaitGroup
	wg.Add(int(n) * sessionNum)
	for j := 0; j < sessionNum; j++ {
		xlist[j] = make([]int64, 4)
		clist[j] = make([]kyber.Scalar, 4)
		for i := int64(0); i < n; i++ {
			xlist[j][i] = i + 1
			go func(i int64) {
				output := havss[j][i].Output()
				if output != nil {
					clist[j][i] = output._Ci.EvalMod(group.Scalar().Zero())
				}
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
	for j := 0; j < sessionNum; j++ {
		originPoly, err := pkg.InterpolationKyber(xlist[j], clist[j], group)
		if err != nil {
			t.Errorf("interpolation failed: %v", err)
		}
		secret, err := originPoly.GetCoefficient(0)
		if err != nil {
			t.Errorf("get coefficient failed: %v", err)
		}
		slog.Info(fmt.Sprintf("secret: %v", secret.String()))
	}

}
