package history

import (
	"reflect"
	"testing"

	"github.com/ardnew/embedit/limit"
	"github.com/ardnew/embedit/line"
	"github.com/ardnew/embedit/volatile"
)

func TestHistory_Configure(t *testing.T) {
	type fields struct {
		line  [limit.LinesPerHistory]line.Line
		head  volatile.Register32
		size  volatile.Register32
		indx  volatile.Register32
		pend  line.Line
		valid bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *History
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				line:  tt.fields.line,
				head:  tt.fields.head,
				size:  tt.fields.size,
				indx:  tt.fields.indx,
				pend:  tt.fields.pend,
				valid: tt.fields.valid,
			}
			if got := h.Configure(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("History.Configure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistory_init(t *testing.T) {
	type fields struct {
		line  [limit.LinesPerHistory]line.Line
		head  volatile.Register32
		size  volatile.Register32
		indx  volatile.Register32
		pend  line.Line
		valid bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *History
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				line:  tt.fields.line,
				head:  tt.fields.head,
				size:  tt.fields.size,
				indx:  tt.fields.indx,
				pend:  tt.fields.pend,
				valid: tt.fields.valid,
			}
			if got := h.init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("History.init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistory_Len(t *testing.T) {
	type fields struct {
		line  [limit.LinesPerHistory]line.Line
		head  volatile.Register32
		size  volatile.Register32
		indx  volatile.Register32
		pend  line.Line
		valid bool
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				line:  tt.fields.line,
				head:  tt.fields.head,
				size:  tt.fields.size,
				indx:  tt.fields.indx,
				pend:  tt.fields.pend,
				valid: tt.fields.valid,
			}
			if got := h.Len(); got != tt.want {
				t.Errorf("History.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistory_Add(t *testing.T) {
	type fields struct {
		line  [limit.LinesPerHistory]line.Line
		head  volatile.Register32
		size  volatile.Register32
		indx  volatile.Register32
		pend  line.Line
		valid bool
	}
	type args struct {
		ln line.Line
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				line:  tt.fields.line,
				head:  tt.fields.head,
				size:  tt.fields.size,
				indx:  tt.fields.indx,
				pend:  tt.fields.pend,
				valid: tt.fields.valid,
			}
			h.Add(tt.args.ln)
		})
	}
}

func TestHistory_Get(t *testing.T) {
	type fields struct {
		line  [limit.LinesPerHistory]line.Line
		head  volatile.Register32
		size  volatile.Register32
		indx  volatile.Register32
		pend  line.Line
		valid bool
	}
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wantLn line.Line
		wantOk bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				line:  tt.fields.line,
				head:  tt.fields.head,
				size:  tt.fields.size,
				indx:  tt.fields.indx,
				pend:  tt.fields.pend,
				valid: tt.fields.valid,
			}
			gotLn, gotOk := h.Get(tt.args.n)
			if !reflect.DeepEqual(gotLn, tt.wantLn) {
				t.Errorf("History.Get() gotLn = %v, want %v", gotLn, tt.wantLn)
			}
			if gotOk != tt.wantOk {
				t.Errorf("History.Get() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
