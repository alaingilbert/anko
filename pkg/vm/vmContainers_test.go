package vm

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/alaingilbert/anko/pkg/parser"
)

var (
	testArrayEmpty []any
	testArray      = []any{nil, true, int64(1), float64(1.1), "a"}
	testMapEmpty   map[any]any
	testMap        = map[any]any{"a": nil, "b": true, "c": int64(1), "d": float64(1.1), "e": "e"}
)

func TestArrays(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `[1++]`, RunError: fmt.Errorf("invalid operation")},
		{Script: `1++[0]`, RunError: fmt.Errorf("invalid operation")},

		{Script: `[]`, RunOutput: []any{}},
		{Script: `[nil]`, RunOutput: []any{nil}},
		{Script: `[true]`, RunOutput: []any{true}},
		{Script: `["a"]`, RunOutput: []any{"a"}},
		{Script: `[1]`, RunOutput: []any{int64(1)}},
		{Script: `[1.1]`, RunOutput: []any{float64(1.1)}},

		{Script: `a = []; a.b`, RunError: fmt.Errorf("type slice does not support member operation")},
		{Script: `a = []; a.b = 1`, RunError: fmt.Errorf("type slice does not support member operation")},

		{Script: `a = []`, RunOutput: []any{}, Output: map[string]any{"a": []any{}}},
		{Script: `a = [nil]`, RunOutput: []any{any(nil)}, Output: map[string]any{"a": []any{any(nil)}}},
		{Script: `a = [true]`, RunOutput: []any{true}, Output: map[string]any{"a": []any{true}}},
		{Script: `a = [1]`, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `a = [1.1]`, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": []any{float64(1.1)}}},
		{Script: `a = ["a"]`, RunOutput: []any{"a"}, Output: map[string]any{"a": []any{"a"}}},

		{Script: `a = [[]]`, RunOutput: []any{[]any{}}, Output: map[string]any{"a": []any{[]any{}}}},
		{Script: `a = [[nil]]`, RunOutput: []any{[]any{any(nil)}}, Output: map[string]any{"a": []any{[]any{any(nil)}}}},
		{Script: `a = [[true]]`, RunOutput: []any{[]any{true}}, Output: map[string]any{"a": []any{[]any{true}}}},
		{Script: `a = [[1]]`, RunOutput: []any{[]any{int64(1)}}, Output: map[string]any{"a": []any{[]any{int64(1)}}}},
		{Script: `a = [[1.1]]`, RunOutput: []any{[]any{float64(1.1)}}, Output: map[string]any{"a": []any{[]any{float64(1.1)}}}},
		{Script: `a = [["a"]]`, RunOutput: []any{[]any{"a"}}, Output: map[string]any{"a": []any{[]any{"a"}}}},

		{Script: `a = []; a += nil`, RunOutput: []any{nil}, Output: map[string]any{"a": []any{nil}}},
		{Script: `a = []; a += true`, RunOutput: []any{true}, Output: map[string]any{"a": []any{true}}},
		{Script: `a = []; a += 1`, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `a = []; a += 1.1`, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": []any{float64(1.1)}}},
		{Script: `a = []; a += "a"`, RunOutput: []any{"a"}, Output: map[string]any{"a": []any{"a"}}},

		{Script: `a = []; a += []`, RunOutput: []any{}, Output: map[string]any{"a": []any{}}},
		{Script: `a = []; a += [nil]`, RunOutput: []any{nil}, Output: map[string]any{"a": []any{nil}}},
		{Script: `a = []; a += [true]`, RunOutput: []any{true}, Output: map[string]any{"a": []any{true}}},
		{Script: `a = []; a += [1]`, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `a = []; a += [1.1]`, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": []any{float64(1.1)}}},
		{Script: `a = []; a += ["a"]`, RunOutput: []any{"a"}, Output: map[string]any{"a": []any{"a"}}},

		{Script: `a = [0]; a[0]++`, RunOutput: int64(1), Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `a = [[0]]; a[0][0]++`, RunOutput: int64(1), Output: map[string]any{"a": []any{[]any{int64(1)}}}},

		{Script: `a = [2]; a[0]--`, RunOutput: int64(1), Output: map[string]any{"a": []any{int64(1)}}},
		{Script: `a = [[2]]; a[0][0]--`, RunOutput: int64(1), Output: map[string]any{"a": []any{[]any{int64(1)}}}},

		{Script: `a`, Input: map[string]any{"a": []bool{}}, RunOutput: []bool{}, Output: map[string]any{"a": []bool{}}},
		{Script: `a`, Input: map[string]any{"a": []int32{}}, RunOutput: []int32{}, Output: map[string]any{"a": []int32{}}},
		{Script: `a`, Input: map[string]any{"a": []int64{}}, RunOutput: []int64{}, Output: map[string]any{"a": []int64{}}},
		{Script: `a`, Input: map[string]any{"a": []float32{}}, RunOutput: []float32{}, Output: map[string]any{"a": []float32{}}},
		{Script: `a`, Input: map[string]any{"a": []float64{}}, RunOutput: []float64{}, Output: map[string]any{"a": []float64{}}},
		{Script: `a`, Input: map[string]any{"a": []string{}}, RunOutput: []string{}, Output: map[string]any{"a": []string{}}},

		{Script: `a`, Input: map[string]any{"a": []bool{true, false}}, RunOutput: []bool{true, false}, Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: []int32{1, 2}, Output: map[string]any{"a": []int32{1, 2}}},
		{Script: `a`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: []int64{1, 2}, Output: map[string]any{"a": []int64{1, 2}}},
		{Script: `a`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: []float32{1.1, 2.2}, Output: map[string]any{"a": []float32{1.1, 2.2}}},
		{Script: `a`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: []float64{1.1, 2.2}, Output: map[string]any{"a": []float64{1.1, 2.2}}},
		{Script: `a`, Input: map[string]any{"a": []string{"a", "b"}}, RunOutput: []string{"a", "b"}, Output: map[string]any{"a": []string{"a", "b"}}},

		{Script: `a += true`, Input: map[string]any{"a": []bool{}}, RunOutput: []bool{true}, Output: map[string]any{"a": []bool{true}}},
		{Script: `a += 1`, Input: map[string]any{"a": []int32{}}, RunOutput: []int32{1}, Output: map[string]any{"a": []int32{1}}},
		{Script: `a += 1.1`, Input: map[string]any{"a": []int32{}}, RunOutput: []int32{1}, Output: map[string]any{"a": []int32{1}}},
		{Script: `a += 1`, Input: map[string]any{"a": []int64{}}, RunOutput: []int64{1}, Output: map[string]any{"a": []int64{1}}},
		{Script: `a += 1.1`, Input: map[string]any{"a": []int64{}}, RunOutput: []int64{1}, Output: map[string]any{"a": []int64{1}}},
		{Script: `a += 1`, Input: map[string]any{"a": []float32{}}, RunOutput: []float32{1}, Output: map[string]any{"a": []float32{1}}},
		{Script: `a += 1.1`, Input: map[string]any{"a": []float32{}}, RunOutput: []float32{1.1}, Output: map[string]any{"a": []float32{1.1}}},
		{Script: `a += 1`, Input: map[string]any{"a": []float64{}}, RunOutput: []float64{1}, Output: map[string]any{"a": []float64{1}}},
		{Script: `a += 1.1`, Input: map[string]any{"a": []float64{}}, RunOutput: []float64{1.1}, Output: map[string]any{"a": []float64{1.1}}},
		{Script: `a += "a"`, Input: map[string]any{"a": []string{}}, RunOutput: []string{"a"}, Output: map[string]any{"a": []string{"a"}}},
		{Script: `a += 97`, Input: map[string]any{"a": []string{}}, RunOutput: []string{"a"}, Output: map[string]any{"a": []string{"a"}}},

		{Script: `a[0]`, Input: map[string]any{"a": []bool{true, false}}, RunOutput: true, Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a[0]`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: int32(1), Output: map[string]any{"a": []int32{1, 2}}},
		{Script: `a[0]`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: int64(1), Output: map[string]any{"a": []int64{1, 2}}},
		{Script: `a[0]`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: float32(1.1), Output: map[string]any{"a": []float32{1.1, 2.2}}},
		{Script: `a[0]`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: float64(1.1), Output: map[string]any{"a": []float64{1.1, 2.2}}},
		{Script: `a[0]`, Input: map[string]any{"a": []string{"a", "b"}}, RunOutput: "a", Output: map[string]any{"a": []string{"a", "b"}}},

		{Script: `a[1]`, Input: map[string]any{"a": []bool{true, false}}, RunOutput: false, Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a[1]`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: int32(2), Output: map[string]any{"a": []int32{1, 2}}},
		{Script: `a[1]`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: int64(2), Output: map[string]any{"a": []int64{1, 2}}},
		{Script: `a[1]`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: float32(2.2), Output: map[string]any{"a": []float32{1.1, 2.2}}},
		{Script: `a[1]`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: float64(2.2), Output: map[string]any{"a": []float64{1.1, 2.2}}},
		{Script: `a[1]`, Input: map[string]any{"a": []string{"a", "b"}}, RunOutput: "b", Output: map[string]any{"a": []string{"a", "b"}}},

		{Script: `a[0]`, Input: map[string]any{"a": []bool{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []bool{}}},
		{Script: `a[0]`, Input: map[string]any{"a": []int32{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []int32{}}},
		{Script: `a[0]`, Input: map[string]any{"a": []int64{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []int64{}}},
		{Script: `a[0]`, Input: map[string]any{"a": []float32{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []float32{}}},
		{Script: `a[0]`, Input: map[string]any{"a": []float64{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []float64{}}},
		{Script: `a[0]`, Input: map[string]any{"a": []string{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []string{}}},

		{Script: `a[1] = true`, Input: map[string]any{"a": []bool{true, false}}, RunOutput: true, Output: map[string]any{"a": []bool{true, true}}},
		{Script: `a[1] = 3`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: int64(3), Output: map[string]any{"a": []int32{1, 3}}},
		{Script: `a[1] = 3`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: int64(3), Output: map[string]any{"a": []int64{1, 3}}},
		{Script: `a[1] = 3.3`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: float64(3.3), Output: map[string]any{"a": []float32{1.1, 3.3}}},
		{Script: `a[1] = 3.3`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: float64(3.3), Output: map[string]any{"a": []float64{1.1, 3.3}}},
		{Script: `a[1] = "c"`, Input: map[string]any{"a": []string{"a", "b"}}, RunOutput: "c", Output: map[string]any{"a": []string{"a", "c"}}},

		{Script: `a = []; a[0]`, RunError: fmt.Errorf("index out of range")},
		{Script: `a = []; a[-1]`, RunError: fmt.Errorf("index out of range")},
		{Script: `a = []; a[1] = 1`, RunError: fmt.Errorf("index out of range")},
		{Script: `a = []; a[-1] = 1`, RunError: fmt.Errorf("index out of range")},

		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": nil}, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": true}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": int(1)}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": int32(1)}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": float32(1.1)}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": float64(1.1)}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": "1"}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": "a"}, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"b": []any{int64(1), int64(2)}}},

		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": testVarBoolP}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": testVarInt32P}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": testVarInt64P}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": testVarFloat32P}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": testVarFloat64P}, RunOutput: int64(2), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]any{"a": testVarStringP}, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"b": []any{int64(1), int64(2)}}},

		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]any{"a": nil}, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]any{"a": true}, RunOutput: int64(3), Output: map[string]any{"b": []any{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]any{"a": int(1)}, RunOutput: int64(3), Output: map[string]any{"b": []any{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]any{"a": int32(1)}, RunOutput: int64(3), Output: map[string]any{"b": []any{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(3), Output: map[string]any{"b": []any{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]any{"a": float32(1.1)}, RunOutput: int64(3), Output: map[string]any{"b": []any{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]any{"a": float64(1.1)}, RunOutput: int64(3), Output: map[string]any{"b": []any{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]any{"a": "1"}, RunOutput: int64(3), Output: map[string]any{"b": []any{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]any{"a": "a"}, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"b": []any{int64(1), int64(2)}}},

		{Script: `a`, Input: map[string]any{"a": testArrayEmpty}, RunOutput: testArrayEmpty, Output: map[string]any{"a": testArrayEmpty}},
		{Script: `a += []`, Input: map[string]any{"a": testArrayEmpty}, RunOutput: []any(nil), Output: map[string]any{"a": testArrayEmpty}},

		{Script: `a`, Input: map[string]any{"a": testArray}, RunOutput: testArray, Output: map[string]any{"a": testArray}},
		{Script: `a[0]`, Input: map[string]any{"a": testArray}, RunOutput: nil, Output: map[string]any{"a": testArray}},
		{Script: `a[0] = 1`, Input: map[string]any{"a": testArray}, RunOutput: int64(1), Output: map[string]any{"a": testArray}},
		{Script: `a[0]`, Input: map[string]any{"a": testArray}, RunOutput: int64(1), Output: map[string]any{"a": testArray}},
		{Script: `a[0] = nil`, Input: map[string]any{"a": testArray}, RunOutput: nil, Output: map[string]any{"a": testArray}},
		{Script: `a[0]`, Input: map[string]any{"a": testArray}, RunOutput: nil, Output: map[string]any{"a": testArray}},

		{Script: `a[1]`, Input: map[string]any{"a": testArray}, RunOutput: true, Output: map[string]any{"a": testArray}},
		{Script: `a[1] = false`, Input: map[string]any{"a": testArray}, RunOutput: false, Output: map[string]any{"a": testArray}},
		{Script: `a[1]`, Input: map[string]any{"a": testArray}, RunOutput: false, Output: map[string]any{"a": testArray}},
		{Script: `a[1] = true`, Input: map[string]any{"a": testArray}, RunOutput: true, Output: map[string]any{"a": testArray}},
		{Script: `a[1]`, Input: map[string]any{"a": testArray}, RunOutput: true, Output: map[string]any{"a": testArray}},

		{Script: `a[2]`, Input: map[string]any{"a": testArray}, RunOutput: int64(1), Output: map[string]any{"a": testArray}},
		{Script: `a[2] = 2`, Input: map[string]any{"a": testArray}, RunOutput: int64(2), Output: map[string]any{"a": testArray}},
		{Script: `a[2]`, Input: map[string]any{"a": testArray}, RunOutput: int64(2), Output: map[string]any{"a": testArray}},
		{Script: `a[2] = 1`, Input: map[string]any{"a": testArray}, RunOutput: int64(1), Output: map[string]any{"a": testArray}},
		{Script: `a[2]`, Input: map[string]any{"a": testArray}, RunOutput: int64(1), Output: map[string]any{"a": testArray}},

		{Script: `a[3]`, Input: map[string]any{"a": testArray}, RunOutput: float64(1.1), Output: map[string]any{"a": testArray}},
		{Script: `a[3] = 2.2`, Input: map[string]any{"a": testArray}, RunOutput: float64(2.2), Output: map[string]any{"a": testArray}},
		{Script: `a[3]`, Input: map[string]any{"a": testArray}, RunOutput: float64(2.2), Output: map[string]any{"a": testArray}},
		{Script: `a[3] = 1.1`, Input: map[string]any{"a": testArray}, RunOutput: float64(1.1), Output: map[string]any{"a": testArray}},
		{Script: `a[3]`, Input: map[string]any{"a": testArray}, RunOutput: float64(1.1), Output: map[string]any{"a": testArray}},

		{Script: `a[4]`, Input: map[string]any{"a": testArray}, RunOutput: "a", Output: map[string]any{"a": testArray}},
		{Script: `a[4] = "x"`, Input: map[string]any{"a": testArray}, RunOutput: "x", Output: map[string]any{"a": testArray}},
		{Script: `a[4]`, Input: map[string]any{"a": testArray}, RunOutput: "x", Output: map[string]any{"a": testArray}},
		{Script: `a[4] = "a"`, Input: map[string]any{"a": testArray}, RunOutput: "a", Output: map[string]any{"a": testArray}},
		{Script: `a[4]`, Input: map[string]any{"a": testArray}, RunOutput: "a", Output: map[string]any{"a": testArray}},

		{Script: `a[0][0] = true`, Input: map[string]any{"a": []any{[]string{"a"}}}, RunError: fmt.Errorf("type bool cannot be assigned to type string for array index"), Output: map[string]any{"a": []any{[]string{"a"}}}},
		{Script: `a[0][0] = "a"`, Input: map[string]any{"a": []any{[]bool{true}}}, RunError: fmt.Errorf("type string cannot be assigned to type bool for array index"), Output: map[string]any{"a": []any{[]bool{true}}}},

		{Script: `a[0][0] = b[0][0]`, Input: map[string]any{"a": []any{[]bool{true}}, "b": []any{[]string{"b"}}}, RunError: fmt.Errorf("type string cannot be assigned to type bool for array index"), Output: map[string]any{"a": []any{[]bool{true}}}},
		{Script: `b[0][0] = a[0][0]`, Input: map[string]any{"a": []any{[]bool{true}}, "b": []any{[]string{"b"}}}, RunError: fmt.Errorf("type bool cannot be assigned to type string for array index"), Output: map[string]any{"a": []any{[]bool{true}}}},

		{Script: `a = make([][]bool); a[0] =  make([]bool); a[0] = [true, 1]`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": [][]bool{{}}}},
		{Script: `a = make([][]bool); a[0] =  make([]bool); a[0] = [true, false]`, RunOutput: []any{true, false}, Output: map[string]any{"a": [][]bool{{true, false}}}},

		{Script: `a = make([][][]bool); a[0] = make([][]bool); a[0][0] = make([]bool); a[0] = [[true, 1]]`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": [][][]bool{{{}}}}},
		{Script: `a = make([][][]bool); a[0] = make([][]bool); a[0][0] = make([]bool); a[0] = [[true, false]]`, RunOutput: []any{[]any{true, false}}, Output: map[string]any{"a": [][][]bool{{{true, false}}}}},
	}
	runTests(t, tests, nil)
}

func TestArraysAutoAppend(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a[2]`, Input: map[string]any{"a": []bool{true, false}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a[2]`, Input: map[string]any{"a": []int32{1, 2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []int32{1, 2}}},
		{Script: `a[2]`, Input: map[string]any{"a": []int64{1, 2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []int64{1, 2}}},
		{Script: `a[2]`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []float32{1.1, 2.2}}},
		{Script: `a[2]`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []float64{1.1, 2.2}}},
		{Script: `a[2]`, Input: map[string]any{"a": []string{"a", "b"}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []string{"a", "b"}}},

		{Script: `a[2] = true`, Input: map[string]any{"a": []bool{true, false}}, RunOutput: true, Output: map[string]any{"a": []bool{true, false, true}}},
		{Script: `a[2] = 3`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: int64(3), Output: map[string]any{"a": []int32{1, 2, 3}}},
		{Script: `a[2] = 3`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: int64(3), Output: map[string]any{"a": []int64{1, 2, 3}}},
		{Script: `a[2] = 3.3`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: float64(3.3), Output: map[string]any{"a": []float32{1.1, 2.2, 3.3}}},
		{Script: `a[2] = 3.3`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: float64(3.3), Output: map[string]any{"a": []float64{1.1, 2.2, 3.3}}},
		{Script: `a[2] = "c"`, Input: map[string]any{"a": []string{"a", "b"}}, RunOutput: "c", Output: map[string]any{"a": []string{"a", "b", "c"}}},

		{Script: `a[2] = 3.3`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: float64(3.3), Output: map[string]any{"a": []int32{1, 2, 3}}},
		{Script: `a[2] = 3.3`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: float64(3.3), Output: map[string]any{"a": []int64{1, 2, 3}}},

		{Script: `a[3] = true`, Input: map[string]any{"a": []bool{true, false}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a[3] = 4`, Input: map[string]any{"a": []int32{1, 2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []int32{1, 2}}},
		{Script: `a[3] = 4`, Input: map[string]any{"a": []int64{1, 2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []int64{1, 2}}},
		{Script: `a[3] = 4.4`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []float32{1.1, 2.2}}},
		{Script: `a[3] = 4.4`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []float64{1.1, 2.2}}},
		{Script: `a[3] = "d"`, Input: map[string]any{"a": []string{"a", "b"}}, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []string{"a", "b"}}},

		{Script: `a[2] = nil`, Input: map[string]any{"a": []bool{true, false}}, RunError: fmt.Errorf("type interface {} cannot be assigned to type bool for array index"), Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a[2] = nil`, Input: map[string]any{"a": []int32{1, 2}}, RunError: fmt.Errorf("type interface {} cannot be assigned to type int32 for array index"), Output: map[string]any{"a": []int32{1, 2}}},
		{Script: `a[2] = "a"`, Input: map[string]any{"a": []int32{1, 2}}, RunError: fmt.Errorf("type string cannot be assigned to type int32 for array index"), Output: map[string]any{"a": []int32{1, 2}}},
		{Script: `a[2] = true`, Input: map[string]any{"a": []int64{1, 2}}, RunError: fmt.Errorf("type bool cannot be assigned to type int64 for array index"), Output: map[string]any{"a": []int64{1, 2}}},
		{Script: `a[2] = "a"`, Input: map[string]any{"a": []int64{1, 2}}, RunError: fmt.Errorf("type string cannot be assigned to type int64 for array index"), Output: map[string]any{"a": []int64{1, 2}}},
		{Script: `a[2] = true`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunError: fmt.Errorf("type bool cannot be assigned to type float32 for array index"), Output: map[string]any{"a": []float32{1.1, 2.2}}},
		{Script: `a[2] = "a"`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunError: fmt.Errorf("type string cannot be assigned to type float64 for array index"), Output: map[string]any{"a": []float64{1.1, 2.2}}},
		{Script: `a[2] = nil`, Input: map[string]any{"a": []string{"a", "b"}}, RunError: fmt.Errorf("type interface {} cannot be assigned to type string for array index"), Output: map[string]any{"a": []string{"a", "b"}}},
		{Script: `a[2] = true`, Input: map[string]any{"a": []string{"a", "b"}}, RunError: fmt.Errorf("type bool cannot be assigned to type string for array index"), Output: map[string]any{"a": []string{"a", "b"}}},
		{Script: `a[2] = 1.1`, Input: map[string]any{"a": []string{"a", "b"}}, RunError: fmt.Errorf("type float64 cannot be assigned to type string for array index"), Output: map[string]any{"a": []string{"a", "b"}}},

		{Script: `a`, Input: map[string]any{"a": [][]any{}}, RunOutput: [][]any{}, Output: map[string]any{"a": [][]any{}}},
		{Script: `a[0] = []`, Input: map[string]any{"a": [][]any{}}, RunOutput: []any{}, Output: map[string]any{"a": [][]any{{}}}},
		{Script: `a[0] = [nil]`, Input: map[string]any{"a": [][]any{}}, RunOutput: []any{nil}, Output: map[string]any{"a": [][]any{{nil}}}},
		{Script: `a[0] = [true]`, Input: map[string]any{"a": [][]any{}}, RunOutput: []any{true}, Output: map[string]any{"a": [][]any{{true}}}},
		{Script: `a[0] = [1]`, Input: map[string]any{"a": [][]any{}}, RunOutput: []any{int64(1)}, Output: map[string]any{"a": [][]any{{int64(1)}}}},
		{Script: `a[0] = [1.1]`, Input: map[string]any{"a": [][]any{}}, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": [][]any{{float64(1.1)}}}},
		{Script: `a[0] = ["b"]`, Input: map[string]any{"a": [][]any{}}, RunOutput: []any{"b"}, Output: map[string]any{"a": [][]any{{"b"}}}},

		{Script: `a[0] = [nil]; a[0][0]`, Input: map[string]any{"a": [][]any{}}, RunOutput: nil, Output: map[string]any{"a": [][]any{{nil}}}},
		{Script: `a[0] = [true]; a[0][0]`, Input: map[string]any{"a": [][]any{}}, RunOutput: true, Output: map[string]any{"a": [][]any{{true}}}},
		{Script: `a[0] = [1]; a[0][0]`, Input: map[string]any{"a": [][]any{}}, RunOutput: int64(1), Output: map[string]any{"a": [][]any{{int64(1)}}}},
		{Script: `a[0] = [1.1]; a[0][0]`, Input: map[string]any{"a": [][]any{}}, RunOutput: float64(1.1), Output: map[string]any{"a": [][]any{{float64(1.1)}}}},
		{Script: `a[0] = ["b"]; a[0][0]`, Input: map[string]any{"a": [][]any{}}, RunOutput: "b", Output: map[string]any{"a": [][]any{{"b"}}}},

		{Script: `a = make([]bool); a[0] = 1`, RunError: fmt.Errorf("type int64 cannot be assigned to type bool for array index"), Output: map[string]any{"a": []bool{}}},
		{Script: `a = make([]bool); a[0] = true; a[1] = false`, RunOutput: false, Output: map[string]any{"a": []bool{true, false}}},

		{Script: `a = make([][]bool); a[0] = [true, 1]`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": [][]bool{}}},
		{Script: `a = make([][]bool); a[0] = [true, false]`, RunOutput: []any{true, false}, Output: map[string]any{"a": [][]bool{{true, false}}}},

		{Script: `a = make([][][]bool); a[0] = [[true, 1]]`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": [][][]bool{}}},
		{Script: `a = make([][][]bool); a[0] = [[true, false]]`, RunOutput: []any{[]any{true, false}}, Output: map[string]any{"a": [][][]bool{{{true, false}}}}},
	}
	runTests(t, tests, nil)
}

func TestMakeArrays(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make([]foo)`, RunError: fmt.Errorf("undefined type 'foo'")},

		{Script: `make([]nilT)`, Types: map[string]any{"nilT": nil}, RunError: fmt.Errorf("type cannot be nil for make")},
		{Script: `make([][]nilT)`, Types: map[string]any{"nilT": nil}, RunError: fmt.Errorf("type cannot be nil for make")},
		{Script: `make([][][]nilT)`, Types: map[string]any{"nilT": nil}, RunError: fmt.Errorf("type cannot be nil for make")},

		{Script: `make([]bool, 1++)`, RunError: fmt.Errorf("invalid operation")},
		{Script: `make([]bool, 0, 1++)`, RunError: fmt.Errorf("invalid operation")},

		{Script: `make(array2x)`, Types: map[string]any{"array2x": [][]any{}}, RunOutput: [][]any{}},

		{Script: `make([]bool)`, RunOutput: []bool{}},
		{Script: `make([]int32)`, RunOutput: []int32{}},
		{Script: `make([]int64)`, RunOutput: []int64{}},
		{Script: `make([]float32)`, RunOutput: []float32{}},
		{Script: `make([]float64)`, RunOutput: []float64{}},
		{Script: `make([]string)`, RunOutput: []string{}},

		{Script: `make([]bool, 0)`, RunOutput: []bool{}},
		{Script: `make([]int32, 0)`, RunOutput: []int32{}},
		{Script: `make([]int64, 0)`, RunOutput: []int64{}},
		{Script: `make([]float32, 0)`, RunOutput: []float32{}},
		{Script: `make([]float64, 0)`, RunOutput: []float64{}},
		{Script: `make([]string, 0)`, RunOutput: []string{}},

		{Script: `make([]bool, 2)`, RunOutput: []bool{false, false}},
		{Script: `make([]int32, 2)`, RunOutput: []int32{int32(0), int32(0)}},
		{Script: `make([]int64, 2)`, RunOutput: []int64{int64(0), int64(0)}},
		{Script: `make([]float32, 2)`, RunOutput: []float32{float32(0), float32(0)}},
		{Script: `make([]float64, 2)`, RunOutput: []float64{float64(0), float64(0)}},
		{Script: `make([]string, 2)`, RunOutput: []string{"", ""}},

		{Script: `make([]bool, 0, 2)`, RunOutput: []bool{}},
		{Script: `make([]int32, 0, 2)`, RunOutput: []int32{}},
		{Script: `make([]int64, 0, 2)`, RunOutput: []int64{}},
		{Script: `make([]float32, 0, 2)`, RunOutput: []float32{}},
		{Script: `make([]float64, 0, 2)`, RunOutput: []float64{}},
		{Script: `make([]string, 0, 2)`, RunOutput: []string{}},

		{Script: `make([]bool, 2, 2)`, RunOutput: []bool{false, false}},
		{Script: `make([]int32, 2, 2)`, RunOutput: []int32{int32(0), int32(0)}},
		{Script: `make([]int64, 2, 2)`, RunOutput: []int64{int64(0), int64(0)}},
		{Script: `make([]float32, 2, 2)`, RunOutput: []float32{float32(0), float32(0)}},
		{Script: `make([]float64, 2, 2)`, RunOutput: []float64{float64(0), float64(0)}},
		{Script: `make([]string, 2, 2)`, RunOutput: []string{"", ""}},

		{Script: `a = make([]bool, 0); a += true; a += false`, RunOutput: []bool{true, false}, Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a = make([]int32, 0); a += 1; a += 2`, RunOutput: []int32{int32(1), int32(2)}, Output: map[string]any{"a": []int32{int32(1), int32(2)}}},
		{Script: `a = make([]int64, 0); a += 1; a += 2`, RunOutput: []int64{int64(1), int64(2)}, Output: map[string]any{"a": []int64{int64(1), int64(2)}}},
		{Script: `a = make([]float32, 0); a += 1.1; a += 2.2`, RunOutput: []float32{float32(1.1), float32(2.2)}, Output: map[string]any{"a": []float32{float32(1.1), float32(2.2)}}},
		{Script: `a = make([]float64, 0); a += 1.1; a += 2.2`, RunOutput: []float64{float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float64{float64(1.1), float64(2.2)}}},
		{Script: `a = make([]string, 0); a += "a"; a += "b"`, RunOutput: []string{"a", "b"}, Output: map[string]any{"a": []string{"a", "b"}}},

		{Script: `a = make([]bool, 2); a[0] = true; a[1] = false`, RunOutput: false, Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a = make([]int32, 2); a[0] = 1; a[1] = 2`, RunOutput: int64(2), Output: map[string]any{"a": []int32{int32(1), int32(2)}}},
		{Script: `a = make([]int64, 2); a[0] = 1; a[1] = 2`, RunOutput: int64(2), Output: map[string]any{"a": []int64{int64(1), int64(2)}}},
		{Script: `a = make([]float32, 2); a[0] = 1.1; a[1] = 2.2`, RunOutput: float64(2.2), Output: map[string]any{"a": []float32{float32(1.1), float32(2.2)}}},
		{Script: `a = make([]float64, 2); a[0] = 1.1; a[1] = 2.2`, RunOutput: float64(2.2), Output: map[string]any{"a": []float64{float64(1.1), float64(2.2)}}},
		{Script: `a = make([]string, 2); a[0] = "a"; a[1] = "b"`, RunOutput: "b", Output: map[string]any{"a": []string{"a", "b"}}},

		{Script: `make([]boolA)`, Types: map[string]any{"boolA": []bool{}}, RunOutput: [][]bool{}},
		{Script: `make([]int32A)`, Types: map[string]any{"int32A": []int32{}}, RunOutput: [][]int32{}},
		{Script: `make([]int64A)`, Types: map[string]any{"int64A": []int64{}}, RunOutput: [][]int64{}},
		{Script: `make([]float32A)`, Types: map[string]any{"float32A": []float32{}}, RunOutput: [][]float32{}},
		{Script: `make([]float64A)`, Types: map[string]any{"float64A": []float64{}}, RunOutput: [][]float64{}},
		{Script: `make([]stringA)`, Types: map[string]any{"stringA": []string{}}, RunOutput: [][]string{}},

		{Script: `make([]array)`, Types: map[string]any{"array": []any{}}, RunOutput: [][]any{}},
		{Script: `a = make([]array)`, Types: map[string]any{"array": []any{}}, RunOutput: [][]any{}, Output: map[string]any{"a": [][]any{}}},

		{Script: `make([][]bool)`, RunOutput: [][]bool{}},
		{Script: `make([][]int32)`, RunOutput: [][]int32{}},
		{Script: `make([][]int64)`, RunOutput: [][]int64{}},
		{Script: `make([][]float32)`, RunOutput: [][]float32{}},
		{Script: `make([][]float64)`, RunOutput: [][]float64{}},
		{Script: `make([][]string)`, RunOutput: [][]string{}},

		{Script: `make([][]bool)`, RunOutput: [][]bool{}},
		{Script: `make([][]int32)`, RunOutput: [][]int32{}},
		{Script: `make([][]int64)`, RunOutput: [][]int64{}},
		{Script: `make([][]float32)`, RunOutput: [][]float32{}},
		{Script: `make([][]float64)`, RunOutput: [][]float64{}},
		{Script: `make([][]string)`, RunOutput: [][]string{}},

		{Script: `make([][][]bool)`, RunOutput: [][][]bool{}},
		{Script: `make([][][]int32)`, RunOutput: [][][]int32{}},
		{Script: `make([][][]int64)`, RunOutput: [][][]int64{}},
		{Script: `make([][][]float32)`, RunOutput: [][][]float32{}},
		{Script: `make([][][]float64)`, RunOutput: [][][]float64{}},
		{Script: `make([][][]string)`, RunOutput: [][][]string{}},
	}
	runTests(t, tests, nil)
}

func TestArraySlice(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = [1, 2]; a[:]`, ParseError: fmt.Errorf("syntax error")},
		{Script: `(1++)[0:0]`, RunError: fmt.Errorf("invalid operation")},
		{Script: `a = [1, 2]; a[1++:0]`, RunError: fmt.Errorf("invalid operation"), Output: map[string]any{"a": []any{int64(1), int64(2)}}},
		{Script: `a = [1, 2]; a[0:1++]`, RunError: fmt.Errorf("invalid operation"), Output: map[string]any{"a": []any{int64(1), int64(2)}}},
		{Script: `a = [1, 2]; a[:0]++`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]any{"a": []any{int64(1), int64(2)}}},
		{Script: `a = [1, 2]; a[:0]--`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]any{"a": []any{int64(1), int64(2)}}},

		{Script: `a = [1, 2, 3]; a[nil:2]`, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:nil]`, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[-1:0]`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:0]`, RunOutput: []any{}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:1]`, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:2]`, RunOutput: []any{int64(1), int64(2)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:3]`, RunOutput: []any{int64(1), int64(2), int64(3)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[1:0]`, RunError: fmt.Errorf("invalid slice index"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:1]`, RunOutput: []any{}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:2]`, RunOutput: []any{int64(2)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:3]`, RunOutput: []any{int64(2), int64(3)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[2:1]`, RunError: fmt.Errorf("invalid slice index"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[2:2]`, RunOutput: []any{}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[2:3]`, RunOutput: []any{int64(3)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[2:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[3:2]`, RunError: fmt.Errorf("invalid slice index"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[3:3]`, RunOutput: []any{}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[3:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[4:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[-1:]`, RunError: fmt.Errorf("index out of range"), RunOutput: nil, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:]`, RunOutput: []any{int64(1), int64(2), int64(3)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:]`, RunOutput: []any{int64(2), int64(3)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[2:]`, RunOutput: []any{int64(3)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[3:]`, RunOutput: []any{}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[4:]`, RunError: fmt.Errorf("index out of range"), RunOutput: nil, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[:-1]`, RunError: fmt.Errorf("index out of range"), RunOutput: nil, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:0]`, RunOutput: []any{}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:1]`, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:2]`, RunOutput: []any{int64(1), int64(2)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:3]`, RunOutput: []any{int64(1), int64(2), int64(3)}, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:4]`, RunError: fmt.Errorf("index out of range"), RunOutput: nil, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `b[1:2] = 4`, RunError: fmt.Errorf("undefined symbol 'b'")},
		{Script: `a = [1, 2, 3]; a[1++:2] = 4`, RunError: fmt.Errorf("invalid operation"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:2++] = 4`, RunError: fmt.Errorf("invalid operation"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[nil:2] = 4`, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:nil] = 4`, RunError: fmt.Errorf("index must be a number"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[0:0] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:1] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:4] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:0] = 4`, RunError: fmt.Errorf("invalid slice index"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:4] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[-1:] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[4:] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[:-1] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:0] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:1] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:4] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}},

		{Script: `a = [{"b": "b"}, {"c": "c"}, {"d": "d"}]; a[0:2].a`, RunError: fmt.Errorf("type slice does not support member operation"), Output: map[string]any{"a": []any{map[any]any{"b": "b"}, map[any]any{"c": "c"}, map[any]any{"d": "d"}}}},

		{Script: `a = [{"b": "b"}, {"c": "c"}, {"d": "d"}]; a[0:2]`, RunOutput: []any{map[any]any{"b": "b"}, map[any]any{"c": "c"}}, Output: map[string]any{"a": []any{map[any]any{"b": "b"}, map[any]any{"c": "c"}, map[any]any{"d": "d"}}}},
		{Script: `a = [{"b": "b"}, {"c": "c"}, {"d": "d"}]; a[0:2][0].b`, RunOutput: "b", Output: map[string]any{"a": []any{map[any]any{"b": "b"}, map[any]any{"c": "c"}, map[any]any{"d": "d"}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:3]`, RunOutput: []any{int64(1), int64(2), int64(3)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][1:3]`, RunOutput: []any{int64(2), int64(3)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][2:3]`, RunOutput: []any{int64(3)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][3:3]`, RunOutput: []any{}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:]`, RunOutput: []any{int64(1), int64(2), int64(3)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][1:]`, RunOutput: []any{int64(2), int64(3)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][2:]`, RunOutput: []any{int64(3)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][3:]`, RunOutput: []any{}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:0]`, RunOutput: []any{}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:1]`, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:2]`, RunOutput: []any{int64(1), int64(2)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:3]`, RunOutput: []any{int64(1), int64(2), int64(3)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][:0]`, RunOutput: []any{}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][:1]`, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][:2]`, RunOutput: []any{int64(1), int64(2)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][:3]`, RunOutput: []any{int64(1), int64(2), int64(3)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:3]`, RunOutput: []any{int64(4), int64(5), int64(6)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][1:3]`, RunOutput: []any{int64(5), int64(6)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][2:3]`, RunOutput: []any{int64(6)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][3:3]`, RunOutput: []any{}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:]`, RunOutput: []any{int64(4), int64(5), int64(6)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][1:]`, RunOutput: []any{int64(5), int64(6)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][2:]`, RunOutput: []any{int64(6)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][3:]`, RunOutput: []any{}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:0]`, RunOutput: []any{}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:1]`, RunOutput: []any{int64(4)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:2]`, RunOutput: []any{int64(4), int64(5)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:3]`, RunOutput: []any{int64(4), int64(5), int64(6)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][:0]`, RunOutput: []any{}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][:1]`, RunOutput: []any{int64(4)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][:2]`, RunOutput: []any{int64(4), int64(5)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][:3]`, RunOutput: []any{int64(4), int64(5), int64(6)}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2), int64(3)}, []any{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [["123"], ["456"]]; a[0][0][0:3]`, RunOutput: "123", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][1:3]`, RunOutput: "23", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][2:3]`, RunOutput: "3", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][3:3]`, RunOutput: "", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[0][0][0:]`, RunOutput: "123", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][1:]`, RunOutput: "23", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][2:]`, RunOutput: "3", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][3:]`, RunOutput: "", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[0][0][0:0]`, RunOutput: "", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][0:1]`, RunOutput: "1", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][0:2]`, RunOutput: "12", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][0:3]`, RunOutput: "123", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[0][0][:0]`, RunOutput: "", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][:1]`, RunOutput: "1", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][:2]`, RunOutput: "12", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][:3]`, RunOutput: "123", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[1][0][0:3]`, RunOutput: "456", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][1:3]`, RunOutput: "56", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][2:3]`, RunOutput: "6", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][3:3]`, RunOutput: "", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[1][0][0:]`, RunOutput: "456", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][1:]`, RunOutput: "56", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][2:]`, RunOutput: "6", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][3:]`, RunOutput: "", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[1][0][0:0]`, RunOutput: "", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][0:1]`, RunOutput: "4", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][0:2]`, RunOutput: "45", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][0:3]`, RunOutput: "456", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[1][0][:0]`, RunOutput: "", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][:1]`, RunOutput: "4", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][:2]`, RunOutput: "45", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][:3]`, RunOutput: "456", Output: map[string]any{"a": []any{[]any{"123"}, []any{"456"}}}},
	}
	runTests(t, tests, nil)
}

func TestArrayAppendArrays(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a += nil`, Input: map[string]any{"a": []bool{true}}, RunError: fmt.Errorf("invalid type conversion")},
		{Script: `a += 1`, Input: map[string]any{"a": []bool{true}}, RunError: fmt.Errorf("invalid type conversion")},
		{Script: `a += 1.1`, Input: map[string]any{"a": []bool{true}}, RunError: fmt.Errorf("invalid type conversion")},
		{Script: `a += "a"`, Input: map[string]any{"a": []bool{true}}, RunError: fmt.Errorf("invalid type conversion")},

		{Script: `a += b`, Input: map[string]any{"a": []bool{true}, "b": []string{"b"}}, RunError: fmt.Errorf("invalid type conversion")},
		{Script: `b += a`, Input: map[string]any{"a": []bool{true}, "b": []string{"b"}}, RunError: fmt.Errorf("invalid type conversion")},

		{Script: `b = []; b += a`, Input: map[string]any{"a": []bool{}}, RunOutput: []any{}, Output: map[string]any{"a": []bool{}, "b": []any{}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []bool{true}}, RunOutput: []any{true}, Output: map[string]any{"a": []bool{true}, "b": []any{true}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []bool{true, false}}, RunOutput: []any{true, false}, Output: map[string]any{"a": []bool{true, false}, "b": []any{true, false}}},

		{Script: `b = [true]; b += a`, Input: map[string]any{"a": []bool{}}, RunOutput: []any{true}, Output: map[string]any{"a": []bool{}, "b": []any{true}}},
		{Script: `b = [true]; b += a`, Input: map[string]any{"a": []bool{true}}, RunOutput: []any{true, true}, Output: map[string]any{"a": []bool{true}, "b": []any{true, true}}},
		{Script: `b = [true]; b += a`, Input: map[string]any{"a": []bool{true, false}}, RunOutput: []any{true, true, false}, Output: map[string]any{"a": []bool{true, false}, "b": []any{true, true, false}}},

		{Script: `b = [true, false]; b += a`, Input: map[string]any{"a": []bool{}}, RunOutput: []any{true, false}, Output: map[string]any{"a": []bool{}, "b": []any{true, false}}},
		{Script: `b = [true, false]; b += a`, Input: map[string]any{"a": []bool{true}}, RunOutput: []any{true, false, true}, Output: map[string]any{"a": []bool{true}, "b": []any{true, false, true}}},
		{Script: `b = [true, false]; b += a`, Input: map[string]any{"a": []bool{true, false}}, RunOutput: []any{true, false, true, false}, Output: map[string]any{"a": []bool{true, false}, "b": []any{true, false, true, false}}},

		{Script: `b = []; b += a`, Input: map[string]any{"a": []int32{}}, RunOutput: []any{}, Output: map[string]any{"a": []int32{}, "b": []any{}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []int32{1}}, RunOutput: []any{int32(1)}, Output: map[string]any{"a": []int32{1}, "b": []any{int32(1)}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: []any{int32(1), int32(2)}, Output: map[string]any{"a": []int32{1, 2}, "b": []any{int32(1), int32(2)}}},

		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []int32{}}, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []int32{}, "b": []any{int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []int32{1}}, RunOutput: []any{int64(1), int32(1)}, Output: map[string]any{"a": []int32{1}, "b": []any{int64(1), int32(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: []any{int64(1), int32(1), int32(2)}, Output: map[string]any{"a": []int32{1, 2}, "b": []any{int64(1), int32(1), int32(2)}}},

		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []int32{}}, RunOutput: []any{int64(1), int64(2)}, Output: map[string]any{"a": []int32{}, "b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []int32{1}}, RunOutput: []any{int64(1), int64(2), int32(1)}, Output: map[string]any{"a": []int32{1}, "b": []any{int64(1), int64(2), int32(1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: []any{int64(1), int64(2), int32(1), int32(2)}, Output: map[string]any{"a": []int32{1, 2}, "b": []any{int64(1), int64(2), int32(1), int32(2)}}},

		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []int32{}}, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": []int32{}, "b": []any{float64(1.1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []int32{1}}, RunOutput: []any{float64(1.1), int32(1)}, Output: map[string]any{"a": []int32{1}, "b": []any{float64(1.1), int32(1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: []any{float64(1.1), int32(1), int32(2)}, Output: map[string]any{"a": []int32{1, 2}, "b": []any{float64(1.1), int32(1), int32(2)}}},

		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []int32{}}, RunOutput: []any{float64(1.1), float64(2.2)}, Output: map[string]any{"a": []int32{}, "b": []any{float64(1.1), float64(2.2)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []int32{1}}, RunOutput: []any{float64(1.1), float64(2.2), int32(1)}, Output: map[string]any{"a": []int32{1}, "b": []any{float64(1.1), float64(2.2), int32(1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: []any{float64(1.1), float64(2.2), int32(1), int32(2)}, Output: map[string]any{"a": []int32{1, 2}, "b": []any{float64(1.1), float64(2.2), int32(1), int32(2)}}},

		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []int32{}}, RunOutput: []any{int64(1), float64(2.2)}, Output: map[string]any{"a": []int32{}, "b": []any{int64(1), float64(2.2)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []int32{1}}, RunOutput: []any{int64(1), float64(2.2), int32(1)}, Output: map[string]any{"a": []int32{1}, "b": []any{int64(1), float64(2.2), int32(1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: []any{int64(1), float64(2.2), int32(1), int32(2)}, Output: map[string]any{"a": []int32{1, 2}, "b": []any{int64(1), float64(2.2), int32(1), int32(2)}}},

		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []int32{}}, RunOutput: []any{float64(1.1), int64(2)}, Output: map[string]any{"a": []int32{}, "b": []any{float64(1.1), int64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []int32{1}}, RunOutput: []any{float64(1.1), int64(2), int32(1)}, Output: map[string]any{"a": []int32{1}, "b": []any{float64(1.1), int64(2), int32(1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []int32{1, 2}}, RunOutput: []any{float64(1.1), int64(2), int32(1), int32(2)}, Output: map[string]any{"a": []int32{1, 2}, "b": []any{float64(1.1), int64(2), int32(1), int32(2)}}},

		{Script: `b = []; b += a`, Input: map[string]any{"a": []int64{}}, RunOutput: []any{}, Output: map[string]any{"a": []int64{}, "b": []any{}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []int64{1}}, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []int64{1}, "b": []any{int64(1)}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: []any{int64(1), int64(2)}, Output: map[string]any{"a": []int64{1, 2}, "b": []any{int64(1), int64(2)}}},

		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []int64{}}, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []int64{}, "b": []any{int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []int64{1}}, RunOutput: []any{int64(1), int64(1)}, Output: map[string]any{"a": []int64{1}, "b": []any{int64(1), int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: []any{int64(1), int64(1), int64(2)}, Output: map[string]any{"a": []int64{1, 2}, "b": []any{int64(1), int64(1), int64(2)}}},

		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []int64{}}, RunOutput: []any{int64(1), int64(2)}, Output: map[string]any{"a": []int64{}, "b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []int64{1}}, RunOutput: []any{int64(1), int64(2), int64(1)}, Output: map[string]any{"a": []int64{1}, "b": []any{int64(1), int64(2), int64(1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: []any{int64(1), int64(2), int64(1), int64(2)}, Output: map[string]any{"a": []int64{1, 2}, "b": []any{int64(1), int64(2), int64(1), int64(2)}}},

		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []int64{}}, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": []int64{}, "b": []any{float64(1.1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []int64{1}}, RunOutput: []any{float64(1.1), int64(1)}, Output: map[string]any{"a": []int64{1}, "b": []any{float64(1.1), int64(1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: []any{float64(1.1), int64(1), int64(2)}, Output: map[string]any{"a": []int64{1, 2}, "b": []any{float64(1.1), int64(1), int64(2)}}},

		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []int64{}}, RunOutput: []any{float64(1.1), float64(2.2)}, Output: map[string]any{"a": []int64{}, "b": []any{float64(1.1), float64(2.2)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []int64{1}}, RunOutput: []any{float64(1.1), float64(2.2), int64(1)}, Output: map[string]any{"a": []int64{1}, "b": []any{float64(1.1), float64(2.2), int64(1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: []any{float64(1.1), float64(2.2), int64(1), int64(2)}, Output: map[string]any{"a": []int64{1, 2}, "b": []any{float64(1.1), float64(2.2), int64(1), int64(2)}}},

		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []int64{}}, RunOutput: []any{int64(1), float64(2.2)}, Output: map[string]any{"a": []int64{}, "b": []any{int64(1), float64(2.2)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []int64{1}}, RunOutput: []any{int64(1), float64(2.2), int64(1)}, Output: map[string]any{"a": []int64{1}, "b": []any{int64(1), float64(2.2), int64(1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: []any{int64(1), float64(2.2), int64(1), int64(2)}, Output: map[string]any{"a": []int64{1, 2}, "b": []any{int64(1), float64(2.2), int64(1), int64(2)}}},

		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []int64{}}, RunOutput: []any{float64(1.1), int64(2)}, Output: map[string]any{"a": []int64{}, "b": []any{float64(1.1), int64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []int64{1}}, RunOutput: []any{float64(1.1), int64(2), int64(1)}, Output: map[string]any{"a": []int64{1}, "b": []any{float64(1.1), int64(2), int64(1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []int64{1, 2}}, RunOutput: []any{float64(1.1), int64(2), int64(1), int64(2)}, Output: map[string]any{"a": []int64{1, 2}, "b": []any{float64(1.1), int64(2), int64(1), int64(2)}}},

		{Script: `b = []; b += a`, Input: map[string]any{"a": []float32{}}, RunOutput: []any{}, Output: map[string]any{"a": []float32{}, "b": []any{}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []float32{1}}, RunOutput: []any{float32(1)}, Output: map[string]any{"a": []float32{1}, "b": []any{float32(1)}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []float32{1, 2}}, RunOutput: []any{float32(1), float32(2)}, Output: map[string]any{"a": []float32{1, 2}, "b": []any{float32(1), float32(2)}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []float32{1.1}}, RunOutput: []any{float32(1.1)}, Output: map[string]any{"a": []float32{1.1}, "b": []any{float32(1.1)}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: []any{float32(1.1), float32(2.2)}, Output: map[string]any{"a": []float32{1.1, 2.2}, "b": []any{float32(1.1), float32(2.2)}}},

		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float32{}}, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []float32{}, "b": []any{int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float32{1}}, RunOutput: []any{int64(1), float32(1)}, Output: map[string]any{"a": []float32{1}, "b": []any{int64(1), float32(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float32{1, 2}}, RunOutput: []any{int64(1), float32(1), float32(2)}, Output: map[string]any{"a": []float32{1, 2}, "b": []any{int64(1), float32(1), float32(2)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float32{1.1}}, RunOutput: []any{int64(1), float32(1.1)}, Output: map[string]any{"a": []float32{1.1}, "b": []any{int64(1), float32(1.1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: []any{int64(1), float32(1.1), float32(2.2)}, Output: map[string]any{"a": []float32{1.1, 2.2}, "b": []any{int64(1), float32(1.1), float32(2.2)}}},

		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float32{}}, RunOutput: []any{int64(1), int64(2)}, Output: map[string]any{"a": []float32{}, "b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float32{1}}, RunOutput: []any{int64(1), int64(2), float32(1)}, Output: map[string]any{"a": []float32{1}, "b": []any{int64(1), int64(2), float32(1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float32{1, 2}}, RunOutput: []any{int64(1), int64(2), float32(1), float32(2)}, Output: map[string]any{"a": []float32{1, 2}, "b": []any{int64(1), int64(2), float32(1), float32(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float32{1.1}}, RunOutput: []any{int64(1), int64(2), float32(1.1)}, Output: map[string]any{"a": []float32{1.1}, "b": []any{int64(1), int64(2), float32(1.1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: []any{int64(1), int64(2), float32(1.1), float32(2.2)}, Output: map[string]any{"a": []float32{1.1, 2.2}, "b": []any{int64(1), int64(2), float32(1.1), float32(2.2)}}},

		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float32{}}, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": []float32{}, "b": []any{float64(1.1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float32{1}}, RunOutput: []any{float64(1.1), float32(1)}, Output: map[string]any{"a": []float32{1}, "b": []any{float64(1.1), float32(1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float32{1, 2}}, RunOutput: []any{float64(1.1), float32(1), float32(2)}, Output: map[string]any{"a": []float32{1, 2}, "b": []any{float64(1.1), float32(1), float32(2)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float32{1.1}}, RunOutput: []any{float64(1.1), float32(1.1)}, Output: map[string]any{"a": []float32{1.1}, "b": []any{float64(1.1), float32(1.1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: []any{float64(1.1), float32(1.1), float32(2.2)}, Output: map[string]any{"a": []float32{1.1, 2.2}, "b": []any{float64(1.1), float32(1.1), float32(2.2)}}},

		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float32{}}, RunOutput: []any{float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float32{}, "b": []any{float64(1.1), float64(2.2)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float32{1}}, RunOutput: []any{float64(1.1), float64(2.2), float32(1)}, Output: map[string]any{"a": []float32{1}, "b": []any{float64(1.1), float64(2.2), float32(1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float32{1, 2}}, RunOutput: []any{float64(1.1), float64(2.2), float32(1), float32(2)}, Output: map[string]any{"a": []float32{1, 2}, "b": []any{float64(1.1), float64(2.2), float32(1), float32(2)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float32{1.1}}, RunOutput: []any{float64(1.1), float64(2.2), float32(1.1)}, Output: map[string]any{"a": []float32{1.1}, "b": []any{float64(1.1), float64(2.2), float32(1.1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: []any{float64(1.1), float64(2.2), float32(1.1), float32(2.2)}, Output: map[string]any{"a": []float32{1.1, 2.2}, "b": []any{float64(1.1), float64(2.2), float32(1.1), float32(2.2)}}},

		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float32{}}, RunOutput: []any{int64(1), float64(2.2)}, Output: map[string]any{"a": []float32{}, "b": []any{int64(1), float64(2.2)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float32{1}}, RunOutput: []any{int64(1), float64(2.2), float32(1)}, Output: map[string]any{"a": []float32{1}, "b": []any{int64(1), float64(2.2), float32(1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float32{1, 2}}, RunOutput: []any{int64(1), float64(2.2), float32(1), float32(2)}, Output: map[string]any{"a": []float32{1, 2}, "b": []any{int64(1), float64(2.2), float32(1), float32(2)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float32{1.1}}, RunOutput: []any{int64(1), float64(2.2), float32(1.1)}, Output: map[string]any{"a": []float32{1.1}, "b": []any{int64(1), float64(2.2), float32(1.1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: []any{int64(1), float64(2.2), float32(1.1), float32(2.2)}, Output: map[string]any{"a": []float32{1.1, 2.2}, "b": []any{int64(1), float64(2.2), float32(1.1), float32(2.2)}}},

		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float32{}}, RunOutput: []any{float64(1.1), int64(2)}, Output: map[string]any{"a": []float32{}, "b": []any{float64(1.1), int64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float32{1}}, RunOutput: []any{float64(1.1), int64(2), float32(1)}, Output: map[string]any{"a": []float32{1}, "b": []any{float64(1.1), int64(2), float32(1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float32{1, 2}}, RunOutput: []any{float64(1.1), int64(2), float32(1), float32(2)}, Output: map[string]any{"a": []float32{1, 2}, "b": []any{float64(1.1), int64(2), float32(1), float32(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float32{1.1}}, RunOutput: []any{float64(1.1), int64(2), float32(1.1)}, Output: map[string]any{"a": []float32{1.1}, "b": []any{float64(1.1), int64(2), float32(1.1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float32{1.1, 2.2}}, RunOutput: []any{float64(1.1), int64(2), float32(1.1), float32(2.2)}, Output: map[string]any{"a": []float32{1.1, 2.2}, "b": []any{float64(1.1), int64(2), float32(1.1), float32(2.2)}}},

		{Script: `b = []; b += a`, Input: map[string]any{"a": []float64{}}, RunOutput: []any{}, Output: map[string]any{"a": []float64{}, "b": []any{}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []float64{1}}, RunOutput: []any{float64(1)}, Output: map[string]any{"a": []float64{1}, "b": []any{float64(1)}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []float64{1, 2}}, RunOutput: []any{float64(1), float64(2)}, Output: map[string]any{"a": []float64{1, 2}, "b": []any{float64(1), float64(2)}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []float64{1.1}}, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": []float64{1.1}, "b": []any{float64(1.1)}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: []any{float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float64{1.1, 2.2}, "b": []any{float64(1.1), float64(2.2)}}},

		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float64{}}, RunOutput: []any{int64(1)}, Output: map[string]any{"a": []float64{}, "b": []any{int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float64{1}}, RunOutput: []any{int64(1), float64(1)}, Output: map[string]any{"a": []float64{1}, "b": []any{int64(1), float64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float64{1, 2}}, RunOutput: []any{int64(1), float64(1), float64(2)}, Output: map[string]any{"a": []float64{1, 2}, "b": []any{int64(1), float64(1), float64(2)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float64{1.1}}, RunOutput: []any{int64(1), float64(1.1)}, Output: map[string]any{"a": []float64{1.1}, "b": []any{int64(1), float64(1.1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: []any{int64(1), float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float64{1.1, 2.2}, "b": []any{int64(1), float64(1.1), float64(2.2)}}},

		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float64{}}, RunOutput: []any{int64(1), int64(2)}, Output: map[string]any{"a": []float64{}, "b": []any{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float64{1}}, RunOutput: []any{int64(1), int64(2), float64(1)}, Output: map[string]any{"a": []float64{1}, "b": []any{int64(1), int64(2), float64(1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float64{1, 2}}, RunOutput: []any{int64(1), int64(2), float64(1), float64(2)}, Output: map[string]any{"a": []float64{1, 2}, "b": []any{int64(1), int64(2), float64(1), float64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float64{1.1}}, RunOutput: []any{int64(1), int64(2), float64(1.1)}, Output: map[string]any{"a": []float64{1.1}, "b": []any{int64(1), int64(2), float64(1.1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: []any{int64(1), int64(2), float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float64{1.1, 2.2}, "b": []any{int64(1), int64(2), float64(1.1), float64(2.2)}}},

		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float64{}}, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": []float64{}, "b": []any{float64(1.1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float64{1}}, RunOutput: []any{float64(1.1), float64(1)}, Output: map[string]any{"a": []float64{1}, "b": []any{float64(1.1), float64(1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float64{1, 2}}, RunOutput: []any{float64(1.1), float64(1), float64(2)}, Output: map[string]any{"a": []float64{1, 2}, "b": []any{float64(1.1), float64(1), float64(2)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float64{1.1}}, RunOutput: []any{float64(1.1), float64(1.1)}, Output: map[string]any{"a": []float64{1.1}, "b": []any{float64(1.1), float64(1.1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: []any{float64(1.1), float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float64{1.1, 2.2}, "b": []any{float64(1.1), float64(1.1), float64(2.2)}}},

		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float64{}}, RunOutput: []any{float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float64{}, "b": []any{float64(1.1), float64(2.2)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float64{1}}, RunOutput: []any{float64(1.1), float64(2.2), float64(1)}, Output: map[string]any{"a": []float64{1}, "b": []any{float64(1.1), float64(2.2), float64(1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float64{1, 2}}, RunOutput: []any{float64(1.1), float64(2.2), float64(1), float64(2)}, Output: map[string]any{"a": []float64{1, 2}, "b": []any{float64(1.1), float64(2.2), float64(1), float64(2)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float64{1.1}}, RunOutput: []any{float64(1.1), float64(2.2), float64(1.1)}, Output: map[string]any{"a": []float64{1.1}, "b": []any{float64(1.1), float64(2.2), float64(1.1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: []any{float64(1.1), float64(2.2), float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float64{1.1, 2.2}, "b": []any{float64(1.1), float64(2.2), float64(1.1), float64(2.2)}}},

		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float64{}}, RunOutput: []any{int64(1), float64(2.2)}, Output: map[string]any{"a": []float64{}, "b": []any{int64(1), float64(2.2)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float64{1}}, RunOutput: []any{int64(1), float64(2.2), float64(1)}, Output: map[string]any{"a": []float64{1}, "b": []any{int64(1), float64(2.2), float64(1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float64{1, 2}}, RunOutput: []any{int64(1), float64(2.2), float64(1), float64(2)}, Output: map[string]any{"a": []float64{1, 2}, "b": []any{int64(1), float64(2.2), float64(1), float64(2)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float64{1.1}}, RunOutput: []any{int64(1), float64(2.2), float64(1.1)}, Output: map[string]any{"a": []float64{1.1}, "b": []any{int64(1), float64(2.2), float64(1.1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: []any{int64(1), float64(2.2), float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float64{1.1, 2.2}, "b": []any{int64(1), float64(2.2), float64(1.1), float64(2.2)}}},

		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float64{}}, RunOutput: []any{float64(1.1), int64(2)}, Output: map[string]any{"a": []float64{}, "b": []any{float64(1.1), int64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float64{1}}, RunOutput: []any{float64(1.1), int64(2), float64(1)}, Output: map[string]any{"a": []float64{1}, "b": []any{float64(1.1), int64(2), float64(1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float64{1, 2}}, RunOutput: []any{float64(1.1), int64(2), float64(1), float64(2)}, Output: map[string]any{"a": []float64{1, 2}, "b": []any{float64(1.1), int64(2), float64(1), float64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float64{1.1}}, RunOutput: []any{float64(1.1), int64(2), float64(1.1)}, Output: map[string]any{"a": []float64{1.1}, "b": []any{float64(1.1), int64(2), float64(1.1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]any{"a": []float64{1.1, 2.2}}, RunOutput: []any{float64(1.1), int64(2), float64(1.1), float64(2.2)}, Output: map[string]any{"a": []float64{1.1, 2.2}, "b": []any{float64(1.1), int64(2), float64(1.1), float64(2.2)}}},

		{Script: `b = []; b += a`, Input: map[string]any{"a": []string{}}, RunOutput: []any{}, Output: map[string]any{"a": []string{}, "b": []any{}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []string{"a"}}, RunOutput: []any{"a"}, Output: map[string]any{"a": []string{"a"}, "b": []any{"a"}}},
		{Script: `b = []; b += a`, Input: map[string]any{"a": []string{"a", "b"}}, RunOutput: []any{"a", "b"}, Output: map[string]any{"a": []string{"a", "b"}, "b": []any{"a", "b"}}},

		{Script: `b = ["a"]; b += a`, Input: map[string]any{"a": []string{}}, RunOutput: []any{"a"}, Output: map[string]any{"a": []string{}, "b": []any{"a"}}},
		{Script: `b = ["a"]; b += a`, Input: map[string]any{"a": []string{"a"}}, RunOutput: []any{"a", "a"}, Output: map[string]any{"a": []string{"a"}, "b": []any{"a", "a"}}},
		{Script: `b = ["a"]; b += a`, Input: map[string]any{"a": []string{"a", "b"}}, RunOutput: []any{"a", "a", "b"}, Output: map[string]any{"a": []string{"a", "b"}, "b": []any{"a", "a", "b"}}},

		{Script: `b = ["a", "b"]; b += a`, Input: map[string]any{"a": []string{}}, RunOutput: []any{"a", "b"}, Output: map[string]any{"a": []string{}, "b": []any{"a", "b"}}},
		{Script: `b = ["a", "b"]; b += a`, Input: map[string]any{"a": []string{"a"}}, RunOutput: []any{"a", "b", "a"}, Output: map[string]any{"a": []string{"a"}, "b": []any{"a", "b", "a"}}},
		{Script: `b = ["a", "b"]; b += a`, Input: map[string]any{"a": []string{"a", "b"}}, RunOutput: []any{"a", "b", "a", "b"}, Output: map[string]any{"a": []string{"a", "b"}, "b": []any{"a", "b", "a", "b"}}},

		{Script: `b = [1, "a"]; b += a`, Input: map[string]any{"a": []string{}}, RunOutput: []any{int64(1), "a"}, Output: map[string]any{"a": []string{}, "b": []any{int64(1), "a"}}},
		{Script: `b = [1, "a"]; b += a`, Input: map[string]any{"a": []string{"a"}}, RunOutput: []any{int64(1), "a", "a"}, Output: map[string]any{"a": []string{"a"}, "b": []any{int64(1), "a", "a"}}},
		{Script: `b = [1, "a"]; b += a`, Input: map[string]any{"a": []string{"a", "b"}}, RunOutput: []any{int64(1), "a", "a", "b"}, Output: map[string]any{"a": []string{"a", "b"}, "b": []any{int64(1), "a", "a", "b"}}},

		{Script: `a = []; a += [nil, nil]; a += [nil, nil]`, RunOutput: []any{any(nil), any(nil), any(nil), any(nil)}, Output: map[string]any{"a": []any{any(nil), any(nil), any(nil), any(nil)}}},
		{Script: `a = []; a += [true, false]; a += [false, true]`, RunOutput: []any{any(true), any(false), any(false), any(true)}, Output: map[string]any{"a": []any{any(true), any(false), any(false), any(true)}}},
		{Script: `a = []; a += [1, 2]; a += [3, 4]`, RunOutput: []any{any(int64(1)), any(int64(2)), any(int64(3)), any(int64(4))}, Output: map[string]any{"a": []any{any(int64(1)), any(int64(2)), any(int64(3)), any(int64(4))}}},
		{Script: `a = []; a += [1.1, 2.2]; a += [3.3, 4.4]`, RunOutput: []any{any(float64(1.1)), any(float64(2.2)), any(float64(3.3)), any(float64(4.4))}, Output: map[string]any{"a": []any{any(float64(1.1)), any(float64(2.2)), any(float64(3.3)), any(float64(4.4))}}},
		{Script: `a = []; a += ["a", "b"]; a += ["c", "d"]`, RunOutput: []any{any("a"), any("b"), any("c"), any("d")}, Output: map[string]any{"a": []any{any("a"), any("b"), any("c"), any("d")}}},

		{Script: `a = []; a += [[nil, nil]]; a += [[nil, nil]]`, RunOutput: []any{[]any{nil, nil}, []any{nil, nil}}, Output: map[string]any{"a": []any{[]any{nil, nil}, []any{nil, nil}}}},
		{Script: `a = []; a += [[true, false]]; a += [[false, true]]`, RunOutput: []any{[]any{true, false}, []any{false, true}}, Output: map[string]any{"a": []any{[]any{true, false}, []any{false, true}}}},
		{Script: `a = []; a += [[1, 2]]; a += [[3, 4]]`, RunOutput: []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}, Output: map[string]any{"a": []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}}},
		{Script: `a = []; a += [[1.1, 2.2]]; a += [[3.3, 4.4]]`, RunOutput: []any{[]any{float64(1.1), float64(2.2)}, []any{float64(3.3), float64(4.4)}}, Output: map[string]any{"a": []any{[]any{float64(1.1), float64(2.2)}, []any{float64(3.3), float64(4.4)}}}},
		{Script: `a = []; a += [["a", "b"]]; a += [["c", "d"]]`, RunOutput: []any{[]any{"a", "b"}, []any{"c", "d"}}, Output: map[string]any{"a": []any{[]any{"a", "b"}, []any{"c", "d"}}}},

		{Script: `a = make([]bool); a += 1`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": []bool{}}},
		{Script: `a = make([]bool); a += true; a += 1`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": []bool{true}}},
		{Script: `a = make([]bool); a += [1]`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": []bool{}}},
		{Script: `a = make([]bool); a += [true, 1]`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": []bool{}}},

		{Script: `a = make([]bool); a += true; a += false`, RunOutput: []bool{true, false}, Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a = make([]bool); a += [true]; a += [false]`, RunOutput: []bool{true, false}, Output: map[string]any{"a": []bool{true, false}}},
		{Script: `a = make([]bool); a += [true, false]`, RunOutput: []bool{true, false}, Output: map[string]any{"a": []bool{true, false}}},

		{Script: `a = make([]bool); a += true; a += b`, Input: map[string]any{"b": int64(1)}, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": []bool{true}, "b": int64(1)}},
		{Script: `a = make([]bool); a += [true]; a += [b]`, Input: map[string]any{"b": int64(1)}, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": []bool{true}, "b": int64(1)}},
		{Script: `a = make([]bool); a += [true, b]`, Input: map[string]any{"b": int64(1)}, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": []bool{}, "b": int64(1)}},

		{Script: `a = make([]bool); a += b; a += b`, Input: map[string]any{"b": true}, RunOutput: []bool{true, true}, Output: map[string]any{"a": []bool{true, true}, "b": true}},
		{Script: `a = make([]bool); a += [b]; a += [b]`, Input: map[string]any{"b": true}, RunOutput: []bool{true, true}, Output: map[string]any{"a": []bool{true, true}, "b": true}},
		{Script: `a = make([]bool); a += [b, b]`, Input: map[string]any{"b": true}, RunOutput: []bool{true, true}, Output: map[string]any{"a": []bool{true, true}, "b": true}},

		{Script: `a = make([]bool); a += [true, false]; b = make([]int64); b += [1, 2]; a += b`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": []bool{true, false}, "b": []int64{int64(1), int64(2)}}},

		{Script: `a = []; b = []; b += true; b += false; a += b`, RunOutput: []any{true, false}, Output: map[string]any{"a": []any{true, false}, "b": []any{true, false}}},
		{Script: `a = []; b = make([]bool); b += true; b += false; a += b`, RunOutput: []any{true, false}, Output: map[string]any{"a": []any{true, false}, "b": []bool{true, false}}},
		{Script: `a = []; b = []; b += [true]; b += [false]; a += [b]`, RunOutput: []any{[]any{true, false}}, Output: map[string]any{"a": []any{[]any{true, false}}, "b": []any{true, false}}},
		{Script: `a = []; b = make([]bool); b += [true]; b += [false]; a += [b]`, RunOutput: []any{[]bool{true, false}}, Output: map[string]any{"a": []any{[]bool{true, false}}, "b": []bool{true, false}}},

		{Script: `a = [true, false]; b = [true, false]; a += b`, RunOutput: []any{true, false, true, false}, Output: map[string]any{"a": []any{true, false, true, false}, "b": []any{true, false}}},
		{Script: `a = make([]bool); a += [true, false]; b = make([]bool); b += [true, false]; a += b`, RunOutput: []bool{true, false, true, false}, Output: map[string]any{"a": []bool{true, false, true, false}, "b": []bool{true, false}}},
		{Script: `a = make([]bool); a += [true, false]; b = [true, false]; a += b`, RunOutput: []bool{true, false, true, false}, Output: map[string]any{"a": []bool{true, false, true, false}, "b": []any{true, false}}},
		{Script: `a = [true, false]; b = make([]bool); b += [true, false]; a += b`, RunOutput: []any{true, false, true, false}, Output: map[string]any{"a": []any{true, false, true, false}, "b": []bool{true, false}}},

		{Script: `a = make([][]bool); b = make([][]bool);  a += b`, RunOutput: [][]bool{}, Output: map[string]any{"a": [][]bool{}, "b": [][]bool{}}},
		{Script: `a = make([][]bool); b = make([][]bool); b += [[]]; a += b`, RunOutput: [][]bool{{}}, Output: map[string]any{"a": [][]bool{{}}, "b": [][]bool{{}}}},
		{Script: `a = make([][]bool); a += [[]]; b = make([][]bool); a += b`, RunOutput: [][]bool{{}}, Output: map[string]any{"a": [][]bool{{}}, "b": [][]bool{}}},
		{Script: `a = make([][]bool); a += [[]]; b = make([][]bool); b += [[]]; a += b`, RunOutput: [][]bool{{}, {}}, Output: map[string]any{"a": [][]bool{{}, {}}, "b": [][]bool{{}}}},

		{Script: `a = make([]bool); a += []; b = make([]bool); b += []; a += b`, RunOutput: []bool{}, Output: map[string]any{"a": []bool{}, "b": []bool{}}},
		{Script: `a = make([]bool); a += [true]; b = make([]bool); b += []; a += b`, RunOutput: []bool{true}, Output: map[string]any{"a": []bool{true}, "b": []bool{}}},
		{Script: `a = make([]bool); a += []; b = make([]bool); b += [true]; a += b`, RunOutput: []bool{true}, Output: map[string]any{"a": []bool{true}, "b": []bool{true}}},
		{Script: `a = make([]bool); a += [true]; b = make([]bool); b += [true]; a += b`, RunOutput: []bool{true, true}, Output: map[string]any{"a": []bool{true, true}, "b": []bool{true}}},

		{Script: `a = make([][]bool); a += [true, false];`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": [][]bool{}}},
		{Script: `a = make([][]bool); a += [[true, false]]; b = make([]bool); b += [true, false]; a += b`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": [][]bool{{true, false}}, "b": []bool{true, false}}},
		{Script: `a = make([]bool); a += [true, false]; b = make([][]bool); b += [[true, false]]; a += b`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]any{"a": []bool{true, false}, "b": [][]bool{{true, false}}}},

		{Script: `a = make([][]interface); a += [[1, 2]]`, RunOutput: [][]any{{int64(1), int64(2)}}, Output: map[string]any{"a": [][]any{{int64(1), int64(2)}}}},
		{Script: `a = make([][]interface); b = [1, 2]; a += [b]`, RunOutput: [][]any{{int64(1), int64(2)}}, Output: map[string]any{"a": [][]any{{int64(1), int64(2)}}, "b": []any{int64(1), int64(2)}}},
		{Script: `a = make([][]interface); a += [[1, 2],[3, 4]]`, RunOutput: [][]any{{int64(1), int64(2)}, {int64(3), int64(4)}}, Output: map[string]any{"a": [][]any{{int64(1), int64(2)}, {int64(3), int64(4)}}}},
		{Script: `a = make([][]interface); b = [1, 2]; a += [b]; b = [3, 4]; a += [b]`, RunOutput: [][]any{{int64(1), int64(2)}, {int64(3), int64(4)}}, Output: map[string]any{"a": [][]any{{int64(1), int64(2)}, {int64(3), int64(4)}}, "b": []any{int64(3), int64(4)}}},

		{Script: `a = [["a", "b"], ["c", "d"]]; b = [["w", "x"], ["y", "z"]]; a += b`, RunOutput: []any{[]any{"a", "b"}, []any{"c", "d"}, []any{"w", "x"}, []any{"y", "z"}}, Output: map[string]any{"a": []any{[]any{"a", "b"}, []any{"c", "d"}, []any{"w", "x"}, []any{"y", "z"}}, "b": []any{[]any{"w", "x"}, []any{"y", "z"}}}},
		{Script: `a = make([][]string); a += [["a", "b"], ["c", "d"]]; b = make([][]string); b += [["w", "x"], ["y", "z"]]; a += b`, RunOutput: [][]string{{"a", "b"}, {"c", "d"}, {"w", "x"}, {"y", "z"}}, Output: map[string]any{"a": [][]string{{"a", "b"}, {"c", "d"}, {"w", "x"}, {"y", "z"}}, "b": [][]string{{"w", "x"}, {"y", "z"}}}},
		{Script: `a = make([][]string); a += [["a", "b"], ["c", "d"]]; b = [["w", "x"], ["y", "z"]]; a += b`, RunOutput: [][]string{{"a", "b"}, {"c", "d"}, {"w", "x"}, {"y", "z"}}, Output: map[string]any{"a": [][]string{{"a", "b"}, {"c", "d"}, {"w", "x"}, {"y", "z"}}, "b": []any{[]any{"w", "x"}, []any{"y", "z"}}}},
		{Script: `a = [["a", "b"], ["c", "d"]]; b = make([][]string); b += [["w", "x"], ["y", "z"]]; a += b`, RunOutput: []any{[]any{"a", "b"}, []any{"c", "d"}, []string{"w", "x"}, []string{"y", "z"}}, Output: map[string]any{"a": []any{[]any{"a", "b"}, []any{"c", "d"}, []string{"w", "x"}, []string{"y", "z"}}, "b": [][]string{{"w", "x"}, {"y", "z"}}}},

		{Script: `a = make([][]int64); a += [[1, 2], [3, 4]]; b = make([][]int32); b += [[5, 6], [7, 8]]; a += b`, RunOutput: [][]int64{{int64(1), int64(2)}, {int64(3), int64(4)}, {int64(5), int64(6)}, {int64(7), int64(8)}}, Output: map[string]any{"a": [][]int64{{int64(1), int64(2)}, {int64(3), int64(4)}, {int64(5), int64(6)}, {int64(7), int64(8)}}, "b": [][]int32{{int32(5), int32(6)}, {int32(7), int32(8)}}}},
		{Script: `a = make([][]int32); a += [[1, 2], [3, 4]]; b = make([][]int64); b += [[5, 6], [7, 8]]; a += b`, RunOutput: [][]int32{{int32(1), int32(2)}, {int32(3), int32(4)}, {int32(5), int32(6)}, {int32(7), int32(8)}}, Output: map[string]any{"a": [][]int32{{int32(1), int32(2)}, {int32(3), int32(4)}, {int32(5), int32(6)}, {int32(7), int32(8)}}, "b": [][]int64{{int64(5), int64(6)}, {int64(7), int64(8)}}}},
		{Script: `a = make([][]int64); a += [[1, 2], [3, 4]]; b = make([][]float64); b += [[5, 6], [7, 8]]; a += b`, RunOutput: [][]int64{{int64(1), int64(2)}, {int64(3), int64(4)}, {int64(5), int64(6)}, {int64(7), int64(8)}}, Output: map[string]any{"a": [][]int64{{int64(1), int64(2)}, {int64(3), int64(4)}, {int64(5), int64(6)}, {int64(7), int64(8)}}, "b": [][]float64{{float64(5), float64(6)}, {float64(7), float64(8)}}}},
		{Script: `a = make([][]float64); a += [[1, 2], [3, 4]]; b = make([][]int64); b += [[5, 6], [7, 8]]; a += b`, RunOutput: [][]float64{{float64(1), float64(2)}, {float64(3), float64(4)}, {float64(5), float64(6)}, {float64(7), float64(8)}}, Output: map[string]any{"a": [][]float64{{float64(1), float64(2)}, {float64(3), float64(4)}, {float64(5), float64(6)}, {float64(7), float64(8)}}, "b": [][]int64{{int64(5), int64(6)}, {int64(7), int64(8)}}}},

		{Script: `a = make([][][]interface); a += [[[1, 2]]]`, RunOutput: [][][]any{{{int64(1), int64(2)}}}, Output: map[string]any{"a": [][][]any{{{int64(1), int64(2)}}}}},
		{Script: `a = make([][][]interface); b = [[1, 2]]; a += [b]`, RunOutput: [][][]any{{{int64(1), int64(2)}}}, Output: map[string]any{"a": [][][]any{{{int64(1), int64(2)}}}, "b": []any{[]any{int64(1), int64(2)}}}},
		{Script: `a = make([][][]interface); b = [1, 2]; a += [[b]]`, RunOutput: [][][]any{{{int64(1), int64(2)}}}, Output: map[string]any{"a": [][][]any{{{int64(1), int64(2)}}}, "b": []any{int64(1), int64(2)}}},

		{Script: `a = make([][][]interface); a += [[[1, 2],[3, 4]]]`, RunOutput: [][][]any{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, Output: map[string]any{"a": [][][]any{{{int64(1), int64(2)}, {int64(3), int64(4)}}}}},
		{Script: `a = make([][][]interface); b = [[1, 2],[3, 4]]; a += [b]`, RunOutput: [][][]any{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, Output: map[string]any{"a": [][][]any{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, "b": []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}}},
		{Script: `a = make([][][]interface); b = [1, 2]; c = [b]; b = [3, 4]; c += [b]; a += [c]`, RunOutput: [][][]any{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, Output: map[string]any{"a": [][][]any{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, "b": []any{int64(3), int64(4)}, "c": []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}}},
		{Script: `a = make([][][]interface); b = [1, 2]; c = []; c += [b]; b = [3, 4]; c += [b]; a += [c]`, RunOutput: [][][]any{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, Output: map[string]any{"a": [][][]any{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, "b": []any{int64(3), int64(4)}, "c": []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}}},
	}
	runTests(t, tests, nil)
}

func TestMaps(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `{"b": 1++}`, RunError: fmt.Errorf("invalid operation")},
		{Script: `{1++: 1}`, RunError: fmt.Errorf("invalid operation")},
		{Script: `a = {}; a.b.c`, RunError: fmt.Errorf("type invalid does not support member operation")},
		{Script: `a = {}; a.b.c = 1`, RunError: fmt.Errorf("type invalid does not support member operation")},
		{Script: `a = {}; a[1++]`, RunError: fmt.Errorf("invalid operation")},
		{Script: `a = {}; a[1++] = 1`, RunError: fmt.Errorf("invalid operation")},
		{Script: `b[1]`, RunError: fmt.Errorf("undefined symbol 'b'")},
		{Script: `b[1] = 1`, RunError: fmt.Errorf("undefined symbol 'b'")},
		{Script: `z.y.x = 1`, RunError: fmt.Errorf("undefined symbol 'z'")},

		{Script: `{}`, RunOutput: map[any]any{}},
		{Script: `{"b": nil}`, RunOutput: map[any]any{"b": nil}},
		{Script: `{"b": true}`, RunOutput: map[any]any{"b": true}},
		{Script: `{"b": 1}`, RunOutput: map[any]any{"b": int64(1)}},
		{Script: `{"b": 1.1}`, RunOutput: map[any]any{"b": float64(1.1)}},
		{Script: `{"b": "b"}`, RunOutput: map[any]any{"b": "b"}},

		{Script: `{1: nil}`, RunOutput: map[any]any{int64(1): nil}},
		{Script: `{1: true}`, RunOutput: map[any]any{int64(1): true}},
		{Script: `{1: 2}`, RunOutput: map[any]any{int64(1): int64(2)}},
		{Script: `{1: 2.2}`, RunOutput: map[any]any{int64(1): float64(2.2)}},
		{Script: `{1: "b"}`, RunOutput: map[any]any{int64(1): "b"}},

		{Script: `a = {}`, RunOutput: map[any]any{}, Output: map[string]any{"a": map[any]any{}}},
		{Script: `a = {"b": nil}`, RunOutput: map[any]any{"b": nil}, Output: map[string]any{"a": map[any]any{"b": nil}}},
		{Script: `a = {"b": true}`, RunOutput: map[any]any{"b": true}, Output: map[string]any{"a": map[any]any{"b": true}}},
		{Script: `a = {"b": 1}`, RunOutput: map[any]any{"b": int64(1)}, Output: map[string]any{"a": map[any]any{"b": int64(1)}}},
		{Script: `a = {"b": 1.1}`, RunOutput: map[any]any{"b": float64(1.1)}, Output: map[string]any{"a": map[any]any{"b": float64(1.1)}}},
		{Script: `a = {"b": "b"}`, RunOutput: map[any]any{"b": "b"}, Output: map[string]any{"a": map[any]any{"b": "b"}}},

		{Script: `a = {"b": {}}`, RunOutput: map[any]any{"b": map[any]any{}}, Output: map[string]any{"a": map[any]any{"b": map[any]any{}}}},
		{Script: `a = {"b": {"c": nil}}`, RunOutput: map[any]any{"b": map[any]any{"c": nil}}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": nil}}}},
		{Script: `a = {"b": {"c": true}}`, RunOutput: map[any]any{"b": map[any]any{"c": true}}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": true}}}},
		{Script: `a = {"b": {"c": 1}}`, RunOutput: map[any]any{"b": map[any]any{"c": int64(1)}}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": int64(1)}}}},
		{Script: `a = {"b": {"c": 1.1}}`, RunOutput: map[any]any{"b": map[any]any{"c": float64(1.1)}}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": float64(1.1)}}}},
		{Script: `a = {"b": {"c": "c"}}`, RunOutput: map[any]any{"b": map[any]any{"c": "c"}}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": "c"}}}},

		{Script: `a = {"b": {}}; a.b`, RunOutput: map[any]any{}, Output: map[string]any{"a": map[any]any{"b": map[any]any{}}}},
		{Script: `a = {"b": {"c": nil}}; a.b`, RunOutput: map[any]any{"c": nil}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": nil}}}},
		{Script: `a = {"b": {"c": true}}; a.b`, RunOutput: map[any]any{"c": true}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": true}}}},
		{Script: `a = {"b": {"c": 1}}; a.b`, RunOutput: map[any]any{"c": int64(1)}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": int64(1)}}}},
		{Script: `a = {"b": {"c": 1.1}}; a.b`, RunOutput: map[any]any{"c": float64(1.1)}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": float64(1.1)}}}},
		{Script: `a = {"b": {"c": "c"}}; a.b`, RunOutput: map[any]any{"c": "c"}, Output: map[string]any{"a": map[any]any{"b": map[any]any{"c": "c"}}}},

		{Script: `a = {"b": []}`, RunOutput: map[any]any{"b": []any{}}, Output: map[string]any{"a": map[any]any{"b": []any{}}}},
		{Script: `a = {"b": [nil]}`, RunOutput: map[any]any{"b": []any{nil}}, Output: map[string]any{"a": map[any]any{"b": []any{nil}}}},
		{Script: `a = {"b": [true]}`, RunOutput: map[any]any{"b": []any{true}}, Output: map[string]any{"a": map[any]any{"b": []any{true}}}},
		{Script: `a = {"b": [1]}`, RunOutput: map[any]any{"b": []any{int64(1)}}, Output: map[string]any{"a": map[any]any{"b": []any{int64(1)}}}},
		{Script: `a = {"b": [1.1]}`, RunOutput: map[any]any{"b": []any{float64(1.1)}}, Output: map[string]any{"a": map[any]any{"b": []any{float64(1.1)}}}},
		{Script: `a = {"b": ["c"]}`, RunOutput: map[any]any{"b": []any{"c"}}, Output: map[string]any{"a": map[any]any{"b": []any{"c"}}}},

		{Script: `a = {}; a.b`, RunOutput: nil, Output: map[string]any{"a": map[any]any{}}},
		{Script: `a = {"b": nil}; a.b`, RunOutput: nil, Output: map[string]any{"a": map[any]any{"b": nil}}},
		{Script: `a = {"b": true}; a.b`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": true}}},
		{Script: `a = {"b": 1}; a.b`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"b": int64(1)}}},
		{Script: `a = {"b": 1.1}; a.b`, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]any{"b": float64(1.1)}}},
		{Script: `a = {"b": "b"}; a.b`, RunOutput: "b", Output: map[string]any{"a": map[any]any{"b": "b"}}},

		{Script: `a = {}; a["b"]`, RunOutput: nil, Output: map[string]any{"a": map[any]any{}}},
		{Script: `a = {"b": nil}; a["b"]`, RunOutput: nil, Output: map[string]any{"a": map[any]any{"b": nil}}},
		{Script: `a = {"b": true}; a["b"]`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": true}}},
		{Script: `a = {"b": 1}; a["b"]`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"b": int64(1)}}},
		{Script: `a = {"b": 1.1}; a["b"]`, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]any{"b": float64(1.1)}}},
		{Script: `a = {"b": "b"}; a["b"]`, RunOutput: "b", Output: map[string]any{"a": map[any]any{"b": "b"}}},

		{Script: `a`, Input: map[string]any{"a": map[string]any{}}, RunOutput: map[string]any{}, Output: map[string]any{"a": map[string]any{}}},
		{Script: `a`, Input: map[string]any{"a": map[string]any{"b": nil}}, RunOutput: map[string]any{"b": nil}, Output: map[string]any{"a": map[string]any{"b": nil}}},
		{Script: `a`, Input: map[string]any{"a": map[string]any{"b": true}}, RunOutput: map[string]any{"b": true}, Output: map[string]any{"a": map[string]any{"b": true}}},
		{Script: `a`, Input: map[string]any{"a": map[string]any{"b": int32(1)}}, RunOutput: map[string]any{"b": int32(1)}, Output: map[string]any{"a": map[string]any{"b": int32(1)}}},
		{Script: `a`, Input: map[string]any{"a": map[string]any{"b": int64(1)}}, RunOutput: map[string]any{"b": int64(1)}, Output: map[string]any{"a": map[string]any{"b": int64(1)}}},
		{Script: `a`, Input: map[string]any{"a": map[string]any{"b": float32(1.1)}}, RunOutput: map[string]any{"b": float32(1.1)}, Output: map[string]any{"a": map[string]any{"b": float32(1.1)}}},
		{Script: `a`, Input: map[string]any{"a": map[string]any{"b": float64(1.1)}}, RunOutput: map[string]any{"b": float64(1.1)}, Output: map[string]any{"a": map[string]any{"b": float64(1.1)}}},
		{Script: `a`, Input: map[string]any{"a": map[string]any{"b": "b"}}, RunOutput: map[string]any{"b": "b"}, Output: map[string]any{"a": map[string]any{"b": "b"}}},

		{Script: `a.b`, Input: map[string]any{"a": map[string]any{}}, RunOutput: nil, Output: map[string]any{"a": map[string]any{}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]any{"b": nil}}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": nil}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]any{"b": true}}, RunOutput: true, Output: map[string]any{"a": map[string]any{"b": true}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]any{"b": int32(1)}}, RunOutput: int32(1), Output: map[string]any{"a": map[string]any{"b": int32(1)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]any{"b": int64(1)}}, RunOutput: int64(1), Output: map[string]any{"a": map[string]any{"b": int64(1)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]any{"b": float32(1.1)}}, RunOutput: float32(1.1), Output: map[string]any{"a": map[string]any{"b": float32(1.1)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]any{"b": float64(1.1)}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[string]any{"b": float64(1.1)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]any{"b": "b"}}, RunOutput: "b", Output: map[string]any{"a": map[string]any{"b": "b"}}},

		{Script: `a.b`, Input: map[string]any{"a": map[string]bool{"a": false, "b": true}}, RunOutput: true, Output: map[string]any{"a": map[string]bool{"a": false, "b": true}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]int32{"a": int32(1), "b": int32(2)}}, RunOutput: int32(2), Output: map[string]any{"a": map[string]int32{"a": int32(1), "b": int32(2)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]int64{"a": int64(1), "b": int64(2)}}, RunOutput: int64(2), Output: map[string]any{"a": map[string]int64{"a": int64(1), "b": int64(2)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]float32{"a": float32(1.1), "b": float32(2.2)}}, RunOutput: float32(2.2), Output: map[string]any{"a": map[string]float32{"a": float32(1.1), "b": float32(2.2)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]float64{"a": float64(1.1), "b": float64(2.2)}}, RunOutput: float64(2.2), Output: map[string]any{"a": map[string]float64{"a": float64(1.1), "b": float64(2.2)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[string]string{"a": "a", "b": "b"}}, RunOutput: "b", Output: map[string]any{"a": map[string]string{"a": "a", "b": "b"}}},

		{Script: `a["b"]`, Input: map[string]any{"a": map[string]any{}}, RunOutput: nil, Output: map[string]any{"a": map[string]any{}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[string]any{"b": reflect.Value{}}}, RunOutput: reflect.Value{}, Output: map[string]any{"a": map[string]any{"b": reflect.Value{}}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[string]any{"b": nil}}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": nil}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[string]any{"b": true}}, RunOutput: true, Output: map[string]any{"a": map[string]any{"b": true}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[string]any{"b": int32(1)}}, RunOutput: int32(1), Output: map[string]any{"a": map[string]any{"b": int32(1)}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[string]any{"b": int64(1)}}, RunOutput: int64(1), Output: map[string]any{"a": map[string]any{"b": int64(1)}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[string]any{"b": float32(1.1)}}, RunOutput: float32(1.1), Output: map[string]any{"a": map[string]any{"b": float32(1.1)}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[string]any{"b": float64(1.1)}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[string]any{"b": float64(1.1)}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[string]any{"b": "b"}}, RunOutput: "b", Output: map[string]any{"a": map[string]any{"b": "b"}}},

		{Script: `a[0:1]`, Input: map[string]any{"a": map[string]any{}}, RunError: fmt.Errorf("type map does not support slice operation"), Output: map[string]any{"a": map[string]any{}}},
		{Script: `a[0:1]`, Input: map[string]any{"a": map[string]any{"b": reflect.Value{}}}, RunError: fmt.Errorf("type map does not support slice operation"), Output: map[string]any{"a": map[string]any{"b": reflect.Value{}}}},
		{Script: `a[0:1]`, Input: map[string]any{"a": map[string]any{"b": nil}}, RunError: fmt.Errorf("type map does not support slice operation"), Output: map[string]any{"a": map[string]any{"b": nil}}},
		{Script: `a[0:1]`, Input: map[string]any{"a": map[string]any{"b": true}}, RunError: fmt.Errorf("type map does not support slice operation"), Output: map[string]any{"a": map[string]any{"b": true}}},
		{Script: `a[0:1]`, Input: map[string]any{"a": map[string]any{"b": int32(1)}}, RunError: fmt.Errorf("type map does not support slice operation"), Output: map[string]any{"a": map[string]any{"b": int32(1)}}},
		{Script: `a[0:1]`, Input: map[string]any{"a": map[string]any{"b": int64(1)}}, RunError: fmt.Errorf("type map does not support slice operation"), Output: map[string]any{"a": map[string]any{"b": int64(1)}}},
		{Script: `a[0:1]`, Input: map[string]any{"a": map[string]any{"b": float32(1.1)}}, RunError: fmt.Errorf("type map does not support slice operation"), Output: map[string]any{"a": map[string]any{"b": float32(1.1)}}},
		{Script: `a[0:1]`, Input: map[string]any{"a": map[string]any{"b": float64(1.1)}}, RunError: fmt.Errorf("type map does not support slice operation"), Output: map[string]any{"a": map[string]any{"b": float64(1.1)}}},
		{Script: `a[0:1]`, Input: map[string]any{"a": map[string]any{"b": "b"}}, RunError: fmt.Errorf("type map does not support slice operation"), Output: map[string]any{"a": map[string]any{"b": "b"}}},

		{Script: `a[c]`, Input: map[string]any{"a": map[string]any{"b": "b"}, "c": nil}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": nil}},
		{Script: `a[c]`, Input: map[string]any{"a": map[string]any{"b": "b"}, "c": true}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": true}},
		{Script: `a[c]`, Input: map[string]any{"a": map[string]any{"b": "b"}, "c": int32(1)}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": int32(1)}},
		{Script: `a[c]`, Input: map[string]any{"a": map[string]any{"b": "b"}, "c": int64(1)}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": int64(1)}},
		{Script: `a[c]`, Input: map[string]any{"a": map[string]any{"b": "b"}, "c": float32(1.1)}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": float32(1.1)}},
		{Script: `a[c]`, Input: map[string]any{"a": map[string]any{"b": "b"}, "c": float64(1.1)}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": float64(1.1)}},
		{Script: `a[c]`, Input: map[string]any{"a": map[string]any{"b": "b"}, "c": "b"}, RunOutput: "b", Output: map[string]any{"a": map[string]any{"b": "b"}, "c": "b"}},
		{Script: `a[c]`, Input: map[string]any{"a": map[string]any{"b": "b"}, "c": "c"}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": "c"}},

		{Script: `a.b = nil`, Input: map[string]any{"a": map[string]any{}}, RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": nil}}},
		{Script: `a.b = true`, Input: map[string]any{"a": map[string]any{}}, RunOutput: true, Output: map[string]any{"a": map[string]any{"b": true}}},
		{Script: `a.b = 1`, Input: map[string]any{"a": map[string]any{}}, RunOutput: int64(1), Output: map[string]any{"a": map[string]any{"b": int64(1)}}},
		{Script: `a.b = 1.1`, Input: map[string]any{"a": map[string]any{}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[string]any{"b": float64(1.1)}}},
		{Script: `a.b = "b"`, Input: map[string]any{"a": map[string]any{}}, RunOutput: "b", Output: map[string]any{"a": map[string]any{"b": "b"}}},

		{Script: `a.b = true`, Input: map[string]any{"a": map[string]bool{"a": true, "b": false}}, RunOutput: true, Output: map[string]any{"a": map[string]bool{"a": true, "b": true}}},
		{Script: `a.b = 3`, Input: map[string]any{"a": map[string]int32{"a": int32(1), "b": int32(2)}}, RunOutput: int64(3), Output: map[string]any{"a": map[string]int32{"a": int32(1), "b": int32(3)}}},
		{Script: `a.b = 3`, Input: map[string]any{"a": map[string]int64{"a": int64(1), "b": int64(2)}}, RunOutput: int64(3), Output: map[string]any{"a": map[string]int64{"a": int64(1), "b": int64(3)}}},
		{Script: `a.b = 3.3`, Input: map[string]any{"a": map[string]float32{"a": float32(1.1), "b": float32(2.2)}}, RunOutput: float64(3.3), Output: map[string]any{"a": map[string]float32{"a": float32(1.1), "b": float32(3.3)}}},
		{Script: `a.b = 3.3`, Input: map[string]any{"a": map[string]float64{"a": float64(1.1), "b": float64(2.2)}}, RunOutput: float64(3.3), Output: map[string]any{"a": map[string]float64{"a": float64(1.1), "b": float64(3.3)}}},
		{Script: `a.b = "c"`, Input: map[string]any{"a": map[string]string{"a": "a", "b": "b"}}, RunOutput: "c", Output: map[string]any{"a": map[string]string{"a": "a", "b": "c"}}},

		{Script: `a["b"] = true`, Input: map[string]any{"a": map[string]bool{"a": true, "b": false}}, RunOutput: true, Output: map[string]any{"a": map[string]bool{"a": true, "b": true}}},
		{Script: `a["b"] = 3`, Input: map[string]any{"a": map[string]int32{"a": int32(1), "b": int32(2)}}, RunOutput: int64(3), Output: map[string]any{"a": map[string]int32{"a": int32(1), "b": int32(3)}}},
		{Script: `a["b"] = 3`, Input: map[string]any{"a": map[string]int64{"a": int64(1), "b": int64(2)}}, RunOutput: int64(3), Output: map[string]any{"a": map[string]int64{"a": int64(1), "b": int64(3)}}},
		{Script: `a["b"] = 3.3`, Input: map[string]any{"a": map[string]float32{"a": float32(1.1), "b": float32(2.2)}}, RunOutput: float64(3.3), Output: map[string]any{"a": map[string]float32{"a": float32(1.1), "b": float32(3.3)}}},
		{Script: `a["b"] = 3.3`, Input: map[string]any{"a": map[string]float64{"a": float64(1.1), "b": float64(2.2)}}, RunOutput: float64(3.3), Output: map[string]any{"a": map[string]float64{"a": float64(1.1), "b": float64(3.3)}}},
		{Script: `a["b"] = "c"`, Input: map[string]any{"a": map[string]string{"a": "a", "b": "b"}}, RunOutput: "c", Output: map[string]any{"a": map[string]string{"a": "a", "b": "c"}}},

		{Script: `a[c] = "x"`, Input: map[string]any{"a": map[string]any{"b": "b"}, "c": true}, RunError: fmt.Errorf("index type bool cannot be used for map index type string"), RunOutput: nil, Output: map[string]any{"a": map[string]any{"b": "b"}, "c": true}},
		{Script: `a[c] = "x"`, Input: map[string]any{"a": map[bool]any{true: "b"}, "c": true}, RunOutput: "x", Output: map[string]any{"a": map[bool]any{true: "x"}, "c": true}},

		// note if passed an uninitialized map there does not seem to be a way to update that map
		{Script: `a`, Input: map[string]any{"a": testMapEmpty}, RunOutput: testMapEmpty, Output: map[string]any{"a": testMapEmpty}},
		{Script: `a.b`, Input: map[string]any{"a": testMapEmpty}, RunOutput: nil, Output: map[string]any{"a": testMapEmpty}},
		{Script: `a.b = 1`, Input: map[string]any{"a": testMapEmpty}, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"b": int64(1)}}},
		{Script: `a.b`, Input: map[string]any{"a": testMapEmpty}, RunOutput: nil, Output: map[string]any{"a": testMapEmpty}},
		{Script: `a["b"]`, Input: map[string]any{"a": testMapEmpty}, RunOutput: nil, Output: map[string]any{"a": testMapEmpty}},
		{Script: `a["b"] = 1`, Input: map[string]any{"a": testMapEmpty}, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"b": int64(1)}}},
		{Script: `a["b"]`, Input: map[string]any{"a": testMapEmpty}, RunOutput: nil, Output: map[string]any{"a": testMapEmpty}},

		{Script: `a`, Input: map[string]any{"a": testMap}, RunOutput: testMap, Output: map[string]any{"a": testMap}},
		{Script: `a.a`, Input: map[string]any{"a": testMap}, RunOutput: nil, Output: map[string]any{"a": testMap}},
		{Script: `a.a = true`, Input: map[string]any{"a": testMap}, RunOutput: true, Output: map[string]any{"a": testMap}},
		{Script: `a.a`, Input: map[string]any{"a": testMap}, RunOutput: true, Output: map[string]any{"a": testMap}},
		{Script: `a.a = nil`, Input: map[string]any{"a": testMap}, RunOutput: nil, Output: map[string]any{"a": testMap}},

		{Script: `a`, Input: map[string]any{"a": testMap}, RunOutput: testMap, Output: map[string]any{"a": testMap}},
		{Script: `a.b`, Input: map[string]any{"a": testMap}, RunOutput: true, Output: map[string]any{"a": testMap}},
		{Script: `a.b = false`, Input: map[string]any{"a": testMap}, RunOutput: false, Output: map[string]any{"a": testMap}},
		{Script: `a.b`, Input: map[string]any{"a": testMap}, RunOutput: false, Output: map[string]any{"a": testMap}},
		{Script: `a.b = true`, Input: map[string]any{"a": testMap}, RunOutput: true, Output: map[string]any{"a": testMap}},

		{Script: `a`, Input: map[string]any{"a": testMap}, RunOutput: testMap, Output: map[string]any{"a": testMap}},
		{Script: `a.c`, Input: map[string]any{"a": testMap}, RunOutput: int64(1), Output: map[string]any{"a": testMap}},
		{Script: `a.c = 2`, Input: map[string]any{"a": testMap}, RunOutput: int64(2), Output: map[string]any{"a": testMap}},
		{Script: `a.c`, Input: map[string]any{"a": testMap}, RunOutput: int64(2), Output: map[string]any{"a": testMap}},
		{Script: `a.c = 1`, Input: map[string]any{"a": testMap}, RunOutput: int64(1), Output: map[string]any{"a": testMap}},

		{Script: `a`, Input: map[string]any{"a": testMap}, RunOutput: testMap, Output: map[string]any{"a": testMap}},
		{Script: `a.d`, Input: map[string]any{"a": testMap}, RunOutput: float64(1.1), Output: map[string]any{"a": testMap}},
		{Script: `a.d = 2.2`, Input: map[string]any{"a": testMap}, RunOutput: float64(2.2), Output: map[string]any{"a": testMap}},
		{Script: `a.d`, Input: map[string]any{"a": testMap}, RunOutput: float64(2.2), Output: map[string]any{"a": testMap}},
		{Script: `a.d = 1.1`, Input: map[string]any{"a": testMap}, RunOutput: float64(1.1), Output: map[string]any{"a": testMap}},

		{Script: `a`, Input: map[string]any{"a": testMap}, RunOutput: testMap, Output: map[string]any{"a": testMap}},
		{Script: `a.e`, Input: map[string]any{"a": testMap}, RunOutput: "e", Output: map[string]any{"a": testMap}},
		{Script: `a.e = "x"`, Input: map[string]any{"a": testMap}, RunOutput: "x", Output: map[string]any{"a": testMap}},
		{Script: `a.e`, Input: map[string]any{"a": testMap}, RunOutput: "x", Output: map[string]any{"a": testMap}},
		{Script: `a.e = "e"`, Input: map[string]any{"a": testMap}, RunOutput: "e", Output: map[string]any{"a": testMap}},

		{Script: `a = {"b": 1, "c": nil}`, RunOutput: map[any]any{"b": int64(1), "c": nil}, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.b`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.c`, RunOutput: nil, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.d`, RunOutput: nil, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},

		{Script: `a = {"b": 1, "c": nil}; a == nil`, RunOutput: false, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a != nil`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.b == nil`, RunOutput: false, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.b != nil`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.c == nil`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.c != nil`, RunOutput: false, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.d == nil`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.d != nil`, RunOutput: false, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},

		{Script: `a = {"b": 1, "c": nil}; a == 1`, RunOutput: false, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a != 1`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.b == 1`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.b != 1`, RunOutput: false, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.c == 1`, RunOutput: false, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.c != 1`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.d == 1`, RunOutput: false, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},
		{Script: `a = {"b": 1, "c": nil}; a.d != 1`, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": int64(1), "c": nil}}},

		{Script: `a["b"]`, Input: map[string]any{"a": map[any]bool{"b": true}}, RunOutput: true, Output: map[string]any{"a": map[any]bool{"b": true}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[any]int32{"b": int32(1)}}, RunOutput: int32(1), Output: map[string]any{"a": map[any]int32{"b": int32(1)}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[any]int64{"b": int64(1)}}, RunOutput: int64(1), Output: map[string]any{"a": map[any]int64{"b": int64(1)}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[any]float32{"b": float32(1.1)}}, RunOutput: float32(1.1), Output: map[string]any{"a": map[any]float32{"b": float32(1.1)}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[any]float64{"b": float64(1.1)}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]float64{"b": float64(1.1)}}},
		{Script: `a["b"]`, Input: map[string]any{"a": map[any]string{"b": "b"}}, RunOutput: "b", Output: map[string]any{"a": map[any]string{"b": "b"}}},

		{Script: `a.b`, Input: map[string]any{"a": map[any]bool{"b": true}}, RunOutput: true, Output: map[string]any{"a": map[any]bool{"b": true}}},
		{Script: `a.b`, Input: map[string]any{"a": map[any]int32{"b": int32(1)}}, RunOutput: int32(1), Output: map[string]any{"a": map[any]int32{"b": int32(1)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[any]int64{"b": int64(1)}}, RunOutput: int64(1), Output: map[string]any{"a": map[any]int64{"b": int64(1)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[any]float32{"b": float32(1.1)}}, RunOutput: float32(1.1), Output: map[string]any{"a": map[any]float32{"b": float32(1.1)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[any]float64{"b": float64(1.1)}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]float64{"b": float64(1.1)}}},
		{Script: `a.b`, Input: map[string]any{"a": map[any]string{"b": "b"}}, RunOutput: "b", Output: map[string]any{"a": map[any]string{"b": "b"}}},

		{Script: `a = {}; a[true] = true`, RunOutput: true, Output: map[string]any{"a": map[any]any{true: true}}},
		{Script: `a = {}; a[1] = 1`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{int64(1): int64(1)}}},
		{Script: `a = {}; a[1.1] = 1.1`, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]any{float64(1.1): float64(1.1)}}},

		{Script: `a = {}; a[true] = true; a[true]`, RunOutput: true, Output: map[string]any{"a": map[any]any{true: true}}},
		{Script: `a = {}; a[1] = 1; a[1]`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{int64(1): int64(1)}}},
		{Script: `a = {}; a[1.1] = 1.1; a[1.1]`, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]any{float64(1.1): float64(1.1)}}},

		{Script: `a = {}; a[b] = b`, Input: map[string]any{"b": nil}, RunOutput: nil, Output: map[string]any{"a": map[any]any{nil: nil}}},
		{Script: `a = {}; a[b] = b`, Input: map[string]any{"b": int32(1)}, RunOutput: int32(1), Output: map[string]any{"a": map[any]any{int32(1): int32(1)}}},
		{Script: `a = {}; a[b] = b`, Input: map[string]any{"b": int64(1)}, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{int64(1): int64(1)}}},
		{Script: `a = {}; a[b] = b`, Input: map[string]any{"b": float32(1.1)}, RunOutput: float32(1.1), Output: map[string]any{"a": map[any]any{float32(1.1): float32(1.1)}}},
		{Script: `a = {}; a[b] = b`, Input: map[string]any{"b": float64(1.1)}, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]any{float64(1.1): float64(1.1)}}},
		{Script: `a = {}; a[b] = b`, Input: map[string]any{"b": "b"}, RunOutput: "b", Output: map[string]any{"a": map[any]any{"b": "b"}}},

		{Script: `a = {}; a[b] = b; a[b]`, Input: map[string]any{"b": nil}, RunOutput: nil, Output: map[string]any{"a": map[any]any{nil: nil}}},
		{Script: `a = {}; a[b] = b; a[b]`, Input: map[string]any{"b": int32(1)}, RunOutput: int32(1), Output: map[string]any{"a": map[any]any{int32(1): int32(1)}}},
		{Script: `a = {}; a[b] = b; a[b]`, Input: map[string]any{"b": int64(1)}, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{int64(1): int64(1)}}},
		{Script: `a = {}; a[b] = b; a[b]`, Input: map[string]any{"b": float32(1.1)}, RunOutput: float32(1.1), Output: map[string]any{"a": map[any]any{float32(1.1): float32(1.1)}}},
		{Script: `a = {}; a[b] = b; a[b]`, Input: map[string]any{"b": float64(1.1)}, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]any{float64(1.1): float64(1.1)}}},
		{Script: `a = {}; a[b] = b; a[b]`, Input: map[string]any{"b": "b"}, RunOutput: "b", Output: map[string]any{"a": map[any]any{"b": "b"}}},

		// test equal nil when not found
		{Script: `a = {"b":"b"}; if a.c == nil { return 1 }`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"b": "b"}}},
		{Script: `a = {"b":"b"}; if a["c"] == nil { return 1 }`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"b": "b"}}},

		// test map create with spacing and comma
		{Script: `a = {"b": "b"
}`, RunOutput: map[any]any{"b": "b"}},
		{Script: `a = {"b": "b",}`, RunOutput: map[any]any{"b": "b"}},
		{Script: `a = {"b": "b", "c": "c"}; a.c`, RunOutput: "c"},
		{Script: `a = {"b": "b", 
"c": "c"}; a.c`, RunOutput: "c"},
		{Script: `a = {"b": "b", "c": "c",}; a.c`, RunOutput: "c"},
		{Script: `a = {"b": "b", 
"c": "c",}; a.c`, RunOutput: "c"},
		{Script: `a = {"b": "b", 
"c": "c",
}; a.c`, RunOutput: "c"},
	}
	runTests(t, tests, nil)
}

func TestExistenceOfKeyInMaps(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = {"b":"b"}; v, ok = a[1++]`, RunError: fmt.Errorf("invalid operation")},
		{Script: `a = {"b":"b"}; b.c, ok = a["b"]`, RunError: fmt.Errorf("undefined symbol 'b'")},
		{Script: `a = {"b":"b"}; v, b.c = a["b"]`, RunError: fmt.Errorf("undefined symbol 'b'")},

		{Script: `a = {"b":"b"}; v, ok = a["a"]`, RunOutput: nil, Output: map[string]any{"a": map[any]any{"b": "b"}, "v": nil, "ok": false}},
		{Script: `a = {"b":"b"}; v, ok = a["b"]`, RunOutput: "b", Output: map[string]any{"a": map[any]any{"b": "b"}, "v": "b", "ok": true}},
		{Script: `a = {"b":"b", "c":"c"}; v, ok = a["a"]`, RunOutput: nil, Output: map[string]any{"a": map[any]any{"b": "b", "c": "c"}, "v": nil, "ok": false}},
		{Script: `a = {"b":"b", "c":"c"}; v, ok = a["b"]`, RunOutput: "b", Output: map[string]any{"a": map[any]any{"b": "b", "c": "c"}, "v": "b", "ok": true}},
	}
	runTests(t, tests, nil)
}

func TestDeleteMaps(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `delete(1++, "b")`, RunError: fmt.Errorf("invalid operation")},
		{Script: `delete({}, 1++)`, RunError: fmt.Errorf("invalid operation")},
		{Script: `delete(nil, "b")`, RunError: fmt.Errorf("first argument to delete cannot be type interface")},
		{Script: `delete(1, "b")`, RunError: fmt.Errorf("first argument to delete cannot be type int64")},

		{Script: `delete(a, "")`, Input: map[string]any{"a": testMapEmpty}, Output: map[string]any{"a": testMapEmpty}},
		{Script: `delete(a, "")`, Input: map[string]any{"a": map[string]any{"b": "b"}}, Output: map[string]any{"a": map[string]any{"b": "b"}}},
		{Script: `delete(a, "a")`, Input: map[string]any{"a": map[string]any{"b": "b"}}, Output: map[string]any{"a": map[string]any{"b": "b"}}},
		{Script: `delete(a, "b")`, Input: map[string]any{"a": map[string]any{"b": "b"}}, Output: map[string]any{"a": map[string]any{}}},
		{Script: `delete(a, "a")`, Input: map[string]any{"a": map[string]any{"b": "b", "c": "c"}}, Output: map[string]any{"a": map[string]any{"b": "b", "c": "c"}}},
		{Script: `delete(a, "b")`, Input: map[string]any{"a": map[string]any{"b": "b", "c": "c"}}, Output: map[string]any{"a": map[string]any{"c": "c"}}},

		{Script: `delete(a, 0)`, Input: map[string]any{"a": map[int64]any{1: 1}}, Output: map[string]any{"a": map[int64]any{1: 1}}},
		{Script: `delete(a, 1)`, Input: map[string]any{"a": map[int64]any{1: 1}}, Output: map[string]any{"a": map[int64]any{}}},
		{Script: `delete(a, 0)`, Input: map[string]any{"a": map[int64]any{1: 1, 2: 2}}, Output: map[string]any{"a": map[int64]any{1: 1, 2: 2}}},
		{Script: `delete(a, 1)`, Input: map[string]any{"a": map[int64]any{1: 1, 2: 2}}, Output: map[string]any{"a": map[int64]any{2: 2}}},

		{Script: `delete({}, "")`},
		{Script: `delete({}, 1)`},
		{Script: `delete({}, "a")`},
		{Script: `delete({"b":"b"}, "")`},
		{Script: `delete({"b":"b"}, 1)`},
		{Script: `delete({"b":"b"}, "a")`},
		{Script: `delete({"b":"b"}, "b")`},

		{Script: `a = {"b": "b"}; delete(a, "a")`, Output: map[string]any{"a": map[any]any{"b": "b"}}},
		{Script: `a = {"b": "b"}; delete(a, "b")`, Output: map[string]any{"a": map[any]any{}}},
		{Script: `a = {"b": "b", "c":"c"}; delete(a, "a")`, Output: map[string]any{"a": map[any]any{"b": "b", "c": "c"}}},
		{Script: `a = {"b": "b", "c":"c"}; delete(a, "b")`, Output: map[string]any{"a": map[any]any{"c": "c"}}},

		{Script: `a = {"b": ["b"]}; delete(a, "a")`, Output: map[string]any{"a": map[any]any{"b": []any{"b"}}}},
		{Script: `a = {"b": ["b"]}; delete(a, "b")`, Output: map[string]any{"a": map[any]any{}}},
		{Script: `a = {"b": ["b"], "c": ["c"]}; delete(a, "a")`, Output: map[string]any{"a": map[any]any{"b": []any{"b"}, "c": []any{"c"}}}},
		{Script: `a = {"b": ["b"], "c": ["c"]}; delete(a, "b")`, Output: map[string]any{"a": map[any]any{"c": []any{"c"}}}},

		{Script: `a = {"b": ["b"]}; b = &a; delete(*b, "a")`, Output: map[string]any{"a": map[any]any{"b": []any{"b"}}}},
		{Script: `a = {"b": ["b"]}; b = &a; delete(*b, "b")`, Output: map[string]any{"a": map[any]any{}}},
		{Script: `a = {"b": ["b"], "c": ["c"]}; b = &a; delete(*b, "a")`, Output: map[string]any{"a": map[any]any{"b": []any{"b"}, "c": []any{"c"}}}},
		{Script: `a = {"b": ["b"], "c": ["c"]}; b = &a; delete(*b, "b")`, Output: map[string]any{"a": map[any]any{"c": []any{"c"}}}},
	}
	runTests(t, tests, nil)
}

func TestMakeMaps(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make(mapStringBool)`, Types: map[string]any{"mapStringBool": map[string]bool{}}, RunOutput: map[string]bool{}},
		{Script: `make(mapStringInt32)`, Types: map[string]any{"mapStringInt32": map[string]int32{}}, RunOutput: map[string]int32{}},
		{Script: `make(mapStringInt64)`, Types: map[string]any{"mapStringInt64": map[string]int64{}}, RunOutput: map[string]int64{}},
		{Script: `make(mapStringFloat32)`, Types: map[string]any{"mapStringFloat32": map[string]float32{}}, RunOutput: map[string]float32{}},
		{Script: `make(mapStringFloat64)`, Types: map[string]any{"mapStringFloat64": map[string]float64{}}, RunOutput: map[string]float64{}},
		{Script: `make(mapStringString)`, Types: map[string]any{"mapStringString": map[string]string{}}, RunOutput: map[string]string{}},

		{Script: `a = make(mapStringBool)`, Types: map[string]any{"mapStringBool": map[string]bool{}}, RunOutput: map[string]bool{}, Output: map[string]any{"a": map[string]bool{}}},
		{Script: `a = make(mapStringInt32)`, Types: map[string]any{"mapStringInt32": map[string]int32{}}, RunOutput: map[string]int32{}, Output: map[string]any{"a": map[string]int32{}}},
		{Script: `a = make(mapStringInt64)`, Types: map[string]any{"mapStringInt64": map[string]int64{}}, RunOutput: map[string]int64{}, Output: map[string]any{"a": map[string]int64{}}},
		{Script: `a = make(mapStringFloat32)`, Types: map[string]any{"mapStringFloat32": map[string]float32{}}, RunOutput: map[string]float32{}, Output: map[string]any{"a": map[string]float32{}}},
		{Script: `a = make(mapStringFloat64)`, Types: map[string]any{"mapStringFloat64": map[string]float64{}}, RunOutput: map[string]float64{}, Output: map[string]any{"a": map[string]float64{}}},
		{Script: `a = make(mapStringString)`, Types: map[string]any{"mapStringString": map[string]string{}}, RunOutput: map[string]string{}, Output: map[string]any{"a": map[string]string{}}},

		{Script: `a = make(mapStringBool); a["b"] = true`, Types: map[string]any{"mapStringBool": map[string]bool{"b": true}}, RunOutput: true, Output: map[string]any{"a": map[string]bool{"b": true}}},
		{Script: `a = make(mapStringInt32); a["b"] = 1`, Types: map[string]any{"mapStringInt32": map[string]int32{"b": int32(1)}}, RunOutput: int64(1), Output: map[string]any{"a": map[string]int32{"b": int32(1)}}},
		{Script: `a = make(mapStringInt64); a["b"] = 1`, Types: map[string]any{"mapStringInt64": map[string]int64{"b": int64(1)}}, RunOutput: int64(1), Output: map[string]any{"a": map[string]int64{"b": int64(1)}}},
		{Script: `a = make(mapStringFloat32); a["b"] = 1.1`, Types: map[string]any{"mapStringFloat32": map[string]float32{"b": float32(1.1)}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[string]float32{"b": float32(1.1)}}},
		{Script: `a = make(mapStringFloat64); a["b"] = 1.1`, Types: map[string]any{"mapStringFloat64": map[string]float64{"b": float64(1.1)}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[string]float64{"b": float64(1.1)}}},
		{Script: `a = make(mapStringString); a["b"] = "b"`, Types: map[string]any{"mapStringString": map[string]string{"b": "b"}}, RunOutput: "b", Output: map[string]any{"a": map[string]string{"b": "b"}}},

		{Script: `a = make(mapStringBool); a.b = true`, Types: map[string]any{"mapStringBool": map[string]bool{"b": true}}, RunOutput: true, Output: map[string]any{"a": map[string]bool{"b": true}}},

		{Script: `a = make(mapInterfaceBool); a["b"] = true; a["b"]`, Types: map[string]any{"mapInterfaceBool": map[any]bool{}}, RunOutput: true, Output: map[string]any{"a": map[any]bool{"b": true}}},
		{Script: `a = make(mapInterfaceInt32); a["b"] = 1; a["b"]`, Types: map[string]any{"mapInterfaceInt32": map[any]int32{}}, RunOutput: int32(1), Output: map[string]any{"a": map[any]int32{"b": int32(1)}}},
		{Script: `a = make(mapInterfaceInt64); a["b"] = 1; a["b"]`, Types: map[string]any{"mapInterfaceInt64": map[any]int64{}}, RunOutput: int64(1), Output: map[string]any{"a": map[any]int64{"b": int64(1)}}},
		{Script: `a = make(mapInterfaceFloat32); a["b"] = 1.1; a["b"]`, Types: map[string]any{"mapInterfaceFloat32": map[any]float32{}}, RunOutput: float32(1.1), Output: map[string]any{"a": map[any]float32{"b": float32(1.1)}}},
		{Script: `a = make(mapInterfaceFloat64); a["b"] = 1.1; a["b"]`, Types: map[string]any{"mapInterfaceFloat64": map[any]float64{}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]float64{"b": float64(1.1)}}},
		{Script: `a = make(mapInterfaceString); a["b"] = "b"; a["b"]`, Types: map[string]any{"mapInterfaceString": map[any]string{}}, RunOutput: "b", Output: map[string]any{"a": map[any]string{"b": "b"}}},

		{Script: `a = make(mapInterfaceBool); a.b = true; a.b`, Types: map[string]any{"mapInterfaceBool": map[any]bool{}}, RunOutput: true, Output: map[string]any{"a": map[any]bool{"b": true}}},
		{Script: `a = make(mapInterfaceInt32); a.b = 1; a.b`, Types: map[string]any{"mapInterfaceInt32": map[any]int32{}}, RunOutput: int32(1), Output: map[string]any{"a": map[any]int32{"b": int32(1)}}},
		{Script: `a = make(mapInterfaceInt64); a.b = 1; a.b`, Types: map[string]any{"mapInterfaceInt64": map[any]int64{}}, RunOutput: int64(1), Output: map[string]any{"a": map[any]int64{"b": int64(1)}}},
		{Script: `a = make(mapInterfaceFloat32); a.b = 1.1; a.b`, Types: map[string]any{"mapInterfaceFloat32": map[any]float32{}}, RunOutput: float32(1.1), Output: map[string]any{"a": map[any]float32{"b": float32(1.1)}}},
		{Script: `a = make(mapInterfaceFloat64); a.b = 1.1; a.b`, Types: map[string]any{"mapInterfaceFloat64": map[any]float64{}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]float64{"b": float64(1.1)}}},
		{Script: `a = make(mapInterfaceString); a.b = "b"; a.b`, Types: map[string]any{"mapInterfaceString": map[any]string{}}, RunOutput: "b", Output: map[string]any{"a": map[any]string{"b": "b"}}},
	}
	runTests(t, tests, nil)
}

func TestArraysAndMaps(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = [{"b": nil}]`, RunOutput: []any{map[any]any{"b": any(nil)}}, Output: map[string]any{"a": []any{map[any]any{"b": any(nil)}}}},
		{Script: `a = [{"b": true}]`, RunOutput: []any{map[any]any{"b": any(true)}}, Output: map[string]any{"a": []any{map[any]any{"b": any(true)}}}},
		{Script: `a = [{"b": 1}]`, RunOutput: []any{map[any]any{"b": any(int64(1))}}, Output: map[string]any{"a": []any{map[any]any{"b": any(int64(1))}}}},
		{Script: `a = [{"b": 1.1}]`, RunOutput: []any{map[any]any{"b": any(float64(1.1))}}, Output: map[string]any{"a": []any{map[any]any{"b": any(float64(1.1))}}}},
		{Script: `a = [{"b": "b"}]`, RunOutput: []any{map[any]any{"b": any("b")}}, Output: map[string]any{"a": []any{map[any]any{"b": any("b")}}}},

		{Script: `a = [{"b": nil}]; a[0]`, RunOutput: map[any]any{"b": any(nil)}, Output: map[string]any{"a": []any{map[any]any{"b": any(nil)}}}},
		{Script: `a = [{"b": true}]; a[0]`, RunOutput: map[any]any{"b": any(true)}, Output: map[string]any{"a": []any{map[any]any{"b": any(true)}}}},
		{Script: `a = [{"b": 1}]; a[0]`, RunOutput: map[any]any{"b": any(int64(1))}, Output: map[string]any{"a": []any{map[any]any{"b": any(int64(1))}}}},
		{Script: `a = [{"b": 1.1}]; a[0]`, RunOutput: map[any]any{"b": any(float64(1.1))}, Output: map[string]any{"a": []any{map[any]any{"b": any(float64(1.1))}}}},
		{Script: `a = [{"b": "b"}]; a[0]`, RunOutput: map[any]any{"b": any("b")}, Output: map[string]any{"a": []any{map[any]any{"b": any("b")}}}},

		{Script: `a = {"b": []}`, RunOutput: map[any]any{"b": []any{}}, Output: map[string]any{"a": map[any]any{"b": []any{}}}},
		{Script: `a = {"b": [nil]}`, RunOutput: map[any]any{"b": []any{nil}}, Output: map[string]any{"a": map[any]any{"b": []any{nil}}}},
		{Script: `a = {"b": [true]}`, RunOutput: map[any]any{"b": []any{true}}, Output: map[string]any{"a": map[any]any{"b": []any{true}}}},
		{Script: `a = {"b": [1]}`, RunOutput: map[any]any{"b": []any{int64(1)}}, Output: map[string]any{"a": map[any]any{"b": []any{int64(1)}}}},
		{Script: `a = {"b": [1.1]}`, RunOutput: map[any]any{"b": []any{float64(1.1)}}, Output: map[string]any{"a": map[any]any{"b": []any{float64(1.1)}}}},
		{Script: `a = {"b": ["b"]}`, RunOutput: map[any]any{"b": []any{"b"}}, Output: map[string]any{"a": map[any]any{"b": []any{"b"}}}},

		{Script: `a = {"b": []}; a.b`, RunOutput: []any{}, Output: map[string]any{"a": map[any]any{"b": []any{}}}},
		{Script: `a = {"b": [nil]}; a.b`, RunOutput: []any{nil}, Output: map[string]any{"a": map[any]any{"b": []any{nil}}}},
		{Script: `a = {"b": [true]}; a.b`, RunOutput: []any{true}, Output: map[string]any{"a": map[any]any{"b": []any{true}}}},
		{Script: `a = {"b": [1]}; a.b`, RunOutput: []any{int64(1)}, Output: map[string]any{"a": map[any]any{"b": []any{int64(1)}}}},
		{Script: `a = {"b": [1.1]}; a.b`, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": map[any]any{"b": []any{float64(1.1)}}}},
		{Script: `a = {"b": ["b"]}; a.b`, RunOutput: []any{"b"}, Output: map[string]any{"a": map[any]any{"b": []any{"b"}}}},

		{Script: `a.b = []`, Input: map[string]any{"a": map[string][]any{}}, RunOutput: []any{}, Output: map[string]any{"a": map[string][]any{"b": {}}}},
		{Script: `a.b = [nil]`, Input: map[string]any{"a": map[string][]any{}}, RunOutput: []any{any(nil)}, Output: map[string]any{"a": map[string][]any{"b": {any(nil)}}}},
		{Script: `a.b = [true]`, Input: map[string]any{"a": map[string][]any{}}, RunOutput: []any{true}, Output: map[string]any{"a": map[string][]any{"b": {true}}}},
		{Script: `a.b = [1]`, Input: map[string]any{"a": map[string][]any{}}, RunOutput: []any{int64(1)}, Output: map[string]any{"a": map[string][]any{"b": {int64(1)}}}},
		{Script: `a.b = [1.1]`, Input: map[string]any{"a": map[string][]any{}}, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": map[string][]any{"b": {float64(1.1)}}}},
		{Script: `a.b = ["b"]`, Input: map[string]any{"a": map[string][]any{}}, RunOutput: []any{"b"}, Output: map[string]any{"a": map[string][]any{"b": {"b"}}}},

		{Script: `b[0] = [nil]; a.b = b`, Input: map[string]any{"a": map[string][][]any{}, "b": [][]any{}}, RunOutput: [][]any{{any(nil)}}, Output: map[string]any{"a": map[string][][]any{"b": {{any(nil)}}}, "b": [][]any{{any(nil)}}}},
		{Script: `b[0] = [true]; a.b = b`, Input: map[string]any{"a": map[string][][]any{}, "b": [][]any{}}, RunOutput: [][]any{{true}}, Output: map[string]any{"a": map[string][][]any{"b": {{true}}}, "b": [][]any{{true}}}},
		{Script: `b[0] = [1]; a.b = b`, Input: map[string]any{"a": map[string][][]any{}, "b": [][]any{}}, RunOutput: [][]any{{int64(1)}}, Output: map[string]any{"a": map[string][][]any{"b": {{int64(1)}}}, "b": [][]any{{int64(1)}}}},
		{Script: `b[0] = [1.1]; a.b = b`, Input: map[string]any{"a": map[string][][]any{}, "b": [][]any{}}, RunOutput: [][]any{{float64(1.1)}}, Output: map[string]any{"a": map[string][][]any{"b": {{float64(1.1)}}}, "b": [][]any{{float64(1.1)}}}},
		{Script: `b[0] = ["b"]; a.b = b`, Input: map[string]any{"a": map[string][][]any{}, "b": [][]any{}}, RunOutput: [][]any{{"b"}}, Output: map[string]any{"a": map[string][][]any{"b": {{"b"}}}, "b": [][]any{{"b"}}}},

		{Script: `a`, Input: map[string]any{"a": map[string][][]any{}}, RunOutput: map[string][][]any{}, Output: map[string]any{"a": map[string][][]any{}}},
		{Script: `a.b = 1`, Input: map[string]any{"a": map[string][][]any{}}, RunError: fmt.Errorf("type int64 cannot be assigned to type [][]interface {} for map"), RunOutput: nil, Output: map[string]any{"a": map[string][][]any{}}},
		{Script: `a["b"] = 1`, Input: map[string]any{"a": map[string][][]any{}}, RunError: fmt.Errorf("type int64 cannot be assigned to type [][]interface {} for map"), RunOutput: nil, Output: map[string]any{"a": map[string][][]any{}}},

		{Script: `a = {}; a.b = []; a.b += 1; a.b[0] = {}; a.b[0].c = []; a.b[0].c += 2; a.b[0].c[0]`, RunOutput: int64(2), Output: map[string]any{"a": map[any]any{"b": []any{map[any]any{"c": []any{int64(2)}}}}}},

		{Script: `a = {}; a.b = b`, Input: map[string]any{"b": [][]any{}}, RunOutput: [][]any{}, Output: map[string]any{"a": map[any]any{"b": [][]any{}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b`, Input: map[string]any{"b": [][]any{}}, RunOutput: [][]any{}, Output: map[string]any{"a": map[any]any{"b": [][]any{}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0]`, Input: map[string]any{"b": [][]any{}}, RunError: fmt.Errorf("index out of range"), RunOutput: nil, Output: map[string]any{"a": map[any]any{"b": [][]any{}}, "b": [][]any{}}},

		{Script: `a = {}; a.b = b; a.b[0] = []`, Input: map[string]any{"b": [][]any{}}, RunOutput: []any{}, Output: map[string]any{"a": map[any]any{"b": [][]any{{}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [nil]`, Input: map[string]any{"b": [][]any{}}, RunOutput: []any{nil}, Output: map[string]any{"a": map[any]any{"b": [][]any{{nil}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [nil]; a.b`, Input: map[string]any{"b": [][]any{}}, RunOutput: [][]any{{nil}}, Output: map[string]any{"a": map[any]any{"b": [][]any{{nil}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [true]; a.b`, Input: map[string]any{"b": [][]any{}}, RunOutput: [][]any{{true}}, Output: map[string]any{"a": map[any]any{"b": [][]any{{true}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [1]; a.b`, Input: map[string]any{"b": [][]any{}}, RunOutput: [][]any{{int64(1)}}, Output: map[string]any{"a": map[any]any{"b": [][]any{{int64(1)}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [1.1]; a.b`, Input: map[string]any{"b": [][]any{}}, RunOutput: [][]any{{float64(1.1)}}, Output: map[string]any{"a": map[any]any{"b": [][]any{{float64(1.1)}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = ["c"]; a.b`, Input: map[string]any{"b": [][]any{}}, RunOutput: [][]any{{"c"}}, Output: map[string]any{"a": map[any]any{"b": [][]any{{"c"}}}, "b": [][]any{}}},

		{Script: `a = {}; a.b = b; a.b[0] = [nil]; a.b[0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: []any{nil}, Output: map[string]any{"a": map[any]any{"b": [][]any{{nil}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [true]; a.b[0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: []any{true}, Output: map[string]any{"a": map[any]any{"b": [][]any{{true}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [1]; a.b[0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: []any{int64(1)}, Output: map[string]any{"a": map[any]any{"b": [][]any{{int64(1)}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [1.1]; a.b[0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: []any{float64(1.1)}, Output: map[string]any{"a": map[any]any{"b": [][]any{{float64(1.1)}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = ["c"]; a.b[0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: []any{"c"}, Output: map[string]any{"a": map[any]any{"b": [][]any{{"c"}}}, "b": [][]any{}}},

		{Script: `a = {}; a.b = b; a.b[0] = [nil]; a.b[0][0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: nil, Output: map[string]any{"a": map[any]any{"b": [][]any{{nil}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [true]; a.b[0][0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": [][]any{{true}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [1]; a.b[0][0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"b": [][]any{{int64(1)}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [1.1]; a.b[0][0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: float64(1.1), Output: map[string]any{"a": map[any]any{"b": [][]any{{float64(1.1)}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = ["c"]; a.b[0][0]`, Input: map[string]any{"b": [][]any{}}, RunOutput: "c", Output: map[string]any{"a": map[any]any{"b": [][]any{{"c"}}}, "b": [][]any{}}},

		{Script: `a = {}; a.b = b; a.b[0] = [nil]; a.b[0][1] = nil`, Input: map[string]any{"b": [][]any{}}, RunOutput: nil, Output: map[string]any{"a": map[any]any{"b": [][]any{{nil, nil}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [true]; a.b[0][1] = true`, Input: map[string]any{"b": [][]any{}}, RunOutput: true, Output: map[string]any{"a": map[any]any{"b": [][]any{{true, true}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [1]; a.b[0][1] = 2`, Input: map[string]any{"b": [][]any{}}, RunOutput: int64(2), Output: map[string]any{"a": map[any]any{"b": [][]any{{int64(1), int64(2)}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = [1.1]; a.b[0][1] = 2.2`, Input: map[string]any{"b": [][]any{}}, RunOutput: float64(2.2), Output: map[string]any{"a": map[any]any{"b": [][]any{{float64(1.1), float64(2.2)}}}, "b": [][]any{}}},
		{Script: `a = {}; a.b = b; a.b[0] = ["c"]; a.b[0][1] = "d"`, Input: map[string]any{"b": [][]any{}}, RunOutput: "d", Output: map[string]any{"a": map[any]any{"b": [][]any{{"c", "d"}}}, "b": [][]any{}}},
	}
	runTests(t, tests, nil)
}

func TestMakeArraysAndMaps(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make([]map)`, Types: map[string]any{"map": map[string]any{}}, RunOutput: []map[string]any{}},
		{Script: `make([][]map)`, Types: map[string]any{"map": map[string]any{}}, RunOutput: [][]map[string]any{}},

		{Script: `make(mapArray2x)`, Types: map[string]any{"mapArray2x": map[string][][]any{}}, RunOutput: map[string][][]any{}},
		{Script: `a = make(mapArray2x)`, Types: map[string]any{"mapArray2x": map[string][][]any{}}, RunOutput: map[string][][]any{}, Output: map[string]any{"a": map[string][][]any{}}},
		{Script: `a = make(mapArray2x); a`, Types: map[string]any{"mapArray2x": map[string][][]any{}}, RunOutput: map[string][][]any{}, Output: map[string]any{"a": map[string][][]any{}}},
		{Script: `a = make(mapArray2x); a.b = b`, Types: map[string]any{"mapArray2x": map[string][][]any{}}, Input: map[string]any{"b": [][]any{}}, RunOutput: [][]any{}, Output: map[string]any{"a": map[string][][]any{"b": {}}, "b": [][]any{}}},
	}
	runTests(t, tests, nil)
}

func TestMakeArraysData(t *testing.T) {
	stmts, err := parser.ParseSrc("make(array)")
	if err != nil {
		t.Errorf("ParseSrc error - received %v - expected: %v", err, nil)
	}

	v := New(nil)
	err = v.DefineType("array", []string{})
	if err != nil {
		t.Errorf("DefineType error - received %v - expected: %v", err, nil)
	}

	value, err := v.Executor().Run(nil, stmts)
	if err != nil {
		t.Errorf("Run error - received %v - expected: %v", err, nil)
	}
	if !reflect.DeepEqual(value, []string{}) {
		t.Errorf("Run value - received %#v - expected: %#v", value, []string{})
	}

	a := value.([]string)
	if len(a) != 0 {
		t.Errorf("len value - received %#v - expected: %#v", len(a), 0)
	}
	a = append(a, "a")
	if a[0] != "a" {
		t.Errorf("Get value - received %#v - expected: %#v", a[0], "a")
	}
	if len(a) != 1 {
		t.Errorf("len value - received %#v - expected: %#v", len(a), 1)
	}

	stmts, err = parser.ParseSrc("make([]string)")
	if err != nil {
		t.Errorf("ParseSrc error - received %v - expected: %v", err, nil)
	}

	v = New(nil)
	err = v.DefineType("string", "a")
	if err != nil {
		t.Errorf("DefineType error - received %v - expected: %v", err, nil)
	}

	value, err = v.Executor().Run(nil, stmts)
	if err != nil {
		t.Errorf("Run error - received %v - expected: %v", err, nil)
	}
	if !reflect.DeepEqual(value, []string{}) {
		t.Errorf("Run value - received %#v - expected: %#v", value, []string{})
	}

	b := value.([]string)
	if len(b) != 0 {
		t.Errorf("len value - received %#v - expected: %#v", len(b), 0)
	}
	b = append(b, "b")
	if b[0] != "b" {
		t.Errorf("Get value - received %#v - expected: %#v", b[0], "b")
	}
	if len(b) != 1 {
		t.Errorf("len value - received %#v - expected: %#v", len(b), 1)
	}
}

func TestMakeMapsData(t *testing.T) {
	stmts, err := parser.ParseSrc("make(map)")
	if err != nil {
		t.Errorf("ParseSrc error - received %v - expected: %v", err, nil)
	}

	// test normal map
	v := New(nil)
	err = v.DefineType("map", map[string]string{})
	if err != nil {
		t.Errorf("DefineType error - received %v - expected: %v", err, nil)
	}

	value, err := v.Executor().Run(nil, stmts)
	if err != nil {
		t.Errorf("Run error - received %v - expected: %v", err, nil)
	}
	if !reflect.DeepEqual(value, map[string]string{}) {
		t.Errorf("Run value - received %#v - expected: %#v", value, map[string]string{})
	}

	a := value.(map[string]string)
	a["a"] = "a"
	if a["a"] != "a" {
		t.Errorf("Get value - received %#v - expected: %#v", a["a"], "a")
	}

	// test url Values map
	v = New(nil)
	err = v.DefineType("map", url.Values{})
	if err != nil {
		t.Errorf("DefineType error - received %v - expected: %v", err, nil)
	}

	value, err = v.Executor().Run(nil, stmts)
	if err != nil {
		t.Errorf("Run error - received %v - expected: %v", err, nil)
	}
	if !reflect.DeepEqual(value, url.Values{}) {
		t.Errorf("Run value - received %#v - expected: %#v", value, url.Values{})
	}

	b := value.(url.Values)
	b.Set("b", "b")
	if b.Get("b") != "b" {
		t.Errorf("Get value - received %#v - expected: %#v", b.Get("b"), "b")
	}
}

func TestStructs(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a["B"]`, Input: map[string]any{"a": struct {
			A any
			B any
		}{}},
			RunError: fmt.Errorf("type struct does not support index operation"),
			Output: map[string]any{"a": struct {
				A any
				B any
			}{}}},
		{Script: `a.C`, Input: map[string]any{"a": struct {
			A any
			B any
		}{}},
			RunError: fmt.Errorf("no member named 'C' for struct"),
			Output: map[string]any{"a": struct {
				A any
				B any
			}{}}},

		{Script: `a.B`, Input: map[string]any{"a": struct {
			A any
			B any
		}{}},
			RunOutput: nil,
			Output: map[string]any{"a": struct {
				A any
				B any
			}{}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A any
			B any
		}{A: nil, B: nil}},
			RunOutput: nil,
			Output: map[string]any{"a": struct {
				A any
				B any
			}{A: nil, B: nil}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A any
			B any
		}{A: int32(1), B: int32(2)}},
			RunOutput: int32(2),
			Output: map[string]any{"a": struct {
				A any
				B any
			}{A: int32(1), B: int32(2)}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A any
			B any
		}{A: int64(1), B: int64(2)}},
			RunOutput: int64(2),
			Output: map[string]any{"a": struct {
				A any
				B any
			}{A: int64(1), B: int64(2)}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A any
			B any
		}{A: float32(1.1), B: float32(2.2)}},
			RunOutput: float32(2.2),
			Output: map[string]any{"a": struct {
				A any
				B any
			}{A: float32(1.1), B: float32(2.2)}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A any
			B any
		}{A: float64(1.1), B: float64(2.2)}},
			RunOutput: float64(2.2),
			Output: map[string]any{"a": struct {
				A any
				B any
			}{A: float64(1.1), B: float64(2.2)}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A any
			B any
		}{A: "a", B: "b"}},
			RunOutput: "b",
			Output: map[string]any{"a": struct {
				A any
				B any
			}{A: "a", B: "b"}}},

		{Script: `a.B`, Input: map[string]any{"a": struct {
			A bool
			B bool
		}{}},
			RunOutput: false,
			Output: map[string]any{"a": struct {
				A bool
				B bool
			}{}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A int32
			B int32
		}{}},
			RunOutput: int32(0),
			Output: map[string]any{"a": struct {
				A int32
				B int32
			}{}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A int64
			B int64
		}{}},
			RunOutput: int64(0),
			Output: map[string]any{"a": struct {
				A int64
				B int64
			}{}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A float32
			B float32
		}{}},
			RunOutput: float32(0),
			Output: map[string]any{"a": struct {
				A float32
				B float32
			}{}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A float64
			B float64
		}{}},
			RunOutput: float64(0),
			Output: map[string]any{"a": struct {
				A float64
				B float64
			}{}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A string
			B string
		}{}},
			RunOutput: "",
			Output: map[string]any{"a": struct {
				A string
				B string
			}{}}},

		{Script: `a.B`, Input: map[string]any{"a": struct {
			A bool
			B bool
		}{A: true, B: true}},
			RunOutput: true,
			Output: map[string]any{"a": struct {
				A bool
				B bool
			}{A: true, B: true}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A int32
			B int32
		}{A: int32(1), B: int32(2)}},
			RunOutput: int32(2),
			Output: map[string]any{"a": struct {
				A int32
				B int32
			}{A: int32(1), B: int32(2)}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A int64
			B int64
		}{A: int64(1), B: int64(2)}},
			RunOutput: int64(2),
			Output: map[string]any{"a": struct {
				A int64
				B int64
			}{A: int64(1), B: int64(2)}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A float32
			B float32
		}{A: float32(1.1), B: float32(2.2)}},
			RunOutput: float32(2.2),
			Output: map[string]any{"a": struct {
				A float32
				B float32
			}{A: float32(1.1), B: float32(2.2)}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A float64
			B float64
		}{A: float64(1.1), B: float64(2.2)}},
			RunOutput: float64(2.2),
			Output: map[string]any{"a": struct {
				A float64
				B float64
			}{A: float64(1.1), B: float64(2.2)}}},
		{Script: `a.B`, Input: map[string]any{"a": struct {
			A string
			B string
		}{A: "a", B: "b"}},
			RunOutput: "b",
			Output: map[string]any{"a": struct {
				A string
				B string
			}{A: "a", B: "b"}}},

		{Script: `a.C = 3`, Input: map[string]any{
			"a": struct {
				A any
				B any
			}{A: int64(1), B: int64(2)},
		},
			RunError: fmt.Errorf("no member named 'C' for struct"),
			Output: map[string]any{"a": struct {
				A any
				B any
			}{A: int64(1), B: int64(2)}}},

		{Script: `a.B = 3`, Input: map[string]any{"a": &struct {
			A any
			B any
		}{A: int64(1), B: int64(2)}},
			RunOutput: int64(3),
			Output: map[string]any{"a": &struct {
				A any
				B any
			}{A: int64(1), B: int64(3)}}},

		{Script: `a.B = 3; a = *a`, Input: map[string]any{"a": &struct {
			A any
			B any
		}{A: int64(1), B: int64(2)}},
			RunOutput: struct {
				A any
				B any
			}{A: int64(1), B: int64(3)},
			Output: map[string]any{"a": struct {
				A any
				B any
			}{A: int64(1), B: int64(3)}}},

		// nil tests
		{Script: `a.A = nil; a.B = nil; a.C = nil;`, Input: map[string]any{"a": &struct {
			A *bool
			B *int32
			C *int64
		}{A: new(bool), B: new(int32), C: new(int64)}},
			RunOutput: nil,
			Output: map[string]any{"a": &struct {
				A *bool
				B *int32
				C *int64
			}{A: nil, B: nil, C: nil}}},
		{Script: `a.A = nil; a.B = nil; a.C = nil;`, Input: map[string]any{"a": &struct {
			A *float32
			B *float64
			C *string
		}{A: new(float32), B: new(float64), C: new(string)}},
			RunOutput: nil,
			Output: map[string]any{"a": &struct {
				A *float32
				B *float64
				C *string
			}{A: nil, B: nil, C: nil}}},
		{Script: `a.A = nil; a.B = nil; a.C = nil;`, Input: map[string]any{"a": &struct {
			A any
			B []any
			C [][]any
		}{A: "a", B: []any{"b"}, C: [][]any{[]any{"c"}}}},
			RunOutput: nil,
			Output: map[string]any{"a": &struct {
				A any
				B []any
				C [][]any
			}{A: nil, B: nil, C: nil}}},
		{Script: `a.A = nil; a.B = nil; a.C = nil; a.D = nil;`, Input: map[string]any{"a": &struct {
			A map[string]string
			B map[string]any
			C map[any]string
			D map[any]any
		}{A: map[string]string{"a": "a"}, B: map[string]any{"b": "b"}, C: map[any]string{"c": "c"}, D: map[any]any{"d": "d"}}},
			RunOutput: nil,
			Output: map[string]any{"a": &struct {
				A map[string]string
				B map[string]any
				C map[any]string
				D map[any]any
			}{A: nil, B: nil, C: nil, D: nil}}},
		{Script: `a.A.AA = nil;`, Input: map[string]any{"a": &struct {
			A struct{ AA *int64 }
		}{A: struct{ AA *int64 }{AA: new(int64)}}},
			RunOutput: nil,
			Output: map[string]any{"a": &struct {
				A struct{ AA *int64 }
			}{A: struct{ AA *int64 }{AA: nil}}}},
		{Script: `a.A = nil;`, Input: map[string]any{"a": &struct {
			A *struct{ AA *int64 }
		}{A: &struct{ AA *int64 }{AA: new(int64)}}},
			RunOutput: nil,
			Output: map[string]any{"a": &struct {
				A *struct{ AA *int64 }
			}{A: nil}}},
	}
	runTests(t, tests, nil)
}

func TestMakeStructs(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `make(struct)`, Types: map[string]any{"struct": &struct {
			A any
			B any
		}{}},
			RunOutput: &struct {
				A any
				B any
			}{}},

		{Script: `a = make(struct)`, Types: map[string]any{"struct": &struct {
			A any
			B any
		}{}},
			RunOutput: &struct {
				A any
				B any
			}{},
			Output: map[string]any{"a": &struct {
				A any
				B any
			}{}}},

		{Script: `a = make(struct); a.A = 3; a.B = 4`, Types: map[string]any{"struct": &struct {
			A any
			B any
		}{}},
			RunOutput: int64(4),
			Output: map[string]any{"a": &struct {
				A any
				B any
			}{A: any(int64(3)), B: any(int64(4))}}},

		{Script: `a = make(struct); a = *a; a.A = 3; a.B = 4`, Types: map[string]any{"struct": &struct {
			A any
			B any
		}{}},
			RunOutput: int64(4),
			Output: map[string]any{"a": struct {
				A any
				B any
			}{A: any(int64(3)), B: any(int64(4))}}},

		{Script: `a = make(struct); a.A = func () { return 1 }; a.A()`, Types: map[string]any{"struct": &struct {
			A any
			B any
		}{}},
			RunOutput: int64(1)},
		{Script: `a = make(struct); a.A = func () { return 1 }; a = *a; a.A()`, Types: map[string]any{"struct": &struct {
			A any
			B any
		}{}},
			RunOutput: int64(1)},
	}
	runTests(t, tests, nil)
}
