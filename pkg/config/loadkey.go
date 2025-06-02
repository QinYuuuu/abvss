package config

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/curve25519"
	"go.dedis.ch/kyber/v4/pairing"
	"go.dedis.ch/kyber/v4/share"
)

func LoadElgamalCurve25519(n int, addr string) ([]kyber.Point, []kyber.Scalar) {
	exist, err := PathExists(addr + "/keypairs/Curve25519")
	if err != nil {
		log.Fatalf("get dir error: %v \n", err)
	}
	if !exist {
		err = os.Mkdir(addr+"/keypairs/Curve25519", 0777)
		if err != nil {
			log.Fatalf("make dir error: %v \n", err)
		}
	}
	pk := make([]kyber.Point, n)
	sk := make([]kyber.Scalar, n)
	suite := curve25519.NewBlakeSHA256Curve25519(true)
	for i := 0; i < n; i++ {
		tmp1 := fmt.Sprintf("%s/keypairs/Curve25519/Node%v.pub", addr, i)
		pkBytes, _ := os.ReadFile(tmp1)
		pki := suite.Point()
		pki.UnmarshalBinary(pkBytes)
		pk[i] = pki

		tmp2 := fmt.Sprintf("%s/keypairs/Curve25519/Node%v.priv", addr, i)
		skBytes, _ := os.ReadFile(tmp2)
		ski := suite.Scalar()
		ski.UnmarshalBinary(skBytes)
		sk[i] = ski
	}
	return pk, sk
}

func LoadSigKey(n int, addr string) ([]*share.PriShare, *share.PubPoly) {
	suit := pairing.NewSuiteBn256()
	path := fmt.Sprintf("%s/keypairs/keys_%v", addr, n)

	sigPK := new(SigPK)
	tmp2 := fmt.Sprintf("%s/spk.key", path)
	input, _ := os.ReadFile(tmp2)
	json.Unmarshal(input, sigPK)

	base := suit.Point()
	base.UnmarshalBinary(sigPK.BaseBytes)
	//fmt.Printf("base: %v\n", base.String())
	commits := make([]kyber.Point, len(sigPK.CommitBytes))
	for i := range commits {
		commit := suit.Point()
		commit.UnmarshalBinary(sigPK.CommitBytes[i])
		commits[i] = commit
	}
	pk := share.NewPubPoly(suit, base, commits)

	sk := make([]*share.PriShare, n)
	for i := range sk {
		tmp1 := fmt.Sprintf("%s/ssk_%v.key", path, i)
		skBytes, _ := os.ReadFile(tmp1)
		tmp := strings.Split(string(skBytes)[1:len(string(skBytes))-1], ":")
		//fmt.Println(tmp[1])
		I, _ := strconv.Atoi(tmp[0])
		V, _ := hex.DecodeString(tmp[1])
		//fmt.Printf("V %v\n", V)
		sk[i] = &share.PriShare{
			I: I,
			V: suit.Scalar().SetBytes(V),
		}
	}
	return sk, pk
}

func LoadEncKey(n int, addr string) (kyber.Point, []*share.PubShare, []*share.PriShare) {
	suit := pairing.NewSuiteBn256()
	path := fmt.Sprintf("%s/keypairs/keys_%v", addr, n)

	tmp3 := fmt.Sprintf("%s/epk.key", path)
	epkBytes, _ := os.ReadFile(tmp3)
	epk := suit.Point()
	epk.UnmarshalBinary(epkBytes)

	evk := make([]*share.PubShare, n)
	esks := make([]*share.PriShare, n)
	for i := range evk {
		tmp4 := fmt.Sprintf("%s/evk_%v.key", path, i)
		file4, _ := os.OpenFile(tmp4, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		encvk := new(EncVK)
		decoder := json.NewDecoder(file4)
		decoder.Decode(encvk)
		file4.Close()
		v := suit.Point()
		v.UnmarshalBinary(encvk.VBytes)
		evki := &share.PubShare{
			I: int(encvk.I),
			V: v,
		}
		evk[i] = evki
	}
	for i := range esks {
		tmp4 := fmt.Sprintf("%s/esk_%v.key", path, i)
		skBytes, _ := os.ReadFile(tmp4)
		tmp := strings.Split(string(skBytes)[1:len(string(skBytes))-1], ":")
		//fmt.Println(tmp[1])
		I, _ := strconv.Atoi(tmp[0])
		V, _ := hex.DecodeString(tmp[1])
		//fmt.Printf("V %v\n", V)
		esks[i] = &share.PriShare{
			I: I,
			V: suit.Scalar().SetBytes(V),
		}
	}
	return epk, evk, esks
}
