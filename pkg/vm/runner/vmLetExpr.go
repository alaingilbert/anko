package runner

import (
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"reflect"
)

func invokeLetExpr(vmp *VmParams, env envPkg.IEnv, stmt *ast.LetsStmt, expr ast.Expr, rv reflect.Value) (reflect.Value, error) {
	switch lhs := expr.(type) {
	case *ast.IdentExpr:
		return invokeLetIdentExpr(env, rv, stmt, lhs)
	case *ast.MemberExpr:
		return invokeLetMemberExpr(vmp, env, rv, stmt, lhs)
	case *ast.ItemExpr:
		return invokeLetItemExpr(vmp, rv, env, stmt, lhs)
	case *ast.SliceExpr:
		return invokeLetSliceExpr(vmp, env, rv, lhs)
	case *ast.DerefExpr:
		return invokeLetDerefExpr(vmp, env, rv, stmt, lhs)
	}
	return nilValue, newError(expr, ErrInvalidOperation)
}

func invokeLetIdentExpr(env envPkg.IEnv, rv reflect.Value, stmt *ast.LetsStmt, lhs *ast.IdentExpr) (vv reflect.Value, err error) {
	stmtTyped := stmt.Typed
	if env.HasValue(lhs.Lit) && stmtTyped {
		return nilValue, newError(lhs, NewSymbolAlreadyDefinedError(lhs.Lit))
	}
	if stmtTyped {
		rv = reflect.ValueOf(vmUtils.NewStronglyTyped(rv, stmt.Mutable))
	}
	if err := env.SetValue(lhs.Lit, rv); err != nil {
		if errors.Is(err, vmUtils.ErrTypeMismatch) {
			return nilValue, newError(lhs, err)
		}
		if errors.Is(err, vmUtils.ErrImmutable) {
			return nilValue, newError(lhs, err)
		}
		if err := env.DefineValue(lhs.Lit, rv); err != nil {
			return nilValue, err
		}
	}
	return rv, nil
}

func invokeLetMemberExpr(vmp *VmParams, env envPkg.IEnv, rv reflect.Value, stmt *ast.LetsStmt, lhs *ast.MemberExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, lhs.Expr)
	if err != nil {
		return nilValueL, newError(lhs, err)
	}

	v = elemIfInterface(v)
	if !v.IsValid() {
		return nilValueL, newError(lhs, ErrNoSupportMemberOpInvalid)
	}
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nilValueL, newError(lhs, ErrNoSupportMemberOpInvalid)
	}

	switch v.Kind() {
	case reflect.Struct:
		return invokeLetMemberStructExpr(vmp, v, rv, lhs)
	case reflect.Map:
		return invokeLetMemberMapExpr(vmp, env, stmt, v, rv, lhs)
	default:
		err := NewNoSupportMemberOpError(v.Kind().String())
		return nilValueL, newError(lhs, err)
	}
}

func invokeLetMemberStructExpr(vmp *VmParams, v, rv reflect.Value, lhs *ast.MemberExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	field, found := v.Type().FieldByName(lhs.Name)
	if !found {
		return nilValueL, newStringError(lhs, "no member named '"+lhs.Name+"' for struct")
	}
	v = v.FieldByIndex(field.Index)
	// From reflect CanSet:
	// A Value can be changed only if it is addressable and was not obtained by the use of unexported struct fields.
	// Often a struct has to be passed as a pointer to be set
	if !v.CanSet() {
		return nilValueL, newStringError(lhs, "struct member '"+lhs.Name+"' cannot be assigned")
	}

	rv, err = convertReflectValueToType(vmp, rv, v.Type())
	if err != nil {
		return nilValueL, newError(lhs, NewTypeCannotBeAssignedError(rv.Type().String(), v.Type().String(), "struct"))
	}

	v.Set(rv)
	return v, nil
}

func invokeLetMemberMapExpr(vmp *VmParams, env envPkg.IEnv, stmt *ast.LetsStmt, v, rv reflect.Value, lhs *ast.MemberExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	if v.Type().Elem() != interfaceType && v.Type().Elem() != rv.Type() {
		rv, err = convertReflectValueToType(vmp, rv, v.Type().Elem())
		if err != nil {
			return nilValueL, newError(lhs, NewTypeCannotBeAssignedError(rv.Type().String(), v.Type().Elem().String(), "map"))
		}
	}
	if v.IsNil() {
		v = reflect.MakeMap(v.Type())
		setMapIndex(v, reflect.ValueOf(lhs.Name), rv, vmp)
		return invokeLetExpr(vmp, env, stmt, lhs.Expr, v)
	}
	setMapIndex(v, reflect.ValueOf(lhs.Name), rv, vmp)
	return v, nil
}

func invokeLetItemExpr(vmp *VmParams, rv reflect.Value, env envPkg.IEnv, stmt *ast.LetsStmt, lhs *ast.ItemExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, lhs.Value)
	if err != nil {
		return nilValueL, newError(lhs, err)
	}
	index, err := invokeExpr(vmp, env, lhs.Index)
	if err != nil {
		return nilValueL, newError(lhs, err)
	}
	v = elemIfInterface(v)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return invokeLetItemSliceExpr(vmp, env, stmt, rv, v, index, lhs)
	case reflect.Map:
		return invokeLetItemMapExpr(vmp, env, stmt, rv, v, index, lhs)
	case reflect.String:
		return invokeLetItemStringExpr(vmp, env, stmt, rv, v, index, lhs)
	default:
		return nilValueL, newStringError(lhs, "type "+v.Kind().String()+" does not support index operation")
	}
}

func invokeLetItemSliceExpr(vmp *VmParams, env envPkg.IEnv, stmt *ast.LetsStmt, rv, v, index reflect.Value, lhs *ast.ItemExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	indexInt, err := tryToInt(index)
	if err != nil {
		return nilValueL, newError(lhs, ErrIndexMustBeNumber)
	}

	if indexInt == v.Len() {
		// try to do automatic append
		if v.Type().Elem() == rv.Type() {
			v = reflect.Append(v, rv)
			return invokeLetExpr(vmp, env, stmt, lhs.Value, v)
		}
		if rv.Type().ConvertibleTo(v.Type().Elem()) {
			v = reflect.Append(v, rv.Convert(v.Type().Elem()))
			return invokeLetExpr(vmp, env, stmt, lhs.Value, v)
		}
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nilValueL, newError(lhs, NewTypeCannotBeAssignedError(rv.Type().String(), v.Type().Elem().String(), "array index"))
		}

		newSlice := reflect.MakeSlice(v.Type().Elem(), 0, rv.Len())
		newSlice, err = appendSlice(lhs, newSlice, rv)
		if err != nil {
			return nilValueL, err
		}
		v = reflect.Append(v, newSlice)
		return invokeLetExpr(vmp, env, stmt, lhs.Value, v)
	}

	if indexInt < 0 || indexInt >= v.Len() {
		return nilValueL, newError(lhs, ErrIndexOutOfRange)
	}
	v = v.Index(indexInt)
	if !v.CanSet() {
		return nilValueL, newStringError(lhs, "index cannot be assigned")
	}

	if v.Type() == rv.Type() {
		v.Set(rv)
		return v, nil
	}
	if rv.Type().ConvertibleTo(v.Type()) {
		v.Set(rv.Convert(v.Type()))
		return v, nil
	}

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nilValueL, newError(lhs, NewTypeCannotBeAssignedError(rv.Type().String(), v.Type().String(), "array index"))
	}

	newSlice := reflect.MakeSlice(v.Type(), 0, rv.Len())
	newSlice, err = appendSlice(lhs, newSlice, rv)
	if err != nil {
		return nilValueL, err
	}
	v.Set(newSlice)
	return v, nil
}

func invokeLetItemMapExpr(vmp *VmParams, env envPkg.IEnv, stmt *ast.LetsStmt, rv, v, index reflect.Value, lhs *ast.ItemExpr) (vv reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = newStringError(lhs, fmt.Sprintf("%v", r))
		}
	}()
	var errr error
	if v.Type().Key() != interfaceType && v.Type().Key() != index.Type() {
		index, errr = convertReflectValueToType(vmp, index, v.Type().Key())
		if errr != nil {
			vv = nilValue
			err = newStringError(lhs, "index type "+index.Type().String()+" cannot be used for map index type "+v.Type().Key().String())
			return
		}
	}
	if v.Type().Elem() != interfaceType && v.Type().Elem() != rv.Type() {
		rv, errr = convertReflectValueToType(vmp, rv, v.Type().Elem())
		if errr != nil {
			vv = nilValue
			err = newError(lhs, NewTypeCannotBeAssignedError(rv.Type().String(), v.Type().Elem().String(), "map"))
			return
		}
	}

	if v.IsNil() {
		v = reflect.MakeMap(v.Type())
		setMapIndex(v, index, rv, vmp)
		vv, err = invokeLetExpr(vmp, env, stmt, lhs.Value, v)
		return
	}
	setMapIndex(v, index, rv, vmp)
	vv = v
	return
}

func invokeLetItemStringExpr(vmp *VmParams, env envPkg.IEnv, stmt *ast.LetsStmt, rv, v, index reflect.Value, lhs *ast.ItemExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	rv, err = convertReflectValueToType(vmp, rv, v.Type())
	if err != nil {
		return nilValueL, newError(lhs, NewTypeCannotBeAssignedError(rv.Type().String(), v.Type().String(), ""))
	}

	indexInt, err := tryToInt(index)
	if err != nil {
		return nilValueL, newError(lhs, ErrIndexMustBeNumber)
	}

	if indexInt == v.Len() {
		// try to do automatic append

		if v.CanSet() {
			v.SetString(v.String() + rv.String())
			return v, nil
		}

		return invokeLetExpr(vmp, env, stmt, lhs.Value, reflect.ValueOf(v.String()+rv.String()))
	}

	if indexInt < 0 || indexInt >= v.Len() {
		return nilValueL, newError(lhs, ErrIndexOutOfRange)
	}

	if v.CanSet() {
		v.SetString(v.Slice(0, indexInt).String() + rv.String() + v.Slice(indexInt+1, v.Len()).String())
		return v, nil
	}

	return invokeLetExpr(vmp, env, stmt, lhs.Value, reflect.ValueOf(v.Slice(0, indexInt).String()+rv.String()+v.Slice(indexInt+1, v.Len()).String()))
}

func invokeLetSliceExpr(vmp *VmParams, env envPkg.IEnv, rv reflect.Value, lhs *ast.SliceExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, lhs.Value)
	if err != nil {
		return nilValueL, newError(lhs, err)
	}
	v = elemIfInterface(v)
	switch v.Kind() {

	// Slice && Array
	case reflect.Slice, reflect.Array:
		v, err = sliceExpr(vmp, env, v, lhs)
		if err != nil {
			return nilValueL, err
		}
		if !v.CanSet() {
			return nilValueL, newStringError(lhs, "slice cannot be assigned")
		}
		v.Set(rv)

	// String
	case reflect.String:
		return nilValueL, newStringError(lhs, "type string does not support slice operation for assignment")

	default:
		return nilValueL, newStringError(lhs, "type "+v.Kind().String()+" does not support slice operation")
	}
	return v, nil
}

func invokeLetDerefExpr(vmp *VmParams, env envPkg.IEnv, rv reflect.Value, stmt *ast.LetsStmt, lhs *ast.DerefExpr) (vv reflect.Value, err error) {
	v, err := invokeExpr(vmp, env, lhs.Expr)
	if err != nil {
		return nilValue, newError(lhs, err)
	}
	v.Elem().Set(rv)
	return v, nil
}
