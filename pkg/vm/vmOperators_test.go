package vm

import (
	"fmt"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"os"
	"reflect"
	"testing"
)

func TestBasicOperators(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `]`, ParseError: fmt.Errorf("unexpected ']'")},

		{Script: `2 + 1`, RunOutput: int64(3)},
		{Script: `2 - 1`, RunOutput: int64(1)},
		{Script: `2 * 1`, RunOutput: int64(2)},
		{Script: `2 / 1`, RunOutput: float64(2)},
		{Script: `2.1 + 1.1`, RunOutput: float64(3.2)},
		{Script: `2.1 - 1.1`, RunOutput: float64(1)},
		{Script: `2.1 * 2.0`, RunOutput: float64(4.2)},
		{Script: `6.5 / 2.0`, RunOutput: float64(3.25)},

		{Script: `a + b`, Input: map[string]any{"a": int64(2), "b": int64(1)}, RunOutput: int64(3)},
		{Script: `a - b`, Input: map[string]any{"a": int64(2), "b": int64(1)}, RunOutput: int64(1)},
		{Script: `a * b`, Input: map[string]any{"a": int64(2), "b": int64(1)}, RunOutput: int64(2)},
		{Script: `a / b`, Input: map[string]any{"a": int64(2), "b": int64(1)}, RunOutput: float64(2)},
		{Script: `a + b`, Input: map[string]any{"a": float64(2.1), "b": float64(1.1)}, RunOutput: float64(3.2)},
		{Script: `a - b`, Input: map[string]any{"a": float64(2.1), "b": float64(1.1)}, RunOutput: float64(1)},
		{Script: `a * b`, Input: map[string]any{"a": float64(2.1), "b": float64(2)}, RunOutput: float64(4.2)},
		{Script: `a / b`, Input: map[string]any{"a": float64(6.5), "b": float64(2)}, RunOutput: float64(3.25)},

		{Script: `a + b`, Input: map[string]any{"a": "a", "b": "b"}, RunOutput: "ab"},
		{Script: `a + b`, Input: map[string]any{"a": "a", "b": int64(1)}, RunOutput: "a1"},
		{Script: `a + b`, Input: map[string]any{"a": "a", "b": float64(1.1)}, RunOutput: "a1.1"},
		{Script: `a + z`, Input: map[string]any{"a": "a"}, RunError: envPkg.NewUndefinedSymbolErr("z"), RunOutput: nil},
		{Script: `z + b`, Input: map[string]any{"a": "a"}, RunError: envPkg.NewUndefinedSymbolErr("z"), RunOutput: nil},

		{Script: `c = a + b`, Input: map[string]any{"a": int64(2), "b": int64(1)}, RunOutput: int64(3), Output: map[string]any{"c": int64(3)}},
		{Script: `c = a - b`, Input: map[string]any{"a": int64(2), "b": int64(1)}, RunOutput: int64(1), Output: map[string]any{"c": int64(1)}},
		{Script: `c = a * b`, Input: map[string]any{"a": int64(2), "b": int64(1)}, RunOutput: int64(2), Output: map[string]any{"c": int64(2)}},
		{Script: `c = a / b`, Input: map[string]any{"a": int64(2), "b": int64(1)}, RunOutput: float64(2), Output: map[string]any{"c": float64(2)}},
		{Script: `c = a + b`, Input: map[string]any{"a": float64(2.1), "b": float64(1.1)}, RunOutput: float64(3.2), Output: map[string]any{"c": float64(3.2)}},
		{Script: `c = a - b`, Input: map[string]any{"a": float64(2.1), "b": float64(1.1)}, RunOutput: float64(1), Output: map[string]any{"c": float64(1)}},
		{Script: `c = a * b`, Input: map[string]any{"a": float64(2.1), "b": float64(2)}, RunOutput: float64(4.2), Output: map[string]any{"c": float64(4.2)}},
		{Script: `c = a / b`, Input: map[string]any{"a": float64(6.5), "b": float64(2)}, RunOutput: float64(3.25), Output: map[string]any{"c": float64(3.25)}},

		{Script: `a = nil; a++`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = false; a++`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = true; a++`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = 1; a++`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = 1.5; a++`, RunOutput: float64(2.5), Output: map[string]any{"a": float64(2.5)}},
		{Script: `a = "1"; a++`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = "a"; a++`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},

		{Script: `a = nil; a--`, RunOutput: int64(-1), Output: map[string]any{"a": int64(-1)}},
		{Script: `a = false; a--`, RunOutput: int64(-1), Output: map[string]any{"a": int64(-1)}},
		{Script: `a = true; a--`, RunOutput: int64(0), Output: map[string]any{"a": int64(0)}},
		{Script: `a = 2; a--`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = 2.5; a--`, RunOutput: float64(1.5), Output: map[string]any{"a": float64(1.5)}},
		{Script: `a = "2"; a--`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = "a"; a--`, RunOutput: int64(-1), Output: map[string]any{"a": int64(-1)}},

		{Script: `a++`, Input: map[string]any{"a": nil}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a++`, Input: map[string]any{"a": false}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a++`, Input: map[string]any{"a": true}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a++`, Input: map[string]any{"a": int32(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a++`, Input: map[string]any{"a": int64(1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a++`, Input: map[string]any{"a": float32(1.5)}, RunOutput: float64(2.5), Output: map[string]any{"a": float64(2.5)}},
		{Script: `a++`, Input: map[string]any{"a": float64(1.5)}, RunOutput: float64(2.5), Output: map[string]any{"a": float64(2.5)}},
		{Script: `a++`, Input: map[string]any{"a": "2"}, RunOutput: int64(3), Output: map[string]any{"a": int64(3)}},
		{Script: `a++`, Input: map[string]any{"a": "a"}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},

		{Script: `a--`, Input: map[string]any{"a": nil}, RunOutput: int64(-1), Output: map[string]any{"a": int64(-1)}},
		{Script: `a--`, Input: map[string]any{"a": false}, RunOutput: int64(-1), Output: map[string]any{"a": int64(-1)}},
		{Script: `a--`, Input: map[string]any{"a": true}, RunOutput: int64(0), Output: map[string]any{"a": int64(0)}},
		{Script: `a--`, Input: map[string]any{"a": int32(2)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a--`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a--`, Input: map[string]any{"a": float32(2.5)}, RunOutput: float64(1.5), Output: map[string]any{"a": float64(1.5)}},
		{Script: `a--`, Input: map[string]any{"a": float64(2.5)}, RunOutput: float64(1.5), Output: map[string]any{"a": float64(1.5)}},
		{Script: `a--`, Input: map[string]any{"a": "2"}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a--`, Input: map[string]any{"a": "a"}, RunOutput: int64(-1), Output: map[string]any{"a": int64(-1)}},

		{Script: `1++`, RunError: fmt.Errorf("invalid operation"), RunOutput: nil},
		{Script: `1--`, RunError: fmt.Errorf("invalid operation"), RunOutput: nil},
		{Script: `z++`, RunError: envPkg.NewUndefinedSymbolErr("z"), RunOutput: nil},
		{Script: `z--`, RunError: envPkg.NewUndefinedSymbolErr("z"), RunOutput: nil},
		{Script: `!(1++)`, RunError: fmt.Errorf("invalid operation"), RunOutput: nil},

		{Script: `a += 1`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(3), Output: map[string]any{"a": int64(3)}},
		{Script: `a -= 1`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a *= 2`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(4), Output: map[string]any{"a": int64(4)}},
		{Script: `a /= 2`, Input: map[string]any{"a": int64(2)}, RunOutput: float64(1), Output: map[string]any{"a": float64(1)}},
		{Script: `a += 1`, Input: map[string]any{"a": 2.1}, RunOutput: float64(3.1), Output: map[string]any{"a": float64(3.1)}},
		{Script: `a -= 1`, Input: map[string]any{"a": 2.1}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}},
		{Script: `a *= 2`, Input: map[string]any{"a": 2.1}, RunOutput: float64(4.2), Output: map[string]any{"a": float64(4.2)}},
		{Script: `a /= 2`, Input: map[string]any{"a": 6.5}, RunOutput: float64(3.25), Output: map[string]any{"a": float64(3.25)}},

		{Script: `a ** 3`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(8), Output: map[string]any{"a": int64(2)}},
		{Script: `a ** 3`, Input: map[string]any{"a": float64(2)}, RunOutput: float64(8), Output: map[string]any{"a": float64(2)}},

		{Script: `a &= 1`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(0), Output: map[string]any{"a": int64(0)}},
		{Script: `a &= 2`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a &= 1`, Input: map[string]any{"a": float64(2.1)}, RunOutput: int64(0), Output: map[string]any{"a": int64(0)}},
		{Script: `a &= 2`, Input: map[string]any{"a": float64(2.1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},

		{Script: `a |= 1`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(3), Output: map[string]any{"a": int64(3)}},
		{Script: `a |= 2`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a |= 1`, Input: map[string]any{"a": float64(2.1)}, RunOutput: int64(3), Output: map[string]any{"a": int64(3)}},
		{Script: `a |= 2`, Input: map[string]any{"a": float64(2.1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},

		{Script: `a << 2`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(8), Output: map[string]any{"a": int64(2)}},
		{Script: `a >> 2`, Input: map[string]any{"a": int64(8)}, RunOutput: int64(2), Output: map[string]any{"a": int64(8)}},
		{Script: `a << 2`, Input: map[string]any{"a": float64(2)}, RunOutput: int64(8), Output: map[string]any{"a": float64(2)}},
		{Script: `a >> 2`, Input: map[string]any{"a": float64(8)}, RunOutput: int64(2), Output: map[string]any{"a": float64(8)}},

		{Script: `a % 2`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(0), Output: map[string]any{"a": int64(2)}},
		{Script: `a % 3`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a % 2`, Input: map[string]any{"a": float64(2.1)}, RunOutput: int64(0), Output: map[string]any{"a": float64(2.1)}},
		{Script: `a % 3`, Input: map[string]any{"a": float64(2.1)}, RunOutput: int64(2), Output: map[string]any{"a": float64(2.1)}},

		{Script: `a * 4`, Input: map[string]any{"a": "a"}, RunOutput: "aaaa", Output: map[string]any{"a": "a"}},
		{Script: `a * 4.0`, Input: map[string]any{"a": "a"}, RunOutput: float64(0), Output: map[string]any{"a": "a"}},

		{Script: `-a`, Input: map[string]any{"a": nil}, RunOutput: float64(-0), Output: map[string]any{"a": nil}},
		{Script: `-a`, Input: map[string]any{"a": "a"}, RunOutput: float64(-0), Output: map[string]any{"a": "a"}},
		{Script: `-a`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(-2), Output: map[string]any{"a": int64(2)}},
		{Script: `-a`, Input: map[string]any{"a": float64(2.1)}, RunOutput: float64(-2.1), Output: map[string]any{"a": float64(2.1)}},

		{Script: `^a`, Input: map[string]any{"a": nil}, RunOutput: int64(-1), Output: map[string]any{"a": nil}},
		{Script: `^a`, Input: map[string]any{"a": "a"}, RunOutput: int64(-1), Output: map[string]any{"a": "a"}},
		{Script: `^a`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(-3), Output: map[string]any{"a": int64(2)}},
		{Script: `^a`, Input: map[string]any{"a": float64(2.1)}, RunOutput: int64(-3), Output: map[string]any{"a": float64(2.1)}},

		{Script: `!true`, RunOutput: false},
		{Script: `!1`, RunOutput: false},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestComparisonOperators(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a == 1`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a == 2`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a != 1`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a != 2`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a == 1.0`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a == 2.0`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a != 1.0`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a != 2.0`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},

		{Script: `a == 1`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a == 2`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a != 1`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a != 2`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a == 1.0`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a == 2.0`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a != 1.0`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a != 2.0`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},

		{Script: `a == nil`, Input: map[string]any{"a": nil}, RunOutput: true, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `a == nil`, Input: map[string]any{"a": nil}, RunOutput: true, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `a == nil`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a == nil`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a == 2`, Input: map[string]any{"a": nil}, RunOutput: false, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `a == 2.0`, Input: map[string]any{"a": nil}, RunOutput: false, Output: map[string]any{"a": nil}, Name: ""},

		{Script: `1 == 1.0`, RunOutput: true, Name: ""},
		{Script: `1 != 1.0`, RunOutput: false, Name: ""},
		{Script: `"a" != "a"`, RunOutput: false, Name: ""},

		{Script: `a > 2`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a > 1`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a < 2`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a < 3`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a > 2.0`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a > 1.0`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a < 2.0`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a < 3.0`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},

		{Script: `a > 2`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a > 1`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a < 2`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a < 3`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a > 2.0`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a > 1.0`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a < 2.0`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a < 3.0`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},

		{Script: `a >= 2`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a >= 3`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a <= 2`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a <= 3`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a >= 2.0`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a >= 3.0`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a <= 2.0`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a <= 3.0`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2)}, Name: ""},

		{Script: `a >= 2`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a >= 3`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a <= 2`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a <= 3`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a >= 2.0`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a >= 3.0`, Input: map[string]any{"a": float64(2)}, RunOutput: false, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a <= 2.0`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},
		{Script: `a <= 3.0`, Input: map[string]any{"a": float64(2)}, RunOutput: true, Output: map[string]any{"a": float64(2)}, Name: ""},

		{Script: `1 == 1 && 1  == 1`, RunOutput: true, Name: ""},
		{Script: `1 == 1 && 1  == 2`, RunOutput: false, Name: ""},
		{Script: `1 == 2 && 1  == 1`, RunOutput: false, Name: ""},
		{Script: `1 == 2 && 1  == 2`, RunOutput: false, Name: ""},
		{Script: `1 == 1 || 1  == 1`, RunOutput: true, Name: ""},
		{Script: `1 == 1 || 1  == 2`, RunOutput: true, Name: ""},
		{Script: `1 == 2 || 1  == 1`, RunOutput: true, Name: ""},
		{Script: `1 == 2 || 1  == 2`, RunOutput: false, Name: ""},

		{Script: `true && func(){throw('abcde')}()`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `false && func(){throw('abcde')}()`, RunOutput: false, Name: ""},
		{Script: `true || func(){throw('abcde')}()`, RunOutput: true, Name: ""},
		{Script: `false || func(){throw('abcde')}()`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `true && true && func(){throw('abcde')}()`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `true && false && func(){throw('abcde')}()`, RunOutput: false, Name: ""},
		{Script: `true && func(){throw('abcde')}() && true`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `false && func(){throw('abcde')}() && func(){throw('abcde')}() `, RunOutput: false, Name: ""},

		{Script: `true && func(){throw('abcde')}() || false`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `true && false || func(){throw('abcde')}()`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `true && true || func(){throw('abcde')}()`, RunOutput: true, Name: ""},

		{Script: `true || func(){throw('abcde')}() || func(){throw('abcde')}()`, RunOutput: true, Name: ""},
		{Script: `false || func(){throw('abcde')}() || true`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `false || true || func(){throw('abcde')}()`, RunOutput: true, Name: ""},
		{Script: `false || false || func(){throw('abcde')}()`, RunError: fmt.Errorf("abcde"), Name: ""},

		{Script: `false || false && func(){throw('abcde')}()`, RunOutput: false, Name: ""},
		{Script: `false || true && func(){throw('abcde')}()`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `false || func(){throw('abcde')}() || true`, RunError: fmt.Errorf("abcde"), Name: ""},

		{Script: `1 == 1 && func(){throw('abcde')}()`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `1 == 2 && func(){throw('abcde')}()`, RunOutput: false, Name: ""},
		{Script: `1 == 1 || func(){throw('abcde')}()`, RunOutput: true, Name: ""},
		{Script: `1 == 2 || func(){throw('abcde')}()`, RunError: fmt.Errorf("abcde"), Name: ""},

		{Script: `(true || func(){throw('abcde')}()) && (true || func(){throw('hello')}())`, RunOutput: true, Name: ""},
		{Script: `(true || func(){throw('abcde')}()) && (true && func(){throw('hello')}())`, RunError: fmt.Errorf("hello"), Name: ""},
		{Script: `(true || func(){throw('abcde')}()) || (true && func(){throw('hello')}())`, RunOutput: true, Name: ""},
		{Script: `(true && func(){throw('abcde')}()) && (true && func(){throw('hello')}())`, RunError: fmt.Errorf("abcde"), Name: ""},
		{Script: `(true || func(){throw('abcde')}()) && (false || func(){throw('hello')}())`, RunError: fmt.Errorf("hello"), Name: ""},

		{Script: `true == "1"`, RunOutput: true, Name: ""},
		{Script: `true == "t"`, RunOutput: true, Name: ""},
		{Script: `true == "T"`, RunOutput: true, Name: ""},
		{Script: `true == "true"`, RunOutput: true, Name: ""},
		{Script: `true == "TRUE"`, RunOutput: true, Name: ""},
		{Script: `true == "True"`, RunOutput: true, Name: ""},
		{Script: `true == "false"`, RunOutput: false, Name: ""},
		{Script: `false == "0"`, RunOutput: true, Name: ""},
		{Script: `false == "f"`, RunOutput: true, Name: ""},
		{Script: `false == "F"`, RunOutput: true, Name: ""},
		{Script: `false == "false"`, RunOutput: true, Name: ""},
		{Script: `false == "false"`, RunOutput: true, Name: ""},
		{Script: `false == "FALSE"`, RunOutput: true, Name: ""},
		{Script: `false == "False"`, RunOutput: true, Name: ""},
		{Script: `false == "true"`, RunOutput: false, Name: ""},
		{Script: `false == "foo"`, RunOutput: false, Name: ""},
		{Script: `true == "foo"`, RunOutput: true, Name: ""},

		{Script: `0 == "0"`, RunOutput: true, Name: ""},
		{Script: `"1.0" == 1`, RunOutput: true, Name: ""},
		{Script: `1 == "1"`, RunOutput: true, Name: ""},
		{Script: `0.0 == "0"`, RunOutput: true, Name: ""},
		{Script: `0.0 == "0.0"`, RunOutput: true, Name: ""},
		{Script: `1.0 == "1.0"`, RunOutput: true, Name: ""},
		{Script: `1.2 == "1.2"`, RunOutput: true, Name: ""},
		{Script: `"7" == 7.2`, RunOutput: false, Name: ""},
		{Script: `1.2 == "1"`, RunOutput: false, Name: ""},
		{Script: `"1.1" == 1`, RunOutput: false, Name: ""},
		{Script: `0 == "1"`, RunOutput: false, Name: ""},

		{Script: `a == b`, Input: map[string]any{"a": reflect.Value{}, "b": reflect.Value{}}, RunOutput: true, Output: map[string]any{"a": reflect.Value{}, "b": reflect.Value{}}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": reflect.Value{}, "b": true}, RunOutput: false, Output: map[string]any{"a": reflect.Value{}, "b": true}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": true, "b": reflect.Value{}}, RunOutput: false, Output: map[string]any{"a": true, "b": reflect.Value{}}, Name: ""},

		{Script: `a == b`, Input: map[string]any{"a": nil, "b": nil}, RunOutput: true, Output: map[string]any{"a": nil, "b": nil}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": nil, "b": true}, RunOutput: false, Output: map[string]any{"a": nil, "b": true}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": true, "b": nil}, RunOutput: false, Output: map[string]any{"a": true, "b": nil}, Name: ""},

		{Script: `a == b`, Input: map[string]any{"a": false, "b": false}, RunOutput: true, Output: map[string]any{"a": false, "b": false}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": false, "b": true}, RunOutput: false, Output: map[string]any{"a": false, "b": true}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": true, "b": false}, RunOutput: false, Output: map[string]any{"a": true, "b": false}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": true, "b": true}, RunOutput: true, Output: map[string]any{"a": true, "b": true}, Name: ""},

		{Script: `a == b`, Input: map[string]any{"a": int32(1), "b": int32(1)}, RunOutput: true, Output: map[string]any{"a": int32(1), "b": int32(1)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": int32(1), "b": int32(2)}, RunOutput: false, Output: map[string]any{"a": int32(1), "b": int32(2)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": int32(2), "b": int32(1)}, RunOutput: false, Output: map[string]any{"a": int32(2), "b": int32(1)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": int32(2), "b": int32(2)}, RunOutput: true, Output: map[string]any{"a": int32(2), "b": int32(2)}, Name: ""},

		{Script: `a == b`, Input: map[string]any{"a": int64(1), "b": int64(1)}, RunOutput: true, Output: map[string]any{"a": int64(1), "b": int64(1)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": int64(1), "b": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(1), "b": int64(2)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": int64(2), "b": int64(1)}, RunOutput: false, Output: map[string]any{"a": int64(2), "b": int64(1)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": int64(2), "b": int64(2)}, RunOutput: true, Output: map[string]any{"a": int64(2), "b": int64(2)}, Name: ""},

		{Script: `a == b`, Input: map[string]any{"a": float32(1.1), "b": float32(1.1)}, RunOutput: true, Output: map[string]any{"a": float32(1.1), "b": float32(1.1)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": float32(1.1), "b": float32(2.2)}, RunOutput: false, Output: map[string]any{"a": float32(1.1), "b": float32(2.2)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": float32(2.2), "b": float32(1.1)}, RunOutput: false, Output: map[string]any{"a": float32(2.2), "b": float32(1.1)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": float32(2.2), "b": float32(2.2)}, RunOutput: true, Output: map[string]any{"a": float32(2.2), "b": float32(2.2)}, Name: ""},

		{Script: `a == b`, Input: map[string]any{"a": float64(1.1), "b": float64(1.1)}, RunOutput: true, Output: map[string]any{"a": float64(1.1), "b": float64(1.1)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": float64(1.1), "b": float64(2.2)}, RunOutput: false, Output: map[string]any{"a": float64(1.1), "b": float64(2.2)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": float64(2.2), "b": float64(1.1)}, RunOutput: false, Output: map[string]any{"a": float64(2.2), "b": float64(1.1)}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": float64(2.2), "b": float64(2.2)}, RunOutput: true, Output: map[string]any{"a": float64(2.2), "b": float64(2.2)}, Name: ""},

		{Script: `a == b`, Input: map[string]any{"a": 'a', "b": 'a'}, RunOutput: true, Output: map[string]any{"a": 'a', "b": 'a'}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": 'a', "b": 'b'}, RunOutput: false, Output: map[string]any{"a": 'a', "b": 'b'}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": 'b', "b": 'a'}, RunOutput: false, Output: map[string]any{"a": 'b', "b": 'a'}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": 'b', "b": 'b'}, RunOutput: true, Output: map[string]any{"a": 'b', "b": 'b'}, Name: ""},

		{Script: `a == b`, Input: map[string]any{"a": "a", "b": "a"}, RunOutput: true, Output: map[string]any{"a": "a", "b": "a"}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": "a", "b": "b"}, RunOutput: false, Output: map[string]any{"a": "a", "b": "b"}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": "b", "b": "a"}, RunOutput: false, Output: map[string]any{"a": "b", "b": "a"}, Name: ""},
		{Script: `a == b`, Input: map[string]any{"a": "b", "b": "b"}, RunOutput: true, Output: map[string]any{"a": "b", "b": "b"}, Name: ""},

		{Script: `b = "\"a\""; a == b`, Input: map[string]any{"a": `"a"`}, RunOutput: true, Output: map[string]any{"a": `"a"`, "b": `"a"`}, Name: ""},

		{Script: `a = "test"; a == "test"`, RunOutput: true, Name: ""},
		{Script: `a = "test"; a[0:1] == "t"`, RunOutput: true, Name: ""},
		{Script: `a = "test"; a[0:2] == "te"`, RunOutput: true, Name: ""},
		{Script: `a = "test"; a[1:3] == "es"`, RunOutput: true, Name: ""},
		{Script: `a = "test"; a[0:4] == "test"`, RunOutput: true, Name: ""},

		{Script: `a = "a b"; a[1] == ' '`, RunOutput: true, Name: ""},
		{Script: `a = "test"; a[0] == 't'`, RunOutput: true, Name: ""},
		{Script: `a = "test"; a[1] == 'e'`, RunOutput: true, Name: ""},
		{Script: `a = "test"; a[3] == 't'`, RunOutput: true, Name: ""},

		{Script: `a = "a b"; a[1] != ' '`, RunOutput: false, Name: ""},
		{Script: `a = "test"; a[0] != 't'`, RunOutput: false, Name: ""},
		{Script: `a = "test"; a[1] != 'e'`, RunOutput: false, Name: ""},
		{Script: `a = "test"; a[3] != 't'`, RunOutput: false, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestTernaryOperator(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = 1 ? 2 : panic(2)`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = c ? a : b`, RunError: envPkg.NewUndefinedSymbolErr("c")},
		{Script: `a = a ? a : b`, RunError: envPkg.NewUndefinedSymbolErr("a")},
		{Script: `a = 0; a = a ? a : b`, RunError: envPkg.NewUndefinedSymbolErr("b")},
		{Script: `a = -1 ? 2 : 1`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = true ? 2 : 1`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = false ? 2 : 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = "true" ? 2 : 1`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = "false" ? 2 : 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = "-1" ? 2 : 1`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = "0" ? 2 : 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = "0.0" ? 2 : 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = "2" ? 2 : 1`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = b ? 2 : 1`, Input: map[string]any{"b": int64(0)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b ? 2 : 1`, Input: map[string]any{"b": int64(2)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = b ? 2 : 1`, Input: map[string]any{"b": float64(0.0)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b ? 2 : 1`, Input: map[string]any{"b": float64(2.0)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = b ? 2 : 1.0`, Input: map[string]any{"b": float64(0.0)}, RunOutput: float64(1.0), Output: map[string]any{"a": float64(1.0)}},
		{Script: `a = b ? 2 : 1.0`, Input: map[string]any{"b": float64(0.1)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = b ? 2 : 1.0`, Input: map[string]any{"b": nil}, RunOutput: float64(1.0), Output: map[string]any{"a": float64(1.0)}},
		{Script: `a = nil ? 2 : 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b ? 2 : 1`, Input: map[string]any{"b": []any{}}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b ? 2 : 1`, Input: map[string]any{"b": map[string]any{}}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b[1] ? 2 : 1`, Input: map[string]any{"b": []any{}}, RunError: fmt.Errorf("index out of range")},
		{Script: `a = b[1][2] ? 2 : 1`, Input: map[string]any{"b": []any{}}, RunError: fmt.Errorf("index out of range")},
		{Script: `a = [] ? 2 : 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = [2] ? 2 : 1`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = b ? 2 : 1`, Input: map[string]any{"b": map[string]any{"test": int64(2)}}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = b["test"] ? 2 : 1`, Input: map[string]any{"b": map[string]any{"test": int64(2)}}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = b["test"][1] ? 2 : 1`, Input: map[string]any{"b": map[string]any{"test": 2}}, RunError: fmt.Errorf("type int does not support index operation")},
		{Script: `b = "test"; a = b ? 2 : "empty"`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `b = "test"; a = b[1:3] ? 2 : "empty"`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `b = "test"; a = b[2:2] ? 2 : "empty"`, RunOutput: "empty", Output: map[string]any{"a": "empty"}},
		{Script: `b = "0.0"; a = false ? 2 : b ? 3 : 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `b = "true"; a = false ? 2 : b ? 3 : 1`, RunOutput: int64(3), Output: map[string]any{"a": int64(3)}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestNilCoalescingOperator(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `a = 1 ?? panic(2)`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = c ?? b`, RunError: envPkg.NewUndefinedSymbolErr("b")},
		{Script: `a = -1 ?? 1`, RunOutput: int64(-1), Output: map[string]any{"a": int64(-1)}},
		{Script: `a = true ?? 1`, RunOutput: true, Output: map[string]any{"a": true}},
		{Script: `a = false ?? 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = "true" ?? 1`, RunOutput: "true", Output: map[string]any{"a": "true"}},
		{Script: `a = "false" ?? 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = "-1" ?? 1`, RunOutput: "-1", Output: map[string]any{"a": "-1"}},
		{Script: `a = "0" ?? 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = "0.0" ?? 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = "2" ?? 1`, RunOutput: "2", Output: map[string]any{"a": "2"}},
		{Script: `a = b ?? 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b ?? 1`, Input: map[string]any{"b": int64(0)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b ?? 1`, Input: map[string]any{"b": int64(2)}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = b ?? 1`, Input: map[string]any{"b": float64(0.0)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b ?? 1`, Input: map[string]any{"b": float64(2.0)}, RunOutput: float64(2.0), Output: map[string]any{"a": float64(2.0)}},
		{Script: `a = b ?? 1.0`, Input: map[string]any{"b": float64(0.0)}, RunOutput: float64(1.0), Output: map[string]any{"a": float64(1.0)}},
		{Script: `a = b ?? 1.0`, Input: map[string]any{"b": float64(0.1)}, RunOutput: float64(0.1), Output: map[string]any{"a": float64(0.1)}},
		{Script: `a = b ?? 1.0`, Input: map[string]any{"b": nil}, RunOutput: float64(1.0), Output: map[string]any{"a": float64(1.0)}},
		{Script: `a = nil ?? 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b ?? 1`, Input: map[string]any{"b": []any{}}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b ?? 1`, Input: map[string]any{"b": map[string]any{}}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b[1] ?? 1`, Input: map[string]any{"b": []any{}}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = b[1][2] ?? 1`, Input: map[string]any{"b": []any{}}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = [] ?? 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `a = [2] ?? 1`, RunOutput: []any{int64(2)}, Output: map[string]any{"a": []any{int64(2)}}},
		{Script: `a = b ?? 1`, Input: map[string]any{"b": map[string]any{"test": int64(2)}}, RunOutput: map[string]any{"test": int64(2)}, Output: map[string]any{"a": map[string]any{"test": int64(2)}}},
		{Script: `a = b["test"] ?? 1`, Input: map[string]any{"b": map[string]any{"test": int64(2)}}, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}},
		{Script: `a = b["test"][1] ?? 1`, Input: map[string]any{"b": map[string]any{"test": 2}}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
		{Script: `b = "test"; a = b ?? "empty"`, RunOutput: "test", Output: map[string]any{"a": "test"}},
		{Script: `b = "test"; a = b[1:3] ?? "empty"`, RunOutput: "es", Output: map[string]any{"a": "es"}},
		{Script: `b = "test"; a = b[2:2] ?? "empty"`, RunOutput: "empty", Output: map[string]any{"a": "empty"}},
		{Script: `b = "0.0"; a = false ?? b ?? 1`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestIf(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `if 1++ {}`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `if false {} else if 1++ {}`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `if false {} else if true { 1++ }`, RunError: fmt.Errorf("invalid operation"), Name: ""},

		{Script: `if true {}`, Input: map[string]any{"a": nil}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `if true {}`, Input: map[string]any{"a": true}, RunOutput: nil, Output: map[string]any{"a": true}, Name: ""},
		{Script: `if true {}`, Input: map[string]any{"a": int64(1)}, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `if true {}`, Input: map[string]any{"a": float64(1.1)}, RunOutput: nil, Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `if true {}`, Input: map[string]any{"a": "a"}, RunOutput: nil, Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `if true {a = nil}`, Input: map[string]any{"a": int64(2)}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `if true {a = true}`, Input: map[string]any{"a": int64(2)}, RunOutput: true, Output: map[string]any{"a": true}, Name: ""},
		{Script: `if true {a = 1}`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `if true {a = 1.1}`, Input: map[string]any{"a": int64(2)}, RunOutput: float64(1.1), Output: map[string]any{"a": float64(1.1)}, Name: ""},
		{Script: `if true {a = "a"}`, Input: map[string]any{"a": int64(2)}, RunOutput: "a", Output: map[string]any{"a": "a"}, Name: ""},

		{Script: `if a == 1 {a = 1}`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `if a == 2 {a = 1}`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `if a == 1 {a = nil}`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `if a == 2 {a = nil}`, Input: map[string]any{"a": int64(2)}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},

		{Script: `if a == 1 {a = 1} else {a = 3}`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(3), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `if a == 2 {a = 1} else {a = 3}`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `if a == 1 {a = 1} else if a == 3 {a = 3}`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `if a == 1 {a = 1} else if a == 2 {a = 3}`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(3), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `if a == 1 {a = 1} else if a == 3 {a = 3} else {a = 4}`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(4), Output: map[string]any{"a": int64(4)}, Name: ""},

		{Script: `if a == 1 {a = 1} else {a = nil}`, Input: map[string]any{"a": int64(2)}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `if a == 2 {a = nil} else {a = 3}`, Input: map[string]any{"a": int64(2)}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `if a == 1 {a = nil} else if a == 3 {a = nil}`, Input: map[string]any{"a": int64(2)}, RunOutput: false, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `if a == 1 {a = 1} else if a == 2 {a = nil}`, Input: map[string]any{"a": int64(2)}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `if a == 1 {a = 1} else if a == 3 {a = 3} else {a = nil}`, Input: map[string]any{"a": int64(2)}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},

		{Script: `if a == 1 {a = 1} else if a == 3 {a = 3} else if a == 4 {a = 4} else {a = 5}`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(5), Output: map[string]any{"a": int64(5)}, Name: ""},
		{Script: `if a == 1 {a = 1} else if a == 3 {a = 3} else if a == 4 {a = 4} else {a = nil}`, Input: map[string]any{"a": int64(2)}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},
		{Script: `if a == 1 {a = 1} else if a == 3 {a = 3} else if a == 2 {a = 4} else {a = 5}`, Input: map[string]any{"a": int64(2)}, RunOutput: int64(4), Output: map[string]any{"a": int64(4)}, Name: ""},
		{Script: `if a == 1 {a = 1} else if a == 3 {a = 3} else if a == 2 {a = nil} else {a = 5}`, Input: map[string]any{"a": int64(2)}, RunOutput: nil, Output: map[string]any{"a": nil}, Name: ""},

		// check scope
		{Script: `a = 1; if a == 1 { b = 2 }; b`, RunError: envPkg.NewUndefinedSymbolErr("b"), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; if a == 2 { b = 3 } else { b = 4 }; b`, RunError: envPkg.NewUndefinedSymbolErr("b"), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; if a == 2 { b = 3 } else if a == 1 { b = 4 }; b`, RunError: envPkg.NewUndefinedSymbolErr("b"), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; if a == 2 { b = 4 } else if a == 5 { b = 6 } else if a == 1 { c = b }`, RunError: envPkg.NewUndefinedSymbolErr("b"), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; if a == 2 { b = 4 } else if a == 5 { b = 6 } else if a == 1 { b = 7 }; b`, RunError: envPkg.NewUndefinedSymbolErr("b"), Output: map[string]any{"a": int64(1)}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestSelect(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		// test parse errors
		{Script: `select {default: return 6; default: return 7}`, ParseError: fmt.Errorf("unexpected DEFAULT")},
		{Script: `a = make(chan int64, 1); a <- 1; select {case <-a: return 5; default: return 6; default: return 7}`, ParseError: fmt.Errorf("unexpected DEFAULT")},

		// test run errors
		{Script: `select {case a = <-b: return 1}`, RunError: envPkg.NewUndefinedSymbolErr("b"), RunOutput: nil},
		{Script: `select {case b = 1: return 1}`, RunError: fmt.Errorf("invalid operation"), RunOutput: nil},
		{Script: `select {case 1: return 1}`, RunError: fmt.Errorf("invalid operation"), RunOutput: nil},
		{Script: `select {case <-1: return 1}`, RunError: fmt.Errorf("invalid operation"), RunOutput: nil},
		{Script: `select {case <-a: return 1}`, RunError: envPkg.NewUndefinedSymbolErr("a"), RunOutput: nil},
		{Script: `select {case if true { }: return 1}`, RunError: fmt.Errorf("invalid operation"), RunOutput: nil},
		{Script: `a = make(chan int64, 1); a <- 1; select {case b.c = <-a: return 1}`, RunError: envPkg.NewUndefinedSymbolErr("b"), RunOutput: nil},
		{Script: `a = make(chan int64, 1); a <- 1; select {case v = <-a:}; v`, RunError: envPkg.NewUndefinedSymbolErr("v"), RunOutput: nil},

		// test 1 channel
		{Script: `a = make(chan int64, 1); a <- 1; select {case <-a: return 1}`, RunOutput: int64(1)},

		// test 2 channels
		{Script: `a = make(chan int64, 1); b = make(chan int64, 1); a <- 1; select {case <-a: return 1; case <-b: return 2}`, RunOutput: int64(1)},
		{Script: `a = make(chan int64, 1); b = make(chan int64, 1); b <- 1; select {case <-a: return 1; case <-b: return 2}`, RunOutput: int64(2)},

		// test default
		{Script: `select {default: return 1}`, RunOutput: int64(1)},
		{Script: `a = make(chan int64, 1); b = make(chan int64, 1); select {case <-a: return 1; case <-b: return 2; default: return 3}`, RunOutput: int64(3)},
		{Script: `a = make(chan int64, 1); b = make(chan int64, 1); a <- 1; select {case <-a: return 1; case <-b: return 2; default: return 3}`, RunOutput: int64(1)},
		{Script: `a = make(chan int64, 1); b = make(chan int64, 1); b <- 1; select {case <-a: return 1; case <-b: return 2; default: return 3}`, RunOutput: int64(2)},

		// test assignment
		{Script: `a = make(chan int64, 1); b = make(chan int64, 1); a <- 1; v = 0; select {case v = <-a:; case v = <-b:}; v`, RunOutput: int64(1), Output: map[string]any{"v": int64(1)}},
		{Script: `a = make(chan int64, 1); a <- 1; select {case v = <-a: return v}`, RunOutput: int64(1), Output: map[string]any{}},

		// test new lines
		{Script: `
		a = make(chan int64, 1)
		a <- 1
		select {
			case <-a:
				return 1
		}`, RunOutput: int64(1)},
		{Script: `
		select {
			case 1:
				return 1
		}`, RunError: fmt.Errorf("invalid operation"), RunErrorLine: 3, RunErrorColumn: 9, RunOutput: nil},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestSwitch(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		// test parse errors
		{Script: `switch {}`, ParseError: fmt.Errorf("unexpected $end"), Name: ""},
		{Script: `a = 1; switch a; {}`, ParseError: fmt.Errorf("unexpected ';'"), Name: ""},
		{Script: `a = 1; switch a = 2 {}`, ParseError: fmt.Errorf("unexpected '='"), Name: ""},
		{Script: `a = 1; switch a {default: return 6; default: return 7}`, ParseError: fmt.Errorf("unexpected DEFAULT"), Name: ""},
		{Script: `a = 1; switch a {case 1: return 5; default: return 6; default: return 7}`, ParseError: fmt.Errorf("unexpected DEFAULT"), Name: ""},

		// test run errors
		{Script: `a = 1; switch 1++ {}`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `a = 1; switch a {case 1++: return 2}`, RunError: fmt.Errorf("invalid operation"), Name: ""},

		// test no or empty cases
		{Script: `a = 1; switch a {}`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; switch a {case: return 2}`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; switch a {case: return 2; case: return 3}`, RunOutput: int64(1), Output: map[string]any{"a": int64(1)}, Name: ""},

		// test 1 case
		{Script: `a = 1; switch a {case 1: return 5}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1: return 5}`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; switch a {case 1,2: return 5}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1,2: return 5}`, RunOutput: int64(5), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 3; switch a {case 1,2: return 5}`, RunOutput: int64(3), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 1; switch a {case 1,2,3: return 5}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1,2,3: return 5}`, RunOutput: int64(5), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 3; switch a {case 1,2,3: return 5}`, RunOutput: int64(5), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 4; switch a {case 1,2,3: return 5}`, RunOutput: int64(4), Output: map[string]any{"a": int64(4)}, Name: ""},
		{Script: `a = func() { return 1 }; switch a() {case 1: return 5}`, RunOutput: int64(5), Name: ""},
		{Script: `a = func() { return 2 }; switch a() {case 1: return 5}`, RunOutput: int64(2), Name: ""},
		{Script: `a = func() { return 5 }; b = 1; switch b {case 1: return a() }`, RunOutput: int64(5), Output: map[string]any{"b": int64(1)}, Name: ""},
		{Script: `a = func() { return 6 }; b = 2; switch b {case 1: return a() }`, RunOutput: int64(2), Output: map[string]any{"b": int64(2)}, Name: ""},

		// test 2 cases
		{Script: `a = 1; switch a {case 1: return 5; case 2: return 6}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1: return 5; case 2: return 6}`, RunOutput: int64(6), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 3; switch a {case 1: return 5; case 2: return 6}`, RunOutput: int64(3), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 1; switch a {case 1: return 5; case 2,3: return 6}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1: return 5; case 2,3: return 6}`, RunOutput: int64(6), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 3; switch a {case 1: return 5; case 2,3: return 6}`, RunOutput: int64(6), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 4; switch a {case 1: return 5; case 2,3: return 6}`, RunOutput: int64(4), Output: map[string]any{"a": int64(4)}, Name: ""},

		// test 3 cases
		{Script: `a = 1; switch a {case 1,2: return 5; case 3: return 6}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1,2: return 5; case 3: return 6}`, RunOutput: int64(5), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 3; switch a {case 1,2: return 5; case 3: return 6}`, RunOutput: int64(6), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 4; switch a {case 1,2: return 5; case 3: return 6}`, RunOutput: int64(4), Output: map[string]any{"a": int64(4)}, Name: ""},
		{Script: `a = 1; switch a {case 1,2: return 5; case 2,3: return 6}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1,2: return 5; case 2,3: return 6}`, RunOutput: int64(5), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 3; switch a {case 1,2: return 5; case 2,3: return 6}`, RunOutput: int64(6), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 4; switch a {case 1,2: return 5; case 2,3: return 6}`, RunOutput: int64(4), Output: map[string]any{"a": int64(4)}, Name: ""},

		// test default
		{Script: `a = 1; switch a {default: return 5}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}},
		{Script: `a = 1; switch a {case 1: return 5; case 2: return 6; default: return 7}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1: return 5; case 2: return 6; default: return 7}`, RunOutput: int64(6), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 3; switch a {case 1: return 5; case 2: return 6; default: return 7}`, RunOutput: int64(7), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 1; switch a {case 1: return 5; case 2,3: return 6; default: return 7}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1: return 5; case 2,3: return 6; default: return 7}`, RunOutput: int64(6), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 3; switch a {case 1: return 5; case 2,3: return 6; default: return 7}`, RunOutput: int64(6), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 4; switch a {case 1: return 5; case 2,3: return 6; default: return 7}`, RunOutput: int64(7), Output: map[string]any{"a": int64(4)}, Name: ""},
		{Script: `a = 1; switch a {case 1,2: return 5; case 3: return 6; default: return 7}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1,2: return 5; case 3: return 6; default: return 7}`, RunOutput: int64(5), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 3; switch a {case 1,2: return 5; case 3: return 6; default: return 7}`, RunOutput: int64(6), Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 4; switch a {case 1,2: return 5; case 3: return 6; default: return 7}`, RunOutput: int64(7), Output: map[string]any{"a": int64(4)}, Name: ""},

		// test scope
		{Script: `a = 1; switch a {case 1: a = 5}`, RunOutput: int64(5), Output: map[string]any{"a": int64(5)}, Name: ""},
		{Script: `a = 2; switch a {case 1: a = 5}`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; b = 5; switch a {case 1: b = 6}`, RunOutput: int64(6), Output: map[string]any{"a": int64(1), "b": int64(6)}, Name: ""},
		{Script: `a = 2; b = 5; switch a {case 1: b = 6}`, RunOutput: int64(2), Output: map[string]any{"a": int64(2), "b": int64(5)}, Name: ""},
		{Script: `a = 1; b = 5; switch a {case 1: b = 6; default: b = 7}`, RunOutput: int64(6), Output: map[string]any{"a": int64(1), "b": int64(6)}, Name: ""},
		{Script: `a = 2; b = 5; switch a {case 1: b = 6; default: b = 7}`, RunOutput: int64(7), Output: map[string]any{"a": int64(2), "b": int64(7)}, Name: ""},

		// test scope without define b
		{Script: `a = 1; switch a {case 1: b = 5}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1: b = 5}`, RunOutput: int64(2), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; switch a {case 1: b = 5}; b`, RunError: envPkg.NewUndefinedSymbolErr("b"), RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1: b = 5}; b`, RunError: envPkg.NewUndefinedSymbolErr("b"), RunOutput: nil, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; switch a {case 1: b = 5; default: b = 6}`, RunOutput: int64(5), Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1: b = 5; default: b = 6}`, RunOutput: int64(6), Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; switch a {case 1: b = 5; default: b = 6}; b`, RunError: envPkg.NewUndefinedSymbolErr("b"), RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 2; switch a {case 1: b = 5; default: b = 6}; b`, RunError: envPkg.NewUndefinedSymbolErr("b"), RunOutput: nil, Output: map[string]any{"a": int64(2)}, Name: ""},

		// test new lines
		{Script: `
		a = 1;
		switch a {
			case 1:
				return 1
		}`, RunOutput: int64(1)},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestForLoop(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `break`, RunError: fmt.Errorf("unexpected break statement"), Name: ""},
		{Script: `continue`, RunError: fmt.Errorf("unexpected continue statement"), Name: ""},
		{Script: `for 1++ { }`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `for { 1++ }`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `for a in 1++ { }`, ParseError: fmt.Errorf("unexpected '{'"), Name: ""},

		{Script: `for { break }`, RunOutput: nil, Name: ""},
		{Script: `for {a = 1; if a == 1 { break } }`, RunOutput: nil, Name: ""},
		{Script: `a = 1; for { if a == 1 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for { if a == 1 { break }; a++ }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},

		{Script: `a = 1; for { a++; if a == 2 { continue } else { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 1; for { a++; if a == 2 { continue }; if a == 3 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(3)}, Name: ""},

		{Script: `for a in [1] { if a == 1 { break } }`, RunOutput: nil, Name: ""},
		{Script: `for a in [1, 2] { if a == 2 { break } }`, RunOutput: nil, Name: ""},
		{Script: `for a in [1, 2, 3] { if a == 3 { break } }`, RunOutput: nil, Name: ""},

		{Script: `b = [1,2,3,4,5]; for a in b[1:3] { if a == 1 { break } }`, RunOutput: nil, Name: ""},
		{Script: `b = [[1,2,3]]; for a in b[0] { if a == 1 { break } }`, RunOutput: nil, Name: ""},
		{Script: `for a in true ? [1,2,3] : [4,5,6] { if a == 1 { break } }`, RunOutput: nil, Name: ""},
		{Script: `for a in (true ? [1,2,3] : [4,5,6]) { if a == 1 { break } }`, RunOutput: nil, Name: ""},

		{Script: `a = [1]; for b in a { if b == 1 { break } }`, RunOutput: nil, Output: map[string]any{"a": []any{int64(1)}}, Name: ""},
		{Script: `a = [1, 2]; for b in a { if b == 2 { break } }`, RunOutput: nil, Output: map[string]any{"a": []any{int64(1), int64(2)}}, Name: ""},
		{Script: `a = [1, 2, 3]; for b in a { if b == 3 { break } }`, RunOutput: nil, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}}, Name: ""},

		{Script: `a = [1]; b = 0; for c in a { b = c }`, RunOutput: nil, Output: map[string]any{"a": []any{int64(1)}, "b": int64(1)}, Name: ""},
		{Script: `a = [1, 2]; b = 0; for c in a { b = c }`, RunOutput: nil, Output: map[string]any{"a": []any{int64(1), int64(2)}, "b": int64(2)}, Name: ""},
		{Script: `a = [1, 2, 3]; b = 0; for c in a { b = c }`, RunOutput: nil, Output: map[string]any{"a": []any{int64(1), int64(2), int64(3)}, "b": int64(3)}, Name: ""},

		{Script: `a = 1; for a < 2 { a++ }`, RunOutput: nil, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; for a < 3 { a++ }`, RunOutput: nil, Output: map[string]any{"a": int64(3)}, Name: ""},

		{Script: `a = 1; for nil { a++; if a > 2 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for nil { a++; if a > 3 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for true { a++; if a > 2 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `a = 1; for true { a++; if a > 3 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(4)}, Name: ""},

		{Script: `func x() { return [1] }; for b in x() { if b == 1 { break } }`, RunOutput: nil, Name: ""},
		{Script: `func x() { return [1, 2] }; for b in x() { if b == 2 { break } }`, RunOutput: nil, Name: ""},
		{Script: `func x() { return [1, 2, 3] }; for b in x() { if b == 3 { break } }`, RunOutput: nil, Name: ""},

		{Script: `func x() { a = 1; for { if a == 1 { return } } }; x()`, RunOutput: nil, Name: ""},
		{Script: `func x() { a = 1; for { if a == 1 { return nil } } }; x()`, RunOutput: nil, Name: ""},
		{Script: `func x() { a = 1; for { if a == 1 { return true } } }; x()`, RunOutput: true, Name: ""},
		{Script: `func x() { a = 1; for { if a == 1 { return 1 } } }; x()`, RunOutput: int64(1), Name: ""},
		{Script: `func x() { a = 1; for { if a == 1 { return 1.1 } } }; x()`, RunOutput: float64(1.1), Name: ""},
		{Script: `func x() { a = 1; for { if a == 1 { return "a" } } }; x()`, RunOutput: "a", Name: ""},

		{Script: `func x() { for a in [1, 2, 3] { if a == 3 { return } } }; x()`, RunOutput: nil, Name: ""},
		{Script: `func x() { for a in [1, 2, 3] { if a == 1 { continue } } }; x()`, RunOutput: nil, Name: ""},
		{Script: `func x() { for a in [1, 2, 3] { if a == 1 { continue };  if a == 3 { return } } }; x()`, RunOutput: nil, Name: ""},

		{Script: `func x() { return [1, 2] }; a = 1; for i in x() { a++ }`, RunOutput: nil, Output: map[string]any{"a": int64(3)}, Name: ""},
		{Script: `func x() { return [1, 2, 3] }; a = 1; for i in x() { a++ }`, RunOutput: nil, Output: map[string]any{"a": int64(4)}, Name: ""},

		{Script: `for a = 1; nil; nil { return }`, Name: ""},
		// TOFIX:
		// {Script: `for a, b = 1; nil; nil { return }`, Name: ""},
		// {Script: `for a, b = 1, 2; nil; nil { return }`, Name: ""},

		{Script: `for var a = 1; nil; nil { return }`, Name: ""},
		{Script: `for var a = 1, 2; nil; nil { return }`, ParseError: fmt.Errorf("unexpected ','"), Name: ""},
		{Script: `for var a, b = 1; nil; nil { return }`, Name: ""},
		{Script: `for var a, b = 1, 2; nil; nil { return }`, Name: ""},

		{Script: `for a.b = 1; nil; nil { return }`, RunError: envPkg.NewUndefinedSymbolErr("a"), Name: ""},

		{Script: `for a = 1; nil; nil { if a == 1 { break } }`, RunOutput: nil, Name: ""},
		{Script: `for a = 1; nil; nil { if a == 2 { break }; a++ }`, RunOutput: nil, Name: ""},
		{Script: `for a = 1; nil; nil { a++; if a == 3 { break } }`, RunOutput: nil, Name: ""},

		{Script: `for a = 1; a < 1; nil { }`, RunOutput: nil, Name: ""},
		{Script: `for a = 1; a > 1; nil { }`, RunOutput: nil, Name: ""},
		{Script: `for a = 1; a == 1; nil { break }`, RunOutput: nil, Name: ""},

		{Script: `for a = 1; a == 1; a++ { }`, RunOutput: nil, Name: ""},
		{Script: `for a = 1; a < 2; a++ { }`, RunOutput: nil, Name: ""},
		{Script: `for a = 1; a < 3; a++ { }`, RunOutput: nil, Name: ""},

		{Script: `a = 1; for b = 1; a < 1; a++ { }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 2; a++ { }`, RunOutput: nil, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 3; a++ { }`, RunOutput: nil, Output: map[string]any{"a": int64(3)}, Name: ""},

		{Script: `a = 1; for b = 1; a < 1; a++ {  if a == 1 { continue } }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 2; a++ {  if a == 1 { continue } }`, RunOutput: nil, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 3; a++ {  if a == 1 { continue } }`, RunOutput: nil, Output: map[string]any{"a": int64(3)}, Name: ""},

		{Script: `a = 1; for b = 1; a < 1; a++ {  if a == 1 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 2; a++ {  if a == 1 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 3; a++ {  if a == 1 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 1; a++ {  if a == 2 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 2; a++ {  if a == 2 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 3; a++ {  if a == 2 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 1; a++ {  if a == 3 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(1)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 2; a++ {  if a == 3 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(2)}, Name: ""},
		{Script: `a = 1; for b = 1; a < 3; a++ {  if a == 3 { break } }`, RunOutput: nil, Output: map[string]any{"a": int64(3)}, Name: ""},

		{Script: `a = ["123", "456", "789"]; b = ""; for i = 0; i < len(a); i++ { b += a[i][len(a[i]) - 2:]; b += a[i][:len(a[i]) - 2] }`, RunOutput: nil, Output: map[string]any{"a": []any{"123", "456", "789"}, "b": "231564897"}, Name: ""},
		{Script: `a = [[["123"], ["456"]], [["789"]]]; b = ""; for i = 0; i < len(a); i++ { for j = 0; j < len(a[i]); j++ {  for k = 0; k < len(a[i][j]); k++ { for l = 0; l < len(a[i][j][k]); l++ { b += a[i][j][k][l] + "-" } } } }`,
			RunOutput: nil, Output: map[string]any{"a": []any{[]any{[]any{"123"}, []any{"456"}}, []any{[]any{"789"}}}, "b": "1-2-3-4-5-6-7-8-9-"}, Name: ""},

		{Script: `func x() { for a = 1; a < 3; a++ { if a == 1 { return a } } }; x()`, RunOutput: int64(1), Name: ""},
		{Script: `func x() { for a = 1; a < 3; a++ { if a == 2 { return a } } }; x()`, RunOutput: int64(2), Name: ""},
		{Script: `func x() { for a = 1; a < 3; a++ { if a == 3 { return a } } }; x()`, RunOutput: nil, Name: ""},
		{Script: `func x() { for a = 1; a < 3; a++ { if a == 4 { return a } } }; x()`, RunOutput: nil, Name: ""},

		{Script: `func x() { a = 1; for b = 1; a < 3; a++ { if a == 1 { continue } }; return a }; x()`, RunOutput: int64(3), Name: ""},
		{Script: `func x() { a = 1; for b = 1; a < 3; a++ { if a == 2 { continue } }; return a }; x()`, RunOutput: int64(3), Name: ""},
		{Script: `func x() { a = 1; for b = 1; a < 3; a++ { if a == 3 { continue } }; return a }; x()`, RunOutput: int64(3), Name: ""},
		{Script: `func x() { a = 1; for b = 1; a < 3; a++ { if a == 4 { continue } }; return a }; x()`, RunOutput: int64(3), Name: ""},

		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{reflect.Value{}}}, RunOutput: nil, Output: map[string]any{"a": []any{reflect.Value{}}, "b": reflect.Value{}}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{nil}}, RunOutput: nil, Output: map[string]any{"a": []any{nil}, "b": nil}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{true}}, RunOutput: nil, Output: map[string]any{"a": []any{true}, "b": true}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{int32(1)}}, RunOutput: nil, Output: map[string]any{"a": []any{int32(1)}, "b": int32(1)}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{int64(1)}}, RunOutput: nil, Output: map[string]any{"a": []any{int64(1)}, "b": int64(1)}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{float32(1.1)}}, RunOutput: nil, Output: map[string]any{"a": []any{float32(1.1)}, "b": float32(1.1)}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{float64(1.1)}}, RunOutput: nil, Output: map[string]any{"a": []any{float64(1.1)}, "b": float64(1.1)}, Name: ""},

		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{any(reflect.Value{})}}, RunOutput: nil, Output: map[string]any{"a": []any{any(reflect.Value{})}, "b": any(reflect.Value{})}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{any(nil)}}, RunOutput: nil, Output: map[string]any{"a": []any{any(nil)}, "b": any(nil)}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{any(true)}}, RunOutput: nil, Output: map[string]any{"a": []any{any(true)}, "b": any(true)}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{any(int32(1))}}, RunOutput: nil, Output: map[string]any{"a": []any{any(int32(1))}, "b": any(int32(1))}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{any(int64(1))}}, RunOutput: nil, Output: map[string]any{"a": []any{any(int64(1))}, "b": any(int64(1))}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{any(float32(1.1))}}, RunOutput: nil, Output: map[string]any{"a": []any{any(float32(1.1))}, "b": any(float32(1.1))}, Name: ""},
		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": []any{any(float64(1.1))}}, RunOutput: nil, Output: map[string]any{"a": []any{any(float64(1.1))}, "b": any(float64(1.1))}, Name: ""},

		{Script: `b = 0; for i in a { b = i }`, Input: map[string]any{"a": any([]any{nil})}, RunOutput: nil, Output: map[string]any{"a": any([]any{nil}), "b": nil}, Name: ""},

		{Script: `for i in nil { }`, ParseError: fmt.Errorf("unexpected '{'"), Name: ""},
		{Script: `for i in true { }`, ParseError: fmt.Errorf("unexpected '{'"), Name: ""},
		{Script: `for i in a { }`, Input: map[string]any{"a": reflect.Value{}}, RunError: fmt.Errorf("for cannot loop over type struct"), Output: map[string]any{"a": reflect.Value{}}, Name: ""},
		{Script: `for i in a { }`, Input: map[string]any{"a": any(nil)}, RunError: fmt.Errorf("for cannot loop over type interface"), Output: map[string]any{"a": any(nil)}, Name: ""},
		{Script: `for i in a { }`, Input: map[string]any{"a": any(true)}, RunError: fmt.Errorf("for cannot loop over type bool"), Output: map[string]any{"a": any(true)}, Name: ""},
		{Script: `for i in [1, 2, 3] { b++ }`, RunError: envPkg.NewUndefinedSymbolErr("b"), Name: ""},
		{Script: `for a = 1; a < 3; a++ { b++ }`, RunError: envPkg.NewUndefinedSymbolErr("b"), Name: ""},
		{Script: `for a = b; a < 3; a++ { }`, RunError: envPkg.NewUndefinedSymbolErr("b"), Name: ""},
		{Script: `for a = 1; b < 3; a++ { }`, RunError: envPkg.NewUndefinedSymbolErr("b"), Name: ""},
		{Script: `for a = 1; a < 3; b++ { }`, RunError: envPkg.NewUndefinedSymbolErr("b"), Name: ""},

		{Script: `a = 1; b = [{"c": "c"}]; for i in b { a = i }`, RunOutput: nil, Output: map[string]any{"a": map[any]any{"c": "c"}, "b": []any{map[any]any{"c": "c"}}}, Name: ""},
		{Script: `a = 1; b = {"x": [{"y": "y"}]};  for i in b.x { a = i }`, RunOutput: nil, Output: map[string]any{"a": map[any]any{"y": "y"}, "b": map[any]any{"x": []any{map[any]any{"y": "y"}}}}, Name: ""},

		{Script: `a = {}; b = 1; for i in a { b = i }; b`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{}, "b": int64(1)}, Name: ""},
		{Script: `a = {"x": 2}; b = 1; for i in a { b = i }; b`, RunOutput: "x", Output: map[string]any{"a": map[any]any{"x": int64(2)}, "b": "x"}, Name: ""},
		{Script: `a = {"x": 2, "y": 3}; b = 0; for i in a { b++ }; b`, RunOutput: int64(2), Output: map[string]any{"a": map[any]any{"x": int64(2), "y": int64(3)}, "b": int64(2)}, Name: ""},
		{Script: `a = {"x": 2, "y": 3}; for i in a { b++ }`, RunError: envPkg.NewUndefinedSymbolErr("b"), Output: map[string]any{"a": map[any]any{"x": int64(2), "y": int64(3)}}, Name: ""},

		{Script: `a = {"x": 2, "y": 3}; b = 0; for i in a { if i ==  "x" { continue }; b = i }; b`, RunOutput: "y", Output: map[string]any{"a": map[any]any{"x": int64(2), "y": int64(3)}, "b": "y"}, Name: ""},
		{Script: `a = {"x": 2, "y": 3}; b = 0; for i in a { if i ==  "y" { continue }; b = i }; b`, RunOutput: "x", Output: map[string]any{"a": map[any]any{"x": int64(2), "y": int64(3)}, "b": "x"}, Name: ""},
		{Script: `a = {"x": 2, "y": 3}; for i in a { if i ==  "x" { return 1 } }`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"x": int64(2), "y": int64(3)}}, Name: ""},
		{Script: `a = {"x": 2, "y": 3}; for i in a { if i ==  "y" { return 2 } }`, RunOutput: int64(2), Output: map[string]any{"a": map[any]any{"x": int64(2), "y": int64(3)}}, Name: ""},
		{Script: `a = {"x": 2, "y": 3}; b = 0; for i in a { if i ==  "x" { break }; b++ }; if b > 1 { return false } else { return true }`, RunOutput: true, Output: map[string]any{"a": map[any]any{"x": int64(2), "y": int64(3)}}, Name: ""},
		{Script: `a = {"x": 2, "y": 3}; b = 0; for i in a { if i ==  "y" { break }; b++ }; if b > 1 { return false } else { return true }`, RunOutput: true, Output: map[string]any{"a": map[any]any{"x": int64(2), "y": int64(3)}}, Name: ""},
		{Script: `a = {"x": 2, "y": 3}; b = 1; for i in a { if (i ==  "x" || i ==  "y") { break }; b++ }; b`, RunOutput: int64(1), Output: map[string]any{"a": map[any]any{"x": int64(2), "y": int64(3)}, "b": int64(1)}, Name: ""},

		{Script: `a = ["123", "456", "789"]; b = ""; for v in a { b += v[len(v) - 2:]; b += v[:len(v) - 2] }`, RunOutput: nil, Output: map[string]any{"a": []any{"123", "456", "789"}, "b": "231564897"}, Name: ""},
		{Script: `a = [[["123"], ["456"]], [["789"]]]; b = ""; for x in a { for y in x  {  for z in y { for i = 0; i < len(z); i++ { b += z[i] + "-" } } } }`,
			RunOutput: nil, Output: map[string]any{"a": []any{[]any{[]any{"123"}, []any{"456"}}, []any{[]any{"789"}}}, "b": "1-2-3-4-5-6-7-8-9-"}, Name: ""},

		{Script: `a = {"x": 2}; b = 0; for k, v in a { b = k }; b`, RunOutput: "x", Output: map[string]any{"a": map[any]any{"x": int64(2)}, "b": "x"}, Name: ""},
		{Script: `a = {"x": 2}; b = 0; for k, v in a { b = v }; b`, RunOutput: int64(2), Output: map[string]any{"a": map[any]any{"x": int64(2)}, "b": int64(2)}, Name: ""},

		{Script: `a = make(chan int64, 1); a <- 1; v = 0; for val in a { v = val; break; }; v`, RunOutput: int64(1), Output: map[string]any{"v": int64(1)}, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}

func TestItemInList(t *testing.T) {
	_ = os.Setenv("ANKO_DEBUG", "1")
	tests := []Test{
		{Script: `"a" in ["a"]`, RunOutput: true, Name: ""},
		{Script: `"a" in ["b"]`, RunOutput: false, Name: ""},
		{Script: `"a" in ["c", "b", "a"]`, RunOutput: true, Name: ""},
		{Script: `"a" in ["a", "b", 1]`, RunOutput: true, Name: ""},
		{Script: `"a" in l`, Input: map[string]any{"l": []any{"a"}}, RunOutput: true, Name: ""},
		{Script: `"a" in l`, Input: map[string]any{"l": []any{"b"}}, RunOutput: false, Name: ""},
		{Script: `"a" in l`, Input: map[string]any{"l": []any{"c", "b", "a"}}, RunOutput: true, Name: ""},
		{Script: `"a" in l`, Input: map[string]any{"l": []any{"a", "b", 1}}, RunOutput: true, Name: ""},

		{Script: `1 in [1]`, RunOutput: true, Name: ""},
		{Script: `1 in [2]`, RunOutput: false, Name: ""},
		{Script: `1 in [3, 2, 1]`, RunOutput: true, Name: ""},
		{Script: `1 in ["1"]`, RunOutput: true, Name: ""},
		{Script: `1 in l`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `"1" in l`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `1 in l`, Input: map[string]any{"l": []any{2}}, RunOutput: false, Name: ""},
		{Script: `1 in l`, Input: map[string]any{"l": []any{3, 2, 1}}, RunOutput: true, Name: ""},
		{Script: `1 in l`, Input: map[string]any{"l": []any{"1"}}, RunOutput: true, Name: ""},

		{Script: `0.9 in [1]`, Input: map[string]any{"l": []any{1}}, RunOutput: false, Name: ""},
		{Script: `1.0 in [1]`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `1.1 in [1]`, Input: map[string]any{"l": []any{1}}, RunOutput: false, Name: ""},
		{Script: `1 in [0.9]`, Input: map[string]any{"l": []any{0.9}}, RunOutput: false, Name: ""},
		{Script: `1 in [1.0]`, Input: map[string]any{"l": []any{1.0}}, RunOutput: true, Name: ""},
		{Script: `1 in [1.1]`, Input: map[string]any{"l": []any{1.1}}, RunOutput: false, Name: ""},
		{Script: `0.9 in [1]`, Input: map[string]any{"l": []any{1}}, RunOutput: false, Name: ""},
		{Script: `1.0 in [1]`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `1.1 in [1]`, Input: map[string]any{"l": []any{1}}, RunOutput: false, Name: ""},
		{Script: `1 in [0.9]`, Input: map[string]any{"l": []any{0.9}}, RunOutput: false, Name: ""},
		{Script: `1 in [1.0]`, Input: map[string]any{"l": []any{1.0}}, RunOutput: true, Name: ""},
		{Script: `1 in [1.1]`, Input: map[string]any{"l": []any{1.1}}, RunOutput: false, Name: ""},

		{Script: `true in ["true"]`, RunOutput: true, Name: ""},
		{Script: `true in [true]`, RunOutput: true, Name: ""},
		{Script: `true in [true, false]`, RunOutput: true, Name: ""},
		{Script: `false in [false, true]`, RunOutput: true, Name: ""},
		{Script: `true in l`, Input: map[string]any{"l": []any{"true"}}, RunOutput: true, Name: ""},
		{Script: `true in l`, Input: map[string]any{"l": []any{true}}, RunOutput: true, Name: ""},
		{Script: `true in l`, Input: map[string]any{"l": []any{true, false}}, RunOutput: true, Name: ""},
		{Script: `false in l`, Input: map[string]any{"l": []any{false, true}}, RunOutput: true, Name: ""},

		{Script: `"a" in ["b", "a", "c"][1:]`, RunOutput: true, Name: ""},
		{Script: `"a" in ["b", "a", "c"][:1]`, RunOutput: false, Name: ""},
		{Script: `"a" in ["b", "a", "c"][1:2]`, RunOutput: true, Name: ""},
		{Script: `l = ["b", "a", "c"];"a" in l[1:]`, RunOutput: true, Name: ""},
		{Script: `l = ["b", "a", "c"];"a" in l[:1]`, RunOutput: false, Name: ""},
		{Script: `l = ["b", "a", "c"];"a" in l[1:2]`, RunOutput: true, Name: ""},
		{Script: `"a" in l[1:]`, Input: map[string]any{"l": []any{"b", "a", "c"}}, RunOutput: true, Name: ""},
		{Script: `"a" in l[:1]`, Input: map[string]any{"l": []any{"b", "a", "c"}}, RunOutput: false, Name: ""},
		{Script: `"a" in l[1:2]`, Input: map[string]any{"l": []any{"b", "a", "c"}}, RunOutput: true, Name: ""},

		// for i in list && item in list
		{Script: `list_of_list = [["a"]];for l in list_of_list { return "a" in l }`, RunOutput: true, Name: ""},
		{Script: `for l in list_of_list { return "a" in l }`, Input: map[string]any{"list_of_list": []any{[]any{"a"}}}, RunOutput: true, Name: ""},

		// not slice or array
		// todo: support `"a" in "aaa"` ?
		{Script: `"a" in "aaa"`, RunError: fmt.Errorf("second argument must be slice or array; but have string"), Name: ""},
		{Script: `1 in 12345`, RunError: fmt.Errorf("type int64 does not support slice operation"), Name: ""},

		// a in item in list
		{Script: `"a" in 5 in [1, 2, 3]`, RunError: fmt.Errorf("type bool does not support slice operation"), Name: ""},

		// applying a in b in several part of expresstion/statement
		{Script: `switch 1 in [1] {case true: return true;default: return false}`, RunOutput: true, Name: ""},
		{Script: `switch 1 in [2,3] {case true: return true;default: return false}`, RunOutput: false, Name: ""},
		{Script: `switch true {case 1 in [1]: return true;default: return false}`, RunOutput: true, Name: ""},
		{Script: `switch false {case 1 in [1]: return true;default: return false}`, RunOutput: false, Name: ""},
		{Script: `if 1 in [1] {return true} else {return false}`, RunOutput: true, Name: ""},
		{Script: `if 1 in [2,3] {return true} else {return false}`, RunOutput: false, Name: ""},
		{Script: `for i in [1,2,3] { i++ }`, Name: ""},
		{Script: `a=1; a=a in [1]`, RunOutput: true, Name: ""},
		{Script: `a=1; a=a in [2,3]`, RunOutput: false, Name: ""},
		{Script: `1 in [1] && true`, RunOutput: true, Name: ""},
		{Script: `1 in [1] && false`, RunOutput: false, Name: ""},
		{Script: `1 in [1] || true`, RunOutput: true, Name: ""},
		{Script: `1 in [1] || false`, RunOutput: true, Name: ""},
		{Script: `1 in [2,3] && true`, RunOutput: false, Name: ""},
		{Script: `1 in [2,3] && false`, RunOutput: false, Name: ""},
		{Script: `1 in [2,3] || true`, RunOutput: true, Name: ""},
		{Script: `1 in [2,3] || false`, RunOutput: false, Name: ""},
		{Script: `1++ in [1, 2, 3]`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `3++ in [1, 2, 3]`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `1 in 1++`, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `a=1;a++ in [1, 2, 3]`, RunOutput: true, Name: ""},
		{Script: `a=3;a++ in [1, 2, 3]`, RunOutput: false, Name: ""},
		{Script: `switch 1 in l {case true: return true;default: return false}`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `switch 1 in l {case true: return true;default: return false}`, Input: map[string]any{"l": []any{2, 3}}, RunOutput: false, Name: ""},
		{Script: `switch true {case 1 in l: return true;default: return false}`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `switch false {case 1 in l: return true;default: return false}`, Input: map[string]any{"l": []any{1}}, RunOutput: false, Name: ""},
		{Script: `if 1 in l {return true} else {return false}`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `if 1 in l {return true} else {return false}`, Input: map[string]any{"l": []any{2, 3}}, RunOutput: false, Name: ""},
		{Script: `for i in l { i++ }`, Input: map[string]any{"l": []any{1, 2, 3}}, Name: ""},
		{Script: `a=1; a=a in l`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `a=1; a=a in l`, Input: map[string]any{"l": []any{2, 3}}, RunOutput: false, Name: ""},
		{Script: `1 in l && true`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `1 in l && false`, Input: map[string]any{"l": []any{1}}, RunOutput: false, Name: ""},
		{Script: `1 in l || true`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `1 in l || false`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `1 in l && true`, Input: map[string]any{"l": []any{2, 3}}, RunOutput: false, Name: ""},
		{Script: `1 in l && false`, Input: map[string]any{"l": []any{2, 3}}, RunOutput: false, Name: ""},
		{Script: `1 in l || true`, Input: map[string]any{"l": []any{2, 3}}, RunOutput: true, Name: ""},
		{Script: `1 in l || false`, Input: map[string]any{"l": []any{2, 3}}, RunOutput: false, Name: ""},
		{Script: `1++ in l`, Input: map[string]any{"l": []any{1, 2, 3}}, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `3++ in l`, Input: map[string]any{"l": []any{1, 2, 3}}, RunError: fmt.Errorf("invalid operation"), Name: ""},
		{Script: `a=1;a++ in l`, Input: map[string]any{"l": []any{1, 2, 3}}, RunOutput: true, Name: ""},
		{Script: `a=3;a++ in l`, Input: map[string]any{"l": []any{1, 2, 3}}, RunOutput: false, Name: ""},

		// multidimensional slice
		{Script: `1 in [1]`, RunOutput: true, Name: ""},
		{Script: `1 in [[1]]`, RunOutput: false, Name: ""},
		{Script: `1 in [[[1]]]`, RunOutput: false, Name: ""},
		{Script: `1 in [[1],[[1]],1]`, RunOutput: true, Name: ""},
		{Script: `[1] in [1]`, RunOutput: false, Name: ""},
		{Script: `[1] in [[1]]`, RunOutput: true, Name: ""},
		{Script: `[1] in [[[1]]]`, RunOutput: false, Name: ""},
		{Script: `[[1]] in [1]`, RunOutput: false, Name: ""},
		{Script: `[[1]] in [[1]]`, RunOutput: false, Name: ""},
		{Script: `[[1]] in [[[1]]]`, RunOutput: true, Name: ""},
		{Script: `1 in [1]`, Input: map[string]any{"l": []any{1}}, RunOutput: true, Name: ""},
		{Script: `1 in [[1]]`, Input: map[string]any{"l": []any{[]any{1}}}, RunOutput: false, Name: ""},
		{Script: `1 in [[[1]]]`, Input: map[string]any{"l": []any{[]any{[]any{1}}}}, RunOutput: false, Name: ""},
		{Script: `1 in [[1],[[1]],1]`, Input: map[string]any{"l": []any{[]any{1}, []any{[]any{1}}, 1}}, RunOutput: true, Name: ""},
		{Script: `[1] in [1]`, Input: map[string]any{"l": []any{1}}, RunOutput: false, Name: ""},
		{Script: `[1] in [[1]]`, Input: map[string]any{"l": []any{[]any{1}}}, RunOutput: true, Name: ""},
		{Script: `[1] in [[[1]]]`, Input: map[string]any{"l": []any{[]any{[]any{1}}}}, RunOutput: false, Name: ""},
		{Script: `[[1]] in [1]`, Input: map[string]any{"l": []any{1}}, RunOutput: false, Name: ""},
		{Script: `[[1]] in [[1]]`, Input: map[string]any{"l": []any{[]any{1}}}, RunOutput: false, Name: ""},
		{Script: `[[1]] in [[[1]]]`, Input: map[string]any{"l": []any{[]any{[]any{1}}}}, RunOutput: true, Name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) { runTest(t, tt, nil) })
	}
}
