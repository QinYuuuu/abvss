package reedsolomn

type ReedSolomonChunk struct {
	DataSize int
	Idx      int
	Data     []byte
}

func (c *ReedSolomonChunk) Index() int {
	return c.Idx
}

func (c *ReedSolomonChunk) GetData() []byte {
	return c.Data
}

func (c *ReedSolomonChunk) Size() int {
	return len(c.Data)
}
