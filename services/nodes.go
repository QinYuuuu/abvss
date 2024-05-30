package services

import "github.com/QinYuuuu/abvss/onesidedvoting"

type Node struct {
	id, n, f int
	onesidedvoting.OSV
}
