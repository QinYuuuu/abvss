package vaba

import (
	"crypto/rand"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/hasher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShare(t *testing.T) {
	// 设置测试参数
	n := int64(4)
	threshold := int64(1)
	prime, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

	// 创建随机秘密
	secret, err := rand.Int(rand.Reader, prime)
	require.NoError(t, err, "生成随机秘密失败")

	rbcList := broadcast.InitLocalMultiOptRBC(n, threshold)
	raList := InitLocalMultiRA(n, threshold, "asks0ra")
	// 创建参与方
	parties := InitLocalMultiASKS(n, threshold, 0, "asks0", secret, prime, rbcList, raList)

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
