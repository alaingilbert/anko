package vm

import (
	"context"
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/packages"
	"github.com/alaingilbert/anko/pkg/utils"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"github.com/alaingilbert/anko/pkg/vm/runner"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alaingilbert/anko/pkg/parser"
)

var (
	testVarValue    = reflect.Value{}
	testVarValueP   = &reflect.Value{}
	testVarBool     = true
	testVarBoolP    = &testVarBool
	testVarInt32    = int32(1)
	testVarInt32P   = &testVarInt32
	testVarInt64    = int64(1)
	testVarInt64P   = &testVarInt64
	testVarFloat32  = float32(1)
	testVarFloat32P = &testVarFloat32
	testVarFloat64  = float64(1)
	testVarFloat64P = &testVarFloat64
	testVarString   = "a"
	testVarStringP  = &testVarString
	testVarFunc     = func() int64 { return 1 }
	testVarFuncP    = &testVarFunc

	testVarValueBool    = reflect.ValueOf(true)
	testVarValueInt32   = reflect.ValueOf(int32(1))
	testVarValueInt64   = reflect.ValueOf(int64(1))
	testVarValueFloat32 = reflect.ValueOf(float32(1.1))
	testVarValueFloat64 = reflect.ValueOf(float64(1.1))
	testVarValueString  = reflect.ValueOf("a")
)

func TestNumbers(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: ``},
		{Script: `;`},
		{Script: `
`},
		{Script: `
1
`, RunOutput: int64(1)},

		{Script: `1..1`, RunError: fmt.Errorf(`strconv.ParseFloat: parsing "1..1": invalid syntax`), Name: ""},
		{Script: `0x1g`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `9223372036854775808`, RunError: fmt.Errorf(`strconv.ParseInt: parsing "9223372036854775808": value out of range`), Name: ""},

		{Script: `1`, RunOutput: int64(1), Name: ""},
		{Script: `-1`, RunOutput: int64(-1), Name: ""},
		{Script: `9223372036854775807`, RunOutput: int64(9223372036854775807), Name: ""},
		{Script: `-9223372036854775807`, RunOutput: int64(-9223372036854775807), Name: ""},
		{Script: `1.1`, RunOutput: float64(1.1), Name: ""},
		{Script: `-1.1`, RunOutput: float64(-1.1), Name: ""},
		{Script: `1e1`, RunOutput: float64(10), Name: ""},
		{Script: `-1e1`, RunOutput: float64(-10), Name: ""},
		{Script: `1e-1`, RunOutput: float64(0.1), Name: ""},
		{Script: `-1e-1`, RunOutput: float64(-0.1), Name: ""},
		{Script: `0x1`, RunOutput: int64(1), Name: ""},
		{Script: `0xc`, RunOutput: int64(12), Name: ""},
		// TOFIX:
		{Script: `0xe`, RunError: fmt.Errorf(`strconv.ParseFloat: parsing "0xe": invalid syntax`), Name: ""},
		{Script: `0xf`, RunOutput: int64(15), Name: ""},
		{Script: `-0x1`, RunOutput: int64(-1), Name: ""},
		{Script: `-0xc`, RunOutput: int64(-12), Name: ""},
		{Script: `-0xf`, RunOutput: int64(-15), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestStrings(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a`, Input: map[string]any{"a": 'a'}, RunOutput: 'a', Output: map[string]any{"a": 'a'}, Name: ""},
		{Script: `a.b`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support member operation"), Output: map[string]any{"a": 'a'}, Name: ""},
		{Script: `a[0]`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support index operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}, Name: ""},
		{Script: `a[0:1]`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support slice operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}, Name: ""},

		{Script: `a.b = "a"`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support member operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}, Name: ""},
		{Script: `a[0] = "a"`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support index operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}, Name: ""},
		{Script: `a[0:1] = "a"`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support slice operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}, Name: ""},

		{Script: `a.b = "a"`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("type string does not support member operation"), Output: map[string]any{"a": "test"}, Name: ""},
		{Script: `a[0:1] = "a"`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("type string does not support slice operation for assignment"), Output: map[string]any{"a": "test"}, Name: ""},

		{Script: `a`, Input: map[string]any{"a": "test"}, RunOutput: "test", Output: map[string]any{"a": "test"}, Name: ""},
		{Script: `a["a"]`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"a": "test"}, Name: ""},
		{Script: `a[0]`, Input: map[string]any{"a": ""}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": ""}, Name: ""},
		{Script: `a[-1]`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test"}, Name: ""},
		{Script: `a[0]`, Input: map[string]any{"a": "test"}, RunOutput: "t", Output: map[string]any{"a": "test"}, Name: ""},
		{Script: `a[1]`, Input: map[string]any{"a": "test"}, RunOutput: "e", Output: map[string]any{"a": "test"}, Name: ""},
		{Script: `a[3]`, Input: map[string]any{"a": "test"}, RunOutput: "t", Output: map[string]any{"a": "test"}, Name: ""},
		{Script: `a[4]`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test"}, Name: ""},

		{Script: `a`, Input: map[string]any{"a": `"a"`}, RunOutput: `"a"`, Output: map[string]any{"a": `"a"`}, Name: ""},
		{Script: `a[0]`, Input: map[string]any{"a": `"a"`}, RunOutput: `"`, Output: map[string]any{"a": `"a"`}, Name: ""},
		{Script: `a[1]`, Input: map[string]any{"a": `"a"`}, RunOutput: "a", Output: map[string]any{"a": `"a"`}, Name: ""},

		{Script: `a = "\"a\""`, RunOutput: `"a"`, Output: map[string]any{"a": `"a"`}, Name: ""},
		{Script: `a = "\"a\""; a`, RunOutput: `"a"`, Output: map[string]any{"a": `"a"`}, Name: ""},
		{Script: `a = "\"a\""; a[0]`, RunOutput: `"`, Output: map[string]any{"a": `"a"`}, Name: ""},
		{Script: `a = "\"a\""; a[1]`, RunOutput: "a", Output: map[string]any{"a": `"a"`}, Name: ""},

		{Script: `a`, Input: map[string]any{"a": "a\\b"}, RunOutput: "a\\b", Output: map[string]any{"a": "a\\b"}, Name: ""},
		{Script: `a`, Input: map[string]any{"a": "a\\\\b"}, RunOutput: "a\\\\b", Output: map[string]any{"a": "a\\\\b"}, Name: ""},
		{Script: `a = "a\b"`, RunOutput: "a\b", Output: map[string]any{"a": "a\b"}, Name: ""},
		{Script: `a = "a\\b"`, RunOutput: "a\\b", Output: map[string]any{"a": "a\\b"}, Name: ""},

		{Script: `a[:]`, Input: map[string]any{"a": "test data"}, ParseError: fmt.Errorf("syntax error"), Output: map[string]any{"a": "test data"}, Name: ""},

		{Script: `a[0:]`, Input: map[string]any{"a": ""}, RunOutput: "", Output: map[string]any{"a": ""}, Name: ""},
		{Script: `a[1:]`, Input: map[string]any{"a": ""}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": ""}, Name: ""},
		{Script: `a[:0]`, Input: map[string]any{"a": ""}, RunOutput: "", Output: map[string]any{"a": ""}, Name: ""},
		{Script: `a[:1]`, Input: map[string]any{"a": ""}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": ""}, Name: ""},
		{Script: `a[0:0]`, Input: map[string]any{"a": ""}, RunOutput: "", Output: map[string]any{"a": ""}, Name: ""},

		{Script: `a[1:0]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("invalid slice index"), Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[-1:2]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[1:-2]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[-1:]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[:-2]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}, Name: ""},

		{Script: `a[0:0]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[0:1]`, Input: map[string]any{"a": "test data"}, RunOutput: "t", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[0:2]`, Input: map[string]any{"a": "test data"}, RunOutput: "te", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[0:3]`, Input: map[string]any{"a": "test data"}, RunOutput: "tes", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[0:7]`, Input: map[string]any{"a": "test data"}, RunOutput: "test da", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[0:8]`, Input: map[string]any{"a": "test data"}, RunOutput: "test dat", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[0:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "test data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[0:10]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}, Name: ""},

		{Script: `a[1:1]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[1:2]`, Input: map[string]any{"a": "test data"}, RunOutput: "e", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[1:3]`, Input: map[string]any{"a": "test data"}, RunOutput: "es", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[1:7]`, Input: map[string]any{"a": "test data"}, RunOutput: "est da", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[1:8]`, Input: map[string]any{"a": "test data"}, RunOutput: "est dat", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[1:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "est data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[1:10]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}, Name: ""},

		{Script: `a[0:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "test data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[1:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "est data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[2:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "st data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[3:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "t data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[7:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "ta", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[8:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "a", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[9:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}, Name: ""},

		{Script: `a[:0]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[:1]`, Input: map[string]any{"a": "test data"}, RunOutput: "t", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[:2]`, Input: map[string]any{"a": "test data"}, RunOutput: "te", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[:3]`, Input: map[string]any{"a": "test data"}, RunOutput: "tes", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[:7]`, Input: map[string]any{"a": "test data"}, RunOutput: "test da", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[:8]`, Input: map[string]any{"a": "test data"}, RunOutput: "test dat", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "test data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[:10]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}, Name: ""},

		{Script: `a[0:]`, Input: map[string]any{"a": "test data"}, RunOutput: "test data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[1:]`, Input: map[string]any{"a": "test data"}, RunOutput: "est data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[2:]`, Input: map[string]any{"a": "test data"}, RunOutput: "st data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[3:]`, Input: map[string]any{"a": "test data"}, RunOutput: "t data", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[7:]`, Input: map[string]any{"a": "test data"}, RunOutput: "ta", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[8:]`, Input: map[string]any{"a": "test data"}, RunOutput: "a", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[9:]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}, Name: ""},
		{Script: `a[10:]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}, Name: ""},

		// index assignment - len 0
		{Script: `a = ""; a[0] = "x"`, RunOutput: "x", Output: map[string]any{"a": "x"}, Name: ""},
		{Script: `a = ""; a[1] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": ""}, Name: ""},

		// index assignment - len 1
		{Script: `a = "a"; a[0] = "x"`, RunOutput: "x", Output: map[string]any{"a": "x"}, Name: ""},
		{Script: `a = "a"; a[1] = "x"`, RunOutput: "x", Output: map[string]any{"a": "ax"}, Name: ""},
		{Script: `a = "a"; a[2] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "a"}, Name: ""},

		// index assignment - len 2
		{Script: `a = "ab"; a[0] = "x"`, RunOutput: "x", Output: map[string]any{"a": "xb"}, Name: ""},
		{Script: `a = "ab"; a[1] = "x"`, RunOutput: "x", Output: map[string]any{"a": "ax"}, Name: ""},
		{Script: `a = "ab"; a[2] = "x"`, RunOutput: "x", Output: map[string]any{"a": "abx"}, Name: ""},
		{Script: `a = "ab"; a[3] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "ab"}, Name: ""},

		// index assignment - len 3
		{Script: `a = "abc"; a[0] = "x"`, RunOutput: "x", Output: map[string]any{"a": "xbc"}, Name: ""},
		{Script: `a = "abc"; a[1] = "x"`, RunOutput: "x", Output: map[string]any{"a": "axc"}, Name: ""},
		{Script: `a = "abc"; a[2] = "x"`, RunOutput: "x", Output: map[string]any{"a": "abx"}, Name: ""},
		{Script: `a = "abc"; a[3] = "x"`, RunOutput: "x", Output: map[string]any{"a": "abcx"}, Name: ""},
		{Script: `a = "abc"; a[4] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "abc"}, Name: ""},

		// index assignment - vm types
		{Script: `a = "abc"; a[1] = nil`, RunOutput: nil, Output: map[string]any{"a": "ac"}, Name: ""},
		{Script: `a = "abc"; a[1] = true`, RunError: fmt.Errorf("type bool cannot be assigned to type string"), Output: map[string]any{"a": "abc"}, Name: ""},
		{Script: `a = "abc"; a[1] = 120`, RunOutput: int64(120), Output: map[string]any{"a": "axc"}, Name: ""},
		{Script: `a = "abc"; a[1] = 2.2`, RunError: fmt.Errorf("type float64 cannot be assigned to type string"), Output: map[string]any{"a": "abc"}, Name: ""},
		{Script: `a = "abc"; a[1] = ["a"]`, RunError: fmt.Errorf("type []interface {} cannot be assigned to type string"), Output: map[string]any{"a": "abc"}, Name: ""},

		// index assignment - Go types
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": reflect.Value{}}, RunError: fmt.Errorf("type reflect.Value cannot be assigned to type string"), Output: map[string]any{"a": "abc"}, Name: ""},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": nil}, RunOutput: nil, Output: map[string]any{"a": "ac"}, Name: ""},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": true}, RunError: fmt.Errorf("type bool cannot be assigned to type string"), Output: map[string]any{"a": "abc"}, Name: ""},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": int32(120)}, RunOutput: int32(120), Output: map[string]any{"a": "axc"}, Name: ""},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": int64(120)}, RunOutput: int64(120), Output: map[string]any{"a": "axc"}, Name: ""},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": float32(1.1)}, RunError: fmt.Errorf("type float32 cannot be assigned to type string"), Output: map[string]any{"a": "abc"}, Name: ""},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": float64(2.2)}, RunError: fmt.Errorf("type float64 cannot be assigned to type string"), Output: map[string]any{"a": "abc"}, Name: ""},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": "x"}, RunOutput: "x", Output: map[string]any{"a": "axc"}, Name: ""},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": 'x'}, RunOutput: 'x', Output: map[string]any{"a": "axc"}, Name: ""},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": struct{}{}}, RunError: fmt.Errorf("type struct {} cannot be assigned to type string"), Output: map[string]any{"a": "abc"}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestVar(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	testInput1 := map[string]any{"b": func() {}}
	tests := []Test{
		// simple one variable
		{Script: `1 = 2`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `var 1 = 2`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `a = 1++`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `var a = 1++`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `var a := 1`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `y = z`, RunError: fmt.Errorf("undefined symbol 'z'"), Name: ""},

		{Script: `a := 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = nil`, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `a = true`, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `a = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1.1`, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `a = "a"`, RunOutput: "a", Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `var a = nil`, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `var a = true`, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `var a = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `var a = 1.1`, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `var a = "a"`, RunOutput: "a", Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `a = b`, Input: map[string]any{"b": reflect.Value{}}, RunOutput: reflect.Value{}, Output: map[string]any{"a": reflect.Value{}, "b": reflect.Value{}}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": nil}, RunOutput: nil, Output: map[string]any{"a": nil, "b": nil}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": true}, RunOutput: true, Output: map[string]any{"a": true, "b": true}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": int32(1)}, RunOutput: int32(1), Output: map[string]any{"a": int32(1), "b": int32(1)}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": int64(1)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1), "b": int64(1)}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": float32(1.1)}, RunOutput: float32(1.1), Output: map[string]any{"a": float32(1.1), "b": float32(1.1)}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": float64(1.1)}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1), "b": float64(1.1)}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": "a"}, RunOutput: "a", Output: map[string]any{"a": "a", "b": "a"}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": 'a'}, RunOutput: 'a', Output: map[string]any{"a": 'a', "b": 'a'}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": struct{}{}}, RunOutput: struct{}{}, Output: map[string]any{"a": struct{}{}, "b": struct{}{}}, Name: ""},

		{Script: `var a = b`, Input: map[string]any{"b": reflect.Value{}}, RunOutput: reflect.Value{}, Output: map[string]any{"a": reflect.Value{}, "b": reflect.Value{}}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": nil}, RunOutput: nil, Output: map[string]any{"a": nil, "b": nil}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": true}, RunOutput: true, Output: map[string]any{"a": true, "b": true}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": int32(1)}, RunOutput: int32(1), Output: map[string]any{"a": int32(1), "b": int32(1)}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": int64(1)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1), "b": int64(1)}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": float32(1.1)}, RunOutput: float32(1.1), Output: map[string]any{"a": float32(1.1), "b": float32(1.1)}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": float64(1.1)}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1), "b": float64(1.1)}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": "a"}, RunOutput: "a", Output: map[string]any{"a": "a", "b": "a"}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": 'a'}, RunOutput: 'a', Output: map[string]any{"a": 'a', "b": 'a'}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": struct{}{}}, RunOutput: struct{}{}, Output: map[string]any{"a": struct{}{}, "b": struct{}{}}, Name: ""},

		// simple one variable overwrite
		{Script: `a = true; a = nil`, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `a = nil; a = true`, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `a = 1; a = 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1.1; a = 2.2`, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}, Name: ""},
		{Script: `a = "a"; a = "b"`, RunOutput: "b", Output: map[string]any{"a": "b"}, Name: ""},

		{Script: `var a = true; var a = nil`, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `var a = nil; var a = true`, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `var a = 1; var a = 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `var a = 1.1; var a = 2.2`, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}, Name: ""},
		{Script: `var a = "a"; var a = "b"`, RunOutput: "b", Output: map[string]any{"a": "b"}, Name: ""},

		{Script: `a = nil`, Input: map[string]any{"a": true}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `a = true`, Input: map[string]any{"a": nil}, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `a = 2`, Input: map[string]any{"a": int32(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 2`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 2.2`, Input: map[string]any{"a": float32(1.1)}, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}, Name: ""},
		{Script: `a = 2.2`, Input: map[string]any{"a": float64(1.1)}, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}, Name: ""},
		{Script: `a = "b"`, Input: map[string]any{"a": "a"}, RunOutput: "b", Output: map[string]any{"a": "b"}},

		{Script: `var a = nil`, Input: map[string]any{"a": true}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `var a = true`, Input: map[string]any{"a": nil}, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `var a = 2`, Input: map[string]any{"a": int32(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `var a = 2`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `var a = 2.2`, Input: map[string]any{"a": float32(1.1)}, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}, Name: ""},
		{Script: `var a = 2.2`, Input: map[string]any{"a": float64(1.1)}, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}, Name: ""},
		{Script: `var a = "b"`, Input: map[string]any{"a": "a"}, RunOutput: "b", Output: map[string]any{"a": "b"}, Name: ""},

		// Go variable copy
		{Script: `a = b`, Input: testInput1, RunOutput: testInput1["b"], Output: map[string]any{"a": testInput1["b"], "b": testInput1["b"]}, Name: ""},
		{Script: `var a = b`, Input: testInput1, RunOutput: testInput1["b"], Output: map[string]any{"a": testInput1["b"], "b": testInput1["b"]}, Name: ""},

		{Script: `a = b`, Input: map[string]any{"b": testVarValue}, RunOutput: testVarValue, Output: map[string]any{"a": testVarValue, "b": testVarValue}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarBoolP}, RunOutput: testVarBoolP, Output: map[string]any{"a": testVarBoolP, "b": testVarBoolP}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarInt32P}, RunOutput: testVarInt32P, Output: map[string]any{"a": testVarInt32P, "b": testVarInt32P}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarInt64P}, RunOutput: testVarInt64P, Output: map[string]any{"a": testVarInt64P, "b": testVarInt64P}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarFloat32P}, RunOutput: testVarFloat32P, Output: map[string]any{"a": testVarFloat32P, "b": testVarFloat32P}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarFloat64P}, RunOutput: testVarFloat64P, Output: map[string]any{"a": testVarFloat64P, "b": testVarFloat64P}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarStringP}, RunOutput: testVarStringP, Output: map[string]any{"a": testVarStringP, "b": testVarStringP}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarFuncP}, RunOutput: testVarFuncP, Output: map[string]any{"a": testVarFuncP, "b": testVarFuncP}, Name: ""},

		{Script: `var a = b`, Input: map[string]any{"b": testVarValue}, RunOutput: testVarValue, Output: map[string]any{"a": testVarValue, "b": testVarValue}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarBoolP}, RunOutput: testVarBoolP, Output: map[string]any{"a": testVarBoolP, "b": testVarBoolP}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarInt32P}, RunOutput: testVarInt32P, Output: map[string]any{"a": testVarInt32P, "b": testVarInt32P}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarInt64P}, RunOutput: testVarInt64P, Output: map[string]any{"a": testVarInt64P, "b": testVarInt64P}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarFloat32P}, RunOutput: testVarFloat32P, Output: map[string]any{"a": testVarFloat32P, "b": testVarFloat32P}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarFloat64P}, RunOutput: testVarFloat64P, Output: map[string]any{"a": testVarFloat64P, "b": testVarFloat64P}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarStringP}, RunOutput: testVarStringP, Output: map[string]any{"a": testVarStringP, "b": testVarStringP}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarFuncP}, RunOutput: testVarFuncP, Output: map[string]any{"a": testVarFuncP, "b": testVarFuncP}, Name: ""},

		{Script: `a = b`, Input: map[string]any{"b": testVarValueBool}, RunOutput: testVarValueBool, Output: map[string]any{"a": testVarValueBool, "b": testVarValueBool}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueInt32}, RunOutput: testVarValueInt32, Output: map[string]any{"a": testVarValueInt32, "b": testVarValueInt32}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueInt64}, RunOutput: testVarValueInt64, Output: map[string]any{"a": testVarValueInt64, "b": testVarValueInt64}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueFloat32}, RunOutput: testVarValueFloat32, Output: map[string]any{"a": testVarValueFloat32, "b": testVarValueFloat32}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueFloat64}, RunOutput: testVarValueFloat64, Output: map[string]any{"a": testVarValueFloat64, "b": testVarValueFloat64}, Name: ""},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueString}, RunOutput: testVarValueString, Output: map[string]any{"a": testVarValueString, "b": testVarValueString}, Name: ""},

		{Script: `var a = b`, Input: map[string]any{"b": testVarValueBool}, RunOutput: testVarValueBool, Output: map[string]any{"a": testVarValueBool, "b": testVarValueBool}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueInt32}, RunOutput: testVarValueInt32, Output: map[string]any{"a": testVarValueInt32, "b": testVarValueInt32}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueInt64}, RunOutput: testVarValueInt64, Output: map[string]any{"a": testVarValueInt64, "b": testVarValueInt64}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueFloat32}, RunOutput: testVarValueFloat32, Output: map[string]any{"a": testVarValueFloat32, "b": testVarValueFloat32}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueFloat64}, RunOutput: testVarValueFloat64, Output: map[string]any{"a": testVarValueFloat64, "b": testVarValueFloat64}, Name: ""},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueString}, RunOutput: testVarValueString, Output: map[string]any{"a": testVarValueString, "b": testVarValueString}, Name: ""},

		// one variable spacing
		{Script: `
a  =  1
`, RunOutput: int64(1), Name: ""},
		{Script: `
a  =  1;
`, RunOutput: int64(1), Name: ""},

		// one variable many values
		{Script: `, b = 1, 2`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(2), Output: map[string]any{"b": int64(1)}, Name: ""},
		{Script: `var , b = 1, 2`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(2), Output: map[string]any{"b": int64(1)}, Name: ""},
		{Script: `a,  = 1, 2`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `var a,  = 1, 2`, ParseError: fmt.Errorf("syntax error"), Name: ""},

		// TOFIX: should not error?
		{Script: `a = 1, 2`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `var a = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1)}, Name: ""},
		// TOFIX: should not error?
		{Script: `a = 1, 2, 3`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `var a = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1)}, Name: ""},

		// two variables many values
		{Script: `a, b  = 1,`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `var a, b  = 1,`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `a, b  = ,1`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(1), Name: ""},
		{Script: `var a, b  = ,1`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(1), Name: ""},
		{Script: `a, b  = 1,,`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `var a, b  = 1,,`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `a, b  = ,1,`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `var a, b  = ,1,`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `a, b  = ,,1`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `var a, b  = ,,1`, ParseError: fmt.Errorf("syntax error"), Name: ""},

		{Script: `a.c, b = 1, 2`, RunError: fmt.Errorf("undefined symbol 'a'"), Name: ""},
		{Script: `a, b.c = 1, 2`, RunError: fmt.Errorf("undefined symbol 'b'"), Name: ""},

		{Script: `a, b = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `var a, b = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a, b = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},
		{Script: `var a, b = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},
		{Script: `a, b = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},
		{Script: `var a, b = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},

		// three variables many values
		{Script: `a, b, c = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `var a, b, c = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a, b, c = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},
		{Script: `var a, b, c = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},
		{Script: `a, b, c = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)}, Name: ""},
		{Script: `var a, b, c = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)}, Name: ""},
		{Script: `a, b, c = 1, 2, 3, 4`, RunOutput: int64(4), Output: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)}, Name: ""},
		{Script: `var a, b, c = 1, 2, 3, 4`, RunOutput: int64(4), Output: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)}, Name: ""},

		// scope
		{Script: `func(){ a = 1 }(); a`, RunError: fmt.Errorf("undefined symbol 'a'"), Name: ""},
		{Script: `func(){ var a = 1 }(); a`, RunError: fmt.Errorf("undefined symbol 'a'"), Name: ""},

		{Script: `a = 1; func(){ a = 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `var a = 1; func(){ a = 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; func(){ var a = 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `var a = 1; func(){ var a = 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(1)}, Name: ""},

		// function return
		{Script: `a, 1++ = func(){ return 1, 2 }()`, RunError: fmt.Errorf("invalid operation"), Output: map[string]any{"a": int64(1)}, Name: ""},

		{Script: `a = func(){ return 1 }()`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `var a = func(){ return 1 }()`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a, b = func(){ return 1, 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},
		{Script: `var a, b = func(){ return 1, 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestModule(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `module a.b { }`, ParseError: fmt.Errorf("syntax error"), Name: ""},
		{Script: `module a { 1++ }`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `module a { }; a.b`, RunError: fmt.Errorf("invalid operation 'b'"), Name: ""},

		{Script: `module a { b = nil }; a.b`, RunOutput: nil, Name: ""},
		{Script: `module a { b = true }; a.b`, RunOutput: true, Name: ""},
		{Script: `module a { b = 1 }; a.b`, RunOutput: int64(1), Name: ""},
		{Script: `module a { b = 1.1 }; a.b`, RunOutput: float64(1.1), Name: ""},
		{Script: `module a { b = "a" }; a.b`, RunOutput: "a", Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestNew(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `new(foo)`, RunError: fmt.Errorf("undefined type 'foo'"), Name: ""},
		{Script: `new(nilT)`, Types: map[string]any{"nilT": nil}, RunError: fmt.Errorf("type cannot be nil for make"), Name: ""},

		{Script: `a = new(bool); *a`, RunOutput: false, Name: ""},
		{Script: `a = new(int32); *a`, RunOutput: int32(0), Name: ""},
		{Script: `a = new(int64); *a`, RunOutput: int64(0), Name: ""},
		{Script: `a = new(float32); *a`, RunOutput: float32(0), Name: ""},
		{Script: `a = new(float64); *a`, RunOutput: float64(0), Name: ""},
		{Script: `a = new(string); *a`, RunOutput: "", Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestMake(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make(foo)`, RunError: fmt.Errorf("undefined type 'foo'"), Name: ""},
		{Script: `make(a.b)`, Types: map[string]any{"a": true}, RunError: fmt.Errorf("no namespace called: a"), Name: ""},
		{Script: `make(a.b)`, Types: map[string]any{"b": true}, RunError: fmt.Errorf("no namespace called: a"), Name: ""},

		{Script: `make(nilT)`, Types: map[string]any{"nilT": nil}, RunError: fmt.Errorf("type cannot be nil for make"), Name: ""},

		{Script: `make(bool)`, RunOutput: false, Name: ""},
		{Script: `make(int32)`, RunOutput: int32(0), Name: ""},
		{Script: `make(int64)`, RunOutput: int64(0), Name: ""},
		{Script: `make(float32)`, RunOutput: float32(0), Name: ""},
		{Script: `make(float64)`, RunOutput: float64(0), Name: ""},
		{Script: `make(string)`, RunOutput: "", Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestMakeType(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make(type a, 1++)`, RunError: fmt.Errorf("invalid operation"), Name: ""},

		{Script: `make(type a, true)`, RunOutput: reflect.TypeOf(true), Name: ""},
		{Script: `a = make(type a, true)`, RunOutput: reflect.TypeOf(true), Output: map[string]any{"a": reflect.TypeOf(true)}, Name: ""},
		{Script: `make(type a, true); a = make([]a)`, RunOutput: []bool{}, Output: map[string]any{"a": []bool{}}, Name: ""},
		{Script: `make(type a, make([]bool))`, RunOutput: reflect.TypeOf([]bool{}), Name: ""},
		{Script: `make(type a, make([]bool)); a = make(a)`, RunOutput: []bool{}, Output: map[string]any{"a": []bool{}}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestReferencingAndDereference(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		// TOFIX:
		// {Script: `a = 1; b = &a; *b = 2; *b`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestMakeChan(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make(chan foobar, 2)`, RunError: fmt.Errorf("undefined type 'foobar'"), Name: ""},
		{Script: `make(chan nilT, 2)`, Types: map[string]any{"nilT": nil}, RunError: fmt.Errorf("type cannot be nil for make"), Name: ""},
		{Script: `make(chan bool, 1++)`, RunError: fmt.Errorf("invalid operation"), Name: ""},

		{Script: `a = make(chan bool); b = func (c) { c <- true }; go b(a); <- a`, RunOutput: true, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestChan(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = make(chan bool, 2); 1++ <- 1`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `a = make(chan bool, 2); a <- 1++`, RunError: fmt.Errorf("invalid operation"), Name: ""},

		// TODO: move close from core into vm code, then update tests

		{Script: `1 <- 1`, RunError: fmt.Errorf("invalid operation for chan"), Name: ""},
		// TODO: this panics, should we capture the panic in a better way?
		// {Script: `a = make(chan bool, 2); close(a); a <- true`, Input: map[string]any{"close": func(b any) { reflect.ValueOf(b).Close() }}, RunError: fmt.Errorf("channel is close")},
		// TODO: add chan syntax ok
		// {Script: `a = make(chan bool, 2); a <- true; close(a); b, ok <- a; ok`, Input: map[string]any{"close": func(b any) { reflect.ValueOf(b).Close() }}, RunOutput: false, Output: map[string]any{"b": nil}},
		{Script: `a = make(chan bool, 2); a <- true; close(a); b = false; b <- a`, Input: map[string]any{"close": func(b any) { reflect.ValueOf(b).Close() }}, RunOutput: true, Output: map[string]any{"b": true}, Name: ""},
		// TOFIX: add chan syntax ok, do not return error. Also b should be nil
		{Script: `a = make(chan bool, 2); a <- true; close(a); b = false; b <- a; b <- a`, Input: map[string]any{"close": func(b any) { reflect.ValueOf(b).Close() }}, RunError: fmt.Errorf("failed to send to channel"), Output: map[string]any{"b": true}, Name: ""},

		{Script: `a = make(chan bool, 2); a <- 1`, RunError: fmt.Errorf("cannot use type int64 as type bool to send to chan"), Name: ""},

		{Script: `a = make(chan interface, 2); a <- nil; <- a`, RunOutput: nil, Name: ""},
		{Script: `a = make(chan bool, 2); a <- true; <- a`, RunOutput: true, Name: ""},
		{Script: `a = make(chan int32, 2); a <- 1; <- a`, RunOutput: int32(1), Name: ""},
		{Script: `a = make(chan int64, 2); a <- 1; <- a`, RunOutput: int64(1), Name: ""},
		{Script: `a = make(chan float32, 2); a <- 1.1; <- a`, RunOutput: float32(1.1), Name: ""},
		{Script: `a = make(chan float64, 2); a <- 1.1; <- a`, RunOutput: float64(1.1), Name: ""},

		{Script: `a = make(chan string, 2); a <- 1; <- a`, RunError: fmt.Errorf("cannot use type int64 as type string to send to chan"), Name: ""},

		{Script: `a <- nil; <- a`, Input: map[string]any{"a": make(chan any, 2)}, RunOutput: nil, Name: ""},
		{Script: `a <- true; <- a`, Input: map[string]any{"a": make(chan bool, 2)}, RunOutput: true, Name: ""},
		{Script: `a <- 1; <- a`, Input: map[string]any{"a": make(chan int32, 2)}, RunOutput: int32(1), Name: ""},
		{Script: `a <- 1; <- a`, Input: map[string]any{"a": make(chan int64, 2)}, RunOutput: int64(1), Name: ""},
		{Script: `a <- 1.1; <- a`, Input: map[string]any{"a": make(chan float32, 2)}, RunOutput: float32(1.1), Name: ""},
		{Script: `a <- 1.1; <- a`, Input: map[string]any{"a": make(chan float64, 2)}, RunOutput: float64(1.1), Name: ""},
		{Script: `a <- "b"; <- a`, Input: map[string]any{"a": make(chan string, 2)}, RunOutput: "b", Name: ""},

		{Script: `a = make(chan bool, 2); a <- true; a <- <- a`, RunOutput: nil, Name: ""},
		{Script: `a = make(chan bool, 2); a <- true; a <- (<- a)`, RunOutput: nil, Name: ""},
		{Script: `a = make(chan bool, 2); a <- true; a <- <- a; <- a`, RunOutput: true, Name: ""},
		{Script: `a = make(chan bool, 2); a <- true; b = false; b <- a`, RunOutput: true, Output: map[string]any{"b": true}, Name: ""},
		// TOFIX: if variable is not created yet, should make variable instead of error
		// {Script: `a = make(chan bool, 2); a <- true; b <- a`, RunOutput: true, Output: map[string]any{"b": true}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestCloseChan(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = make(chan int64, 1); close(a); <- a`, RunOutput: int64(0), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestVMDelete(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `delete(1++)`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `delete("a", 1++)`, RunError: fmt.Errorf("invalid operation"), Name: ""},

		{Script: `a = 1; delete("a"); a`, RunError: fmt.Errorf("undefined symbol 'a'"), Name: ""},

		{Script: `delete("a")`, Name: ""},
		{Script: `delete("a", false)`, Name: ""},
		{Script: `delete("a", true)`, Name: ""},
		{Script: `delete("a", nil)`, Name: ""},

		{Script: `a = 1; func b() { delete("a") }; b()`, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; func b() { delete("a", false) }; b()`, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; func b() { delete("a", true) }; b(); a`, RunError: fmt.Errorf("undefined symbol 'a'"), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestComment(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `# 1`, Name: ""},
		{Script: `# 1;`, Name: ""},
		{Script: `# 1 // 2`, Name: ""},
		{Script: `# 1 \n 2`, Name: ""},
		{Script: `# 1 # 2`, Name: ""},

		{Script: `1# 1`, RunOutput: int64(1), Name: ""},
		{Script: `1# 1;`, RunOutput: int64(1), Name: ""},
		{Script: `1# 1 // 2`, RunOutput: int64(1), Name: ""},
		{Script: `1# 1 \n 2`, RunOutput: int64(1), Name: ""},
		{Script: `1# 1 # 2`, RunOutput: int64(1), Name: ""},

		{Script: `1
# 1`, RunOutput: int64(1), Name: ""},
		{Script: `1
# 1;`, RunOutput: int64(1), Name: ""},
		{Script: `1
# 1 // 2`, RunOutput: int64(1), Name: ""},
		{Script: `1
# 1 \n 2`, RunOutput: int64(1), Name: ""},
		{Script: `1
# 1 # 2`, RunOutput: int64(1), Name: ""},

		{Script: `// 1`, Name: ""},
		{Script: `// 1;`, Name: ""},
		{Script: `// 1 // 2`, Name: ""},
		{Script: `// 1 \n 2`, Name: ""},
		{Script: `// 1 # 2`, Name: ""},

		{Script: `1// 1`, RunOutput: int64(1), Name: ""},
		{Script: `1// 1;`, RunOutput: int64(1), Name: ""},
		{Script: `1// 1 // 2`, RunOutput: int64(1), Name: ""},
		{Script: `1// 1 \n 2`, RunOutput: int64(1), Name: ""},
		{Script: `1// 1 # 2`, RunOutput: int64(1), Name: ""},

		{Script: `1
// 1`, RunOutput: int64(1), Name: ""},
		{Script: `1
// 1;`, RunOutput: int64(1), Name: ""},
		{Script: `1
// 1 // 2`, RunOutput: int64(1), Name: ""},
		{Script: `1
// 1 \n 2`, RunOutput: int64(1), Name: ""},
		{Script: `1
// 1 # 2`, RunOutput: int64(1), Name: ""},

		{Script: `/* 1 */`, Name: ""},
		{Script: `/* * 1 */`, Name: ""},
		{Script: `/* 1 * */`, Name: ""},
		{Script: `/** 1 */`, Name: ""},
		{Script: `/*** 1 */`, Name: ""},
		{Script: `/**** 1 */`, Name: ""},
		{Script: `/* 1 **/`, Name: ""},
		{Script: `/* 1 ***/`, Name: ""},
		{Script: `/* 1 ****/`, Name: ""},
		{Script: `/** 1 ****/`, Name: ""},
		{Script: `/*** 1 ****/`, Name: ""},
		{Script: `/**** 1 ****/`, Name: ""},

		{Script: `1/* 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1/* * 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1/* 1 * */`, RunOutput: int64(1), Name: ""},
		{Script: `1/** 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1/*** 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1/**** 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1/* 1 **/`, RunOutput: int64(1), Name: ""},
		{Script: `1/* 1 ***/`, RunOutput: int64(1), Name: ""},
		{Script: `1/* 1 ****/`, RunOutput: int64(1), Name: ""},
		{Script: `1/** 1 ****/`, RunOutput: int64(1), Name: ""},
		{Script: `1/*** 1 ****/`, RunOutput: int64(1), Name: ""},
		{Script: `1/**** 1 ****/`, RunOutput: int64(1), Name: ""},

		{Script: `/* 1 */1`, RunOutput: int64(1), Name: ""},
		{Script: `/* * 1 */1`, RunOutput: int64(1), Name: ""},
		{Script: `/* 1 * */1`, RunOutput: int64(1), Name: ""},
		{Script: `/** 1 */1`, RunOutput: int64(1), Name: ""},
		{Script: `/*** 1 */1`, RunOutput: int64(1), Name: ""},
		{Script: `/**** 1 */1`, RunOutput: int64(1), Name: ""},
		{Script: `/* 1 **/1`, RunOutput: int64(1), Name: ""},
		{Script: `/* 1 ***/1`, RunOutput: int64(1), Name: ""},
		{Script: `/* 1 ****/1`, RunOutput: int64(1), Name: ""},
		{Script: `/** 1 ****/1`, RunOutput: int64(1), Name: ""},
		{Script: `/*** 1 ****/1`, RunOutput: int64(1), Name: ""},
		{Script: `/**** 1 ****/1`, RunOutput: int64(1), Name: ""},

		{Script: `1
/* 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1
/* * 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1
/* 1 * */`, RunOutput: int64(1), Name: ""},
		{Script: `1
/** 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1
/*** 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1
/**** 1 */`, RunOutput: int64(1), Name: ""},
		{Script: `1
/* 1 **/`, RunOutput: int64(1), Name: ""},
		{Script: `1
/* 1 ***/`, RunOutput: int64(1), Name: ""},
		{Script: `1
/* 1 ****/`, RunOutput: int64(1), Name: ""},
		{Script: `1
/** 1 ****/`, RunOutput: int64(1), Name: ""},
		{Script: `1
/*** 1 ****/`, RunOutput: int64(1), Name: ""},
		{Script: `1
/**** 1 ****/`, RunOutput: int64(1), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestValidate(t *testing.T) {
	scripts := []string{
		`for {}`,
		`for i = 0; i > -1; i++ {}`,
		`select {}`,
		`a = make(chan int64); <-a `,
		`a = make(chan int64); for <-a {} `,
		`a = make(chan int64); for b in a {} `,
		`
		a = make(chan int64)
		b = make(chan int64)
		select {
			case <-a: toString(1);
			//case <-b: test2();
		}`,
	}
	for _, script := range scripts {
		runValidateTest(t, script)
	}
}

func runValidateTest(t *testing.T, script string) {
	toString := func(value any) string {
		return fmt.Sprintf("%v", value)
	}
	v := New(nil)
	_ = v.Define("toString", toString)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := v.Validate(ctx, script)
	if err != nil {
		t.Errorf("execute error - received %#v", err)
	}
}

func TestCancelWithContext(t *testing.T) {
	scripts := []string{
		`
b = 0
closeWaitChan()
for {
	b = 1
}
`,
		`
b = 0
closeWaitChan()
for {
	for {
		b = 1
	}
}
`,
		`
a = []
for i = 0; i < 10000; i++ {
	a += 1
}
b = 0
closeWaitChan()
for i in a {
	b = i
}
`,
		`
a = []
for i = 0; i < 10000; i++ {
	a += 1
}
b = 0
closeWaitChan()
for i in a {
	for j in a {
		b = j
	}
}
`,
		`
closeWaitChan()
b = 0
for i = 0; true; nil {
}
`,
		`
b = 0
closeWaitChan()
for i = 0; true; nil {
	for j = 0; true; nil {
		b = 1
	}
}
`,
		`
a = {}
for i = 0; i < 10000; i++ {
	a[toString(i)] = 1
}
b = 0
closeWaitChan()
for i in a {
	b = 1
}
`,
		`
a = {}
for i = 0; i < 10000; i++ {
	a[toString(i)] = 1
}
b = 0
closeWaitChan()
for i in a {
	for j in a {
		b = 1
	}
}
`,
		`
closeWaitChan()
<-make(chan string)
`,
		`
a = ""
closeWaitChan()
a <-make(chan string)
`,
		`
for {
	a = ""
	closeWaitChan()
	a <-make(chan string)
}
`,
		`
a = make(chan int)
closeWaitChan()
a <- 1
`,
		`
a = make(chan interface)
closeWaitChan()
a <- nil
`,
		`
a = make(chan int64, 1)
closeWaitChan()
for v in a { }
`,
		`
closeWaitChan()
select { }
`,
	}
	for _, script := range scripts {
		runCancelTestWithContext(t, script)
	}
}

func runCancelTestWithContext(t *testing.T, script string) {
	waitChan := make(chan struct{}, 1)
	closeWaitChan := func() {
		close(waitChan)
	}
	toString := func(value any) string {
		return fmt.Sprintf("%v", value)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-waitChan
		cancel()
	}()
	v := New(nil)
	err := v.Define("closeWaitChan", closeWaitChan)
	if err != nil {
		t.Errorf("Define error: %v", err)
	}
	err = v.Define("toString", toString)
	if err != nil {
		t.Errorf("Define error: %v", err)
	}

	_, err = v.Executor().Run(ctx, script)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("execute error - received %#v - expected: %#v", err, context.Canceled)
	}
}

func TestTwoContextSameEnv(t *testing.T) {
	var waitGroup sync.WaitGroup
	env := envPkg.NewEnv()
	v := New(&Config{Env: env})
	ctx1, cancel := context.WithCancel(context.Background())
	e := v.Executor()
	waitGroup.Add(1)
	go func() {
		_, err := e.Run(ctx1, "func myFn(a) { return 123; }; for { }")
		if err == nil || !errors.Is(err, context.Canceled) {
			t.Errorf("execute error - received %#v - expected: %#v", err, context.Canceled)
		}
		waitGroup.Done()
	}()
	time.Sleep(time.Millisecond)
	cancel()
	waitGroup.Wait()

	ctx2 := context.Background()
	_, err2 := e.Run(ctx2, "a = myFn(1)")
	if err2 != nil {
		t.Error("")
	}
}

func TestContextConcurrency(t *testing.T) {
	var waitGroup sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	v := New(nil)
	waitGroup.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			_, err := v.Executor().Run(ctx, "for { }")
			if err == nil || !errors.Is(err, context.Canceled) {
				t.Errorf("execute error - received %#v - expected: %#v", err, context.Canceled)
			}
			waitGroup.Done()
		}()
	}
	time.Sleep(time.Millisecond)
	cancel()
	waitGroup.Wait()

	//_, err := env.ExecuteContext(ctx, "for { }")
	//if err == nil || err.Error() != ErrInterrupt.Error() {
	//	t.Errorf("execute error - received %#v - expected: %#v", err, ErrInterrupt)
	//}
	//
	//ctx, cancel = context.WithCancel(context.Background())
	//
	//_, err = env.Execute("for i = 0; i < 1000; i ++ {}")
	//if err != nil {
	//	t.Errorf("execute error - received: %v - expected: %v", err, nil)
	//}
	//
	//waitGroup.Add(100)
	//for i := 0; i < 100; i++ {
	//	go func() {
	//		_, err := env.ExecuteContext(ctx, "for { }")
	//		if err == nil || err.Error() != ErrInterrupt.Error() {
	//			t.Errorf("execute error - received %#v - expected: %#v", err, ErrInterrupt)
	//		}
	//		waitGroup.Done()
	//	}()
	//}
	//time.Sleep(time.Millisecond)
	//cancel()
	//waitGroup.Wait()
}

func TestAssignToInterface(t *testing.T) {
	v := New(nil)
	X := new(struct {
		Stdout io.Writer
	})
	err := v.Define("X", X)
	if err != nil {
		t.Errorf("Define error: %v", err)
	}
	err = v.Define("a", new(os.File))
	if err != nil {
		t.Errorf("Define error: %v", err)
	}
	_, err = v.Executor().Run(nil, `X.Stdout = a`)
	if err != nil {
		t.Errorf("execute error - received %#v - expected: %#v", err, context.Canceled)
	}
}

// TestValueEqual do some basic ValueEqual tests for coverage
func TestValueEqual(t *testing.T) {
	result := valueEqual(true, true)
	if result != true {
		t.Fatal("ValueEqual")
	}
	result = valueEqual(true, false)
	if result != false {
		t.Fatal("ValueEqual")
	}
	result = valueEqual(false, true)
	if result != false {
		t.Fatal("ValueEqual")
	}
}

func TestLetsStatementPosition(t *testing.T) {
	src := `a, b = 1, 2`
	stmts, _ := parser.ParseSrc(src)
	stmt := stmts.(*ast.StmtsStmt)
	e := stmt.Stmts[0].(*ast.LetsStmt)
	pos := e.Position()
	if pos.Line != 1 {
		t.Fatalf("%v != %v", pos.Line, 1)
	}
	if pos.Column != 1 {
		t.Fatalf("%v != %v", pos.Column, 1)
	}
}

type Foo struct {
	Value int
}
type Bar struct {
	Foo
	Ref *int
}

func (f Foo) ValueReceiver() int {
	return f.Value
}
func (b *Bar) PointerReceiver() (int, int) {
	b.Value = 0
	*b.Ref = 0
	return b.Value, *b.Ref
}

func TestCallStructMethod(t *testing.T) {
	t.Parallel()

	ref := 10
	ptr := &Bar{
		Foo: Foo{
			Value: 100,
		},
		Ref: &ref,
	}
	val := Bar{
		Foo: Foo{
			Value: 200,
		},
		Ref: &ref,
	}

	// execution in native go
	v := ptr.ValueReceiver()
	if v != 100 {
		t.Fatal("ptr: call value receiver failed, v should equal to 100")
	}
	v, r := ptr.PointerReceiver()
	if v != 0 || r != 0 {
		t.Fatal("ptr: call pointer receiver failed, both should be 0")
	}

	ref = 10
	v = val.ValueReceiver()
	if v != 200 {
		t.Fatal("val: call value receiver failed, v should equal to 200")
	}
	v, r = val.PointerReceiver()
	if v != 0 || r != 0 {
		t.Fatal("val: call pointer receiver failed, both should be 0")
	}

	// reinitialize values before executing script in VM
	ptr.Value = 100
	val.Value = 200

	// Define pointer 'ptr' to struct Bar in VM
	ref = 10
	tests := []Test{
		{Script: "ptr.ValueReceiver()", Input: map[string]any{"ptr": ptr}, RunOutput: 100},
		{Script: "ptr.PointerReceiver()", Input: map[string]any{"ptr": ptr}, RunOutput: []any{0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}

	// Define value 'val' to struct Bar in VM
	ref = 10
	tests = []Test{
		{Script: "val.ValueReceiver()", Input: map[string]interface{}{"val": val}, RunOutput: 200},
		{Script: "val.PointerReceiver()", Input: map[string]interface{}{"val": val}, RunOutput: []interface{}{0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestURL(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	value, err := New(&Config{DefineImport: utils.Ptr(true)}).Executor().Run(nil, `
url = import("net/url")
v1 = make(url.Values)
v1.Set("a", "a")
if v1.Get("a") != "a" {
	return "value a not set"
} 
v2 = make(url.Values) 
v2.Set("b", "b")
if v2.Get("b") != "b" {
	return "value b not set"
} 
v2.Get("a")
`)
	if err != nil {
		t.Errorf("execute error - received: %v expected: %v", err, nil)
	}
	if value != "" {
		t.Errorf("execute value - received: %#v expected: %#v", value, "")
	}
}

var testPackagesEnvSetupFunc = func(t *testing.T, env *envPkg.Env) { runner.DefineImport(env) }

func TestDefineImport(t *testing.T) {
	value, err := New(&Config{DefineImport: utils.Ptr(true)}).Executor().Run(nil, `strings = import("strings"); strings.ToLower("TEST")`)
	if err != nil {
		t.Errorf("execute error - received: %v - expected: %v", err, nil)
	}
	if value != "test" {
		t.Errorf("execute value - received: %v - expected: %v", value, int(4))
	}
}

func TestDefineImportPackageNotFound(t *testing.T) {
	_ = os.Unsetenv("ANKO_DEBUG")
	value, err := New(&Config{DefineImport: utils.Ptr(true)}).Executor().Run(nil, `a = import("a")`)
	expectedError := "package 'a' not found"
	if err == nil || err.Error() != expectedError {
		t.Errorf("execute error - received: %v - expected: %v", err, expectedError)
	}
	if value != nil {
		t.Errorf("execute value - received: %v - expected: %v", value, nil)
	}
}

func TestDefineImportPackageDefineError(t *testing.T) {
	_ = os.Unsetenv("ANKO_DEBUG")
	packages.Packages.Insert("testPackage", map[string]any{"bad.name": testing.Coverage})

	value, err := New(&Config{DefineImport: utils.Ptr(true)}).Executor().Run(nil, `a = import("testPackage")`)
	expectedError := "import error: unknown symbol 'bad.name'"
	if err == nil || err.Error() != expectedError {
		t.Errorf("execute error - received: %v - expected: %v", err, expectedError)
	}
	if value != nil {
		t.Errorf("execute value - received: %v - expected: %v", value, nil)
	}

	packages.Packages.Insert("testPackage", map[string]any{"Coverage": testing.Coverage})
	packages.PackageTypes.Insert("testPackage", map[string]any{"bad.name": int64(1)})

	value, err = New(&Config{DefineImport: utils.Ptr(true)}).Executor().Run(nil, `a = import("testPackage")`)
	expectedError = "import error: unknown symbol 'bad.name'"
	if err == nil || err.Error() != expectedError {
		t.Errorf("execute error - received: %v - expected: %v", err, expectedError)
	}
	if value != nil {
		t.Errorf("execute value - received: %v - expected: %v", value, nil)
	}
}

func TestTime(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `time = import("time"); a = make(time.Time); a.IsZero()`, RunOutput: true},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true}) })
	}
}

func TestSync(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `sync = import("sync"); once = make(sync.Once); a = []; func add() { a += "a" }; once.Do(add); once.Do(add); a`, RunOutput: []any{"a"}, Output: map[string]any{"a": []any{"a"}}},
		{Script: `sync = import("sync"); waitGroup = make(sync.WaitGroup); waitGroup.Add(2);  func done() { waitGroup.Done() }; go done(); go done(); waitGroup.Wait(); "a"`, RunOutput: "a"},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true}) })
	}
}

func TestStringsPkg(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `strings = import("strings"); a = " one two "; b = strings.TrimSpace(a)`, RunOutput: "one two", Output: map[string]any{"a": " one two ", "b": "one two"}},
		{Script: `strings = import("strings"); a = "a b c d"; b = strings.Split(a, " ")`, RunOutput: []string{"a", "b", "c", "d"}, Output: map[string]any{"a": "a b c d", "b": []string{"a", "b", "c", "d"}}},
		{Script: `strings = import("strings"); a = "a b c d"; b = strings.SplitN(a, " ", 3)`, RunOutput: []string{"a", "b", "c d"}, Output: map[string]any{"a": "a b c d", "b": []string{"a", "b", "c d"}}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true}) })
	}
}

func TestStrconv(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	var toRune = func(s string) rune {
		if len(s) == 0 {
			return 0
		}
		return []rune(s)[0]
	}
	var toString = func(v any) string {
		if b, ok := v.([]byte); ok {
			return string(b)
		}
		return fmt.Sprint(v)
	}
	tests := []Test{
		{Script: `strconv = import("strconv"); a = true; b = strconv.FormatBool(a)`, RunOutput: "true", Output: map[string]any{"a": true, "b": "true"}, Name: ""},
		{Script: `strconv = import("strconv"); a = 1.1; b = strconv.FormatFloat(a, toRune("f"), -1, 64)`, Input: map[string]any{"toRune": toRune}, RunOutput: "1.1", Output: map[string]any{"a": float64(1.1), "b": "1.1"}, Name: ""},
		{Script: `strconv = import("strconv"); a = 1; b = strconv.FormatInt(a, 10)`, RunOutput: "1", Output: map[string]any{"a": int64(1), "b": "1"}, Name: ""},
		{Script: `strconv = import("strconv"); b = strconv.FormatInt(a, 10)`, Input: map[string]any{"a": uint64(1)}, RunOutput: "1", Output: map[string]any{"a": uint64(1), "b": "1"}, Name: ""},

		{Script: `strconv = import("strconv"); a = "true"; b, err = strconv.ParseBool(a); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "true", "b": true, "err": "<nil>"}, Name: ""},
		{Script: `strconv = import("strconv"); a = "2"; b, err = strconv.ParseBool(a); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseBool: parsing "2": invalid syntax`, Output: map[string]any{"a": "2", "b": false, "err": `strconv.ParseBool: parsing "2": invalid syntax`}, Name: ""},
		{Script: `strconv = import("strconv"); a = "1.1"; b, err = strconv.ParseFloat(a, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1.1", "b": float64(1.1), "err": "<nil>"}, Name: ""},
		{Script: `strconv = import("strconv"); a = "a"; b, err = strconv.ParseFloat(a, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseFloat: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": float64(0), "err": `strconv.ParseFloat: parsing "a": invalid syntax`}, Name: ""},
		{Script: `strconv = import("strconv"); a = "1"; b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1", "b": int64(1), "err": "<nil>"}, Name: ""},
		{Script: `strconv = import("strconv"); a = "1.1"; b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseInt: parsing "1.1": invalid syntax`, Output: map[string]any{"a": "1.1", "b": int64(0), "err": `strconv.ParseInt: parsing "1.1": invalid syntax`}, Name: ""},
		{Script: `strconv = import("strconv"); a = "a"; b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseInt: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": int64(0), "err": `strconv.ParseInt: parsing "a": invalid syntax`}, Name: ""},
		{Script: `strconv = import("strconv"); a = "1"; b, err = strconv.ParseUint(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1", "b": uint64(1), "err": "<nil>"}, Name: ""},
		{Script: `strconv = import("strconv"); a = "a"; b, err = strconv.ParseUint(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseUint: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": uint64(0), "err": `strconv.ParseUint: parsing "a": invalid syntax`}, Name: ""},

		{Script: `strconv = import("strconv"); a = "true"; var b, err = strconv.ParseBool(a); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "true", "b": true, "err": "<nil>"}, Name: ""},
		{Script: `strconv = import("strconv"); a = "2"; var b, err = strconv.ParseBool(a); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseBool: parsing "2": invalid syntax`, Output: map[string]any{"a": "2", "b": false, "err": `strconv.ParseBool: parsing "2": invalid syntax`}, Name: ""},
		{Script: `strconv = import("strconv"); a = "1.1"; var b, err = strconv.ParseFloat(a, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1.1", "b": float64(1.1), "err": "<nil>"}, Name: ""},
		{Script: `strconv = import("strconv"); a = "a"; var b, err = strconv.ParseFloat(a, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseFloat: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": float64(0), "err": `strconv.ParseFloat: parsing "a": invalid syntax`}, Name: ""},
		{Script: `strconv = import("strconv"); a = "1"; var b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1", "b": int64(1), "err": "<nil>"}, Name: ""},
		{Script: `strconv = import("strconv"); a = "1.1"; var b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseInt: parsing "1.1": invalid syntax`, Output: map[string]any{"a": "1.1", "b": int64(0), "err": `strconv.ParseInt: parsing "1.1": invalid syntax`}, Name: ""},
		{Script: `strconv = import("strconv"); a = "a"; var b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseInt: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": int64(0), "err": `strconv.ParseInt: parsing "a": invalid syntax`}, Name: ""},
		{Script: `strconv = import("strconv"); a = "1"; var b, err = strconv.ParseUint(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1", "b": uint64(1), "err": "<nil>"}, Name: ""},
		{Script: `strconv = import("strconv"); a = "a"; var b, err = strconv.ParseUint(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseUint: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": uint64(0), "err": `strconv.ParseUint: parsing "a": invalid syntax`}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true}) })
	}
}

func TestSort(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `sort = import("sort"); a = make([]int); a += [5, 3, 1, 4, 2]; sort.Ints(a); a`, RunOutput: []int{1, 2, 3, 4, 5}, Output: map[string]any{"a": []int{1, 2, 3, 4, 5}}, Name: ""},
		{Script: `sort = import("sort"); a = make([]float64); a += [5.5, 3.3, 1.1, 4.4, 2.2]; sort.Float64s(a); a`, RunOutput: []float64{1.1, 2.2, 3.3, 4.4, 5.5}, Output: map[string]any{"a": []float64{1.1, 2.2, 3.3, 4.4, 5.5}}, Name: ""},
		{Script: `sort = import("sort"); a = make([]string); a += ["e", "c", "a", "d", "b"]; sort.Strings(a); a`, RunOutput: []string{"a", "b", "c", "d", "e"}, Output: map[string]any{"a": []string{"a", "b", "c", "d", "e"}}, Name: ""},
		{Script: `
sort = import("sort")
a = [5, 1.1, 3, "f", "2", "4.4"]
sortFuncs = make(sort.SortFuncsStruct)
sortFuncs.LenFunc = func() { return len(a) }
sortFuncs.LessFunc = func(i, j) { return a[i] < a[j] }
sortFuncs.SwapFunc = func(i, j) { temp = a[i]; a[i] = a[j]; a[j] = temp }
sort.Sort(sortFuncs)
a
`,
			RunOutput: []any{"f", float64(1.1), "2", int64(3), "4.4", int64(5)}, Output: map[string]any{"a": []any{"f", float64(1.1), "2", int64(3), "4.4", int64(5)}}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true}) })
	}
}

func TestRegexp(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^simple$"); re.MatchString("simple")`, RunOutput: true, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^simple$"); re.MatchString("no match")`, RunOutput: false, Name: ""},

		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^simple$", "b": "simple"}, RunOutput: true, Output: map[string]any{"a": "^simple$", "b": "simple"}, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^simple$", "b": "no match"}, RunOutput: false, Output: map[string]any{"a": "^simple$", "b": "no match"}, Name: ""},

		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.String()`, RunOutput: "^a\\.\\d+\\.b$", Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("a.1.b")`, RunOutput: true, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("a.22.b")`, RunOutput: true, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("a.333.b")`, RunOutput: true, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("no match")`, RunOutput: false, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("a+1+b")`, RunOutput: false, Name: ""},

		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.String()`, Input: map[string]any{"a": "^a\\.\\d+\\.b$"}, RunOutput: "^a\\.\\d+\\.b$", Output: map[string]any{"a": "^a\\.\\d+\\.b$"}, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.1.b"}, RunOutput: true, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.1.b"}, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.22.b"}, RunOutput: true, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.22.b"}, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.333.b"}, RunOutput: true, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.333.b"}, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "no match"}, RunOutput: false, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "no match"}, Name: ""},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a+1+b"}, RunOutput: false, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a+1+b"}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true}) })
	}
}

func TestJson(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `json = import("encoding/json"); a = make(mapStringInterface); a["b"] = "b"; c, err = json.Marshal(a); err`, Types: map[string]any{"mapStringInterface": map[string]any{}}, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": []byte(`{"b":"b"}`)}, Name: ""},
		{Script: `json = import("encoding/json"); b = 1; err = json.Unmarshal(a, &b); err`, Input: map[string]any{"a": []byte(`{"b": "b"}`)}, Output: map[string]any{"a": []byte(`{"b": "b"}`), "b": map[string]any{"b": "b"}}, Name: ""},
		{Script: `json = import("encoding/json"); b = 1; err = json.Unmarshal(a, &b); err`, Input: map[string]any{"a": `{"b": "b"}`}, Output: map[string]any{"a": `{"b": "b"}`, "b": map[string]any{"b": "b"}}, Name: ""},
		{Script: `json = import("encoding/json"); b = 1; err = json.Unmarshal(a, &b); err`, Input: map[string]any{"a": `[["1", "2"],["3", "4"]]`}, Output: map[string]any{"a": `[["1", "2"],["3", "4"]]`, "b": []any{[]any{"1", "2"}, []any{"3", "4"}}}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true}) })
	}
}

func TestBytes(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `bytes = import("bytes"); a = make(bytes.Buffer); n, err = a.WriteString("a"); if err != nil { return err }; n`, RunOutput: 1, Name: ""},
		{Script: `bytes = import("bytes"); a = make(bytes.Buffer); n, err = a.WriteString("a"); if err != nil { return err }; a.String()`, RunOutput: "a", Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true}) })
	}
}

func TestAnk(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "")
	tests := []Test{
		{Script: `load('testdata/testing.ank'); load('testdata/let.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/toString.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/op.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/func.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/len.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/for.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/switch.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/if.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/toBytes.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/toRunes.ank')`, Name: ""},
		{Script: `load('testdata/testing.ank'); load('testdata/chan.ank')`, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true, ImportCore: true}) })
	}
}

func TestKeys(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = {}; b = keys(a)`, RunOutput: []any{}, Output: map[string]any{"a": map[any]any{}}, Name: ""},
		{Script: `a = {"a": nil}; b = keys(a)`, RunOutput: []any{"a"}, Output: map[string]any{"a": map[any]any{"a": nil}}, Name: ""},
		{Script: `a = {"a": 1}; b = keys(a)`, RunOutput: []any{"a"}, Output: map[string]any{"a": map[any]any{"a": int64(1)}}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true, ImportCore: true}) })
	}
}

func TestKindOf(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `kindOf(a)`, Input: map[string]any{"a": reflect.Value{}}, RunOutput: "struct", Output: map[string]any{"a": reflect.Value{}}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": nil}, RunOutput: "nil", Output: map[string]any{"a": nil}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": true}, RunOutput: "bool", Output: map[string]any{"a": true}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": int32(1)}, RunOutput: "int32", Output: map[string]any{"a": int32(1)}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": int64(1)}, RunOutput: "int64", Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": float32(1.1)}, RunOutput: "float32", Output: map[string]any{"a": float32(1.1)}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": float64(1.1)}, RunOutput: "float64", Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": "a"}, RunOutput: "string", Output: map[string]any{"a": "a"}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": 'a'}, RunOutput: "int32", Output: map[string]any{"a": 'a'}, Name: ""},

		{Script: `kindOf(a)`, Input: map[string]any{"a": any(nil)}, RunOutput: "nil", Output: map[string]any{"a": any(nil)}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(true)}, RunOutput: "bool", Output: map[string]any{"a": any(true)}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(int32(1))}, RunOutput: "int32", Output: map[string]any{"a": any(int32(1))}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(int64(1))}, RunOutput: "int64", Output: map[string]any{"a": any(int64(1))}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(float32(1))}, RunOutput: "float32", Output: map[string]any{"a": any(float32(1))}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(float64(1))}, RunOutput: "float64", Output: map[string]any{"a": any(float64(1))}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any("a")}, RunOutput: "string", Output: map[string]any{"a": any("a")}, Name: ""},

		{Script: `kindOf(a)`, Input: map[string]any{"a": []any{}}, RunOutput: "slice", Output: map[string]any{"a": []any{}}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": []any{nil}}, RunOutput: "slice", Output: map[string]any{"a": []any{nil}}, Name: ""},

		{Script: `kindOf(a)`, Input: map[string]any{"a": map[string]any{}}, RunOutput: "map", Output: map[string]any{"a": map[string]any{}}, Name: ""},
		{Script: `kindOf(a)`, Input: map[string]any{"a": map[string]any{"b": "b"}}, RunOutput: "map", Output: map[string]any{"a": map[string]any{"b": "b"}}, Name: ""},

		{Script: `a = make(interface); kindOf(a)`, RunOutput: "nil", Output: map[string]any{"a": any(nil)}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true, ImportCore: true}) })
	}
}

func TestRange(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "")
	tests := []Test{
		// 0 arguments
		{Script: `range()`, RunError: fmt.Errorf("range expected at least 1 argument, got 0")},
		// 1 arguments(step == 1, start == 0)
		{Script: `range(-1)`, RunOutput: []int64{}},
		{Script: `range(0)`, RunOutput: []int64{}},
		{Script: `range(1)`, RunOutput: []int64{0}},
		{Script: `range(2)`, RunOutput: []int64{0, 1}},
		{Script: `range(10)`, RunOutput: []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		// 2 arguments(step == 1)
		{Script: `range(-5,-1)`, RunOutput: []int64{-5, -4, -3, -2}},
		{Script: `range(-1,1)`, RunOutput: []int64{-1, 0}},
		{Script: `range(1,5)`, RunOutput: []int64{1, 2, 3, 4}},
		// 3 arguments
		// step == 2
		{Script: `range(-5,-1,2)`, RunOutput: []int64{-5, -3}},
		{Script: `range(1,5,2)`, RunOutput: []int64{1, 3}},
		{Script: `range(-1,5,2)`, RunOutput: []int64{-1, 1, 3}},
		// step < 0 and from small to large
		{Script: `range(-5,-1,-1)`, RunOutput: []int64{}},
		{Script: `range(1,5,-1)`, RunOutput: []int64{}},
		{Script: `range(-1,5,-1)`, RunOutput: []int64{}},
		// step < 0 and from large to small
		{Script: `range(-1,-5,-1)`, RunOutput: []int64{-1, -2, -3, -4}},
		{Script: `range(5,1,-1)`, RunOutput: []int64{5, 4, 3, 2}},
		{Script: `range(5,-1,-1)`, RunOutput: []int64{5, 4, 3, 2, 1, 0}},
		// 4,5 arguments
		{Script: `range(1,5,1,1)`, RunError: fmt.Errorf("range expected at most 3 arguments, got 4")},
		{Script: `range(1,5,1,1,1)`, RunError: fmt.Errorf("range expected at most 3 arguments, got 5")},
		// more 0 test
		{Script: `range(0,1,2)`, RunOutput: []int64{0}},
		{Script: `range(1,0,2)`, RunOutput: []int64{}},
		{Script: `range(1,2,0)`, RunError: fmt.Errorf("range argument 3 must not be zero")},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true, ImportCore: true}) })
	}
}

func TestLoad(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "")
	notFoundRunErrorFunc := func(t *testing.T, err error) {
		if err == nil || !strings.HasPrefix(err.Error(), "open testdata/not-found.ank:") {
			t.Errorf("load not-found.ank failed - received: %v", err)
		}
	}
	tests := []Test{
		{Script: `load('testdata/test.ank'); X(1)`, RunOutput: int64(2)},
		{Script: `load('testdata/not-found.ank'); X(1)`, RunErrorFunc: &notFoundRunErrorFunc},
		{Script: `load('testdata/broken.ank'); X(1)`, RunError: fmt.Errorf("syntax error")},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true, ImportCore: true}) })
	}
}

func TestDefined(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "")
	tests := []Test{
		{Script: `var a = 1; defined("a")`, RunOutput: true},
		{Script: `defined("a")`, RunOutput: false},
		{Script: `func(){ var a = 1 }(); defined("a")`, RunOutput: false},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true, ImportCore: true}) })
	}
}

func TestToX(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `toBool(-2)`, RunOutput: false, Name: ""},
		{Script: `toBool(-1.5)`, RunOutput: false, Name: ""},
		{Script: `toBool(-1)`, RunOutput: false, Name: ""},
		{Script: `toBool(-0.4)`, RunOutput: false, Name: ""},
		{Script: `toBool(0)`, RunOutput: false, Name: ""},
		{Script: `toBool(0.4)`, RunOutput: true, Name: ""},
		{Script: `toBool(1)`, RunOutput: true, Name: ""},
		{Script: `toBool(1.5)`, RunOutput: true, Name: ""},
		{Script: `toBool(2)`, RunOutput: true, Name: ""},
		{Script: `toBool(true)`, RunOutput: true, Name: ""},
		{Script: `toBool(false)`, RunOutput: false, Name: ""},
		{Script: `toBool("true")`, RunOutput: true, Name: ""},
		{Script: `toBool("false")`, RunOutput: false, Name: ""},
		{Script: `toBool("yes")`, RunOutput: true, Name: ""},
		{Script: `toBool("ye")`, RunOutput: false, Name: ""},
		{Script: `toBool("y")`, RunOutput: true, Name: ""},
		{Script: `toBool("false")`, RunOutput: false, Name: ""},
		{Script: `toBool("f")`, RunOutput: false, Name: ""},
		{Script: `toBool("")`, RunOutput: false, Name: ""},
		{Script: `toBool(nil)`, RunOutput: false, Name: ""},
		{Script: `toBool({})`, RunOutput: false, Name: ""},
		{Script: `toBool([])`, RunOutput: false, Name: ""},
		{Script: `toBool([true])`, RunOutput: false, Name: ""},
		{Script: `toBool({"true": "true"})`, RunOutput: false, Name: ""},
		{Script: `toString(nil)`, RunOutput: "<nil>", Name: ""},
		{Script: `toString("")`, RunOutput: "", Name: ""},
		{Script: `toString(1)`, RunOutput: "1", Name: ""},
		{Script: `toString(1.2)`, RunOutput: "1.2", Name: ""},
		{Script: `toString(1/3)`, RunOutput: "0.3333333333333333", Name: ""},
		{Script: `toString(false)`, RunOutput: "false", Name: ""},
		{Script: `toString(true)`, RunOutput: "true", Name: ""},
		{Script: `toString({})`, RunOutput: "map[]", Name: ""},
		{Script: `toString({"foo": "bar"})`, RunOutput: "map[foo:bar]", Name: ""},
		{Script: `toString([true,nil])`, RunOutput: "[true <nil>]", Name: ""},
		{Script: `toString(toByteSlice("foo"))`, RunOutput: "foo", Name: ""},
		{Script: `toInt(nil)`, RunOutput: int64(0), Name: ""},
		{Script: `toInt(-2)`, RunOutput: int64(-2), Name: ""},
		{Script: `toInt(-1.4)`, RunOutput: int64(-1), Name: ""},
		{Script: `toInt(-1)`, RunOutput: int64(-1), Name: ""},
		{Script: `toInt(0)`, RunOutput: int64(0), Name: ""},
		{Script: `toInt(1)`, RunOutput: int64(1), Name: ""},
		{Script: `toInt(1.4)`, RunOutput: int64(1), Name: ""},
		{Script: `toInt(1.5)`, RunOutput: int64(1), Name: ""},
		{Script: `toInt(1.9)`, RunOutput: int64(1), Name: ""},
		{Script: `toInt(2)`, RunOutput: int64(2), Name: ""},
		{Script: `toInt(2.1)`, RunOutput: int64(2), Name: ""},
		{Script: `toInt("2")`, RunOutput: int64(2), Name: ""},
		{Script: `toInt("2.1")`, RunOutput: int64(2), Name: ""},
		{Script: `toInt(true)`, RunOutput: int64(1), Name: ""},
		{Script: `toInt(false)`, RunOutput: int64(0), Name: ""},
		{Script: `toInt({})`, RunOutput: int64(0), Name: ""},
		{Script: `toInt([])`, RunOutput: int64(0), Name: ""},
		{Script: `toFloat(nil)`, RunOutput: float64(0.0), Name: ""},
		{Script: `toFloat(-2)`, RunOutput: float64(-2.0), Name: ""},
		{Script: `toFloat(-1.4)`, RunOutput: float64(-1.4), Name: ""},
		{Script: `toFloat(-1)`, RunOutput: float64(-1.0), Name: ""},
		{Script: `toFloat(0)`, RunOutput: float64(0.0), Name: ""},
		{Script: `toFloat(1)`, RunOutput: float64(1.0), Name: ""},
		{Script: `toFloat(1.4)`, RunOutput: float64(1.4), Name: ""},
		{Script: `toFloat(1.5)`, RunOutput: float64(1.5), Name: ""},
		{Script: `toFloat(1.9)`, RunOutput: float64(1.9), Name: ""},
		{Script: `toFloat(2)`, RunOutput: float64(2.0), Name: ""},
		{Script: `toFloat(2.1)`, RunOutput: float64(2.1), Name: ""},
		{Script: `toFloat("2")`, RunOutput: float64(2.0), Name: ""},
		{Script: `toFloat("2.1")`, RunOutput: float64(2.1), Name: ""},
		{Script: `toFloat(true)`, RunOutput: float64(1.0), Name: ""},
		{Script: `toFloat(false)`, RunOutput: float64(0.0), Name: ""},
		{Script: `toFloat({})`, RunOutput: float64(0.0), Name: ""},
		{Script: `toFloat([])`, RunOutput: float64(0.0), Name: ""},
		{Script: `toChar(0x1F431)`, RunOutput: "🐱", Name: ""},
		{Script: `toChar(0)`, RunOutput: "\x00", Name: ""},
		{Script: `toRune("")`, RunOutput: rune(0), Name: ""},
		{Script: `toRune("🐱")`, RunOutput: rune(0x1F431), Name: ""},
		{Script: `toBoolSlice(nil)`, RunOutput: []bool{}, Name: ""},
		{Script: `toBoolSlice(1)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type int64"), Name: ""},
		{Script: `toBoolSlice(1.2)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type float64"), Name: ""},
		{Script: `toBoolSlice(false)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type bool"), Name: ""},
		{Script: `toBoolSlice({})`, RunError: fmt.Errorf("function wants argument type []interface {} but received type map[interface {}]interface {}"), Name: ""},
		{Script: `toBoolSlice([])`, RunOutput: []bool{}, Name: ""},
		{Script: `toBoolSlice([nil])`, RunOutput: []bool{false}, Name: ""},
		{Script: `toBoolSlice([1])`, RunOutput: []bool{false}, Name: ""},
		{Script: `toBoolSlice([1.1])`, RunOutput: []bool{false}, Name: ""},
		{Script: `toBoolSlice([true])`, RunOutput: []bool{true}, Name: ""},
		{Script: `toBoolSlice([[]])`, RunOutput: []bool{false}, Name: ""},
		{Script: `toBoolSlice([{}])`, RunOutput: []bool{false}, Name: ""},
		{Script: `toIntSlice(nil)`, RunOutput: []int64{}, Name: ""},
		{Script: `toIntSlice(1)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type int64"), Name: ""},
		{Script: `toIntSlice(1.2)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type float64"), Name: ""},
		{Script: `toIntSlice(false)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type bool"), Name: ""},
		{Script: `toIntSlice({})`, RunError: fmt.Errorf("function wants argument type []interface {} but received type map[interface {}]interface {}"), Name: ""},
		{Script: `toIntSlice([])`, RunOutput: []int64{}, Name: ""},
		{Script: `toIntSlice([nil])`, RunOutput: []int64{0}, Name: ""},
		{Script: `toIntSlice([1])`, RunOutput: []int64{1}, Name: ""},
		{Script: `toIntSlice([1.1])`, RunOutput: []int64{1}, Name: ""},
		{Script: `toIntSlice([true])`, RunOutput: []int64{0}, Name: ""},
		{Script: `toIntSlice([[]])`, RunOutput: []int64{0}, Name: ""},
		{Script: `toIntSlice([{}])`, RunOutput: []int64{0}, Name: ""},
		{Script: `toFloatSlice(nil)`, RunOutput: []float64{}, Name: ""},
		{Script: `toFloatSlice(1)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type int64"), Name: ""},
		{Script: `toFloatSlice(1.2)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type float64"), Name: ""},
		{Script: `toFloatSlice(false)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type bool"), Name: ""},
		{Script: `toFloatSlice({})`, RunError: fmt.Errorf("function wants argument type []interface {} but received type map[interface {}]interface {}"), Name: ""},
		{Script: `toFloatSlice([])`, RunOutput: []float64{}, Name: ""},
		{Script: `toFloatSlice([nil])`, RunOutput: []float64{0.0}, Name: ""},
		{Script: `toFloatSlice([1])`, RunOutput: []float64{1.0}, Name: ""},
		{Script: `toFloatSlice([1.1])`, RunOutput: []float64{1.1}, Name: ""},
		{Script: `toFloatSlice([true])`, RunOutput: []float64{0.0}, Name: ""},
		{Script: `toFloatSlice([[]])`, RunOutput: []float64{0.0}, Name: ""},
		{Script: `toFloatSlice([{}])`, RunOutput: []float64{0.0}, Name: ""},
		{Script: `toByteSlice(nil)`, RunOutput: []byte{}, Name: ""},
		{Script: `toByteSlice([])`, RunError: fmt.Errorf("function wants argument type string but received type []interface {}"), Name: ""},
		{Script: `toByteSlice(1)`, RunOutput: []byte{0x01}, Name: ""}, // FIXME?
		{Script: `toByteSlice(1.1)`, RunError: fmt.Errorf("function wants argument type string but received type float64"), Name: ""},
		{Script: `toByteSlice(true)`, RunError: fmt.Errorf("function wants argument type string but received type bool"), Name: ""},
		{Script: `toByteSlice("foo")`, RunOutput: []byte{'f', 'o', 'o'}, Name: ""},
		{Script: `toByteSlice("世界")`, RunOutput: []byte{0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c}, Name: ""},
		{Script: `toRuneSlice(nil)`, RunOutput: []rune{}, Name: ""},
		{Script: `toRuneSlice([])`, RunError: fmt.Errorf("function wants argument type string but received type []interface {}"), Name: ""},
		{Script: `toRuneSlice(1)`, RunOutput: []rune{0x01}, Name: ""}, // FIXME?
		{Script: `toRuneSlice(1.1)`, RunError: fmt.Errorf("function wants argument type string but received type float64"), Name: ""},
		{Script: `toRuneSlice(true)`, RunError: fmt.Errorf("function wants argument type string but received type bool"), Name: ""},
		{Script: `toRuneSlice("foo")`, RunOutput: []rune{'f', 'o', 'o'}, Name: ""},
		{Script: `toRuneSlice("世界")`, RunOutput: []rune{'世', '界'}, Name: ""},
		{Script: `toStringSlice([true,false,1])`, RunOutput: []string{"", "", "\x01"}, Name: ""}, // FIXME?
		{Script: `toDuration(nil)`, RunOutput: time.Duration(0), Name: ""},
		{Script: `toDuration(0)`, RunOutput: time.Duration(0), Name: ""},
		{Script: `toDuration(true)`, RunError: fmt.Errorf("function wants argument type int64 but received type bool"), Name: ""},
		{Script: `toDuration([])`, RunError: fmt.Errorf("function wants argument type int64 but received type []interface {}"), Name: ""},
		{Script: `toDuration({})`, RunError: fmt.Errorf("function wants argument type int64 but received type map[interface {}]interface {}"), Name: ""},
		{Script: `toDuration("")`, RunError: fmt.Errorf("function wants argument type int64 but received type string"), Name: ""},
		{Script: `toDuration("1s")`, RunError: fmt.Errorf("function wants argument type int64 but received type string"), Name: ""}, // TODO
		{Script: `toDuration(a)`, Input: map[string]any{"a": int64(time.Duration(123 * time.Minute))}, RunOutput: time.Duration(123 * time.Minute), Name: ""},
		{Script: `toDuration(a)`, Input: map[string]any{"a": float64(time.Duration(123 * time.Minute))}, RunOutput: time.Duration(123 * time.Minute), Name: ""},
		{Script: `toDuration(a)`, Input: map[string]any{"a": time.Duration(123 * time.Minute)}, RunOutput: time.Duration(123 * time.Minute), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true, ImportCore: true}) })
	}
}

func TestAddPackage(t *testing.T) {
	// empty
	v := New(nil)
	pack, _ := v.AddPackage("empty", map[string]any{}, map[string]any{})
	value, err := v.Executor().Run(nil, "empty")
	if err != nil {
		t.Errorf("AddPackage error - received: %v - expected: %v", err, nil)
	}
	val, _ := pack.GetValue("empty")
	switch data := value.(type) {
	case *envPkg.Env:
		if data.Values().Len() != 0 {
			t.Errorf("AddPackage value - received: %#v - expected: %#v", value, val)
		} else if data.Types().Len() != 0 {
			t.Errorf("AddPackage value - received: %#v - expected: %#v", value, val)
		}
	default:
		t.Errorf("AddPackage value - received: %#v - expected: %#v", value, val)
	}

	// bad package name
	v = New(nil)
	_, err = v.AddPackage("bad.package.name", map[string]any{}, map[string]any{})
	if err == nil {
		t.Errorf("AddPackage error - received: %v - expected: %v", err, "unknown symbol 'bad.package.name'")
	} else {
		if err.Error() != "unknown symbol 'bad.package.name'" {
			t.Errorf("AddPackage error - received: %v - expected: %v", err, "unknown symbol 'bad.package.name'")
		}
	}

	// bad method name
	v = New(nil)
	_, err = v.AddPackage("badMethodName", map[string]any{"a.b": "a"}, map[string]any{})
	if err == nil {
		t.Errorf("AddPackage error - received: %v - expected: %v", err, "unknown symbol 'a.b'")
	} else {
		if err.Error() != "unknown symbol 'a.b'" {
			t.Errorf("AddPackage error - received: %v - expected: %v", err, "unknown symbol 'a.b'")
		}
	}

	// bad type name
	v = New(nil)
	_, err = v.AddPackage("badTypeName", map[string]any{}, map[string]any{"a.b": "a"})
	if err == nil {
		t.Errorf("AddPackage error - received: %v - expected: %v", err, "unknown symbol 'a.b'")
	} else {
		if err.Error() != "unknown symbol 'a.b'" {
			t.Errorf("AddPackage error - received: %v - expected: %v", err, "unknown symbol 'a.b'")
		}
	}

	// method
	v = New(nil)
	_, _ = v.AddPackage("strings", map[string]any{"ToLower": strings.ToLower}, map[string]any{})
	value, err = v.Executor().Run(nil, `strings.ToLower("TEST")`)
	if err != nil {
		t.Errorf("AddPackage error - received: %v - expected: %v", err, nil)
	}
	if value != "test" {
		t.Errorf("AddPackage value - received: %v - expected: %v", value, int(4))
	}

	// type
	v = New(nil)
	_, _ = v.AddPackage("test", map[string]any{}, map[string]any{"array2x": [][]any{}})
	value, err = v.Executor().Run(nil, "a = make(test.array2x); a += [[1]]")
	if err != nil {
		t.Errorf("AddPackage error - received: %v - expected: %v", err, nil)
	}
	if !reflect.DeepEqual(value, [][]any{{int64(1)}}) {
		t.Errorf("AddPackage value - received: %#v - expected: %#v", value, [][]any{{int64(1)}})
	}
}

func TestEnvRef(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `func(x){return func(){return x}}(1)()`, RunOutput: int64(1), Name: ""},
		{Script: `func(x){if true {return func(){return x}}}(1)()`, RunOutput: int64(1), Name: ""},
		{Script: `func(x){if false {} else if true {return func(){return x}}}(1)()`, RunOutput: int64(1), Name: ""},
		{Script: `func(x){if false {} else {return func(){return x}}}(1)()`, RunOutput: int64(1), Name: ""},
		{Script: `func(x){for a = 1; a < 2; a++ { return func(){return x}}}(1)()`, RunOutput: int64(1), Name: ""},
		{Script: `func(x){for a in [1] { return func(){return x}}}(1)()`, RunOutput: int64(1), Name: ""},
		{Script: `func(x){for { return func(){return x}}}(1)()`, RunOutput: int64(1), Name: ""},
		{Script: `module m { f=func(x){return func(){return x}} }; m.f(1)()`, RunOutput: int64(1), Name: ""},
		{Script: `module m { f=func(x){return func(){return x}}(1)() }; m.f`, RunOutput: int64(1), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestFuncTypedParams(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `func a(x any){return x}; a(1)`, RunOutput: int64(1), Name: ""},
		{Script: `func a(x any){return x}; a("1")`, RunOutput: "1", Name: ""},
		{Script: `func a(x int64){return x}; a(1)`, RunOutput: int64(1), Name: ""},
		{Script: `func a(x int64){return x}; a("1")`, RunError: fmt.Errorf("function wants argument type int64 but received type string"), Name: ""},
		{Script: `func a(x string){return x}; a(1)`, RunError: fmt.Errorf("function wants argument type string but received type int64"), Name: ""},
		{Script: `func a(x string){return x}; a("1")`, RunOutput: "1", Name: ""},
		{Script: `func a(b, x int64){return x}; a("", 1)`, RunOutput: int64(1), Name: ""},
		{Script: `func a(b, x int64){return x}; a("", "1")`, RunError: fmt.Errorf("function wants argument type int64 but received type string"), Name: ""},
		{Script: `func a(x int64, b){return x}; a(1, "")`, RunOutput: int64(1), Name: ""},
		{Script: `func a(x int64, b){return x}; a("1", "")`, RunError: fmt.Errorf("function wants argument type int64 but received type string"), Name: ""},
		{Script: `func a(x...){return x}; a([]string{"1","2","3"}...)`, RunOutput: []any{"1", "2", "3"}, Name: ""},
		{Script: `func a(x int64...){return x}; a([]int64{1,2,3}...)`, RunOutput: []int64{1, 2, 3}, Name: ""},
		{Script: `func a(x string...){return x}; a([]string{"1","2","3"}...)`, RunOutput: []string{"1", "2", "3"}, Name: ""},
		{Script: `func a(x string...){return x}; a([]int64{1,2,3}...)`, RunError: fmt.Errorf("function wants argument type []string but received type []int64"), Name: ""},
		{Script: `func a(x int64...){return x}; a(1,2,3)`, RunOutput: []int64{1, 2, 3}, Name: ""},
		{Script: `func a(x int64...){return x}; a(1,"2",3)`, RunError: fmt.Errorf("function wants argument type []int64 but received type string"), Name: ""},
		{Script: `func a(x int64, y int64, z int64){return x}; a([]int64{1,2,3}...)`, RunOutput: int64(1), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestFuncTypedReturns(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `errors = import("errors");func a() {return errors.New("err")}; a()`, RunOutput: fmt.Errorf("err"), Name: ""},
		{Script: `errors = import("errors");func a() error {return errors.New("err")}; a()`, RunOutput: fmt.Errorf("err"), Name: ""},
		{Script: `errors = import("errors");func a() (error) {return errors.New("err")}; a()`, RunOutput: fmt.Errorf("err"), Name: ""},
		{Script: `errors = import("errors");func a() (int) {return errors.New("err")}; a()`, RunError: fmt.Errorf("invalid type for returned value, have: error, expected: int"), Name: ""},
		{Script: `errors = import("errors");func a() int64 {return 1}; a()`, RunOutput: int64(1), Name: ""},
		{Script: `errors = import("errors");func a() any {return 1}; a()`, RunOutput: int64(1), Name: ""},
		{Script: `errors = import("errors");func a() any {return "1"}; a()`, RunOutput: "1", Name: ""},
		{Script: `errors = import("errors");func a() () {return 1}; a()`, RunError: fmt.Errorf("invalid number of returned values, have 1, expected: 0"), Name: ""},
		{Script: `errors = import("errors");func a() () {}; a()`, Name: ""},
		{Script: `errors = import("errors");func a() (int64) {return 1}; a()`, RunOutput: int64(1), Name: ""},
		{Script: `errors = import("errors");func a() (int64,error) {return 1, nil}; b,c=a();b`, RunOutput: int64(1), Name: ""},
		{Script: `errors = import("errors");func a() (int64) {return 1,2}; a()`, RunError: fmt.Errorf("invalid number of returned values, have 2, expected: 1"), Name: ""},
		{Script: `errors = import("errors");func a() (int64,int64,int64) {return 1,2}; a()`, RunError: fmt.Errorf("invalid number of returned values, have 2, expected: 3"), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true, ImportCore: true}) })
	}
}

func TestTypedValues(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a := 1; a = 2`, RunOutput: int64(2)},
		{Script: `a = 1; a = "1"`, RunOutput: "1"},
		{Script: `a := 1; a = "1"`, RunError: vmUtils.ErrTypeMismatch, Name: ""},
		{Script: `a := 1; delete("a"); a = "1"`, RunOutput: "1"},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{DefineImport: true, ImportCore: true, ResetEnv: true}) })
	}
}

func TestExecuteError(t *testing.T) {
	script := "a]]"
	_, err := New(nil).Executor().Run(nil, script)
	if err == nil {
		t.Errorf("execute error - received: %v - expected: %v", err, fmt.Errorf("syntax error"))
	} else if err.Error() != "syntax error" {
		t.Errorf("execute error - received: %v - expected: %v", err, fmt.Errorf("syntax error"))
	}
}

func TestHas(t *testing.T) {
	toString := func(value any) string {
		return fmt.Sprintf("%v", value)
	}
	otherFn := func(value any) string {
		return fmt.Sprintf("%v", value)
	}
	third := func() (int, int, int) {
		return 1, 2, 3
	}
	fourth := func(fn func()) {
		fn()
	}
	fifth := func() runner.TestStruct {
		return runner.TestStruct{}
	}
	sixth := func() runner.TestInterface {
		return runner.TestStruct{}
	}
	vmprint := func(args ...any) {
		fmt.Println(fmt.Sprint(args))
	}
	arr := func() []int {
		return []int{0, 1, 2}
	}
	arr2 := func() []runner.TestInterface {
		return []runner.TestInterface{}
	}
	newEnv := func() *VM {
		v := New(&Config{DefineImport: utils.Ptr(true)})
		_ = v.Define("toString", toString)
		_ = v.Define("otherFn", otherFn)
		_ = v.Define("third", third)
		_ = v.Define("fourth", fourth)
		_ = v.Define("fifth", fifth)
		_ = v.Define("sixth", sixth)
		_ = v.Define("print", vmprint)
		_ = v.Define("arr", arr)
		_ = v.Define("arr2", arr2)
		return v
	}

	oks, err := newEnv().Has(context.Background(), `func a() { b() }; func b() {}; a()`, []any{toString})
	assert.NoError(t, err)
	oks, err = newEnv().Has(context.Background(), `a = arr2()[0]; a.TestMember()`, []any{toString})
	assert.NoError(t, err)
	oks, err = newEnv().Has(context.Background(), `arr()[0]`, []any{toString})
	assert.NoError(t, err)
	_, err = newEnv().Has(context.Background(), `print("test")`, []any{toString})
	assert.NoError(t, err)
	_, err = newEnv().Has(context.Background(), `t = sixth(); t.TestMember()`, []any{toString})
	assert.NoError(t, err)
	oks, err = newEnv().Has(context.Background(), `t = fifth(); t.TestMember()`, []any{toString})
	assert.NoError(t, err)
	assert.False(t, oks[0])
	oks, _ = newEnv().Has(context.Background(), `toString(123)`, []any{toString})
	assert.True(t, oks[0])
	oks, err = newEnv().Has(context.Background(), `a = toString; a(123)`, []any{toString})
	assert.True(t, oks[0])
	assert.NoError(t, err)
	oks, _ = newEnv().Has(context.Background(), `otherFn(123)`, []any{toString})
	assert.False(t, oks[0])
	oks, _ = newEnv().Has(context.Background(), `otherFn(123)`, []any{toString, otherFn})
	assert.False(t, oks[0])
	assert.True(t, oks[1])
	oks, _ = newEnv().Has(context.Background(), `a = toString; a(123); otherFn(123)`, []any{toString, otherFn})
	assert.True(t, oks[0])
	assert.True(t, oks[1])
	oks, err = newEnv().Has(context.Background(), `b = toString; a = func() { b(123) }; a()`, []any{toString})
	assert.NoError(t, err)
	assert.True(t, oks[0])
	oks, _ = newEnv().Has(context.Background(), `b = toString; a = func() { b(123) }`, []any{toString})
	assert.True(t, oks[0])
	oks, err = newEnv().Has(context.Background(), `h, m, s = third(); otherFn(s)`, []any{toString})
	assert.NoError(t, err)
	assert.False(t, oks[0])
	oks, err = newEnv().Has(context.Background(), `a = func() { for {  } }; a(); toString(123)`, []any{toString})
	assert.NoError(t, err)
	assert.True(t, oks[0])
	oks, err = newEnv().Has(context.Background(), `a = func() { for { toString(123) } }; a()`, []any{toString})
	assert.NoError(t, err)
	assert.True(t, oks[0])
	oks, err = newEnv().Has(context.Background(), `fourth(func() { toString(123) })`, []any{toString})
	assert.NoError(t, err)
	assert.True(t, oks[0])
	oks, err = newEnv().Has(context.Background(), `fmt = import("fmt"); fmt.Print("test")`, []any{toString})
	assert.NoError(t, err)
	assert.False(t, oks[0])
}

func TestTest(t *testing.T) {
	v := New(nil)
	e1 := v.Executor()
	_, _ = e1.Run(nil, "a = 1")
	assert.Equal(t, int64(1), utils.Must(e1.GetEnv().Get("a")).(int64))
	assert.Error(t, utils.MustErr(v.env.Get("b")))
	e2 := v.Executor()
	_, _ = e2.Run(nil, "b = 2")
	assert.Error(t, utils.MustErr(v.env.Get("a")))
	assert.Equal(t, int64(2), utils.Must(e2.GetEnv().Get("b")).(int64))
}

func TestDefineCtx(t *testing.T) {
	v := New(nil)
	_ = v.DefineCtx("a", func(context.Context) int64 { return 1 })
	_ = v.DefineCtx("b", func(context.Context, int64) int64 { return 1 })
	e := v.Executor()
	tests := []Test{
		{Script: `a()`, RunOutput: int64(1), Name: ""},
		{Script: `a(1)`, RunError: fmt.Errorf("function wants 0 arguments but received 1"), Name: ""},
		{Script: `b(1, 2)`, RunError: fmt.Errorf("function wants 1 arguments but received 2"), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{Executor: e}) })
	}
}

func TestRateLimitPeriod(t *testing.T) {
	v := New(&Config{RateLimitPeriod: time.Minute})
	_ = v.DefineCtx("a", func(context.Context) int64 { return 1 })
	e := v.Executor()
	tests := []Test{
		{Script: `a()`, RunOutput: int64(1), Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, &Options{Executor: e}) })
	}
}
