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

func TestDefault(t *testing.T) {
	assert.Equal(t, true, Default((*bool)(nil), true))
	assert.Equal(t, false, Default((*bool)(nil), false))
	assert.Equal(t, true, Default(Ptr(true), false))
	assert.Equal(t, false, Default(Ptr(false), true))
}

func TestOverride(t *testing.T) {
	assert.Equal(t, (*bool)(nil), Override((*bool)(nil), (*bool)(nil)))
	assert.Equal(t, true, *Override((*bool)(nil), Ptr(true)))
	assert.Equal(t, false, *Override((*bool)(nil), Ptr(false)))
	assert.Equal(t, true, *Override(Ptr(true), (*bool)(nil)))
	assert.Equal(t, true, *Override(Ptr(true), Ptr(true)))
	assert.Equal(t, false, *Override(Ptr(true), Ptr(false)))
	assert.Equal(t, false, *Override(Ptr(false), (*bool)(nil)))
	assert.Equal(t, true, *Override(Ptr(false), Ptr(true)))
	assert.Equal(t, false, *Override(Ptr(false), Ptr(false)))
}
