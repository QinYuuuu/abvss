package harts

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
)

const (
	Proposal int64 = iota
	Signature
)

// Party represents a participant in the protocol
type Party struct {
	ID       int
	Secret   *big.Int
	Dealers  map[int]bool
	Prop     map[int]bool
	Sigs     map[int][]byte
	Sharings map[int]Sharing
	Nonces   []*big.Int
	mu       sync.Mutex
}

// Sharing represents a reconstructed sharing
type Sharing struct {
	Party     int
	OriginalS *big.Int
	Shares    map[int]*big.Int
}

// Message represents a message exchanged between parties
type Message struct {
	Type     string
	From     int
	To       int
	Proposal map[int]bool
	Sig      []byte
}

// Network simulates a network between parties
type Network struct {
	Parties []*Party
	Msgs    chan Message
	wg      sync.WaitGroup
}

// NewParty creates a new party with the given ID
func NewParty(id int) *Party {
	return &Party{
		ID:       id,
		Dealers:  make(map[int]bool),
		Prop:     make(map[int]bool),
		Sigs:     make(map[int][]byte),
		Sharings: make(map[int]Sharing),
	}
}

// NewNetwork creates a new network with n parties
func NewNetwork(n int) *Network {
	network := &Network{
		Parties: make([]*Party, n),
		Msgs:    make(chan Message, 1000),
	}
	for i := 0; i < n; i++ {
		network.Parties[i] = NewParty(i)
	}
	return network
}

// Sig simulates creating a signature
func Sig(secret *big.Int, proposal map[int]bool) []byte {
	return []byte(fmt.Sprintf("sig-%v-%v", secret, proposal))
}

// Ver simulates verifying a signature
func Ver(vk int, proposal map[int]bool, sig []byte) bool {
	// In a real implementation, this would verify the signature
	return true
}

// Share simulates sharing a secret via AVSS
func (p *Party) Share(network *Network, n, t int) {
	// Sample a random secret
	max := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	secret, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	p.Secret = secret

	// Simulate AVSS completion by having this party become a dealer for others
	for i := 0; i < n; i++ {
		if i != p.ID {
			msg := Message{
				Type: "avss-complete",
				From: p.ID,
				To:   i,
			}
			network.Msgs <- msg
		}
	}
}

// MVBA simulates Multi-Value Byzantine Agreement
func MVBA(proposal map[int]bool, sigs map[int][]byte, checkValidity func(map[int]bool, map[int][]byte) bool) (map[int]bool, map[int][]byte) {
	// In a real implementation, this would be a Byzantine agreement protocol
	// Here we simply return the inputs as they are valid by assumption
	return proposal, sigs
}

// Rec simulates reconstructing secrets from AVSS
func Rec(dealer int, parties []int) (map[int]*big.Int, *big.Int) {
	// In a real implementation, this would reconstruct the secret from shares
	// Here we just simulate it by creating random shares
	shares := make(map[int]*big.Int)
	for _, p := range parties {
		max := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
		share, _ := rand.Int(rand.Reader, max)
		shares[p] = share
	}

	// Original secret (simulated)
	max := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	original, _ := rand.Int(rand.Reader, max)

	return shares, original
}

// ApplySI applies a pseudo-random funciton to generate nonces
func ApplySI(sharings map[int]Sharing) []*big.Int {
	// In a real implementation, this would apply a pseudo-random function
	// Here we just create random values
	nonces := make([]*big.Int, 0)
	for i := 0; i < 10; i++ { // Generate 10 nonces for example
		max := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
		nonce, _ := rand.Int(rand.Reader, max)
		nonces = append(nonces, nonce)
	}
	return nonces
}

// RunPADKG runs the PADKG algorithm for a party
func (p *Party) RunPADKG(network *Network, n, t int) {
	// Initialize
	p.mu.Lock()
	p.Dealers = make(map[int]bool)
	p.Prop = make(map[int]bool)
	p.Sigs = make(map[int][]byte)
	p.Sharings = make(map[int]Sharing)
	p.Nonces = nil
	p.mu.Unlock()

	// Share a random secret via AVSS
	go p.Share(network, n, t)

	// Process messages
	for msg := range network.Msgs {
		if msg.To != p.ID && msg.To != -1 {
			continue
		}

		switch msg.Type {
		case "avss-complete":
			p.mu.Lock()
			p.Dealers[msg.From] = true

			// If we have n-t dealers, send a proposal
			if len(p.Dealers) == n-t {
				p.Prop = make(map[int]bool)
				for d := range p.Dealers {
					p.Prop[d] = true
				}

				// Send proposal to all parties
				for i := 0; i < n; i++ {
					propCopy := make(map[int]bool)
					for k, v := range p.Prop {
						propCopy[k] = v
					}
					network.Msgs <- Message{
						Type:     "proposal",
						From:     p.ID,
						To:       i,
						Proposal: propCopy,
					}
				}
			}
			p.mu.Unlock()

		case "proposal":
			p.mu.Lock()

			// Check if proposal dealers are a subset of our dealers
			isSubset := true
			for d := range msg.Proposal {
				if !p.Dealers[d] {
					isSubset = false
					break
				}
			}

			if isSubset {
				// Send signature to the proposer
				signature := Sig(p.Secret, msg.Proposal)
				network.Msgs <- Message{
					Type:     "signature",
					From:     p.ID,
					To:       msg.From,
					Proposal: msg.Proposal,
					Sig:      signature,
				}
			}
			p.mu.Unlock()

		case "signature":
			p.mu.Lock()

			// Verify signature
			if len(p.Prop) > 0 && Ver(msg.From, p.Prop, msg.Sig) {
				p.Sigs[msg.From] = msg.Sig

				// If we have t+1 signatures, run MVBA
				if len(p.Sigs) == t+1 {
					// Check validity function (dummy implementation)
					checkValidity := func(prop map[int]bool, sigs map[int][]byte) bool {
						return len(prop) >= n-t && len(sigs) >= t+1
					}

					// Run MVBA
					agreedProp, agreedSigs := MVBA(p.Prop, p.Sigs, checkValidity)

					// If our dealers include all in the agreed proposal
					allIncluded := true
					for d := range agreedProp {
						if !p.Dealers[d] {
							allIncluded = false
							break
						}
					}

					if allIncluded {
						// Reconstruct secrets for all dealers in the proposal
						propSlice := make([]int, 0)
						for d := range agreedProp {
							propSlice = append(propSlice, d)
						}

						for _, dealer := range propSlice {
							shares, originalS := Rec(dealer, propSlice)
							p.Sharings[dealer] = Sharing{
								Party:     dealer,
								OriginalS: originalS,
								Shares:    shares,
							}
						}

						// If we have n-t sharings, generate nonces and terminate
						if len(p.Sharings) == n-t {
							p.Nonces = ApplySI(p.Sharings)
							network.wg.Done() // Signal completion
							fmt.Printf("Party %d completed with %d nonces\n", p.ID, len(p.Nonces))
						}
					}
				}
			}
			p.mu.Unlock()
		}
	}
}

func main() {
	n := 10 // Total number of parties
	t := 3  // Fault tolerance (can handle up to t Byzantine parties)

	network := NewNetwork(n)

	// Set up wait group for completion
	network.wg.Add(n)

	// Start each party's PADKG process
	for i := 0; i < n; i++ {
		go network.Parties[i].RunPADKG(network, n, t)
	}

	// Wait for all parties to complete
	network.wg.Wait()
	close(network.Msgs)

	fmt.Println("PADKG protocol completed successfully")

	// Print some sample results
	for i := 0; i < 3; i++ {
		fmt.Printf("Party %d generated %d nonces\n", i, len(network.Parties[i].Nonces))
	}
}
