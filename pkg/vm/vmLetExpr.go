package vm

import (
	"context"
	"fmt"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"reflect"
	"strings"

	"github.com/alaingilbert/anko/pkg/ast"
)

func invokeLetExpr(vmp *vmParams, env envPkg.IEnv, expr ast.Expr, rv reflect.Value) (reflect.Value, error) {
	switch lhs := expr.(type) {
	case *ast.IdentExpr:
		return invokeLetIdentExpr(expr, rv, env, lhs)
	case *ast.MemberExpr:
		return invokeLetMemberExpr(vmp, rv, env, expr, lhs)
	case *ast.ItemExpr:
		return invokeLetItemExpr(vmp, rv, env, expr, lhs)
	case *ast.SliceExpr:
		return invokeLetSliceExpr(vmp, rv, env, expr, lhs)
	case *ast.DerefExpr:
		return invokeLetDerefExpr(vmp, rv, env, expr, lhs)
	}
	return nilValue, newStringError(expr, "invalid operation")
}

func invokeLetIdentExpr(expr ast.Expr, rv reflect.Value, env envPkg.IEnv, lhs *ast.IdentExpr) (vv reflect.Value, err error) {
	if env.SetValue(lhs.Lit, rv) != nil {
		if strings.Contains(lhs.Lit, ".") {
			return nilValue, newErrorf(expr, "undefined symbol '%s'", lhs.Lit)
		}
		_ = env.DefineValue(lhs.Lit, rv)
	}
	return rv, nil
}

func invokeLetMemberExpr(vmp *vmParams, rv reflect.Value, env envPkg.IEnv, expr ast.Expr, lhs *ast.MemberExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, lhs.Expr)
	if err != nil {
		return nilValueL, newError(expr, err)
	}

	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nilValueL, newStringError(expr, "type invalid does not support member operation")
	}
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nilValueL, newStringError(expr, "type invalid does not support member operation")
	}

	switch v.Kind() {
	case reflect.Struct:
		return invokeLetMemberStructExpr(vmp.ctx, expr, v, rv, lhs)
	case reflect.Map:
		return invokeLetMemberMapExpr(vmp, expr, v, rv, env, lhs)
	default:
		return nilValueL, newStringError(expr, "type "+v.Kind().String()+" does not support member operation")
	}
}

func invokeLetMemberStructExpr(ctx context.Context, expr ast.Expr, v, rv reflect.Value, lhs *ast.MemberExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	field, found := v.Type().FieldByName(lhs.Name)
	if !found {
		return nilValueL, newStringError(expr, "no member named '"+lhs.Name+"' for struct")
	}
	v = v.FieldByIndex(field.Index)
	// From reflect CanSet:
	// A Value can be changed only if it is addressable and was not obtained by the use of unexported struct fields.
	// Often a struct has to be passed as a pointer to be set
	if !v.CanSet() {
		return nilValueL, newStringError(expr, "struct member '"+lhs.Name+"' cannot be assigned")
	}

	rv, err = convertReflectValueToType(ctx, rv, v.Type())
	if err != nil {
		return nilValueL, newStringError(expr, "type "+rv.Type().String()+" cannot be assigned to type "+v.Type().String()+" for struct")
	}

	v.Set(rv)
	return v, nil
}

func invokeLetMemberMapExpr(vmp *vmParams, expr ast.Expr, v, rv reflect.Value, env envPkg.IEnv, lhs *ast.MemberExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	if v.Type().Elem() != interfaceType && v.Type().Elem() != rv.Type() {
		rv, err = convertReflectValueToType(vmp.ctx, rv, v.Type().Elem())
		if err != nil {
			return nilValueL, newStringError(expr, "type "+rv.Type().String()+" cannot be assigned to type "+v.Type().Elem().String()+" for map")
		}
	}
	if v.IsNil() {
		v = reflect.MakeMap(v.Type())
		vmp.mapMutex.Lock()
		defer vmp.mapMutex.Unlock()
		v.SetMapIndex(reflect.ValueOf(lhs.Name), rv)
		return invokeLetExpr(vmp, env, lhs.Expr, v)
	}
	vmp.mapMutex.Lock()
	defer vmp.mapMutex.Unlock()
	v.SetMapIndex(reflect.ValueOf(lhs.Name), rv)
	return v, nil
}

func invokeLetItemExpr(vmp *vmParams, rv reflect.Value, env envPkg.IEnv, expr ast.Expr, lhs *ast.ItemExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, lhs.Value)
	if err != nil {
		return nilValueL, newError(expr, err)
	}
	index, err := invokeExpr(vmp, env, lhs.Index)
	if err != nil {
		return nilValueL, newError(expr, err)
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return invokeLetItemSliceExpr(vmp, expr, rv, v, index, env, lhs)
	case reflect.Map:
		return invokeLetItemMapExpr(vmp, expr, rv, v, index, env, lhs)
	case reflect.String:
		return invokeLetItemStringExpr(vmp, expr, rv, v, index, env, lhs)
	default:
		return nilValueL, newStringError(expr, "type "+v.Kind().String()+" does not support index operation")
	}
}

func invokeLetItemSliceExpr(vmp *vmParams, expr ast.Expr, rv, v, index reflect.Value, env envPkg.IEnv, lhs *ast.ItemExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	indexInt, err := tryToInt(index)
	if err != nil {
		return nilValueL, newStringError(expr, "index must be a number")
	}

	if indexInt == v.Len() {
		// try to do automatic append
		if v.Type().Elem() == rv.Type() {
			v = reflect.Append(v, rv)
			return invokeLetExpr(vmp, env, lhs.Value, v)
		}
		if rv.Type().ConvertibleTo(v.Type().Elem()) {
			v = reflect.Append(v, rv.Convert(v.Type().Elem()))
			return invokeLetExpr(vmp, env, lhs.Value, v)
		}
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nilValueL, newStringError(expr, "type "+rv.Type().String()+" cannot be assigned to type "+v.Type().Elem().String()+" for array index")
		}

		newSlice := reflect.MakeSlice(v.Type().Elem(), 0, rv.Len())
		newSlice, err = appendSlice(expr, newSlice, rv)
		if err != nil {
			return nilValueL, err
		}
		v = reflect.Append(v, newSlice)
		return invokeLetExpr(vmp, env, lhs.Value, v)
	}

	if indexInt < 0 || indexInt >= v.Len() {
		return nilValueL, newStringError(expr, "index out of range")
	}
	v = v.Index(indexInt)
	if !v.CanSet() {
		return nilValueL, newStringError(expr, "index cannot be assigned")
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
		return nilValueL, newStringError(expr, "type "+rv.Type().String()+" cannot be assigned to type "+v.Type().String()+" for array index")
	}

	newSlice := reflect.MakeSlice(v.Type(), 0, rv.Len())
	newSlice, err = appendSlice(expr, newSlice, rv)
	if err != nil {
		return nilValueL, err
	}
	v.Set(newSlice)
	return v, nil
}

func invokeLetItemMapExpr(vmp *vmParams, expr ast.Expr, rv, v, index reflect.Value, env envPkg.IEnv, lhs *ast.ItemExpr) (vv reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = newStringError(expr, fmt.Sprintf("%v", r))
		}
	}()
	var errr error
	if v.Type().Key() != interfaceType && v.Type().Key() != index.Type() {
		index, errr = convertReflectValueToType(vmp.ctx, index, v.Type().Key())
		if errr != nil {
			vv = nilValue
			err = newStringError(expr, "index type "+index.Type().String()+" cannot be used for map index type "+v.Type().Key().String())
			return
		}
	}
	if v.Type().Elem() != interfaceType && v.Type().Elem() != rv.Type() {
		rv, errr = convertReflectValueToType(vmp.ctx, rv, v.Type().Elem())
		if errr != nil {
			vv = nilValue
			err = newStringError(expr, "type "+rv.Type().String()+" cannot be assigned to type "+v.Type().Elem().String()+" for map")
			return
		}
	}

	if v.IsNil() {
		v = reflect.MakeMap(v.Type())
		vmp.mapMutex.Lock()
		defer vmp.mapMutex.Unlock()
		v.SetMapIndex(index, rv)
		vv, err = invokeLetExpr(vmp, env, lhs.Value, v)
		return
	}
	vmp.mapMutex.Lock()
	defer vmp.mapMutex.Unlock()
	v.SetMapIndex(index, rv)
	vv = v
	return
}

func invokeLetItemStringExpr(vmp *vmParams, expr ast.Expr, rv, v, index reflect.Value, env envPkg.IEnv, lhs *ast.ItemExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	rv, err = convertReflectValueToType(vmp.ctx, rv, v.Type())
	if err != nil {
		return nilValueL, newStringError(expr, "type "+rv.Type().String()+" cannot be assigned to type "+v.Type().String())
	}

	indexInt, err := tryToInt(index)
	if err != nil {
		return nilValueL, newStringError(expr, "index must be a number")
	}

	if indexInt == v.Len() {
		// try to do automatic append

		if v.CanSet() {
			v.SetString(v.String() + rv.String())
			return v, nil
		}

		return invokeLetExpr(vmp, env, lhs.Value, reflect.ValueOf(v.String()+rv.String()))
	}

	if indexInt < 0 || indexInt >= v.Len() {
		return nilValueL, newStringError(expr, "index out of range")
	}

	if v.CanSet() {
		v.SetString(v.Slice(0, indexInt).String() + rv.String() + v.Slice(indexInt+1, v.Len()).String())
		return v, nil
	}

	return invokeLetExpr(vmp, env, lhs.Value, reflect.ValueOf(v.Slice(0, indexInt).String()+rv.String()+v.Slice(indexInt+1, v.Len()).String()))
}

func invokeLetSliceExpr(vmp *vmParams, rv reflect.Value, env envPkg.IEnv, expr ast.Expr, lhs *ast.SliceExpr) (vv reflect.Value, err error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, lhs.Value)
	if err != nil {
		return nilValueL, newError(expr, err)
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {

	// Slice && Array
	case reflect.Slice, reflect.Array:
		var rbi, rei int
		if lhs.Begin != nil {
			rb, err := invokeExpr(vmp, env, lhs.Begin)
			if err != nil {
				return nilValueL, newError(expr, err)
			}
			rbi, err = tryToInt(rb)
			if err != nil {
				return nilValueL, newStringError(expr, "index must be a number")
			}
			if rbi < 0 || rbi > v.Len() {
				return nilValueL, newStringError(expr, "index out of range")
			}
		} else {
			rbi = 0
		}
		if lhs.End != nil {
			re, err := invokeExpr(vmp, env, lhs.End)
			if err != nil {
				return nilValueL, newError(expr, err)
			}
			rei, err = tryToInt(re)
			if err != nil {
				return nilValueL, newStringError(expr, "index must be a number")
			}
			if rei < 0 || rei > v.Len() {
				return nilValueL, newStringError(expr, "index out of range")
			}
		} else {
			rei = v.Len()
		}
		if rbi > rei {
			return nilValueL, newStringError(expr, "invalid slice index")
		}
		v = v.Slice(rbi, rei)
		if !v.CanSet() {
			return nilValueL, newStringError(expr, "slice cannot be assigned")
		}
		v.Set(rv)

	// String
	case reflect.String:
		return nilValueL, newStringError(expr, "type string does not support slice operation for assignment")

	default:
		return nilValueL, newStringError(expr, "type "+v.Kind().String()+" does not support slice operation")
	}
	return v, nil
}

func invokeLetDerefExpr(vmp *vmParams, rv reflect.Value, env envPkg.IEnv, expr ast.Expr, lhs *ast.DerefExpr) (vv reflect.Value, err error) {
	v, err := invokeExpr(vmp, env, lhs.Expr)
	if err != nil {
		return nilValue, newError(expr, err)
	}
	v.Elem().Set(rv)
	return v, nil
}
