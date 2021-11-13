package cryptography

import (
	"bytes"
	"math"
)

const partSize = 160
const padDelim = 0x80
const padByte = 0x00

func getPaddedMessageLength(originalLen int) int {
	originalLen += 1
	numParts := int(math.Floor(float64(originalLen) / partSize))
	if numParts%partSize != 0 {
		numParts += 1
	}
	return numParts * partSize
}

func addPadding(data []byte) []byte {
	dlen := len(data)
	msglen := getPaddedMessageLength(len(data)+1) - 1
	padlen := msglen - dlen
	data = append(data, padDelim)
	for padlen > 0 {
		data = append(data, padByte)
		padlen--
	}
	return data
}

func delPadding(data []byte) []byte {
	idx := bytes.LastIndexByte(data, padDelim)
	if idx <= 0 {
		return nil
	}
	return data[0:idx]
}
