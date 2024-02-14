package vm

import (
	"context"
	"errors"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/compiler"
	"github.com/alaingilbert/anko/pkg/parser"
	"reflect"
	"testing"
	"time"
)

// Test is utility struct to make tests easy.
type Test struct {
	Script         string
	ParseError     error
	ParseErrorFunc *func(*testing.T, error)
	Types          map[string]any
	Input          map[string]any
	RunError       error
	RunErrorLine   int
	RunErrorColumn int
	RunErrorFunc   *func(*testing.T, error)
	RunOutput      any
	Output         map[string]any
}

// Options is utility struct to pass options to the test.
type Options struct {
	DefineImport bool
	ImportCore   bool
}

// Run runs VM tests
func runTests(t *testing.T, tests []Test, testingOptions *Options) {
	for _, test := range tests {
		runTest(t, test, testingOptions)
	}
}

func runTest(t *testing.T, test Test, testingOptions *Options) {
	runTestFromSource(t, test, testingOptions)
	runTestFromCompiledSource(t, test, testingOptions)
}

func runTestFromCompiledSource(t *testing.T, test Test, testingOptions *Options) {
	compiled, err := compiler.Compile(test.Script, false)
	if test.ParseErrorFunc != nil {
		(*test.ParseErrorFunc)(t, err)
	} else if err != nil && test.ParseError != nil {
		if err.Error() != test.ParseError.Error() {
			t.Errorf("ParseSrc error - received: %v - expected: %v - script: %v", err, test.ParseError, test.Script)
			return
		}
	} else if !errors.Is(err, test.ParseError) {
		t.Errorf("ParseSrc error - received: %v - expected: %v - script: %v", err, test.ParseError, test.Script)
		return
	}
	if err != nil {
		return
	}
	stmts := compiler.Decode(compiled)
	runTest1(t, test, testingOptions, stmts)
}

func runTestFromSource(t *testing.T, test Test, testingOptions *Options) {
	stmts, err := parser.ParseSrc(test.Script)
	if test.ParseErrorFunc != nil {
		(*test.ParseErrorFunc)(t, err)
	} else if err != nil && test.ParseError != nil {
		if err.Error() != test.ParseError.Error() {
			t.Errorf("ParseSrc error - received: %v - expected: %v - script: %v", err, test.ParseError, test.Script)
			return
		}
	} else if !errors.Is(err, test.ParseError) {
		t.Errorf("ParseSrc error - received: %v - expected: %v - script: %v", err, test.ParseError, test.Script)
		return
	}
	// Note: Still want to run the code even after a parse error to see what happens
	runTest1(t, test, testingOptions, stmts)
}

func runTest1(t *testing.T, test Test, testingOptions *Options, stmts []ast.Stmt) {
	// parser.EnableErrorVerbose()
	var err error

	configs := &Configs{}
	if testingOptions != nil && testingOptions.DefineImport {
		configs.DefineImport = true
	}
	if testingOptions != nil && testingOptions.ImportCore {
		configs.ImportCore = true
	}
	v := New(configs)

	for typeName, typeValue := range test.Types {
		err = v.DefineType(typeName, typeValue)
		if err != nil {
			t.Errorf("DefineType error: %v - typeName: %v - script: %v", err, typeName, test.Script)
			return
		}
	}

	for inputName, inputValue := range test.Input {
		err = v.Define(inputName, inputValue)
		if err != nil {
			t.Errorf("Define error: %v - inputName: %v - script: %v", err, inputName, test.Script)
			return
		}
	}

	var value any
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	e := v.Executor()
	value, err = e.Run(ctx, stmts)
	if test.RunErrorFunc != nil {
		(*test.RunErrorFunc)(t, err)
	} else if err != nil && test.RunError != nil {
		if err.Error() != test.RunError.Error() {
			t.Errorf("Run error - received: %v - expected: %v - script: %v", err, test.RunError, test.Script)
			return
		}
		var err *Error
		if err != nil && errors.As(err, &err) {
			if test.RunErrorLine != 0 && err.Pos.Line != test.RunErrorLine {
				t.Errorf("Run error line - received: %v - expected: %v - script: %v", err.Pos.Line, test.RunErrorLine, test.Script)
				return
			} else if test.RunErrorColumn != 0 && err.Pos.Column != test.RunErrorColumn {
				t.Errorf("Run error line - received: %v - expected: %v - script: %v", err.Pos.Column, test.RunErrorColumn, test.Script)
				return
			}
		}
	} else if !errors.Is(err, test.RunError) {
		t.Errorf("Run error - received: %v - expected: %v - script: %v", err, test.RunError, test.Script)
		return
	}

	if !valueEqual(value, test.RunOutput) {
		t.Errorf("Run output - received: %#v - expected: %#v - script: %v", value, test.RunOutput, test.Script)
		t.Errorf("received type: %T - expected: %T", value, test.RunOutput)
		return
	}

	for outputName, outputValue := range test.Output {
		value, err = e.GetEnv().Get(outputName)
		if err != nil {
			t.Errorf("Get error: %v - outputName: %v - script: %v", err, outputName, test.Script)
			return
		}

		if !valueEqual(value, outputValue) {
			t.Errorf("outputName %v - received: %#v - expected: %#v - script: %v", outputName, value, outputValue, test.Script)
			t.Errorf("received type: %T - expected: %T", value, outputValue)
			continue
		}
	}
}

// ValueEqual return true if v1 and v2 is same value. If passed function, does
// extra checks otherwise just doing reflect.DeepEqual
func valueEqual(v1 any, v2 any) bool {
	v1RV := reflect.ValueOf(v1)
	switch v1RV.Kind() {
	case reflect.Func:
		// This is best effort to check if functions match, but it could be wrong
		v2RV := reflect.ValueOf(v2)
		if !v1RV.IsValid() || !v2RV.IsValid() {
			if v1RV.IsValid() != !v2RV.IsValid() {
				return false
			}
			return true
		} else if v1RV.Kind() != v2RV.Kind() {
			return false
		} else if v1RV.Type() != v2RV.Type() {
			return false
		} else if v1RV.Pointer() != v2RV.Pointer() {
			// From reflect: If v's Kind is Func, the returned pointer is an underlying code pointer, but not necessarily enough to identify a single function uniquely.
			return false
		}
		return true
	default:
		return reflect.DeepEqual(v1, v2)
	}
}
