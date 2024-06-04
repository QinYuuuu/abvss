package osv

import (
	"context"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"log"
)

type OSVService struct {
	*OSV
	Clients []protobuf.OSVClient
	protobuf.UnimplementedOSVServer
}

func NewOSVService(n int) *OSVService {
	return &OSVService{
		Clients: make([]protobuf.OSVClient, n),
	}
}

func (s *OSVService) Init() {
	msgs := s.OSV.Init()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, msg := range msgs {
		protonewmsg := &protobuf.OSVMsg{
			FromID: int64(msg.FromID),
			DestID: int64(msg.DestID),
			Mtype:  msg.Mtype,
		}
		_, err := s.Clients[msg.DestID].Receive(ctx, protonewmsg)
		if err != nil {
			log.Printf("node %v init err: %v", s.id, err)
		}
	}
}

func (s *OSVService) Receive(ctx context.Context, osvmsg *protobuf.OSVMsg) (*protobuf.AckMsg, error) {
	msg := Message{
		FromID: int(osvmsg.GetFromID()),
		DestID: int(osvmsg.GetDestID()),
		Mtype:  osvmsg.GetMtype(),
	}
	recvmsgs, err := s.Recv(msg)
	if err != nil {
		log.Printf("node %v receive msg err: %v", s.id, err)
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, newmsg := range recvmsgs {
		protonewmsg := &protobuf.OSVMsg{
			FromID: int64(newmsg.FromID),
			DestID: int64(newmsg.DestID),
			Mtype:  newmsg.Mtype,
		}
		_, err := s.Clients[newmsg.DestID].Receive(ctx, protonewmsg)
		if err != nil {
			log.Printf("node %v receive msg err: %v", s.id, err)
		}
	}
	return &protobuf.AckMsg{}, nil
}
