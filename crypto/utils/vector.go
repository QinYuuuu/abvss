package utils

import (
	"errors"
	"fmt"
	"math/big"
)

func arrayOfZeroes(n int) []*big.Int {
	r := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		r[i] = new(big.Int).SetInt64(0)
	}
	return r[:]
}

func DotProduct(v1, v2 []*big.Int) (*big.Int, error) {
	if len(v1) != len(v2) {
		return nil, errors.New("the input length is different")
	}
	dot := zero
	for i := 0; i < len(v1); i++ {
		dot = new(big.Int).Add(dot, new(big.Int).Mul(v1[i], v2[i]))
	}
	return dot, nil
}

func VecPow(v1, v2 []*big.Int, m *big.Int) (*big.Int, error) {
	if len(v1) != len(v2) {
		return nil, errors.New("the input length is different")
	}
	dot := one
	for i := 0; i < len(v1); i++ {
		dot = new(big.Int).Mul(dot, new(big.Int).Exp(v1[i], v2[i], m))
	}
	return dot, nil
}

// VecAdd returns v1 + v2
func VecAdd(v1, v2 []*big.Int) ([]*big.Int, error) {
	if len(v1) != len(v2) {
		return nil, errors.New("the input length is different")
	}
	v := make([]*big.Int, len(v1))
	for i := 0; i < len(v); i++ {
		v[i] = new(big.Int).Add(v1[i], v2[i])
	}
	return v, nil
}

func MatrixMulVector(matrix [][]*big.Int, vector []*big.Int) ([]*big.Int, error) {
	// 检查矩阵和向量的维度是否匹配
	if len(matrix) == 0 || len(matrix[0]) != len(vector) {
		return nil, fmt.Errorf("矩阵列数必须等于向量长度")
	}

	// 初始化结果向量
	result := make([]*big.Int, len(matrix))
	for i := range result {
		result[i] = new(big.Int)
	}

	// 执行矩阵乘向量运算
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < len(vector); j++ {
			// 计算矩阵元素与向量元素的乘积
			temp := new(big.Int).Mul(matrix[i][j], vector[j])
			// 将乘积累加到结果向量的对应位置
			result[i].Add(result[i], temp)
		}
	}

	return result, nil
}

// Helper function to compare two big.Int vectors
func CompareVectors(a, b []*big.Int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i].Cmp(b[i]) != 0 {
			return false
		}
	}

	return true
}
