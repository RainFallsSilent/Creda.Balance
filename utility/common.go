package utility

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

func BytesToHexString(data []byte) string {
	return hex.EncodeToString(data)
}

func BytesReverse(u []byte) []byte {
	for i, j := 0, len(u)-1; i < j; i, j = i+1, j-1 {
		u[i], u[j] = u[j], u[i]
	}
	return u
}

func BytesToUint32(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data)
}

func Uint32ToBytes(data uint32) []byte {
	var r [4]byte
	binary.LittleEndian.PutUint32(r[:], data)
	return r[:]
}

func Uint32ArrayToBytes(data []uint32) []byte {
	var r [4]byte
	binary.LittleEndian.PutUint32(r[:], uint32(len(data)))
	var buffer bytes.Buffer
	buffer.Write(r[:])
	for i := 0; i < len(data); i++ {
		buffer.Write(Uint32ToBytes(data[i]))
	}
	return buffer.Bytes()
}
