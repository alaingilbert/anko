package utils

import (
	"fmt"
	"github.com/alaingilbert/anko/pkg/utils"
	"reflect"
	"strings"
)

func FormatValue(value reflect.Value) string {
	if !value.IsValid() {
		return "<nil>"
	}
	replaceInterface := func(in string) string { return strings.ReplaceAll(in, "interface {}", "any") }
	if value.Kind() == reflect.Func {
		numIn := value.Type().NumIn()
		inParams := make([]string, numIn)
		for i := 0; i < numIn; i++ {
			inParams[i] = value.Type().In(i).String()
			inParams[i] = replaceInterface(inParams[i])
		}
		numOut := value.Type().NumOut()
		outParams := make([]string, numOut)
		for i := 0; i < numOut; i++ {
			outParams[i] = value.Type().Out(i).String()
			outParams[i] = replaceInterface(outParams[i])
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
	s = replaceInterface(s)
	return s
}
