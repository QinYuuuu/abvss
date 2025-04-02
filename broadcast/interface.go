package broadcast

type RBCNode interface {
	StartNewBroadcast(msg []byte, sessionID int, leaderID int) error
	Output(sessionID int) ([]byte, error)
}

type RBCMessage struct {
	InstanceID     int64
	FromID, DestID int64
	MsgType        string
	Data           []byte
}
