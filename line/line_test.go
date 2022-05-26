package line

import (
	"reflect"
	"testing"

	"github.com/ardnew/embedit/limit"
	"github.com/ardnew/embedit/volatile"
)

func TestLine_Configure(t *testing.T) {
	type fields struct {
		Rune  [limit.RunesPerLine]rune
		Pos   volatile.Register32
		valid bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *Line
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Line{
				Rune:  tt.fields.Rune,
				Pos:   tt.fields.Pos,
				valid: tt.fields.valid,
			}
			if got := l.Configure(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.Configure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_init(t *testing.T) {
	type fields struct {
		Rune  [limit.RunesPerLine]rune
		Pos   volatile.Register32
		valid bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *Line
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Line{
				Rune:  tt.fields.Rune,
				Pos:   tt.fields.Pos,
				valid: tt.fields.valid,
			}
			if got := l.init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_Reset(t *testing.T) {
	type fields struct {
		Rune  [limit.RunesPerLine]rune
		Pos   volatile.Register32
		valid bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *Line
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Line{
				Rune:  tt.fields.Rune,
				Pos:   tt.fields.Pos,
				valid: tt.fields.valid,
			}
			if got := l.Reset(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.Reset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_String(t *testing.T) {
	type fields struct {
		Rune  [limit.RunesPerLine]rune
		Pos   volatile.Register32
		valid bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Line{
				Rune:  tt.fields.Rune,
				Pos:   tt.fields.Pos,
				valid: tt.fields.valid,
			}
			if got := l.String(); got != tt.want {
				t.Errorf("Line.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
