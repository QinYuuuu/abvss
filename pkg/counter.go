package pkg

import "sync/atomic"

type Counter struct {
	count int32
}

func (c *Counter) Add() {
	atomic.AddInt32(&c.count, 1)
}

func (c *Counter) Get() int {
	return int(atomic.LoadInt32(&c.count))
}
