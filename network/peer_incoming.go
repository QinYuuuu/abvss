package network

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type incomingPeer struct {
	ourID    int
	ourAddr  string
	num      int
	mutex    *sync.RWMutex
	nReadys  int
	transmit chan HttpMessage
	*http.Server
}

func (p *incomingPeer) ready(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	//data := r.URL.Query()
	//id := data.Get("id")
	//addr := data.Get("addr")
	answer := `{"status": "ok"}`
	w.Write([]byte(answer))
	//log.Printf("[node %d] get connect from node %s on %s", p.ourID, id, addr)
	p.mutex.Lock()
	p.nReadys++
	p.mutex.Unlock()
}

func (p *incomingPeer) ReceivePost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET")
		fmt.Fprintln(w, "Hello test!")
	case "POST":
		//fmt.Println("POST")
		r.ParseForm()
		//fromid := r.Form.Get("From")
		lenth, _ := strconv.Atoi(r.Form.Get("Lenth"))
		dataType := r.Form.Get("Type")
		content := r.Form.Get("Content")
		if lenth != len([]byte(content)) {
			fmt.Fprintln(w, "message too long")
		} else {
			//fmt.Fprintf(w, "node %d receive message", p.ourID)
			p.transmit <- HttpMessage{
				DataType: dataType,
				Content:  []byte(content),
			}
		}
		//log.Printf("message: id=%s, lenth=%s, content=%s\n", fromid, lenth, content)
	}
}

func SayHello(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET")
		fmt.Fprintln(w, "Hello test!")
	case "POST":
		fmt.Println("POST")
		fmt.Fprintln(w, []byte("Hello test!"))
	}

}

func NewIncomingPeer(ourID int, num int, ourAddr string, mutex *sync.RWMutex) *incomingPeer {
	server := &http.Server{
		Addr:         ":" + strings.Split(ourAddr, ":")[1],
		Handler:      http.NewServeMux(),
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
	incomingPeer := &incomingPeer{
		ourID:    ourID,
		ourAddr:  ourAddr,
		num:      num,
		mutex:    mutex,
		nReadys:  0,
		transmit: make(chan HttpMessage),
		Server:   server,
	}

	return incomingPeer
}

func (p *incomingPeer) serve() {
	var wg sync.WaitGroup
	p.Handler.(*http.ServeMux).HandleFunc("/ready", p.ready)
	p.Handler.(*http.ServeMux).HandleFunc("/sayhello", SayHello)
	go func() {
		err := p.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe: ", err.Error())
		}
	}()
	wg.Add(1)
	log.Printf("[node %d] handler ready\n", p.ourID)
	go func() {
		for {
			p.mutex.Lock()
			n := p.nReadys
			p.mutex.Unlock()
			if n == p.num-1 {
				break
			} else {
				time.Sleep(3 * time.Second)
			}
			//log.Printf("[node %d] waiting connect\n", p.ourID)
		}
		wg.Done()
	}()
	wg.Wait()
	log.Printf("[node %d] listen ready\n", p.ourID)
}
