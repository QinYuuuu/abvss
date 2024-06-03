package osv

import (
	"context"
	"github.com/QinYuuuu/abvss/protobuf"
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
			FromID: int64(msg.fromID),
			DestID: int64(msg.destID),
			Mtype:  msg.mtype,
		}
		_, err := s.Clients[msg.destID].Receive(ctx, protonewmsg)
		if err != nil {
			log.Printf("node %v init err: %v", s.id, err)
		}
	}
}

func (s *OSVService) Receive(ctx context.Context, osvmsg *protobuf.OSVMsg) (*protobuf.AckMsg, error) {
	msg := Message{
		fromID: int(osvmsg.GetFromID()),
		destID: int(osvmsg.GetDestID()),
		mtype:  osvmsg.GetMtype(),
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
			FromID: int64(newmsg.fromID),
			DestID: int64(newmsg.destID),
			Mtype:  newmsg.mtype,
		}
		_, err := s.Clients[newmsg.destID].Receive(ctx, protonewmsg)
		if err != nil {
			log.Printf("node %v receive msg err: %v", s.id, err)
		}
	}
	return &protobuf.AckMsg{}, nil
}
