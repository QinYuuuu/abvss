package network

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

type Server struct {
	ID       int
	N        int
	nodeInfo string
	mutex    *sync.RWMutex
	Outgoing *outgoingPeer
	Incoming *incomingPeer
}

type peer struct {
	Id   int
	Addr string
}

func NewServer(id, n int, path string, mutex *sync.RWMutex) *Server {
	s := &Server{
		ID:       id,
		N:        n,
		nodeInfo: path,
		mutex:    mutex,
	}

	var peers []peer

	content, err := os.ReadFile(s.nodeInfo)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
		return nil
	}
	err = json.Unmarshal(content, &peers)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
		return nil
	}

	if len(peers) != s.N {
		log.Fatalln("Number of specified nodes is not equal to N")
		return nil
	}
	var ourAddr string
	for _, peer := range peers {
		if peer.Id == s.ID {
			ourAddr = peer.Addr
			break
		}
	}
	s.Incoming = NewIncomingPeer(s.ID, s.N, ourAddr, s.mutex)
	s.Outgoing = NewOutgoingPeer(s.ID, s.N, ourAddr, peers, s.mutex)

	return s
}

func (s *Server) Init() error {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		s.Outgoing.init()
		wg.Done()
	}()
	go func() {
		s.Incoming.serve()
		wg.Done()
	}()
	wg.Wait()
	log.Printf("[node %d] ready!\n", s.ID)
	return nil
}

func (s *Server) Register(api string, handler func(http.ResponseWriter, *http.Request)) {
	s.Incoming.Handler.(*http.ServeMux).HandleFunc(api, handler)
}

func (s *Server) RecvChan() chan HttpMessage {
	return s.Incoming.transmit
}

func (s *Server) GetAmount() int {
	return s.Outgoing.totalBytes
}

func (s *Server) GetBandwidth() int {
	return s.Outgoing.maxBytes
}
