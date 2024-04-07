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

// Complex128SliceToBytes 将 []complex128 转换为 []byte
func Complex128SliceToBytes(complexSlice []complex128) ([]byte, error) {
	var buf bytes.Buffer
	for _, c := range complexSlice {
		// 写入实部
		if err := binary.Write(&buf, binary.LittleEndian, real(c)); err != nil {
			return nil, err
		}
		// 写入虚部
		if err := binary.Write(&buf, binary.LittleEndian, imag(c)); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// BytesToComplex128Slice 将 []byte 反序列化为 []complex128
func BytesToComplex128Slice(data []byte) ([]complex128, error) {
	if len(data)%16 != 0 {
		return nil, fmt.Errorf("输入的字节长度不是复数的正确倍数")
	}

	var complexSlice []complex128
	reader := bytes.NewReader(data)

	for reader.Len() > 0 {
		var realPart, imagPart float64
		// 读取实部
		if err := binary.Read(reader, binary.LittleEndian, &realPart); err != nil {
			return nil, err
		}
		// 读取虚部
		if err := binary.Read(reader, binary.LittleEndian, &imagPart); err != nil {
			return nil, err
		}
		complexSlice = append(complexSlice, complex(realPart, imagPart))
	}

	return complexSlice, nil
}
