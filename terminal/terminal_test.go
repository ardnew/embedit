package terminal

import (
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTerminal_Configure(t *testing.T) {
	t.Parallel()
	type args struct {
		rw     io.ReadWriter
		width  int
		height int
	}
	for name, tt := range map[string]struct {
		tr   *Terminal
		args args
		want *Terminal
	}{
		// TODO: Add test cases.
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.tr.Configure(tt.args.rw, tt.args.width, tt.args.height)); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}
