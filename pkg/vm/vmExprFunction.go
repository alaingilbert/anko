package vm

import (
	"context"
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/utils"
	"github.com/alaingilbert/anko/pkg/vm/env"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"

	"github.com/alaingilbert/anko/pkg/ast"
)

// funcExpr creates a function that reflect Call can use.
// When called, it will run runVMFunction, to run the function statements
func funcExpr(vmp *vmParams, env env.IEnv, funcExpr *ast.FuncExpr) (reflect.Value, error) {
	nilValueL := nilValue
	// create the inTypes needed by reflect.FuncOf
	inTypes := make([]reflect.Type, len(funcExpr.Params)+1, len(funcExpr.Params)+1)
	// for runVMFunction first arg is always context
	inTypes[0] = contextType
	for i := 1; i < len(inTypes); i++ {
		if funcExpr.Params[i-1].TypeData != nil {
			t, err := makeType(vmp, env, funcExpr.Params[i-1].TypeData)
			if err != nil {
				return nilValue, err
			}
			inTypes[i] = t
		} else {
			inTypes[i] = reflectValueType
		}
	}
	if funcExpr.VarArg {
		lastType := inTypes[len(inTypes)-1]
		if lastType == reflectValueType {
			inTypes[len(inTypes)-1] = interfaceSliceType
		} else {
			inTypes[len(inTypes)-1] = reflect.SliceOf(lastType)
		}
	}
	// create funcType, output is always slice of reflect.Type with two values
	funcType := reflect.FuncOf(inTypes, []reflect.Type{reflectValueType, reflectValueType}, funcExpr.VarArg)

	// create a function that can be used by reflect.MakeFunc
	// this function is a translator that converts a function call into a vm run
	// returns slice of reflect.Type with two values:
	// return value of the function and error value of the run
	runVMFunction := func(in []reflect.Value) []reflect.Value {
		var err error
		var rv reflect.Value

		// create newEnv for run
		newEnv := env.NewEnv()
		// add Params to newEnv, except last Params
		for i := 0; i < len(funcExpr.Params)-1; i++ {
			var ok bool
			inInterface := in[i+1].Interface()
			if rv, ok = inInterface.(reflect.Value); ok {
				err = newEnv.DefineValue(funcExpr.Params[i].Name, rv)
			} else {
				err = newEnv.DefineValue(funcExpr.Params[i].Name, reflect.ValueOf(inInterface))
			}
			if err != nil {
				return []reflect.Value{reflect.ValueOf(nilValueL), reflect.ValueOf(reflect.ValueOf(newError(funcExpr, err)))}
			}
		}
		// add last Params to newEnv
		if len(funcExpr.Params) > 0 {
			if funcExpr.VarArg {
				// function is variadic, add last Params to newEnv without convert to Interface and then reflect.Value
				rv = in[len(funcExpr.Params)]
				err = newEnv.DefineValue(funcExpr.Params[len(funcExpr.Params)-1].Name, rv)
				if err != nil {
					return []reflect.Value{reflect.ValueOf(nilValueL), reflect.ValueOf(reflect.ValueOf(newError(funcExpr, err)))}
				}
			} else {
				// function is not variadic, add last Params to newEnv
				inInterface := in[len(funcExpr.Params)].Interface()
				if newRv, ok := inInterface.(reflect.Value); ok {
					rv = newRv
				} else {
					rv = reflect.ValueOf(inInterface)
				}
				err = newEnv.DefineValue(funcExpr.Params[len(funcExpr.Params)-1].Name, rv)
				if err != nil {
					return []reflect.Value{reflect.ValueOf(nilValueL), reflect.ValueOf(reflect.ValueOf(newError(funcExpr, err)))}
				}
			}
		}

		ctx := in[0].Interface().(context.Context)
		// run function statements
		newVmp := newVmParams(ctx, vmp.rvCh, vmp.stats, vmp.doNotProtectMaps, vmp.mapMutex, vmp.pause, vmp.rateLimit, vmp.validate, vmp.has, vmp.validateLater)
		rv, err = runSingleStmt(newVmp, newEnv, funcExpr.Stmt)

		for i := newEnv.Defers().Len() - 1; i >= 0; i-- {
			cf := newEnv.Defers().Get(i)
			if cf.CallSlice {
				cf.Func.CallSlice(cf.Args)
			} else {
				cf.Func.Call(cf.Args)
			}
			newEnv.Defers().Remove(i)
		}
		//env.defers = nil

		if err != nil && !errors.Is(err, ErrReturn) {
			err = newError(funcExpr, err)
			// return nil value and error
			// need to do single reflect.ValueOf because nilValue is already reflect.Value of nil
			// need to do double reflect.ValueOf of newError in order to match
			return []reflect.Value{reflect.ValueOf(nilValueL), reflect.ValueOf(reflect.ValueOf(newError(funcExpr, err)))}
		}

		// Validate return values types
		if len(funcExpr.Returns) > 0 {
			if rv.Len() != len(funcExpr.Returns) {
				err = fmt.Errorf("invalid number of returned values")
				return []reflect.Value{reflect.ValueOf(nilValueL), reflect.ValueOf(reflect.ValueOf(newError(funcExpr, err)))}
			}
			for i := 0; i < rv.Len(); i++ {
				rvv := rv.Index(i)
				rvvT := reflect.TypeOf(rvv.Interface())
				expectedT, err := makeType(vmp, env, funcExpr.Returns[i].TypeData)
				if err != nil {
					return []reflect.Value{reflect.ValueOf(nilValueL), reflect.ValueOf(reflect.ValueOf(newError(funcExpr, err)))}
				}
				if rvvT != expectedT {
					err = fmt.Errorf("invalid type for returned value %d", i)
					return []reflect.Value{reflect.ValueOf(nilValueL), reflect.ValueOf(reflect.ValueOf(newError(funcExpr, err)))}
				}
			}
		}

		// the reflect.ValueOf of rv is needed to work in the reflect.Value slice
		// reflectValueErrorNilValue is already a double reflect.ValueOf
		return []reflect.Value{reflect.ValueOf(rv), reflectValueErrorNilValue}
	}

	// make the reflect.Value function that calls runVMFunction
	rv := reflect.MakeFunc(funcType, runVMFunction)

	// if function name is not empty, define it in the env
	if funcExpr.Name != "" {
		err := env.DefineValue(funcExpr.Name, rv)
		if err != nil {
			return nilValueL, newError(funcExpr, err)
		}
	}

	if vmp.validate {
		k := utils.Ternary(funcExpr.Name != "", funcExpr.Name, "invar_"+strconv.Itoa(rand.Intn(math.MaxInt32)))
		vmp.validateLater[k] = funcExpr.Stmt
	}

	// return the reflect.Value created
	return rv, nil
}

// anonCallExpr handles ast.AnonCallExpr which calls a function anonymously
func anonCallExpr(vmp *vmParams, env env.IEnv, e *ast.AnonCallExpr) (reflect.Value, error) {
	f, err := invokeExpr(vmp, env, e.Expr)
	if err != nil {
		return nilValue, newError(e, err)
	}
	if f.Kind() == reflect.Interface && !f.IsNil() {
		f = f.Elem()
	}
	if f.Kind() == reflect.Func {
		callExpr := &ast.CallExpr{Func: f, SubExprs: e.SubExprs, VarArg: e.VarArg, Go: e.Go, Defer: e.Defer}
		callExpr.SetPosition(e.Position())
		return invokeExpr(vmp, env, callExpr)
	}
	if !f.IsValid() {
		return nilValue, newStringError(e, "cannot call type invalid")
	}
	return nilValue, newStringError(e, "cannot call type "+f.Type().String())
}

// callExpr handles *ast.CallExpr which calls a function
func callExpr(vmp *vmParams, env env.IEnv, callExpr *ast.CallExpr) (rv reflect.Value, err error) {
	// Note that if the function type looks the same as the VM function type, the returned values will probably be wrong
	nilValueL := nilValue

	rv = nilValueL

	f := callExpr.Func
	if !f.IsValid() {
		// if function is not valid try to get by function name
		f, err = env.GetValue(callExpr.Name)
		if err != nil {
			err = newError(callExpr, err)
			return
		}
	}

	if vmp.validate {
		for h := range vmp.has {
			if h == fmt.Sprintf("%v", f) {
				vmp.has[h] = true
			}
		}
	}

	if f.Kind() == reflect.Interface && !f.IsNil() {
		f = f.Elem()
	}
	if !f.IsValid() {
		err = newStringError(callExpr, "cannot call type invalid")
		return
	}
	if f.Kind() != reflect.Func {
		err = newStringError(callExpr, "cannot call type "+f.Type().String())
		return
	}

	var rvs []reflect.Value
	var args []reflect.Value
	var useCallSlice bool
	fType := f.Type()
	// check if this is a runVMFunction type
	isRunVMFunction := checkIfRunVMFunction(fType)
	// create/convert the args to the function
	args, _, useCallSlice, err = makeCallArgs(vmp, env, fType, isRunVMFunction, callExpr)
	if err != nil {
		return
	}

	// capture panics if not in debug mode
	if os.Getenv("ANKO_DEBUG") == "" {
		defer func() {
			if recoverResult := recover(); recoverResult != nil {
				err = fmt.Errorf("%v", recoverResult)
			}
		}()
	}

	// useCallSlice lets us know to use CallSlice instead of Call because of the format of the args
	callFn := utils.Ternary(useCallSlice, f.CallSlice, f.Call)
	if callExpr.Go {
		if !vmp.validate {
			go func() {
				rvs := callFn(args)
				// call processCallReturnValues to process runVMFunction return values
				// returns normal VM reflect.Value form
				_, err := processCallReturnValues(rvs, checkIfRunVMFunction(f.Type()), false)
				if err != nil {
					vmp.rvCh <- Result{Value: nilValueL, Error: err}
				}
			}()
		}
		return
	}
	if !vmp.validate {
		rvs = callFn(args)
	} else {
		if isRunVMFunction || callExpr.Name == "import" {
			rvs = f.Call(args)
		} else {
			for i := 0; i < fType.NumOut(); i++ {
				rvs = append(rvs, zeroOfType(fType.Out(i)))
			}
		}
	}

	// TOFIX: how VM pointers/addressing work
	// Until then, this is a work around to set pointers back to VM variables
	// This will probably panic for some functions and/or calls that are variadic
	if !isRunVMFunction {
		for i, expr := range callExpr.SubExprs {
			if addrExpr, ok := expr.(*ast.AddrExpr); ok {
				if identExpr, ok := addrExpr.Expr.(*ast.IdentExpr); ok {
					_, _ = invokeLetExpr(vmp, env, identExpr, args[i].Elem())
				}
			}
		}
	}

	// processCallReturnValues to get/convert return values to normal rv form
	rv, err = processCallReturnValues(rvs, isRunVMFunction, true)
	return
}

// checkIfRunVMFunction checking the number and types of the reflect.Type.
// If it matches the types for a runVMFunction this will return true, otherwise false
func checkIfRunVMFunction(rt reflect.Type) bool {
	if rt.NumIn() < 1 || rt.In(0) != contextType || rt.NumOut() != 2 || rt.Out(0) != reflectValueType || rt.Out(1) != reflectValueType {
		return false
	}
	//if rt.NumIn() > 1 {
	//	if rt.IsVariadic() {
	//		if rt.In(rt.NumIn()-1) != interfaceSliceType {
	//			return false
	//		}
	//	} else {
	//		if rt.In(rt.NumIn()-1) != reflectValueType {
	//			return false
	//		}
	//	}
	//	for i := 1; i < rt.NumIn()-1; i++ {
	//		if rt.In(i) != reflectValueType {
	//			return false
	//		}
	//	}
	//}
	return true
}

// makeCallArgs creates the arguments reflect.Value slice for the four different kinds of functions.
// Also returns true if CallSlice should be used on the arguments, or false if Call should be used.
func makeCallArgs(vmp *vmParams, env env.IEnv, rt reflect.Type, isRunVMFunction bool, callExpr *ast.CallExpr) ([]reflect.Value, []reflect.Type, bool, error) {
	// number of arguments
	numInReal := rt.NumIn()
	numIn := numInReal
	if isRunVMFunction {
		// for runVMFunction the first arg is context so does not count against number of SubExprs
		numIn--
	}
	if numIn < 1 {
		// no arguments needed
		if isRunVMFunction {
			// for runVMFunction first arg is always context
			return []reflect.Value{reflect.ValueOf(vmp.ctx)}, []reflect.Type{reflect.TypeOf(vmp.ctx)}, false, nil
		}
		return []reflect.Value{}, []reflect.Type{}, false, nil
	}

	// number of expressions
	numExprs := len(callExpr.SubExprs)
	// checks to short circuit wrong number of arguments
	if (!rt.IsVariadic() && !callExpr.VarArg && numIn != numExprs) ||
		(rt.IsVariadic() && callExpr.VarArg && (numIn < numExprs || numIn > numExprs+1)) ||
		(rt.IsVariadic() && !callExpr.VarArg && numIn > numExprs+1) ||
		(!rt.IsVariadic() && callExpr.VarArg && numIn < numExprs) {
		err := newStringError(callExpr, fmt.Sprintf("function wants %v arguments but received %v", numIn, numExprs))
		return []reflect.Value{}, []reflect.Type{}, false, err
	}
	if rt.IsVariadic() && rt.In(numInReal-1).Kind() != reflect.Slice && rt.In(numInReal-1).Kind() != reflect.Array {
		err := newStringError(callExpr, "function is variadic but last parameter is of type "+rt.In(numInReal-1).String())
		return []reflect.Value{}, []reflect.Type{}, false, err
	}

	var err error
	var arg reflect.Value
	var args []reflect.Value
	var types []reflect.Type
	indexIn := 0
	indexInReal := 0
	indexExpr := 0

	args = make([]reflect.Value, 0, utils.Ternary(numInReal > numExprs, numInReal, numExprs))
	types = make([]reflect.Type, 0, utils.Ternary(numInReal > numExprs, numInReal, numExprs))
	if isRunVMFunction {
		// for runVMFunction first arg is always context
		argCtx := reflect.ValueOf(vmp.ctx)
		args = append(args, argCtx)
		types = append(types, argCtx.Type())
		indexInReal++
	}

	// create arguments except the last one
	for indexInReal < numInReal-1 && indexExpr < numExprs-1 {
		subExpr := callExpr.SubExprs[indexExpr]
		arg, err = invokeExpr(vmp, env, subExpr)
		if err != nil {
			return []reflect.Value{}, []reflect.Type{}, false, newError(subExpr, err)
		}
		if isRunVMFunction {
			if rt.In(indexInReal) != reflectValueType {
				if arg.Type() != rt.In(indexInReal) {
					err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Type().String())
					return []reflect.Value{}, []reflect.Type{}, false, err
				}
				types = append(types, arg.Type())
				args = append(args, arg)
			} else {
				args = append(args, reflect.ValueOf(arg))
				types = append(types, arg.Type())
			}
		} else {
			arg, err = convertReflectValueToType(vmp, arg, rt.In(indexInReal))
			if err != nil {
				err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Type().String())
				return []reflect.Value{}, []reflect.Type{}, false, err
			}
			types = append(types, arg.Type())
			args = append(args, arg)
		}
		indexIn++
		indexInReal++
		indexExpr++
	}

	if !rt.IsVariadic() && !callExpr.VarArg {
		return makeCallArgsFnNotVarCallNotVar(vmp, env, rt, isRunVMFunction, callExpr, indexInReal, indexExpr, args, types)
	} else if !rt.IsVariadic() && callExpr.VarArg {
		return makeCallArgsFnNotVarCallVar(vmp, env, rt, isRunVMFunction, callExpr, numInReal, indexInReal, indexExpr, numIn, indexIn, numExprs, args, types)
	} else if indexExpr == numExprs {
		return makeCallArgsNoMoreExprs(args, types)
	} else if numIn > numExprs {
		return makeCallArgsDoNotCare(vmp, env, rt, isRunVMFunction, callExpr, indexInReal, indexExpr, args, types)
	} else if rt.IsVariadic() && !callExpr.VarArg {
		return makeCallArgsFnVarCallNotVar(vmp, env, rt, numInReal, indexInReal, indexExpr, numExprs, callExpr, args, types)
	}
	return makeCallArgsFnVarCallVar(vmp, env, rt, arg, callExpr, numInReal, indexInReal, indexExpr, args, types)
}

func makeCallArgsFnNotVarCallNotVar(vmp *vmParams, env env.IEnv, rt reflect.Type, isRunVMFunction bool,
	callExpr *ast.CallExpr, indexInReal, indexExpr int, args []reflect.Value, types []reflect.Type) ([]reflect.Value, []reflect.Type, bool, error) {
	// function is not variadic and call is not variadic
	// add last arguments and return
	subExpr := callExpr.SubExprs[indexExpr]
	arg, err := invokeExpr(vmp, env, subExpr)
	if err != nil {
		return []reflect.Value{}, []reflect.Type{}, false, newError(subExpr, err)
	}
	if isRunVMFunction {
		if rt.In(indexInReal) != reflectValueType {
			if arg.Type() != rt.In(indexInReal) {
				err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Type().String())
				return []reflect.Value{}, []reflect.Type{}, false, err
			}
			args = append(args, arg)
			types = append(types, arg.Type())
		} else {
			args = append(args, reflect.ValueOf(arg))
			types = append(types, arg.Type())
		}
	} else {
		arg, err = convertReflectValueToType(vmp, arg, rt.In(indexInReal))
		if err != nil {
			err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Type().String())
			return []reflect.Value{}, []reflect.Type{}, false, err
		}
		args = append(args, arg)
		types = append(types, arg.Type())
	}
	return args, types, false, nil
}

func makeCallArgsFnNotVarCallVar(vmp *vmParams, env env.IEnv, rt reflect.Type, isRunVMFunction bool, callExpr *ast.CallExpr,
	numInReal, indexInReal, indexExpr, numIn, indexIn, numExprs int, args []reflect.Value, types []reflect.Type) ([]reflect.Value, []reflect.Type, bool, error) {
	// function is not variadic and call is variadic
	subExpr := callExpr.SubExprs[indexExpr]
	arg, err := invokeExpr(vmp, env, subExpr)
	if err != nil {
		return []reflect.Value{}, []reflect.Type{}, false, newError(subExpr, err)
	}
	if arg.Kind() != reflect.Slice && arg.Kind() != reflect.Array {
		return []reflect.Value{}, []reflect.Type{}, false, newStringError(callExpr, "call is variadic but last parameter is of type "+arg.Type().String())
	}
	if arg.Len() < numIn-indexIn {
		err := newStringError(callExpr, fmt.Sprintf("function wants %v arguments but received %v", numIn, numExprs+arg.Len()-1))
		return []reflect.Value{}, []reflect.Type{}, false, err
	}

	indexSlice := 0
	for indexInReal < numInReal {
		if isRunVMFunction {
			if rt.In(indexInReal) != reflectValueType {
				if arg.Index(indexSlice).Type() != rt.In(indexInReal) {
					err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Index(indexSlice).Type().String())
					return []reflect.Value{}, []reflect.Type{}, false, err
				}
				args = append(args, arg.Index(indexSlice))
				types = append(types, arg.Index(indexSlice).Type())
			} else {
				args = append(args, reflect.ValueOf(arg.Index(indexSlice)))
				types = append(types, arg.Index(indexSlice).Type())
			}
		} else {
			arg, err = convertReflectValueToType(vmp, arg.Index(indexSlice), rt.In(indexInReal))
			if err != nil {
				err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Type().String())
				return []reflect.Value{}, []reflect.Type{}, false, err
			}
			args = append(args, arg)
			types = append(types, arg.Type())
		}
		indexIn++
		indexInReal++
		indexSlice++
	}
	return args, types, false, nil
}

func makeCallArgsNoMoreExprs(args []reflect.Value, types []reflect.Type) ([]reflect.Value, []reflect.Type, bool, error) {
	// no more expressions, return what we have and let reflect Call handle if call is variadic or not
	return args, types, false, nil
}

func makeCallArgsDoNotCare(vmp *vmParams, env env.IEnv, rt reflect.Type, isRunVMFunction bool, callExpr *ast.CallExpr,
	indexInReal, indexExpr int, args []reflect.Value, types []reflect.Type) ([]reflect.Value, []reflect.Type, bool, error) {
	// there are more arguments after this one, so does not matter if call is variadic or not
	// add the last argument then return what we have and let reflect Call handle if call is variadic or not
	subExpr := callExpr.SubExprs[indexExpr]
	arg, err := invokeExpr(vmp, env, subExpr)
	if err != nil {
		return []reflect.Value{}, []reflect.Type{}, false, newError(subExpr, err)
	}
	if isRunVMFunction {
		args = append(args, reflect.ValueOf(arg))
		types = append(types, arg.Type())
	} else {
		arg, err = convertReflectValueToType(vmp, arg, rt.In(indexInReal))
		if err != nil {
			err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Type().String())
			return []reflect.Value{}, []reflect.Type{}, false, err
		}
		args = append(args, arg)
		types = append(types, arg.Type())
	}
	return args, types, false, nil
}

func makeCallArgsFnVarCallNotVar(vmp *vmParams, env env.IEnv, rt reflect.Type, numInReal, indexInReal, indexExpr, numExprs int,
	callExpr *ast.CallExpr, args []reflect.Value, types []reflect.Type) ([]reflect.Value, []reflect.Type, bool, error) {
	// function is variadic and call is not variadic
	sliceType := rt.In(numInReal - 1).Elem()
	for indexExpr < numExprs {
		subExpr := callExpr.SubExprs[indexExpr]
		arg, err := invokeExpr(vmp, env, subExpr)
		if err != nil {
			return []reflect.Value{}, []reflect.Type{}, false, newError(subExpr, err)
		}
		arg, err = convertReflectValueToType(vmp, arg, sliceType)
		if err != nil {
			err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Type().String())
			return []reflect.Value{}, []reflect.Type{}, false, err
		}
		args = append(args, arg)
		types = append(types, arg.Type())
		indexExpr++
	}
	return args, types, false, nil
}

func makeCallArgsFnVarCallVar(vmp *vmParams, env env.IEnv, rt reflect.Type, arg reflect.Value,
	callExpr *ast.CallExpr, numInReal, indexInReal, indexExpr int, args []reflect.Value, types []reflect.Type) ([]reflect.Value, []reflect.Type, bool, error) {
	// function is variadic and call is variadic
	// the only time we return CallSlice is true
	var err error
	sliceType := rt.In(numInReal - 1)
	if sliceType.Kind() == reflect.Interface && !arg.IsNil() {
		sliceType = sliceType.Elem()
	}
	subExpr := callExpr.SubExprs[indexExpr]
	arg, err = invokeExpr(vmp, env, subExpr)
	if err != nil {
		return []reflect.Value{}, []reflect.Type{}, false, newError(subExpr, err)
	}
	if sliceType != interfaceSliceType && arg.Type() != sliceType {
		err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Type().String())
		return []reflect.Value{}, []reflect.Type{}, false, err
	}
	arg, err = convertReflectValueToType(vmp, arg, sliceType)
	if err != nil {
		err := newStringError(subExpr, "function wants argument type "+rt.In(indexInReal).String()+" but received type "+arg.Type().String())
		return []reflect.Value{}, []reflect.Type{}, false, err
	}
	args = append(args, arg)
	types = append(types, arg.Type())

	return args, types, true, nil
}

// processCallReturnValues get/converts the values returned from a function call into our normal reflect.Value, error
func processCallReturnValues(rvs []reflect.Value, isRunVMFunction bool, convertToInterfaceSlice bool) (reflect.Value, error) {
	nilValueL := nilValue
	// check if it is not runVMFunction
	if !isRunVMFunction {
		// the function was a Go function, convert to our normal reflect.Value, error
		switch len(rvs) {
		case 0:
			// no return values so return nil reflect.Value and nil error
			return nilValueL, nil
		case 1:
			// one return value but need to add nil error
			return rvs[0], nil
		}
		if convertToInterfaceSlice {
			// need to convert from a slice of reflect.Value to slice of interface
			return reflectValueSliceToInterfaceSlice(rvs), nil
		}
		// need to keep as slice of reflect.Value
		return reflect.ValueOf(rvs), nil
	}

	// is a runVMFunction, expect return in the runVMFunction format
	// convertToInterfaceSlice is ignored
	// some of the below checks probably can be removed because they are done in checkIfRunVMFunction

	if len(rvs) != 2 {
		return nilValueL, fmt.Errorf("VM function did not return 2 values but returned %v values", len(rvs))
	}
	if !rvs[0].IsValid() {
		return nilValueL, fmt.Errorf("VM function value 1 did not return reflect value type but returned invalid type")
	}
	if !rvs[1].IsValid() {
		return nilValueL, fmt.Errorf("VM function value 2 did not return reflect value type but returned invalid type")
	}
	if rvs[0].Type() != reflectValueType {
		return nilValueL, fmt.Errorf("VM function value 1 did not return reflect value type but returned %v type", rvs[0].Type().String())
	}
	if rvs[1].Type() != reflectValueType {
		return nilValueL, fmt.Errorf("VM function value 2 did not return reflect value type but returned %v type", rvs[1].Type().String())
	}

	rvError := rvs[1].Interface().(reflect.Value)
	if !rvError.IsValid() {
		return nilValueL, fmt.Errorf("VM function error type is invalid")
	}
	if rvError.Type() != errorType && rvError.Type() != vmErrorType {
		return nilValueL, fmt.Errorf("VM function error type is %v", rvError.Type())
	}

	if rvError.IsNil() {
		// no error, so return the normal VM reflect.Value form
		return rvs[0].Interface().(reflect.Value), nil
	}

	// VM returns two types of errors, check to see which type
	if rvError.Type() == vmErrorType {
		// convert to VM *Error
		return nilValueL, rvError.Interface().(*Error)
	}
	// convert to error
	return nilValueL, rvError.Interface().(error)
}
