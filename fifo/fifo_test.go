package fifo

import (
	"reflect"
	"strings"
	"testing"
)

type RuneBuffer []rune

const z = rune(0)

// NewRuneBuffer returns a pointer to a RuneBuffer that is allocated based on
// the given var args size.
//
// The same semantics used to allocate a slice with the make built-in function
// are also used[1][2] to allocate the slice returned from NewRuneBuffer.
//
// Namely, the first argument is used as the size of the created slice, and
// the second argument is used as the capacity of the created slice.
// +———————————————————————————————————————————————————————————————————————————
// | [1] Except for the case in which no arguments are given:
// |     • make          — syntax/compiler error
// |     • NewRuneBuffer — allocates the zero value (RuneBuffer{})
// | [2] Except for the case in which invalid (negative) arguments are given:
// |     • make          — syntax/compiler error
// |     • NewRuneBuffer — Uses a value of 0 in place of the invalid value.
func NewRuneBuffer(size int, buff []rune) *RuneBuffer {
	if size < 0 {
		size = 0
	}
	rb := make(RuneBuffer, size)
	copy(rb, buff)
	return &rb
}

func (rb *RuneBuffer) Len() int {
	return len(*rb)
}

func (rb *RuneBuffer) Get(i int) (data Data, ok bool) {
	if rb != nil && 0 <= i && i < rb.Len() {
		data, ok = (*rb)[i], true
	}
	return
}

func (rb *RuneBuffer) Set(i int, data Data) (ok bool) {
	if rb != nil && 0 <= i && i < rb.Len() {
		(*rb)[i], ok = data.(rune)
	}
	return
}

func (rb *RuneBuffer) String() string {
	if rb == nil {
		return "<nil>"
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < rb.Len(); i++ {
		if (*rb)[i] == z {
			sb.WriteRune('.')
		} else {
			sb.WriteRune((*rb)[i])
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func TestFifo_Len(t *testing.T) {
	tests := []struct {
		name   string
		fifo   *State
		expLen int
	}{
		{
			name:   "struct-zero",
			fifo:   &State{},
			expLen: 0,
		},
		{
			name:   "greater-than-cap",
			fifo:   &State{head: 2450, tail: 2500},
			expLen: 50,
		},
		{
			name:   "uint32-overflow",
			fifo:   &State{head: (1 << 32) - 1, tail: 0},
			expLen: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len := tt.fifo.Len(); len != tt.expLen {
				t.Errorf("Fifo.Len() %v != %v", len, tt.expLen)
			}
		})
	}
}

func TestFifo_Deq(t *testing.T) {
	tests := []struct {
		name    string
		fifo    *State
		expData Data
		expOK   bool
	}{
		{
			name:    "struct-nil",
			fifo:    nil,
			expData: nil,
			expOK:   false,
		},
		{
			name:    "struct-zero",
			fifo:    &State{},
			expData: nil,
			expOK:   false,
		},
		{
			name:    "buff-nil",
			fifo:    &State{capa: 10, head: 0, tail: 9, buff: nil},
			expData: nil,
			expOK:   false,
		},
		{
			name:    "buff-empty",
			fifo:    &State{capa: 10, head: 0, tail: 0, buff: NewRuneBuffer(10, []rune{})},
			expData: nil,
			expOK:   false,
		},
		{
			name:    "buff-single",
			fifo:    &State{capa: 10, head: 9, tail: 10, buff: NewRuneBuffer(10, []rune{z, z, z, z, z, z, z, z, z, 'X'})},
			expData: 'X',
			expOK:   true,
		},
		{
			name:    "buff-full",
			fifo:    &State{capa: 10, head: 12, tail: 21, buff: NewRuneBuffer(10, []rune("abcdefghij"))},
			expData: 'c',
			expOK:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fifo
			data, ok := q.Deq()
			if !reflect.DeepEqual(data, tt.expData) {
				t.Errorf("Fifo.Deq(): data: %v != %v", data, tt.expData)
			}
			if ok != tt.expOK {
				t.Errorf("Fifo.Deq(): ok: %v != %v", ok, tt.expOK)
			}
		})
	}
}

func TestFifo_Enq(t *testing.T) {
	tests := []struct {
		name    string
		fifo    *State
		data    Data
		expFifo *State
		expOK   bool
	}{
		{
			name:    "struct-nil",
			fifo:    nil,
			data:    'X',
			expFifo: nil,
			expOK:   false,
		},
		{
			name:    "struct-zero",
			fifo:    &State{},
			data:    'X',
			expFifo: &State{},
			expOK:   false,
		},
		{
			name:    "buff-nil",
			fifo:    &State{capa: 10, mode: DiscardLast, head: 0, tail: 9, buff: nil},
			data:    'X',
			expFifo: &State{capa: 10, mode: DiscardLast, head: 0, tail: 9, buff: nil},
			expOK:   false,
		},
		{
			name:    "buff-empty",
			fifo:    &State{capa: 10, mode: DiscardLast, head: 0, tail: 0, buff: NewRuneBuffer(10, []rune{})},
			data:    'X',
			expFifo: &State{capa: 10, mode: DiscardLast, head: 0, tail: 1, buff: NewRuneBuffer(10, []rune{'X'})},
			expOK:   true,
		},
		{
			name:    "buff-single",
			fifo:    &State{capa: 10, mode: DiscardLast, head: 9, tail: 10, buff: NewRuneBuffer(10, []rune{z, z, z, z, z, z, z, z, z, 'X'})},
			data:    'X',
			expFifo: &State{capa: 10, mode: DiscardLast, head: 9, tail: 11, buff: NewRuneBuffer(10, []rune{'X', z, z, z, z, z, z, z, z, 'X'})},
			expOK:   true,
		},
		{
			name:    "buff-full-last",
			fifo:    &State{capa: 10, mode: DiscardLast, head: 12, tail: 21, buff: NewRuneBuffer(10, []rune("abcdefghij"))},
			data:    'X',
			expFifo: &State{capa: 10, mode: DiscardLast, head: 12, tail: 21, buff: NewRuneBuffer(10, []rune("abcdefghij"))},
			expOK:   false,
		},
		{
			name:    "buff-full-first",
			fifo:    &State{capa: 10, mode: DiscardFirst, head: 12, tail: 21, buff: NewRuneBuffer(10, []rune("abcdefghij"))},
			data:    'X',
			expFifo: &State{capa: 10, mode: DiscardFirst, head: 13, tail: 22, buff: NewRuneBuffer(10, []rune("aXcdefghij"))},
			expOK:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fifo
			ok := q.Enq(tt.data)
			if !reflect.DeepEqual(q, tt.expFifo) {
				t.Errorf("Fifo.Enq(%v): fifo: %s != %s", tt.data, q.String(), tt.expFifo.String())
			}
			if ok != tt.expOK {
				t.Errorf("Fifo.Enq(%v): ok: %v != %v", tt.data, ok, tt.expOK)
			}
		})
	}
}
