package fifo

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type RuneBuffer []rune

const z = rune(0)

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
		// don't write to the buffer unless type assertion succeeds
		var c rune
		if c, ok = data.(rune); ok {
			(*rb)[i] = c
		}
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

func TestOverflowMode_String(t *testing.T) {
	type in struct {
		OverflowMode
	}
	type out struct {
		string
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"discard-last",
			in{DiscardLast},
			out{"OverflowMode(DiscardLast)"},
		}, {
			"discard-first",
			in{DiscardFirst},
			out{"OverflowMode(DiscardFirst)"},
		}, {
			"mode-invalid",
			in{OverflowMode(^byte(0))},
			out{"<OverflowMode(0xFF):invalid>"},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			if string := tt.in.OverflowMode.String(); string != tt.out.string {
				t.Errorf("OverflowMode.String(): string got:%q != want:%q", string, tt.out.string)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type in struct {
		Buffer
		int
		OverflowMode
	}
	type out struct {
		*State
		cap int
		rem int
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"buff-nil",
			in{nil, 0, DiscardLast},
			out{
				State: &State{capa: 0, head: 0, tail: 0, mode: DiscardLast, buff: nil},
				cap:   0,
				rem:   0,
			},
		}, {
			"buff-zero",
			in{&RuneBuffer{}, 0, DiscardLast},
			out{
				State: &State{capa: 0, head: 0, tail: 0, mode: DiscardLast, buff: &RuneBuffer{}},
				cap:   0,
				rem:   0,
			},
		}, {
			"greater-than-phy",
			in{&RuneBuffer{'a', 'b', 'c', 'd', 'e'}, 10, DiscardFirst},
			out{
				State: &State{capa: 5, head: 0, tail: 0, mode: DiscardFirst, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}},
				cap:   5,
				rem:   4,
			},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			state := New(tt.in.Buffer, tt.in.int, tt.in.OverflowMode)
			if !reflect.DeepEqual(state, tt.out.State) {
				t.Errorf("New(%s, %d, %s): got:%s != want:%s",
					tt.in.Buffer.String(), tt.in.int, tt.in.OverflowMode.String(),
					state.String(), tt.out.State.String())
			}
			if cap := state.Cap(); cap != tt.out.cap {
				t.Errorf("Cap(): int: got:%d != want:%d", cap, tt.out.cap)
			}
			if rem := state.Rem(); rem != tt.out.rem {
				t.Errorf("Rem(): int: got:%d != want:%d", rem, tt.out.rem)
			}
		})
	}
}

func TestState_String(t *testing.T) {
	type in struct {
		*State
	}
	type out struct {
		string
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"basic",
			in{&State{capa: 5, mode: DiscardLast, head: 3, tail: 6, buff: &RuneBuffer{z, 'a', 'b', 'c', z}}},
			out{"{mode:OverflowMode(DiscardLast), capa:5, head:3, tail:6[1], size:3, buff:[.abc.]}"},
		}, {
			"state-nil",
			in{nil},
			out{"<nil>"},
		}, {
			"buff-nil",
			in{&State{capa: 2, mode: DiscardLast, head: 5, tail: 100, buff: nil}},
			out{"{mode:OverflowMode(DiscardLast), capa:2, head:5[1], tail:100[0], size:95, buff:<nil>}"},
		}, {
			"buff-zero",
			in{&State{buff: &RuneBuffer{}}},
			out{"{mode:OverflowMode(DiscardLast), capa:0, head:0, tail:0, size:0, buff:[]}"},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			if string := tt.in.State.String(); string != tt.out.string {
				t.Errorf("State.String(): string: got:%q != want:%q", string, tt.out.string)
			}
		})
	}
}

func TestState_Len(t *testing.T) {
	type in struct {
		*State
	}
	type out struct {
		int
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-zero",
			in{&State{}},
			out{0},
		}, {
			"greater-than-cap",
			in{&State{head: 2450, tail: 2500}},
			out{50},
		}, {
			"uint32-overflow",
			in{&State{head: (1 << 32) - 1, tail: 0}},
			out{1},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			if int := tt.in.State.Len(); int != tt.out.int {
				t.Errorf("Fifo.Len(): int: got:%v != want:%v", int, tt.out.int)
			}
		})
	}
}

func TestState_Deq(t *testing.T) {
	type in struct {
		*State
	}
	type out struct {
		Data
		bool
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-nil",
			in{nil},
			out{nil, false},
		}, {
			"state-zero",
			in{&State{}},
			out{nil, false},
		}, {
			"buff-nil",
			in{&State{capa: 10, head: 0, tail: 9, buff: nil}},
			out{nil, false},
		}, {
			"buff-empty",
			in{&State{capa: 10, head: 0, tail: 0, buff: &RuneBuffer{}}},
			out{nil, false},
		}, {
			"buff-single",
			in{&State{capa: 10, head: 9, tail: 10, buff: &RuneBuffer{z, z, z, z, z, z, z, z, z, 'X'}}},
			out{'X', true},
		}, {
			"buff-full",
			in{&State{capa: 10, head: 12, tail: 21, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}}},
			out{'c', true},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			Data, bool := tt.in.State.Deq()
			if !reflect.DeepEqual(Data, tt.out.Data) {
				t.Errorf("Fifo.Deq(): Data: got:%+v != want:%+v", Data, tt.out.Data)
			}
			if bool != tt.out.bool {
				t.Errorf("Fifo.Deq(): bool: got:%v != want:%v", bool, tt.out.bool)
			}
		})
	}
}

func TestState_Enq(t *testing.T) {
	type in struct {
		*State
		Data
	}
	type out struct {
		*State
		bool
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-nil",
			in{nil, 'X'},
			out{nil, false},
		}, {
			"state-zero",
			in{&State{}, 'X'},
			out{&State{}, false},
		}, {
			"buff-nil",
			in{&State{capa: 10, mode: DiscardLast, head: 0, tail: 9, buff: nil}, 'X'},
			out{&State{capa: 10, mode: DiscardLast, head: 0, tail: 9, buff: nil}, false},
		}, {
			"buff-empty",
			in{&State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}}, 'X'},
			out{&State{capa: 5, mode: DiscardLast, head: 0, tail: 1, buff: &RuneBuffer{'X', z, z, z, z}}, true},
		}, {
			"buff-single",
			in{&State{capa: 10, mode: DiscardLast, head: 9, tail: 10, buff: &RuneBuffer{z, z, z, z, z, z, z, z, z, 'X'}}, 'X'},
			out{&State{capa: 10, mode: DiscardLast, head: 9, tail: 11, buff: &RuneBuffer{'X', z, z, z, z, z, z, z, z, 'X'}}, true},
		}, {
			"discard-last",
			in{&State{capa: 10, mode: DiscardLast, head: 12, tail: 21, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}}, 'X'},
			out{&State{capa: 10, mode: DiscardLast, head: 12, tail: 21, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}}, false},
		}, {
			"discard-first",
			in{&State{capa: 10, mode: DiscardFirst, head: 12, tail: 21, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}}, 'X'},
			out{&State{capa: 10, mode: DiscardFirst, head: 13, tail: 22, buff: &RuneBuffer{'a', 'X', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}}, true},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			bool := tt.in.State.Enq(tt.in.Data)
			if !reflect.DeepEqual(tt.in.State, tt.out.State) {
				t.Errorf("Fifo.Enq(%+v): State: got:%s != want:%s", tt.in.Data, tt.in.State.String(), tt.out.State.String())
			}
			if bool != tt.out.bool {
				t.Errorf("Fifo.Enq(%+v): bool: got:%v != want:%v", tt.in.Data, bool, tt.out.bool)
			}
		})
	}
}

func TestState_Read(t *testing.T) {
	type in struct {
		*State
		Data []Data
	}
	type out struct {
		int
		bool
		Data []Data
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-nil",
			in{nil, make([]Data, 10)},
			out{0, true, []Data{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}},
		}, {
			"buff-nil",
			in{&State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: nil}, make([]Data, 10)},
			out{0, true, []Data{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}},
		}, {
			"data-nil",
			in{&State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}}, nil},
			out{0, true, nil},
		}, {
			"data-empty",
			in{&State{capa: 5, mode: DiscardLast, head: 0, tail: 1, buff: &RuneBuffer{'X', z, z, z, z}}, []Data{}},
			out{0, true, []Data{}},
		}, {
			"buff-empty",
			in{&State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}}, make([]Data, 10)},
			out{0, true, []Data{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}},
		}, {
			"buff-short",
			in{&State{capa: 5, mode: DiscardLast, head: 8, tail: 10, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}, make([]Data, 10)},
			out{2, false, []Data{'d', 'e', nil, nil, nil, nil, nil, nil, nil, nil}},
		}, {
			"data-short",
			in{&State{capa: 5, mode: DiscardLast, head: 6, tail: 10, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}, make([]Data, 2)},
			out{2, false, []Data{'b', 'c'}},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			int, error := tt.in.State.Read(tt.in.Data)
			if (error != nil) != tt.out.bool {
				t.Errorf("State.Read(): error: %v (want-error: %v)", error, tt.out.bool)
			}
			if int != tt.out.int {
				t.Errorf("State.Read(): int: got:%d != want:%d", int, tt.out.int)
			}
			if !reflect.DeepEqual(tt.in.Data, tt.out.Data) {
				t.Errorf("State.Read(): Data: got:%+v != want:%+v", tt.in.Data, tt.out.Data)
			}
		})
	}
}

func TestState_Write(t *testing.T) {
	type in struct {
		*State
		Data []Data
	}
	type out struct {
		int
		bool
		*State
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-nil",
			in{
				nil,
				[]Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			},
			out{0, true, nil},
		}, {
			"buff-nil",
			in{
				&State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: nil},
				[]Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			},
			out{0, true, &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: nil}},
		}, {
			"data-nil",
			in{
				&State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
				nil,
			},
			out{0, true, &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}}},
		}, {
			"data-empty",
			in{
				&State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
				[]Data{},
			},
			out{0, true, &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}}},
		}, {
			"zero-cap",
			in{
				&State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
				[]Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			},
			out{0, true, &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}}},
		}, {
			"mode-invalid",
			in{
				&State{capa: 5, mode: OverflowMode(^byte(0)), head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
				[]Data{'a'},
			},
			out{0, true, &State{capa: 5, mode: OverflowMode(^byte(0)), head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}}},
		}, {
			"discard-last",
			in{
				&State{capa: 5, mode: DiscardLast, head: 2, tail: 6, buff: &RuneBuffer{'d', z, 'a', 'b', 'c'}},
				[]Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			},
			out{0, true, &State{capa: 5, mode: DiscardLast, head: 2, tail: 6, buff: &RuneBuffer{'d', z, 'a', 'b', 'c'}}},
		}, {
			"short-write",
			in{
				&State{capa: 10, mode: DiscardLast, head: 0, tail: 7, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', z, z, z}},
				[]Data{'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'},
			},
			out{2, false, &State{capa: 10, mode: DiscardLast, head: 0, tail: 9, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'q', 'r', z}}},
		}, {
			"discard-first",
			in{
				&State{capa: 5, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'d', z, 'a', 'b', 'c'}},
				[]Data{'q', 'r', 's'},
			},
			out{3, false, &State{capa: 5, mode: DiscardFirst, head: 5, tail: 9, buff: &RuneBuffer{'d', 'q', 'r', 's', 'c'}}},
		}, {
			"discard-cap",
			in{
				&State{capa: 5, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'d', z, 'a', 'b', 'c'}},
				[]Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			},
			out{10, false, &State{capa: 5, mode: DiscardFirst, head: 0, tail: 4, buff: &RuneBuffer{'g', 'h', 'i', 'j', 'c'}}},
		}, {
			"last-zero-cap",
			in{
				&State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
				make([]Data, 10),
			},
			out{0, true, &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}}},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			int, error := tt.in.State.Write(tt.in.Data)
			if (error != nil) != tt.out.bool {
				t.Errorf("State.Write(): error: %v (want-error: %v)", error, tt.out.bool)
			}
			if int != tt.out.int {
				t.Errorf("State.Write(): int: got:%d != want:%d", int, tt.out.int)
			}
			if !reflect.DeepEqual(tt.in.State, tt.out.State) {
				t.Errorf("State.Write(): State: got:%s != want:%s", tt.in.State.String(), tt.out.State.String())
			}
		})
	}
}

func TestState_Get(t *testing.T) {
	type in struct {
		*State
		int []int
	}
	type out struct {
		Data []Data
		bool []bool
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-nil",
			in{nil, []int{0}},
			out{[]Data{nil}, []bool{false}},
		}, {
			"<nil>FIFO:",
			in{
				&State{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: nil},
				[]int{0, 1, -1},
			},
			out{
				[]Data{nil, nil, nil},
				[]bool{false, false, false},
			},
		}, {
			"[0]FIFO:",
			in{
				&State{capa: 10, mode: DiscardFirst, head: 6, tail: 6, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
				[]int{0, 1, -1},
			},
			out{
				[]Data{nil, nil, nil},
				[]bool{false, false, false},
			},
		}, {
			"[4]FIFO:",
			in{
				&State{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
				[]int{0, 3, -4, -1, 4, -5},
			},
			out{
				[]Data{'c', 'f', 'c', 'f', nil, nil},
				[]bool{true, true, true, true, false, false},
			},
		},
	} {
		// find the maximum index n common to all 3 parameter slices
		nI, nD, nB := len(tt.in.int), len(tt.out.Data), len(tt.out.bool)
		n := nI
		if nD < n {
			n = nD
		}
		if nB < n {
			n = nB
		}
		for i := 0; i < n; i++ {
			string := strconv.FormatInt(int64(tt.in.int[i]), 10)
			if tt.in.int[i] >= 0 {
				string = "+" + string
			}
			t.Run(tt.string+string, func(t *testing.T) {
				Data, bool := tt.in.State.Get(tt.in.int[i])
				if !reflect.DeepEqual(Data, tt.out.Data[i]) {
					t.Errorf("State.Get(): Data: got:%+v != want:%+v", Data, tt.out.Data[i])
				}
				if bool != tt.out.bool[i] {
					t.Errorf("State.Get(): bool: got:%v != want:%v", bool, tt.out.bool[i])
				}
			})
		}
	}
}

func TestState_Set(t *testing.T) {
	type in struct {
		*State
		int  []int
		Data []Data
	}
	type out struct {
		bool  []bool
		State []*State
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-nil",
			in{nil, []int{0}, []Data{nil}},
			out{[]bool{false}, nil},
		}, {
			"<nil>FIFO:",
			in{
				&State{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: nil},
				[]int{0, 1, -1},
				[]Data{nil, nil, nil},
			},
			out{
				[]bool{false, false, false},
				[]*State{
					{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: nil},
					{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: nil},
					{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: nil},
				},
			},
		}, {
			"[0]FIFO:",
			in{
				&State{capa: 10, mode: DiscardFirst, head: 6, tail: 6, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
				[]int{0, 1, -1},
				[]Data{nil, nil, nil},
			},
			out{
				[]bool{false, false, false},
				[]*State{
					{capa: 10, mode: DiscardFirst, head: 6, tail: 6, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
					{capa: 10, mode: DiscardFirst, head: 6, tail: 6, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
					{capa: 10, mode: DiscardFirst, head: 6, tail: 6, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
				},
			},
		}, {
			"[4]FIFO:",
			in{
				&State{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
				[]int{0, 3, 1, 4},
				[]Data{'x', 'y', nil, nil},
			},
			out{
				[]bool{true, true, false, false},
				[]*State{
					{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'a', 'b', 'x', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
					{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'a', 'b', 'x', 'd', 'e', 'y', 'g', 'h', 'i', 'j'}},
					{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'a', 'b', 'x', 'd', 'e', 'y', 'g', 'h', 'i', 'j'}},
					{capa: 10, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'a', 'b', 'x', 'd', 'e', 'y', 'g', 'h', 'i', 'j'}},
				},
			},
		}, {
			"[5]FIFO:",
			in{
				&State{capa: 10, mode: DiscardFirst, head: 2, tail: 7, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
				[]int{-1, -5, -2, -6},
				[]Data{'x', 'y', nil, nil},
			},
			out{
				[]bool{true, true, false, false},
				[]*State{
					{capa: 10, mode: DiscardFirst, head: 2, tail: 7, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'x', 'h', 'i', 'j'}},
					{capa: 10, mode: DiscardFirst, head: 2, tail: 7, buff: &RuneBuffer{'a', 'b', 'y', 'd', 'e', 'f', 'x', 'h', 'i', 'j'}},
					{capa: 10, mode: DiscardFirst, head: 2, tail: 7, buff: &RuneBuffer{'a', 'b', 'y', 'd', 'e', 'f', 'x', 'h', 'i', 'j'}},
					{capa: 10, mode: DiscardFirst, head: 2, tail: 7, buff: &RuneBuffer{'a', 'b', 'y', 'd', 'e', 'f', 'x', 'h', 'i', 'j'}},
				},
			},
		},
	} {
		// find the maximum index n common to all 3 parameter slices
		nI, nD, nB, nS := len(tt.in.int), len(tt.in.Data), len(tt.out.bool), len(tt.out.State)
		n := nI
		if nD < n {
			n = nD
		}
		if nB < n {
			n = nB
		}
		if nS < n {
			n = nS
		}
		for i := 0; i < n; i++ {
			string := strconv.FormatInt(int64(tt.in.int[i]), 10)
			if tt.in.int[i] >= 0 {
				string = "+" + string
			}
			t.Run(tt.string+string, func(t *testing.T) {
				bool := tt.in.State.Set(tt.in.int[i], tt.in.Data[i])
				if bool != tt.out.bool[i] {
					t.Errorf("State.Set(%d): bool: got:%v != want:%v", tt.in.int[i], bool, tt.out.bool[i])
				}
				if !reflect.DeepEqual(tt.in.State, tt.out.State[i]) {
					t.Errorf("State.Set(%v): State: got:%s != want:%s", tt.in.int[0:i+1], tt.in.State.String(), tt.out.State[i].String())
				}
			})
		}
	}
}

func TestState_First(t *testing.T) {
	type in struct {
		*State
	}
	type out struct {
		Data
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-nil",
			in{nil},
			out{nil},
		}, {
			"buff-nil",
			in{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 3, buff: nil}},
			out{nil},
		}, {
			"buff-zero",
			in{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 3, buff: &RuneBuffer{}}},
			out{nil},
		}, {
			"normal",
			in{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 3, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}},
			out{'b'},
		}, {
			"overflow",
			in{&State{capa: 5, mode: DiscardFirst, head: 19, tail: 22, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}},
			out{'e'},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			if Data := tt.in.State.First(); !reflect.DeepEqual(Data, tt.out.Data) {
				t.Errorf("State.First(): Data: got:%+v != want:%+v", Data, tt.out.Data)
			}
		})
	}
}

func TestState_Last(t *testing.T) {
	type in struct {
		*State
	}
	type out struct {
		Data
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-nil",
			in{nil},
			out{nil},
		}, {
			"buff-nil",
			in{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 3, buff: nil}},
			out{nil},
		}, {
			"buff-zero",
			in{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 3, buff: &RuneBuffer{}}},
			out{nil},
		}, {
			"normal",
			in{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 3, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}},
			out{'c'},
		}, {
			"overflow",
			in{&State{capa: 5, mode: DiscardFirst, head: 19, tail: 22, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}},
			out{'b'},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			if Data := tt.in.State.Last(); !reflect.DeepEqual(Data, tt.out.Data) {
				t.Errorf("State.Last(): Data: got:%+v != want:%+v", Data, tt.out.Data)
			}
		})
	}
}

func TestState_Remove(t *testing.T) {
	type in struct {
		*State
		int
	}
	type out struct {
		*State
		Data
		bool
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"state-nil",
			in{nil, 0},
			out{nil, nil, false},
		}, {
			"buff-nil",
			in{&State{capa: 5, mode: DiscardFirst, head: 0, tail: 0, buff: nil}, 0},
			out{&State{capa: 5, mode: DiscardFirst, head: 0, tail: 0, buff: nil}, nil, false},
		}, {
			"buff-zero",
			in{&State{capa: 5, mode: DiscardFirst, head: 0, tail: 0, buff: &RuneBuffer{}}, 0},
			out{&State{capa: 5, mode: DiscardFirst, head: 0, tail: 0, buff: &RuneBuffer{}}, nil, false},
		}, {
			"invalid-index",
			in{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 3, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}, 2},
			out{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 3, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}, nil, false},
		}, {
			"circular",
			in{&State{capa: 5, mode: DiscardFirst, head: 19, tail: 22, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}, 0},
			out{&State{capa: 5, mode: DiscardFirst, head: 19, tail: 21, buff: &RuneBuffer{'b', 'b', 'c', 'd', 'a'}}, 'e', true},
		}, {
			"first",
			in{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 3, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}, 0},
			out{&State{capa: 5, mode: DiscardFirst, head: 1, tail: 2, buff: &RuneBuffer{'a', 'c', 'c', 'd', 'e'}}, 'b', true},
		}, {
			"last",
			in{&State{capa: 5, mode: DiscardFirst, head: 4, tail: 6, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}, 1},
			out{&State{capa: 5, mode: DiscardFirst, head: 4, tail: 5, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}}, 'a', true},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			Data, bool := tt.in.State.Remove(tt.in.int)
			if !reflect.DeepEqual(tt.in.State, tt.out.State) {
				t.Errorf("State.Remove(%d): State: got:%s != want:%s", tt.in.int, tt.in.State, tt.out.State)
			}
			if !reflect.DeepEqual(Data, tt.out.Data) {
				t.Errorf("State.Remove(%d): Data: got:%+v != want:%+v", tt.in.int, Data, tt.out.Data)
			}
			if bool != tt.out.bool {
				t.Errorf("State.Remove(%d): bool: got:%v != want:%v", tt.in.int, bool, tt.out.bool)
			}
		})
	}
}
