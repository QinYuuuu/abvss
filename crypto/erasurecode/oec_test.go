package erasurecode

import (
	"fmt"
	"testing"
	"time"
)

func fillChannel(ch chan int, monitor *ChannelMonitor, n int) {
	for i := 0; i < n; i++ {
		time.Sleep(time.Second) // Simulate some work
		ch <- i
		fmt.Printf("Sent %d to channel\n", i)
		monitor.Increment() // Increment the count and signal the condition variable
	}
}

func TestNewChannelMonitor(t *testing.T) {
	n := 5                  // Desired number of elements in the channel
	ch := make(chan int, n) // Buffered channel with capacity n
	monitor := NewChannelMonitor()

	go fillChannel(ch, monitor, n) // Fill the channel with n elements

	fmt.Println("Waiting for channel to fill...")
	monitor.WaitUntil(n) // Block until the channel has n elements
	fmt.Println("Channel has reached the desired number of elements")
}
