package erasurecode

// ErasureCode is the interface that wraps the methods that an erasure code should support.
type ErasureCode interface {
	// Encode the given object into a slice of ErasureCodeChunk.
	Encode(input Payload) ([]ErasureCodeChunk, error)

	// Decode a slice of ErasureCodeChunk into the original object. Note that the chunks may be unordered, but
	// all chunks in the slice should be valid.
	Decode(shards []ErasureCodeChunk, v *Payload) error
}

// ErasureCodeChunk VIDChunk is the interface that an erasure coded chunk should implement.
type ErasureCodeChunk interface {
	Size() int // Size returns the size of the object in the emulator.
	GetData() []byte
	Index() int
}

// Payload is the interface that a payload of the protocol should implement.
type Payload interface{}
