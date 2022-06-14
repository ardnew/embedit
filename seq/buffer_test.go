package seq

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/seq/eol"
	"github.com/ardnew/embedit/volatile"
)

var seqByte = [config.BytesPerBuffer]byte{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
	'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p',
	'q', 'r', 's', 't', 'u', 'v', 'w', 'x',
	'y', 'z', '0', '1', '2', '3', '4', '5',
}

func TestBuffer_Configure(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Buffer
		want *Buffer
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

func TestBuffer_Len(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Buffer
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

func TestBuffer_Reset(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Buffer
		want *Buffer
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

func TestBuffer_Read(t *testing.T) {
	t.Parallel()
	type args struct {
		p []byte
	}
	for name, tt := range map[string]struct {
		s       *Buffer
		args    args
		wantN   int
		wantErr bool
		want    *Buffer
	}{
		// TODO: Add test cases.
		"full-seq": {
			s: &Buffer{
				Byte:  seqByte,
				head:  volatile.Register32{Reg: 0},
				tail:  volatile.Register32{Reg: 32},
				valid: true,
			},
			args:    args{p: make([]byte, config.BytesPerBuffer-1)},
			wantN:   31,
			wantErr: false,
			want: &Buffer{
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
				t.Errorf("diff Buffer got:%+v != want:%+v\n", tt.s, tt.want)
			}
		})
	}
}

func TestBuffer_Write(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Buffer
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

func TestBuffer_ReadFrom(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s    *Buffer
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

func TestBuffer_WriteTo(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		s       *Buffer
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
