package main

import (
	"fmt"
	osv "github.com/QinYuuuu/abvss/crypto/onesidedvoting"
	"sync"
)

/*
	func main() {
		var c curve.Curve
		c = elliptic.P256()
		param := c.Params()
		generator := curve.NewECPoint(param.Gx, param.Gy)
		batchsize := 3
		//n := new(big.Int).Mul(param.P, big.NewInt(5))
		//degree := 1
		//randstate := rand.New(rand.NewSource(1))
		_, pk, _ := paillier.KeyGen()
		zk := zkp.NewBatchNIZK(c, generator, batchsize, pk)

		fij := make([]*big.Int, batchsize)
		zij := make([]*big.Int, batchsize)
		rij := make([]*big.Int, batchsize)
		Aijx := make([]*big.Int, batchsize)
		Aijy := make([]*big.Int, batchsize)
		for i := 0; i < batchsize; i++ {
			fij[i] = utils.RandomNum(param.P)
			zij[i], rij[i], _ = pk.Encrypt(fij[i])
			Aijx[i], Aijy[i] = c.ScalarMult(generator.X(), generator.Y(), fij[i].Bytes())
		}
		pi, _ := zk.Prove(fij, rij)
		result, _ := zk.Verify(Aijx, Aijy, zij, pi)
		fmt.Println(result)
	}
*/

type node struct {
	in  chan osv.Message
	out chan osv.Message
}

func main() {
	n := 4
	tnum := 1
	msgchans := make([]chan osv.Message, n)
	ch := make(chan osv.Message)
	osvs := make([]*osv.OSV, n)
	var wait sync.WaitGroup
	wait.Add(n)
	for i := 0; i < n; i++ {
		osvs[i] = osv.NewOSV(n, tnum, i)
		msgchans[i] = make(chan osv.Message, 10)
		go func(i int) {
			msgs := osvs[i].Init()
			fmt.Println(msgs)
			for _, msg := range msgs {
				msgchans[msg.Dest()] <- msg
			}
			fmt.Println("Init done")
			wait.Done()
		}(i)
	}
	//msgs := make([][]Message, n)
	wait.Wait()

	wait.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			for msg := range msgchans[i] {
				//fmt.Printf("message from %v to %v %v\n", msg.From(), msg.Dest(), msg.Type())
				replymsgs, err := osvs[i].Recv(msg)
				if err != nil {
					fmt.Printf("recv err: %v", err)
				}
				for _, replymsg := range replymsgs {
					ch <- replymsg
				}
				if osvs[i].Done() {
					wait.Done()
				}
			}
		}(i)
	}
	go func() {
		for replymsg := range ch {
			msgchans[replymsg.Dest()] <- replymsg
			fmt.Printf("out message %v\n", replymsg)
		}
	}()
	wait.Wait()
}
