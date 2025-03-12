package config

import (
<<<<<<< HEAD
<<<<<<< HEAD
	"encoding/json"
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/elgamal"
	"github.com/QinYuuuu/abvss/internal/party"
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	"abvss/crypto/elgamal"
	"abvss/internal/party"
	"encoding/json"
	"fmt"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/curve25519"
	"os"
)

type ElgamalKeyPair struct {
	SK kyber.Scalar
	PK kyber.Point
}

func ElgamalCurve25519KeyGen(n int, addr string) {
	exist, err := PathExists(addr + "/keypairs/Curve25519")
	if err != nil {
		fmt.Printf("get dir error: %v \n", err)
		return
	}
	if !exist {
		err = os.Mkdir(addr+"/keypairs/Curve25519", 0777)
		if err != nil {
			fmt.Printf("make dir error: %v \n", err)
			return
		}
	}
	suite := curve25519.NewBlakeSHA256Curve25519(true)
	for i := 0; i < n; i++ {
		pk, sk := elgamal.KeyGenCurve25519(suite)

		tmp1 := fmt.Sprintf("%s/keypairs/Curve25519/Node%v.pub", addr, i)
		file1, _ := os.OpenFile(tmp1, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		pkBytes, _ := pk.MarshalBinary()
		_, err := file1.Write(pkBytes)
		if err != nil {
			fmt.Printf("encode error: %v \n", err)
		}
		err = file1.Close()
		if err != nil {
			fmt.Printf("close error: %v \n", err)
		}
		tmp2 := fmt.Sprintf("%s/keypairs/Curve25519/Node%v.priv", addr, i)
		file2, _ := os.OpenFile(tmp2, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		skBytes, _ := sk.MarshalBinary()
		_, err = file2.Write(skBytes)
		if err != nil {
			fmt.Printf("encode error: %v \n", err)
		}
		err = file2.Close()
		if err != nil {
			fmt.Printf("close error: %v \n", err)
		}
	}
}

/*
func ElgamalBLS12381KeyGen(n int, addr string) {
	exist, err := PathExists(addr + "/keypairs/BLS12381")
	if err != nil {
		fmt.Printf("get dir error: %v \n", err)
		return
	}
	if !exist {
		err = os.Mkdir(addr+"/keypairs/BLS12381", 0777)
		if err != nil {
			fmt.Printf("make dir error: %v \n", err)
			return
		}
	}
	suite := bls12381.NewG1()
	for i := 0; i < n; i++ {
		pk, sk := elgamal.KeyGenCurve25519(suite)
		tmp1 := fmt.Sprintf("%s/keypairs/BLS12381/Node%v.pub", addr, i)
		file1, _ := os.OpenFile(tmp1, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		pkBytes, _ := pk.MarshalBinary()
		_, err := file1.Write(pkBytes)
		if err != nil {
			fmt.Printf("encode error: %v \n", err)
		}
		err = file1.Close()
		if err != nil {
			fmt.Printf("close error: %v \n", err)
		}
		tmp2 := fmt.Sprintf("%s/keypairs/BLS12381/Node%v.priv", addr, i)
		file2, _ := os.OpenFile(tmp2, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		skBytes, _ := sk.MarshalBinary()
		_, err = file2.Write(skBytes)
		if err != nil {
			fmt.Printf("encode error: %v \n", err)
		}
		err = file2.Close()
		if err != nil {
			fmt.Printf("close error: %v \n", err)
		}
	}
}*/

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

type SigPK struct {
	BaseBytes   []byte
	CommitBytes [][]byte
}

type EncVK struct {
	I      int
	VBytes []byte
}

func SigKeyGen(n, f int, addr string) {
	path := fmt.Sprintf("%s/keypairs/keys_%v", addr, n)
	exist, err := PathExists(path)
	if err != nil {
		fmt.Printf("get dir error: %v \n", err)
	}
	if !exist {
		err = os.Mkdir(fmt.Sprintf("%s/keypairs/keys_%v", addr, n), 0777)
		if err != nil {
			fmt.Printf("make dir error: %v \n", err)
			return
		}
	}
	sk, pk := party.SigKeyGen(uint32(n), uint32(2*f+1))
	for i, ski := range sk {
		skString := ski.String()
		tmp1 := fmt.Sprintf("%s/ssk_%v.key", path, i)
		file1, _ := os.OpenFile(tmp1, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)

		_, err := file1.Write([]byte(skString))
		if err != nil {
			fmt.Printf("encode error: %v \n", err)
		}
		file1.Close()
	}
	pkbase, commits := pk.Info()
	pkbaseBytes, err := pkbase.MarshalBinary()

	if err != nil {
		fmt.Printf("encode error: %v \n", err)
	}
	commitsBytes := make([][]byte, len(commits))
	for i, commit := range commits {
		commitsBytes[i], _ = commit.MarshalBinary()
	}
	sigPK := SigPK{
		BaseBytes:   pkbaseBytes,
		CommitBytes: commitsBytes,
	}
	tmp2 := fmt.Sprintf("%s/spk.key", path)
	file2, _ := os.OpenFile(tmp2, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	output, _ := json.Marshal(sigPK)
	file2.Write(output)
	file2.Close()

	//_, getpk := LoadSigKey(n, addr)
	/*for i := 0; i < n; i++ {
		fmt.Println(sk[i].String())
		fmt.Println(getsk[i].String())
	}*/
}

func EncKeyGen(n, f int, addr string) {
	path := fmt.Sprintf("%s/keypairs/keys_%v", addr, n)
	exist, err := PathExists(path)
	if err != nil {
		fmt.Printf("get dir error: %v \n", err)
	}
	if !exist {
		err = os.Mkdir(fmt.Sprintf("%s/keypairs/keys_%v", addr, n), 0777)
		if err != nil {
			fmt.Printf("make dir error: %v \n", err)
			return
		}
	}
	epk, evk, esks := party.EncKeyGen(uint32(n), uint32(f+1))
	tmp3 := fmt.Sprintf("%s/epk.key", path)
	file3, _ := os.OpenFile(tmp3, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	epkBytes, _ := epk.MarshalBinary()
	file3.Write(epkBytes)
	file3.Close()
	for i, evki := range evk {
		tmp4 := fmt.Sprintf("%s/evk_%v.key", path, i)
		file4, _ := os.OpenFile(tmp4, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		vbytes, _ := evki.V.MarshalBinary()
		encvk := &EncVK{
			I:      evki.I,
			VBytes: vbytes,
		}
		encoder := json.NewEncoder(file4)
		encoder.Encode(encvk)
		file4.Close()
	}
	for i, esk := range esks {
		tmp4 := fmt.Sprintf("%s/esk_%v.key", path, i)
		file4, _ := os.OpenFile(tmp4, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		eskString := esk.String()
		file4.Write([]byte(eskString))
		file4.Close()
	}
}
