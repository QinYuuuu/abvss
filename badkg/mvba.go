package badkg

type MVBA struct {
	output chan []int64
}

func (mvba *MVBA) Run() {}

func (mvba *MVBA) Input(set []int64) {}

func (mvba *MVBA) Output() chan []int64 {
	return mvba.output
}
