package curve

import (
	"crypto/elliptic"
	"fmt"
	"testing"

<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/QinYuuuu/abvss/crypto/utils"
=======
	"abvss/crypto/utils"
>>>>>>> 19b0d27 (Initial commit)
=======
	"abvss/crypto/utils"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
	"github.com/stretchr/testify/assert"
)

func TestP256CurveScalarMult(t *testing.T) {
	var c Curve
	c = elliptic.P256()
	param := c.Params()
	generator := NewECPoint(param.Gx, param.Gy)
	k := utils.RandomNum(param.P)
	wantx, wanty := c.ScalarMult(generator.x, generator.y, k.Bytes())
	getx, gety := c.ScalarBaseMult(k.Bytes())
	assert.Equal(t, wantx, getx, "generator mul")
	assert.Equal(t, wanty, gety, "generator mul")
}

func TestP256CurveAdd(t *testing.T) {
	var c Curve
	c = elliptic.P256()
	dotx := zero
	doty := zero
	fmt.Println(c.IsOnCurve(dotx, doty))
}
<<<<<<< HEAD
<<<<<<< HEAD


=======
>>>>>>> 19b0d27 (Initial commit)
=======
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
