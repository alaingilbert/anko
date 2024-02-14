package vm

import (
	"context"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// invokeExpr evaluates one expression.
func invokeExpr(vmp *vmParams, env *envPkg.Env, expr ast.Expr) (reflect.Value, error) {
	if err := incrCycle(vmp); err != nil {
		return nilValue, ErrInterrupt
	}
	//fmt.Println("invokeExpr", reflect.ValueOf(expr).String())

	switch e := expr.(type) {
	case *ast.NumberExpr:
		return invokeNumberExpr(vmp.ctx, env, e)
	case *ast.IdentExpr:
		return invokeIdentExpr(vmp.ctx, env, e)
	case *ast.StringExpr:
		return invokeStringExpr(vmp.ctx, env, e)
	case *ast.ArrayExpr:
		return invokeArrayExpr(vmp, env, e)
	case *ast.MapExpr:
		return invokeMapExpr(vmp, env, e)
	case *ast.DerefExpr:
		return invokeDerefExpr(vmp, env, e)
	case *ast.AddrExpr:
		return invokeAddrExpr(vmp, env, e)
	case *ast.UnaryExpr:
		return invokeUnaryExpr(vmp, env, e)
	case *ast.ParenExpr:
		return invokeParenExpr(vmp, env, e)
	case *ast.MemberExpr:
		return invokeMemberExpr(vmp, env, e)
	case *ast.ItemExpr:
		return invokeItemExpr(vmp, env, e)
	case *ast.SliceExpr:
		return invokeSliceExpr(vmp, env, e)
	case *ast.AssocExpr:
		return invokeAssocExpr(vmp, env, e)
	case *ast.LetsExpr:
		return invokeLetsExpr(vmp, env, e)
	case *ast.BinOpExpr:
		return invokeBinOpExpr(vmp, e, env, expr)
	case *ast.ConstExpr:
		return invokeConstExpr(vmp.ctx, env, e)
	case *ast.TernaryOpExpr:
		return invokeTernaryOpExpr(vmp, env, e)
	case *ast.NilCoalescingOpExpr:
		return invokeNilCoalescingOpExpr(vmp, env, e)
	case *ast.LenExpr:
		return invokeLenExpr(vmp, env, e)
	case *ast.NewExpr:
		return invokeNewExpr(vmp.ctx, e, env, expr)
	case *ast.MakeExpr:
		return invokeMakeExpr(vmp, e, env, expr)
	case *ast.MakeTypeExpr:
		return invokeMakeTypeExpr(vmp, e, env, expr)
	case *ast.MakeChanExpr:
		return invokeMakeChanExpr(vmp, e, env, expr)
	case *ast.ChanExpr:
		return invokeChanExpr(vmp, e, env, expr)
	case *ast.FuncExpr:
		return invokeFuncExpr(vmp, env, e)
	case *ast.AnonCallExpr:
		return invokeAnonCallExpr(vmp, env, e)
	case *ast.CallExpr:
		return invokeCallExpr(vmp, env, e)
	case *ast.CloseExpr:
		return invokeCloseExpr(vmp, env, e)
	case *ast.DeleteExpr:
		return invokeDeleteExpr(vmp, env, e)
	case *ast.IncludeExpr:
		return invokeIncludeExpr(vmp, env, e)
	default:
		return nilValue, newStringError(e, "unknown expression")
	}
}

func invokeNumberExpr(ctx context.Context, env *envPkg.Env, e *ast.NumberExpr) (reflect.Value, error) {
	nilValueL := nilValue
	if strings.Contains(e.Lit, ".") || strings.Contains(e.Lit, "e") {
		v, err := strconv.ParseFloat(e.Lit, 64)
		if err != nil {
			return nilValueL, newError(e, err)
		}
		return reflect.ValueOf(float64(v)), nil
	}
	var i int64
	var err error
	if strings.HasPrefix(e.Lit, "0x") {
		i, err = strconv.ParseInt(e.Lit[2:], 16, 64)
	} else {
		i, err = strconv.ParseInt(e.Lit, 10, 64)
	}
	if err != nil {
		return nilValueL, newError(e, err)
	}
	return reflect.ValueOf(i), nil
}

func invokeIdentExpr(ctx context.Context, env *envPkg.Env, e *ast.IdentExpr) (reflect.Value, error) {
	return env.GetValue(e.Lit)
}

func invokeStringExpr(ctx context.Context, env *envPkg.Env, e *ast.StringExpr) (reflect.Value, error) {
	return reflect.ValueOf(e.Lit), nil
}

func makeType(vmp *vmParams, env *envPkg.Env, typeStruct *ast.TypeStruct) (reflect.Type, error) {
	switch typeStruct.Kind {
	case ast.TypeDefault:
		return getTypeFromString(env, typeStruct.Name)
	case ast.TypePtr:
		var t reflect.Type
		var err error
		if typeStruct.SubType != nil {
			t, err = makeType(vmp, env, typeStruct.SubType)
		} else {
			t, err = getTypeFromString(env, typeStruct.Name)
		}
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, err
		}
		return reflect.PtrTo(t), nil
	case ast.TypeSlice:
		var t reflect.Type
		var err error
		if typeStruct.SubType != nil {
			t, err = makeType(vmp, env, typeStruct.SubType)
		} else {
			t, err = getTypeFromString(env, typeStruct.Name)
		}
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, err
		}
		for i := 1; i < typeStruct.Dimensions; i++ {
			t = reflect.SliceOf(t)
		}
		return reflect.SliceOf(t), nil
	//case ast.TypeMap:
	//	key := makeType(runInfo, typeStruct.Key)
	//	if runInfo.err != nil {
	//		return nil
	//	}
	//	if key == nil {
	//		return nil
	//	}
	//	t := makeType(runInfo, typeStruct.SubType)
	//	if runInfo.err != nil {
	//		return nil
	//	}
	//	if t == nil {
	//		return nil
	//	}
	//	if !runInfo.options.Debug {
	//		// captures panic
	//		defer recoverFunc(runInfo)
	//	}
	//	t = reflect.MapOf(key, t)
	//	return t
	//case ast.TypeChan:
	//	var t reflect.Type
	//	if typeStruct.SubType != nil {
	//		t = makeType(runInfo, typeStruct.SubType)
	//	} else {
	//		t = getTypeFromEnv(runInfo, typeStruct)
	//	}
	//	if runInfo.err != nil {
	//		return nil
	//	}
	//	if t == nil {
	//		return nil
	//	}
	//	return reflect.ChanOf(reflect.BothDir, t)
	//case ast.TypeStructType:
	//	var t reflect.Type
	//	fields := make([]reflect.StructField, 0, len(typeStruct.StructNames))
	//	for i := 0; i < len(typeStruct.StructNames); i++ {
	//		t = makeType(runInfo, typeStruct.StructTypes[i])
	//		if runInfo.err != nil {
	//			return nil
	//		}
	//		if t == nil {
	//			return nil
	//		}
	//		fields = append(fields, reflect.StructField{Name: typeStruct.StructNames[i], Type: t})
	//	}
	//	if !runInfo.options.Debug {
	//		// captures panic
	//		defer recoverFunc(runInfo)
	//	}
	//	t = reflect.StructOf(fields)
	//	return t
	default:
		return nil, fmt.Errorf("unknown kind")
	}
}

func invokeArrayExpr(vmp *vmParams, env *envPkg.Env, e *ast.ArrayExpr) (reflect.Value, error) {
	if e.TypeData == nil {
		a := make([]any, len(e.Exprs))
		for i, expr := range e.Exprs {
			arg, err := invokeExpr(vmp, env, expr)
			if err != nil {
				return nilValue, newError(expr, err)
			}
			a[i] = arg.Interface()
		}
		return reflect.ValueOf(a), nil
	}

	t, err := makeType(vmp, env, e.TypeData)
	if err != nil {
		return nilValue, err
	}
	if t == nil {
		return nilValue, newStringError(e, "cannot make type nil")
	}

	slice := reflect.MakeSlice(t, len(e.Exprs), len(e.Exprs))
	valueType := t.Elem()
	for i, ee := range e.Exprs {
		rv, err := invokeExpr(vmp, env, ee)
		if err != nil {
			return nilValue, err
		}
		rv, err = convertReflectValueToType(vmp.ctx, rv, valueType)
		if err != nil {
			return nilValue, newStringError(e, "cannot use type "+rv.Type().String()+" as type "+valueType.String()+" as slice value")
		}
		slice.Index(i).Set(rv)
	}
	return slice, nil
}

func invokeMapExpr(vmp *vmParams, env *envPkg.Env, e *ast.MapExpr) (reflect.Value, error) {
	nilValueL := nilValue
	var err error
	var key reflect.Value
	var value reflect.Value
	m := make(map[any]any, len(e.MapExpr))
	for keyExpr, valueExpr := range e.MapExpr {
		key, err = invokeExpr(vmp, env, keyExpr)
		if err != nil {
			return nilValueL, newError(keyExpr, err)
		}
		value, err = invokeExpr(vmp, env, valueExpr)
		if err != nil {
			return nilValueL, newError(valueExpr, err)
		}
		m[key.Interface()] = value.Interface()
	}
	return reflect.ValueOf(m), nil
}

func invokeDerefExpr(vmp *vmParams, env *envPkg.Env, e *ast.DerefExpr) (reflect.Value, error) {
	v := nilValue
	switch ee := e.Expr.(type) {
	case *ast.IdentExpr:
		return invokeDeferExprIdentExpr(v, env, e, ee)
	case *ast.MemberExpr:
		return invokeDeferExprMemberExpr(vmp, env, e, ee)
	default:
		return nilValue, newStringError(e, "invalid operation for the value")
	}
}

func invokeDeferExprIdentExpr(v reflect.Value, env *envPkg.Env, e *ast.DerefExpr, ee *ast.IdentExpr) (reflect.Value, error) {
	var err error
	v, err = env.GetValue(ee.Lit)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if v.Kind() != reflect.Pointer {
		return nilValue, newStringError(e, "cannot deference for the value")
	}
	return v.Elem(), nil
}

func invokeDeferExprMemberExpr(vmp *vmParams, env *envPkg.Env, e *ast.DerefExpr, ee *ast.MemberExpr) (reflect.Value, error) {
	v, err := invokeExpr(vmp, env, ee.Expr)
	if err != nil {
		return nilValue, newError(ee.Expr, err)
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice {
		v = v.Index(0)
	}
	if v.IsValid() && v.CanInterface() {
		if vme, ok := v.Interface().(*envPkg.Env); ok {
			m, err := vme.GetValue(ee.Name)
			if !m.IsValid() || err != nil {
				return nilValue, newStringError(e, fmt.Sprintf("invalid operation '%s'", ee.Name))
			}
			return m, nil
		}
	}

	m := v.MethodByName(ee.Name)
	if !m.IsValid() {
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			field, found := v.Type().FieldByName(ee.Name)
			if !found {
				return nilValue, newStringError(e, "no member named '"+ee.Name+"' for struct")
			}
			return v.FieldByIndex(field.Index), nil
		} else if v.Kind() == reflect.Map {
			// From reflect MapIndex:
			// It returns the zero Value if key is not found in the map or if v represents a nil map.
			vmp.mapMutex.Lock()
			m = v.MapIndex(reflect.ValueOf(ee.Name))
			vmp.mapMutex.Unlock()
		} else {
			return nilValue, newStringError(e, fmt.Sprintf("invalid operation '%s'", ee.Name))
		}
		v = m
	} else {
		v = m
	}
	if v.Kind() != reflect.Pointer {
		return nilValue, newStringError(e, "cannot deference for the value")
	}
	return v.Elem(), nil
}

func invokeAddrExpr(vmp *vmParams, env *envPkg.Env, e *ast.AddrExpr) (reflect.Value, error) {
	nilValueL := nilValue
	switch ee := e.Expr.(type) {
	case *ast.IdentExpr:
		return invokeAddrExprIdentExpr(env, e, ee)
	case *ast.MemberExpr:
		return invokeAddrExprMemberExpr(vmp, env, e, ee)
	default:
		return nilValueL, newStringError(e, "invalid operation for the value")
	}
}

func invokeAddrExprIdentExpr(env *envPkg.Env, e *ast.AddrExpr, ee *ast.IdentExpr) (reflect.Value, error) {
	v, err := env.GetValue(ee.Lit)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if !v.CanAddr() {
		i := v.Interface()
		return reflect.ValueOf(&i), nil
	}
	return v.Addr(), nil
}

func invokeAddrExprMemberExpr(vmp *vmParams, env *envPkg.Env, e *ast.AddrExpr, ee *ast.MemberExpr) (reflect.Value, error) {
	v, err := invokeExpr(vmp, env, ee.Expr)
	if err != nil {
		return nilValue, newError(ee.Expr, err)
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice {
		v = v.Index(0)
	}
	if v.IsValid() && v.CanInterface() {
		if vme, ok := v.Interface().(*envPkg.Env); ok {
			m, err := vme.GetValue(ee.Name)
			if !m.IsValid() || err != nil {
				return nilValue, newStringError(e, fmt.Sprintf("invalid operation '%s'", ee.Name))
			}
			return m, nil
		}
	}

	m := v.MethodByName(ee.Name)
	if !m.IsValid() {
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			m = v.FieldByName(ee.Name)
			if !m.IsValid() {
				return nilValue, newStringError(e, fmt.Sprintf("invalid operation '%s'", ee.Name))
			}
		} else if v.Kind() == reflect.Map {
			// From reflect MapIndex:
			// It returns the zero Value if key is not found in the map or if v represents a nil map.
			vmp.mapMutex.Lock()
			m = v.MapIndex(reflect.ValueOf(ee.Name))
			vmp.mapMutex.Unlock()
		} else {
			return nilValue, newStringError(e, fmt.Sprintf("invalid operation '%s'", ee.Name))
		}
		v = m
	} else {
		v = m
	}
	if !v.CanAddr() {
		i := v.Interface()
		return reflect.ValueOf(&i), nil
	}
	return v.Addr(), nil
}

func invokeUnaryExpr(vmp *vmParams, env *envPkg.Env, e *ast.UnaryExpr) (reflect.Value, error) {
	v, err := invokeExpr(vmp, env, e.Expr)
	if err != nil {
		return nilValue, newError(e.Expr, err)
	}
	switch e.Operator {
	case "-":
		if v.Kind() == reflect.Int64 {
			return reflect.ValueOf(-v.Int()), nil
		}
		if v.Kind() == reflect.Float64 {
			return reflect.ValueOf(-v.Float()), nil
		}
		return reflect.ValueOf(-toFloat64(v)), nil
	case "^":
		return reflect.ValueOf(^toInt64(v)), nil
	case "!":
		return reflect.ValueOf(!toBool(v)), nil
	default:
		return nilValue, newStringError(e, "unknown operator ''")
	}
}
func invokeParenExpr(vmp *vmParams, env *envPkg.Env, e *ast.ParenExpr) (reflect.Value, error) {
	v, err := invokeExpr(vmp, env, e.SubExpr)
	if err != nil {
		return nilValue, newError(e.SubExpr, err)
	}
	return v, nil
}

func invokeMemberExpr(vmp *vmParams, env *envPkg.Env, e *ast.MemberExpr) (reflect.Value, error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, e.Expr)
	if err != nil {
		return nilValueL, newError(e.Expr, err)
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nilValueL, newStringError(e, "type invalid does not support member operation")
	}
	if v.CanInterface() {
		if vme, ok := v.Interface().(*envPkg.Env); ok {
			m, err := vme.GetValue(e.Name)
			if !m.IsValid() || err != nil {
				return nilValueL, newStringError(e, fmt.Sprintf("invalid operation '%s'", e.Name))
			}
			return m, nil
		}
	}

	method, found := v.Type().MethodByName(e.Name)
	if found {
		return v.Method(method.Index), nil
	}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		field, found := v.Type().FieldByName(e.Name)
		if found {
			return v.FieldByIndex(field.Index), nil
		}
		if v.CanAddr() {
			v = v.Addr()
			method, found := v.Type().MethodByName(e.Name)
			if found {
				return v.Method(method.Index), nil
			}
		} else {
			// Check if method with pointer receiver is defined,
			// if yes, invoke it in the copied instance
			method, found := reflect.PtrTo(v.Type()).MethodByName(e.Name)
			if found {
				// Create pointer value to given struct type which were passed by value
				cv := reflect.New(v.Type())
				cv.Elem().Set(v)
				return cv.Method(method.Index), nil
			}
		}
		return nilValueL, newStringError(e, "no member named '"+e.Name+"' for struct")
	case reflect.Map:
		v = getMapIndex(vmp, reflect.ValueOf(e.Name), v)
		// Note if the map is of reflect.Value, it will incorrectly return nil when zero value
		if v == zeroValue {
			return nilValueL, nil
		}
		return v, nil
	default:
		return nilValueL, newStringError(e, "type "+v.Kind().String()+" does not support member operation")
	}
}

func invokeItemExpr(vmp *vmParams, env *envPkg.Env, e *ast.ItemExpr) (reflect.Value, error) {
	v, err := invokeExpr(vmp, env, e.Value)
	if err != nil {
		return nilValue, newError(e.Value, err)
	}
	i, err := invokeExpr(vmp, env, e.Index)
	if err != nil {
		return nilValue, newError(e.Index, err)
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		ii, err := tryToInt(i)
		if err != nil {
			return nilValue, newStringError(e, "index must be a number")
		}
		if !vmp.validate {
			if ii < 0 || ii >= v.Len() {
				return nilValue, newStringError(e, "index out of range")
			}
		}
		if v.Kind() != reflect.String {
			if vmp.validate {
				return zeroOfType(v.Type().Elem()), nil
			}
			return v.Index(ii), nil
		}
		// String
		return v.Index(ii).Convert(stringType), nil
	case reflect.Map:
		v = getMapIndex(vmp, i, v)
		// Note if the map is of reflect.Value, it will incorrectly return nil when zero value
		if v == zeroValue {
			return nilValue, nil
		}
		return v, nil
	default:
		return nilValue, newStringError(e, "type "+v.Kind().String()+" does not support index operation")
	}
}

type TestStruct struct {
}

func (t TestStruct) TestMember() {
}

type TestInterface interface {
	TestMember()
}

func zeroOfType(t reflect.Type) (v reflect.Value) {
	v = reflect.Zero(t)
	if t.Kind() == reflect.Interface {
		yType := reflect.TypeOf((*TestInterface)(nil)).Elem()
		if t.Implements(yType) {
			v = reflect.ValueOf(TestStruct{})
		}
	}
	return
}

func invokeSliceExpr(vmp *vmParams, env *envPkg.Env, e *ast.SliceExpr) (reflect.Value, error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, e.Value)
	if err != nil {
		return nilValueL, newError(e.Value, err)
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		var rbi, rei int
		if e.Begin != nil {
			rb, err := invokeExpr(vmp, env, e.Begin)
			if err != nil {
				return nilValueL, newError(e.Begin, err)
			}
			rbi, err = tryToInt(rb)
			if err != nil {
				return nilValueL, newStringError(e, "index must be a number")
			}
			if rbi < 0 || rbi > v.Len() {
				return nilValueL, newStringError(e, "index out of range")
			}
		} else {
			rbi = 0
		}
		if e.End != nil {
			re, err := invokeExpr(vmp, env, e.End)
			if err != nil {
				return nilValueL, newError(e.End, err)
			}
			rei, err = tryToInt(re)
			if err != nil {
				return nilValueL, newStringError(e, "index must be a number")
			}
			if rei < 0 || rei > v.Len() {
				return nilValueL, newStringError(e, "index out of range")
			}
		} else {
			rei = v.Len()
		}
		if rbi > rei {
			return nilValueL, newStringError(e, "invalid slice index")
		}
		return v.Slice(rbi, rei), nil
	default:
		return nilValueL, newStringError(e, "type "+v.Kind().String()+" does not support slice operation")
	}
}

func invokeAssocExpr(vmp *vmParams, env *envPkg.Env, e *ast.AssocExpr) (reflect.Value, error) {
	switch e.Operator {
	case "++":
		if alhs, ok := e.Lhs.(*ast.IdentExpr); ok {
			v, err := env.GetValue(alhs.Lit)
			if err != nil {
				return nilValue, newError(e, err)
			}
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				v = reflect.ValueOf(v.Float() + 1)
			case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
				v = reflect.ValueOf(v.Int() + 1)
			case reflect.Bool:
				if v.Bool() {
					v = reflect.ValueOf(int64(2))
				} else {
					v = reflect.ValueOf(int64(1))
				}
			default:
				v = reflect.ValueOf(toInt64(v) + 1)
			}
			// not checking err because checked above in get
			_ = env.SetValue(alhs.Lit, v)
			return v, nil
		}
	case "--":
		if alhs, ok := e.Lhs.(*ast.IdentExpr); ok {
			v, err := env.GetValue(alhs.Lit)
			if err != nil {
				return nilValue, newError(e, err)
			}
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				v = reflect.ValueOf(v.Float() - 1)
			case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
				v = reflect.ValueOf(v.Int() - 1)
			case reflect.Bool:
				if v.Bool() {
					v = reflect.ValueOf(int64(0))
				} else {
					v = reflect.ValueOf(int64(-1))
				}
			default:
				v = reflect.ValueOf(toInt64(v) - 1)
			}
			// not checking err because checked above in get
			_ = env.SetValue(alhs.Lit, v)
			return v, nil
		}
	}

	if e.Rhs == nil {
		// TODO: Can this be fixed in the parser so that Rhs is not nil?
		e.Rhs = &ast.NumberExpr{Lit: "1"}
	}
	v, err := invokeExpr(vmp, env, &ast.BinOpExpr{Lhs: e.Lhs, Operator: e.Operator[0:1], Rhs: e.Rhs})
	if err != nil {
		return nilValue, newError(e, err)
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return invokeLetExpr(vmp, env, e.Lhs, v)
}

func invokeLetsExpr(vmp *vmParams, env *envPkg.Env, e *ast.LetsExpr) (reflect.Value, error) {
	var err error
	rvs := make([]reflect.Value, len(e.Rhss))
	for i, rhs := range e.Rhss {
		rvs[i], err = invokeExpr(vmp, env, rhs)
		if err != nil {
			return nilValue, newError(rhs, err)
		}
	}
	for i, lhs := range e.Lhss {
		if i >= len(rvs) {
			break
		}
		v := rvs[i]
		if v.Kind() == reflect.Interface && !v.IsNil() {
			v = v.Elem()
		}
		_, err = invokeLetExpr(vmp, env, lhs, v)
		if err != nil {
			return nilValue, newError(lhs, err)
		}
	}
	return rvs[len(rvs)-1], nil
}

func invokeBinOpExpr(vmp *vmParams, e *ast.BinOpExpr, env *envPkg.Env, expr ast.Expr) (reflect.Value, error) {
	nilValueL := nilValue
	lhsV := nilValueL
	rhsV := nilValueL
	var err error

	lhsV, err = invokeExpr(vmp, env, e.Lhs)
	if err != nil {
		return nilValueL, newError(e.Lhs, err)
	}
	if lhsV.Kind() == reflect.Interface && !lhsV.IsNil() {
		lhsV = lhsV.Elem()
	}
	switch e.Operator {
	case "&&":
		if !toBool(lhsV) {
			return lhsV, nil
		}
	case "||":
		if toBool(lhsV) {
			return lhsV, nil
		}
	}
	if e.Rhs != nil {
		rhsV, err = invokeExpr(vmp, env, e.Rhs)
		if err != nil {
			return nilValueL, newError(e.Rhs, err)
		}
		if rhsV.Kind() == reflect.Interface && !rhsV.IsNil() {
			rhsV = rhsV.Elem()
		}
	}
	switch e.Operator {
	case "+":
		if (lhsV.Kind() == reflect.Slice || lhsV.Kind() == reflect.Array) && (rhsV.Kind() != reflect.Slice && rhsV.Kind() != reflect.Array) {
			rhsT := rhsV.Type()
			lhsT := lhsV.Type().Elem()
			if lhsT.Kind() != rhsT.Kind() {
				if !rhsT.ConvertibleTo(lhsT) {
					return nilValueL, newStringError(e, "invalid type conversion")
				}
				rhsV = rhsV.Convert(lhsT)
			}
			return reflect.Append(lhsV, rhsV), nil
		}
		if (lhsV.Kind() == reflect.Slice || lhsV.Kind() == reflect.Array) && (rhsV.Kind() == reflect.Slice || rhsV.Kind() == reflect.Array) {
			return appendSlice(expr, lhsV, rhsV)
		}
		if lhsV.Kind() == reflect.String || rhsV.Kind() == reflect.String {
			return reflect.ValueOf(toString(lhsV) + toString(rhsV)), nil
		}
		if lhsV.Kind() == reflect.Float64 || rhsV.Kind() == reflect.Float64 {
			return reflect.ValueOf(toFloat64(lhsV) + toFloat64(rhsV)), nil
		}
		return reflect.ValueOf(toInt64(lhsV) + toInt64(rhsV)), nil
	case "-":
		if lhsV.Kind() == reflect.Float64 || rhsV.Kind() == reflect.Float64 {
			return reflect.ValueOf(toFloat64(lhsV) - toFloat64(rhsV)), nil
		}
		return reflect.ValueOf(toInt64(lhsV) - toInt64(rhsV)), nil
	case "*":
		if lhsV.Kind() == reflect.String && (rhsV.Kind() == reflect.Int || rhsV.Kind() == reflect.Int32 || rhsV.Kind() == reflect.Int64) {
			return reflect.ValueOf(strings.Repeat(toString(lhsV), int(toInt64(rhsV)))), nil
		}
		if lhsV.Kind() == reflect.Float64 || rhsV.Kind() == reflect.Float64 {
			return reflect.ValueOf(toFloat64(lhsV) * toFloat64(rhsV)), nil
		}
		return reflect.ValueOf(toInt64(lhsV) * toInt64(rhsV)), nil
	case "/":
		return reflect.ValueOf(toFloat64(lhsV) / toFloat64(rhsV)), nil
	case "%":
		return reflect.ValueOf(toInt64(lhsV) % toInt64(rhsV)), nil
	case "==":
		return reflect.ValueOf(equal(lhsV, rhsV)), nil
	case "!=":
		return reflect.ValueOf(equal(lhsV, rhsV) == false), nil
	case ">":
		return reflect.ValueOf(toFloat64(lhsV) > toFloat64(rhsV)), nil
	case ">=":
		return reflect.ValueOf(toFloat64(lhsV) >= toFloat64(rhsV)), nil
	case "<":
		return reflect.ValueOf(toFloat64(lhsV) < toFloat64(rhsV)), nil
	case "<=":
		return reflect.ValueOf(toFloat64(lhsV) <= toFloat64(rhsV)), nil
	case "|":
		return reflect.ValueOf(toInt64(lhsV) | toInt64(rhsV)), nil
	case "||":
		return rhsV, nil
	case "&":
		return reflect.ValueOf(toInt64(lhsV) & toInt64(rhsV)), nil
	case "&&":
		return rhsV, nil
	case "**":
		if lhsV.Kind() == reflect.Float64 {
			return reflect.ValueOf(math.Pow(lhsV.Float(), toFloat64(rhsV))), nil
		}
		return reflect.ValueOf(int64(math.Pow(toFloat64(lhsV), toFloat64(rhsV)))), nil
	case ">>":
		return reflect.ValueOf(toInt64(lhsV) >> uint64(toInt64(rhsV))), nil
	case "<<":
		return reflect.ValueOf(toInt64(lhsV) << uint64(toInt64(rhsV))), nil
	default:
		return nilValueL, newStringError(e, "unknown operator")
	}
}

func invokeConstExpr(ctx context.Context, env *envPkg.Env, e *ast.ConstExpr) (reflect.Value, error) {
	switch e.Value {
	case "true":
		return trueValue, nil
	case "false":
		return falseValue, nil
	}
	return nilValue, nil
}

func invokeTernaryOpExpr(vmp *vmParams, env *envPkg.Env, e *ast.TernaryOpExpr) (reflect.Value, error) {
	rv, err := invokeExpr(vmp, env, e.Expr)
	if err != nil {
		return nilValue, newError(e.Expr, err)
	}
	if toBool(rv) {
		lhsV, err := invokeExpr(vmp, env, e.Lhs)
		if err != nil {
			return nilValue, newError(e.Lhs, err)
		}
		return lhsV, nil
	}
	rhsV, err := invokeExpr(vmp, env, e.Rhs)
	if err != nil {
		return nilValue, newError(e.Rhs, err)
	}
	return rhsV, nil
}

func invokeNilCoalescingOpExpr(vmp *vmParams, env *envPkg.Env, e *ast.NilCoalescingOpExpr) (reflect.Value, error) {
	var err error
	rv, _ := invokeExpr(vmp, env, e.Lhs)
	if toBool(rv) {
		return rv, nil
	}
	rv, err = invokeExpr(vmp, env, e.Rhs)
	if err != nil {
		return nilValue, newError(e.Rhs, err)
	}
	return rv, nil
}

func invokeLenExpr(vmp *vmParams, env *envPkg.Env, e *ast.LenExpr) (reflect.Value, error) {
	rv, err := invokeExpr(vmp, env, e.Expr)
	if err != nil {
		return nilValue, newError(e.Expr, err)
	}

	if rv.Kind() == reflect.Interface && !rv.IsNil() {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		return reflect.ValueOf(int64(rv.Len())), nil
	default:
		return nilValue, newStringError(e, "type "+rv.Kind().String()+" does not support len operation")
	}
}

func invokeNewExpr(ctx context.Context, e *ast.NewExpr, env *envPkg.Env, expr ast.Expr) (reflect.Value, error) {
	t, err := getTypeFromString(env, e.Type)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if t == nil {
		return nilValue, newErrorf(expr, "type cannot be nil for new")
	}

	return reflect.New(t), nil
}

func invokeMakeExpr(vmp *vmParams, e *ast.MakeExpr, env *envPkg.Env, expr ast.Expr) (reflect.Value, error) {
	t, err := getTypeFromString(env, e.Type)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if t == nil {
		return nilValue, newErrorf(expr, "type cannot be nil for make")
	}

	for i := 1; i < e.Dimensions; i++ {
		t = reflect.SliceOf(t)
	}
	if e.Dimensions < 1 {
		v, err := makeValue(t)
		if err != nil {
			return nilValue, newError(e, err)
		}
		return v, nil
	}

	var alen int
	if e.LenExpr != nil {
		rv, err := invokeExpr(vmp, env, e.LenExpr)
		if err != nil {
			return nilValue, newError(e.LenExpr, err)
		}
		alen = toInt(rv)
	}

	var acap int
	if e.CapExpr != nil {
		rv, err := invokeExpr(vmp, env, e.CapExpr)
		if err != nil {
			return nilValue, newError(e.CapExpr, err)
		}
		acap = toInt(rv)
	} else {
		acap = alen
	}

	return reflect.MakeSlice(reflect.SliceOf(t), alen, acap), nil
}

func invokeMakeTypeExpr(vmp *vmParams, e *ast.MakeTypeExpr, env *envPkg.Env, expr ast.Expr) (reflect.Value, error) {
	rv, err := invokeExpr(vmp, env, e.Type)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if !rv.IsValid() || rv.Type() == nil {
		return nilValue, newErrorf(expr, "type cannot be nil for make type")
	}

	// if e.Name has a dot in it, it should give a syntax error
	// so no needs to check err
	_ = env.DefineReflectType(e.Name, rv.Type())

	return reflect.ValueOf(rv.Type()), nil
}

func invokeMakeChanExpr(vmp *vmParams, e *ast.MakeChanExpr, env *envPkg.Env, expr ast.Expr) (reflect.Value, error) {
	t, err := getTypeFromString(env, e.Type)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if t == nil {
		return nilValue, newErrorf(expr, "type cannot be nil for make chan")
	}

	var size int
	if e.SizeExpr != nil {
		rv, err := invokeExpr(vmp, env, e.SizeExpr)
		if err != nil {
			return nilValue, newError(e.SizeExpr, err)
		}
		size = int(toInt64(rv))
	}

	return reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t), size), nil
}

func invokeChanExpr(vmp *vmParams, e *ast.ChanExpr, env *envPkg.Env, expr ast.Expr) (reflect.Value, error) {
	rhs, err := invokeExpr(vmp, env, e.Rhs)
	if err != nil {
		return nilValue, newError(e.Rhs, err)
	}

	if e.Lhs == nil {
		if rhs.Kind() == reflect.Chan {
			if vmp.validate {
				return nilValue, nil
			}
			cases := []reflect.SelectCase{{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(vmp.ctx.Done()),
				Send: zeroValue,
			}, {
				Dir:  reflect.SelectRecv,
				Chan: rhs,
				Send: zeroValue,
			}}
			chosen, rv, _ := reflect.Select(cases)
			if chosen == 0 {
				return nilValue, ErrInterrupt
			}
			return rv, nil
		}
	} else {
		lhs, err := invokeExpr(vmp, env, e.Lhs)
		if err != nil {
			return nilValue, newError(e.Lhs, err)
		}
		if lhs.Kind() == reflect.Chan {
			chanType := lhs.Type().Elem()
			if chanType == interfaceType || (rhs.IsValid() && rhs.Type() == chanType) {
				cases := []reflect.SelectCase{{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(vmp.ctx.Done()),
					Send: zeroValue,
				}, {
					Dir:  reflect.SelectSend,
					Chan: lhs,
					Send: rhs,
				}}
				if !vmp.validate {
					if chosen, _, _ := reflect.Select(cases); chosen == 0 {
						return nilValue, ErrInterrupt
					}
				}
			} else {
				rhs, err = convertReflectValueToType(vmp.ctx, rhs, chanType)
				if err != nil {
					return nilValue, newStringError(e, "cannot use type "+rhs.Type().String()+" as type "+chanType.String()+" to send to chan")
				}
				cases := []reflect.SelectCase{{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(vmp.ctx.Done()),
					Send: zeroValue,
				}, {
					Dir:  reflect.SelectSend,
					Chan: lhs,
					Send: rhs,
				}}
				if !vmp.validate {
					if chosen, _, _ := reflect.Select(cases); chosen == 0 {
						return nilValue, ErrInterrupt
					}
				}
			}
			return nilValue, nil
		} else if rhs.Kind() == reflect.Chan {
			var rv reflect.Value
			cases := []reflect.SelectCase{{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(vmp.ctx.Done()),
				Send: zeroValue,
			}, {
				Dir:  reflect.SelectRecv,
				Chan: rhs,
				Send: zeroValue,
			}}
			if !vmp.validate {
				var chosen int
				var ok bool
				chosen, rv, ok = reflect.Select(cases)
				if chosen == 0 {
					return nilValue, ErrInterrupt
				}
				if !ok {
					return nilValue, newErrorf(expr, "failed to send to channel")
				}
			}
			return invokeLetExpr(vmp, env, e.Lhs, rv)
		}
	}

	return nilValue, newStringError(e, "invalid operation for chan")
}

func invokeFuncExpr(vmp *vmParams, env *envPkg.Env, e *ast.FuncExpr) (reflect.Value, error) {
	return funcExpr(vmp, env, e)
}

func invokeAnonCallExpr(vmp *vmParams, env *envPkg.Env, e *ast.AnonCallExpr) (reflect.Value, error) {
	return anonCallExpr(vmp, env, e)
}

func invokeCallExpr(vmp *vmParams, env *envPkg.Env, e *ast.CallExpr) (reflect.Value, error) {
	return callExpr(vmp, env, e)
}

func invokeCloseExpr(vmp *vmParams, env *envPkg.Env, e *ast.CloseExpr) (reflect.Value, error) {
	nilValueL := nilValue
	whatExpr, err := invokeExpr(vmp, env, e.WhatExpr)
	if err != nil {
		return nilValueL, newError(e.WhatExpr, err)
	}
	if whatExpr.Kind() == reflect.Interface && !whatExpr.IsNil() {
		whatExpr = whatExpr.Elem()
	}

	switch whatExpr.Kind() {
	case reflect.Chan:
		if whatExpr.IsNil() {
			return nilValueL, nil
		}
		whatExpr.Close()
		return nilValueL, nil
	default:
		return nilValueL, newStringError(e, "first argument to close cannot be type "+whatExpr.Kind().String())
	}
}

func invokeDeleteExpr(vmp *vmParams, env *envPkg.Env, e *ast.DeleteExpr) (reflect.Value, error) {
	nilValueL := nilValue
	whatExpr, err := invokeExpr(vmp, env, e.WhatExpr)
	if err != nil {
		return nilValueL, newError(e.WhatExpr, err)
	}
	var keyExpr reflect.Value
	if e.KeyExpr != nil {
		keyExpr, err = invokeExpr(vmp, env, e.KeyExpr)
		if err != nil {
			return nilValueL, newError(e.KeyExpr, err)
		}
	}
	if whatExpr.Kind() == reflect.Interface && !whatExpr.IsNil() {
		whatExpr = whatExpr.Elem()
	}

	switch whatExpr.Kind() {
	case reflect.String:
		if e.KeyExpr == nil {
			return nilValueL, env.Delete(whatExpr.String())
		}
		if keyExpr.Kind() == reflect.Bool && keyExpr.Bool() {
			return nilValueL, env.DeleteGlobal(whatExpr.String())
		}
		return nilValueL, env.Delete(whatExpr.String())

	case reflect.Map:
		if whatExpr.IsNil() {
			return nilValueL, nil
		}
		if whatExpr.Type().Key() != keyExpr.Type() {
			keyExpr, err = convertReflectValueToType(vmp.ctx, keyExpr, whatExpr.Type().Key())
			if err != nil {
				return nilValueL, newStringError(e, "cannot use type "+whatExpr.Type().Key().String()+" as type "+keyExpr.Type().String()+" in delete")
			}
		}
		vmp.mapMutex.Lock()
		whatExpr.SetMapIndex(keyExpr, reflect.Value{})
		vmp.mapMutex.Unlock()
		return nilValueL, nil
	default:
		return nilValueL, newStringError(e, "first argument to delete cannot be type "+whatExpr.Kind().String())
	}
}

func invokeIncludeExpr(vmp *vmParams, env *envPkg.Env, e *ast.IncludeExpr) (reflect.Value, error) {
	nilValueL := nilValue
	itemExpr, err := invokeExpr(vmp, env, e.ItemExpr)
	if err != nil {
		return nilValueL, newError(e.ItemExpr, err)
	}
	listExpr, err := invokeExpr(vmp, env, e.ListExpr.(*ast.SliceExpr))
	if err != nil {
		return nilValueL, newError(e.ListExpr.(*ast.SliceExpr), err)
	}

	if listExpr.Kind() != reflect.Slice && listExpr.Kind() != reflect.Array {
		return nilValueL, newStringError(e, "second argument must be slice or array; but have "+listExpr.Kind().String())
	}

	for i := 0; i < listExpr.Len(); i++ {
		// todo: https://github.com/alaingilbert/anko/issues/192
		if equal(itemExpr, listExpr.Index(i)) {
			return trueValue, nil
		}
	}
	return falseValue, nil
}
