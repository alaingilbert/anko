package env

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestSetError(t *testing.T) {
	envParent := NewEnv()
	envChild := envParent.newEnv()
	err := envChild.set("a", "a")
	if err == nil {
		t.Errorf("Set error - received: %v - expected: %v", err, fmt.Errorf("unknown symbol 'a'"))
	} else if err.Error() != "unknown symbol 'a'" {
		t.Errorf("Set error - received: %v - expected: %v", err, fmt.Errorf("unknown symbol 'a'"))
	}
}

func TestAddrError(t *testing.T) {
	envParent := NewEnv()
	envChild := envParent.newEnv()
	_, err := envChild.addr("a")
	if err == nil {
		t.Errorf("Addr error - received: %v - expected: %v", err, fmt.Errorf("undefined symbol 'a'"))
	} else if err.Error() != "undefined symbol 'a'" {
		t.Errorf("Addr error - received: %v - expected: %v", err, fmt.Errorf("undefined symbol 'a'"))
	}
}

func TestGetInvalid(t *testing.T) {
	env := NewEnv()
	env.values.Insert("a", reflect.Value{})
	value, err := env.Get("a")
	if err != nil {
		t.Errorf("Get error - received: %v - expected: %v", err, nil)
	}
	if value != nil {
		t.Errorf("Get value - received: %v - expected: %v", value, nil)
	}
}

func TestDefineAndGet(t *testing.T) {
	var err error
	var value any
	tests := []struct {
		testInfo       string
		varName        string
		varDefineValue any
		varGetValue    any
		varKind        reflect.Kind
		defineError    error
		getError       error
	}{
		{testInfo: "nil", varName: "a", varDefineValue: nil, varGetValue: nil, varKind: reflect.Interface},
		{testInfo: "bool", varName: "a", varDefineValue: true, varGetValue: true, varKind: reflect.Bool},
		{testInfo: " int16", varName: "a", varDefineValue: int16(1), varGetValue: int16(1), varKind: reflect.Int16},
		{testInfo: "int32", varName: "a", varDefineValue: int32(1), varGetValue: int32(1), varKind: reflect.Int32},
		{testInfo: "int64", varName: "a", varDefineValue: int64(1), varGetValue: int64(1), varKind: reflect.Int64},
		{testInfo: "uint32", varName: "a", varDefineValue: uint32(1), varGetValue: uint32(1), varKind: reflect.Uint32},
		{testInfo: "uint64", varName: "a", varDefineValue: uint64(1), varGetValue: uint64(1), varKind: reflect.Uint64},
		{testInfo: "float32", varName: "a", varDefineValue: float32(1), varGetValue: float32(1), varKind: reflect.Float32},
		{testInfo: "float64", varName: "a", varDefineValue: float64(1), varGetValue: float64(1), varKind: reflect.Float64},
		{testInfo: "string", varName: "a", varDefineValue: "a", varGetValue: "a", varKind: reflect.String},

		{testInfo: "string with dot", varName: "a.a", varDefineValue: "a", varGetValue: nil, varKind: reflect.Interface, defineError: fmt.Errorf("unknown symbol 'a.a'"), getError: fmt.Errorf("undefined symbol 'a.a'")},
		{testInfo: "string with quotes", varName: "a", varDefineValue: `"a"`, varGetValue: `"a"`, varKind: reflect.String},
	}

	// DefineAndGet
	for _, test := range tests {
		env := NewEnv()

		err = env.Define(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineAndGet %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineAndGet %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		value, err = env.Get(test.varName)
		if err != nil && test.getError != nil {
			if err.Error() != test.getError.Error() {
				t.Errorf("DefineAndGet %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
				continue
			}
		} else if !errors.Is(test.getError, err) {
			t.Errorf("DefineAndGet %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
			continue
		}
		if value != test.varGetValue {
			t.Errorf("DefineAndGet %v - value check - received %#v expected: %#v", test.testInfo, value, test.varGetValue)
		}
	}

	// DefineAndGet NewPackage
	for _, test := range tests {
		env := NewEnv()

		err = env.Define(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineAndGet NewPackage %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineAndGet NewPackage %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		value, err = env.Get(test.varName)
		if err != nil && test.getError != nil {
			if err.Error() != test.getError.Error() {
				t.Errorf("DefineAndGet NewPackage %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
				continue
			}
		} else if !errors.Is(test.getError, err) {
			t.Errorf("DefineAndGet NewPackage %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
			continue
		}
		if value != test.varGetValue {
			t.Errorf("DefineAndGet NewPackage %v - value check - received %#v expected: %#v", test.testInfo, value, test.varGetValue)
		}
	}

	// DefineAndGet NewEnv
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.NewEnv()

		err = envParent.Define(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineAndGet NewEnv %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineAndGet NewEnv %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		value, err = envChild.Get(test.varName)
		if err != nil && test.getError != nil {
			if err.Error() != test.getError.Error() {
				t.Errorf("DefineAndGet NewEnv %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
				continue
			}
		} else if !errors.Is(test.getError, err) {
			t.Errorf("DefineAndGet NewEnv %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
			continue
		}
		if value != test.varGetValue {
			t.Errorf("DefineAndGet NewEnv %v - value check - received %#v expected: %#v", test.testInfo, value, test.varGetValue)
		}
	}

	// DefineAndGet DefineGlobal
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.newEnv()

		err = envChild.defineGlobal(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineAndGet DefineGlobal %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineAndGet DefineGlobal %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		value, err = envParent.Get(test.varName)
		if err != nil && test.getError != nil {
			if err.Error() != test.getError.Error() {
				t.Errorf("DefineAndGet DefineGlobal %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
				continue
			}
		} else if !errors.Is(test.getError, err) {
			t.Errorf("DefineAndGet DefineGlobal %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
			continue
		}
		if value != test.varGetValue {
			t.Errorf("DefineAndGet DefineGlobal %v - value check - received %#v expected: %#v", test.testInfo, value, test.varGetValue)
		}
	}

}

func TestDefineModify(t *testing.T) {
	var err error
	var value any
	tests := []struct {
		testInfo       string
		varName        string
		varDefineValue any
		varGetValue    any
		varKind        reflect.Kind
		defineError    error
		getError       error
	}{
		{testInfo: "nil", varName: "a", varDefineValue: nil, varGetValue: nil, varKind: reflect.Interface},
		{testInfo: "bool", varName: "a", varDefineValue: true, varGetValue: true, varKind: reflect.Bool},
		{testInfo: "int64", varName: "a", varDefineValue: int64(1), varGetValue: int64(1), varKind: reflect.Int64},
		{testInfo: "float64", varName: "a", varDefineValue: float64(1), varGetValue: float64(1), varKind: reflect.Float64},
		{testInfo: "string", varName: "a", varDefineValue: "a", varGetValue: "a", varKind: reflect.String},
	}
	changeTests := []struct {
		varDefineValue any
		varGetValue    any
		varKind        reflect.Kind
		defineError    error
		getError       error
	}{
		{varDefineValue: nil, varGetValue: nil, varKind: reflect.Interface},
		{varDefineValue: "a", varGetValue: "a", varKind: reflect.String},
		{varDefineValue: int64(1), varGetValue: int64(1), varKind: reflect.Int64},
		{varDefineValue: float64(1), varGetValue: float64(1), varKind: reflect.Float64},
		{varDefineValue: true, varGetValue: true, varKind: reflect.Bool},
	}

	// DefineModify
	for _, test := range tests {
		env := NewEnv()

		err = env.Define(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineModify %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineModify %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		value, err = env.Get(test.varName)
		if err != nil && test.getError != nil {
			if err.Error() != test.getError.Error() {
				t.Errorf("DefineModify %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
				continue
			}
		} else if !errors.Is(test.getError, err) {
			t.Errorf("DefineModify %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
			continue
		}
		if value != test.varGetValue {
			t.Errorf("DefineModify %v - value check - received %#v expected: %#v", test.testInfo, value, test.varGetValue)
		}

		// DefineModify changeTest
		for _, changeTest := range changeTests {
			err = env.set(test.varName, changeTest.varDefineValue)
			if err != nil && changeTest.defineError != nil {
				if err.Error() != changeTest.defineError.Error() {
					t.Errorf("DefineModify changeTest %v - Set error - received: %v - expected: %v", test.testInfo, err, changeTest.defineError)
					continue
				}
			} else if err != changeTest.defineError {
				t.Errorf("DefineModify changeTest %v - Set error - received: %v - expected: %v", test.testInfo, err, changeTest.defineError)
				continue
			}

			value, err = env.Get(test.varName)
			if err != nil && changeTest.getError != nil {
				if err.Error() != changeTest.getError.Error() {
					t.Errorf("DefineModify changeTest  %v - Get error - received: %v - expected: %v", test.testInfo, err, changeTest.getError)
					continue
				}
			} else if err != changeTest.getError {
				t.Errorf("DefineModify changeTest  %v - Get error - received: %v - expected: %v", test.testInfo, err, changeTest.getError)
				continue
			}
			if value != changeTest.varGetValue {
				t.Errorf("DefineModify changeTest  %v - value check - received %#v expected: %#v", test.testInfo, value, changeTest.varGetValue)
			}
		}
	}

	// DefineModify envParent
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.NewEnv()

		err = envParent.Define(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineModify envParent %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineModify envParent %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		value, err = envChild.Get(test.varName)
		if err != nil && test.getError != nil {
			if err.Error() != test.getError.Error() {
				t.Errorf("DefineModify envParent  %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
				continue
			}
		} else if !errors.Is(test.getError, err) {
			t.Errorf("DefineModify envParent  %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
			continue
		}
		if value != test.varGetValue {
			t.Errorf("DefineModify envParent  %v - value check - received %#v expected: %#v", test.testInfo, value, test.varGetValue)
		}

		for _, changeTest := range changeTests {
			err = envParent.set(test.varName, changeTest.varDefineValue)
			if err != nil && changeTest.defineError != nil {
				if err.Error() != changeTest.defineError.Error() {
					t.Errorf("DefineModify envParent changeTest %v - Set error - received: %v - expected: %v", test.testInfo, err, changeTest.defineError)
					continue
				}
			} else if !errors.Is(changeTest.defineError, err) {
				t.Errorf("DefineModify envParent changeTest %v - Set error - received: %v - expected: %v", test.testInfo, err, changeTest.defineError)
				continue
			}

			value, err = envChild.Get(test.varName)
			if err != nil && changeTest.getError != nil {
				if err.Error() != changeTest.getError.Error() {
					t.Errorf("DefineModify envParent changeTest %v - Get error - received: %v - expected: %v", test.testInfo, err, changeTest.getError)
					continue
				}
			} else if !errors.Is(changeTest.getError, err) {
				t.Errorf("ChanDefineModify envParent changeTestgeTest %v - Get error - received: %v - expected: %v", test.testInfo, err, changeTest.getError)
				continue
			}
			if value != changeTest.varGetValue {
				t.Errorf("DefineModify envParent changeTest %v - value check - received %#v expected: %#v", test.testInfo, value, changeTest.varGetValue)
			}
		}
	}

	// DefineModify envChild
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.newEnv()

		err = envParent.Define(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineModify envChild %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineModify envChild %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		value, err = envChild.Get(test.varName)
		if err != nil && test.getError != nil {
			if err.Error() != test.getError.Error() {
				t.Errorf("DefineModify envChild  %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
				continue
			}
		} else if !errors.Is(test.getError, err) {
			t.Errorf("DefineModify envChild  %v - Get error - received: %v - expected: %v", test.testInfo, err, test.getError)
			continue
		}
		if value != test.varGetValue {
			t.Errorf("DefineModify envChild  %v - value check - received %#v expected: %#v", test.testInfo, value, test.varGetValue)
		}

		for _, changeTest := range changeTests {
			err = envChild.set(test.varName, changeTest.varDefineValue)
			if err != nil && changeTest.defineError != nil {
				if err.Error() != changeTest.defineError.Error() {
					t.Errorf("DefineModify envChild changeTest %v - Set error - received: %v - expected: %v", test.testInfo, err, changeTest.defineError)
					continue
				}
			} else if !errors.Is(changeTest.defineError, err) {
				t.Errorf("DefineModify envChild changeTest %v - Set error - received: %v - expected: %v", test.testInfo, err, changeTest.defineError)
				continue
			}

			value, err = envChild.Get(test.varName)
			if err != nil && changeTest.getError != nil {
				if err.Error() != changeTest.getError.Error() {
					t.Errorf("DefineModify envChild changeTest %v - Get error - received: %v - expected: %v", test.testInfo, err, changeTest.getError)
					continue
				}
			} else if !errors.Is(changeTest.getError, err) {
				t.Errorf("ChanDefineModify envChild changeTestgeTest %v - Get error - received: %v - expected: %v", test.testInfo, err, changeTest.getError)
				continue
			}
			if value != changeTest.varGetValue {
				t.Errorf("DefineModify envChild changeTest %v - value check - received %#v expected: %#v", test.testInfo, value, changeTest.varGetValue)
			}
		}
	}
}

func TestDefineType(t *testing.T) {
	var err error
	var valueType reflect.Type
	tests := []struct {
		testInfo       string
		varName        string
		varDefineValue any
		defineError    error
		typeError      error
	}{
		{testInfo: "nil", varName: "a", varDefineValue: nil},
		{testInfo: "bool", varName: "a", varDefineValue: true},
		{testInfo: "int16", varName: "a", varDefineValue: int16(1)},
		{testInfo: "int32", varName: "a", varDefineValue: int32(1)},
		{testInfo: "int64", varName: "a", varDefineValue: int64(1)},
		{testInfo: "uint32", varName: "a", varDefineValue: uint32(1)},
		{testInfo: "uint64", varName: "a", varDefineValue: uint64(1)},
		{testInfo: "float32", varName: "a", varDefineValue: float32(1)},
		{testInfo: "float64", varName: "a", varDefineValue: float64(1)},
		{testInfo: "string", varName: "a", varDefineValue: "a"},

		{testInfo: "string with dot", varName: "a.a", varDefineValue: nil, defineError: fmt.Errorf("unknown symbol 'a.a'"), typeError: fmt.Errorf("undefined type 'a.a'")},
	}

	// DefineType
	for _, test := range tests {
		env := NewEnv()

		err = env.DefineType(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineType %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineType %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		valueType, err = env.Type(test.varName)
		if err != nil && test.typeError != nil {
			if err.Error() != test.typeError.Error() {
				t.Errorf("DefineType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
				continue
			}
		} else if !errors.Is(test.typeError, err) {
			t.Errorf("DefineType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
			continue
		}
		if valueType == nil || test.varDefineValue == nil {
			if valueType != reflect.TypeOf(test.varDefineValue) {
				t.Errorf("DefineType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
			}
		} else if valueType.String() != reflect.TypeOf(test.varDefineValue).String() {
			t.Errorf("DefineType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
		}
	}

	// DefineType NewEnv
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.NewEnv()

		err = envParent.DefineType(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineType NewEnv %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineType NewEnv %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		valueType, err = envChild.Type(test.varName)
		if err != nil && test.typeError != nil {
			if err.Error() != test.typeError.Error() {
				t.Errorf("DefineType NewEnv %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
				continue
			}
		} else if !errors.Is(test.typeError, err) {
			t.Errorf("DefineType NewEnv %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
			continue
		}
		if valueType == nil || test.varDefineValue == nil {
			if valueType != reflect.TypeOf(test.varDefineValue) {
				t.Errorf("DefineType NewEnv %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
			}
		} else if valueType.String() != reflect.TypeOf(test.varDefineValue).String() {
			t.Errorf("DefineType NewEnv %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
		}
	}

	// DefineType NewPackage
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.NewEnv()

		err = envParent.DefineType(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineType NewPackage  %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineType NewPackage %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		valueType, err = envChild.Type(test.varName)
		if err != nil && test.typeError != nil {
			if err.Error() != test.typeError.Error() {
				t.Errorf("DefineType NewPackage %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
				continue
			}
		} else if !errors.Is(test.typeError, err) {
			t.Errorf("DefineType NewPackage %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
			continue
		}
		if valueType == nil || test.varDefineValue == nil {
			if valueType != reflect.TypeOf(test.varDefineValue) {
				t.Errorf("DefineType NewPackage %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
			}
		} else if valueType.String() != reflect.TypeOf(test.varDefineValue).String() {
			t.Errorf("DefineType NewPackage %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
		}
	}

	// DefineType NewModule
	for _, test := range tests {
		envParent := NewEnv()
		envChild, _ := envParent.NewModule("child")

		err = envParent.DefineType(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineType NewModule %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineType NewModule %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		valueType, err = envChild.Type(test.varName)
		if err != nil && test.typeError != nil {
			if err.Error() != test.typeError.Error() {
				t.Errorf("DefineType NewModule %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
				continue
			}
		} else if !errors.Is(test.typeError, err) {
			t.Errorf("DefineType NewModule %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
			continue
		}
		if valueType == nil || test.varDefineValue == nil {
			if valueType != reflect.TypeOf(test.varDefineValue) {
				t.Errorf("DefineType NewModule %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
			}
		} else if valueType.String() != reflect.TypeOf(test.varDefineValue).String() {
			t.Errorf("DefineType NewModule %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
		}
	}

	// DefineGlobalType
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.newEnv()

		err = envChild.defineGlobalType(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineGlobalType %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineGlobalType %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		valueType, err = envParent.Type(test.varName)
		if err != nil && test.typeError != nil {
			if err.Error() != test.typeError.Error() {
				t.Errorf("DefineGlobalType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
				continue
			}
		} else if !errors.Is(test.typeError, err) {
			t.Errorf("DefineGlobalType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
			continue
		}
		if valueType == nil || test.varDefineValue == nil {
			if valueType != reflect.TypeOf(test.varDefineValue) {
				t.Errorf("DefineGlobalType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
			}
		} else if valueType.String() != reflect.TypeOf(test.varDefineValue).String() {
			t.Errorf("DefineGlobalType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
		}

		valueType, err = envChild.Type(test.varName)
		if err != nil && test.typeError != nil {
			if err.Error() != test.typeError.Error() {
				t.Errorf("DefineGlobalType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
				continue
			}
		} else if !errors.Is(test.typeError, err) {
			t.Errorf("DefineGlobalType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
			continue
		}
		if valueType == nil || test.varDefineValue == nil {
			if valueType != reflect.TypeOf(test.varDefineValue) {
				t.Errorf("DefineGlobalType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
			}
		} else if valueType.String() != reflect.TypeOf(test.varDefineValue).String() {
			t.Errorf("DefineGlobalType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
		}
	}

	// DefineGlobalReflectType
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.newEnv()

		err = envChild.defineGlobalReflectType(test.varName, reflect.TypeOf(test.varDefineValue))
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("DefineGlobalReflectType %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("DefineGlobalReflectType %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		valueType, err = envParent.Type(test.varName)
		if err != nil && test.typeError != nil {
			if err.Error() != test.typeError.Error() {
				t.Errorf("DefineGlobalReflectType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
				continue
			}
		} else if !errors.Is(test.typeError, err) {
			t.Errorf("DefineGlobalReflectType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
			continue
		}
		if valueType == nil || test.varDefineValue == nil {
			if valueType != reflect.TypeOf(test.varDefineValue) {
				t.Errorf("DefineGlobalReflectType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
			}
		} else if valueType.String() != reflect.TypeOf(test.varDefineValue).String() {
			t.Errorf("DefineGlobalReflectType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
		}

		valueType, err = envChild.Type(test.varName)
		if err != nil && test.typeError != nil {
			if err.Error() != test.typeError.Error() {
				t.Errorf("DefineGlobalReflectType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
				continue
			}
		} else if !errors.Is(test.typeError, err) {
			t.Errorf("DefineGlobalReflectType %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
			continue
		}
		if valueType == nil || test.varDefineValue == nil {
			if valueType != reflect.TypeOf(test.varDefineValue) {
				t.Errorf("DefineGlobalReflectType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
			}
		} else if valueType.String() != reflect.TypeOf(test.varDefineValue).String() {
			t.Errorf("DefineGlobalReflectType %v - Type check - received: %v - expected: %v", test.testInfo, valueType, reflect.TypeOf(test.varDefineValue))
		}
	}
}

func TestDefineTypeFail(t *testing.T) {
	var err error
	tests := []struct {
		testInfo       string
		varName        string
		varDefineValue any
		defineError    error
		typeError      error
	}{
		{testInfo: "nil", varName: "a", varDefineValue: nil, typeError: fmt.Errorf("undefined type 'a'")},
		{testInfo: "bool", varName: "a", varDefineValue: true, typeError: fmt.Errorf("undefined type 'a'")},
		{testInfo: "int16", varName: "a", varDefineValue: int16(1), typeError: fmt.Errorf("undefined type 'a'")},
		{testInfo: "int32", varName: "a", varDefineValue: int32(1), typeError: fmt.Errorf("undefined type 'a'")},
		{testInfo: "int64", varName: "a", varDefineValue: int64(1), typeError: fmt.Errorf("undefined type 'a'")},
		{testInfo: "uint32", varName: "a", varDefineValue: uint32(1), typeError: fmt.Errorf("undefined type 'a'")},
		{testInfo: "uint64", varName: "a", varDefineValue: uint64(1), typeError: fmt.Errorf("undefined type 'a'")},
		{testInfo: "float32", varName: "a", varDefineValue: float32(1), typeError: fmt.Errorf("undefined type 'a'")},
		{testInfo: "float64", varName: "a", varDefineValue: float64(1), typeError: fmt.Errorf("undefined type 'a'")},
		{testInfo: "string", varName: "a", varDefineValue: "a", typeError: fmt.Errorf("undefined type 'a'")},
	}

	// DefineTypeFail
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.NewEnv()

		err = envChild.DefineType(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("TestDefineTypeFail %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("TestDefineTypeFail %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		_, err = envParent.Type(test.varName)
		if err != nil && test.typeError != nil {
			if err.Error() != test.typeError.Error() {
				t.Errorf("TestDefineTypeFail %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
				continue
			}
		} else if !errors.Is(test.typeError, err) {
			t.Errorf("TestDefineTypeFail %v - Type error - received: %v - expected: %v", test.testInfo, err, test.typeError)
		}
	}
}

func TestAddr(t *testing.T) {
	var err error
	tests := []struct {
		testInfo       string
		varName        string
		varDefineValue any
		defineError    error
		addrError      error
	}{
		{testInfo: "nil", varName: "a", varDefineValue: nil, addrError: nil},
		{testInfo: "string", varName: "a", varDefineValue: "a", addrError: fmt.Errorf("unaddressable")},
		{testInfo: "int64", varName: "a", varDefineValue: int64(1), addrError: fmt.Errorf("unaddressable")},
		{testInfo: "float64", varName: "a", varDefineValue: float64(1), addrError: fmt.Errorf("unaddressable")},
		{testInfo: "bool", varName: "a", varDefineValue: true, addrError: fmt.Errorf("unaddressable")},
	}

	// TestAddr
	for _, test := range tests {
		envParent := NewEnv()
		envChild := envParent.newEnv()

		err = envParent.Define(test.varName, test.varDefineValue)
		if err != nil && test.defineError != nil {
			if err.Error() != test.defineError.Error() {
				t.Errorf("TestAddr %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
				continue
			}
		} else if !errors.Is(test.defineError, err) {
			t.Errorf("TestAddr %v - Define error - received: %v - expected: %v", test.testInfo, err, test.defineError)
			continue
		}

		_, err = envChild.addr(test.varName)
		if err != nil && test.addrError != nil {
			if err.Error() != test.addrError.Error() {
				t.Errorf("TestAddr %v - Addr error - received: %v - expected: %v", test.testInfo, err, test.addrError)
				continue
			}
		} else if !errors.Is(test.addrError, err) {
			t.Errorf("TestAddr %v - Addr error - received: %v - expected: %v", test.testInfo, err, test.addrError)
			continue
		}
	}
}

func TestDelete(t *testing.T) {
	// empty
	env := NewEnv()
	err := env.Delete("a")
	if err != nil {
		t.Errorf("Delete error - received: %v - expected: %v", err, nil)
	}

	// bad name
	err = env.Delete("a.b")
	expectedError := "unknown symbol 'a.b'"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Delete error - received: %v - expected: %v", err, expectedError)
	}

	// add & delete a
	_ = env.Define("a", "a")
	err = env.Delete("a")
	if err != nil {
		t.Errorf("Delete error - received: %v - expected: %v", err, nil)
	}
	value, err := env.Get("a")
	expectedError = "undefined symbol 'a'"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Get error - received: %v - expected: %v", err, expectedError)
	}
	if value != nil {
		t.Errorf("Get value - received: %#v - expected: %#v", value, nil)
	}
}

func TestDeleteGlobal(t *testing.T) {
	// empty
	env := NewEnv()
	err := env.DeleteGlobal("a")
	if err != nil {
		t.Errorf("DeleteGlobal error - received: %v - expected: %v", err, nil)
	}

	// bad name
	err = env.DeleteGlobal("a.b")
	expectedError := "unknown symbol 'a.b'"
	if err == nil || err.Error() != expectedError {
		t.Errorf("DeleteGlobal error - received: %v - expected: %v", err, expectedError)
	}

	// add & delete a
	_ = env.Define("a", "a")
	err = env.DeleteGlobal("a")
	if err != nil {
		t.Errorf("DeleteGlobal error - received: %v - expected: %v", err, nil)
	}
	value, err := env.Get("a")
	expectedError = "undefined symbol 'a'"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Get error - received: %v - expected: %v", err, expectedError)
	}
	if value != nil {
		t.Errorf("Get value - received: %#v - expected: %#v", value, nil)
	}

	// parent & child, var in child, delete in parent
	envChild := env.NewEnv()
	_ = envChild.Define("a", "a")
	err = env.DeleteGlobal("a")
	if err != nil {
		t.Errorf("DeleteGlobal error - received: %v - expected: %v", err, nil)
	}
	value, err = envChild.Get("a")
	if err != nil {
		t.Errorf("Get error - received: %v - expected: %v", err, nil)
	}
	if value != "a" {
		t.Errorf("Get value - received: %#v - expected: %#v", value, "a")
	}

	// parent & child, var in child, delete in child
	err = envChild.DeleteGlobal("a")
	if err != nil {
		t.Errorf("DeleteGlobal error - received: %v - expected: %v", err, nil)
	}
	value, err = envChild.Get("a")
	if err == nil || err.Error() != expectedError {
		t.Errorf("Get error - received: %v - expected: %v", err, expectedError)
	}
	if value != nil {
		t.Errorf("Get value - received: %#v - expected: %#v", value, nil)
	}

	// parent & child, var in parent, delete in child
	_ = env.Define("a", "a")
	err = envChild.DeleteGlobal("a")
	if err != nil {
		t.Errorf("DeleteGlobal error - received: %v - expected: %v", err, nil)
	}
	value, err = envChild.Get("a")
	if err == nil || err.Error() != expectedError {
		t.Errorf("Get error - received: %v - expected: %v", err, expectedError)
	}
	if value != nil {
		t.Errorf("Get value - received: %#v - expected: %#v", value, nil)
	}

	// parent & child, var in parent, delete in parent
	_ = env.Define("a", "a")
	err = env.DeleteGlobal("a")
	if err != nil {
		t.Errorf("DeleteGlobal error - received: %v - expected: %v", err, nil)
	}
	value, err = envChild.Get("a")
	if err == nil || err.Error() != expectedError {
		t.Errorf("Get error - received: %v - expected: %v", err, expectedError)
	}
	if value != nil {
		t.Errorf("Get value - received: %#v - expected: %#v", value, nil)
	}
}

type ExternalResolver struct {
	values map[string]reflect.Value
	types  map[string]reflect.Type
}

func NewExternalResolver() *ExternalResolver {
	return &ExternalResolver{values: make(map[string]reflect.Value), types: make(map[string]reflect.Type)}
}

func (er *ExternalResolver) SetValue(name string, value any) error {
	if strings.Contains(name, ".") {
		return fmt.Errorf("unknown symbol '%s'", name)
	}

	if value == nil {
		er.values[name] = nilValue
	} else {
		er.values[name] = reflect.ValueOf(value)
	}
	return nil
}

func (er *ExternalResolver) Get(name string) (reflect.Value, error) {
	if v, ok := er.values[name]; ok {
		return v, nil
	}
	return nilValue, fmt.Errorf("undefined symbol '%s'", name)
}

func (er *ExternalResolver) DefineType(name string, t any) error {
	if strings.Contains(name, ".") {
		return fmt.Errorf("unknown symbol '%s'", name)
	}

	var typ reflect.Type
	if t == nil {
		typ = nilType
	} else {
		var ok bool
		typ, ok = t.(reflect.Type)
		if !ok {
			typ = reflect.TypeOf(t)
		}
	}

	er.types[name] = typ
	return nil
}

func (er *ExternalResolver) Type(name string) (reflect.Type, error) {
	if v, ok := er.types[name]; ok {
		return v, nil
	}
	return nilType, fmt.Errorf("undefined symbol '%s'", name)
}

func TestRaceCreateSameVariable(t *testing.T) {
	// Test creating same variable in parallel

	waitChan := make(chan struct{}, 1)
	var waitGroup sync.WaitGroup

	env := NewEnv()

	for i := 0; i < 100; i++ {
		waitGroup.Add(1)
		go func(i int) {
			<-waitChan
			err := env.Define("a", i)
			if err != nil {
				t.Errorf("Define error: %v", err)
			}
			_, err = env.Get("a")
			if err != nil {
				t.Errorf("Get error: %v", err)
			}
			waitGroup.Done()
		}(i)
	}

	close(waitChan)
	waitGroup.Wait()

	_, err := env.Get("a")
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
}

func TestRaceCreateDifferentVariables(t *testing.T) {
	// Test creating different variables in parallel

	waitChan := make(chan struct{}, 1)
	var waitGroup sync.WaitGroup

	env := NewEnv()

	for i := 0; i < 100; i++ {
		waitGroup.Add(1)
		go func(i int) {
			<-waitChan
			err := env.Define(fmt.Sprint(i), i)
			if err != nil {
				t.Errorf("Define error: %v", err)
			}
			_, err = env.Get(fmt.Sprint(i))
			if err != nil {
				t.Errorf("Get error: %v", err)
			}
			waitGroup.Done()
		}(i)
	}

	close(waitChan)
	waitGroup.Wait()

	for i := 0; i < 100; i++ {
		_, err := env.Get(fmt.Sprint(i))
		if err != nil {
			t.Errorf("Get error: %v", err)
		}
	}
}

func TestRaceReadDifferentVariables(t *testing.T) {
	// Test reading different variables in parallel

	waitChan := make(chan struct{}, 1)
	var waitGroup sync.WaitGroup

	env := NewEnv()

	for i := 0; i < 100; i++ {
		err := env.Define(fmt.Sprint(i), i)
		if err != nil {
			t.Errorf("Define error: %v", err)
		}
		_, err = env.Get(fmt.Sprint(i))
		if err != nil {
			t.Errorf("Get error: %v", err)
		}
	}

	for i := 0; i < 100; i++ {
		waitGroup.Add(1)
		go func(i int) {
			<-waitChan
			_, err := env.Get(fmt.Sprint(i))
			if err != nil {
				t.Errorf("Get error: %v", err)
			}
			waitGroup.Done()
		}(i)
	}

	close(waitChan)
	waitGroup.Wait()
}

func TestRaceSetSameVariable(t *testing.T) {
	// Test setting same variable in parallel

	waitChan := make(chan struct{}, 1)
	var waitGroup sync.WaitGroup

	env := NewEnv()

	err := env.Define("a", 0)
	if err != nil {
		t.Errorf("Define error: %v", err)
	}
	_, err = env.Get("a")
	if err != nil {
		t.Errorf("Get error: %v", err)
	}

	for i := 0; i < 100; i++ {
		waitGroup.Add(1)
		go func(i int) {
			<-waitChan
			err := env.set("a", i)
			if err != nil {
				t.Errorf("Set error: %v", err)
			}
			waitGroup.Done()
		}(i)
	}

	close(waitChan)
	waitGroup.Wait()

	_, err = env.Get("a")
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
}

func TestRaceSetSameVariableNewEnv(t *testing.T) {
	// Test setting same variable in parallel with NewEnv

	waitChan := make(chan struct{}, 1)
	var waitGroup sync.WaitGroup

	env := NewEnv()

	err := env.Define("a", 0)
	if err != nil {
		t.Errorf("Define error: %v", err)
	}
	_, err = env.Get("a")
	if err != nil {
		t.Errorf("Get error: %v", err)
	}

	for i := 0; i < 100; i++ {
		waitGroup.Add(1)
		go func(i int) {
			<-waitChan
			env = env.newEnv().newEnv()
			err := env.set("a", i)
			if err != nil {
				t.Errorf("Set error: %v", err)
			}
			waitGroup.Done()
		}(i)
	}
}

func TestRaceDefineAndSetSameVariable(t *testing.T) {
	// Test defining and setting same variable in parallel
	for i := 0; i < 100; i++ {
		raceDefineAndSetSameVariable(t)
	}
}

func raceDefineAndSetSameVariable(t *testing.T) {
	waitChan := make(chan struct{}, 1)
	var waitGroup sync.WaitGroup

	envParent := NewEnv()
	envChild := envParent.newEnv()

	for i := 0; i < 2; i++ {
		waitGroup.Add(1)
		go func() {
			<-waitChan
			err := envParent.set("a", 1)
			if err != nil && err.Error() != "unknown symbol 'a'" {
				t.Errorf("Set error: %v", err)
			}
			waitGroup.Done()
		}()
		waitGroup.Add(1)
		go func() {
			<-waitChan
			err := envParent.Define("a", 2)
			if err != nil {
				t.Errorf("Define error: %v", err)
			}
			waitGroup.Done()
		}()
		waitGroup.Add(1)
		go func() {
			<-waitChan
			err := envChild.set("a", 3)
			if err != nil && err.Error() != "unknown symbol 'a'" {
				t.Errorf("Set error: %v", err)
			}
			waitGroup.Done()
		}()
		waitGroup.Add(1)
		go func() {
			<-waitChan
			err := envChild.Define("a", 4)
			if err != nil {
				t.Errorf("Define error: %v", err)
			}
			waitGroup.Done()
		}()
	}

	close(waitChan)
	waitGroup.Wait()

	_, err := envParent.Get("a") // value of a could be 1, 2, or 3
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
	_, err = envChild.Get("a") // value of a could be 3 or 4
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
}

func BenchmarkDefine(b *testing.B) {
	var err error
	env := NewEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := env.Define("a", 1)
		if err != nil {
			b.Errorf("Set error: %v", err)
		}
	}
	b.StopTimer()
	_, err = env.Get("a")
	if err != nil {
		b.Errorf("Get error: %v", err)
	}
}

func BenchmarkSet(b *testing.B) {
	env := NewEnv()
	err := env.Define("a", 1)
	if err != nil {
		b.Errorf("Define error: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = env.set("a", 1)
		if err != nil {
			b.Errorf("Set error: %v", err)
		}
	}
	b.StopTimer()
	_, err = env.Get("a")
	if err != nil {
		b.Errorf("Get error: %v", err)
	}
}

func TestCopy(t *testing.T) {
	parent := NewEnv()
	_ = parent.Define("b", "a")
	env := parent.newEnv()
	_ = env.Define("a", "a")
	copyEnv := env.copy()
	if v, e := copyEnv.Get("a"); e != nil || v != "a" {
		t.Errorf("copy doesn't retain original values")
	}
	_ = copyEnv.set("a", "b")
	if v, e := env.Get("a"); e != nil || v != "a" {
		t.Errorf("original was modified")
	}
	if v, e := copyEnv.Get("a"); e != nil || v != "b" {
		t.Errorf("copy kept the old value")
	}
	_ = env.set("a", "c")
	if v, e := env.Get("a"); e != nil || v != "c" {
		t.Errorf("original was not modified")
	}
	if v, e := copyEnv.Get("a"); e != nil || v != "b" {
		t.Errorf("copy was modified")
	}
	if v, e := copyEnv.Get("b"); e != nil || v != "a" {
		t.Errorf("copy parent doesn't retain original value")
	}
	_ = parent.set("b", "b")
	if v, e := copyEnv.Get("b"); e != nil || v != "b" {
		t.Errorf("copy parent not modified")
	}
}

func TestDeepCopy(t *testing.T) {
	parent := NewEnv()
	_ = parent.Define("a", "a")
	env := parent.NewEnv()
	copyEnv := env.DeepCopy()
	if v, e := copyEnv.Get("a"); e != nil || v != "a" {
		t.Errorf("copy doesn't retain original values")
	}
	_ = parent.set("a", "b")
	if v, e := env.Get("a"); e != nil || v != "b" {
		t.Errorf("son was not modified")
	}
	if v, e := copyEnv.Get("a"); e != nil || v != "a" {
		t.Errorf("copy got the new value")
	}
	_ = parent.set("a", "c")
	if v, e := env.Get("a"); e != nil || v != "c" {
		t.Errorf("original was not modified")
	}
	if v, e := copyEnv.Get("a"); e != nil || v != "a" {
		t.Errorf("copy was modified")
	}
	_ = parent.Define("b", "b")
	if _, e := copyEnv.Get("b"); e == nil {
		t.Errorf("copy parent was modified")
	}
}

func TestGetParentValue(t *testing.T) {
	env := NewEnv()
	val, _ := env.Type("bool")
	assert.Equal(t, "bool", val.String())
}

func TestEnv_String(t *testing.T) {
	env := NewEnv()
	_ = env.Define("a", 1)
	_ = env.Define("b", func() {})
	_, _ = env.NewModule("c")
	_ = env.DefineType("t", "bool")
	_ = env.DefineType("u", "bool")
	assert.Equal(t, "No parent\na = 1\nb = func()\nc = module<c>\nt = string\nu = string\n", env.String())
}

func TestEnv_AddPackage(t *testing.T) {
	env := NewEnv()
	_, _ = env.AddPackage("a", map[string]any{"m1": 1}, map[string]any{"t1": "bool"})
	val, _ := env.Get("a")
	pack := val.(*Env)
	assert.Equal(t, 1, pack.Values().Len())

	_, err := env.AddPackage("a.b", map[string]any{}, map[string]any{})
	assert.Error(t, err)

	_, err = env.AddPackage("a", map[string]any{"in.valid": 1}, map[string]any{})
	assert.Error(t, err)

	_, err = env.AddPackage("a", map[string]any{}, map[string]any{"in.valid": "bool"})
	assert.Error(t, err)
}

func TestEnv_NewModule(t *testing.T) {
	env := NewEnv()
	_, err := env.NewModule("in.valid")
	assert.Error(t, err)
}

func TestEnv_GetEnvFromPath(t *testing.T) {
	env := NewEnv()
	_, err := env.GetEnvFromPath([]string{})
	assert.NoError(t, err)
	_, err = env.GetEnvFromPath([]string{"a"})
	assert.Error(t, err)
	_, _ = env.NewModule("a")
	_, err = env.GetEnvFromPath([]string{"a"})
	assert.NoError(t, err)
}

func TestEnv_DefineGlobalValue(t *testing.T) {
	env := NewEnv()
	newEnv1 := env.NewEnv()
	_ = newEnv1.DefineGlobalValue("a", reflect.ValueOf(1))
	val, _ := env.Get("a")
	assert.Equal(t, 1, val)
}

func TestEnv_Types(t *testing.T) {
	env := NewEnv()
	assert.Equal(t, 0, env.Types().Len())
}

func TestEnv_Defers(t *testing.T) {
	env := NewEnv()
	assert.Equal(t, 0, env.Defers().Len())
}

func TestEnv_DefineReflectType(t *testing.T) {
	env := NewEnv()
	err := env.DefineReflectType("a", reflect.TypeOf("a"))
	assert.NoError(t, err)
}

func TestEnv_DefineValue(t *testing.T) {
	env := NewEnv()
	err := env.DefineValue("a", reflect.ValueOf("a"))
	assert.NoError(t, err)
}

func TestEnv_Set(t *testing.T) {
	env := NewEnv()
	err := env.set("a", reflect.ValueOf("a"))
	assert.Error(t, err)
	_ = env.define("a", "a")
	err = env.set("a", reflect.ValueOf("a"))
	assert.NoError(t, err)
	assert.Equal(t, 1, env.Values().Len())
}

func TestEnv_SetValue(t *testing.T) {
	env := NewEnv()
	err := env.SetValue("a", reflect.ValueOf("a"))
	assert.Error(t, err)
	_ = env.define("a", "a")
	err = env.SetValue("a", reflect.ValueOf("a"))
	assert.NoError(t, err)
	assert.Equal(t, 1, env.Values().Len())
}

func TestEnv_Name(t *testing.T) {
	env := NewEnv()
	assert.Equal(t, "n/a", env.Name())
}
