package vm

import (
	"fmt"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ImportToX adds all the toX to the env given
func ImportToX(env envPkg.IEnv) {
	_ = env.Define("toBool", toBoolFn)
	_ = env.Define("toString", toStringFn)
	_ = env.Define("toInt", toIntFn)
	_ = env.Define("toFloat", toFloatFn)
	_ = env.Define("toChar", toCharFn)
	_ = env.Define("toRune", toRuneFn)
	_ = env.Define("toBoolSlice", toBoolSliceFn)
	_ = env.Define("toStringSlice", toStringSliceFn)
	_ = env.Define("toIntSlice", toIntSliceFn)
	_ = env.Define("toFloatSlice", foFloatSliceFn)
	_ = env.Define("toByteSlice", toByteSliceFn)
	_ = env.Define("toRuneSlice", toRuneSliceFn)
	_ = env.Define("toDuration", toDurationFn)
}

func toBoolFn(v any) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return false
	}
	nt := reflect.TypeOf(true)
	if rv.Type().ConvertibleTo(nt) {
		return rv.Convert(nt).Bool()
	}
	if rv.Type().ConvertibleTo(reflect.TypeOf(1.0)) && rv.Convert(reflect.TypeOf(1.0)).Float() > 0.0 {
		return true
	}
	if rv.Kind() == reflect.String {
		s := strings.ToLower(v.(string))
		if s == "y" || s == "yes" {
			return true
		}
		b, err := strconv.ParseBool(s)
		if err == nil {
			return b
		}
	}
	return false
}

func toStringFn(v any) string {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return fmt.Sprint(v)
}

func toIntFn(v any) int64 {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return 0
	}
	nt := reflect.TypeOf(1)
	if rv.Type().ConvertibleTo(nt) {
		return rv.Convert(nt).Int()
	}
	if rv.Kind() == reflect.String {
		i, err := strconv.ParseInt(v.(string), 10, 64)
		if err == nil {
			return i
		}
		f, err := strconv.ParseFloat(v.(string), 64)
		if err == nil {
			return int64(f)
		}
	}
	if rv.Kind() == reflect.Bool {
		if v.(bool) {
			return 1
		}
	}
	return 0
}

func toFloatFn(v any) float64 {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return 0
	}
	nt := reflect.TypeOf(1.0)
	if rv.Type().ConvertibleTo(nt) {
		return rv.Convert(nt).Float()
	}
	if rv.Kind() == reflect.String {
		f, err := strconv.ParseFloat(v.(string), 64)
		if err == nil {
			return f
		}
	}
	if rv.Kind() == reflect.Bool {
		if v.(bool) {
			return 1.0
		}
	}
	return 0.0
}

func toCharFn(s rune) string {
	return string(s)
}

func toRuneFn(s string) rune {
	if len(s) == 0 {
		return 0
	}
	return []rune(s)[0]
}

func toBoolSliceFn(v []any) (out []bool) {
	toSlice(v, &out)
	return
}

func toStringSliceFn(v []any) (out []string) {
	toSlice(v, &out)
	return
}

func toIntSliceFn(v []any) (out []int64) {
	toSlice(v, &out)
	return
}

func foFloatSliceFn(v []any) (out []float64) {
	toSlice(v, &out)
	return
}

func toByteSliceFn(s string) []byte {
	return []byte(s)
}

func toRuneSliceFn(s string) []rune {
	return []rune(s)
}

func toDurationFn(v int64) time.Duration {
	return time.Duration(v)
}

// toSlice takes in a "generic" slice and converts and copies
// it's elements into the typed slice pointed at by ptr.
// Note that this is a costly operation.
func toSlice(from []any, ptr any) {
	// Value of the pointer to the target
	obj := reflect.Indirect(reflect.ValueOf(ptr))
	// We can't just convert from any to whatever the target is (diff memory layout),
	// so we need to create a New slice of the proper type and copy the values individually
	t := reflect.TypeOf(ptr).Elem()
	tt := t.Elem()
	slice := reflect.MakeSlice(t, len(from), len(from))
	// Copying the data, val is an addressable Pointer of the actual target type
	val := reflect.Indirect(reflect.New(tt))
	for i := 0; i < len(from); i++ {
		v := reflect.ValueOf(from[i])
		if v.IsValid() && v.Type().ConvertibleTo(tt) {
			val.Set(v.Convert(tt))
		} else {
			val.Set(reflect.Zero(tt))
		}
		slice.Index(i).Set(val)
	}
	// Ok now assign our slice to the target pointer
	obj.Set(slice)
}
