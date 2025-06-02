package harts

import (
	"github.com/QinYuuuu/abvss/pkg"
	"go.dedis.ch/kyber/v4"
)

const (
	Commit string = "harts.Commit"
	Row    string = "harts.Row"
	Column string = "harts.Column"
	Vote   string = "harts.Vote"
	Done   string = "harts.Done"
)

type HavssOutput struct {
	_S  []kyber.Point
	_Ci *pkg.PolyKyberImpl
}
