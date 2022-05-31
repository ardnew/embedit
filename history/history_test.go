package history

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ardnew/embedit/terminal/line"
)

func TestHistory_Configure(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		h    *History
		want *History
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.h.Configure(nil, nil)); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}

func TestHistory_Len(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		h    *History
		want int
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.h.Len()); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}

func TestHistory_Add(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		h    *History
		args struct {
			ln line.Line
		}
		want *History
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tt.h.Add(tt.args.ln)
			if diff := cmp.Diff(tt.want, tt.h); len(diff) > 0 {
				t.Errorf("dirr (-want +got):%s\n", diff)
			}
		})
	}
}

func TestHistory_Get(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		h    *History
		args struct {
			n int
		}
		wantLn line.Line
		wantOk bool
	}{
		// TODO: Add test cases.
		"": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotLn, gotOk := tt.h.Get(tt.args.n)
			if diff := cmp.Diff(tt.wantLn, gotLn); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
			if diff := cmp.Diff(tt.wantOk, gotOk); len(diff) > 0 {
				t.Errorf("diff (-want +got):%s\n", diff)
			}
		})
	}
}
