package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

func First[T any](a T, _ ...any) T { return a }

func Ternary[T any](predicate bool, a, b T) T {
	if predicate {
		return a
	}
	return b
}

// TernaryZ returns a if the predicate is true (not its zero value)
func TernaryZ[T comparable](predicate T, a, b T) T {
	var zero T
	return Ternary(predicate != zero, a, b)
}

// MD5 returns md5 hex sum as a string
func MD5(in []byte) string {
	h := md5.New()
	h.Write(in)
	return hex.EncodeToString(h.Sum(nil))
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func MustErr[T any](v T, err error) error {
	if err == nil {
		panic("error expected")
	}
	return err
}

func DefaultCtx(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx
}

func Default[T any](v *T, d T) T {
	if v == nil {
		return d
	}
	return *v
}

func Bool(v bool) *bool { return &v }

func DoParseI64(v string) int64 {
	parsed, _ := strconv.ParseInt(v, 10, 64)
	return parsed
}

func Clamp[T ~int | ~int64](val, minVal, maxVal T) T {
	val = min(val, maxVal)
	val = max(val, minVal)
	return val
}
