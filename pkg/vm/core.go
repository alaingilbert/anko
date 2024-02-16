// Package core implements core interface for anko script.
package vm

import (
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/vm/env"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"os"
	"reflect"

	"github.com/alaingilbert/anko/pkg/parser"
)

// Import defines core language builtins - keys, range, println,  etc.
func Import(env env.IEnv) env.IEnv {
	_ = env.Define("keys", keysFn)
	_ = env.Define("range", rangeFn)
	_ = env.Define("typeOf", typeOfFn)
	_ = env.Define("kindOf", kindOfFn)
	_ = env.Define("chanOf", chanOfFn)
	_ = env.Define("defined", definedFn(env))
	_ = env.Define("load", loadFn(env))
	_ = env.Define("panic", panicFn)
	_ = env.Define("print", fmt.Print)
	_ = env.Define("println", fmt.Println)
	_ = env.Define("printf", fmt.Printf)
	_ = env.Define("close", closeFn)
	_ = env.Define("dbg", dbgFn)

	ImportToX(env)

	return env
}

func dbgFn(v any) {
	if e, ok := v.(*env.Env); ok {
		print(e.String())
		return
	}
	println(vmUtils.FormatValue(reflect.ValueOf(v)))
}

// Given a map, returns its keys
// > keys({"a": 1, "b": 2})
// []interface {}{"a", "b"}
func keysFn(v any) []any {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
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

// Dynamically load a file and execute it, return the RV value
func loadFn(env env.IEnv) func(s string) any {
	return func(s string) any {
		body, err := os.ReadFile(s)
		if err != nil {
			panic(err)
		}
		scanner := new(parser.Scanner)
		scanner.Init(string(body))
		stmts, err := parser.Parse(scanner)
		if err != nil {
			var pe *parser.Error
			if errors.As(err, &pe) {
				pe.Filename = s
				panic(pe)
			}
			panic(err)
		}
		rv, err := New(&Configs{Env: env, DoNotDeepCopyEnv: true}).Executor().Run(nil, stmts)
		if err != nil {
			panic(err)
		}
		return rv
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
