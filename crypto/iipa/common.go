package iipa

import "go.dedis.ch/kyber/v3"

type CRS struct {
	G []kyber.Point
	H kyber.Point
}
