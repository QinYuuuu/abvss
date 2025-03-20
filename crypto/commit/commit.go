package commit

type Commit interface {
	Commit(data any) (commitment Commitment, witness Witness, err error)
	Verify(commitment Commitment, witness Witness, data []byte) bool
	Open(commitment Commitment, witness Witness) ([]byte, error)
}

type Commitment any

type Witness any
