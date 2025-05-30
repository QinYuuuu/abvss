package pkg

import (
	"errors"
	"fmt"
	"log/slog"

	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/util/random"
)

// PolyKyberImpl 使用 kyber.Scalar 实现多项式
type PolyKyberImpl struct {
	group kyber.Group
	coeff []kyber.Scalar // coefficients P(x) = coeff[0] + coeff[1] x + ... + coeff[degree] x^degree ...
}

// NewPolyKyber 返回一个多项式 P(x) = 0，容量为 degree + 1
func NewPolyKyber(degree int, group kyber.Group) (*PolyKyberImpl, error) {
	if degree < 0 {
		return nil, fmt.Errorf(fmt.Sprintf("degree must be non-negative, got %d", degree))
	}

	coeff := make([]kyber.Scalar, degree+1)

	for i := 0; i < len(coeff); i++ {
		zero := group.Scalar().Zero()
		coeff[i] = zero
	}

	return &PolyKyberImpl{
		coeff: coeff,
		group: group,
	}, nil
}

// NewRandPolyKyber 返回一个指定次数的随机多项式
// 系数是 [0, n) 范围内的伪随机数
func NewRandPolyKyber(degree int, suite kyber.Group) (*PolyKyberImpl, error) {
	p, e := NewPolyKyber(degree, suite)
	if e != nil {
		return nil, e
	}

	p.Rand(suite)

	return p, nil
}

// NewConstantPolyKyber 返回一个常数多项式 P(x) = c
func NewConstantPolyKyber(suite kyber.Group, c int64) *PolyKyberImpl {
	zero, err := NewPolyKyber(0, suite)
	if err != nil {
		slog.Error("NewConstantPolyKyber", slog.String("error", err.Error()))
		return nil
	}

	constant := suite.Scalar().SetInt64(c)
	zero.coeff[0] = constant
	return zero
}

// NewOneKyber 创建一个常数多项式 P(x) = 1
func NewOneKyber(suite kyber.Group) *PolyKyberImpl {
	return NewConstantPolyKyber(suite, 1)
}

// NewEmptyKyber 创建一个常数多项式 P(x) = 0
func NewEmptyKyber(suite kyber.Group) *PolyKyberImpl {
	return NewConstantPolyKyber(suite, 0)
}

// GetDegree 返回多项式的次数，忽略前导零
func (poly *PolyKyberImpl) GetDegree() int {
	deg := len(poly.coeff) - 1

	// note: i == 0 is not tested, because even the constant term is zero, we consider it's degree 0
	for i := deg; i > 0; i-- {
		if poly.coeff[i].Equal(poly.coeff[i].Zero()) {
			deg--
		} else {
			break
		}
	}
	return deg
}

// GetLeadingCoefficient 返回最高次项的系数
func (poly *PolyKyberImpl) GetLeadingCoefficient() kyber.Scalar {
	lc := poly.coeff[poly.GetDegree()].Clone()
	return lc
}

// GetCoefficient 返回 coeff[i]
func (poly *PolyKyberImpl) GetCoefficient(i int) (kyber.Scalar, error) {
	if i < 0 || i >= len(poly.coeff) {
		zero := poly.coeff[0].Zero()
		return zero, errors.New("out of boundary")
	}

	return poly.coeff[i], nil
}

// SetCoefficient 将 poly.coeff[i] 设置为 ci
func (poly *PolyKyberImpl) SetCoefficient(i int, suite kyber.Group, ci int64) error {
	if i < 0 || i >= len(poly.coeff) {
		return errors.New("out of boundary")
	}

	newScalar := suite.Scalar().SetInt64(ci)
	poly.coeff[i] = newScalar

	return nil
}

// SetCoefficientScalar 将 poly.coeff[i] 设置为 ci (一个 kyber.Scalar)
func (poly *PolyKyberImpl) SetCoefficientScalar(i int, ci kyber.Scalar) error {
	if i < 0 || i >= len(poly.coeff) {
		return errors.New("out of boundary")
	}

	poly.coeff[i] = ci

	return nil
}

// Reset 将系数设置为零
func (poly *PolyKyberImpl) Reset(suite kyber.Group) {
	for i := 0; i < len(poly.coeff); i++ {
		zero := suite.Scalar().Zero()
		poly.coeff[i] = zero
	}
}

func (poly *PolyKyberImpl) DeepCopy(suite kyber.Group, other *PolyKyberImpl) {
	poly.resetToDegree(suite, other.GetDegree())

	for i := 0; i < other.GetDegree()+1; i++ {
		newScalar := suite.Scalar().Set(other.coeff[i])
		poly.coeff[i] = newScalar
	}
}

// resetToDegree 调整切片大小为 degree
func (poly *PolyKyberImpl) resetToDegree(suite kyber.Group, degree int) {
	// 如果只需要缩小大小
	if degree+1 <= len(poly.coeff) {
		poly.coeff = poly.coeff[:degree+1]
	} else {
		// 如果需要增大切片
		needed := degree + 1 - len(poly.coeff)
		neededPointers := make([]kyber.Scalar, needed)
		for i := 0; i < len(neededPointers); i++ {
			zero := suite.Scalar().Zero()
			neededPointers[i] = zero
		}

		poly.coeff = append(poly.coeff, neededPointers...)
	}

	poly.Reset(suite)
}

func (poly *PolyKyberImpl) Equal(suite kyber.Group, op *PolyKyberImpl) bool {
	if op.GetDegree() != poly.GetDegree() {
		return false
	}

	for i := 0; i <= op.GetDegree(); i++ {
		if !poly.coeff[i].Equal(op.coeff[i]) {
			return false
		}
	}

	return true
}

// IsZero 返回 poly 是否为零多项式
func (poly *PolyKyberImpl) IsZero(suite kyber.Group) bool {
	if poly.GetDegree() != 0 {
		return false
	}

	zero := suite.Scalar().Zero()
	return poly.coeff[0].Equal(zero)
}

// Rand 将多项式的系数设置为 [0, n) 范围内的伪随机数
// WARNING: Rand 确保最高次项系数不为零
func (poly *PolyKyberImpl) Rand(suite kyber.Group) {
	rand := random.New()
	for i := range poly.coeff {
		random := suite.Scalar().Pick(rand)
		poly.coeff[i] = random
	}

	highest := len(poly.coeff) - 1

	for {
		if poly.coeff[highest].Equal(poly.coeff[highest].Zero()) {
			random := suite.Scalar().Pick(rand)
			poly.coeff[highest] = random
		} else {
			break
		}
	}
}

func (poly *PolyKyberImpl) GetCap() int {
	return len(poly.coeff)
}

func (poly *PolyKyberImpl) GrowCapTo(suite kyber.Group, cap int) {
	if cap > len(poly.coeff) {
		needed := cap - len(poly.coeff)
		neededPointers := make([]kyber.Scalar, needed)
		for i := 0; i < len(neededPointers); i++ {
			zero := suite.Scalar().Zero()
			neededPointers[i] = zero
		}
		poly.coeff = append(poly.coeff, neededPointers...)
	}
}

// EvalMod returns poly(x) using Horner's rule. If p != nil, returns poly(x) mod p
func (poly *PolyKyberImpl) EvalMod(x kyber.Scalar) kyber.Scalar {
	result := poly.group.Scalar().Set(poly.coeff[poly.GetDegree()])

	for i := poly.GetDegree(); i >= 1; i-- {
		result.Mul(result, x)
		result.Add(result, poly.coeff[i-1])
	}
	return result
}
