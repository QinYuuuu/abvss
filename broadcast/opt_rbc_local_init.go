package broadcast

import "github.com/QinYuuuu/abvss/pkg/protobuf"

func InitLocalMultiOptRBC(n int64, t int64) []*OptRBC {
	rbcMsgChans := make([]chan *protobuf.OptRBCMessage, n)
	sendRBCMsg := func(destID int64, msg *protobuf.OptRBCMessage) {
		rbcMsgChans[msg.DestID] <- msg
	}
	rbcList := make([]*OptRBC, n)
	for i := int64(0); i < n; i++ {
		rbcMsgChans[i] = make(chan *protobuf.OptRBCMessage, n*100)
		receive := func() (*protobuf.OptRBCMessage, bool) {
			msg, ok := <-rbcMsgChans[i]
			return msg, ok
		}
		rbcList[i] = NewOptRBC(i, n, t, sendRBCMsg, receive)
	}
	return rbcList
}
