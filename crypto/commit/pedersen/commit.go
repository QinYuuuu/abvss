package pedersen

import (
	"go.dedis.ch/kyber/v3"
)

type Committer struct {
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

func Setup(group kyber.Group) *Committer {
	params := &Committer{
		group: group,
		g:     group.Point().Base(),
		h:     group.Point(),
	}
	return params
}

func (param *Committer) Commit(m []kyber.Scalar) (*Comm, *Pi) {
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

func (param *Committer) Verify(m []kyber.Scalar, c *Comm, pi *Pi) bool {
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
