package executor

import (
	"context"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/compiler"
	"github.com/alaingilbert/anko/pkg/parser"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCycles(t *testing.T) {
	script := "a = 1; b = 2; if a == b { return a; }; return b"
	env := envPkg.NewEnv()
	e := NewExecutor(&Config{Env: env})
	val, err := e.Run(context.Background(), script)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
	assert.Equal(t, int64(11), e.getCycles())
}

func TestInvalidInput(t *testing.T) {
	env := envPkg.NewEnv()
	e := NewExecutor(&Config{Env: env})
	stmt := &ast.ForStmt{}
	val, err := e.Run(context.Background(), stmt)
	assert.ErrorIs(t, err, ErrInvalidInput)
	assert.Equal(t, nil, val)

	_, err = e.Run(context.Background(), 123)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestRunStmts(t *testing.T) {
	script := "a = 1; b = 2; if a == b { return a; }; return b"
	env := envPkg.NewEnv()
	e := NewExecutor(&Config{Env: env})
	stmts, _ := parser.ParseSrc(script)
	val, err := e.Run(context.Background(), stmts)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
	assert.Equal(t, int64(11), e.getCycles())
}

func TestRunCompiled(t *testing.T) {
	script := "a = 1; b = 2; if a == b { return a; }; return b"
	by, _ := compiler.Compile(script, false)
	env := envPkg.NewEnv()
	e := NewExecutor(&Config{Env: env})
	val, err := e.Run(context.Background(), by)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
	assert.Equal(t, int64(11), e.getCycles())
}

func TestInvalidString(t *testing.T) {
	script := "a ==== 1"
	env := envPkg.NewEnv()
	e := NewExecutor(&Config{Env: env})
	_, err := e.Run(context.Background(), script)
	var target *parser.Error
	assert.ErrorAs(t, err, &target)
}

func TestExecutor_Validate(t *testing.T) {
	invalidFnErr := envPkg.NewUndefinedSymbolErr("invalidFn")
	tests := []struct {
		name    string
		input   any
		wantErr error
	}{
		{input: "a = 1; b = 2; if a == b { return a; }; return b", wantErr: nil, name: ""},
		{input: "a = func(){ return 1; }; a()", wantErr: nil, name: ""},
		{input: "a = func(){ invalidFn() }; a()", wantErr: invalidFnErr, name: ""},
		{input: "a = func(){ if (true) { invalidFn() } }; a()", wantErr: invalidFnErr, name: ""},
		{input: "a = func(){ if (false) { invalidFn() } }; a()", wantErr: invalidFnErr, name: ""},
		// {input: "a = func(){ if (true) { return 1; } else { invalidFn() } }; a()", wantErr: invalidFnErr, name: ""},
		{input: "a = func(){ if (true) { } else { invalidFn() } }; a()", wantErr: invalidFnErr, name: ""},
		{input: "a = func(){ if true { } else if true { } else { invalidFn() } }; a()", wantErr: invalidFnErr, name: ""},
		{input: "a = func(){ if true { } else if true { invalidFn() } else { } }; a()", wantErr: invalidFnErr, name: ""},
		{input: "a = func(){ try { } catch err { invalidFn() } finally { } }; a()", wantErr: invalidFnErr, name: ""},
		{input: "switch 0 { case 1: invalidFn(); }", wantErr: invalidFnErr, name: ""},
		{input: "switch 0 { case 1: 1; default: invalidFn(); }", wantErr: invalidFnErr, name: ""},
		{input: "ch1 = make(chan int); ch2 = make(chan int); select { case <-ch1: invalidFn(); case <-ch2: 2; }", wantErr: invalidFnErr, name: ""},
		{input: "ch1 = make(chan int); ch2 = make(chan int); select { case <-ch1: 1; case <-ch2: invalidFn(); }", wantErr: invalidFnErr, name: ""},
		{input: "ch1 = make(chan int); select { case <-ch1: 1; default: invalidFn(); }", wantErr: invalidFnErr, name: ""},
		{input: "a = func() { defer func() { invalidFn(); }(); }; a()", wantErr: invalidFnErr, name: ""},
	}
	e := NewExecutor(&Config{Env: envPkg.NewEnv()})
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := e.Validate(ctx, tt.input)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr.Error())
			}
		})
	}
}

func TestExecutor_Has(t *testing.T) {
	env := envPkg.NewEnv()
	myFn1 := func() {}
	myFn2 := func() {}
	_ = env.Define("myFn1", myFn1)
	_ = env.Define("myFn2", myFn2)
	e := NewExecutor(&Config{Env: env})

	ctx := context.Background()
	tests := []struct {
		name    string
		input   any
		wantErr error
		targets []any
		wantOks []bool
	}{
		{input: "myFn1()", targets: []any{myFn1}, wantOks: []bool{true}, wantErr: nil, name: ""},
		{input: "myFn1()", targets: []any{myFn2}, wantOks: []bool{false}, wantErr: nil, name: ""},
		{input: "myFn1()", targets: []any{myFn1, myFn2}, wantOks: []bool{true, false}, wantErr: nil, name: ""},
		{input: "myFn2()", targets: []any{myFn1, myFn2}, wantOks: []bool{false, true}, wantErr: nil, name: ""},
		{input: "myFn1(); myFn2()", targets: []any{myFn1, myFn2}, wantOks: []bool{true, true}, wantErr: nil, name: ""},
		{input: "a = myFn1; b = myFn2", targets: []any{myFn1, myFn2}, wantOks: []bool{false, false}, wantErr: nil, name: ""},
		{input: "a = myFn1; b = myFn2; a()", targets: []any{myFn1, myFn2}, wantOks: []bool{true, false}, wantErr: nil, name: ""},
		{input: "a = myFn1; b = myFn2; a(); b()", targets: []any{myFn1, myFn2}, wantOks: []bool{true, true}, wantErr: nil, name: ""},
		{input: "a = func() { myFn1(); }", targets: []any{myFn1, myFn2}, wantOks: []bool{true, false}, wantErr: nil, name: ""},
		{input: "a = func() { defer func() { myFn1(); }() }", targets: []any{myFn1, myFn2}, wantOks: []bool{true, false}, wantErr: nil, name: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oks, err := e.Has(ctx, tt.input, tt.targets)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr.Error())
			}
			assert.Equal(t, tt.wantOks, oks)
		})
	}
}
