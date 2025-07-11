package badkg

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"strconv"

	"github.com/QinYuuuu/abvss/crypto/elgamal"
	"github.com/QinYuuuu/abvss/pkg"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v4"
	"google.golang.org/protobuf/proto"
)

type dealer struct {
	fPoly []*pkg.PolyBigIntImpl
	gPoly []*pkg.PolyBigIntImpl
}

func NewACSSImplDealer(id, degree, nodeNum, batchSize, r, sessionID int64, s []*big.Int, p *big.Int, group kyber.Group) *ACSSImpl {
	acss := NewACSSImpl(id, degree, nodeNum, batchSize, r, sessionID, id, p, group)
	acss.dealerInit(s, p, degree, batchSize, r)
	return acss
}

func (vss *ACSSImpl) dealerInit(s []*big.Int, p *big.Int, degree, batchSize, r int64) {
	fPoly := make([]*pkg.PolyBigIntImpl, batchSize)
	gPoly := make([]*pkg.PolyBigIntImpl, r)
	var err error
	for i := range fPoly {
		fPoly[i], err = pkg.NewRandPoly(int(degree), p)
		if err != nil {
			slog.Error("init fPoly", slog.Any("error", err))
		}
		err = fPoly[i].SetCoefficientBig(0, s[i])
		if err != nil {
			slog.Error("set secret", slog.Any("error", err))
		}
	}
	for i := range gPoly {
		gPoly[i], err = pkg.NewRandPoly(int(degree), p)
		if err != nil {
			slog.Error("init gPoly", slog.Any("error", err))
		}
	}
	vss.dealer = &dealer{
		fPoly: fPoly,
		gPoly: gPoly,
	}
	hpoly := make([]*pkg.PolyBigIntImpl, r)
	for i, tmpGPoly := range gPoly {
		hpoly[i], err = pkg.NewPoly(int(degree))
		if err != nil {
			slog.Error("init hPoly", slog.Any("error", err))
		}
		hpoly[i].DeepCopy(tmpGPoly)
		for j, tmpFPoly := range fPoly {
			hpoly[i].AddMul(tmpFPoly, vss.theta[i][j])
		}
	}
	vss.challengePoly = hpoly
}

func (vss *ACSSImpl) Share() {
	if vss.dealer == nil {
		slog.Error("dealer is nil")
		return
	}
	slog.Debug(fmt.Sprintf("[node %v] [acss %v] dealer", vss.id, vss.sessionID))
	var i int64
	aesEncShares := make([]*protobuf.AESEncShare, vss.nodeNum)
	encShares := make([]*protobuf.ElgamalEncShare, vss.nodeNum)
	for i = 0; i < vss.nodeNum; i++ {
		sS24Share := vss.generateShare(i + 1)
		slog.Debug(fmt.Sprintf("[node %v] [acss %v] generate", vss.id, vss.sessionID), slog.Any("share", sS24Share.FShare))
		shareByte, err := proto.Marshal(sS24Share)
		if err != nil {
			slog.Error("proto marshal", slog.Any("error", err))
		}
		key := []byte("exampleKey123456")
		encShare, err := aesEnc(key, shareByte)
		if err != nil {
			slog.Error("aes enc", slog.Any("error", err))
		}
		c1, c2, _ := elgamal.Encrypt(vss.group, vss.pkList[i], key)
		c1Bytes, err := c1.MarshalBinary()
		if err != nil {
			slog.Error("encrypt cipher", slog.Any("error", err), slog.Any("len of aes key", len(key)))
		}
		c2Bytes, err := c2.MarshalBinary()
		if err != nil {
			slog.Error("encrypt cipher", slog.Any("error", err), slog.Any("len of aes key", len(key)))
		}
		aesEncShares[i] = &protobuf.AESEncShare{
			Index:  sS24Share.Index,
			Cipher: encShare,
		}
		encShares[i] = &protobuf.ElgamalEncShare{
			Index: sS24Share.Index,
			C1:    c1Bytes,
			C2:    c2Bytes,
		}
		if i == vss.id {
			vss.myShare = sS24Share
		}
	}
	encMultiShare := &protobuf.AESEncMultiShare{
		AesEncShares:     aesEncShares,
		ElGamalEncShares: encShares,
	}
	encMultiShareBytes, err := proto.Marshal(encMultiShare)
	if err != nil {
		slog.Error("proto marshal", slog.Any("error", err))
		return
	}
	slog.Debug(fmt.Sprintf("[node %v] [acss %v] dealer broadcast encMultiShareBytes ", vss.id, vss.sessionID))
	vss.rbc.StartNewBroadcast(encMultiShareBytes, vss.id, strconv.FormatInt(vss.id, 10)+"0")
	challengePolys := make([]*protobuf.Poly, vss.r)
	for i = 0; i < vss.r; i++ {
		hPoly := make([][]byte, vss.degree+1)
		for j := int64(0); j < vss.degree+1; j++ {
			coeff, err := vss.challengePoly[i].GetCoefficient(int(j))
			if err != nil {
				slog.Error("GetCoefficient", slog.Any("error", err))
			}
			hPoly[j] = coeff.Bytes()
		}
		challengePolys[i] = &protobuf.Poly{
			Coefficient: hPoly,
		}
	}
	challenge := &protobuf.ChallengePoly{
		Polys: challengePolys,
	}
	challengeByte, err := proto.Marshal(challenge)
	if err != nil {
		slog.Error("proto marshal", slog.Any("error", err))
	}
	// RBC challenge poly
	vss.rbc.StartNewBroadcast(challengeByte, vss.id, strconv.FormatInt(vss.id, 10)+"1")
}

func (vss *ACSSImpl) generateShare(index int64) *protobuf.SS24Share {
	if vss.dealer == nil {
		slog.Error("not dealer")
		return nil
	}
	//f := make([]*big.Int, vss.batchSize)
	fByte := make([][]byte, vss.batchSize)
	for i, poly := range vss.dealer.fPoly {
		f := poly.EvalMod(new(big.Int).SetInt64(index), vss.p)
		// slog.Debug(fmt.Sprintf("fShare for %v at %v : %v", index-1, i, f.Int64()))
		fByte[i] = f.Bytes()
	}
	//g := make([]*big.Int, vss.r)
	gByte := make([][]byte, vss.r)
	for i, poly := range vss.dealer.gPoly {
		g := poly.EvalMod(new(big.Int).SetInt64(index), vss.p)
		gByte[i] = g.Bytes()
	}
	return &protobuf.SS24Share{
		InstanceID: strconv.FormatInt(vss.sessionID, 10),
		Index:      index,
		FShare:     fByte,
		GShare:     gByte,
	}
}

func aesEnc(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(cryptoRand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func aesDec(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
