package vm

import (
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/packages"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"reflect"
	"sync"
)

func isNil(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		// from reflect IsNil:
		// Note that IsNil is not always equivalent to a regular comparison with nil in Go.
		// For example, if v was created by calling ValueOf with an uninitialized interface variable i,
		// i==nil will be true but v.IsNil will panic as v will be the zero Value.
		return v.IsNil()
	default:
		return false
	}
}

func isNum(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// equal returns true when lhsV and rhsV is same value.
func equal(lhsV, rhsV reflect.Value) bool {
	lhsNotValid, rhsVNotValid := !lhsV.IsValid(), !rhsV.IsValid()
	if lhsNotValid && rhsVNotValid {
		return true
	}
	if (!lhsNotValid && rhsVNotValid) || (lhsNotValid && !rhsVNotValid) {
		return false
	}

	lhsIsNil, rhsIsNil := isNil(lhsV), isNil(rhsV)
	if lhsIsNil && rhsIsNil {
		return true
	}
	if (!lhsIsNil && rhsIsNil) || (lhsIsNil && !rhsIsNil) {
		return false
	}
	if lhsV.Kind() == reflect.Interface || lhsV.Kind() == reflect.Pointer {
		lhsV = lhsV.Elem()
	}
	if rhsV.Kind() == reflect.Interface || rhsV.Kind() == reflect.Pointer {
		rhsV = rhsV.Elem()
	}

	// Compare a string and a number.
	// This will attempt to convert the string to a number,
	// while leaving the other side alone. Code further
	// down takes care of converting ints and floats as needed.
	if isNum(lhsV) && rhsV.Kind() == reflect.String {
		rhsF, err := tryToFloat64(rhsV)
		if err != nil {
			// Couldn't convert RHS to a float, they can't be compared.
			return false
		}
		rhsV = reflect.ValueOf(rhsF)
	} else if lhsV.Kind() == reflect.String && isNum(rhsV) {
		// If the LHS is a string formatted as an int, try that before trying float
		lhsI, err := tryToInt64(lhsV)
		if err != nil {
			// if LHS is a float, e.g. "1.2", we need to set lhsV to a float64
			lhsF, err := tryToFloat64(lhsV)
			if err != nil {
				return false
			}
			lhsV = reflect.ValueOf(lhsF)
		} else {
			lhsV = reflect.ValueOf(lhsI)
		}
	}

	if isNum(lhsV) && isNum(rhsV) {
		return fmt.Sprintf("%v", lhsV) == fmt.Sprintf("%v", rhsV)
	}

	// Try to compare bools to strings and numbers
	if lhsV.Kind() == reflect.Bool || rhsV.Kind() == reflect.Bool {
		lhsB, err := tryToBool(lhsV)
		if err != nil {
			return false
		}
		rhsB, err := tryToBool(rhsV)
		if err != nil {
			return false
		}
		return lhsB == rhsB
	}

	if lhsV.CanInterface() && rhsV.CanInterface() {
		return reflect.DeepEqual(lhsV.Interface(), rhsV.Interface())
	}
	return reflect.DeepEqual(lhsV, rhsV)
}

func readMapIndex(aMap, key reflect.Value, m sync.Locker) (out reflect.Value) {
	m.Lock()
	defer m.Unlock()
	return aMap.MapIndex(key)
}

func setMapIndex(aMap, key, value reflect.Value, m sync.Locker) {
	m.Lock()
	defer m.Unlock()
	aMap.SetMapIndex(key, value)
}

func mapIterNext(mapIter *reflect.MapIter, m sync.Locker) bool {
	m.Lock()
	defer m.Unlock()
	return mapIter.Next()
}

func getMapIndex(vmp *vmParams, key reflect.Value, aMap reflect.Value) reflect.Value {
	nilValueL := nilValue
	if !aMap.IsValid() || aMap.IsNil() {
		return nilValueL
	}

	keyType := key.Type()
	if keyType == interfaceType && aMap.Type().Key() != interfaceType {
		if key.Elem().IsValid() && !key.Elem().IsNil() {
			keyType = key.Elem().Type()
		}
	}
	if keyType != aMap.Type().Key() && aMap.Type().Key() != interfaceType {
		return nilValueL
	}

	// From reflect MapIndex:
	// It returns the zero Value if key is not found in the map or if v represents a nil map.
	value := readMapIndex(aMap, key, vmp.mapMutex)

	if value.IsValid() && value.CanInterface() && aMap.Type().Elem() == interfaceType && !value.IsNil() {
		value = reflect.ValueOf(value.Interface())
	}

	return value
}

func appendSlice(expr ast.Expr, lhsV reflect.Value, rhsV reflect.Value) (reflect.Value, error) {
	nilValueL := nilValue
	lhsT := lhsV.Type().Elem()
	rhsT := rhsV.Type().Elem()

	if lhsT == rhsT {
		return reflect.AppendSlice(lhsV, rhsV), nil
	}

	if rhsT.ConvertibleTo(lhsT) {
		for i := 0; i < rhsV.Len(); i++ {
			lhsV = reflect.Append(lhsV, rhsV.Index(i).Convert(lhsT))
		}
		return lhsV, nil
	}

	leftHasSubArray := lhsT.Kind() == reflect.Slice || lhsT.Kind() == reflect.Array
	rightHasSubArray := rhsT.Kind() == reflect.Slice || rhsT.Kind() == reflect.Array

	if leftHasSubArray != rightHasSubArray && lhsT != interfaceType && rhsT != interfaceType {
		return nilValueL, newStringError(expr, "invalid type conversion")
	}

	if !leftHasSubArray && !rightHasSubArray {
		for i := 0; i < rhsV.Len(); i++ {
			value := rhsV.Index(i)
			if rhsT == interfaceType {
				value = value.Elem()
			}
			if lhsT == value.Type() {
				lhsV = reflect.Append(lhsV, value)
			} else if value.Type().ConvertibleTo(lhsT) {
				lhsV = reflect.Append(lhsV, value.Convert(lhsT))
			} else {
				return nilValueL, newStringError(expr, "invalid type conversion")
			}
		}
		return lhsV, nil
	}

	if (leftHasSubArray || lhsT == interfaceType) && (rightHasSubArray || rhsT == interfaceType) {
		for i := 0; i < rhsV.Len(); i++ {
			value := rhsV.Index(i)
			if rhsT == interfaceType {
				value = value.Elem()
				if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
					return nilValueL, newStringError(expr, "invalid type conversion")
				}
			}
			newSlice, err := appendSlice(expr, reflect.MakeSlice(lhsT, 0, value.Len()), value)
			if err != nil {
				return nilValueL, err
			}
			lhsV = reflect.Append(lhsV, newSlice)
		}
		return lhsV, nil
	}

	return nilValueL, newStringError(expr, "invalid type conversion")
}

func getTypeFromEnv(env envPkg.IEnv, typeStruct *ast.TypeStruct) (reflect.Type, error) {
	e, err := env.GetEnvFromPath(typeStruct.Env)
	if err != nil {
		return nil, err
	}
	t, err := e.Type(typeStruct.Name)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func makeValue(t reflect.Type) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Chan:
		return makeValueChan(t)
	case reflect.Func:
		return makeValueFunc(t)
	case reflect.Map:
		return makeValueMap(t)
	case reflect.Pointer:
		return makeValuePointer(t)
	case reflect.Slice:
		return makeValueSlice(t)
	case reflect.Struct:
		return makeValueStruct(t)
	default:
		return makeValueDefault(t)
	}
}

func makeValueChan(t reflect.Type) (reflect.Value, error) {
	return reflect.MakeChan(t, 0), nil
}

func makeValueFunc(t reflect.Type) (reflect.Value, error) {
	return reflect.MakeFunc(t, nil), nil
}

func makeValueMap(t reflect.Type) (reflect.Value, error) {
	// note creating slice as work around to create map
	// just doing MakeMap can give incorrect type for defined types
	value := reflect.MakeSlice(reflect.SliceOf(t), 0, 1)
	value = reflect.Append(value, reflect.MakeMap(reflect.MapOf(t.Key(), t.Elem())))
	return value.Index(0), nil
}

func makeValuePointer(t reflect.Type) (reflect.Value, error) {
	ptrV := reflect.New(t.Elem())
	v, err := makeValue(t.Elem())
	if err != nil {
		return nilValue, err
	}
	if !ptrV.Elem().CanSet() {
		return nilValue, fmt.Errorf("type " + t.String() + " cannot be assigned")
	}
	ptrV.Elem().Set(v)
	return ptrV, nil
}

func makeValueSlice(t reflect.Type) (reflect.Value, error) {
	return reflect.MakeSlice(t, 0, 0), nil
}

func makeValueStruct(t reflect.Type) (reflect.Value, error) {
	structV := reflect.New(t).Elem()
	for i := 0; i < structV.NumField(); i++ {
		if structV.Field(i).Kind() == reflect.Ptr {
			continue
		}
		v, err := makeValue(structV.Field(i).Type())
		if err != nil {
			return nilValue, err
		}
		if structV.Field(i).CanSet() {
			structV.Field(i).Set(v)
		}
	}
	return structV, nil
}

func makeValueDefault(t reflect.Type) (reflect.Value, error) {
	return reflect.New(t).Elem(), nil
}

// DefineImport defines the vm import command that will import packages and package types when wanted
func DefineImport(e envPkg.IEnv) {
	_ = e.Define("import", importFn(e))
}

func importFn(e envPkg.IEnv) func(string) envPkg.IEnv {
	return func(source string) envPkg.IEnv {
		methods, ok := packages.Packages.Get(source)
		if !ok {
			panic(fmt.Sprintf("package '%s' not found", source))
		}
		var err error
		pack, _ := e.NewModule(source)
		for methodName, methodValue := range methods {
			err = pack.Define(methodName, methodValue)
			if err != nil {
				panic(fmt.Sprintf("import Define error: %v", err))
			}
		}
		types, ok := packages.PackageTypes.Get(source)
		if ok {
			for typeName, typeValue := range types {
				err = pack.DefineType(typeName, typeValue)
				if err != nil {
					panic(fmt.Sprintf("import DefineType error: %v", err))
				}
			}
		}
		return pack
	}
}
