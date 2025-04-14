package vaba

import (
	"crypto/rand"
	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/hasher"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/big"
	"sync"
	"testing"
	"time"
)

func TestShare(t *testing.T) {
	// 设置测试参数
	n := int64(4)
	threshold := int64(1)
	prime, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

	// 创建随机秘密
	secret, err := rand.Int(rand.Reader, prime)
	require.NoError(t, err, "生成随机秘密失败")

	// 创建通信通道
	channels := make(map[int64]chan *protobuf.ASKSMessage)
	rbcChannels := make(map[int64]chan *protobuf.OptRBCMessage)
	raChannels := make(map[int64]chan *protobuf.RAMessage)
	for i := int64(0); i < n; i++ {
		channels[i] = make(chan *protobuf.ASKSMessage, 100)
		rbcChannels[i] = make(chan *protobuf.OptRBCMessage, 100)
		raChannels[i] = make(chan *protobuf.RAMessage, 100)
	}

	// 创建参与方
	parties := make([]*ASKSImpl, n)
	for i := int64(0); i < n; i++ {
		id := i
		if id == 0 {
			parties[i] = NewASKSDealer(i, n, threshold, secret, prime)
		} else {
			parties[i] = NewASKS(i, n, threshold, 0, prime)
		}

		// 设置通信函数
		parties[i].send = func(id int64) func(message *protobuf.ASKSMessage) {
			return func(message *protobuf.ASKSMessage) {
				channels[message.DestID] <- message
			}
		}(i)

		parties[i].receive = func(id int64) func() chan *protobuf.ASKSMessage {
			return func() chan *protobuf.ASKSMessage {
				return channels[id]
			}
		}(i)

		// 创建输出通道
		parties[i].output = make(chan sharePhaseOutput, 1)

		// 模拟RBC和RA组件
		rbcSend := func(destID int64, msg *protobuf.OptRBCMessage) {
			rbcChannels[msg.DestID] <- msg
		}
		rbcRecv := func() (*protobuf.OptRBCMessage, bool) {
			select {
			case msg := <-rbcChannels[id]:
				return msg, true
			default:
				return nil, false
			}
		}
		parties[i].rbc = broadcast.NewOptRBC(i, n, threshold, rbcSend, rbcRecv)
		parties[i].ra = NewRAImpl(id, n, threshold, "asks0ra")
		parties[i].ra.send = func(msg *protobuf.RAMessage) {
			raChannels[msg.DestID] <- msg
		}
		parties[i].ra.receive = func() chan *protobuf.RAMessage {
			return raChannels[id]
		}
	}

	// 启动所有参与方
	for i := int64(0); i < n; i++ {
		parties[i].Run()
	}

	// 执行Share方法
	parties[0].Share()
	var wg sync.WaitGroup
	wg.Add(int(n))
	done := make(chan bool)
	// 验证结果
	for i := int64(0); i < n; i++ {
		go func(id int64) {
			defer wg.Done()
			select {
			case output := <-parties[id].output:
				// 验证每个参与方都收到了有效的份额
				assert.NotNil(t, output.share, "参与方 %d 没有收到有效份额", id)
				assert.NotNil(t, output.hashVector, "参与方 %d 没有收到哈希向量", id)
				// 验证哈希值是否正确
				hash := hasher.MD5Hasher(append([]byte{byte(id)}, output.share.Bytes()...))
				assert.Equal(t, hash, output.hashVector[id], "参与方 %d 的哈希值不匹配", id)
			}
		}(i)
	}
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// 验证多项式是否正确设置
		assert.NotNil(t, parties[0].poly, "dealer的多项式没有设置")
		// 验证多项式的度数
		assert.Equal(t, int(threshold), parties[0].poly.GetDegree(), "多项式度数不正确")
		// 验证多项式在0点的值是否等于秘密
		secretFromPoly := parties[0].poly.EvalMod(big.NewInt(0), prime)
		// TODO: add lagrange test
		assert.Equal(t, 0, secret.Cmp(secretFromPoly), "多项式在0点的值不等于秘密")
		return
	case <-time.After(30 * time.Second):
		t.Error("测试超时，10秒后强制终止")
		return
	}
}
