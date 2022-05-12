package fifo

import (
	"reflect"
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

func TestOverflowMode_String(t *testing.T) {
	tests := []struct {
		name string
		m    OverflowMode
		want string
	}{
		{
			name: "discard-last",
			m:    DiscardLast,
			want: "OverflowMode(DiscardLast)",
		},
		{
			name: "discard-first",
			m:    DiscardFirst,
			want: "OverflowMode(DiscardFirst)",
		},
		{
			name: "mode-invalid",
			m:    OverflowMode(^byte(0)),
			want: "<OverflowMode(0xFF):invalid>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("OverflowMode.String(): got:%q != want:%q", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		buff Buffer
		capa int
		mode OverflowMode
	}
	tests := []struct {
		name      string
		args      args
		wantState *State
		wantCap   int
		wantRem   int
	}{
		{
			name:      "state-nil",
			args:      args{},
			wantState: &State{capa: 0, head: 0, tail: 0, mode: DiscardLast, buff: nil},
			wantCap:   0,
			wantRem:   0,
		},
		{
			name:      "buff-zero",
			args:      args{buff: &RuneBuffer{}},
			wantState: &State{capa: 0, head: 0, tail: 0, mode: DiscardLast, buff: &RuneBuffer{}},
			wantCap:   0,
			wantRem:   0,
		},
		{
			name:      "greater-than-phy",
			args:      args{buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}, capa: 10, mode: DiscardFirst},
			wantState: &State{capa: 5, head: 0, tail: 0, mode: DiscardFirst, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}},
			wantCap:   5,
			wantRem:   4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := New(tt.args.buff, tt.args.capa, tt.args.mode)
			if !reflect.DeepEqual(state, tt.wantState) {
				t.Errorf("New(%s, %d, %s): got:%s != want:%s",
					tt.args.buff.String(), tt.args.capa, tt.args.mode.String(), state.String(), tt.wantState.String())
			}
			if cap := state.Cap(); cap != tt.wantCap {
				t.Errorf("Cap(): got:%d != want:%d", cap, tt.wantCap)
			}
			if rem := state.Rem(); rem != tt.wantRem {
				t.Errorf("Rem(): got:%d != want:%d", rem, tt.wantRem)
			}
		})
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		name  string
		state *State
		want  string
	}{
		{
			name:  "basic",
			state: &State{capa: 5, mode: DiscardLast, head: 1, tail: 4, buff: &RuneBuffer{z, 'a', 'b', 'c', z}},
			want:  "{mode:OverflowMode(DiscardLast), capa:5, head:1[1], tail:4[4], size:3, buff:[.abc.]}",
		},
		{
			name:  "state-nil",
			state: nil,
			want:  "<nil>",
		},
		{
			name:  "buff-nil",
			state: &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: nil},
			want:  "{mode:OverflowMode(DiscardLast), capa:0, head:0[0], tail:0[0], size:0, buff:<nil>}",
		},
		{
			name:  "buff-zero",
			state: &State{buff: &RuneBuffer{}},
			want:  "{mode:OverflowMode(DiscardLast), capa:0, head:0[0], tail:0[0], size:0, buff:[]}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("State.String(): got:%q != want:%q", got, tt.want)
			}
		})
	}
}

func TestState_Len(t *testing.T) {
	tests := []struct {
		name    string
		state   *State
		wantLen int
	}{
		{
			name:    "state-zero",
			state:   &State{},
			wantLen: 0,
		},
		{
			name:    "greater-than-cap",
			state:   &State{head: 2450, tail: 2500},
			wantLen: 50,
		},
		{
			name:    "uint32-overflow",
			state:   &State{head: (1 << 32) - 1, tail: 0},
			wantLen: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len := tt.state.Len(); len != tt.wantLen {
				t.Errorf("Fifo.Len(): got:%v != want:%v", len, tt.wantLen)
			}
		})
	}
}

func TestState_Deq(t *testing.T) {
	tests := []struct {
		name     string
		state    *State
		wantData Data
		wantOK   bool
	}{
		{
			name:     "state-nil",
			state:    nil,
			wantData: nil,
			wantOK:   false,
		},
		{
			name:     "state-zero",
			state:    &State{},
			wantData: nil,
			wantOK:   false,
		},
		{
			name:     "buff-nil",
			state:    &State{capa: 10, head: 0, tail: 9, buff: nil},
			wantData: nil,
			wantOK:   false,
		},
		{
			name:   "buff-empty",
			state:  &State{capa: 10, head: 0, tail: 0, buff: &RuneBuffer{}},
			wantOK: false,
		},
		{
			name:     "buff-single",
			state:    &State{capa: 10, head: 9, tail: 10, buff: &RuneBuffer{z, z, z, z, z, z, z, z, z, 'X'}},
			wantData: 'X',
			wantOK:   true,
		},
		{
			name:     "buff-full",
			state:    &State{capa: 10, head: 12, tail: 21, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
			wantData: 'c',
			wantOK:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := tt.state
			data, ok := sc.Deq()
			if !reflect.DeepEqual(data, tt.wantData) {
				t.Errorf("Fifo.Deq(): data: got:%+v != want:%+v", data, tt.wantData)
			}
			if ok != tt.wantOK {
				t.Errorf("Fifo.Deq(): ok: got:%v != want:%v", ok, tt.wantOK)
			}
		})
	}
}

func TestState_Enq(t *testing.T) {
	tests := []struct {
		name      string
		state     *State
		data      Data
		wantState *State
		wantOK    bool
	}{
		{
			name:      "state-nil",
			state:     nil,
			data:      'X',
			wantState: nil,
			wantOK:    false,
		},
		{
			name:      "state-zero",
			state:     &State{},
			data:      'X',
			wantState: &State{},
			wantOK:    false,
		},
		{
			name:      "buff-nil",
			state:     &State{capa: 10, mode: DiscardLast, head: 0, tail: 9, buff: nil},
			data:      'X',
			wantState: &State{capa: 10, mode: DiscardLast, head: 0, tail: 9, buff: nil},
			wantOK:    false,
		},
		{
			name:      "buff-empty",
			state:     &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
			data:      'X',
			wantState: &State{capa: 5, mode: DiscardLast, head: 0, tail: 1, buff: &RuneBuffer{'X', z, z, z, z}},
			wantOK:    true,
		},
		{
			name:      "buff-single",
			state:     &State{capa: 10, mode: DiscardLast, head: 9, tail: 10, buff: &RuneBuffer{z, z, z, z, z, z, z, z, z, 'X'}},
			data:      'X',
			wantState: &State{capa: 10, mode: DiscardLast, head: 9, tail: 11, buff: &RuneBuffer{'X', z, z, z, z, z, z, z, z, 'X'}},
			wantOK:    true,
		},
		{
			name:      "discard-last",
			state:     &State{capa: 10, mode: DiscardLast, head: 12, tail: 21, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
			data:      'X',
			wantState: &State{capa: 10, mode: DiscardLast, head: 12, tail: 21, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
			wantOK:    false,
		},
		{
			name:      "discard-first",
			state:     &State{capa: 10, mode: DiscardFirst, head: 12, tail: 21, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
			data:      'X',
			wantState: &State{capa: 10, mode: DiscardFirst, head: 13, tail: 22, buff: &RuneBuffer{'a', 'X', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}},
			wantOK:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := tt.state
			ok := sc.Enq(tt.data)
			if !reflect.DeepEqual(sc, tt.wantState) {
				t.Errorf("Fifo.Enq(%+v): State: got:%s != want:%s", tt.data, sc.String(), tt.wantState.String())
			}
			if ok != tt.wantOK {
				t.Errorf("Fifo.Enq(%+v): ok: got:%v != want:%v", tt.data, ok, tt.wantOK)
			}
		})
	}
}

func TestState_Read(t *testing.T) {
	// &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}}
	tests := []struct {
		name     string
		state    *State
		data     []Data
		want     int
		wantErr  bool
		wantData []Data
	}{
		{
			name:     "state-nil",
			state:    nil,
			data:     make([]Data, 10),
			want:     0,
			wantErr:  true,
			wantData: []Data{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		},
		{
			name:     "buff-nil",
			state:    &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: nil},
			data:     make([]Data, 10),
			want:     0,
			wantErr:  true,
			wantData: []Data{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		},
		{
			name:     "data-nil",
			state:    &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
			data:     nil,
			want:     0,
			wantErr:  true,
			wantData: nil,
		},
		{
			name:     "data-empty",
			state:    &State{capa: 5, mode: DiscardLast, head: 0, tail: 1, buff: &RuneBuffer{'X', z, z, z, z}},
			data:     []Data{},
			want:     0,
			wantErr:  true,
			wantData: []Data{},
		},
		{
			name:     "buff-empty",
			state:    &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
			data:     make([]Data, 10),
			want:     0,
			wantErr:  true,
			wantData: []Data{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		},
		{
			name:     "buff-short",
			state:    &State{capa: 5, mode: DiscardLast, head: 8, tail: 10, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}},
			data:     make([]Data, 10),
			want:     2,
			wantErr:  false,
			wantData: []Data{'d', 'e', nil, nil, nil, nil, nil, nil, nil, nil},
		},
		{
			name:     "data-short",
			state:    &State{capa: 5, mode: DiscardLast, head: 6, tail: 10, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e'}},
			data:     make([]Data, 2),
			want:     2,
			wantErr:  false,
			wantData: []Data{'b', 'c'},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.state.Read(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("State.Read(): error: %v (want-error: %v)", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("State.Read(): int: got:%d != want:%d", got, tt.want)
			}
			if !reflect.DeepEqual(tt.data, tt.wantData) {
				t.Errorf("State.Read(): data: got:%+v != want:%+v", tt.data, tt.wantData)
			}
		})
	}
}

func TestState_Write(t *testing.T) {
	tests := []struct {
		name      string
		state     *State
		data      []Data
		want      int
		wantErr   bool
		wantState *State
	}{
		{
			name:      "state-nil",
			state:     nil,
			data:      []Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			want:      0,
			wantErr:   true,
			wantState: nil,
		},
		{
			name:      "buff-nil",
			state:     &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: nil},
			data:      []Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			want:      0,
			wantErr:   true,
			wantState: &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: nil},
		},
		{
			name:      "data-nil",
			state:     &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
			data:      nil,
			want:      0,
			wantErr:   true,
			wantState: &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
		},
		{
			name:      "data-empty",
			state:     &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
			data:      []Data{},
			want:      0,
			wantErr:   true,
			wantState: &State{capa: 5, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
		},
		{
			name:      "zero-cap",
			state:     &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
			data:      []Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			want:      0,
			wantErr:   true,
			wantState: &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
		},
		{
			name:      "mode-invalid",
			state:     &State{capa: 5, mode: OverflowMode(^byte(0)), head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
			data:      []Data{'a'},
			want:      0,
			wantErr:   true,
			wantState: &State{capa: 5, mode: OverflowMode(^byte(0)), head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
		},
		{
			name:      "discard-last",
			state:     &State{capa: 5, mode: DiscardLast, head: 2, tail: 6, buff: &RuneBuffer{'d', z, 'a', 'b', 'c'}},
			data:      []Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			want:      0,
			wantErr:   true,
			wantState: &State{capa: 5, mode: DiscardLast, head: 2, tail: 6, buff: &RuneBuffer{'d', z, 'a', 'b', 'c'}},
		},
		{
			name:      "short-write",
			state:     &State{capa: 10, mode: DiscardLast, head: 0, tail: 7, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', z, z, z}},
			data:      []Data{'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'},
			want:      2,
			wantErr:   false,
			wantState: &State{capa: 10, mode: DiscardLast, head: 0, tail: 9, buff: &RuneBuffer{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'q', 'r', z}},
		},
		{
			name:      "discard-first",
			state:     &State{capa: 5, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'d', z, 'a', 'b', 'c'}},
			data:      []Data{'q', 'r', 's'},
			want:      3,
			wantErr:   false,
			wantState: &State{capa: 5, mode: DiscardFirst, head: 5, tail: 9, buff: &RuneBuffer{'d', 'q', 'r', 's', 'c'}},
		},
		{
			name:      "discard-cap",
			state:     &State{capa: 5, mode: DiscardFirst, head: 2, tail: 6, buff: &RuneBuffer{'d', z, 'a', 'b', 'c'}},
			data:      []Data{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'},
			want:      10,
			wantErr:   false,
			wantState: &State{capa: 5, mode: DiscardFirst, head: 0, tail: 4, buff: &RuneBuffer{'g', 'h', 'i', 'j', 'c'}},
		},
		{
			name:      "last-zero-cap",
			state:     &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
			data:      make([]Data, 10),
			want:      0,
			wantErr:   true,
			wantState: &State{capa: 0, mode: DiscardLast, head: 0, tail: 0, buff: &RuneBuffer{z, z, z, z, z}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.state.Write(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("State.Write(): error: %v (want-error: %v)", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("State.Write(): int: got:%d != want:%d", got, tt.want)
			}
			if !reflect.DeepEqual(tt.state, tt.wantState) {
				t.Errorf("State.Write(): State: got:%s != want:%s", tt.state.String(), tt.wantState.String())
			}
		})
	}
}
