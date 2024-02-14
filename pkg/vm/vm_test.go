package vm

import (
	"context"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/packages"
	"github.com/alaingilbert/anko/pkg/utils"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
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

		{Script: `1..1`, RunError: fmt.Errorf(`strconv.ParseFloat: parsing "1..1": invalid syntax`)},
		{Script: `0x1g`, ParseError: fmt.Errorf("syntax error")},
		{Script: `9223372036854775808`, RunError: fmt.Errorf(`strconv.ParseInt: parsing "9223372036854775808": value out of range`)},

		{Script: `1`, RunOutput: int64(1)},
		{Script: `-1`, RunOutput: int64(-1)},
		{Script: `9223372036854775807`, RunOutput: int64(9223372036854775807)},
		{Script: `-9223372036854775807`, RunOutput: int64(-9223372036854775807)},
		{Script: `1.1`, RunOutput: float64(1.1)},
		{Script: `-1.1`, RunOutput: float64(-1.1)},
		{Script: `1e1`, RunOutput: float64(10)},
		{Script: `-1e1`, RunOutput: float64(-10)},
		{Script: `1e-1`, RunOutput: float64(0.1)},
		{Script: `-1e-1`, RunOutput: float64(-0.1)},
		{Script: `0x1`, RunOutput: int64(1)},
		{Script: `0xc`, RunOutput: int64(12)},
		// TOFIX:
		{Script: `0xe`, RunError: fmt.Errorf(`strconv.ParseFloat: parsing "0xe": invalid syntax`)},
		{Script: `0xf`, RunOutput: int64(15)},
		{Script: `-0x1`, RunOutput: int64(-1)},
		{Script: `-0xc`, RunOutput: int64(-12)},
		{Script: `-0xf`, RunOutput: int64(-15)},
	}
	runTests(t, tests, nil)
}

func TestStrings(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a`, Input: map[string]any{"a": 'a'}, RunOutput: 'a', Output: map[string]any{"a": 'a'}},
		{Script: `a.b`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support member operation"), Output: map[string]any{"a": 'a'}},
		{Script: `a[0]`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support index operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}},
		{Script: `a[0:1]`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support slice operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}},

		{Script: `a.b = "a"`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support member operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}},
		{Script: `a[0] = "a"`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support index operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}},
		{Script: `a[0:1] = "a"`, Input: map[string]any{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support slice operation"), RunOutput: nil, Output: map[string]any{"a": 'a'}},

		{Script: `a.b = "a"`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("type string does not support member operation"), Output: map[string]any{"a": "test"}},
		{Script: `a[0:1] = "a"`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("type string does not support slice operation for assignment"), Output: map[string]any{"a": "test"}},

		{Script: `a`, Input: map[string]any{"a": "test"}, RunOutput: "test", Output: map[string]any{"a": "test"}},
		{Script: `a["a"]`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"a": "test"}},
		{Script: `a[0]`, Input: map[string]any{"a": ""}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": ""}},
		{Script: `a[-1]`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test"}},
		{Script: `a[0]`, Input: map[string]any{"a": "test"}, RunOutput: "t", Output: map[string]any{"a": "test"}},
		{Script: `a[1]`, Input: map[string]any{"a": "test"}, RunOutput: "e", Output: map[string]any{"a": "test"}},
		{Script: `a[3]`, Input: map[string]any{"a": "test"}, RunOutput: "t", Output: map[string]any{"a": "test"}},
		{Script: `a[4]`, Input: map[string]any{"a": "test"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test"}},

		{Script: `a`, Input: map[string]any{"a": `"a"`}, RunOutput: `"a"`, Output: map[string]any{"a": `"a"`}},
		{Script: `a[0]`, Input: map[string]any{"a": `"a"`}, RunOutput: `"`, Output: map[string]any{"a": `"a"`}},
		{Script: `a[1]`, Input: map[string]any{"a": `"a"`}, RunOutput: "a", Output: map[string]any{"a": `"a"`}},

		{Script: `a = "\"a\""`, RunOutput: `"a"`, Output: map[string]any{"a": `"a"`}},
		{Script: `a = "\"a\""; a`, RunOutput: `"a"`, Output: map[string]any{"a": `"a"`}},
		{Script: `a = "\"a\""; a[0]`, RunOutput: `"`, Output: map[string]any{"a": `"a"`}},
		{Script: `a = "\"a\""; a[1]`, RunOutput: "a", Output: map[string]any{"a": `"a"`}},

		{Script: `a`, Input: map[string]any{"a": "a\\b"}, RunOutput: "a\\b", Output: map[string]any{"a": "a\\b"}},
		{Script: `a`, Input: map[string]any{"a": "a\\\\b"}, RunOutput: "a\\\\b", Output: map[string]any{"a": "a\\\\b"}},
		{Script: `a = "a\b"`, RunOutput: "a\b", Output: map[string]any{"a": "a\b"}},
		{Script: `a = "a\\b"`, RunOutput: "a\\b", Output: map[string]any{"a": "a\\b"}},

		{Script: `a[:]`, Input: map[string]any{"a": "test data"}, ParseError: fmt.Errorf("syntax error"), Output: map[string]any{"a": "test data"}},

		{Script: `a[0:]`, Input: map[string]any{"a": ""}, RunOutput: "", Output: map[string]any{"a": ""}},
		{Script: `a[1:]`, Input: map[string]any{"a": ""}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": ""}},
		{Script: `a[:0]`, Input: map[string]any{"a": ""}, RunOutput: "", Output: map[string]any{"a": ""}},
		{Script: `a[:1]`, Input: map[string]any{"a": ""}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": ""}},
		{Script: `a[0:0]`, Input: map[string]any{"a": ""}, RunOutput: "", Output: map[string]any{"a": ""}},

		{Script: `a[1:0]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("invalid slice index"), Output: map[string]any{"a": "test data"}},
		{Script: `a[-1:2]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}},
		{Script: `a[1:-2]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}},
		{Script: `a[-1:]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}},
		{Script: `a[:-2]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}},

		{Script: `a[0:0]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}},
		{Script: `a[0:1]`, Input: map[string]any{"a": "test data"}, RunOutput: "t", Output: map[string]any{"a": "test data"}},
		{Script: `a[0:2]`, Input: map[string]any{"a": "test data"}, RunOutput: "te", Output: map[string]any{"a": "test data"}},
		{Script: `a[0:3]`, Input: map[string]any{"a": "test data"}, RunOutput: "tes", Output: map[string]any{"a": "test data"}},
		{Script: `a[0:7]`, Input: map[string]any{"a": "test data"}, RunOutput: "test da", Output: map[string]any{"a": "test data"}},
		{Script: `a[0:8]`, Input: map[string]any{"a": "test data"}, RunOutput: "test dat", Output: map[string]any{"a": "test data"}},
		{Script: `a[0:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "test data", Output: map[string]any{"a": "test data"}},
		{Script: `a[0:10]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}},

		{Script: `a[1:1]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}},
		{Script: `a[1:2]`, Input: map[string]any{"a": "test data"}, RunOutput: "e", Output: map[string]any{"a": "test data"}},
		{Script: `a[1:3]`, Input: map[string]any{"a": "test data"}, RunOutput: "es", Output: map[string]any{"a": "test data"}},
		{Script: `a[1:7]`, Input: map[string]any{"a": "test data"}, RunOutput: "est da", Output: map[string]any{"a": "test data"}},
		{Script: `a[1:8]`, Input: map[string]any{"a": "test data"}, RunOutput: "est dat", Output: map[string]any{"a": "test data"}},
		{Script: `a[1:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "est data", Output: map[string]any{"a": "test data"}},
		{Script: `a[1:10]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}},

		{Script: `a[0:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "test data", Output: map[string]any{"a": "test data"}},
		{Script: `a[1:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "est data", Output: map[string]any{"a": "test data"}},
		{Script: `a[2:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "st data", Output: map[string]any{"a": "test data"}},
		{Script: `a[3:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "t data", Output: map[string]any{"a": "test data"}},
		{Script: `a[7:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "ta", Output: map[string]any{"a": "test data"}},
		{Script: `a[8:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "a", Output: map[string]any{"a": "test data"}},
		{Script: `a[9:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}},

		{Script: `a[:0]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}},
		{Script: `a[:1]`, Input: map[string]any{"a": "test data"}, RunOutput: "t", Output: map[string]any{"a": "test data"}},
		{Script: `a[:2]`, Input: map[string]any{"a": "test data"}, RunOutput: "te", Output: map[string]any{"a": "test data"}},
		{Script: `a[:3]`, Input: map[string]any{"a": "test data"}, RunOutput: "tes", Output: map[string]any{"a": "test data"}},
		{Script: `a[:7]`, Input: map[string]any{"a": "test data"}, RunOutput: "test da", Output: map[string]any{"a": "test data"}},
		{Script: `a[:8]`, Input: map[string]any{"a": "test data"}, RunOutput: "test dat", Output: map[string]any{"a": "test data"}},
		{Script: `a[:9]`, Input: map[string]any{"a": "test data"}, RunOutput: "test data", Output: map[string]any{"a": "test data"}},
		{Script: `a[:10]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}},

		{Script: `a[0:]`, Input: map[string]any{"a": "test data"}, RunOutput: "test data", Output: map[string]any{"a": "test data"}},
		{Script: `a[1:]`, Input: map[string]any{"a": "test data"}, RunOutput: "est data", Output: map[string]any{"a": "test data"}},
		{Script: `a[2:]`, Input: map[string]any{"a": "test data"}, RunOutput: "st data", Output: map[string]any{"a": "test data"}},
		{Script: `a[3:]`, Input: map[string]any{"a": "test data"}, RunOutput: "t data", Output: map[string]any{"a": "test data"}},
		{Script: `a[7:]`, Input: map[string]any{"a": "test data"}, RunOutput: "ta", Output: map[string]any{"a": "test data"}},
		{Script: `a[8:]`, Input: map[string]any{"a": "test data"}, RunOutput: "a", Output: map[string]any{"a": "test data"}},
		{Script: `a[9:]`, Input: map[string]any{"a": "test data"}, RunOutput: "", Output: map[string]any{"a": "test data"}},
		{Script: `a[10:]`, Input: map[string]any{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "test data"}},

		// index assignment - len 0
		{Script: `a = ""; a[0] = "x"`, RunOutput: "x", Output: map[string]any{"a": "x"}},
		{Script: `a = ""; a[1] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": ""}},

		// index assignment - len 1
		{Script: `a = "a"; a[0] = "x"`, RunOutput: "x", Output: map[string]any{"a": "x"}},
		{Script: `a = "a"; a[1] = "x"`, RunOutput: "x", Output: map[string]any{"a": "ax"}},
		{Script: `a = "a"; a[2] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "a"}},

		// index assignment - len 2
		{Script: `a = "ab"; a[0] = "x"`, RunOutput: "x", Output: map[string]any{"a": "xb"}},
		{Script: `a = "ab"; a[1] = "x"`, RunOutput: "x", Output: map[string]any{"a": "ax"}},
		{Script: `a = "ab"; a[2] = "x"`, RunOutput: "x", Output: map[string]any{"a": "abx"}},
		{Script: `a = "ab"; a[3] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "ab"}},

		// index assignment - len 3
		{Script: `a = "abc"; a[0] = "x"`, RunOutput: "x", Output: map[string]any{"a": "xbc"}},
		{Script: `a = "abc"; a[1] = "x"`, RunOutput: "x", Output: map[string]any{"a": "axc"}},
		{Script: `a = "abc"; a[2] = "x"`, RunOutput: "x", Output: map[string]any{"a": "abx"}},
		{Script: `a = "abc"; a[3] = "x"`, RunOutput: "x", Output: map[string]any{"a": "abcx"}},
		{Script: `a = "abc"; a[4] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": "abc"}},

		// index assignment - vm types
		{Script: `a = "abc"; a[1] = nil`, RunOutput: nil, Output: map[string]any{"a": "ac"}},
		{Script: `a = "abc"; a[1] = true`, RunError: fmt.Errorf("type bool cannot be assigned to type string"), Output: map[string]any{"a": "abc"}},
		{Script: `a = "abc"; a[1] = 120`, RunOutput: int64(120), Output: map[string]any{"a": "axc"}},
		{Script: `a = "abc"; a[1] = 2.2`, RunError: fmt.Errorf("type float64 cannot be assigned to type string"), Output: map[string]any{"a": "abc"}},
		{Script: `a = "abc"; a[1] = ["a"]`, RunError: fmt.Errorf("type []interface {} cannot be assigned to type string"), Output: map[string]any{"a": "abc"}},

		// index assignment - Go types
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": reflect.Value{}}, RunError: fmt.Errorf("type reflect.Value cannot be assigned to type string"), Output: map[string]any{"a": "abc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": nil}, RunOutput: nil, Output: map[string]any{"a": "ac"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": true}, RunError: fmt.Errorf("type bool cannot be assigned to type string"), Output: map[string]any{"a": "abc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": int32(120)}, RunOutput: int32(120), Output: map[string]any{"a": "axc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": int64(120)}, RunOutput: int64(120), Output: map[string]any{"a": "axc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": float32(1.1)}, RunError: fmt.Errorf("type float32 cannot be assigned to type string"), Output: map[string]any{"a": "abc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": float64(2.2)}, RunError: fmt.Errorf("type float64 cannot be assigned to type string"), Output: map[string]any{"a": "abc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": "x"}, RunOutput: "x", Output: map[string]any{"a": "axc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": 'x'}, RunOutput: 'x', Output: map[string]any{"a": "axc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]any{"b": struct{}{}}, RunError: fmt.Errorf("type struct {} cannot be assigned to type string"), Output: map[string]any{"a": "abc"}},
	}
	runTests(t, tests, nil)
}

func TestVar(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	testInput1 := map[string]any{"b": func() {}}
	tests := []Test{
		// simple one variable
		{Script: `1 = 2`, RunError: fmt.Errorf("invalid operation")},
		{Script: `var 1 = 2`, ParseError: fmt.Errorf("syntax error")},
		{Script: `a = 1++`, RunError: fmt.Errorf("invalid operation")},
		{Script: `var a = 1++`, RunError: fmt.Errorf("invalid operation")},
		{Script: `a := 1`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a := 1`, ParseError: fmt.Errorf("syntax error")},
		{Script: `y = z`, RunError: fmt.Errorf("undefined symbol 'z'")},

		{Script: `a = nil`, RunOutput: nil, Output: map[string]any{"a": nil}},
		{Script: `a = true`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1.1`, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}},
		{Script: `a = "a"`, RunOutput: "a", Output: map[string]any{"a": "a"}},

		{Script: `var a = nil`, RunOutput: nil, Output: map[string]any{"a": nil}},
		{Script: `var a = true`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `var a = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `var a = 1.1`, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}},
		{Script: `var a = "a"`, RunOutput: "a", Output: map[string]any{"a": "a"}},

		{Script: `a = b`, Input: map[string]any{"b": reflect.Value{}}, RunOutput: reflect.Value{}, Output: map[string]any{"a": reflect.Value{}, "b": reflect.Value{}}},
		{Script: `a = b`, Input: map[string]any{"b": nil}, RunOutput: nil, Output: map[string]any{"a": nil, "b": nil}},
		{Script: `a = b`, Input: map[string]any{"b": true}, RunOutput: true, Output: map[string]any{"a": true, "b": true}},
		{Script: `a = b`, Input: map[string]any{"b": int32(1)}, RunOutput: int32(1), Output: map[string]any{"a": int32(1), "b": int32(1)}},
		{Script: `a = b`, Input: map[string]any{"b": int64(1)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1), "b": int64(1)}},
		{Script: `a = b`, Input: map[string]any{"b": float32(1.1)}, RunOutput: float32(1.1), Output: map[string]any{"a": float32(1.1), "b": float32(1.1)}},
		{Script: `a = b`, Input: map[string]any{"b": float64(1.1)}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1), "b": float64(1.1)}},
		{Script: `a = b`, Input: map[string]any{"b": "a"}, RunOutput: "a", Output: map[string]any{"a": "a", "b": "a"}},
		{Script: `a = b`, Input: map[string]any{"b": 'a'}, RunOutput: 'a', Output: map[string]any{"a": 'a', "b": 'a'}},
		{Script: `a = b`, Input: map[string]any{"b": struct{}{}}, RunOutput: struct{}{}, Output: map[string]any{"a": struct{}{}, "b": struct{}{}}},

		{Script: `var a = b`, Input: map[string]any{"b": reflect.Value{}}, RunOutput: reflect.Value{}, Output: map[string]any{"a": reflect.Value{}, "b": reflect.Value{}}},
		{Script: `var a = b`, Input: map[string]any{"b": nil}, RunOutput: nil, Output: map[string]any{"a": nil, "b": nil}},
		{Script: `var a = b`, Input: map[string]any{"b": true}, RunOutput: true, Output: map[string]any{"a": true, "b": true}},
		{Script: `var a = b`, Input: map[string]any{"b": int32(1)}, RunOutput: int32(1), Output: map[string]any{"a": int32(1), "b": int32(1)}},
		{Script: `var a = b`, Input: map[string]any{"b": int64(1)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1), "b": int64(1)}},
		{Script: `var a = b`, Input: map[string]any{"b": float32(1.1)}, RunOutput: float32(1.1), Output: map[string]any{"a": float32(1.1), "b": float32(1.1)}},
		{Script: `var a = b`, Input: map[string]any{"b": float64(1.1)}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1), "b": float64(1.1)}},
		{Script: `var a = b`, Input: map[string]any{"b": "a"}, RunOutput: "a", Output: map[string]any{"a": "a", "b": "a"}},
		{Script: `var a = b`, Input: map[string]any{"b": 'a'}, RunOutput: 'a', Output: map[string]any{"a": 'a', "b": 'a'}},
		{Script: `var a = b`, Input: map[string]any{"b": struct{}{}}, RunOutput: struct{}{}, Output: map[string]any{"a": struct{}{}, "b": struct{}{}}},

		// simple one variable overwrite
		{Script: `a = true; a = nil`, RunOutput: nil, Output: map[string]any{"a": nil}},
		{Script: `a = nil; a = true`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = 1; a = 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = 1.1; a = 2.2`, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}},
		{Script: `a = "a"; a = "b"`, RunOutput: "b", Output: map[string]any{"a": "b"}},

		{Script: `var a = true; var a = nil`, RunOutput: nil, Output: map[string]any{"a": nil}},
		{Script: `var a = nil; var a = true`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `var a = 1; var a = 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `var a = 1.1; var a = 2.2`, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}},
		{Script: `var a = "a"; var a = "b"`, RunOutput: "b", Output: map[string]any{"a": "b"}},

		{Script: `a = nil`, Input: map[string]any{"a": true}, RunOutput: nil, Output: map[string]any{"a": nil}},
		{Script: `a = true`, Input: map[string]any{"a": nil}, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = 2`, Input: map[string]any{"a": int32(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = 2`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = 2.2`, Input: map[string]any{"a": float32(1.1)}, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}},
		{Script: `a = 2.2`, Input: map[string]any{"a": float64(1.1)}, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}},
		{Script: `a = "b"`, Input: map[string]any{"a": "a"}, RunOutput: "b", Output: map[string]any{"a": "b"}},

		{Script: `var a = nil`, Input: map[string]any{"a": true}, RunOutput: nil, Output: map[string]any{"a": nil}},
		{Script: `var a = true`, Input: map[string]any{"a": nil}, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `var a = 2`, Input: map[string]any{"a": int32(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `var a = 2`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `var a = 2.2`, Input: map[string]any{"a": float32(1.1)}, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}},
		{Script: `var a = 2.2`, Input: map[string]any{"a": float64(1.1)}, RunOutput: float64(2.2), Output: map[string]any{"a": float64(2.2)}},
		{Script: `var a = "b"`, Input: map[string]any{"a": "a"}, RunOutput: "b", Output: map[string]any{"a": "b"}},

		// Go variable copy
		{Script: `a = b`, Input: testInput1, RunOutput: testInput1["b"], Output: map[string]any{"a": testInput1["b"], "b": testInput1["b"]}},
		{Script: `var a = b`, Input: testInput1, RunOutput: testInput1["b"], Output: map[string]any{"a": testInput1["b"], "b": testInput1["b"]}},

		{Script: `a = b`, Input: map[string]any{"b": testVarValue}, RunOutput: testVarValue, Output: map[string]any{"a": testVarValue, "b": testVarValue}},
		{Script: `a = b`, Input: map[string]any{"b": testVarBoolP}, RunOutput: testVarBoolP, Output: map[string]any{"a": testVarBoolP, "b": testVarBoolP}},
		{Script: `a = b`, Input: map[string]any{"b": testVarInt32P}, RunOutput: testVarInt32P, Output: map[string]any{"a": testVarInt32P, "b": testVarInt32P}},
		{Script: `a = b`, Input: map[string]any{"b": testVarInt64P}, RunOutput: testVarInt64P, Output: map[string]any{"a": testVarInt64P, "b": testVarInt64P}},
		{Script: `a = b`, Input: map[string]any{"b": testVarFloat32P}, RunOutput: testVarFloat32P, Output: map[string]any{"a": testVarFloat32P, "b": testVarFloat32P}},
		{Script: `a = b`, Input: map[string]any{"b": testVarFloat64P}, RunOutput: testVarFloat64P, Output: map[string]any{"a": testVarFloat64P, "b": testVarFloat64P}},
		{Script: `a = b`, Input: map[string]any{"b": testVarStringP}, RunOutput: testVarStringP, Output: map[string]any{"a": testVarStringP, "b": testVarStringP}},
		{Script: `a = b`, Input: map[string]any{"b": testVarFuncP}, RunOutput: testVarFuncP, Output: map[string]any{"a": testVarFuncP, "b": testVarFuncP}},

		{Script: `var a = b`, Input: map[string]any{"b": testVarValue}, RunOutput: testVarValue, Output: map[string]any{"a": testVarValue, "b": testVarValue}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarBoolP}, RunOutput: testVarBoolP, Output: map[string]any{"a": testVarBoolP, "b": testVarBoolP}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarInt32P}, RunOutput: testVarInt32P, Output: map[string]any{"a": testVarInt32P, "b": testVarInt32P}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarInt64P}, RunOutput: testVarInt64P, Output: map[string]any{"a": testVarInt64P, "b": testVarInt64P}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarFloat32P}, RunOutput: testVarFloat32P, Output: map[string]any{"a": testVarFloat32P, "b": testVarFloat32P}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarFloat64P}, RunOutput: testVarFloat64P, Output: map[string]any{"a": testVarFloat64P, "b": testVarFloat64P}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarStringP}, RunOutput: testVarStringP, Output: map[string]any{"a": testVarStringP, "b": testVarStringP}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarFuncP}, RunOutput: testVarFuncP, Output: map[string]any{"a": testVarFuncP, "b": testVarFuncP}},

		{Script: `a = b`, Input: map[string]any{"b": testVarValueBool}, RunOutput: testVarValueBool, Output: map[string]any{"a": testVarValueBool, "b": testVarValueBool}},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueInt32}, RunOutput: testVarValueInt32, Output: map[string]any{"a": testVarValueInt32, "b": testVarValueInt32}},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueInt64}, RunOutput: testVarValueInt64, Output: map[string]any{"a": testVarValueInt64, "b": testVarValueInt64}},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueFloat32}, RunOutput: testVarValueFloat32, Output: map[string]any{"a": testVarValueFloat32, "b": testVarValueFloat32}},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueFloat64}, RunOutput: testVarValueFloat64, Output: map[string]any{"a": testVarValueFloat64, "b": testVarValueFloat64}},
		{Script: `a = b`, Input: map[string]any{"b": testVarValueString}, RunOutput: testVarValueString, Output: map[string]any{"a": testVarValueString, "b": testVarValueString}},

		{Script: `var a = b`, Input: map[string]any{"b": testVarValueBool}, RunOutput: testVarValueBool, Output: map[string]any{"a": testVarValueBool, "b": testVarValueBool}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueInt32}, RunOutput: testVarValueInt32, Output: map[string]any{"a": testVarValueInt32, "b": testVarValueInt32}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueInt64}, RunOutput: testVarValueInt64, Output: map[string]any{"a": testVarValueInt64, "b": testVarValueInt64}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueFloat32}, RunOutput: testVarValueFloat32, Output: map[string]any{"a": testVarValueFloat32, "b": testVarValueFloat32}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueFloat64}, RunOutput: testVarValueFloat64, Output: map[string]any{"a": testVarValueFloat64, "b": testVarValueFloat64}},
		{Script: `var a = b`, Input: map[string]any{"b": testVarValueString}, RunOutput: testVarValueString, Output: map[string]any{"a": testVarValueString, "b": testVarValueString}},

		// one variable spacing
		{Script: `
a  =  1
`, RunOutput: int64(1)},
		{Script: `
a  =  1;
`, RunOutput: int64(1)},

		// one variable many values
		{Script: `, b = 1, 2`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(2), Output: map[string]any{"b": int64(1)}},
		{Script: `var , b = 1, 2`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(2), Output: map[string]any{"b": int64(1)}},
		{Script: `a,  = 1, 2`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a,  = 1, 2`, ParseError: fmt.Errorf("syntax error")},

		// TOFIX: should not error?
		{Script: `a = 1, 2`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1)}},
		// TOFIX: should not error?
		{Script: `a = 1, 2, 3`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1)}},

		// two variables many values
		{Script: `a, b  = 1,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a, b  = 1,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `a, b  = ,1`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(1)},
		{Script: `var a, b  = ,1`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(1)},
		{Script: `a, b  = 1,,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a, b  = 1,,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `a, b  = ,1,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a, b  = ,1,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `a, b  = ,,1`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a, b  = ,,1`, ParseError: fmt.Errorf("syntax error")},

		{Script: `a.c, b = 1, 2`, RunError: fmt.Errorf("undefined symbol 'a'")},
		{Script: `a, b.c = 1, 2`, RunError: fmt.Errorf("undefined symbol 'b'")},

		{Script: `a, b = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `var a, b = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a, b = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}},
		{Script: `var a, b = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}},
		{Script: `a, b = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1), "b": int64(2)}},
		{Script: `var a, b = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1), "b": int64(2)}},

		// three variables many values
		{Script: `a, b, c = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `var a, b, c = 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a, b, c = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}},
		{Script: `var a, b, c = 1, 2`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}},
		{Script: `a, b, c = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)}},
		{Script: `var a, b, c = 1, 2, 3`, RunOutput: int64(3), Output: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)}},
		{Script: `a, b, c = 1, 2, 3, 4`, RunOutput: int64(4), Output: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)}},
		{Script: `var a, b, c = 1, 2, 3, 4`, RunOutput: int64(4), Output: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)}},

		// scope
		{Script: `func(){ a = 1 }(); a`, RunError: fmt.Errorf("undefined symbol 'a'")},
		{Script: `func(){ var a = 1 }(); a`, RunError: fmt.Errorf("undefined symbol 'a'")},

		{Script: `a = 1; func(){ a = 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `var a = 1; func(){ a = 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = 1; func(){ var a = 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(1)}},
		{Script: `var a = 1; func(){ var a = 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(1)}},

		// function return
		{Script: `a, 1++ = func(){ return 1, 2 }()`, RunError: fmt.Errorf("invalid operation"), Output: map[string]any{"a": int64(1)}},

		{Script: `a = func(){ return 1 }()`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `var a = func(){ return 1 }()`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a, b = func(){ return 1, 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}},
		{Script: `var a, b = func(){ return 1, 2 }()`, RunOutput: int64(2), Output: map[string]any{"a": int64(1), "b": int64(2)}},
	}
	runTests(t, tests, nil)
}

func TestModule(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `module a.b { }`, ParseError: fmt.Errorf("syntax error")},
		{Script: `module a { 1++ }`, RunError: fmt.Errorf("invalid operation")},
		{Script: `module a { }; a.b`, RunError: fmt.Errorf("invalid operation 'b'")},

		{Script: `module a { b = nil }; a.b`, RunOutput: nil},
		{Script: `module a { b = true }; a.b`, RunOutput: true},
		{Script: `module a { b = 1 }; a.b`, RunOutput: int64(1)},
		{Script: `module a { b = 1.1 }; a.b`, RunOutput: float64(1.1)},
		{Script: `module a { b = "a" }; a.b`, RunOutput: "a"},
	}
	runTests(t, tests, nil)
}

func TestNew(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `new(foo)`, RunError: fmt.Errorf("undefined type 'foo'")},
		{Script: `new(nilT)`, Types: map[string]any{"nilT": nil}, RunError: fmt.Errorf("type cannot be nil for new")},

		{Script: `a = new(bool); *a`, RunOutput: false},
		{Script: `a = new(int32); *a`, RunOutput: int32(0)},
		{Script: `a = new(int64); *a`, RunOutput: int64(0)},
		{Script: `a = new(float32); *a`, RunOutput: float32(0)},
		{Script: `a = new(float64); *a`, RunOutput: float64(0)},
		{Script: `a = new(string); *a`, RunOutput: ""},
	}
	runTests(t, tests, nil)
}

func TestMake(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make(foo)`, RunError: fmt.Errorf("undefined type 'foo'")},
		{Script: `make(a.b)`, Types: map[string]any{"a": true}, RunError: fmt.Errorf("no namespace called: a")},
		{Script: `make(a.b)`, Types: map[string]any{"b": true}, RunError: fmt.Errorf("no namespace called: a")},

		{Script: `make(nilT)`, Types: map[string]any{"nilT": nil}, RunError: fmt.Errorf("type cannot be nil for make")},

		{Script: `make(bool)`, RunOutput: false},
		{Script: `make(int32)`, RunOutput: int32(0)},
		{Script: `make(int64)`, RunOutput: int64(0)},
		{Script: `make(float32)`, RunOutput: float32(0)},
		{Script: `make(float64)`, RunOutput: float64(0)},
		{Script: `make(string)`, RunOutput: ""},
	}
	runTests(t, tests, nil)
}

func TestMakeType(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make(type a, 1++)`, RunError: fmt.Errorf("invalid operation")},

		{Script: `make(type a, true)`, RunOutput: reflect.TypeOf(true)},
		{Script: `a = make(type a, true)`, RunOutput: reflect.TypeOf(true), Output: map[string]any{"a": reflect.TypeOf(true)}},
		{Script: `make(type a, true); a = make([]a)`, RunOutput: []bool{}, Output: map[string]any{"a": []bool{}}},
		{Script: `make(type a, make([]bool))`, RunOutput: reflect.TypeOf([]bool{})},
		{Script: `make(type a, make([]bool)); a = make(a)`, RunOutput: []bool{}, Output: map[string]any{"a": []bool{}}},
	}
	runTests(t, tests, nil)
}

func TestReferencingAndDereference(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		// TOFIX:
		// {Script: `a = 1; b = &a; *b = 2; *b`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
	}
	runTests(t, tests, nil)
}

func TestMakeChan(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make(chan foobar, 2)`, RunError: fmt.Errorf("undefined type 'foobar'")},
		{Script: `make(chan nilT, 2)`, Types: map[string]any{"nilT": nil}, RunError: fmt.Errorf("type cannot be nil for make chan")},
		{Script: `make(chan bool, 1++)`, RunError: fmt.Errorf("invalid operation")},

		{Script: `a = make(chan bool); b = func (c) { c <- true }; go b(a); <- a`, RunOutput: true},
	}
	runTests(t, tests, nil)
}

func TestChan(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = make(chan bool, 2); 1++ <- 1`, RunError: fmt.Errorf("invalid operation")},
		{Script: `a = make(chan bool, 2); a <- 1++`, RunError: fmt.Errorf("invalid operation")},

		// TODO: move close from core into vm code, then update tests

		{Script: `1 <- 1`, RunError: fmt.Errorf("invalid operation for chan")},
		// TODO: this panics, should we capture the panic in a better way?
		// {Script: `a = make(chan bool, 2); close(a); a <- true`, Input: map[string]any{"close": func(b any) { reflect.ValueOf(b).Close() }}, RunError: fmt.Errorf("channel is close")},
		// TODO: add chan syntax ok
		// {Script: `a = make(chan bool, 2); a <- true; close(a); b, ok <- a; ok`, Input: map[string]any{"close": func(b any) { reflect.ValueOf(b).Close() }}, RunOutput: false, Output: map[string]any{"b": nil}},
		{Script: `a = make(chan bool, 2); a <- true; close(a); b = false; b <- a`, Input: map[string]any{"close": func(b any) { reflect.ValueOf(b).Close() }}, RunOutput: true, Output: map[string]any{"b": true}},
		// TOFIX: add chan syntax ok, do not return error. Also b should be nil
		{Script: `a = make(chan bool, 2); a <- true; close(a); b = false; b <- a; b <- a`, Input: map[string]any{"close": func(b any) { reflect.ValueOf(b).Close() }}, RunError: fmt.Errorf("failed to send to channel"), Output: map[string]any{"b": true}},

		{Script: `a = make(chan bool, 2); a <- 1`, RunError: fmt.Errorf("cannot use type int64 as type bool to send to chan")},

		{Script: `a = make(chan interface, 2); a <- nil; <- a`, RunOutput: nil},
		{Script: `a = make(chan bool, 2); a <- true; <- a`, RunOutput: true},
		{Script: `a = make(chan int32, 2); a <- 1; <- a`, RunOutput: int32(1)},
		{Script: `a = make(chan int64, 2); a <- 1; <- a`, RunOutput: int64(1)},
		{Script: `a = make(chan float32, 2); a <- 1.1; <- a`, RunOutput: float32(1.1)},
		{Script: `a = make(chan float64, 2); a <- 1.1; <- a`, RunOutput: float64(1.1)},

		{Script: `a <- nil; <- a`, Input: map[string]any{"a": make(chan any, 2)}, RunOutput: nil},
		{Script: `a <- true; <- a`, Input: map[string]any{"a": make(chan bool, 2)}, RunOutput: true},
		{Script: `a <- 1; <- a`, Input: map[string]any{"a": make(chan int32, 2)}, RunOutput: int32(1)},
		{Script: `a <- 1; <- a`, Input: map[string]any{"a": make(chan int64, 2)}, RunOutput: int64(1)},
		{Script: `a <- 1.1; <- a`, Input: map[string]any{"a": make(chan float32, 2)}, RunOutput: float32(1.1)},
		{Script: `a <- 1.1; <- a`, Input: map[string]any{"a": make(chan float64, 2)}, RunOutput: float64(1.1)},
		{Script: `a <- "b"; <- a`, Input: map[string]any{"a": make(chan string, 2)}, RunOutput: "b"},

		{Script: `a = make(chan bool, 2); a <- true; a <- <- a`, RunOutput: nil},
		{Script: `a = make(chan bool, 2); a <- true; a <- (<- a)`, RunOutput: nil},
		{Script: `a = make(chan bool, 2); a <- true; a <- <- a; <- a`, RunOutput: true},
		{Script: `a = make(chan bool, 2); a <- true; b = false; b <- a`, RunOutput: true, Output: map[string]any{"b": true}},
		// TOFIX: if variable is not created yet, should make variable instead of error
		// {Script: `a = make(chan bool, 2); a <- true; b <- a`, RunOutput: true, Output: map[string]any{"b": true}},
	}
	runTests(t, tests, nil)
}

func TestCloseChan(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = make(chan int64, 1); close(a); <- a`, RunOutput: int64(0)},
	}
	runTests(t, tests, nil)
}

func TestVMDelete(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `delete(1++)`, RunError: fmt.Errorf("invalid operation")},
		{Script: `delete("a", 1++)`, RunError: fmt.Errorf("invalid operation")},

		{Script: `a = 1; delete("a"); a`, RunError: fmt.Errorf("undefined symbol 'a'")},

		{Script: `delete("a")`},
		{Script: `delete("a", false)`},
		{Script: `delete("a", true)`},
		{Script: `delete("a", nil)`},

		{Script: `a = 1; func b() { delete("a") }; b()`, Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1; func b() { delete("a", false) }; b()`, Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1; func b() { delete("a", true) }; b(); a`, RunError: fmt.Errorf("undefined symbol 'a'")},
	}
	runTests(t, tests, nil)
}

func TestComment(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `# 1`},
		{Script: `# 1;`},
		{Script: `# 1 // 2`},
		{Script: `# 1 \n 2`},
		{Script: `# 1 # 2`},

		{Script: `1# 1`, RunOutput: int64(1)},
		{Script: `1# 1;`, RunOutput: int64(1)},
		{Script: `1# 1 // 2`, RunOutput: int64(1)},
		{Script: `1# 1 \n 2`, RunOutput: int64(1)},
		{Script: `1# 1 # 2`, RunOutput: int64(1)},

		{Script: `1
# 1`, RunOutput: int64(1)},
		{Script: `1
# 1;`, RunOutput: int64(1)},
		{Script: `1
# 1 // 2`, RunOutput: int64(1)},
		{Script: `1
# 1 \n 2`, RunOutput: int64(1)},
		{Script: `1
# 1 # 2`, RunOutput: int64(1)},

		{Script: `// 1`},
		{Script: `// 1;`},
		{Script: `// 1 // 2`},
		{Script: `// 1 \n 2`},
		{Script: `// 1 # 2`},

		{Script: `1// 1`, RunOutput: int64(1)},
		{Script: `1// 1;`, RunOutput: int64(1)},
		{Script: `1// 1 // 2`, RunOutput: int64(1)},
		{Script: `1// 1 \n 2`, RunOutput: int64(1)},
		{Script: `1// 1 # 2`, RunOutput: int64(1)},

		{Script: `1
// 1`, RunOutput: int64(1)},
		{Script: `1
// 1;`, RunOutput: int64(1)},
		{Script: `1
// 1 // 2`, RunOutput: int64(1)},
		{Script: `1
// 1 \n 2`, RunOutput: int64(1)},
		{Script: `1
// 1 # 2`, RunOutput: int64(1)},

		{Script: `/* 1 */`},
		{Script: `/* * 1 */`},
		{Script: `/* 1 * */`},
		{Script: `/** 1 */`},
		{Script: `/*** 1 */`},
		{Script: `/**** 1 */`},
		{Script: `/* 1 **/`},
		{Script: `/* 1 ***/`},
		{Script: `/* 1 ****/`},
		{Script: `/** 1 ****/`},
		{Script: `/*** 1 ****/`},
		{Script: `/**** 1 ****/`},

		{Script: `1/* 1 */`, RunOutput: int64(1)},
		{Script: `1/* * 1 */`, RunOutput: int64(1)},
		{Script: `1/* 1 * */`, RunOutput: int64(1)},
		{Script: `1/** 1 */`, RunOutput: int64(1)},
		{Script: `1/*** 1 */`, RunOutput: int64(1)},
		{Script: `1/**** 1 */`, RunOutput: int64(1)},
		{Script: `1/* 1 **/`, RunOutput: int64(1)},
		{Script: `1/* 1 ***/`, RunOutput: int64(1)},
		{Script: `1/* 1 ****/`, RunOutput: int64(1)},
		{Script: `1/** 1 ****/`, RunOutput: int64(1)},
		{Script: `1/*** 1 ****/`, RunOutput: int64(1)},
		{Script: `1/**** 1 ****/`, RunOutput: int64(1)},

		{Script: `/* 1 */1`, RunOutput: int64(1)},
		{Script: `/* * 1 */1`, RunOutput: int64(1)},
		{Script: `/* 1 * */1`, RunOutput: int64(1)},
		{Script: `/** 1 */1`, RunOutput: int64(1)},
		{Script: `/*** 1 */1`, RunOutput: int64(1)},
		{Script: `/**** 1 */1`, RunOutput: int64(1)},
		{Script: `/* 1 **/1`, RunOutput: int64(1)},
		{Script: `/* 1 ***/1`, RunOutput: int64(1)},
		{Script: `/* 1 ****/1`, RunOutput: int64(1)},
		{Script: `/** 1 ****/1`, RunOutput: int64(1)},
		{Script: `/*** 1 ****/1`, RunOutput: int64(1)},
		{Script: `/**** 1 ****/1`, RunOutput: int64(1)},

		{Script: `1
/* 1 */`, RunOutput: int64(1)},
		{Script: `1
/* * 1 */`, RunOutput: int64(1)},
		{Script: `1
/* 1 * */`, RunOutput: int64(1)},
		{Script: `1
/** 1 */`, RunOutput: int64(1)},
		{Script: `1
/*** 1 */`, RunOutput: int64(1)},
		{Script: `1
/**** 1 */`, RunOutput: int64(1)},
		{Script: `1
/* 1 **/`, RunOutput: int64(1)},
		{Script: `1
/* 1 ***/`, RunOutput: int64(1)},
		{Script: `1
/* 1 ****/`, RunOutput: int64(1)},
		{Script: `1
/** 1 ****/`, RunOutput: int64(1)},
		{Script: `1
/*** 1 ****/`, RunOutput: int64(1)},
		{Script: `1
/**** 1 ****/`, RunOutput: int64(1)},
	}
	runTests(t, tests, nil)
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
	if err == nil || err.Error() != ErrInterrupt.Error() {
		t.Errorf("execute error - received %#v - expected: %#v", err, ErrInterrupt)
	}
}

func TestTwoContextSameEnv(t *testing.T) {
	var waitGroup sync.WaitGroup
	env := envPkg.NewEnv()
	v := New(&Configs{Env: env})
	ctx1, cancel := context.WithCancel(context.Background())
	e := v.Executor()
	waitGroup.Add(1)
	go func() {
		_, err := e.Run(ctx1, "func myFn(a) { return 123; }; for { }")
		if err == nil || err.Error() != ErrInterrupt.Error() {
			t.Errorf("execute error - received %#v - expected: %#v", err, ErrInterrupt)
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
			if err == nil || err.Error() != ErrInterrupt.Error() {
				t.Errorf("execute error - received %#v - expected: %#v", err, ErrInterrupt)
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
		t.Errorf("execute error - received %#v - expected: %#v", err, ErrInterrupt)
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
	e := stmts[0].(*ast.LetsStmt)
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
	runTests(t, tests, nil)

	// Define value 'val' to struct Bar in VM
	ref = 10
	tests = []Test{
		{Script: "val.ValueReceiver()", Input: map[string]interface{}{"val": val}, RunOutput: 200},
		{Script: "val.PointerReceiver()", Input: map[string]interface{}{"val": val}, RunOutput: []interface{}{0, 0}},
	}
	runTests(t, tests, nil)
}

func TestURL(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	value, err := New(&Configs{DefineImport: true}).Executor().Run(nil, `
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

var testPackagesEnvSetupFunc = func(t *testing.T, env *envPkg.Env) { DefineImport(env) }

func TestDefineImport(t *testing.T) {
	value, err := New(&Configs{DefineImport: true}).Executor().Run(nil, `strings = import("strings"); strings.ToLower("TEST")`)
	if err != nil {
		t.Errorf("execute error - received: %v - expected: %v", err, nil)
	}
	if value != "test" {
		t.Errorf("execute value - received: %v - expected: %v", value, int(4))
	}
}

func TestDefineImportPackageNotFound(t *testing.T) {
	_ = os.Unsetenv("ANKO_DEBUG")
	value, err := New(&Configs{DefineImport: true}).Executor().Run(nil, `a = import("a")`)
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

	value, err := New(&Configs{DefineImport: true}).Executor().Run(nil, `a = import("testPackage")`)
	expectedError := "import Define error: unknown symbol 'bad.name'"
	if err == nil || err.Error() != expectedError {
		t.Errorf("execute error - received: %v - expected: %v", err, expectedError)
	}
	if value != nil {
		t.Errorf("execute value - received: %v - expected: %v", value, nil)
	}

	packages.Packages.Insert("testPackage", map[string]any{"Coverage": testing.Coverage})
	packages.PackageTypes.Insert("testPackage", map[string]any{"bad.name": int64(1)})

	value, err = New(&Configs{DefineImport: true}).Executor().Run(nil, `a = import("testPackage")`)
	expectedError = "import DefineType error: unknown symbol 'bad.name'"
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
	runTests(t, tests, &Options{DefineImport: true})
}

func TestSync(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `sync = import("sync"); once = make(sync.Once); a = []; func add() { a += "a" }; once.Do(add); once.Do(add); a`, RunOutput: []any{"a"}, Output: map[string]any{"a": []any{"a"}}},
		{Script: `sync = import("sync"); waitGroup = make(sync.WaitGroup); waitGroup.Add(2);  func done() { waitGroup.Done() }; go done(); go done(); waitGroup.Wait(); "a"`, RunOutput: "a"},
	}
	runTests(t, tests, &Options{DefineImport: true})
}

func TestStringsPkg(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `strings = import("strings"); a = " one two "; b = strings.TrimSpace(a)`, RunOutput: "one two", Output: map[string]any{"a": " one two ", "b": "one two"}},
		{Script: `strings = import("strings"); a = "a b c d"; b = strings.Split(a, " ")`, RunOutput: []string{"a", "b", "c", "d"}, Output: map[string]any{"a": "a b c d", "b": []string{"a", "b", "c", "d"}}},
		{Script: `strings = import("strings"); a = "a b c d"; b = strings.SplitN(a, " ", 3)`, RunOutput: []string{"a", "b", "c d"}, Output: map[string]any{"a": "a b c d", "b": []string{"a", "b", "c d"}}},
	}
	runTests(t, tests, &Options{DefineImport: true})
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
		{Script: `strconv = import("strconv"); a = true; b = strconv.FormatBool(a)`, RunOutput: "true", Output: map[string]any{"a": true, "b": "true"}},
		{Script: `strconv = import("strconv"); a = 1.1; b = strconv.FormatFloat(a, toRune("f"), -1, 64)`, Input: map[string]any{"toRune": toRune}, RunOutput: "1.1", Output: map[string]any{"a": float64(1.1), "b": "1.1"}},
		{Script: `strconv = import("strconv"); a = 1; b = strconv.FormatInt(a, 10)`, RunOutput: "1", Output: map[string]any{"a": int64(1), "b": "1"}},
		{Script: `strconv = import("strconv"); b = strconv.FormatInt(a, 10)`, Input: map[string]any{"a": uint64(1)}, RunOutput: "1", Output: map[string]any{"a": uint64(1), "b": "1"}},

		{Script: `strconv = import("strconv"); a = "true"; b, err = strconv.ParseBool(a); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "true", "b": true, "err": "<nil>"}},
		{Script: `strconv = import("strconv"); a = "2"; b, err = strconv.ParseBool(a); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseBool: parsing "2": invalid syntax`, Output: map[string]any{"a": "2", "b": false, "err": `strconv.ParseBool: parsing "2": invalid syntax`}},
		{Script: `strconv = import("strconv"); a = "1.1"; b, err = strconv.ParseFloat(a, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1.1", "b": float64(1.1), "err": "<nil>"}},
		{Script: `strconv = import("strconv"); a = "a"; b, err = strconv.ParseFloat(a, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseFloat: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": float64(0), "err": `strconv.ParseFloat: parsing "a": invalid syntax`}},
		{Script: `strconv = import("strconv"); a = "1"; b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1", "b": int64(1), "err": "<nil>"}},
		{Script: `strconv = import("strconv"); a = "1.1"; b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseInt: parsing "1.1": invalid syntax`, Output: map[string]any{"a": "1.1", "b": int64(0), "err": `strconv.ParseInt: parsing "1.1": invalid syntax`}},
		{Script: `strconv = import("strconv"); a = "a"; b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseInt: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": int64(0), "err": `strconv.ParseInt: parsing "a": invalid syntax`}},
		{Script: `strconv = import("strconv"); a = "1"; b, err = strconv.ParseUint(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1", "b": uint64(1), "err": "<nil>"}},
		{Script: `strconv = import("strconv"); a = "a"; b, err = strconv.ParseUint(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseUint: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": uint64(0), "err": `strconv.ParseUint: parsing "a": invalid syntax`}},

		{Script: `strconv = import("strconv"); a = "true"; var b, err = strconv.ParseBool(a); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "true", "b": true, "err": "<nil>"}},
		{Script: `strconv = import("strconv"); a = "2"; var b, err = strconv.ParseBool(a); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseBool: parsing "2": invalid syntax`, Output: map[string]any{"a": "2", "b": false, "err": `strconv.ParseBool: parsing "2": invalid syntax`}},
		{Script: `strconv = import("strconv"); a = "1.1"; var b, err = strconv.ParseFloat(a, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1.1", "b": float64(1.1), "err": "<nil>"}},
		{Script: `strconv = import("strconv"); a = "a"; var b, err = strconv.ParseFloat(a, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseFloat: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": float64(0), "err": `strconv.ParseFloat: parsing "a": invalid syntax`}},
		{Script: `strconv = import("strconv"); a = "1"; var b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1", "b": int64(1), "err": "<nil>"}},
		{Script: `strconv = import("strconv"); a = "1.1"; var b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseInt: parsing "1.1": invalid syntax`, Output: map[string]any{"a": "1.1", "b": int64(0), "err": `strconv.ParseInt: parsing "1.1": invalid syntax`}},
		{Script: `strconv = import("strconv"); a = "a"; var b, err = strconv.ParseInt(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseInt: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": int64(0), "err": `strconv.ParseInt: parsing "a": invalid syntax`}},
		{Script: `strconv = import("strconv"); a = "1"; var b, err = strconv.ParseUint(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: "<nil>", Output: map[string]any{"a": "1", "b": uint64(1), "err": "<nil>"}},
		{Script: `strconv = import("strconv"); a = "a"; var b, err = strconv.ParseUint(a, 10, 64); err = toString(err)`, Input: map[string]any{"toString": toString}, RunOutput: `strconv.ParseUint: parsing "a": invalid syntax`, Output: map[string]any{"a": "a", "b": uint64(0), "err": `strconv.ParseUint: parsing "a": invalid syntax`}},
	}
	runTests(t, tests, &Options{DefineImport: true})
}

func TestSort(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `sort = import("sort"); a = make([]int); a += [5, 3, 1, 4, 2]; sort.Ints(a); a`, RunOutput: []int{1, 2, 3, 4, 5}, Output: map[string]any{"a": []int{1, 2, 3, 4, 5}}},
		{Script: `sort = import("sort"); a = make([]float64); a += [5.5, 3.3, 1.1, 4.4, 2.2]; sort.Float64s(a); a`, RunOutput: []float64{1.1, 2.2, 3.3, 4.4, 5.5}, Output: map[string]any{"a": []float64{1.1, 2.2, 3.3, 4.4, 5.5}}},
		{Script: `sort = import("sort"); a = make([]string); a += ["e", "c", "a", "d", "b"]; sort.Strings(a); a`, RunOutput: []string{"a", "b", "c", "d", "e"}, Output: map[string]any{"a": []string{"a", "b", "c", "d", "e"}}},
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
			RunOutput: []any{"f", float64(1.1), "2", int64(3), "4.4", int64(5)}, Output: map[string]any{"a": []any{"f", float64(1.1), "2", int64(3), "4.4", int64(5)}}},
	}
	runTests(t, tests, &Options{DefineImport: true})
}

func TestRegexp(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^simple$"); re.MatchString("simple")`, RunOutput: true},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^simple$"); re.MatchString("no match")`, RunOutput: false},

		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^simple$", "b": "simple"}, RunOutput: true, Output: map[string]any{"a": "^simple$", "b": "simple"}},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^simple$", "b": "no match"}, RunOutput: false, Output: map[string]any{"a": "^simple$", "b": "no match"}},

		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.String()`, RunOutput: "^a\\.\\d+\\.b$"},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("a.1.b")`, RunOutput: true},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("a.22.b")`, RunOutput: true},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("a.333.b")`, RunOutput: true},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("no match")`, RunOutput: false},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile("^a\\.\\d+\\.b$"); re.MatchString("a+1+b")`, RunOutput: false},

		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.String()`, Input: map[string]any{"a": "^a\\.\\d+\\.b$"}, RunOutput: "^a\\.\\d+\\.b$", Output: map[string]any{"a": "^a\\.\\d+\\.b$"}},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.1.b"}, RunOutput: true, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.1.b"}},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.22.b"}, RunOutput: true, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.22.b"}},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.333.b"}, RunOutput: true, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a.333.b"}},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "no match"}, RunOutput: false, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "no match"}},
		{Script: `regexp = import("regexp"); re = regexp.MustCompile(a); re.MatchString(b)`, Input: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a+1+b"}, RunOutput: false, Output: map[string]any{"a": "^a\\.\\d+\\.b$", "b": "a+1+b"}},
	}
	runTests(t, tests, &Options{DefineImport: true})
}

func TestJson(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `json = import("encoding/json"); a = make(mapStringInterface); a["b"] = "b"; c, err = json.Marshal(a); err`, Types: map[string]any{"mapStringInterface": map[string]any{}}, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": []byte(`{"b":"b"}`)}},
		{Script: `json = import("encoding/json"); b = 1; err = json.Unmarshal(a, &b); err`, Input: map[string]any{"a": []byte(`{"b": "b"}`)}, Output: map[string]any{"a": []byte(`{"b": "b"}`), "b": map[string]any{"b": "b"}}},
		{Script: `json = import("encoding/json"); b = 1; err = json.Unmarshal(a, &b); err`, Input: map[string]any{"a": `{"b": "b"}`}, Output: map[string]any{"a": `{"b": "b"}`, "b": map[string]any{"b": "b"}}},
		{Script: `json = import("encoding/json"); b = 1; err = json.Unmarshal(a, &b); err`, Input: map[string]any{"a": `[["1", "2"],["3", "4"]]`}, Output: map[string]any{"a": `[["1", "2"],["3", "4"]]`, "b": []any{[]any{"1", "2"}, []any{"3", "4"}}}},
	}
	runTests(t, tests, &Options{DefineImport: true})
}

func TestBytes(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `bytes = import("bytes"); a = make(bytes.Buffer); n, err = a.WriteString("a"); if err != nil { return err }; n`, RunOutput: 1},
		{Script: `bytes = import("bytes"); a = make(bytes.Buffer); n, err = a.WriteString("a"); if err != nil { return err }; a.String()`, RunOutput: "a"},
	}
	runTests(t, tests, &Options{DefineImport: true})
}

func TestAnk(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "")
	tests := []Test{
		{Script: `load('testdata/testing.ank'); load('testdata/let.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/toString.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/op.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/func.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/len.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/for.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/switch.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/if.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/toBytes.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/toRunes.ank')`},
		{Script: `load('testdata/testing.ank'); load('testdata/chan.ank')`},
	}
	runTests(t, tests, &Options{DefineImport: true, ImportCore: true})
}

func TestKeys(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = {}; b = keys(a)`, RunOutput: []any{}, Output: map[string]any{"a": map[any]any{}}},
		{Script: `a = {"a": nil}; b = keys(a)`, RunOutput: []any{"a"}, Output: map[string]any{"a": map[any]any{"a": nil}}},
		{Script: `a = {"a": 1}; b = keys(a)`, RunOutput: []any{"a"}, Output: map[string]any{"a": map[any]any{"a": int64(1)}}},
	}
	runTests(t, tests, &Options{DefineImport: true, ImportCore: true})
}

func TestKindOf(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `kindOf(a)`, Input: map[string]any{"a": reflect.Value{}}, RunOutput: "struct", Output: map[string]any{"a": reflect.Value{}}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": nil}, RunOutput: "nil", Output: map[string]any{"a": nil}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": true}, RunOutput: "bool", Output: map[string]any{"a": true}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": int32(1)}, RunOutput: "int32", Output: map[string]any{"a": int32(1)}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": int64(1)}, RunOutput: "int64", Output: map[string]any{"a": int64(1)}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": float32(1.1)}, RunOutput: "float32", Output: map[string]any{"a": float32(1.1)}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": float64(1.1)}, RunOutput: "float64", Output: map[string]any{"a": float64(1.1)}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": "a"}, RunOutput: "string", Output: map[string]any{"a": "a"}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": 'a'}, RunOutput: "int32", Output: map[string]any{"a": 'a'}},

		{Script: `kindOf(a)`, Input: map[string]any{"a": any(nil)}, RunOutput: "nil", Output: map[string]any{"a": any(nil)}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(true)}, RunOutput: "bool", Output: map[string]any{"a": any(true)}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(int32(1))}, RunOutput: "int32", Output: map[string]any{"a": any(int32(1))}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(int64(1))}, RunOutput: "int64", Output: map[string]any{"a": any(int64(1))}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(float32(1))}, RunOutput: "float32", Output: map[string]any{"a": any(float32(1))}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any(float64(1))}, RunOutput: "float64", Output: map[string]any{"a": any(float64(1))}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": any("a")}, RunOutput: "string", Output: map[string]any{"a": any("a")}},

		{Script: `kindOf(a)`, Input: map[string]any{"a": []any{}}, RunOutput: "slice", Output: map[string]any{"a": []any{}}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": []any{nil}}, RunOutput: "slice", Output: map[string]any{"a": []any{nil}}},

		{Script: `kindOf(a)`, Input: map[string]any{"a": map[string]any{}}, RunOutput: "map", Output: map[string]any{"a": map[string]any{}}},
		{Script: `kindOf(a)`, Input: map[string]any{"a": map[string]any{"b": "b"}}, RunOutput: "map", Output: map[string]any{"a": map[string]any{"b": "b"}}},

		{Script: `a = make(interface); kindOf(a)`, RunOutput: "nil", Output: map[string]any{"a": any(nil)}},
	}
	runTests(t, tests, &Options{DefineImport: true, ImportCore: true})
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
	runTests(t, tests, &Options{DefineImport: true, ImportCore: true})
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
	runTests(t, tests, &Options{DefineImport: true, ImportCore: true})
}

func TestDefined(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "")
	tests := []Test{
		{Script: `var a = 1; defined("a")`, RunOutput: true},
		{Script: `defined("a")`, RunOutput: false},
		{Script: `func(){ var a = 1 }(); defined("a")`, RunOutput: false},
	}
	runTests(t, tests, &Options{DefineImport: true, ImportCore: true})
}

func TestToX(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `toBool(-2)`, RunOutput: false},
		{Script: `toBool(-1.5)`, RunOutput: false},
		{Script: `toBool(-1)`, RunOutput: false},
		{Script: `toBool(-0.4)`, RunOutput: false},
		{Script: `toBool(0)`, RunOutput: false},
		{Script: `toBool(0.4)`, RunOutput: true},
		{Script: `toBool(1)`, RunOutput: true},
		{Script: `toBool(1.5)`, RunOutput: true},
		{Script: `toBool(2)`, RunOutput: true},
		{Script: `toBool(true)`, RunOutput: true},
		{Script: `toBool(false)`, RunOutput: false},
		{Script: `toBool("true")`, RunOutput: true},
		{Script: `toBool("false")`, RunOutput: false},
		{Script: `toBool("yes")`, RunOutput: true},
		{Script: `toBool("ye")`, RunOutput: false},
		{Script: `toBool("y")`, RunOutput: true},
		{Script: `toBool("false")`, RunOutput: false},
		{Script: `toBool("f")`, RunOutput: false},
		{Script: `toBool("")`, RunOutput: false},
		{Script: `toBool(nil)`, RunOutput: false},
		{Script: `toBool({})`, RunOutput: false},
		{Script: `toBool([])`, RunOutput: false},
		{Script: `toBool([true])`, RunOutput: false},
		{Script: `toBool({"true": "true"})`, RunOutput: false},
		{Script: `toString(nil)`, RunOutput: "<nil>"},
		{Script: `toString("")`, RunOutput: ""},
		{Script: `toString(1)`, RunOutput: "1"},
		{Script: `toString(1.2)`, RunOutput: "1.2"},
		{Script: `toString(1/3)`, RunOutput: "0.3333333333333333"},
		{Script: `toString(false)`, RunOutput: "false"},
		{Script: `toString(true)`, RunOutput: "true"},
		{Script: `toString({})`, RunOutput: "map[]"},
		{Script: `toString({"foo": "bar"})`, RunOutput: "map[foo:bar]"},
		{Script: `toString([true,nil])`, RunOutput: "[true <nil>]"},
		{Script: `toString(toByteSlice("foo"))`, RunOutput: "foo"},
		{Script: `toInt(nil)`, RunOutput: int64(0)},
		{Script: `toInt(-2)`, RunOutput: int64(-2)},
		{Script: `toInt(-1.4)`, RunOutput: int64(-1)},
		{Script: `toInt(-1)`, RunOutput: int64(-1)},
		{Script: `toInt(0)`, RunOutput: int64(0)},
		{Script: `toInt(1)`, RunOutput: int64(1)},
		{Script: `toInt(1.4)`, RunOutput: int64(1)},
		{Script: `toInt(1.5)`, RunOutput: int64(1)},
		{Script: `toInt(1.9)`, RunOutput: int64(1)},
		{Script: `toInt(2)`, RunOutput: int64(2)},
		{Script: `toInt(2.1)`, RunOutput: int64(2)},
		{Script: `toInt("2")`, RunOutput: int64(2)},
		{Script: `toInt("2.1")`, RunOutput: int64(2)},
		{Script: `toInt(true)`, RunOutput: int64(1)},
		{Script: `toInt(false)`, RunOutput: int64(0)},
		{Script: `toInt({})`, RunOutput: int64(0)},
		{Script: `toInt([])`, RunOutput: int64(0)},
		{Script: `toFloat(nil)`, RunOutput: float64(0.0)},
		{Script: `toFloat(-2)`, RunOutput: float64(-2.0)},
		{Script: `toFloat(-1.4)`, RunOutput: float64(-1.4)},
		{Script: `toFloat(-1)`, RunOutput: float64(-1.0)},
		{Script: `toFloat(0)`, RunOutput: float64(0.0)},
		{Script: `toFloat(1)`, RunOutput: float64(1.0)},
		{Script: `toFloat(1.4)`, RunOutput: float64(1.4)},
		{Script: `toFloat(1.5)`, RunOutput: float64(1.5)},
		{Script: `toFloat(1.9)`, RunOutput: float64(1.9)},
		{Script: `toFloat(2)`, RunOutput: float64(2.0)},
		{Script: `toFloat(2.1)`, RunOutput: float64(2.1)},
		{Script: `toFloat("2")`, RunOutput: float64(2.0)},
		{Script: `toFloat("2.1")`, RunOutput: float64(2.1)},
		{Script: `toFloat(true)`, RunOutput: float64(1.0)},
		{Script: `toFloat(false)`, RunOutput: float64(0.0)},
		{Script: `toFloat({})`, RunOutput: float64(0.0)},
		{Script: `toFloat([])`, RunOutput: float64(0.0)},
		{Script: `toChar(0x1F431)`, RunOutput: "🐱"},
		{Script: `toChar(0)`, RunOutput: "\x00"},
		{Script: `toRune("")`, RunOutput: rune(0)},
		{Script: `toRune("🐱")`, RunOutput: rune(0x1F431)},
		{Script: `toBoolSlice(nil)`, RunOutput: []bool{}},
		{Script: `toBoolSlice(1)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type int64")},
		{Script: `toBoolSlice(1.2)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type float64")},
		{Script: `toBoolSlice(false)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type bool")},
		{Script: `toBoolSlice({})`, RunError: fmt.Errorf("function wants argument type []interface {} but received type map[interface {}]interface {}")},
		{Script: `toBoolSlice([])`, RunOutput: []bool{}},
		{Script: `toBoolSlice([nil])`, RunOutput: []bool{false}},
		{Script: `toBoolSlice([1])`, RunOutput: []bool{false}},
		{Script: `toBoolSlice([1.1])`, RunOutput: []bool{false}},
		{Script: `toBoolSlice([true])`, RunOutput: []bool{true}},
		{Script: `toBoolSlice([[]])`, RunOutput: []bool{false}},
		{Script: `toBoolSlice([{}])`, RunOutput: []bool{false}},
		{Script: `toIntSlice(nil)`, RunOutput: []int64{}},
		{Script: `toIntSlice(1)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type int64")},
		{Script: `toIntSlice(1.2)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type float64")},
		{Script: `toIntSlice(false)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type bool")},
		{Script: `toIntSlice({})`, RunError: fmt.Errorf("function wants argument type []interface {} but received type map[interface {}]interface {}")},
		{Script: `toIntSlice([])`, RunOutput: []int64{}},
		{Script: `toIntSlice([nil])`, RunOutput: []int64{0}},
		{Script: `toIntSlice([1])`, RunOutput: []int64{1}},
		{Script: `toIntSlice([1.1])`, RunOutput: []int64{1}},
		{Script: `toIntSlice([true])`, RunOutput: []int64{0}},
		{Script: `toIntSlice([[]])`, RunOutput: []int64{0}},
		{Script: `toIntSlice([{}])`, RunOutput: []int64{0}},
		{Script: `toFloatSlice(nil)`, RunOutput: []float64{}},
		{Script: `toFloatSlice(1)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type int64")},
		{Script: `toFloatSlice(1.2)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type float64")},
		{Script: `toFloatSlice(false)`, RunError: fmt.Errorf("function wants argument type []interface {} but received type bool")},
		{Script: `toFloatSlice({})`, RunError: fmt.Errorf("function wants argument type []interface {} but received type map[interface {}]interface {}")},
		{Script: `toFloatSlice([])`, RunOutput: []float64{}},
		{Script: `toFloatSlice([nil])`, RunOutput: []float64{0.0}},
		{Script: `toFloatSlice([1])`, RunOutput: []float64{1.0}},
		{Script: `toFloatSlice([1.1])`, RunOutput: []float64{1.1}},
		{Script: `toFloatSlice([true])`, RunOutput: []float64{0.0}},
		{Script: `toFloatSlice([[]])`, RunOutput: []float64{0.0}},
		{Script: `toFloatSlice([{}])`, RunOutput: []float64{0.0}},
		{Script: `toByteSlice(nil)`, RunOutput: []byte{}},
		{Script: `toByteSlice([])`, RunError: fmt.Errorf("function wants argument type string but received type []interface {}")},
		{Script: `toByteSlice(1)`, RunOutput: []byte{0x01}}, // FIXME?
		{Script: `toByteSlice(1.1)`, RunError: fmt.Errorf("function wants argument type string but received type float64")},
		{Script: `toByteSlice(true)`, RunError: fmt.Errorf("function wants argument type string but received type bool")},
		{Script: `toByteSlice("foo")`, RunOutput: []byte{'f', 'o', 'o'}},
		{Script: `toByteSlice("世界")`, RunOutput: []byte{0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c}},
		{Script: `toRuneSlice(nil)`, RunOutput: []rune{}},
		{Script: `toRuneSlice([])`, RunError: fmt.Errorf("function wants argument type string but received type []interface {}")},
		{Script: `toRuneSlice(1)`, RunOutput: []rune{0x01}}, // FIXME?
		{Script: `toRuneSlice(1.1)`, RunError: fmt.Errorf("function wants argument type string but received type float64")},
		{Script: `toRuneSlice(true)`, RunError: fmt.Errorf("function wants argument type string but received type bool")},
		{Script: `toRuneSlice("foo")`, RunOutput: []rune{'f', 'o', 'o'}},
		{Script: `toRuneSlice("世界")`, RunOutput: []rune{'世', '界'}},
		{Script: `toStringSlice([true,false,1])`, RunOutput: []string{"", "", "\x01"}}, // FIXME?
		{Script: `toDuration(nil)`, RunOutput: time.Duration(0)},
		{Script: `toDuration(0)`, RunOutput: time.Duration(0)},
		{Script: `toDuration(true)`, RunError: fmt.Errorf("function wants argument type int64 but received type bool")},
		{Script: `toDuration([])`, RunError: fmt.Errorf("function wants argument type int64 but received type []interface {}")},
		{Script: `toDuration({})`, RunError: fmt.Errorf("function wants argument type int64 but received type map[interface {}]interface {}")},
		{Script: `toDuration("")`, RunError: fmt.Errorf("function wants argument type int64 but received type string")},
		{Script: `toDuration("1s")`, RunError: fmt.Errorf("function wants argument type int64 but received type string")}, // TODO
		{Script: `toDuration(a)`, Input: map[string]any{"a": int64(time.Duration(123 * time.Minute))}, RunOutput: time.Duration(123 * time.Minute)},
		{Script: `toDuration(a)`, Input: map[string]any{"a": float64(time.Duration(123 * time.Minute))}, RunOutput: time.Duration(123 * time.Minute)},
		{Script: `toDuration(a)`, Input: map[string]any{"a": time.Duration(123 * time.Minute)}, RunOutput: time.Duration(123 * time.Minute)},
	}
	runTests(t, tests, &Options{DefineImport: true, ImportCore: true})
}

func TestAddPackage(t *testing.T) {
	// empty
	v := New(nil)
	_, _ = v.AddPackage("empty", map[string]any{}, map[string]any{})
	value, err := v.Executor().Run(nil, "empty")
	if err != nil {
		t.Errorf("AddPackage error - received: %v - expected: %v", err, nil)
	}
	val, _ := v.GetEnv().GetValue("empty")
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
		{Script: `func(x){return func(){return x}}(1)()`, RunOutput: int64(1)},
		{Script: `func(x){if true {return func(){return x}}}(1)()`, RunOutput: int64(1)},
		{Script: `func(x){if false {} else if true {return func(){return x}}}(1)()`, RunOutput: int64(1)},
		{Script: `func(x){if false {} else {return func(){return x}}}(1)()`, RunOutput: int64(1)},
		{Script: `func(x){for a = 1; a < 2; a++ { return func(){return x}}}(1)()`, RunOutput: int64(1)},
		{Script: `func(x){for a in [1] { return func(){return x}}}(1)()`, RunOutput: int64(1)},
		{Script: `func(x){for { return func(){return x}}}(1)()`, RunOutput: int64(1)},
		{Script: `module m { f=func(x){return func(){return x}} }; m.f(1)()`, RunOutput: int64(1)},
		{Script: `module m { f=func(x){return func(){return x}}(1)() }; m.f`, RunOutput: int64(1)},
	}
	runTests(t, tests, nil)
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

func TestCyclesExecuted(t *testing.T) {
	script := "a = 1; b = 2; if a == b { return a; }; return b"
	v := New(nil)
	e := v.Executor()
	_, _ = e.Run(nil, script)
	assert.Equal(t, int64(10), e.GetCycles())
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
	fifth := func() TestStruct {
		return TestStruct{}
	}
	sixth := func() TestInterface {
		return TestStruct{}
	}
	vmprint := func(args ...any) {
		fmt.Println(fmt.Sprint(args))
	}
	arr := func() []int {
		return []int{0, 1, 2}
	}
	arr2 := func() []TestInterface {
		return []TestInterface{}
	}
	newEnv := func() *VM {
		v := New(&Configs{DefineImport: true})
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
	assert.Error(t, utils.MustErr(v.GetEnv().Get("b")))
	e2 := v.Executor()
	_, _ = e2.Run(nil, "b = 2")
	assert.Error(t, utils.MustErr(v.GetEnv().Get("a")))
	assert.Equal(t, int64(2), utils.Must(e2.GetEnv().Get("b")).(int64))
}
