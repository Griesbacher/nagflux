package factorlog

import (
	"testing"
)

// allDigits converts an integer d to its ascii presentation,
// no matter how big the number is
// i is the deinstation index in buf
func allDigits(buf *[]byte, i, d int) int {
	j := len(*buf)
	// reverse order
	for {
		j--
		(*buf)[j] = digits[d%10]
		d /= 10
		if d == 0 {
			break
		}
	}
	return copy((*buf)[i:], (*buf)[j:])
}

func BenchmarkLangAllDigits(b *testing.B) {
	var buf []byte
	tmp := make([]byte, 64)
	for x := 0; x < b.N; x++ {
		allDigits(&tmp, 0, 3456)
		buf = append(buf, tmp...)
		buf = []byte{}
	}
}

func BenchmarkLangItoa(b *testing.B) {
	var buf []byte
	tmp := make([]byte, 64)
	for x := 0; x < b.N; x++ {
		Itoa(&tmp, 0, 3456)
		buf = append(buf, tmp...)
		buf = []byte{}
	}
}

func BenchmarkLangNDigits(b *testing.B) {
	var buf []byte
	tmp := make([]byte, 64)
	for x := 0; x < b.N; x++ {
		NDigits(&tmp, 4, 0, 3456)
		buf = append(buf, tmp...)
		buf = []byte{}
	}
}
