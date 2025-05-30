package pedersen

import (
	"go.dedis.ch/kyber/v4"
)

type Param struct {
	group kyber.Group
	g     kyber.Point
	h     kyber.Point
}

type Comm struct {
	c []kyber.Point
}

type Pi struct {
	r []kyber.Scalar
}

func Setup(group kyber.Group) *Param {
	params := &Param{
		group: group,
		g:     group.Point().Base(),
		h:     group.Point(),
	}
	return params
}

func (param *Param) Commit(m []kyber.Scalar) (*Comm, *Pi) {
	comm := make([]kyber.Point, len(m))
	r := make([]kyber.Scalar, len(m))
	for i, val := range m {
		r[i] = param.group.Scalar()
		left := param.group.Point().Mul(r[i], param.h)
		right := param.group.Point().Mul(val, param.g)
		comm[i] = param.group.Point().Add(left, right)
	}
	return &Comm{c: comm}, &Pi{r: r}
}

func (param *Param) Verify(m []kyber.Scalar, c *Comm, pi *Pi) bool {
	for i, val := range m {
		left := param.group.Point().Mul(pi.r[i], param.h)
		right := param.group.Point().Mul(val, param.g)
		tmp := param.group.Point().Add(left, right)
		if !c.c[i].Equal(tmp) {
			return false
		}
	}
	return true
}
