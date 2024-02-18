package executor

import (
	"context"
	"github.com/alaingilbert/anko/pkg/ast"
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
}
