package sequence

import (
	"reflect"
	"testing"

	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/volatile"
)

func TestSequence_Append(t *testing.T) {
	tests := []struct {
		name string
		seq  *Sequence
		args []byte
		want []byte
		err  bool
	}{
		{
			name: "nil-seq",
			seq:  nil,
			args: []byte{1, 2, 3, 4, 5},
			want: []byte{1, 2, 3, 4, 5},
			err:  true,
		}, {
			name: "zero-seq",
			seq:  (&Sequence{}).Configure(),
			args: []byte{1, 2, 3, 4, 5},
			want: []byte{},
			err:  false,
		}, {
			name: "nil-arg",
			seq:  (&Sequence{}).Configure(),
			args: nil,
			want: nil,
			err:  true,
		}, {
			name: "zero-arg",
			seq:  (&Sequence{}).Configure(),
			args: []byte{},
			want: []byte{},
			err:  false,
		}, {
			name: "full-seq",
			seq:  &Sequence{tail: volatile.Register32{Reg: config.BytesPerSequence}, valid: true},
			args: []byte{1, 2, 3, 4, 5},
			want: []byte{1, 2, 3, 4, 5},
			err:  true,
		}, {
			name: "part-seq",
			seq:  &Sequence{tail: volatile.Register32{Reg: config.BytesPerSequence - 3}, valid: true},
			args: []byte{1, 2, 3, 4, 5},
			want: []byte{4, 5},
			err:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.seq.Append(tt.args); !reflect.DeepEqual(tt.args[got:], tt.want) {
				t.Errorf("Sequence.Append() = %v, want %v", tt.args[got:], tt.want)
			} else if (err != nil) != tt.err {
				t.Errorf("Sequence.Append(): error(%t) != %v", tt.err, err)
			}

			if tt.seq == nil {
				if seq := tt.seq.Configure(); seq != nil {
					t.Errorf("Sequence.Configure() = %v, want %v", seq, nil)
				}
				if len := tt.seq.Len(); len != 0 {
					t.Errorf("Sequence.Len() = %d, want %d", len, 0)
				}
			}
			if seq := tt.seq.Reset(); seq.Len() != 0 {
				t.Errorf("Sequence.Reset(): Len=%d != %d", seq.Len(), 0)
			}
		})
	}
}
