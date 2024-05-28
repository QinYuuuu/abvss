package erasurecode

import "sync"

type ChannelMonitor struct {
	mu    sync.Mutex
	count int
	cond  *sync.Cond
}

type OEC struct {
	tuples  chan ReedSolomonChunk
	n       int
	t       int
	monitor *ChannelMonitor
	rscode  *ReedSolomonCode
}

func NewChannelMonitor() *ChannelMonitor {
	m := &ChannelMonitor{
		count: 0,
	}
	m.cond = sync.NewCond(&m.mu)
	return m
}

func (m *ChannelMonitor) Increment() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.count++
	m.cond.Broadcast() // Wake up all goroutines waiting on the condition variable
}

func (m *ChannelMonitor) WaitUntil(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for m.count < n {
		m.cond.Wait() // Wait until the condition variable is signaled
	}
}

func NewOEC(n, t int, monitor *ChannelMonitor) *OEC {
	return &OEC{
		tuples: make(chan ReedSolomonChunk, n),
		n:      n,
		t:      t,
	}
}

func (oec *OEC) Input(t ReedSolomonChunk) {
	oec.tuples <- t
	oec.monitor.Increment()
}

func (oec *OEC) Run() {
	for r := 0; r < oec.t; r++ {
		oec.monitor.WaitUntil(2*oec.t + r + 1)
		shards := make([]ErasureCodeChunk, 2*oec.t+r+1)
		i := 0
		for chunk := range oec.tuples {
			i++
			shards[i] = &chunk
		}
	}
}
