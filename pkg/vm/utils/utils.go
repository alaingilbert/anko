package utils

import (
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/utils"
	"reflect"
	"strings"
)

var (
	NilType  = reflect.TypeOf(nil)
	NilValue = reflect.New(reflect.TypeOf((*any)(nil)).Elem()).Elem()
)

var ErrTypeMismatch = errors.New("type mismatch")

// StronglyTyped is a special type that let the vm know that the value is strongly typed and should keep its type
type StronglyTyped struct{ V reflect.Value }

func ReplaceInterface(in string) string { return strings.ReplaceAll(in, "interface {}", "any") }

func FormatValue(value reflect.Value) string {
	if !value.IsValid() {
		return "<nil>"
	}
	if value.Kind() == reflect.Func {
		numIn := value.Type().NumIn()
		inParams := make([]string, numIn)
		for i := 0; i < numIn; i++ {
			inParams[i] = value.Type().In(i).String()
			inParams[i] = ReplaceInterface(inParams[i])
		}
		numOut := value.Type().NumOut()
		outParams := make([]string, numOut)
		for i := 0; i < numOut; i++ {
			outParams[i] = value.Type().Out(i).String()
			outParams[i] = ReplaceInterface(outParams[i])
		}

		inParamsStr := utils.Ternary(numIn > 0, strings.Join(inParams, ", "), "")
		sign := fmt.Sprintf("func(%s)", inParamsStr)
		if numOut > 0 {
			outParamsStr := strings.Join(outParams, ", ")
			if numOut > 1 {
				outParamsStr = fmt.Sprintf("(%s)", outParamsStr)
			}
			sign += fmt.Sprintf(" %s", outParamsStr)
		}
		return sign
	}
	s := fmt.Sprintf("%#v", value)
	s = ReplaceInterface(s)
	return s
}

func KindIsNumeric(kind reflect.Kind) bool {
	return kind == reflect.Int ||
		kind == reflect.Int8 ||
		kind == reflect.Int16 ||
		kind == reflect.Int32 ||
		kind == reflect.Int64 ||
		kind == reflect.Uint ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64 ||
		kind == reflect.Uintptr ||
		kind == reflect.Float32 ||
		kind == reflect.Float64
}
