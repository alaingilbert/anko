package utils

import (
	"reflect"
	"testing"
)

func TestFormatValue(t *testing.T) {
	f1 := func() any { return nil }
	f2 := func(any, any) (any, any) { return nil, nil }
	type args struct {
		value reflect.Value
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{reflect.ValueOf(1)}, "1"},
		{"", args{reflect.ValueOf(f1)}, "func() any"},
		{"", args{reflect.ValueOf(f2)}, "func(any, any) (any, any)"},
		{"", args{reflect.ValueOf(nil)}, "<nil>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatValue(tt.args.value); got != tt.want {
				t.Errorf("FormatValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
