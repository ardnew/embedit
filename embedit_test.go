package embedit

import (
	"reflect"
	"testing"
)

func TestCommandLine_String(t *testing.T) {
	type in struct {
		*CommandLine
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
			"cli-nil",
			in{nil},
			out{"<nil>"},
		}, {
			"cli-zero",
			in{&CommandLine{}},
			out{"{History:[]}"},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			if string := tt.in.CommandLine.String(); string != tt.out.string {
				t.Errorf("CommandLine.String(): string: got:%q != want:%q", string, tt.out.string)
			}
		})
	}
}

func TestControlSequence_Set(t *testing.T) {
	type in struct {
		*ControlSequence
		param []byte
		inter []byte
		code  byte
	}
	type out struct {
		*ControlSequence
		int
	}
	for _, tt := range []struct {
		string
		in
		out
	}{
		{
			"nil-receiver",
			in{nil, []byte("1;\x02;\x0a"), []byte(" !/"), 'Z'},
			out{nil, 0},
		}, {
			"param-recode",
			in{&ControlSequence{}, []byte("1;\x02;:"), []byte(" !/"), 'Z'},
			out{&ControlSequence{[16]byte{0x1B, '[', '1', ';', '2', ';', ':', ' ', '!', '/', 'Z', 0, 0, 0, 0, 0}}, 11},
		}, {
			"param-invalid",
			in{&ControlSequence{}, []byte("1 "), []byte(" !/"), 'Z'},
			out{&ControlSequence{[16]byte{0x1B, '[', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}, 3},
		}, {
			"inter-invalid",
			in{&ControlSequence{}, []byte("1;\x02;:"), []byte(" !/1"), 'Z'},
			out{&ControlSequence{[16]byte{0x1B, '[', '1', ';', '2', ';', ':', ' ', '!', '/', 0, 0, 0, 0, 0, 0}}, 10},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			if int := tt.in.ControlSequence.Set(tt.in.param, tt.in.inter, tt.in.code); int != tt.out.int {
				t.Errorf("ControlSequence.Set(): int: got:%d != want:%d", int, tt.out.int)
			}
			if !reflect.DeepEqual(tt.in.ControlSequence, tt.out.ControlSequence) {
				t.Errorf("ControlSequence.Set(%+v, %+v, %d): got:%s != want:%s",
					tt.in.param, tt.in.inter, tt.in.code,
					tt.in.ControlSequence.String(), tt.out.ControlSequence.String())
			}
		})
	}
}

func TestControlSequence_String(t *testing.T) {
	type in struct {
		*ControlSequence
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
			"nil-receiver",
			in{nil},
			out{"<nil>"},
		}, {
			"zero-receiver",
			in{&ControlSequence{}},
			out{"<undef>"},
		}, {
			"basic",
			in{&ControlSequence{[16]byte{0x1B, '[', '1', ';', '2', ';', ':', ' ', '!', '/', 'Z', 0, 0, 0, 0, 0}}},
			out{"ESC[ 1;2;:  !/ Z"},
		}, {
			"param-ooo",
			in{&ControlSequence{[16]byte{0x1B, '[', '1', ';', '2', ';', '!', ':', '!', '/', 'Z', 0, 0, 0, 0, 0}}},
			out{"<invalid>"},
		},
	} {
		t.Run(tt.string, func(t *testing.T) {
			if string := tt.in.ControlSequence.String(); string != tt.out.string {
				t.Errorf("ControlSequence.String(): string got:%q != want:%q", string, tt.out.string)
			}
		})
	}
}
