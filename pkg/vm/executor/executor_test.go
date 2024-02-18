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

func TestValidate(t *testing.T) {
	e := NewExecutor(&Config{Env: envPkg.NewEnv()})
	ctx := context.Background()
	script := "a = 1; b = 2; if a == b { return a; }; return b"
	err := e.Validate(ctx, script)
	assert.NoError(t, err)

	script = "a = func(){ return 1; }; a()"
	assert.NoError(t, e.Validate(ctx, script))

	script = "a = func(){ invalidFn() }; a()"
	assert.ErrorContains(t, e.Validate(ctx, script), "undefined symbol 'invalidFn'")

	script = "a = func(){ if (true) { invalidFn() } }; a()"
	assert.ErrorContains(t, e.Validate(ctx, script), "undefined symbol 'invalidFn'")

	script = "a = func(){ if (false) { invalidFn() } }; a()"
	assert.ErrorContains(t, e.Validate(ctx, script), "undefined symbol 'invalidFn'")
}
