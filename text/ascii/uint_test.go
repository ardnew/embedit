package ascii

import (
	"fmt"
	"strings"
	"testing"
)

func TestWriteUint32(t *testing.T) {
	buf := [][]byte{
		nil,
		make([]byte, 0),
		make([]byte, 0, 10),
		make([]byte, 5, 10),
		make([]byte, 10),
		[]byte(""),
		[]byte("abc"),
		[]byte("0"),
		[]byte("1234567890"),
	}
	for _, tt := range []uint32{
		0,
		1,
		12,
		123,
		1234,
		12345,
		123456,
		1234567,
		12345678,
		123456789,
		987654321,
		98765432,
		9876543,
		987654,
		98765,
		9876,
		987,
		98,
		9,
		^uint32(0),
	} {
		for _, tb := range buf {
			s := fmt.Sprintf("%s%d", tb, tt)
			t.Run(fmt.Sprintf("%#v,%d", tb, tt), func(t *testing.T) {
				var sb strings.Builder
				_, _ = sb.Write(tb)
				_, _ = WriteUint32(&sb, tt)
				if sb.String() != s {
					t.Fatalf("Utoa() = %v, want %v", sb.String(), s)
				}
			})
		}
	}
}
