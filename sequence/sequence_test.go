package sequence

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/sequence/eol"
	"github.com/ardnew/embedit/volatile"
)

var seqByte = [config.BytesPerSequence]byte{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
	'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p',
	'q', 'r', 's', 't', 'u', 'v', 'w', 'x',
	'y', 'z', '0', '1', '2', '3', '4', '5',
}

func TestSequence_Configure(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Sequence
		want *Sequence
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.s.Configure(eol.CRLF)); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}

func TestSequence_Len(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Sequence
		want int
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.s.Len()); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}

func TestSequence_Reset(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Sequence
		want *Sequence
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.s.reset()); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}

func TestSequence_Read(t *testing.T) {
	t.Parallel()
	type args struct {
		p []byte
	}
	for name, tt := range map[string]struct {
		s       *Sequence
		args    args
		wantN   int
		wantErr bool
		want    *Sequence
	}{
		// TODO: Add test cases.
		"full-seq": {
			s: &Sequence{
				Byte:  seqByte,
				head:  volatile.Register32{Reg: 0},
				tail:  volatile.Register32{Reg: 32},
				valid: true,
			},
			args:    args{p: make([]byte, config.BytesPerSequence-1)},
			wantN:   31,
			wantErr: false,
			want: &Sequence{
				Byte:  seqByte,
				head:  volatile.Register32{Reg: 31},
				tail:  volatile.Register32{Reg: 32},
				valid: true,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotN, err := tt.s.Read(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr(%t): error = %+v", tt.wantErr, err)
			}
			if diff := cmp.Diff(tt.wantN, gotN); len(diff) > 0 {
				t.Errorf("diff N (-want +got):%s\n", diff)
			}
			if !reflect.DeepEqual(tt.s, tt.want) {
				t.Errorf("diff Sequence got:%+v != want:%+v\n", tt.s, tt.want)
			}
		})
	}
}

func TestSequence_Write(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Sequence
		args struct {
			p []byte
		}
		wantN   int
		wantErr bool
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotN, err := tt.s.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr(%t): error = %+v", tt.wantErr, err)
			}
			if diff := cmp.Diff(tt.wantN, gotN); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}

func TestSequence_ReadFrom(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Sequence
		args struct {
			r io.Reader
		}
		wantN   int64
		wantErr bool
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotN, err := tt.s.ReadFrom(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr(%t): error = %+v", tt.wantErr, err)
			}
			if diff := cmp.Diff(tt.wantN, gotN); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}

func TestSequence_WriteTo(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s       *Sequence
		wantN   int64
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			w := &bytes.Buffer{}
			gotN, err := tt.s.WriteTo(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr(%t): error = %+v", tt.wantErr, err)
			}
			if diff := cmp.Diff(tt.wantN, gotN); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
			if diff := cmp.Diff(tt.wantW,
				w.String()); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}
