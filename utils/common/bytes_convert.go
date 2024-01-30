package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Int64ToBytes converts an int64 to a byte slice
func Int64ToBytes(num int64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, num)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

// BytesToInt64 converts a byte slice to an int64
func BytesToInt64(b []byte) int64 {
	buf := bytes.NewBuffer(b)
	var num int64
	err := binary.Read(buf, binary.LittleEndian, &num)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	return num
}

// Uint64ToBytes converts an uint64 to a byte slice
func Uint64ToBytes(num uint64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, num)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

// BytesToUint64 converts a byte slice to an uint64
func BytesToUint64(b []byte) uint64 {
	buf := bytes.NewBuffer(b)
	var num uint64
	err := binary.Read(buf, binary.LittleEndian, &num)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	return num
}
