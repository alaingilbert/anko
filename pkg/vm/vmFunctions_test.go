package vm

import (
	"bytes"
	"context"
	"fmt"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"github.com/alaingilbert/anko/pkg/vm/runner"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestReturns(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `return 1++`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `return 1, 1++`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `return 1, 2, 1++`, RunError: fmt.Errorf("invalid operation"), Name: ""},

		{Script: `return`, RunOutput: nil, Name: ""},
		{Script: `return nil`, RunOutput: nil, Name: ""},
		{Script: `return true`, RunOutput: true, Name: ""},
		{Script: `return 1`, RunOutput: int64(1), Name: ""},
		{Script: `return 1.1`, RunOutput: float64(1.1), Name: ""},
		{Script: `return "a"`, RunOutput: "a", Name: ""},

		{Script: `b()`, Input: map[string]any{"b": func() {}}, RunOutput: nil, Name: ""},
		{Script: `b()`, Input: map[string]any{"b": func() reflect.Value { return reflect.Value{} }}, RunOutput: reflect.Value{}, Name: ""},
		{Script: `b()`, Input: map[string]any{"b": func() any { return nil }}, RunOutput: nil, Name: ""},
		{Script: `b()`, Input: map[string]any{"b": func() bool { return true }}, RunOutput: true, Name: ""},
		{Script: `b()`, Input: map[string]any{"b": func() int32 { return int32(1) }}, RunOutput: int32(1), Name: ""},
		{Script: `b()`, Input: map[string]any{"b": func() int64 { return int64(1) }}, RunOutput: int64(1), Name: ""},
		{Script: `b()`, Input: map[string]any{"b": func() float32 { return float32(1.1) }}, RunOutput: float32(1.1), Name: ""},
		{Script: `b()`, Input: map[string]any{"b": func() float64 { return float64(1.1) }}, RunOutput: float64(1.1), Name: ""},
		{Script: `b()`, Input: map[string]any{"b": func() string { return "a" }}, RunOutput: "a", Name: ""},

		{Script: `b(a)`, Input: map[string]any{"a": reflect.Value{}, "b": func(c reflect.Value) reflect.Value { return c }}, RunOutput: reflect.Value{}, Output: map[string]any{"a": reflect.Value{}}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": nil, "b": func(c any) any { return c }}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": true, "b": func(c bool) bool { return c }}, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": int32(1), "b": func(c int32) int32 { return c }}, RunOutput: int32(1), Output: map[string]any{"a": int32(1)}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": int64(1), "b": func(c int64) int64 { return c }}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": float32(1.1), "b": func(c float32) float32 { return c }}, RunOutput: float32(1.1), Output: map[string]any{"a": float32(1.1)}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": float64(1.1), "b": func(c float64) float64 { return c }}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": "a", "b": func(c string) string { return c }}, RunOutput: "a", Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `b(a)`, Input: map[string]any{"a": "a", "b": func(c bool) bool { return c }}, RunError: fmt.Errorf("function wants argument type bool but received type string"), Output: map[string]any{"a": "a"}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": int64(1), "b": func(c int32) int32 { return c }}, RunOutput: int32(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": int32(1), "b": func(c int64) int64 { return c }}, RunOutput: int64(1), Output: map[string]any{"a": int32(1)}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": float64(1.25), "b": func(c float32) float32 { return c }}, RunOutput: float32(1.25), Output: map[string]any{"a": float64(1.25)}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": float32(1.25), "b": func(c float64) float64 { return c }}, RunOutput: float64(1.25), Output: map[string]any{"a": float32(1.25)}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": true, "b": func(c string) string { return c }}, RunError: fmt.Errorf("function wants argument type string but received type bool"), Output: map[string]any{"a": true}, Name: ""},

		{Script: `b(a)`, Input: map[string]any{"a": testVarValueBool, "b": func(c any) any { return c }}, RunOutput: testVarValueBool, Output: map[string]any{"a": testVarValueBool}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueInt32, "b": func(c any) any { return c }}, RunOutput: testVarValueInt32, Output: map[string]any{"a": testVarValueInt32}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueInt64, "b": func(c any) any { return c }}, RunOutput: testVarValueInt64, Output: map[string]any{"a": testVarValueInt64}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueFloat32, "b": func(c any) any { return c }}, RunOutput: testVarValueFloat32, Output: map[string]any{"a": testVarValueFloat32}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueFloat64, "b": func(c any) any { return c }}, RunOutput: testVarValueFloat64, Output: map[string]any{"a": testVarValueFloat64}, Name: ""},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueString, "b": func(c any) any { return c }}, RunOutput: testVarValueString, Output: map[string]any{"a": testVarValueString}, Name: ""},

		{Script: `func aFunc() {}; aFunc()`, RunOutput: nil, Name: ""},
		{Script: `func aFunc() {return}; aFunc()`, RunOutput: nil, Name: ""},
		{Script: `func aFunc() {return}; a = aFunc()`, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},

		{Script: `func aFunc() {return nil}; aFunc()`, RunOutput: nil, Name: ""},
		{Script: `func aFunc() {return true}; aFunc()`, RunOutput: true, Name: ""},
		{Script: `func aFunc() {return 1}; aFunc()`, RunOutput: int64(1), Name: ""},
		{Script: `func aFunc() {return 1.1}; aFunc()`, RunOutput: float64(1.1), Name: ""},
		{Script: `func aFunc() {return "a"}; aFunc()`, RunOutput: "a", Name: ""},

		{Script: `func aFunc() {return 1 + 2}; aFunc()`, RunOutput: int64(3), Name: ""},
		{Script: `func aFunc() {return 1.25 + 2.25}; aFunc()`, RunOutput: float64(3.5), Name: ""},
		{Script: `func aFunc() {return "a" + "b"}; aFunc()`, RunOutput: "ab", Name: ""},

		{Script: `func aFunc() {return 1 + 2, 3 + 4}; aFunc()`, RunOutput: []any{int64(3), int64(7)}, Name: ""},
		{Script: `func aFunc() {return 1.25 + 2.25, 3.25 + 4.25}; aFunc()`, RunOutput: []any{float64(3.5), float64(7.5)}, Name: ""},
		{Script: `func aFunc() {return "a" + "b", "c" + "d"}; aFunc()`, RunOutput: []any{"ab", "cd"}, Name: ""},

		{Script: `func aFunc() {return nil, nil}; aFunc()`, RunOutput: []any{nil, nil}, Name: ""},
		{Script: `func aFunc() {return true, false}; aFunc()`, RunOutput: []any{true, false}, Name: ""},
		{Script: `func aFunc() {return 1, 2}; aFunc()`, RunOutput: []any{int64(1), int64(2)}, Name: ""},
		{Script: `func aFunc() {return 1.1, 2.2}; aFunc()`, RunOutput: []any{float64(1.1), float64(2.2)}, Name: ""},
		{Script: `func aFunc() {return "a", "b"}; aFunc()`, RunOutput: []any{"a", "b"}, Name: ""},

		{Script: `func aFunc() {return [nil]}; aFunc()`, RunOutput: []any{nil}, Name: ""},
		{Script: `func aFunc() {return [nil, nil]}; aFunc()`, RunOutput: []any{nil, nil}, Name: ""},
		{Script: `func aFunc() {return [nil, nil, nil]}; aFunc()`, RunOutput: []any{nil, nil, nil}, Name: ""},
		{Script: `func aFunc() {return [nil, nil], [nil, nil]}; aFunc()`, RunOutput: []any{[]any{nil, nil}, []any{nil, nil}}, Name: ""},

		{Script: `func aFunc() {return [true]}; aFunc()`, RunOutput: []any{true}, Name: ""},
		{Script: `func aFunc() {return [true, false]}; aFunc()`, RunOutput: []any{true, false}, Name: ""},
		{Script: `func aFunc() {return [true, false, true]}; aFunc()`, RunOutput: []any{true, false, true}, Name: ""},
		{Script: `func aFunc() {return [true, false], [false, true]}; aFunc()`, RunOutput: []any{[]any{true, false}, []any{false, true}}, Name: ""},

		{Script: `func aFunc() {return []}; aFunc()`, RunOutput: []any{}, Name: ""},
		{Script: `func aFunc() {return [1]}; aFunc()`, RunOutput: []any{int64(1)}, Name: ""},
		{Script: `func aFunc() {return [1, 2]}; aFunc()`, RunOutput: []any{int64(1), int64(2)}, Name: ""},
		{Script: `func aFunc() {return [1, 2, 3]}; aFunc()`, RunOutput: []any{int64(1), int64(2), int64(3)}, Name: ""},
		{Script: `func aFunc() {return [1, 2], [3, 4]}; aFunc()`, RunOutput: []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}, Name: ""},

		{Script: `func aFunc() {return [1.1]}; aFunc()`, RunOutput: []any{float64(1.1)}, Name: ""},
		{Script: `func aFunc() {return [1.1, 2.2]}; aFunc()`, RunOutput: []any{float64(1.1), float64(2.2)}, Name: ""},
		{Script: `func aFunc() {return [1.1, 2.2, 3.3]}; aFunc()`, RunOutput: []any{float64(1.1), float64(2.2), float64(3.3)}, Name: ""},
		{Script: `func aFunc() {return [1.1, 2.2], [3.3, 4.4]}; aFunc()`, RunOutput: []any{[]any{float64(1.1), float64(2.2)}, []any{float64(3.3), float64(4.4)}}, Name: ""},

		{Script: `func aFunc() {return ["a"]}; aFunc()`, RunOutput: []any{"a"}, Name: ""},
		{Script: `func aFunc() {return ["a", "b"]}; aFunc()`, RunOutput: []any{"a", "b"}, Name: ""},
		{Script: `func aFunc() {return ["a", "b", "c"]}; aFunc()`, RunOutput: []any{"a", "b", "c"}, Name: ""},
		{Script: `func aFunc() {return ["a", "b"], ["c", "d"]}; aFunc()`, RunOutput: []any{[]any{"a", "b"}, []any{"c", "d"}}, Name: ""},

		{Script: `func aFunc() {return nil, nil}; aFunc()`, RunOutput: []any{any(nil), any(nil)}, Name: ""},
		{Script: `func aFunc() {return true, false}; aFunc()`, RunOutput: []any{true, false}, Name: ""},
		{Script: `func aFunc() {return 1, 2}; aFunc()`, RunOutput: []any{int64(1), int64(2)}, Name: ""},
		{Script: `func aFunc() {return 1.1, 2.2}; aFunc()`, RunOutput: []any{float64(1.1), float64(2.2)}, Name: ""},
		{Script: `func aFunc() {return "a", "b"}; aFunc()`, RunOutput: []any{"a", "b"}, Name: ""},

		{Script: `func aFunc() {return a}; aFunc()`, Input: map[string]any{"a": reflect.Value{}}, RunOutput: reflect.Value{}, Output: map[string]any{"a": reflect.Value{}}, Name: ""},

		{Script: `func aFunc() {return a}; aFunc()`, Input: map[string]any{"a": nil}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `func aFunc() {return a}; aFunc()`, Input: map[string]any{"a": true}, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `func aFunc() {return a}; aFunc()`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `func aFunc() {return a}; aFunc()`, Input: map[string]any{"a": float64(1.1)}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `func aFunc() {return a}; aFunc()`, Input: map[string]any{"a": "a"}, RunOutput: "a", Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `func aFunc() {return a, a}; aFunc()`, Input: map[string]any{"a": reflect.Value{}}, RunOutput: []any{reflect.Value{}, reflect.Value{}}, Output: map[string]any{"a": reflect.Value{}}, Name: ""},
		{Script: `func aFunc() {return a, a}; aFunc()`, Input: map[string]any{"a": nil}, RunOutput: []any{nil, nil}, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `func aFunc() {return a, a}; aFunc()`, Input: map[string]any{"a": true}, RunOutput: []any{true, true}, Output: map[string]any{"a": true}, Name: ""},
		{Script: `func aFunc() {return a, a}; aFunc()`, Input: map[string]any{"a": int32(1)}, RunOutput: []any{int32(1), int32(1)}, Output: map[string]any{"a": int32(1)}, Name: ""},
		{Script: `func aFunc() {return a, a}; aFunc()`, Input: map[string]any{"a": int64(1)}, RunOutput: []any{int64(1), int64(1)}, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `func aFunc() {return a, a}; aFunc()`, Input: map[string]any{"a": float32(1.1)}, RunOutput: []any{float32(1.1), float32(1.1)}, Output: map[string]any{"a": float32(1.1)}, Name: ""},
		{Script: `func aFunc() {return a, a}; aFunc()`, Input: map[string]any{"a": float64(1.1)}, RunOutput: []any{float64(1.1), float64(1.1)}, Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `func aFunc() {return a, a}; aFunc()`, Input: map[string]any{"a": "a"}, RunOutput: []any{"a", "a"}, Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `func a(x) { return x}; a(nil)`, RunOutput: nil, Name: ""},
		{Script: `func a(x) { return x}; a(true)`, RunOutput: true, Name: ""},
		{Script: `func a(x) { return x}; a(1)`, RunOutput: int64(1), Name: ""},
		{Script: `func a(x) { return x}; a(1.1)`, RunOutput: float64(1.1), Name: ""},
		{Script: `func a(x) { return x}; a("a")`, RunOutput: "a", Name: ""},

		{Script: `func aFunc() {return a}; for {aFunc(); break}`, Input: map[string]any{"a": nil}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `func aFunc() {return a}; for {aFunc(); break}`, Input: map[string]any{"a": true}, RunOutput: nil, Output: map[string]any{"a": true}, Name: ""},
		{Script: `func aFunc() {return a}; for {aFunc(); break}`, Input: map[string]any{"a": int64(1)}, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `func aFunc() {return a}; for {aFunc(); break}`, Input: map[string]any{"a": float64(1.1)}, RunOutput: nil, Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `func aFunc() {return a}; for {aFunc(); break}`, Input: map[string]any{"a": "a"}, RunOutput: nil, Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `func aFunc() {for {return a}}; aFunc()`, Input: map[string]any{"a": nil}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `func aFunc() {for {return a}}; aFunc()`, Input: map[string]any{"a": true}, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `func aFunc() {for {return a}}; aFunc()`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `func aFunc() {for {return a}}; aFunc()`, Input: map[string]any{"a": float64(1.1)}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `func aFunc() {for {return a}}; aFunc()`, Input: map[string]any{"a": "a"}, RunOutput: "a", Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `func aFunc() {for {if true {return a}}}; aFunc()`, Input: map[string]any{"a": nil}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `func aFunc() {for {if true {return a}}}; aFunc()`, Input: map[string]any{"a": true}, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `func aFunc() {for {if true {return a}}}; aFunc()`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `func aFunc() {for {if true {return a}}}; aFunc()`, Input: map[string]any{"a": float64(1.1)}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `func aFunc() {for {if true {return a}}}; aFunc()`, Input: map[string]any{"a": "a"}, RunOutput: "a", Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `func aFunc() {return nil, nil}; a, b = aFunc()`, RunOutput: nil, Output: map[string]any{"a": nil, "b": nil}, Name: ""},
		{Script: `func aFunc() {return true, false}; a, b = aFunc()`, RunOutput: false, Output: map[string]any{"a": true, "b": false}, Name: ""},
		{Script: `func aFunc() {return 1, 2}; a, b = aFunc()`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},
		{Script: `func aFunc() {return 1.1, 2.2}; a, b = aFunc()`, RunOutput: float64(2.2), Output: map[string]any{"a": float64(1.1), "b": float64(2.2)}, Name: ""},
		{Script: `func aFunc() {return "a", "b"}; a, b = aFunc()`, RunOutput: "b", Output: map[string]any{"a": "a", "b": "b"}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestFunctions(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a()`, Input: map[string]any{"a": reflect.Value{}}, RunError: fmt.Errorf("cannot call type reflect.Value")},
		{Script: `a = nil; a()`, RunError: fmt.Errorf("cannot call type interface {}"), Output: map[string]any{"a": nil}},
		{Script: `a = true; a()`, RunError: fmt.Errorf("cannot call type bool"), Output: map[string]any{"a": true}},
		{Script: `a = nil; b = func c(d) { return d == nil }; c = nil; c(a)`, RunError: fmt.Errorf("cannot call type interface {}"), Output: map[string]any{"a": nil}},
		{Script: `a = [true]; a()`, RunError: fmt.Errorf("cannot call type []interface {}")},
		{Script: `a = [true]; func b(c) { return c() }; b(a)`, RunError: fmt.Errorf("cannot call type []interface {}")},
		{Script: `a = {}; a.missing()`, RunError: fmt.Errorf("cannot call type interface {}"), Output: map[string]any{"a": map[any]any{}}},
		{Script: `a = 1; b = func(,a){}; a`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(1)},

		{Script: `func a(b) { }; a()`, RunError: fmt.Errorf("function wants 1 arguments but received 0")},
		{Script: `func a(b) { }; a(true, true)`, RunError: fmt.Errorf("function wants 1 arguments but received 2")},
		{Script: `func a(b, c) { }; a()`, RunError: fmt.Errorf("function wants 2 arguments but received 0")},
		{Script: `func a(b, c) { }; a(true)`, RunError: fmt.Errorf("function wants 2 arguments but received 1")},
		{Script: `func a(b, c) { }; a(true, true, true)`, RunError: fmt.Errorf("function wants 2 arguments but received 3")},

		{Script: `func a() { return "a" }; a.b()`, RunError: fmt.Errorf("type func does not support member operation")},
		{Script: `a = [func () { return nil}]; func b(c) { return c() }; b(a[1])`, RunError: fmt.Errorf("index out of range")},
		{Script: `func a() { return "a" }; b()`, RunError: envPkg.NewUndefinedSymbolErr("b")},
		{Script: ` func a() { return "a" }; 1++()`, RunError: fmt.Errorf("invalid operation")},
		{Script: ` func a(b) { return b }; a(1++)`, RunError: fmt.Errorf("invalid operation")},

		{Script: `a`, Input: map[string]any{"a": testVarFunc}, RunOutput: testVarFunc, Output: map[string]any{"a": testVarFunc}},
		{Script: `a()`, Input: map[string]any{"a": testVarFunc}, RunOutput: int64(1), Output: map[string]any{"a": testVarFunc}},
		{Script: `a`, Input: map[string]any{"a": testVarFuncP}, RunOutput: testVarFuncP, Output: map[string]any{"a": testVarFuncP}},
		// TOFIX:
		// {Script: `a()`, Input: map[string]any{"a": testVarFuncP}, RunOutput: int64(1), Output: map[string]any{"a": testVarFuncP}},

		{Script: `module a { func b() { return } }; a.b()`, RunOutput: nil},
		{Script: `module a { func b() { return nil} }; a.b()`, RunOutput: nil},
		{Script: `module a { func b() { return true} }; a.b()`, RunOutput: true},
		{Script: `module a { func b() { return 1} }; a.b()`, RunOutput: int64(1)},
		{Script: `module a { func b() { return 1.1} }; a.b()`, RunOutput: float64(1.1)},
		{Script: `module a { func b() { return "a"} }; a.b()`, RunOutput: "a"},

		{Script: `if true { module a { func b() { return } } }; a.b()`, RunOutput: nil},
		{Script: `if true { module a { func b() { return nil} } }; a.b()`, RunOutput: nil},
		{Script: `if true { module a { func b() { return true} } }; a.b()`, RunOutput: true},
		{Script: `if true { module a { func b() { return 1} } }; a.b()`, RunOutput: int64(1)},
		{Script: `if true { module a { func b() { return 1.1} } }; a.b()`, RunOutput: float64(1.1)},
		{Script: `if true { module a { func b() { return "a"} } }; a.b()`, RunOutput: "a"},

		{Script: `if true { module a { func b() { return 1} } }; a.b()`, RunOutput: int64(1)},

		{Script: `a = 1; func b() { a = 2 }; b()`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `b(a); a`, Input: map[string]any{"a": int64(1), "b": func(c any) { c = int64(2); _ = c }}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `func b() { }; go b()`, RunOutput: nil},

		{Script: `b(a)`, Input: map[string]any{"a": nil, "b": func(c any) bool { return c == nil }}, RunOutput: true, Output: map[string]any{"a": nil}},
		{Script: `b(a)`, Input: map[string]any{"a": true, "b": func(c bool) bool { return c == true }}, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `b(a)`, Input: map[string]any{"a": int32(1), "b": func(c int32) bool { return c == 1 }}, RunOutput: true, Output: map[string]any{"a": int32(1)}},
		{Script: `b(a)`, Input: map[string]any{"a": int64(1), "b": func(c int64) bool { return c == 1 }}, RunOutput: true, Output: map[string]any{"a": int64(1)}},
		{Script: `b(a)`, Input: map[string]any{"a": float32(1.1), "b": func(c float32) bool { return c == 1.1 }}, RunOutput: true, Output: map[string]any{"a": float32(1.1)}},
		{Script: `b(a)`, Input: map[string]any{"a": float64(1.1), "b": func(c float64) bool { return c == 1.1 }}, RunOutput: true, Output: map[string]any{"a": float64(1.1)}},
		{Script: `b(a)`, Input: map[string]any{"a": "a", "b": func(c string) bool { return c == "a" }}, RunOutput: true, Output: map[string]any{"a": "a"}},

		{Script: `b(a)`, Input: map[string]any{"a": testVarValueBool, "b": func(c reflect.Value) bool { return c == testVarValueBool }}, RunOutput: true, Output: map[string]any{"a": testVarValueBool}},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueInt32, "b": func(c reflect.Value) bool { return c == testVarValueInt32 }}, RunOutput: true, Output: map[string]any{"a": testVarValueInt32}},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueInt64, "b": func(c reflect.Value) bool { return c == testVarValueInt64 }}, RunOutput: true, Output: map[string]any{"a": testVarValueInt64}},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueFloat32, "b": func(c reflect.Value) bool { return c == testVarValueFloat32 }}, RunOutput: true, Output: map[string]any{"a": testVarValueFloat32}},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueFloat64, "b": func(c reflect.Value) bool { return c == testVarValueFloat64 }}, RunOutput: true, Output: map[string]any{"a": testVarValueFloat64}},
		{Script: `b(a)`, Input: map[string]any{"a": testVarValueString, "b": func(c reflect.Value) bool { return c == testVarValueString }}, RunOutput: true, Output: map[string]any{"a": testVarValueString}},

		{Script: `x(a, b, c, d, e, f, g)`, Input: map[string]any{"a": nil, "b": true, "c": int32(1), "d": int64(2), "e": float32(1.1), "f": float64(2.2), "g": "g",
			"x": func(a any, b bool, c int32, d int64, e float32, f float64, g string) bool {
				return a == nil && b == true && c == 1 && d == 2 && e == 1.1 && f == 2.2 && g == "g"
			}}, RunOutput: true, Output: map[string]any{"a": nil, "b": true, "c": int32(1), "d": int64(2), "e": float32(1.1), "f": float64(2.2), "g": "g"}},
		{Script: `x(a, b, c, d, e, f, g)`, Input: map[string]any{"a": nil, "b": true, "c": int32(1), "d": int64(2), "e": float32(1.1), "f": float64(2.2), "g": "g",
			"x": func(a any, b bool, c int32, d int64, e float32, f float64, g string) (any, bool, int32, int64, float32, float64, string) {
				return a, b, c, d, e, f, g
			}}, RunOutput: []any{nil, true, int32(1), int64(2), float32(1.1), float64(2.2), "g"}, Output: map[string]any{"a": nil, "b": true, "c": int32(1), "d": int64(2), "e": float32(1.1), "f": float64(2.2), "g": "g"}},

		{Script: `b = a()`, Input: map[string]any{"a": func() (bool, int32, int64, float32, float64, string) { return true, 1, 2, 3.3, 4.4, "5" }}, RunOutput: []any{true, int32(1), int64(2), float32(3.3), float64(4.4), "5"}, Output: map[string]any{"b": []any{true, int32(1), int64(2), float32(3.3), float64(4.4), "5"}}},
		{Script: `b = a(); b`, Input: map[string]any{"a": func() (bool, int32, int64, float32, float64, string) { return true, 1, 2, 3.3, 4.4, "5" }}, RunOutput: []any{true, int32(1), int64(2), float32(3.3), float64(4.4), "5"}, Output: map[string]any{"b": []any{true, int32(1), int64(2), float32(3.3), float64(4.4), "5"}}},
		{Script: `b, c = a(); b`, Input: map[string]any{"a": func() (bool, int32, int64, float32, float64, string) { return true, 1, 2, 3.3, 4.4, "5" }}, RunOutput: true, Output: map[string]any{"b": true, "c": int32(1)}},
		{Script: `b, c, d = a(); b`, Input: map[string]any{"a": func() (bool, int32, int64, float32, float64, string) { return true, 1, 2, 3.3, 4.4, "5" }}, RunOutput: true, Output: map[string]any{"b": true, "c": int32(1), "d": int64(2)}},
		{Script: `b, c, d, e = a(); b`, Input: map[string]any{"a": func() (bool, int32, int64, float32, float64, string) { return true, 1, 2, 3.3, 4.4, "5" }}, RunOutput: true, Output: map[string]any{"b": true, "c": int32(1), "d": int64(2), "e": float32(3.3)}},
		{Script: `b, c, d, e, f = a(); b`, Input: map[string]any{"a": func() (bool, int32, int64, float32, float64, string) { return true, 1, 2, 3.3, 4.4, "5" }}, RunOutput: true, Output: map[string]any{"b": true, "c": int32(1), "d": int64(2), "e": float32(3.3), "f": float64(4.4)}},
		{Script: `b, c, d, e, f, g = a(); b`, Input: map[string]any{"a": func() (bool, int32, int64, float32, float64, string) { return true, 1, 2, 3.3, 4.4, "5" }}, RunOutput: true, Output: map[string]any{"b": true, "c": int32(1), "d": int64(2), "e": float32(3.3), "f": float64(4.4), "g": "5"}},

		{Script: `a = nil; b(a)`, Input: map[string]any{"b": func(c any) bool { return c == nil }}, RunOutput: true, Output: map[string]any{"a": nil}},
		{Script: `a = true; b(a)`, Input: map[string]any{"b": func(c bool) bool { return c == true }}, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = 1; b(a)`, Input: map[string]any{"b": func(c int64) bool { return c == 1 }}, RunOutput: true, Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1.1; b(a)`, Input: map[string]any{"b": func(c float64) bool { return c == 1.1 }}, RunOutput: true, Output: map[string]any{"a": float64(1.1)}},
		{Script: `a = "a"; b(a)`, Input: map[string]any{"b": func(c string) bool { return c == "a" }}, RunOutput: true, Output: map[string]any{"a": "a"}},

		{Script: `func b(c) { return c == nil }; b(a)`, Input: map[string]any{"a": nil}, RunOutput: true, Output: map[string]any{"a": nil}},
		{Script: `func b(c) { return c == true }; b(a)`, Input: map[string]any{"a": true}, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `func b(c) { return c == 1 }; b(a)`, Input: map[string]any{"a": int32(1)}, RunOutput: true, Output: map[string]any{"a": int32(1)}},
		{Script: `func b(c) { return c == 1 }; b(a)`, Input: map[string]any{"a": int64(1)}, RunOutput: true, Output: map[string]any{"a": int64(1)}},
		{Script: `func b(c) { return c == 1.1 }; b(a)`, Input: map[string]any{"a": float32(1.1)}, RunOutput: true, Output: map[string]any{"a": float32(1.1)}},
		{Script: `func b(c) { return c == 1.1 }; b(a)`, Input: map[string]any{"a": float64(1.1)}, RunOutput: true, Output: map[string]any{"a": float64(1.1)}},
		{Script: `func b(c) { return c == "a" }; b(a)`, Input: map[string]any{"a": "a"}, RunOutput: true, Output: map[string]any{"a": "a"}},

		{Script: `a = nil; func b(c) { return c == nil }; b(a)`, RunOutput: true, Output: map[string]any{"a": nil}},
		{Script: `a = true; func b(c) { return c == true }; b(a)`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = 1; func b(c) { return c == 1 }; b(a)`, Input: map[string]any{"a": int64(1)}, RunOutput: true, Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1.1; func b(c) { return c == 1.1 }; b(a)`, Input: map[string]any{"a": float64(1.1)}, RunOutput: true, Output: map[string]any{"a": float64(1.1)}},
		{Script: `a = "a"; func b(c) { return c == "a" }; b(a)`, Input: map[string]any{"a": "a"}, RunOutput: true, Output: map[string]any{"a": "a"}},

		{Script: `b(a[0])`, Input: map[string]any{"a": []any{nil}, "b": func(c any) bool { return c == nil }}, RunOutput: true, Output: map[string]any{"a": []any{nil}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{true}, "b": func(c any) bool { return c == true }}, RunOutput: true, Output: map[string]any{"a": []any{true}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{int32(1)}, "b": func(c any) bool { return c == int32(1) }}, RunOutput: true, Output: map[string]any{"a": []any{int32(1)}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{int64(1)}, "b": func(c any) bool { return c == int64(1) }}, RunOutput: true, Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{float32(1.1)}, "b": func(c any) bool { return c == float32(1.1) }}, RunOutput: true, Output: map[string]any{"a": []any{float32(1.1)}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{float64(1.1)}, "b": func(c any) bool { return c == float64(1.1) }}, RunOutput: true, Output: map[string]any{"a": []any{float64(1.1)}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{"a"}, "b": func(c any) bool { return c == "a" }}, RunOutput: true, Output: map[string]any{"a": []any{"a"}}},

		// TOFIX:
		//		{Script: `b(a)`,
		//			Input:     map[string]any{"a": []bool{true, false, true}, "b": func(c ...bool) bool { return c[len(c)-1] }},
		//			RunOutput: true, Output: map[string]any{"a": true}},

		{Script: `b(a[0])`, Input: map[string]any{"a": []any{true}, "b": func(c bool) bool { return c == true }}, RunOutput: true, Output: map[string]any{"a": []any{true}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{int32(1)}, "b": func(c int32) bool { return c == int32(1) }}, RunOutput: true, Output: map[string]any{"a": []any{int32(1)}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{int64(1)}, "b": func(c int64) bool { return c == int64(1) }}, RunOutput: true, Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{float32(1.1)}, "b": func(c float32) bool { return c == float32(1.1) }}, RunOutput: true, Output: map[string]any{"a": []any{float32(1.1)}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{float64(1.1)}, "b": func(c float64) bool { return c == float64(1.1) }}, RunOutput: true, Output: map[string]any{"a": []any{float64(1.1)}}},
		{Script: `b(a[0])`, Input: map[string]any{"a": []any{"a"}, "b": func(c string) bool { return c == "a" }}, RunOutput: true, Output: map[string]any{"a": []any{"a"}}},

		{Script: `a = [nil]; b(a[0])`, Input: map[string]any{"b": func(c any) bool { return c == nil }}, RunOutput: true, Output: map[string]any{"a": []any{nil}}},
		{Script: `a = [true]; b(a[0])`, Input: map[string]any{"b": func(c bool) bool { return c == true }}, RunOutput: true, Output: map[string]any{"a": []any{true}}},
		{Script: `a = [1]; b(a[0])`, Input: map[string]any{"b": func(c int64) bool { return c == int64(1) }}, RunOutput: true, Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `a = [1.1]; b(a[0])`, Input: map[string]any{"b": func(c float64) bool { return c == float64(1.1) }}, RunOutput: true, Output: map[string]any{"a": []any{float64(1.1)}}},
		{Script: `a = ["a"]; b(a[0])`, Input: map[string]any{"b": func(c string) bool { return c == "a" }}, RunOutput: true, Output: map[string]any{"a": []any{"a"}}},

		{Script: `a = [nil]; func b(c) { c == nil }; b(a[0])`, RunOutput: true, Output: map[string]any{"a": []any{nil}}},
		{Script: `a = [true]; func b(c) { c == true }; b(a[0])`, RunOutput: true, Output: map[string]any{"a": []any{true}}},
		{Script: `a = [1]; func b(c) { c == 1 }; b(a[0])`, RunOutput: true, Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `a = [1.1]; func b(c) { c == 1.1 }; b(a[0])`, RunOutput: true, Output: map[string]any{"a": []any{float64(1.1)}}},
		{Script: `a = ["a"]; func b(c) { c == "a" }; b(a[0])`, RunOutput: true, Output: map[string]any{"a": []any{"a"}}},

		{Script: `a = nil; b = func (d) { return d == nil }; b(a)`, RunOutput: true, Output: map[string]any{"a": nil}},
		{Script: `a = true; b = func (d) { return d == true }; b(a)`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = 1; b = func (d) { return d == 1 }; b(a)`, RunOutput: true, Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1.1; b = func (d) { return d == 1.1 }; b(a)`, RunOutput: true, Output: map[string]any{"a": float64(1.1)}},
		{Script: `a = "a"; b = func (d) { return d == "a" }; b(a)`, RunOutput: true, Output: map[string]any{"a": "a"}},

		{Script: `a = nil; b = func c(d) { return d == nil }; b(a)`, RunOutput: true, Output: map[string]any{"a": nil}},
		{Script: `a = true; b = func c(d) { return d == true }; b(a)`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = 1; b = func c(d) { return d == 1 }; b(a)`, RunOutput: true, Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1.1; b = func c(d) { return d == 1.1 }; b(a)`, RunOutput: true, Output: map[string]any{"a": float64(1.1)}},
		{Script: `a = "a"; b = func c(d) { return d == "a" }; b(a)`, RunOutput: true, Output: map[string]any{"a": "a"}},

		{Script: `a = nil; b = func c(d) { return d == nil }; c(a)`, RunOutput: true, Output: map[string]any{"a": nil}},
		{Script: `a = true; b = func c(d) { return d == true }; c(a)`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = 1; b = func c(d) { return d == 1 }; c(a)`, RunOutput: true, Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1.1; b = func c(d) { return d == 1.1 }; c(a)`, RunOutput: true, Output: map[string]any{"a": float64(1.1)}},
		{Script: `a = "a"; b = func c(d) { return d == "a" }; c(a)`, RunOutput: true, Output: map[string]any{"a": "a"}},

		{Script: `a = nil; func b() { return func c(d) { d == nil } }; e = b(); e(a)`, RunOutput: true, Output: map[string]any{"a": nil}},
		{Script: `a = true; func b() { return func c(d) { d == true } }; e = b(); e(a)`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = 1; func b() { return func c(d) { d == 1 } }; e = b(); e(a)`, RunOutput: true, Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1.1; func b() { return func c(d) { d == 1.1 } }; e = b(); e(a)`, RunOutput: true, Output: map[string]any{"a": float64(1.1)}},
		{Script: `a = "a"; func b() { return func c(d) { d == "a" } }; e = b(); e(a)`, RunOutput: true, Output: map[string]any{"a": "a"}},

		{Script: `a = func () { return nil }; func b(c) { return c() }; b(a)`, RunOutput: nil},
		{Script: `a = func () { return true }; func b(c) { return c() }; b(a)`, RunOutput: true},
		{Script: `a = func () { return 1 }; func b(c) { return c() }; b(a)`, RunOutput: int64(1)},
		{Script: `a = func () { return 1.1 }; func b(c) { return c() }; b(a)`, RunOutput: float64(1.1)},
		{Script: `a = func () { return "a" }; func b(c) { return c() }; b(a)`, RunOutput: "a"},

		{Script: `a = [nil]; func c(d) { return d[0] }; c(a)`, RunOutput: nil},
		{Script: `a = [true]; func c(d) { return d[0] }; c(a)`, RunOutput: true},
		{Script: `a = [1]; func c(d) { return d[0] }; c(a)`, RunOutput: int64(1)},
		{Script: `a = [1.1]; func c(d) { return d[0] }; c(a)`, RunOutput: float64(1.1)},
		{Script: `a = ["a"]; func c(d) { return d[0] }; c(a)`, RunOutput: "a"},

		{Script: `a = {"b": nil}; func c(d) { return d.b }; c(a)`, RunOutput: nil},
		{Script: `a = {"b": true}; func c(d) { return d.b }; c(a)`, RunOutput: true},
		{Script: `a = {"b": 1}; func c(d) { return d.b }; c(a)`, RunOutput: int64(1)},
		{Script: `a = {"b": 1.1}; func c(d) { return d.b }; c(a)`, RunOutput: float64(1.1)},
		{Script: `a = {"b": "a"}; func c(d) { return d.b }; c(a)`, RunOutput: "a"},

		{Script: `a = func() { return func(c) { return c + "c"} }; a()("a")`, RunOutput: "ac"},
		{Script: `a = func() { return func(c) { return c + "c"} }(); a("a")`, RunOutput: "ac"},
		{Script: `a = func() { return func(c) { return c + "c"} }()("a")`, RunOutput: "ac"},
		{Script: `func() { return func(c) { return c + "c"} }()("a")`, RunOutput: "ac"},

		{Script: `a = func(b) { return func() { return b + "c"} }; b = a("a"); b()`, RunOutput: "ac"},
		{Script: `a = func(b) { return func() { return b + "c"} }("a"); a()`, RunOutput: "ac"},
		{Script: `a = func(b) { return func() { return b + "c"} }("a")()`, RunOutput: "ac"},
		{Script: `func(b) { return func() { return b + "c"} }("a")()`, RunOutput: "ac"},

		{Script: `a = func(b) { return func(c) { return b[c] } }; b = a({"x": "x"}); b("x")`, RunOutput: "x"},
		{Script: `a = func(b) { return func(c) { return b[c] } }({"x": "x"}); a("x")`, RunOutput: "x"},
		{Script: `a = func(b) { return func(c) { return b[c] } }({"x": "x"})("x")`, RunOutput: "x"},
		{Script: `func(b) { return func(c) { return b[c] } }({"x": "x"})("x")`, RunOutput: "x"},

		{Script: `a = func(b) { return func(c) { return b[c] } }; x = {"y": "y"}; b = a(x); x = {"y": "y"}; b("y")`, RunOutput: "y"},
		{Script: `a = func(b) { return func(c) { return b[c] } }; x = {"y": "y"}; b = a(x); x.y = "z"; b("y")`, RunOutput: "z"},

		{Script: ` func a() { return "a" }; a()`, RunOutput: "a"},
		{Script: `a = func a() { return "a" }; a = func() { return "b" }; a()`, RunOutput: "b"},
		{Script: `a = "a.b"; func a() { return "a" }; a()`, RunOutput: "a"},

		{Script: `a = func() { b = "b"; return func() { b += "c" } }(); a()`, RunOutput: "bc"},
		{Script: `a = func() { b = "b"; return func() { b += "c"; return b} }(); a()`, RunOutput: "bc"},
		{Script: `a = func(b) { return func(c) { return func(d) { return d + "d" }(c) + "c" }(b) + "b" }("a")`, RunOutput: "adcb"},
		{Script: `a = func(b) { return "b" + func(c) { return "c" + func(d) { return "d" + d  }(c) }(b) }("a")`, RunOutput: "bcda"},
		{Script: `a = func(b) { return b + "b" }; a( func(c) { return c + "c" }("a") )`, RunOutput: "acb"},

		{Script: `a = func(x, y) { return func() { x(y) } }; b = a(func (z) { return z + "z" }, "b"); b()`, RunOutput: "bz"},

		{Script: `a = make(Time); a.IsZero()`, Types: map[string]any{"Time": time.Time{}}, RunOutput: true},

		{Script: `a = make(Buffer); n, err = a.WriteString("a"); if err != nil { return err }; n`, Types: map[string]any{"Buffer": bytes.Buffer{}}, RunOutput: 1},
		{Script: `a = make(Buffer); n, err = a.WriteString("a"); if err != nil { return err }; a.String()`, Types: map[string]any{"Buffer": bytes.Buffer{}}, RunOutput: "a"},

		{Script: `b = {}; c = a(b.c); c`, Input: map[string]any{"a": func(b string) bool {
			if b == "" {
				return true
			}
			return false
		}}, RunOutput: true},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestPointerFunctions(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	testFunctionPointer := func(b any) string {
		rv := reflect.ValueOf(b)
		if !rv.IsValid() {
			return "invalid"
		}
		if rv.Kind() != reflect.Pointer {
			return fmt.Sprintf("not ptr: " + rv.Kind().String())
		}
		if rv.IsNil() {
			return "IsNil"
		}
		if !rv.Elem().CanInterface() {
			return "cannot interface"
		}
		if rv.Elem().Interface() != int64(1) {
			return fmt.Sprintf("not 1: %v", rv.Elem().Interface())
		}
		if !rv.Elem().CanSet() {
			return "cannot set"
		}
		slice := reflect.MakeSlice(runner.InterfaceSliceType, 0, 1)
		value, _ := runner.MakeValue(runner.StringType)
		value.SetString("b")
		slice = reflect.Append(slice, value)
		rv.Elem().Set(slice)
		return "good"
	}
	tests := []Test{
		{Script: `b = 1; a(&b)`, Input: map[string]any{"a": testFunctionPointer}, RunOutput: "good", Output: map[string]any{"b": []any{"b"}}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestVariadicFunctions(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		// params Variadic arg !Variadic
		{Script: `func a(b...) { return b }; a()`, RunOutput: []any{}},
		{Script: `func a(b...) { return b }; a(true)`, RunOutput: []any{true}},
		{Script: `func a(b...) { return b }; a(true, true)`, RunOutput: []any{true, true}},
		{Script: `func a(b...) { return b }; a([true])`, RunOutput: []any{[]any{true}}},
		{Script: `func a(b...) { return b }; a([true, true])`, RunOutput: []any{[]any{true, true}}},
		{Script: `func a(b...) { return b }; a([true, true], [true, true])`, RunOutput: []any{[]any{true, true}, []any{true, true}}},

		// params Variadic arg !Variadic
		{Script: `func a(b, c...) { return c }; a()`, RunError: fmt.Errorf("function wants 2 arguments but received 0")},
		{Script: `func a(b, c...) { return c }; a(true)`, RunOutput: []any{}},
		{Script: `func a(b, c...) { return c }; a(true, true)`, RunOutput: []any{true}},
		{Script: `func a(b, c...) { return c }; a(true, true, true)`, RunOutput: []any{true, true}},
		{Script: `func a(b, c...) { return c }; a([true])`, RunOutput: []any{}},
		{Script: `func a(b, c...) { return c }; a([true], [true])`, RunOutput: []any{[]any{true}}},
		{Script: `func a(b, c...) { return c }; a([true], [true], [true])`, RunOutput: []any{[]any{true}, []any{true}}},
		{Script: `func a(b, c...) { return c }; a([true], [true, true], [true, true])`, RunOutput: []any{[]any{true, true}, []any{true, true}}},

		// params Variadic arg Variadic
		{Script: `func a(b...) { return b }; a([true]...)`, RunOutput: []any{true}},
		{Script: `func a(b...) { return b }; a([true, true]...)`, RunOutput: []any{true, true}},
		{Script: `func a(b...) { return b }; a(true, [true]...)`, RunError: fmt.Errorf("function wants 1 arguments but received 2")},

		// params Variadic arg Variadic
		{Script: `func a(b, c...) { return c }; a([true]...)`, RunOutput: []any{}},
		{Script: `func a(b, c...) { return c }; a([true, true]...)`, RunOutput: []any{}},
		{Script: `func a(b, c...) { return c }; a(true, [true]...)`, RunOutput: []any{true}},
		{Script: `func a(b, c...) { return c }; a(true, [true, true]...)`, RunOutput: []any{true, true}},

		// params !Variadic arg Variadic
		{Script: `func a() { return "a" }; a([true]...)`, RunOutput: "a"},
		{Script: `func a() { return "a" }; a(true, [true]...)`, RunOutput: "a"},
		{Script: `func a() { return "a" }; a(true, [true, true]...)`, RunOutput: "a"},

		// params !Variadic arg Variadic
		{Script: `func a(b) { return b }; a(true...)`, RunError: fmt.Errorf("call is variadic but last parameter is of type bool")},
		{Script: `func a(b) { return b }; a([true]...)`, RunOutput: true},
		{Script: `func a(b) { return b }; a(true, false...)`, RunError: fmt.Errorf("function wants 1 arguments but received 2")},
		{Script: `func a(b) { return b }; a(true, [1]...)`, RunError: fmt.Errorf("function wants 1 arguments but received 2")},
		{Script: `func a(b) { return b }; a(true, [1, 2]...)`, RunError: fmt.Errorf("function wants 1 arguments but received 2")},
		{Script: `func a(b) { return b }; a([true, 1]...)`, RunOutput: true},
		{Script: `func a(b) { return b }; a([true, 1, 2]...)`, RunOutput: true},

		// params !Variadic arg Variadi
		{Script: `func a(b, c) { return c }; a(false...)`, RunError: fmt.Errorf("call is variadic but last parameter is of type bool")},
		{Script: `func a(b, c) { return c }; a([1]...)`, RunError: fmt.Errorf("function wants 2 arguments but received 1")},
		{Script: `func a(b, c) { return c }; a(1, true...)`, RunError: fmt.Errorf("call is variadic but last parameter is of type bool")},
		{Script: `func a(b, c) { return c }; a(1, [true]...)`, RunOutput: true},
		{Script: `func a(b, c) { return c }; a([1, true]...)`, RunOutput: true},
		{Script: `func a(b, c) { return c }; a(1, true...)`, RunError: fmt.Errorf("call is variadic but last parameter is of type bool")},
		{Script: `func a(b, c) { return c }; a(1, [true]...)`, RunOutput: true},
		{Script: `func a(b, c) { return c }; a(1, true, false...)`, RunError: fmt.Errorf("function wants 2 arguments but received 3")},
		{Script: `func a(b, c) { return c }; a(1, true, [2]...)`, RunError: fmt.Errorf("function wants 2 arguments but received 3")},
		{Script: `func a(b, c) { return c }; a(1, [true, 2]...)`, RunOutput: true},
		{Script: `func a(b, c) { return c }; a([1, true, 2]...)`, RunOutput: true},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestFunctionsInArraysAndMaps(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = [func () { return nil }]; a[0]()`, RunOutput: nil},
		{Script: `a = [func () { return true }]; a[0]()`, RunOutput: true},
		{Script: `a = [func () { return 1 }]; a[0]()`, RunOutput: int64(1)},
		{Script: `a = [func () { return 1.1 }]; a[0]()`, RunOutput: float64(1.1)},
		{Script: `a = [func () { return "a" }]; a[0]()`, RunOutput: "a"},

		{Script: `a = [func () { return nil }]; b = a[0]; b()`, RunOutput: nil},
		{Script: `a = [func () { return true }]; b = a[0]; b()`, RunOutput: true},
		{Script: `a = [func () { return 1 }]; b = a[0]; b()`, RunOutput: int64(1)},
		{Script: `a = [func () { return 1.1 }]; b = a[0]; b()`, RunOutput: float64(1.1)},
		{Script: `a = [func () { return "a" }]; b = a[0]; b()`, RunOutput: "a"},

		{Script: `a = [func () { return nil}]; func b(c) { return c() }; b(a[0])`, RunOutput: nil},
		{Script: `a = [func () { return true}]; func b(c) { return c() }; b(a[0])`, RunOutput: true},
		{Script: `a = [func () { return 1}]; func b(c) { return c() }; b(a[0])`, RunOutput: int64(1)},
		{Script: `a = [func () { return 1.1}]; func b(c) { return c() }; b(a[0])`, RunOutput: float64(1.1)},
		{Script: `a = [func () { return "a"}]; func b(c) { return c() }; b(a[0])`, RunOutput: "a"},

		{Script: `a = {"b": func () { return nil }}; a["b"]()`, RunOutput: nil},
		{Script: `a = {"b": func () { return true }}; a["b"]()`, RunOutput: true},
		{Script: `a = {"b": func () { return 1 }}; a["b"]()`, RunOutput: int64(1)},
		{Script: `a = {"b": func () { return 1.1 }}; a["b"]()`, RunOutput: float64(1.1)},
		{Script: `a = {"b": func () { return "a" }}; a["b"]()`, RunOutput: "a"},

		{Script: `a = {"b": func () { return nil }}; a.b()`, RunOutput: nil},
		{Script: `a = {"b": func () { return true }}; a.b()`, RunOutput: true},
		{Script: `a = {"b": func () { return 1 }}; a.b()`, RunOutput: int64(1)},
		{Script: `a = {"b": func () { return 1.1 }}; a.b()`, RunOutput: float64(1.1)},
		{Script: `a = {"b": func () { return "a" }}; a.b()`, RunOutput: "a"},

		{Script: `a = {"b": func () { return nil }}; func c(d) { return d() }; c(a.b)`, RunOutput: nil},
		{Script: `a = {"b": func () { return true }}; func c(d) { return d() }; c(a.b)`, RunOutput: true},
		{Script: `a = {"b": func () { return 1 }}; func c(d) { return d() }; c(a.b)`, RunOutput: int64(1)},
		{Script: `a = {"b": func () { return 1.1 }}; func c(d) { return d() }; c(a.b)`, RunOutput: float64(1.1)},
		{Script: `a = {"b": func () { return "a" }}; func c(d) { return d() }; c(a.b)`, RunOutput: "a"},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestFunctionConvertions(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `b = func(c){ return c }; a("x", b)`, Input: map[string]any{"a": func(b string, c func(string) string) string { return c(b) }}, RunOutput: "x"},
		{Script: `b = make(struct1); b.A = func (c, d) { return c == d }; b.A(2, 2)`, Types: map[string]any{"struct1": &struct {
			A func(int, int) bool
		}{}},
			RunOutput: true},
		{Script: `b = 1; a(&b)`, Input: map[string]any{"a": func(b *int64) { *b = int64(2) }}, Output: map[string]any{"b": int64(2)}},
		{Script: `b = func(){ return true, 1, 2, 3.3, 4.4, "5" }; c, d, e, f, g, h = a(b); c`, Input: map[string]any{"a": func(b func() (bool, int32, int64, float32, float64, string)) (bool, int32, int64, float32, float64, string) {
			return b()
		}}, RunOutput: true, Output: map[string]any{"c": true, "d": int32(1), "e": int64(2), "f": float32(3.3), "g": float64(4.4), "h": "5"}},

		// slice inteface unable to convert to int
		{Script: `b = [1, 2.2, "3"]; a(b)`, Input: map[string]any{"a": func(b []int) int { return len(b) }}, RunError: fmt.Errorf("function wants argument type []int but received type []interface {}"), Output: map[string]any{"b": []any{int64(1), float64(2.2), "3"}}},
		// slice no sub convertible convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b []int) int { return len(b) }, "b": []int64{1}}, RunOutput: int(1), Output: map[string]any{"b": []int64{1}}},
		// array no sub convertible convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b [2]int) int { return len(b) }, "b": [2]int64{1, 2}}, RunOutput: int(2), Output: map[string]any{"b": [2]int64{1, 2}}},
		// slice no sub to interface convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b []any) int { return len(b) }, "b": []int64{1}}, RunOutput: int(1), Output: map[string]any{"b": []int64{1}}},
		// array no sub to interface convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b [2]any) int { return len(b) }, "b": [2]int64{1, 2}}, RunOutput: int(2), Output: map[string]any{"b": [2]int64{1, 2}}},
		// slice no sub from interface convertion
		{Script: `b = [1]; a(b)`, Input: map[string]any{"a": func(b []int) int { return len(b) }}, RunOutput: int(1), Output: map[string]any{"b": []any{int64(1)}}},
		// array no sub from interface convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b [2]int) int { return len(b) }, "b": [2]any{1, 2}}, RunOutput: int(2), Output: map[string]any{"b": [2]any{1, 2}}},

		// slice sub mismatch
		{Script: `a(b)`, Input: map[string]any{"a": func(b []int) int { return len(b) }, "b": [][]int64{[]int64{1, 2}}}, RunError: fmt.Errorf("function wants argument type []int but received type [][]int64"), Output: map[string]any{"b": [][]int64{[]int64{1, 2}}}},
		// array sub mismatch
		{Script: `a(b)`, Input: map[string]any{"a": func(b [2]int) int { return len(b) }, "b": [1][2]int64{[2]int64{1, 2}}}, RunError: fmt.Errorf("function wants argument type [2]int but received type [1][2]int64"), Output: map[string]any{"b": [1][2]int64{[2]int64{1, 2}}}},

		// slice with sub int64 to int convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b [][]int) int { return len(b) }, "b": [][]int64{[]int64{1, 2}, []int64{3, 4}}}, RunOutput: int(2), Output: map[string]any{"b": [][]int64{[]int64{1, 2}, []int64{3, 4}}}},
		// array with sub int64 to int convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b [][]int) int { return len(b) }, "b": [2][2]int64{[2]int64{1, 2}, [2]int64{3, 4}}}, RunOutput: int(2), Output: map[string]any{"b": [2][2]int64{[2]int64{1, 2}, [2]int64{3, 4}}}},
		// slice with sub interface to int convertion
		{Script: `b = [[1, 2], [3, 4]]; a(b)`, Input: map[string]any{"a": func(b [][]int) int { return len(b) }}, RunOutput: int(2), Output: map[string]any{"b": []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}}},
		// slice with sub interface to int convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b [][]int) int { return len(b) }, "b": [][]any{[]any{int64(1), int32(2)}, []any{float64(3.3), float32(4.4)}}}, RunOutput: int(2), Output: map[string]any{"b": [][]any{[]any{int64(1), int32(2)}, []any{float64(3.3), float32(4.4)}}}},
		// array with sub interface to int convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b [][]int) int { return len(b) }, "b": [2][2]any{[2]any{1, 2}, [2]any{3, 4}}}, RunOutput: int(2), Output: map[string]any{"b": [2][2]any{[2]any{1, 2}, [2]any{3, 4}}}},
		// slice with single interface to double interface
		{Script: `b = [[1, 2], [3, 4]]; a(b)`, Input: map[string]any{"a": func(b [][]any) int { return len(b) }}, RunOutput: int(2), Output: map[string]any{"b": []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}}},
		// slice with sub int64 to double interface convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b [][]any) int { return len(b) }, "b": [][]int64{[]int64{1, 2}, []int64{3, 4}}}, RunOutput: int(2), Output: map[string]any{"b": [][]int64{[]int64{1, 2}, []int64{3, 4}}}},
		// array with sub int64 to double interface convertion
		{Script: `a(b)`, Input: map[string]any{"a": func(b [][]any) int { return len(b) }, "b": [2][2]int64{[2]int64{1, 2}, [2]int64{3, 4}}}, RunOutput: int(2), Output: map[string]any{"b": [2][2]int64{[2]int64{1, 2}, [2]int64{3, 4}}}},

		// TOFIX: not able to change pointer value
		// {Script: `b = 1; c = &b; a(c); *c`, Input: map[string]any{"a": func(b *int64) { *b = int64(2) }}, RunOutput: int64(2), Output: map[string]any{"b": int64(2)}},

		// map [interface]interface to [interface]interface
		{Script: `b = {nil:nil}; c = nil; d = a(b, c)`, Input: map[string]any{"a": func(b map[any]any, c any) any { return b[c] }}, RunOutput: nil, Output: map[string]any{"b": map[any]any{nil: nil}, "c": nil}},
		{Script: `b = {true:true}; c = true; d = a(b, c)`, Input: map[string]any{"a": func(b map[any]any, c any) any { return b[c] }}, RunOutput: true, Output: map[string]any{"b": map[any]any{true: true}, "c": true}},
		{Script: `b = {1:2}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[any]any, c any) any { return b[c] }}, RunOutput: int64(2), Output: map[string]any{"b": map[any]any{int64(1): int64(2)}, "c": int64(1)}},
		{Script: `b = {1.1:2.2}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[any]any, c any) any { return b[c] }}, RunOutput: float64(2.2), Output: map[string]any{"b": map[any]any{float64(1.1): float64(2.2)}, "c": float64(1.1)}},
		{Script: `b = {"a":"b"}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[any]any, c any) any { return b[c] }}, RunOutput: "b", Output: map[string]any{"b": map[any]any{"a": "b"}, "c": "a"}},

		// map [interface]interface to [bool]interface
		{Script: `b = {"a":"b"}; c = true; d = a(b, c)`, Input: map[string]any{"a": func(b map[bool]any, c bool) any { return b[c] }}, RunError: fmt.Errorf("function wants argument type map[bool]interface {} but received type map[interface {}]interface {}"), Output: map[string]any{"b": map[any]any{"a": "b"}, "c": true}},
		{Script: `b = {"a":"b"}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[bool]any, c bool) any { return b[c] }}, RunError: fmt.Errorf("function wants argument type map[bool]interface {} but received type map[interface {}]interface {}"), Output: map[string]any{"b": map[any]any{"a": "b"}, "c": "a"}},
		{Script: `b = {true:"b"}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[bool]any, c bool) any { return b[c] }}, RunError: fmt.Errorf("function wants argument type bool but received type string"), Output: map[string]any{"b": map[any]any{true: "b"}, "c": "a"}},
		{Script: `b = {true:nil}; c = true; d = a(b, c)`, Input: map[string]any{"a": func(b map[bool]any, c bool) any { return b[c] }}, RunOutput: nil, Output: map[string]any{"b": map[any]any{true: nil}, "c": true}},
		{Script: `b = {true:true}; c = true; d = a(b, c)`, Input: map[string]any{"a": func(b map[bool]any, c bool) any { return b[c] }}, RunOutput: true, Output: map[string]any{"b": map[any]any{true: true}, "c": true}},
		{Script: `b = {true:2}; c = true; d = a(b, c)`, Input: map[string]any{"a": func(b map[bool]any, c bool) any { return b[c] }}, RunOutput: int64(2), Output: map[string]any{"b": map[any]any{true: int64(2)}, "c": true}},
		{Script: `b = {true:2.2}; c = true; d = a(b, c)`, Input: map[string]any{"a": func(b map[bool]any, c bool) any { return b[c] }}, RunOutput: float64(2.2), Output: map[string]any{"b": map[any]any{true: float64(2.2)}, "c": true}},
		{Script: `b = {true:"b"}; c = true; d = a(b, c)`, Input: map[string]any{"a": func(b map[bool]any, c bool) any { return b[c] }}, RunOutput: "b", Output: map[string]any{"b": map[any]any{true: "b"}, "c": true}},

		// map [interface]interface to [int32]interface
		{Script: `b = {1:nil}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int32]any, c int32) any { return b[c] }}, RunOutput: nil, Output: map[string]any{"b": map[any]any{int64(1): nil}, "c": int64(1)}},
		{Script: `b = {1:true}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int32]any, c int32) any { return b[c] }}, RunOutput: true, Output: map[string]any{"b": map[any]any{int64(1): true}, "c": int64(1)}},
		{Script: `b = {1:2}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int32]any, c int32) any { return b[c] }}, RunOutput: int64(2), Output: map[string]any{"b": map[any]any{int64(1): int64(2)}, "c": int64(1)}},
		{Script: `b = {1:2.2}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int32]any, c int32) any { return b[c] }}, RunOutput: float64(2.2), Output: map[string]any{"b": map[any]any{int64(1): float64(2.2)}, "c": int64(1)}},
		{Script: `b = {1:"b"}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int32]any, c int32) any { return b[c] }}, RunOutput: "b", Output: map[string]any{"b": map[any]any{int64(1): "b"}, "c": int64(1)}},

		// map [interface]interface to [int64]interface
		{Script: `b = {1:nil}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int64]any, c int64) any { return b[c] }}, RunOutput: nil, Output: map[string]any{"b": map[any]any{int64(1): nil}, "c": int64(1)}},
		{Script: `b = {1:true}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int64]any, c int64) any { return b[c] }}, RunOutput: true, Output: map[string]any{"b": map[any]any{int64(1): true}, "c": int64(1)}},
		{Script: `b = {1:2}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int64]any, c int64) any { return b[c] }}, RunOutput: int64(2), Output: map[string]any{"b": map[any]any{int64(1): int64(2)}, "c": int64(1)}},
		{Script: `b = {1:2.2}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int64]any, c int64) any { return b[c] }}, RunOutput: float64(2.2), Output: map[string]any{"b": map[any]any{int64(1): float64(2.2)}, "c": int64(1)}},
		{Script: `b = {1:"b"}; c = 1; d = a(b, c)`, Input: map[string]any{"a": func(b map[int64]any, c int64) any { return b[c] }}, RunOutput: "b", Output: map[string]any{"b": map[any]any{int64(1): "b"}, "c": int64(1)}},

		// map [interface]interface to [float32]interface
		{Script: `b = {1.1:nil}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float32]any, c float32) any { return b[c] }}, RunOutput: nil, Output: map[string]any{"b": map[any]any{float64(1.1): nil}, "c": float64(1.1)}},
		{Script: `b = {1.1:true}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float32]any, c float32) any { return b[c] }}, RunOutput: true, Output: map[string]any{"b": map[any]any{float64(1.1): true}, "c": float64(1.1)}},
		{Script: `b = {1.1:2}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float32]any, c float32) any { return b[c] }}, RunOutput: int64(2), Output: map[string]any{"b": map[any]any{float64(1.1): int64(2)}, "c": float64(1.1)}},
		{Script: `b = {1.1:2.2}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float32]any, c float32) any { return b[c] }}, RunOutput: float64(2.2), Output: map[string]any{"b": map[any]any{float64(1.1): float64(2.2)}, "c": float64(1.1)}},
		{Script: `b = {1.1:"b"}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float32]any, c float32) any { return b[c] }}, RunOutput: "b", Output: map[string]any{"b": map[any]any{float64(1.1): "b"}, "c": float64(1.1)}},

		// map [interface]interface to [float64]interface
		{Script: `b = {1.1:nil}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float64]any, c float64) any { return b[c] }}, RunOutput: nil, Output: map[string]any{"b": map[any]any{float64(1.1): nil}, "c": float64(1.1)}},
		{Script: `b = {1.1:true}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float64]any, c float64) any { return b[c] }}, RunOutput: true, Output: map[string]any{"b": map[any]any{float64(1.1): true}, "c": float64(1.1)}},
		{Script: `b = {1.1:2}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float64]any, c float64) any { return b[c] }}, RunOutput: int64(2), Output: map[string]any{"b": map[any]any{float64(1.1): int64(2)}, "c": float64(1.1)}},
		{Script: `b = {1.1:2.2}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float64]any, c float64) any { return b[c] }}, RunOutput: float64(2.2), Output: map[string]any{"b": map[any]any{float64(1.1): float64(2.2)}, "c": float64(1.1)}},
		{Script: `b = {1.1:"b"}; c = 1.1; d = a(b, c)`, Input: map[string]any{"a": func(b map[float64]any, c float64) any { return b[c] }}, RunOutput: "b", Output: map[string]any{"b": map[any]any{float64(1.1): "b"}, "c": float64(1.1)}},

		// map [interface]interface to [string]interface
		{Script: `b = {"a":nil}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]any, c string) any { return b[c] }}, RunOutput: nil, Output: map[string]any{"b": map[any]any{"a": nil}, "c": "a"}},
		{Script: `b = {"a":true}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]any, c string) any { return b[c] }}, RunOutput: true, Output: map[string]any{"b": map[any]any{"a": true}, "c": "a"}},
		{Script: `b = {"a":2}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]any, c string) any { return b[c] }}, RunOutput: int64(2), Output: map[string]any{"b": map[any]any{"a": int64(2)}, "c": "a"}},
		{Script: `b = {"a":2.2}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]any, c string) any { return b[c] }}, RunOutput: float64(2.2), Output: map[string]any{"b": map[any]any{"a": float64(2.2)}, "c": "a"}},
		{Script: `b = {"a":"b"}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]any, c string) any { return b[c] }}, RunOutput: "b", Output: map[string]any{"b": map[any]any{"a": "b"}, "c": "a"}},

		// map [interface]interface to [string]X
		{Script: `b = {"a":"b"}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]bool, c string) bool { return b[c] }}, RunError: fmt.Errorf("function wants argument type map[string]bool but received type map[interface {}]interface {}"), Output: map[string]any{"b": map[any]any{"a": "b"}, "c": "a"}},
		{Script: `b = {"a":true}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]bool, c string) bool { return b[c] }}, RunOutput: true, Output: map[string]any{"b": map[any]any{"a": true}, "c": "a"}},
		{Script: `b = {"a":1}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]int32, c string) int32 { return b[c] }}, RunOutput: int32(1), Output: map[string]any{"b": map[any]any{"a": int64(1)}, "c": "a"}},
		{Script: `b = {"a":1}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]int64, c string) int64 { return b[c] }}, RunOutput: int64(1), Output: map[string]any{"b": map[any]any{"a": int64(1)}, "c": "a"}},
		{Script: `b = {"a":1.1}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]float32, c string) float32 { return b[c] }}, RunOutput: float32(1.1), Output: map[string]any{"b": map[any]any{"a": float64(1.1)}, "c": "a"}},
		{Script: `b = {"a":1.1}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]float64, c string) float64 { return b[c] }}, RunOutput: float64(1.1), Output: map[string]any{"b": map[any]any{"a": float64(1.1)}, "c": "a"}},
		{Script: `b = {"a":"b"}; c = "a"; d = a(b, c)`, Input: map[string]any{"a": func(b map[string]string, c string) string { return b[c] }}, RunOutput: "b", Output: map[string]any{"b": map[any]any{"a": "b"}, "c": "a"}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}

	_ = os.Unsetenv("ANKO_DEBUG")
	tests = []Test{
		{Script: `c = a(b)`,
			Input: map[string]any{"a": func(b func() bool) bool {
				return b()
			}, "b": func(c func(bool)) { c(true) }}, RunError: fmt.Errorf("function wants argument type func() bool but received type func(func(bool))")},
		{Script: `b = func(){ return 1++ }; c = a(b)`,
			Input: map[string]any{"a": func(b func() bool) bool {
				return b()
			}}, RunError: fmt.Errorf("invalid operation")},
		{Script: `b = func(){ return true }; c = a(b)`,
			Input: map[string]any{"a": func(b func() string) string {
				return b()
			}}, RunError: fmt.Errorf("function wants return type string but received type bool")},
		{Script: `b = func(){ return true }; c = a(b)`,
			Input: map[string]any{"a": func(b func() (bool, string)) (bool, string) {
				return b()
			}}, RunError: fmt.Errorf("function wants 2 return values but received bool")},
		{Script: `b = func(){ return true, 1 }; c = a(b)`,
			Input: map[string]any{"a": func(b func() (bool, int64, string)) (bool, int64, string) {
				return b()
			}}, RunError: fmt.Errorf("function wants 3 return values but received 2 values")},
		{Script: `b = func(){ return "1", true }; c = a(b)`,
			Input: map[string]any{"a": func(b func() (bool, string)) (bool, string) {
				return b()
			}}, RunError: fmt.Errorf("function wants return type bool but received type string")},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestVariadicFunctionConvertions(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	testSumFunc := func(nums ...int64) int64 {
		var total int64
		for _, num := range nums {
			total += num
		}
		return total
	}
	tests := []Test{
		// params Variadic arg !Variadic
		{Script: `a(true)`, Input: map[string]any{"a": func(b ...any) []any { return b }}, RunOutput: []any{true}},

		{Script: `a()`, Input: map[string]any{"a": testSumFunc}, RunOutput: int64(0)},
		{Script: `a(1)`, Input: map[string]any{"a": testSumFunc}, RunOutput: int64(1)},
		{Script: `a(1, 2)`, Input: map[string]any{"a": testSumFunc}, RunOutput: int64(3)},
		{Script: `a(1, 2, 3)`, Input: map[string]any{"a": testSumFunc}, RunOutput: int64(6)},

		// TODO: add more tests
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestLen(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `len(1++)`, RunError: fmt.Errorf("invalid operation")},
		{Script: `len(true)`, RunError: fmt.Errorf("type bool does not support len operation")},

		{Script: `a = ""; len(a)`, RunOutput: int64(0)},
		{Script: `a = "test"; len(a)`, RunOutput: int64(4)},
		{Script: `a = []; len(a)`, RunOutput: int64(0)},
		{Script: `a = [nil]; len(a)`, RunOutput: int64(1)},
		{Script: `a = [true]; len(a)`, RunOutput: int64(1)},
		{Script: `a = ["test"]; len(a)`, RunOutput: int64(1)},
		{Script: `a = [1]; len(a)`, RunOutput: int64(1)},
		{Script: `a = [1.1]; len(a)`, RunOutput: int64(1)},

		{Script: `a = [[]]; len(a)`, RunOutput: int64(1)},
		{Script: `a = [[nil]]; len(a)`, RunOutput: int64(1)},
		{Script: `a = [[true]]; len(a)`, RunOutput: int64(1)},
		{Script: `a = [["test"]]; len(a)`, RunOutput: int64(1)},
		{Script: `a = [[1]]; len(a)`, RunOutput: int64(1)},
		{Script: `a = [[1.1]]; len(a)`, RunOutput: int64(1)},

		{Script: `a = [[]]; len(a[0])`, RunOutput: int64(0)},
		{Script: `a = [[nil]]; len(a[0])`, RunOutput: int64(1)},
		{Script: `a = [[true]]; len(a[0])`, RunOutput: int64(1)},
		{Script: `a = [["test"]]; len(a[0])`, RunOutput: int64(1)},
		{Script: `a = [[1]]; len(a[0])`, RunOutput: int64(1)},
		{Script: `a = [[1.1]]; len(a[0])`, RunOutput: int64(1)},

		{Script: `len(a)`, Input: map[string]any{"a": "a"}, RunOutput: int64(1), Output: map[string]any{"a": "a"}},
		{Script: `len(a)`, Input: map[string]any{"a": map[string]any{}}, RunOutput: int64(0), Output: map[string]any{"a": map[string]any{}}},
		{Script: `len(a)`, Input: map[string]any{"a": map[string]any{"test": "test"}}, RunOutput: int64(1), Output: map[string]any{"a": map[string]any{"test": "test"}}},
		{Script: `len(a["test"])`, Input: map[string]any{"a": map[string]any{"test": "test"}}, RunOutput: int64(4), Output: map[string]any{"a": map[string]any{"test": "test"}}},

		{Script: `len(a)`, Input: map[string]any{"a": []any{}}, RunOutput: int64(0), Output: map[string]any{"a": []any{}}},
		{Script: `len(a)`, Input: map[string]any{"a": []any{nil}}, RunOutput: int64(1), Output: map[string]any{"a": []any{nil}}},
		{Script: `len(a)`, Input: map[string]any{"a": []any{true}}, RunOutput: int64(1), Output: map[string]any{"a": []any{true}}},
		{Script: `len(a)`, Input: map[string]any{"a": []any{int32(1)}}, RunOutput: int64(1), Output: map[string]any{"a": []any{int32(1)}}},
		{Script: `len(a)`, Input: map[string]any{"a": []any{int64(1)}}, RunOutput: int64(1), Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `len(a)`, Input: map[string]any{"a": []any{float32(1.1)}}, RunOutput: int64(1), Output: map[string]any{"a": []any{float32(1.1)}}},
		{Script: `len(a)`, Input: map[string]any{"a": []any{float64(1.1)}}, RunOutput: int64(1), Output: map[string]any{"a": []any{float64(1.1)}}},
		{Script: `len(a)`, Input: map[string]any{"a": []any{"a"}}, RunOutput: int64(1), Output: map[string]any{"a": []any{"a"}}},

		{Script: `len(a[0])`, Input: map[string]any{"a": []any{"test"}}, RunOutput: int64(4), Output: map[string]any{"a": []any{"test"}}},

		{Script: `len(a)`, Input: map[string]any{"a": [][]any{}}, RunOutput: int64(0), Output: map[string]any{"a": [][]any{}}},
		{Script: `len(a)`, Input: map[string]any{"a": [][]any{nil}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{nil}}},
		{Script: `len(a)`, Input: map[string]any{"a": [][]any{{nil}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{nil}}}},
		{Script: `len(a)`, Input: map[string]any{"a": [][]any{{true}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{true}}}},
		{Script: `len(a)`, Input: map[string]any{"a": [][]any{{int32(1)}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{int32(1)}}}},
		{Script: `len(a)`, Input: map[string]any{"a": [][]any{{int64(1)}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{int64(1)}}}},
		{Script: `len(a)`, Input: map[string]any{"a": [][]any{{float32(1.1)}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{float32(1.1)}}}},
		{Script: `len(a)`, Input: map[string]any{"a": [][]any{{float64(1.1)}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{float64(1.1)}}}},
		{Script: `len(a)`, Input: map[string]any{"a": [][]any{{"a"}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{"a"}}}},

		{Script: `len(a[0])`, Input: map[string]any{"a": [][]any{nil}}, RunOutput: int64(0), Output: map[string]any{"a": [][]any{nil}}},
		{Script: `len(a[0])`, Input: map[string]any{"a": [][]any{{nil}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{nil}}}},
		{Script: `len(a[0])`, Input: map[string]any{"a": [][]any{{true}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{true}}}},
		{Script: `len(a[0])`, Input: map[string]any{"a": [][]any{{int32(1)}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{int32(1)}}}},
		{Script: `len(a[0])`, Input: map[string]any{"a": [][]any{{int64(1)}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{int64(1)}}}},
		{Script: `len(a[0])`, Input: map[string]any{"a": [][]any{{float32(1.1)}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{float32(1.1)}}}},
		{Script: `len(a[0])`, Input: map[string]any{"a": [][]any{{float64(1.1)}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{float64(1.1)}}}},
		{Script: `len(a[0])`, Input: map[string]any{"a": [][]any{{"a"}}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{"a"}}}},

		{Script: `len(a[0][0])`, Input: map[string]any{"a": [][]any{{"test"}}}, RunOutput: int64(4), Output: map[string]any{"a": [][]any{{"test"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestCallFunctionWithVararg(t *testing.T) {
	v := New(nil)
	err := v.Define("X", func(args ...string) []string {
		return args
	})
	if err != nil {
		t.Errorf("Define error: %v", err)
	}
	want := []string{"foo", "bar", "baz"}
	err = v.Define("a", want)
	if err != nil {
		t.Errorf("Define error: %v", err)
	}
	got, err := v.Executor(nil).Run(nil, `X(a...)`)
	if err != nil {
		t.Errorf("execute error - received %#v - expected: %#v", err, context.Canceled)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("execute error - received %#v - expected: %#v", got, want)
	}
}

func TestDeferFunction(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `
a = []
func add(n) { a += n }
func aFunc() {
  var i = 0
  defer add(i)
  i++
  defer add(i)
  i++
  defer add(i)
  i++
  defer add(i)
  i++
  defer add(i)
}
aFunc()`,
			Output: map[string]any{"a": []any{int64(4), int64(3), int64(2), int64(1), int64(0)}}},
		{Script: `
a = 0
func aFunc() {
  defer func(){
    a = 123
  }()
}
aFunc()`,
			Output: map[string]any{"a": int64(123)}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}
