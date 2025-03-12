<<<<<<< HEAD
<<<<<<< HEAD
package utils

import (
	"bytes"
	"encoding/binary"
)

//Uint32ToBytes convert uint32 to bytes
func Uint32ToBytes(n uint32) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, n)
	return bytebuf.Bytes()
}

//BytesToUint32 convert bytes to uint32
func BytesToUint32(byt []byte) uint32 {
	bytebuff := bytes.NewBuffer(byt)
	var data uint32
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}

//BytesToInt convert bytes to int
func BytesToInt(byt []byte) int {
	bytebuff := bytes.NewBuffer(byt)
	var data uint32
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

//IntToBytes convert int to bytes
func IntToBytes(n int) []byte {
	data := uint32(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

//Uint32sToBytes convert uint32s to bytes
func Uint32sToBytes(ns []uint32) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	for _, n := range ns {
		binary.Write(bytebuf, binary.BigEndian, n)
	}
	return bytebuf.Bytes()
}

//BytesToUint32s convert bytes to uint32s
func BytesToUint32s(byt []byte) []uint32 {
	bytebuff := bytes.NewBuffer(byt)
	data := make([]uint32, len(byt)/4)
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}
=======
=======
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
package utils

import (
	"bytes"
	"encoding/binary"
)

//Uint32ToBytes convert uint32 to bytes
func Uint32ToBytes(n uint32) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, n)
	return bytebuf.Bytes()
}

//BytesToUint32 convert bytes to uint32
func BytesToUint32(byt []byte) uint32 {
	bytebuff := bytes.NewBuffer(byt)
	var data uint32
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}

//BytesToInt convert bytes to int
func BytesToInt(byt []byte) int {
	bytebuff := bytes.NewBuffer(byt)
	var data uint32
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

//IntToBytes convert int to bytes
func IntToBytes(n int) []byte {
	data := uint32(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

//Uint32sToBytes convert uint32s to bytes
func Uint32sToBytes(ns []uint32) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	for _, n := range ns {
		binary.Write(bytebuf, binary.BigEndian, n)
	}
	return bytebuf.Bytes()
}

//BytesToUint32s convert bytes to uint32s
func BytesToUint32s(byt []byte) []uint32 {
	bytebuff := bytes.NewBuffer(byt)
	data := make([]uint32, len(byt)/4)
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}
<<<<<<< HEAD
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
=======
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
