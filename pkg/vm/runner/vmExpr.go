package runner

import (
	"bytes"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/utils"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// invokeExpr evaluates one expression.
func invokeExpr(vmp *VmParams, env envPkg.IEnv, expr ast.Expr) (reflect.Value, error) {
	if err := incrCycle(vmp); err != nil {
		return nilValue, err
	}
	//fmt.Println("invokeExpr", reflect.ValueOf(expr).String())

	switch e := expr.(type) {
	case *ast.NumberExpr:
		return invokeNumberExpr(vmp, env, e)
	case *ast.IdentExpr:
		return invokeIdentExpr(vmp, env, e)
	case *ast.StringExpr:
		return invokeStringExpr(vmp, env, e)
	case *ast.ArrayExpr:
		return invokeArrayExpr(vmp, env, e)
	case *ast.MapExpr:
		return invokeMapExpr(vmp, env, e)
	case *ast.DerefExpr:
		return invokeDerefExpr(vmp, env, e)
	case *ast.AddrExpr:
		return invokeAddrExpr(vmp, env, e)
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
	case *ast.UnaryExpr:
		return invokeUnaryExpr(vmp, env, e)
	case *ast.BinOpExpr:
		return invokeBinOpExpr(vmp, env, e, expr)
	case *ast.ConstExpr:
		return invokeConstExpr(vmp, env, e)
	case *ast.TernaryOpExpr:
		return invokeTernaryOpExpr(vmp, env, e)
	case *ast.NilCoalescingOpExpr:
		return invokeNilCoalescingOpExpr(vmp, env, e)
	case *ast.LenExpr:
		return invokeLenExpr(vmp, env, e)
	case *ast.DbgExpr:
		return invokeDbgExpr(vmp, env, e)
	case *ast.MakeExpr:
		return invokeMakeExpr(vmp, env, e)
	case *ast.MakeTypeExpr:
		return invokeMakeTypeExpr(vmp, env, e, expr)
	case *ast.ChanExpr:
		return invokeChanExpr(vmp, env, e, expr)
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
		return nilValue, newError(e, ErrUnknownExpr)
	}
}

func invokeNumberExpr(_ *VmParams, env envPkg.IEnv, e *ast.NumberExpr) (reflect.Value, error) {
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

func invokeIdentExpr(_ *VmParams, env envPkg.IEnv, e *ast.IdentExpr) (reflect.Value, error) {
	return env.GetValue(e.Lit)
}

func invokeStringExpr(_ *VmParams, _ envPkg.IEnv, e *ast.StringExpr) (reflect.Value, error) {
	return reflect.ValueOf(e.Lit), nil
}

func makeType(vmp *VmParams, env envPkg.IEnv, typeStruct *ast.TypeStruct) (reflect.Type, error) {
	switch typeStruct.Kind {
	case ast.TypeDefault:
		return getTypeFromEnv(env, typeStruct)
	case ast.TypePtr:
		var t reflect.Type
		var err error
		if typeStruct.SubType != nil {
			t, err = makeType(vmp, env, typeStruct.SubType)
		} else {
			t, err = getTypeFromEnv(env, typeStruct)
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
			t, err = getTypeFromEnv(env, typeStruct)
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
	case ast.TypeMap:
		key, err := makeType(vmp, env, typeStruct.Key)
		if err != nil {
			return nil, err
		}
		if key == nil {
			return nil, err
		}
		t, err := makeType(vmp, env, typeStruct.SubType)
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, err
		}
		//if !runInfo.options.Debug {
		//	// captures panic
		//	defer recoverFunc(runInfo)
		//}
		return reflect.MapOf(key, t), nil
	case ast.TypeChan:
		var t reflect.Type
		var err error
		if typeStruct.SubType != nil {
			t, err = makeType(vmp, env, typeStruct.SubType)
		} else {
			t, err = getTypeFromEnv(env, typeStruct)
		}
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, err
		}
		return reflect.ChanOf(reflect.BothDir, t), nil
	case ast.TypeStructType:
		var t reflect.Type
		fields := make([]reflect.StructField, 0, len(typeStruct.StructNames))
		for i := 0; i < len(typeStruct.StructNames); i++ {
			var err error
			t, err = makeType(vmp, env, typeStruct.StructTypes[i])
			if err != nil {
				return nil, err
			}
			if t == nil {
				return nil, err
			}
			fields = append(fields, reflect.StructField{Name: typeStruct.StructNames[i], Type: t})
		}
		// captures panic
		//defer recoverFunc(runInfo)
		return reflect.StructOf(fields), nil
	default:
		return nil, fmt.Errorf("unknown kind")
	}
}

// recoverFunc generic recover function
//func recoverFunc(runInfo *runInfoStruct) {
//	recoverInterface := recover()
//	if recoverInterface == nil {
//		return
//	}
//	switch value := recoverInterface.(type) {
//	case *Error:
//		runInfo.err = value
//	case error:
//		runInfo.err = value
//	default:
//		runInfo.err = fmt.Errorf("%v", recoverInterface)
//	}
//}

func invokeArrayExpr(vmp *VmParams, env envPkg.IEnv, e *ast.ArrayExpr) (reflect.Value, error) {
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
		rv, err = convertReflectValueToType(vmp, rv, valueType)
		if err != nil {
			return nilValue, newStringError(e, "cannot use type "+rv.Type().String()+" as type "+valueType.String()+" as slice value")
		}
		slice.Index(i).Set(rv)
	}
	return slice, nil
}

func invokeMapExpr(vmp *VmParams, env envPkg.IEnv, e *ast.MapExpr) (reflect.Value, error) {
	nilValueL := nilValue
	m := make(map[any]any, len(e.Keys))
	for i, ee := range e.Keys {
		key, err := invokeExpr(vmp, env, ee)
		if err != nil {
			return nilValueL, newError(ee, err)
		}
		valueExpr := e.Values[i]
		rv, err := invokeExpr(vmp, env, valueExpr)
		if err != nil {
			return nilValueL, newError(valueExpr, err)
		}
		m[key.Interface()] = rv.Interface()
	}
	return reflect.ValueOf(m), nil
}

func invokeDerefExpr(vmp *VmParams, env envPkg.IEnv, e *ast.DerefExpr) (reflect.Value, error) {
	v := nilValue
	switch ee := e.Expr.(type) {
	case *ast.IdentExpr:
		return invokeDeferExprIdentExpr(v, env, ee)
	case *ast.MemberExpr:
		return invokeDeferExprMemberExpr(vmp, env, ee)
	default:
		return nilValue, newError(e, ErrInvalidOperationForTheValue)
	}
}

func invokeDeferExprIdentExpr(v reflect.Value, env envPkg.IEnv, e *ast.IdentExpr) (reflect.Value, error) {
	var err error
	v, err = env.GetValue(e.Lit)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if v.Kind() != reflect.Pointer {
		return nilValue, newStringError(e, "cannot deference for the value")
	}
	return v.Elem(), nil
}

func invokeDeferExprMemberExpr(vmp *VmParams, env envPkg.IEnv, e *ast.MemberExpr) (reflect.Value, error) {
	invalidOperationErr := newInvalidOperation(e)
	v, err := invokeExpr(vmp, env, e.Expr)
	if err != nil {
		return nilValue, newError(e.Expr, err)
	}
	v = elemIfInterface(v)
	if v.Kind() == reflect.Slice {
		v = v.Index(0)
	}
	if v.IsValid() && v.CanInterface() {
		if vme, ok := v.Interface().(envPkg.IEnv); ok {
			m, err := vme.GetValue(e.Name)
			if !m.IsValid() || err != nil {
				return nilValue, invalidOperationErr
			}
			return m, nil
		}
	}

	m := v.MethodByName(e.Name)
	if !m.IsValid() {
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			field, found := v.Type().FieldByName(e.Name)
			if !found {
				return nilValue, newStringError(e, "no member named '"+e.Name+"' for struct")
			}
			return v.FieldByIndex(field.Index), nil
		} else if v.Kind() == reflect.Map {
			// From reflect MapIndex:
			// It returns the zero Value if key is not found in the map or if v represents a nil map.
			m = readMapIndex(v, reflect.ValueOf(e.Name), vmp)
		} else {
			return nilValue, invalidOperationErr
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

func invokeAddrExpr(vmp *VmParams, env envPkg.IEnv, e *ast.AddrExpr) (reflect.Value, error) {
	nilValueL := nilValue
	switch ee := e.Expr.(type) {
	case *ast.IdentExpr:
		return invokeAddrExprIdentExpr(env, ee)
	case *ast.MemberExpr:
		return invokeAddrExprMemberExpr(vmp, env, ee)
	default:
		return nilValueL, newError(e, ErrInvalidOperationForTheValue)
	}
}

func invokeAddrExprIdentExpr(env envPkg.IEnv, e *ast.IdentExpr) (reflect.Value, error) {
	v, err := env.GetValue(e.Lit)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if !v.CanAddr() {
		i := v.Interface()
		return reflect.ValueOf(&i), nil
	}
	return v.Addr(), nil
}

func invokeAddrExprMemberExpr(vmp *VmParams, env envPkg.IEnv, e *ast.MemberExpr) (reflect.Value, error) {
	memberExprName := e.Name
	invalidOperationErr := newInvalidOperation(e)
	v, err := invokeExpr(vmp, env, e.Expr)
	if err != nil {
		return nilValue, newError(e.Expr, err)
	}
	v = elemIfInterface(v)
	if v.Kind() == reflect.Slice {
		v = v.Index(0)
	}
	if v.IsValid() && v.CanInterface() {
		if vme, ok := v.Interface().(envPkg.IEnv); ok {
			m, err := vme.GetValue(memberExprName)
			if !m.IsValid() || err != nil {
				return nilValue, invalidOperationErr
			}
			return m, nil
		}
	}

	m := v.MethodByName(memberExprName)
	if !m.IsValid() {
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			m = v.FieldByName(memberExprName)
			if !m.IsValid() {
				return nilValue, invalidOperationErr
			}
		} else if v.Kind() == reflect.Map {
			// From reflect MapIndex:
			// It returns the zero Value if key is not found in the map or if v represents a nil map.
			m = readMapIndex(v, reflect.ValueOf(memberExprName), vmp)
		} else {
			return nilValue, invalidOperationErr
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

func invokeParenExpr(vmp *VmParams, env envPkg.IEnv, e *ast.ParenExpr) (reflect.Value, error) {
	v, err := invokeExpr(vmp, env, e.SubExpr)
	if err != nil {
		return nilValue, newError(e.SubExpr, err)
	}
	return v, nil
}

func newInvalidOperation(e *ast.MemberExpr) error {
	return newStringError(e, fmt.Sprintf("invalid operation '%s'", e.Name))
}

func invokeMemberExpr(vmp *VmParams, env envPkg.IEnv, e *ast.MemberExpr) (reflect.Value, error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, e.Expr)
	if err != nil {
		return nilValueL, newError(e.Expr, err)
	}
	v = elemIfInterface(v)
	if !v.IsValid() {
		return nilValueL, newError(e, ErrNoSupportMemberOpInvalid)
	}
	if v.CanInterface() {
		if vme, ok := v.Interface().(envPkg.IEnv); ok {
			m, err := vme.GetValue(e.Name)
			if !m.IsValid() || err != nil {
				return nilValueL, newInvalidOperation(e)
			}
			return m, nil
		}
	}

	if method, found := v.Type().MethodByName(e.Name); found {
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
		err := NewNoSupportMemberOpError(v.Kind().String())
		return nilValueL, newError(e, err)
	}
}

func invokeItemExpr(vmp *VmParams, env envPkg.IEnv, e *ast.ItemExpr) (reflect.Value, error) {
	v, err := invokeExpr(vmp, env, e.Value)
	if err != nil {
		return nilValue, newError(e.Value, err)
	}
	i, err := invokeExpr(vmp, env, e.Index)
	if err != nil {
		return nilValue, newError(e.Index, err)
	}
	v = elemIfInterface(v)
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		ii, err := tryToInt(i)
		if err != nil {
			return nilValue, newError(e, ErrIndexMustBeNumber)
		}
		if !vmp.Validate {
			if ii < 0 || ii >= v.Len() {
				return nilValue, newError(e, ErrIndexOutOfRange)
			}
		}
		if v.Kind() != reflect.String {
			if vmp.Validate {
				return zeroOfType(v.Type().Elem()), nil
			}
			return v.Index(ii), nil
		}
		// String
		return v.Index(ii).Convert(StringType), nil
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

func invokeSliceExpr(vmp *VmParams, env envPkg.IEnv, e *ast.SliceExpr) (reflect.Value, error) {
	nilValueL := nilValue
	v, err := invokeExpr(vmp, env, e.Value)
	if err != nil {
		return nilValueL, newError(e.Value, err)
	}
	v = elemIfInterface(v)
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		return sliceExpr(vmp, env, v, e)
	default:
		return nilValueL, newStringError(e, "type "+v.Kind().String()+" does not support slice operation")
	}
}

func sliceExpr(vmp *VmParams, env envPkg.IEnv, v reflect.Value, lhs *ast.SliceExpr) (reflect.Value, error) {
	nilValueL := nilValue
	tryToIntL := tryToInt
	getIdx := func(e ast.Expr, b int) (int, error) {
		var idx int
		if e != nil {
			rb, err := invokeExpr(vmp, env, e)
			if err != nil {
				return 0, newError(lhs, err)
			}
			idx, err = tryToIntL(rb)
			if err != nil {
				return 0, newError(lhs, ErrIndexMustBeNumber)
			}
			if idx < 0 || idx > v.Len() {
				return 0, newError(lhs, ErrIndexOutOfRange)
			}
		} else {
			idx = b
		}
		return idx, nil
	}
	i, err := getIdx(lhs.Begin, 0)
	if err != nil {
		return nilValueL, err
	}
	j, err := getIdx(lhs.End, v.Len())
	if err != nil {
		return nilValueL, err
	}
	if i > j {
		return nilValueL, newError(lhs, ErrInvalidSliceIndex)
	}
	return v.Slice(i, j), nil
}

func invokeAssocExpr(vmp *VmParams, env envPkg.IEnv, e *ast.AssocExpr) (reflect.Value, error) {
	boolToI64 := func(v bool) int64 { return utils.Ternary[int64](v, 1, 0) }
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
				v = reflect.ValueOf(boolToI64(v.Bool()) + 1)
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
				v = reflect.ValueOf(boolToI64(v.Bool()) - 1)
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
	v = elemIfInterface(v)
	return invokeLetExpr(vmp, env, &ast.LetsStmt{Typed: false}, e.Lhs, v)
}

func invokeLetsExpr(vmp *VmParams, env envPkg.IEnv, e *ast.LetsExpr) (reflect.Value, error) {
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
		v = elemIfInterfaceNNil(v)
		_, err = invokeLetExpr(vmp, env, &ast.LetsStmt{Typed: false}, lhs, v)
		if err != nil {
			return nilValue, newError(lhs, err)
		}
	}
	return rvs[len(rvs)-1], nil
}

func invokeUnaryExpr(vmp *VmParams, env envPkg.IEnv, e *ast.UnaryExpr) (reflect.Value, error) {
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
		return nilValue, newError(e, ErrUnknownOperator)
	}
}

func invokeBinOpExpr(vmp *VmParams, env envPkg.IEnv, e *ast.BinOpExpr, expr ast.Expr) (reflect.Value, error) {
	nilValueL := nilValue
	lhsV := nilValueL
	rhsV := nilValueL
	var err error

	lhsV, err = invokeExpr(vmp, env, e.Lhs)
	if err != nil {
		return nilValueL, newError(e.Lhs, err)
	}
	lhsV = elemIfInterfaceNNil(lhsV)
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
		rhsV = elemIfInterfaceNNil(rhsV)
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
		return nilValueL, newError(e, ErrUnknownOperator)
	}
}

func invokeConstExpr(_ *VmParams, _ envPkg.IEnv, e *ast.ConstExpr) (reflect.Value, error) {
	switch e.Value {
	case "true":
		return trueValue, nil
	case "false":
		return falseValue, nil
	}
	return nilValue, nil
}

func invokeTernaryOpExpr(vmp *VmParams, env envPkg.IEnv, e *ast.TernaryOpExpr) (reflect.Value, error) {
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

func invokeNilCoalescingOpExpr(vmp *VmParams, env envPkg.IEnv, e *ast.NilCoalescingOpExpr) (reflect.Value, error) {
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

func invokeLenExpr(vmp *VmParams, env envPkg.IEnv, e *ast.LenExpr) (reflect.Value, error) {
	rv, err := invokeExpr(vmp, env, e.Expr)
	if err != nil {
		return nilValue, newError(e.Expr, err)
	}

	rv = elemIfInterfaceNNil(rv)

	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		return reflect.ValueOf(int64(rv.Len())), nil
	default:
		return nilValue, newStringError(e, "type "+rv.Kind().String()+" does not support len operation")
	}
}

func invokeDbgExpr(vmp *VmParams, env envPkg.IEnv, e *ast.DbgExpr) (reflect.Value, error) {
	if !vmp.DbgEnabled {
		return nilValue, nil
	}
	if e.Expr == nil && e.TypeData == nil {
		println(env.String())
		return nilValue, nil
	} else if e.Expr != nil {
		val, err := invokeExpr(vmp, env, e.Expr)
		if err != nil {
			return nilValue, err
		}
		println(val.String())
		return nilValue, nil
	} else if e.TypeData != nil {
		typeEnv, err := env.GetEnvFromPath(e.TypeData.Env)
		if err != nil {
			return nilValue, err
		}
		if rv, err := typeEnv.GetValue(e.TypeData.Name); err == nil {
			if e, ok := rv.Interface().(*envPkg.Env); ok {
				print(e.String())
				return nilValue, nil
			}
			out := vmUtils.FormatValue(rv)
			if rv.Kind() != reflect.Func {
				out += fmt.Sprintf(" | %s", vmUtils.ReplaceInterface(reflect.TypeOf(rv.Interface()).String()))
			}
			println(out)
			return nilValue, nil
		}

		rt, err := typeEnv.Type(e.TypeData.Name)
		if err != nil {
			return nilValue, err
		}
		if rt.Kind() == reflect.Interface {
			nb := rt.NumMethod()
			methodsArr := make([][]string, 0)
			for i := 0; i < nb; i++ {
				method := rt.Method(i)
				methodsArr = append(methodsArr, []string{method.Name, method.Type.String()})
			}
			maxSymbolLen := sortAndMax(methodsArr)

			buf := new(bytes.Buffer)
			buf.WriteString("type " + rt.Name() + " interface {\n")
			format := "    %-" + strconv.Itoa(maxSymbolLen) + "v %s\n"
			for _, v := range methodsArr {
				buf.WriteString(fmt.Sprintf(format, v[0], v[1]))
			}
			buf.WriteString("}")
			println(buf.String())
			return nilValue, nil
		} else if rt.Kind() == reflect.Struct {
			nb := rt.NumField()
			fieldsArr := make([][]string, 0)
			for i := 0; i < nb; i++ {
				field := rt.Field(i)
				fieldsArr = append(fieldsArr, []string{field.Name, field.Type.String()})
			}
			maxSymbolLen := sortAndMax(fieldsArr)

			buf := new(bytes.Buffer)
			buf.WriteString("type " + rt.Name() + " struct {\n")
			format := "    %-" + strconv.Itoa(maxSymbolLen) + "v %s\n"
			for _, v := range fieldsArr {
				buf.WriteString(fmt.Sprintf(format, v[0], v[1]))
			}
			buf.WriteString("}")
			println(buf.String())
			return nilValue, nil
		}
		println(rt.String())
		return nilValue, nil
	}
	return nilValue, nil
}

func invokeMakeExpr(vmp *VmParams, env envPkg.IEnv, e *ast.MakeExpr) (reflect.Value, error) {
	toIntL := toInt
	t, err := makeType(vmp, env, e.TypeData)
	if err != nil {
		return nilValue, err
	}
	if t == nil {
		return nilValue, newStringError(e, "type cannot be nil for make")
	}
	switch e.TypeData.Kind {
	case ast.TypeSlice:
		aLen := 0
		if e.LenExpr != nil {
			ee := e.LenExpr
			rv, err := invokeExpr(vmp, env, ee)
			if err != nil {
				return nilValue, err
			}
			aLen = toIntL(rv)
		}
		aCap := aLen
		if e.CapExpr != nil {
			ee := e.CapExpr
			rv, err := invokeExpr(vmp, env, ee)
			if err != nil {
				return nilValue, err
			}
			aCap = toIntL(rv)
		}
		if aLen > aCap {
			return nilValue, newStringError(e, "make slice len > cap")
		}
		rv := reflect.MakeSlice(t, aLen, aCap)
		return rv, nil
	case ast.TypeChan:
		aLen := 0
		if e.LenExpr != nil {
			ee := e.LenExpr
			rv, err := invokeExpr(vmp, env, ee)
			if err != nil {
				return nilValue, err
			}
			aLen = toIntL(rv)
		}
		return reflect.MakeChan(t, aLen), nil
	default:
		return MakeValue(t)
	}
}

func invokeMakeTypeExpr(vmp *VmParams, env envPkg.IEnv, e *ast.MakeTypeExpr, expr ast.Expr) (reflect.Value, error) {
	rv, err := invokeExpr(vmp, env, e.Type)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if !rv.IsValid() || rv.Type() == nil {
		return nilValue, newStringError(expr, "type cannot be nil for make type")
	}

	// if e.Name has a dot in it, it should give a syntax error
	// so no needs to check err
	_ = env.DefineReflectType(e.Name, rv.Type())

	return reflect.ValueOf(rv.Type()), nil
}

func invokeChanExpr(vmp *VmParams, env envPkg.IEnv, e *ast.ChanExpr, expr ast.Expr) (reflect.Value, error) {
	rhs, err := invokeExpr(vmp, env, e.Rhs)
	if err != nil {
		return nilValue, newError(e.Rhs, err)
	}

	if e.Lhs == nil {
		if rhs.Kind() == reflect.Chan {
			if vmp.Validate {
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
				return nilValue, vmp.ctx.Err()
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
				if !vmp.Validate {
					if chosen, _, _ := reflect.Select(cases); chosen == 0 {
						return nilValue, vmp.ctx.Err()
					}
				}
			} else {
				buildErr := func(rhs reflect.Value, chanType reflect.Type) error {
					return newStringError(e, "cannot use type "+rhs.Type().String()+" as type "+chanType.String()+" to send to chan")
				}
				if !vmUtils.KindIsNumeric(chanType.Kind()) || !vmUtils.KindIsNumeric(rhs.Type().Kind()) {
					return nilValue, buildErr(rhs, chanType)
				}
				rhs, err = convertReflectValueToType(vmp, rhs, chanType)
				if err != nil {
					return nilValue, buildErr(rhs, chanType)
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
				if !vmp.Validate {
					if chosen, _, _ := reflect.Select(cases); chosen == 0 {
						return nilValue, vmp.ctx.Err()
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
			if !vmp.Validate {
				var chosen int
				var ok bool
				chosen, rv, ok = reflect.Select(cases)
				if chosen == 0 {
					return nilValue, vmp.ctx.Err()
				}
				if !ok {
					return nilValue, newStringError(expr, "failed to send to channel")
				}
			}
			return invokeLetExpr(vmp, env, &ast.LetsStmt{Typed: false}, e.Lhs, rv)
		}
	}

	return nilValue, newStringError(e, "invalid operation for chan")
}

func invokeFuncExpr(vmp *VmParams, env envPkg.IEnv, e *ast.FuncExpr) (reflect.Value, error) {
	return funcExpr(vmp, env, e)
}

func invokeAnonCallExpr(vmp *VmParams, env envPkg.IEnv, e *ast.AnonCallExpr) (reflect.Value, error) {
	return anonCallExpr(vmp, env, e)
}

func invokeCallExpr(vmp *VmParams, env envPkg.IEnv, e *ast.CallExpr) (reflect.Value, error) {
	return callExpr(vmp, env, e)
}

func invokeCloseExpr(vmp *VmParams, env envPkg.IEnv, e *ast.CloseExpr) (reflect.Value, error) {
	nilValueL := nilValue
	whatExpr, err := invokeExpr(vmp, env, e.WhatExpr)
	if err != nil {
		return nilValueL, newError(e.WhatExpr, err)
	}
	whatExpr = elemIfInterfaceNNil(whatExpr)

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

func invokeDeleteExpr(vmp *VmParams, env envPkg.IEnv, e *ast.DeleteExpr) (reflect.Value, error) {
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
	whatExpr = elemIfInterfaceNNil(whatExpr)

	switch whatExpr.Kind() {
	case reflect.String:
		if e.KeyExpr != nil && keyExpr.Kind() == reflect.Bool && keyExpr.Bool() {
			return nilValueL, env.DeleteGlobal(whatExpr.String())
		}
		return nilValueL, env.Delete(whatExpr.String())

	case reflect.Map:
		if whatExpr.IsNil() {
			return nilValueL, nil
		}
		if whatExpr.Type().Key() != keyExpr.Type() {
			keyExpr, err = convertReflectValueToType(vmp, keyExpr, whatExpr.Type().Key())
			if err != nil {
				return nilValueL, newStringError(e, "cannot use type "+whatExpr.Type().Key().String()+" as type "+keyExpr.Type().String()+" in delete")
			}
		}
		setMapIndex(whatExpr, keyExpr, reflect.Value{}, vmp)
		return nilValueL, nil
	default:
		return nilValueL, newStringError(e, "first argument to delete cannot be type "+whatExpr.Kind().String())
	}
}

func invokeIncludeExpr(vmp *VmParams, env envPkg.IEnv, e *ast.IncludeExpr) (reflect.Value, error) {
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
