// Package core implements core interface for anko script.

package runner

import (
	"fmt"
	"github.com/alaingilbert/anko/pkg/vm/env"
	"os"
	"reflect"
	"sort"
)

// Import defines core language builtins - keys, range, println,  etc.
func Import(env env.IEnv) env.IEnv {
	_ = env.Define("keys", keysFn)
	_ = env.Define("range", rangeFn)
	_ = env.Define("typeOf", typeOfFn)
	_ = env.Define("kindOf", kindOfFn)
	_ = env.Define("chanOf", chanOfFn)
	_ = env.Define("defined", definedFn(env))
	_ = env.Define("panic", panicFn)
	_ = env.Define("print", fmt.Print)
	_ = env.Define("println", fmt.Println)
	_ = env.Define("printf", fmt.Printf)
	_ = env.Define("close", closeFn)

	ImportToX(env)

	return env
}

func sortAndMax(arr [][]string) (maxLen int) {
	sort.Slice(arr, func(i, j int) bool { return arr[i][0] < arr[j][0] })
	for _, v := range arr {
		maxLen = max(maxLen, len(v[0]))
	}
	return maxLen
}

// Given a map, returns its keys
// > keys({"a": 1, "b": 2})
// []interface {}{"a", "b"}
func keysFn(v any) []any {
	rv := reflect.ValueOf(v)
	rv = elemIfInterface(rv)
	mapKeysValue := rv.MapKeys()
	mapKeys := make([]any, len(mapKeysValue))
	for i := 0; i < len(mapKeysValue); i++ {
		mapKeys[i] = mapKeysValue[i].Interface()
	}
	return mapKeys
}

func rangeFn(args ...int64) []int64 {
	var start, stop int64
	var step int64 = 1

	switch len(args) {
	case 0:
		panic("range expected at least 1 argument, got 0")
	case 1:
		stop = args[0]
	case 2:
		start = args[0]
		stop = args[1]
	case 3:
		start = args[0]
		stop = args[1]
		step = args[2]
		if step == 0 {
			panic("range argument 3 must not be zero")
		}
	default:
		panic(fmt.Sprintf("range expected at most 3 arguments, got %d", len(args)))
	}

	arr := make([]int64, 0)
	for i := start; (step > 0 && i < stop) || (step < 0 && i > stop); i += step {
		arr = append(arr, i)
	}
	return arr
}

func typeOfFn(v any) string {
	return reflect.TypeOf(v).String()
}

func kindOfFn(v any) string {
	typeOf := reflect.TypeOf(v)
	if typeOf == nil {
		return "nil"
	}
	return typeOf.Kind().String()
}

// Create a chan of something, not sure how to use this?
func chanOfFn(t reflect.Type) reflect.Value {
	return reflect.MakeChan(t, 1)
}

// Returns either or not a value (variable) is defined in our env
func definedFn(env env.IEnv) func(s string) bool {
	return func(s string) bool {
		_, err := env.Get(s)
		return err == nil
	}
}

// Panic and crash the vm
func panicFn(e any) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	panic(e)
}

// Close a channel (or anything that can be closed)
func closeFn(e any) {
	reflect.ValueOf(e).Close()
}
