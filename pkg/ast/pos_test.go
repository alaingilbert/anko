package ast

import (
	"reflect"
	"testing"
)

func TestPosImpl_Position(t *testing.T) {
	type fields struct {
		pos Position
	}
	tests := []struct {
		name   string
		fields fields
		want   Position
	}{
		{"", fields{Position{1, 2}}, Position{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := &PosImpl{
				pos: tt.fields.pos,
			}
			if got := x.Position(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Position() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPosImpl_SetPosition(t *testing.T) {
	type fields struct {
		pos Position
	}
	type args struct {
		pos Position
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"", fields{Position{1, 2}}, args{Position{1, 2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := &PosImpl{
				pos: tt.fields.pos,
			}
			x.SetPosition(tt.args.pos)
		})
	}
}
