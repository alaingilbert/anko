package utils

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMust(t *testing.T) {
	resFn := func() (int, error) { return 1, nil }
	errFn := func() (int, error) { return 0, errors.New("err") }
	assert.NotPanics(t, func() { Must(resFn()) })
	assert.Panics(t, func() { Must(errFn()) })
}

func TestMustErr(t *testing.T) {
	resFn := func() (int, error) { return 1, nil }
	errFn := func() (int, error) { return 0, errors.New("err") }
	assert.Panics(t, func() { _ = MustErr(resFn()) })
	assert.NotPanics(t, func() { _ = MustErr(errFn()) })
}

func TestTernary(t *testing.T) {
	assert.Equal(t, 1, Ternary(true, 1, 2))
	assert.Equal(t, 2, Ternary(false, 1, 2))
}

func TestTernaryZ(t *testing.T) {
	assert.Equal(t, 1, TernaryZ(10, 1, 2))
	assert.Equal(t, 2, TernaryZ(0, 1, 2))
}

func TestMD5(t *testing.T) {
	assert.Equal(t, "acbd18db4cc2f85cedef654fccc4a4d8", MD5([]byte("foo")))
}
