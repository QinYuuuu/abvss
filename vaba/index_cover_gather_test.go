package vaba

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"testing"
)

func TestIndexCoverGatherImpl(t *testing.T) {
	// Test parameters
	n := int64(4)
	threshold := int64(1) // t = 1, so we need n-t = 3 parties
	sessionID := "test"

	// Create nodes
	indexGatherInstances := InitLocalMultiIG(n, threshold, "IG_for_ICG")
	reliableAgreementInstances := make([][]*RAImpl, n)
	for i := int64(0); i < n; i++ {
		reliableAgreementInstances[i] = InitLocalMultiRA(n, threshold, "RA_for_ICG"+strconv.FormatInt(i, 10))
	}
	nodes := InitLocalMultiICG(n, threshold, sessionID, reliableAgreementInstances, indexGatherInstances)
	for i := int64(0); i < n; i++ {
		nodes[i].Run()
	}

	// Create test validations
	// All nodes validate node 0 and 1
	// For successful termination, at least n-t nodes should be validated
	for i := int64(0); i < n-1; i++ {
		nodes[i].ValidateParty(0)
		nodes[i].ValidateParty(1)
		nodes[i].ValidateParty(2)
	}

	// Keep track of how many nodes terminated
	terminated := int64(0)

	var wg sync.WaitGroup
	wg.Add(1)
	outputChan := make(chan []int64, 4)
	for i := int64(0); i < n; i++ {
		go func(index int64) {
			xi := <-nodes[index].Output()
			outputChan <- xi
		}(i)
	}
	for i := n - 1; i < n; i++ {
		nodes[i].ValidateParty(0)
		nodes[i].ValidateParty(1)
		nodes[i].ValidateParty(2)
	}
	go func() {
		for {
			select {
			case <-outputChan:
				terminated++
				if terminated == n {
					slog.Info(fmt.Sprintf("Test finished - %d of %d nodes terminated", terminated, n))
					wg.Done()
				}
			}
		}
	}()
	wg.Wait()
}
