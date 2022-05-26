package cursor

import (
	"reflect"
	"testing"

	"github.com/ardnew/embedit/volatile"
)

func TestCursor_Configure(t *testing.T) {
	type fields struct {
		screen Screener
		X      volatile.Register32
		Y      volatile.Register32
		MaxY   volatile.Register32
		valid  bool
	}
	type args struct {
		s Screener
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Cursor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cursor{
				screen: tt.fields.screen,
				X:      tt.fields.X,
				Y:      tt.fields.Y,
				MaxY:   tt.fields.MaxY,
				valid:  tt.fields.valid,
			}
			if got := c.Configure(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cursor.Configure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCursor_init(t *testing.T) {
	type fields struct {
		screen Screener
		X      volatile.Register32
		Y      volatile.Register32
		MaxY   volatile.Register32
		valid  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *Cursor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cursor{
				screen: tt.fields.screen,
				X:      tt.fields.X,
				Y:      tt.fields.Y,
				MaxY:   tt.fields.MaxY,
				valid:  tt.fields.valid,
			}
			if got := c.init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cursor.init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCursor_Set(t *testing.T) {
	type fields struct {
		screen Screener
		X      volatile.Register32
		Y      volatile.Register32
		MaxY   volatile.Register32
		valid  bool
	}
	type args struct {
		x int
		y int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wantX  int
		wantY  int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cursor{
				screen: tt.fields.screen,
				X:      tt.fields.X,
				Y:      tt.fields.Y,
				MaxY:   tt.fields.MaxY,
				valid:  tt.fields.valid,
			}
			gotX, gotY := c.Set(tt.args.x, tt.args.y)
			if gotX != tt.wantX {
				t.Errorf("Cursor.Set() gotX = %v, want %v", gotX, tt.wantX)
			}
			if gotY != tt.wantY {
				t.Errorf("Cursor.Set() gotY = %v, want %v", gotY, tt.wantY)
			}
		})
	}
}
