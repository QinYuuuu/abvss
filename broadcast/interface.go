package broadcast

type RBCNode interface {
	StartNewBroadcast(msg []byte, sessionID int, leaderID int) error
	Output(sessionID int) ([]byte, error)
}

type RBCMessaege interface {
	GetSessionID() int
	GetLeaderID() int
	GetMessage() []byte
	GetType() string
	GetSenderID() int
	GetReceiverID() int
}
