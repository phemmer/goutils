package ple

import (
	"encoding/binary"
	"io"
)

func Encode32(in []byte) []byte {
	buf := make([]byte, len(in) + 4)
	Encode32To(in, buf)
	return buf
}
func Encode32To(in []byte, out []byte) {
	binary.BigEndian.PutUint32(out, uint32(len(in)))
	copy(out[4:], in)
}
func Write32(in []byte, out io.Writer) (int, error) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(len(in)))
	n, err := out.Write(buf[:])
	if err != nil {
		return n, err
	}
	n2, err := out.Write(in)
	return n+n2, err
}

func EncodeAll32(ins ...[]byte) []byte {
	l := 0
	for _, bs := range ins {
		l += len(bs) + 4
	}
	out := make([]byte, l)
	l = 0
	for _, bs := range ins {
		size := len(bs)+4
		Encode32To(bs, out[l:l+size])
		l += size
	}
	return out
}

// returns the decoded slice and any remainder.
func Decode32(in []byte) ([]byte, []byte) {
	size := binary.BigEndian.Uint32(in)
	return in[4:size+4], in[size+4:]
}

func DecodeAll32(in []byte) [][]byte {
	var outs [][]byte
	for len(in) > 0 {
		var out []byte
		out, in = Decode32(in)
		outs = append(outs, out)
	}
	return outs
}
