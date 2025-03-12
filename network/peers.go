package network

import (
<<<<<<< HEAD
<<<<<<< HEAD
	"errors"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"github.com/QinYuuuu/abvss/pkg/utils"
	"golang.org/x/net/context"
=======
	"abvss/pkg/protobuf"
	"abvss/pkg/utils"
	"errors"
>>>>>>> 19b0d27 (Initial commit)
=======
	"abvss/pkg/protobuf"
	"abvss/pkg/utils"
	"errors"
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Peer struct {
	n, id          int
<<<<<<< HEAD
<<<<<<< HEAD
	Lis            *net.TCPListener
	Conns          []*net.TCPConn
	ipList         []string
	portList       []string // Node IP Address List
	ReceiveChannel chan *protobuf.Message
	SendChannels   []chan *protobuf.Message
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	lis            *net.TCPListener
	conns          []*net.TCPConn
	iplist         []string
	portlist       []string // Node IP Address list
	receiveChannel chan *protobuf.Message
	sendChannels   []chan *protobuf.Message
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	Closed         bool
	Ready          bool
	Bandwidth      []uint64
}

<<<<<<< HEAD
<<<<<<< HEAD
func NewPeer(n, id int, iplist []string, portList []string) (*Peer, error) {
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
func (p *Peer) GetSendChannel() []chan *protobuf.Message {
	return p.sendChannels
}

func (p *Peer) GetReceiveChannel() chan *protobuf.Message {
	return p.receiveChannel
}

func NewPeer(n, id int, iplist []string, portlist []string) (*Peer, error) {
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	if n != len(iplist) {
		return nil, errors.New("n does not match iplist ")
	}
	return &Peer{
		n:              n,
		id:             id,
<<<<<<< HEAD
<<<<<<< HEAD
		Conns:          make([]*net.TCPConn, n),
		ReceiveChannel: make(chan *protobuf.Message),
		SendChannels:   make([]chan *protobuf.Message, n),
		ipList:         iplist,
		portList:       portList,
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
		conns:          make([]*net.TCPConn, n),
		receiveChannel: make(chan *protobuf.Message),
		sendChannels:   make([]chan *protobuf.Message, n),
		iplist:         iplist,
		portlist:       portlist,
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
		Ready:          false,
		Bandwidth:      make([]uint64, n),
	}, nil
}

<<<<<<< HEAD
<<<<<<< HEAD
func (p *Peer) Serve() {
	addr, err1 := net.ResolveTCPAddr("tcp4", ":"+p.portList[p.id])
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
func (p *Peer) Send(destID int, m *protobuf.Message) {
	p.sendChannels[destID] <- m
}

func (p *Peer) Broadcast(m *protobuf.Message) {
	for _, chn := range p.sendChannels {
		chn <- m
	}
}

func (p *Peer) Serve() {
	addr, err1 := net.ResolveTCPAddr("tcp4", ":"+p.portlist[p.id])
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	if err1 != nil {
		log.Fatalf("node %v create addr err: %v\n", p.id, err1)
	}
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("node %v failed to listen %v", p.id, err)
	}
<<<<<<< HEAD
<<<<<<< HEAD
	//log.Printf("node %d listen on %s", p.id, addr)
=======
	log.Printf("node %d listen on %s", p.id, addr)
>>>>>>> 19b0d27 (Initial commit)
=======
	log.Printf("node %d listen on %s", p.id, addr)
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792

	//Make the receive channel and the handle func
	var conn *net.TCPConn
	var err3 error
<<<<<<< HEAD
<<<<<<< HEAD
	p.ReceiveChannel = make(chan *protobuf.Message, 2048)
=======
	p.receiveChannel = make(chan *protobuf.Message, 2048)
>>>>>>> 19b0d27 (Initial commit)
=======
	p.receiveChannel = make(chan *protobuf.Message, 2048)
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	go func() {
		for {
			//The handle func run forever
			conn, err3 = lis.AcceptTCP()
			if err3 != nil {
				log.Fatalln(err3)
			}
			conn.SetKeepAlive(true)

			//Once connect to a node, make a sub-handle func to handle this connection
			go func(conn *net.TCPConn, channel chan *protobuf.Message) {
				for {
					//Receive bytes
					lengthBuf := make([]byte, 4)
					_, err1 := io.ReadFull(conn, lengthBuf)
					length := utils.BytesToInt(lengthBuf)
					buf := make([]byte, length)
					_, err2 := io.ReadFull(conn, buf)
					if err1 != nil || err2 != nil {
						//log.Fatal("The come in conn has break down", err1, err2)
						break
					}
					//Do Unmarshal
					var m protobuf.Message
					err3 := proto.Unmarshal(buf, &m)
					if err3 != nil {
						//log.Fatalln(err3)
						break
					}
					//log.Printf("node %v receive msg: %v from node %v", p.id, m.GetType(), m.Sender)
<<<<<<< HEAD
<<<<<<< HEAD
					//Push protobuf.Message to receivechannel
					channel <- &m
				}

			}(conn, p.ReceiveChannel)
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
					//Push protobuf.Message to receiveChannel
					channel <- &m
				}

			}(conn, p.receiveChannel)
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
		}
	}()
}

func isBrokenPipeError(err error) bool {
	if err == nil {
		return false
	}
	// 检查错误信息中是否包含 "broken pipe"
	return err.Error() == "write: broken pipe" || err.Error() == "write: connection reset by peer"
}

func (p *Peer) Connect() {
	var wg sync.WaitGroup
	wg.Add(p.n - 1)
<<<<<<< HEAD
<<<<<<< HEAD
	for i := 0; i < len(p.ipList); i++ {
=======
	for i := 0; i < len(p.iplist); i++ {
>>>>>>> 19b0d27 (Initial commit)
=======
	for i := 0; i < len(p.iplist); i++ {
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
		p.Bandwidth[i] = 0
		if i == p.id {
			continue

		}
		go func(i int) {
<<<<<<< HEAD
<<<<<<< HEAD
			addr, err1 := net.ResolveTCPAddr("tcp4", p.ipList[i]+":"+p.portList[i])
			if err1 != nil {
				log.Fatalf("node %v create addr err: %v\n", p.id, err1)
			}
			//log.Printf("node %v try connect to node %v on %v", p.id, i, p.ipList[i])
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
			addr, err1 := net.ResolveTCPAddr("tcp4", p.iplist[i]+":"+p.portlist[i])
			if err1 != nil {
				log.Fatalf("node %v create addr err: %v\n", p.id, err1)
			}
			log.Printf("node %v try connect to node %v on %v", p.id, i, p.iplist[i])
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
			for {
				nConn, err := net.DialTCP("tcp", nil, addr)
				if err != nil {
					//log.Printf("node %v did not connect to node %v: %v", p.id, i, err)
					time.Sleep(3 * time.Second)
					continue
				} else {
					nConn.SetKeepAlive(true)
<<<<<<< HEAD
<<<<<<< HEAD
					p.Conns[i] = nConn
=======
					p.conns[i] = nConn
>>>>>>> 19b0d27 (Initial commit)
=======
					p.conns[i] = nConn
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
					//log.Printf("node %v connect to node %v", p.id, i)
					break
				}

			}
			wg.Done()
		}(i)
	}
	wg.Wait()
<<<<<<< HEAD
<<<<<<< HEAD
	//log.Printf("node %v connect to other nodes", p.id)
	for i := 0; i < len(p.ipList); i++ {
		if i == p.id {
			continue
		}
		conn := p.Conns[i]
		p.SendChannels[i] = make(chan *protobuf.Message, 2048)
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	log.Printf("node %v connect to other nodes", p.id)
	for i := 0; i < len(p.iplist); i++ {
		if i == p.id {
			continue
		}
		conn := p.conns[i]
		p.sendChannels[i] = make(chan *protobuf.Message, 2048)
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
		go func(conn *net.TCPConn, channel chan *protobuf.Message, i int) {
			for {
				//Pop protobuf.Message form sendchannel
				m := <-(channel)
				//Do Marshal
				byt, err1 := proto.Marshal(m)
				if err1 != nil {
					log.Fatalln("do marshal failed", err1)
				}
				//Send bytes
				length := len(byt)
				_, err2 := conn.Write(utils.IntToBytes(length))
				_, err3 := conn.Write(byt)
				p.Bandwidth[i] += uint64(length)
				if err2 != nil || err3 != nil {
					//log.Fatalln("The send channel has break down!", err2, err3)
					break
				}
			}
<<<<<<< HEAD
<<<<<<< HEAD
		}(conn, p.SendChannels[i], i)
	}
	//fmt.Println(p.Conns)
}

func (p *Peer) Close() {
	for i, Conn := range p.Conns {
		if i == p.id {
			continue
		}
		close(p.SendChannels[i])
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
		}(conn, p.sendChannels[i], i)
	}
	//fmt.Println(p.conns)
}

func (p *Peer) Close() {
	for i, Conn := range p.conns {
		if i == p.id {
			continue
		}
		close(p.sendChannels[i])
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
		err := Conn.Close()
		if err != nil {
			log.Printf("node %v close %v", p.id, err)
			continue
		}
		//log.Printf("node %v close success", p.id)
	}
}

type Service struct {
	Id int
}

<<<<<<< HEAD
<<<<<<< HEAD
func (n Service) Receive(ctx context.Context, req *protobuf.TestHelloMessage) (*protobuf.TestResMessage, error) {
=======
func (n Service) Receive(req *protobuf.TestHelloMessage) (*protobuf.TestResMessage, error) {
>>>>>>> 19b0d27 (Initial commit)
=======
func (n Service) Receive(req *protobuf.TestHelloMessage) (*protobuf.TestResMessage, error) {
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	//log.Printf("node %v receive request from node %v: %v", n.Id, req.GetFromID(), req.GetContent())
	return &protobuf.TestResMessage{Content: "have received Hello", FromID: int64(n.Id), DestID: req.GetFromID()}, nil
}
