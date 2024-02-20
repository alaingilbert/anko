package runner

import (
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"reflect"
)

// runSingleStmt executes one statement in the specified environment.
func runSingleStmt(vmp *VmParams, env envPkg.IEnv, stmt ast.Stmt) (reflect.Value, error) {
	if err := incrCycle(vmp); err != nil {
		return nilValue, err
	}
	//fmt.Println("runSingleStmt", reflect.ValueOf(stmt).String())
	switch stmt := stmt.(type) {
	case nil:
		return nilValue, nil
	case *ast.StmtsStmt:
		return runStmtsStmt(vmp, env, stmt)
	case *ast.ExprStmt:
		return runExprStmt(vmp, env, stmt)
	case *ast.VarStmt:
		return runVarStmt(vmp, env, stmt)
	case *ast.LetsStmt:
		return runLetsStmt(vmp, env, stmt)
	case *ast.LetMapItemStmt:
		return runLetMapItemStmt(vmp, env, stmt)
	case *ast.IfStmt:
		return runIfStmt(vmp, env, stmt)
	case *ast.TryStmt:
		return runTryStmt(vmp, env, stmt)
	case *ast.LoopStmt:
		return runLoopStmt(vmp, env, stmt)
	case *ast.ForStmt:
		return runForStmt(vmp, env, stmt)
	case *ast.CForStmt:
		return runCForStmt(vmp, env, stmt)
	case *ast.ReturnStmt:
		return runReturnStmt(vmp, env, stmt)
	case *ast.ThrowStmt:
		return runThrowStmt(vmp, env, stmt)
	case *ast.ModuleStmt:
		return runModuleStmt(vmp, env, stmt)
	case *ast.SelectStmt:
		return runSelectStmt(vmp, env, stmt)
	case *ast.SwitchStmt:
		return runSwitchStmt(vmp, env, stmt)
	case *ast.GoroutineStmt:
		return runGoroutineStmt(vmp, env, stmt)
	case *ast.DeferStmt:
		return runDeferStmt(vmp, env, stmt)
	default:
		return nilValue, newError(stmt, vmUtils.ErrUnknownStmt)
	}
}

func runStmtsStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.StmtsStmt) (reflect.Value, error) {
	rv := nilValue
	var err error
	for _, s := range stmt.Stmts {
		switch s.(type) {
		case *ast.BreakStmt:
			return nilValue, ErrBreak
		case *ast.ContinueStmt:
			return nilValue, ErrContinue
		case *ast.ReturnStmt:
			rv, err = runSingleStmt(vmp, env, s)
			if err != nil {
				return rv, err
			}
			return rv, ErrReturn
		default:
			rv, err = runSingleStmt(vmp, env, s)
			if err != nil {
				return rv, err
			}
		}
	}
	return rv, nil
}

func runExprStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.ExprStmt) (reflect.Value, error) {
	rv, err := invokeExpr(vmp, env, stmt.Expr)
	if err != nil {
		return rv, newError(stmt.Expr, err)
	}
	return rv, nil
}

func runVarStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.VarStmt) (reflect.Value, error) {
	var err error

	// get right side expression values
	rvs := make([]reflect.Value, len(stmt.Exprs))
	for i, expr := range stmt.Exprs {
		rvs[i], err = invokeExpr(vmp, env, expr)
		if err != nil {
			return nilValue, newError(expr, err)
		}
	}

	if len(rvs) == 1 && len(stmt.Names) > 1 {
		// only one right side value but many left side names
		value := rvs[0]
		if value.Kind() == reflect.Interface && !value.IsNil() {
			value = value.Elem()
		}
		if (value.Kind() == reflect.Slice || value.Kind() == reflect.Array) && value.Len() > 0 {
			// value is slice/array, add each value to left side names
			for i := 0; i < value.Len() && i < len(stmt.Names); i++ {
				_ = env.DefineValue(stmt.Names[i], value.Index(i))
			}
			// return last value of slice/array
			return value.Index(value.Len() - 1), nil
		}
	}

	// define all names with right side values
	for i := 0; i < len(rvs) && i < len(stmt.Names); i++ {
		_ = env.DefineValue(stmt.Names[i], rvs[i])
	}

	// return last right side value
	return rvs[len(rvs)-1], nil
}

func runLetsStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.LetsStmt) (reflect.Value, error) {
	nilValueL := nilValue
	var err error

	// get right side expression values
	rvs := make([]reflect.Value, len(stmt.Rhss))
	for i, rhs := range stmt.Rhss {
		rvs[i], err = invokeExpr(vmp, env, rhs)
		if err != nil {
			return nilValueL, newError(rhs, err)
		}
	}

	if len(rvs) == 1 && len(stmt.Lhss) > 1 {
		// only one right side value but many left side expressions
		value := rvs[0]
		if value.Kind() == reflect.Interface && !value.IsNil() {
			value = value.Elem()
		}
		if (value.Kind() == reflect.Slice || value.Kind() == reflect.Array) && value.Len() > 0 {
			// value is slice/array, add each value to left side expression
			for i := 0; i < value.Len() && i < len(stmt.Lhss); i++ {
				_, err = invokeLetExpr(vmp, env, stmt, stmt.Lhss[i], value.Index(i))
				if err != nil {
					return nilValueL, newError(stmt.Lhss[i], err)
				}
			}
			// return last value of slice/array
			return value.Index(value.Len() - 1), nil
		}
	}

	// invoke all left side expressions with right side values
	for i := 0; i < len(rvs) && i < len(stmt.Lhss); i++ {
		value := rvs[i]
		if value.Kind() == reflect.Interface && !value.IsNil() {
			value = value.Elem()
		}
		_, err = invokeLetExpr(vmp, env, stmt, stmt.Lhss[i], value)
		if err != nil {
			return nilValueL, newError(stmt.Lhss[i], err)
		}
	}

	if len(rvs) == 0 {
		return nilValueL, newError(stmt.Lhss[0], errors.New("invalid syntax"))
	}

	// return last right side value
	return rvs[len(rvs)-1], nil
}

func runLetMapItemStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.LetMapItemStmt) (reflect.Value, error) {
	nilValueL := nilValue
	rv, err := invokeExpr(vmp, env, stmt.Rhs)
	if err != nil {
		return nilValueL, newError(stmt, err)
	}
	var rvs []reflect.Value
	if isNil(rv) {
		rvs = []reflect.Value{nilValueL, falseValue}
	} else {
		rvs = []reflect.Value{rv, trueValue}
	}
	for i, lhs := range stmt.Lhss {
		v := rvs[i]
		if v.Kind() == reflect.Interface && !v.IsNil() {
			v = v.Elem()
		}
		_, err = invokeLetExpr(vmp, env, &ast.LetsStmt{Typed: false}, lhs, v)
		if err != nil {
			return nilValueL, newError(lhs, err)
		}
	}
	return rvs[0], nil
}

func runIfStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.IfStmt) (reflect.Value, error) {
	validate := vmp.Validate
	// if
	rv, err := invokeExpr(vmp, env, stmt.If)
	if err != nil {
		return rv, newError(stmt.If, err)
	}

	if toBool(rv) || validate {
		// then
		env.WithNewEnv(func(newenv envPkg.IEnv) {
			rv, err = runSingleStmt(vmp, newenv, stmt.Then)
		})
		if err != nil {
			return rv, newError(stmt, err)
		}
		if !validate {
			return rv, nil
		}
	}

	for _, statement := range stmt.ElseIf {
		elseIf := statement.(*ast.IfStmt)
		// else if - if
		env.WithNewEnv(func(newenv envPkg.IEnv) {
			rv, err = invokeExpr(vmp, newenv, elseIf.If)
		})
		if err != nil {
			return rv, newError(elseIf.If, err)
		}
		if !toBool(rv) && !validate {
			continue
		}

		// else if - then
		env.WithNewEnv(func(newenv envPkg.IEnv) {
			rv, err = runSingleStmt(vmp, newenv, elseIf.Then)
		})
		if err != nil {
			return rv, newError(elseIf, err)
		}
		if !validate {
			return rv, nil
		}
	}

	if stmt.Else != nil {
		// else
		env.WithNewEnv(func(newenv envPkg.IEnv) {
			rv, err = runSingleStmt(vmp, newenv, stmt.Else)
		})
		if err != nil {
			return rv, newError(stmt, err)
		}
	}

	return rv, nil
}

func runTryStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.TryStmt) (reflect.Value, error) {
	validate := vmp.Validate
	newenv := env.NewEnv()
	defer newenv.Destroy()
	_, err := runSingleStmt(vmp, newenv, stmt.Try)
	if err != nil || validate {
		// Catch
		env.WithNewEnv(func(catchEnv envPkg.IEnv) {
			if stmt.Var != "" {
				_ = catchEnv.DefineValue(stmt.Var, reflect.ValueOf(err))
			}
			if _, catchErr := runSingleStmt(vmp, catchEnv, stmt.Catch); catchErr != nil {
				err = newError(stmt.Catch, catchErr)
			} else {
				err = nil
			}
		})
	}
	if stmt.Finally != nil {
		// Finally
		_, e2 := runSingleStmt(vmp, newenv, stmt.Finally)
		if e2 != nil {
			err = newError(stmt.Finally, e2)
		}
	}
	if err != nil {
		return nilValue, newError(stmt, err)
	}
	return nilValue, nil
}

func runLoopStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.LoopStmt) (reflect.Value, error) {
	nilValueL := nilValue
	newenv := env.NewEnv()
	defer newenv.Destroy()
	for {
		if err := incrCycle(vmp); err != nil {
			return nilValueL, err
		}
		if stmt.Expr != nil {
			ev, ee := invokeExpr(vmp, newenv, stmt.Expr)
			if ee != nil {
				return ev, ee
			}
			if !toBool(ev) {
				break
			}
		}

		rv, err := runSingleStmt(vmp, newenv, stmt.Stmt)
		if err != nil && !errors.Is(err, ErrContinue) {
			if errors.Is(err, ErrBreak) {
				break
			}
			if errors.Is(err, ErrReturn) {
				return rv, err
			}
			return nilValueL, newError(stmt, err)
		}
		if vmp.Validate {
			break
		}
	}
	return nilValueL, nil
}

func runForStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.ForStmt) (reflect.Value, error) {
	val, ee := invokeExpr(vmp, env, stmt.Value)
	if ee != nil {
		return val, ee
	}
	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		return runForStmtSlice(vmp, env, stmt, val)
	case reflect.Map:
		return runForStmtMap(vmp, env, stmt, val)
	case reflect.Chan:
		return runForStmtChan(vmp, env, stmt, val)
	default:
		return nilValue, newStringError(stmt, "for cannot loop over type "+val.Kind().String())
	}
}

func runForStmtSlice(vmp *VmParams, env envPkg.IEnv, stmt *ast.ForStmt, val reflect.Value) (reflect.Value, error) {
	nilValueL := nilValue
	newenv := env.NewEnv()
	defer newenv.Destroy()
	for i := 0; i < val.Len(); i++ {
		if err := incrCycle(vmp); err != nil {
			return nilValueL, err
		}
		iv := val.Index(i)
		if iv.Kind() == reflect.Pointer || (iv.Kind() == reflect.Interface && !iv.IsNil()) {
			iv = iv.Elem()
		}
		_ = newenv.DefineValue(stmt.Vars[0], iv)
		rv, err := runSingleStmt(vmp, newenv, stmt.Stmt)
		if err != nil && !errors.Is(err, ErrContinue) {
			if errors.Is(err, ErrBreak) {
				break
			}
			if errors.Is(err, ErrReturn) {
				return rv, err
			}
			return nilValueL, newError(stmt, err)
		}
	}
	return nilValueL, nil
}

func runForStmtMap(vmp *VmParams, env envPkg.IEnv, stmt *ast.ForStmt, val reflect.Value) (reflect.Value, error) {
	nilValueL := nilValue
	newenv := env.NewEnv()
	defer newenv.Destroy()
	keys := val.MapKeys()
	for i := 0; i < len(keys); i++ {
		if err := incrCycle(vmp); err != nil {
			return nilValueL, err
		}
		_ = newenv.DefineValue(stmt.Vars[0], keys[i])
		if len(stmt.Vars) > 1 {
			m := readMapIndex(val, keys[i], vmp)
			_ = newenv.DefineValue(stmt.Vars[1], m)
		}
		rv, err := runSingleStmt(vmp, newenv, stmt.Stmt)
		if err != nil && !errors.Is(err, ErrContinue) {
			if errors.Is(err, ErrBreak) {
				break
			}
			if errors.Is(err, ErrReturn) {
				return rv, err
			}
			return nilValueL, newError(stmt, err)
		}
	}
	return nilValueL, nil
}

func runForStmtChan(vmp *VmParams, env envPkg.IEnv, stmt *ast.ForStmt, val reflect.Value) (reflect.Value, error) {
	newenv := env.NewEnv()
	defer newenv.Destroy()
	for {
		iv := nilValue
		if !vmp.Validate {
			cases := []reflect.SelectCase{{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(vmp.ctx.Done()),
				Send: zeroValue,
			}, {
				Dir:  reflect.SelectRecv,
				Chan: val,
				Send: zeroValue,
			}}
			var chosen int
			var ok bool
			chosen, iv, ok = reflect.Select(cases)
			if chosen == 0 {
				return nilValue, vmp.ctx.Err()
			}
			if !ok {
				break
			}
		}
		if iv.Kind() == reflect.Interface || iv.Kind() == reflect.Pointer {
			iv = iv.Elem()
		}
		_ = newenv.DefineValue(stmt.Vars[0], iv)
		rv, err := runSingleStmt(vmp, newenv, stmt.Stmt)
		if err != nil && !errors.Is(err, ErrContinue) {
			if errors.Is(err, ErrBreak) {
				break
			}
			if errors.Is(err, ErrReturn) {
				return rv, err
			}
			return nilValue, newError(stmt, err)
		}
		if vmp.Validate {
			break
		}
	}
	return nilValue, nil
}

func runCForStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.CForStmt) (reflect.Value, error) {
	nilValueL := nilValue
	newenv := env.NewEnv()
	defer newenv.Destroy()
	_, err := runSingleStmt(vmp, newenv, stmt.Stmt1)
	if err != nil {
		return nilValueL, err
	}
	for {
		if err := incrCycle(vmp); err != nil {
			return nilValueL, err
		}
		fb, err := invokeExpr(vmp, newenv, stmt.Expr2)
		if err != nil {
			return nilValueL, err
		}
		if !toBool(fb) {
			break
		}

		rv, err := runSingleStmt(vmp, newenv, stmt.Stmt)
		if err != nil && !errors.Is(err, ErrContinue) {
			if errors.Is(err, ErrBreak) {
				break
			}
			if errors.Is(err, ErrReturn) {
				return rv, err
			}
			return nilValueL, newError(stmt, err)
		}
		_, err = invokeExpr(vmp, newenv, stmt.Expr3)
		if err != nil {
			return nilValueL, err
		}
		if vmp.Validate {
			break
		}
	}
	return nilValueL, nil
}

func runReturnStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.ReturnStmt) (reflect.Value, error) {
	var err error
	rv := nilValue
	switch len(stmt.Exprs) {
	case 0:
		return rv, nil
	case 1:
		rv, err = invokeExpr(vmp, env, stmt.Exprs[0])
		if err != nil {
			return rv, newError(stmt, err)
		}
		return rv, nil
	}
	rvs := make([]any, len(stmt.Exprs))
	for i, expr := range stmt.Exprs {
		rv, err = invokeExpr(vmp, env, expr)
		if err != nil {
			return rv, newError(stmt, err)
		}
		if !rv.IsValid() || !rv.CanInterface() {
			rvs[i] = nil
		} else {
			rvs[i] = rv.Interface()
		}
	}
	return reflect.ValueOf(rvs), nil
}

func runThrowStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.ThrowStmt) (reflect.Value, error) {
	rv, err := invokeExpr(vmp, env, stmt.Expr)
	if err != nil {
		return rv, newError(stmt, err)
	}
	if !rv.IsValid() {
		return nilValue, newStringError(stmt, "invalid type")
	}
	return rv, newStringError(stmt, fmt.Sprint(rv.Interface()))
}

func runModuleStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.ModuleStmt) (reflect.Value, error) {
	newenv := env.NewEnv()
	defer newenv.Destroy()
	rv, err := runSingleStmt(vmp, newenv, stmt.Stmt)
	if err != nil {
		return rv, newError(stmt, err)
	}
	_ = env.DefineGlobalValue(stmt.Name, reflect.ValueOf(newenv))
	return rv, nil
}

var ErrInvalidOperation = errors.New("invalid operation")
var ErrInvalidOperationForTheValue = errors.New("invalid operation for the value")

func runSelectStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.SelectStmt) (reflect.Value, error) {
	nilValueL := nilValue
	var err error
	newenv := env.NewEnv()
	defer newenv.Destroy()
	body := stmt.Body.(*ast.SelectBodyStmt)
	letsStmts := []*ast.LetsStmt{nil}
	bodies := []ast.Stmt{nil}
	cases := []reflect.SelectCase{{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(vmp.ctx.Done()),
		Send: zeroValue,
	}}
	for _, selectCaseStmt := range body.Cases {
		caseStmt := selectCaseStmt.(*ast.SelectCaseStmt)
		var letStmt *ast.LetsStmt
		var che *ast.ChanExpr
		var ok bool
		var pos ast.Pos = caseStmt.Expr
		switch e := caseStmt.Expr.(type) {
		case *ast.LetsStmt:
			letStmt = e
			pos = e.Rhss[0]
			che, ok = e.Rhss[0].(*ast.ChanExpr)
		case *ast.ExprStmt:
			pos = e.Expr
			che, ok = e.Expr.(*ast.ChanExpr)
		}
		invalidOperationErr := newError(pos, ErrInvalidOperation)
		if !ok {
			return nilValueL, invalidOperationErr
		}
		ident, ok := che.Rhs.(*ast.IdentExpr)
		if !ok {
			return nilValueL, invalidOperationErr
		}
		v, err := newenv.GetValue(ident.Lit)
		if err != nil {
			return nilValueL, newError(che, err)
		}
		letsStmts = append(letsStmts, letStmt)
		bodies = append(bodies, caseStmt.Stmt)
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: v,
			Send: zeroValue,
		})
	}
	if body.Default != nil {
		letsStmts = append(letsStmts, nil)
		bodies = append(bodies, body.Default)
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectDefault,
			Chan: zeroValue,
			Send: zeroValue,
		})
	}

	tmp := func(chosen int, rv reflect.Value) (reflect.Value, error) {
		var err error
		if letStmt := letsStmts[chosen]; letStmt != nil {
			rv, err = invokeLetExpr(vmp, newenv, letStmt, letStmt.Lhss[0], rv)
			if err != nil {
				return nilValueL, newError(letStmt.Lhss[0], err)
			}
		}
		if statements := bodies[chosen]; statements != nil {
			rv, err = runSingleStmt(vmp, newenv, statements)
			if err != nil {
				return rv, err
			}
		}
		return rv, nil
	}

	if vmp.Validate {
		var rv reflect.Value
		for chosen := range cases {
			rv, err = tmp(chosen, rv)
			if err != nil {
				return rv, err
			}
		}
		return rv, nil
	}
	chosen, rv, _ := reflect.Select(cases)
	if chosen == 0 {
		return nilValueL, vmp.ctx.Err()
	}
	return tmp(chosen, rv)
}

func runSwitchStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.SwitchStmt) (reflect.Value, error) {
	validate := vmp.Validate
	newenv := env.NewEnv()
	defer newenv.Destroy()
	rv, err := invokeExpr(vmp, newenv, stmt.Expr)
	if err != nil {
		return rv, newError(stmt, err)
	}

	for _, switchCaseStmt := range stmt.Cases {
		caseStmt := switchCaseStmt.(*ast.SwitchCaseStmt)
		for _, expr := range caseStmt.Exprs {
			caseValue, err := invokeExpr(vmp, newenv, expr)
			if err != nil {
				return nilValue, newError(expr, err)
			}
			if equal(rv, caseValue) || validate {
				val, err := runSingleStmt(vmp, newenv, caseStmt.Stmt)
				if err != nil {
					return val, newError(expr, err)
				}
				if !validate {
					return val, err
				}
			}
		}
	}

	if stmt.Default != nil {
		rv, err = runSingleStmt(vmp, newenv, stmt.Default)
		if err != nil {
			return rv, err
		}
	}

	return rv, nil
}

func runGoroutineStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.GoroutineStmt) (reflect.Value, error) {
	return invokeExpr(vmp, env, stmt.Expr)
}

func runDeferStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.DeferStmt) (reflect.Value, error) {
	switch t := stmt.Expr.(type) {
	case *ast.AnonCallExpr:
		return runDeferStmtAnonCallExpr(vmp, env, t)
	case *ast.CallExpr:
		return runDeferStmtCallExpr(vmp, t, env)
	}
	return invokeExpr(vmp, env, stmt.Expr)
}

func runDeferStmtAnonCallExpr(vmp *VmParams, env envPkg.IEnv, t *ast.AnonCallExpr) (reflect.Value, error) {
	f, err := invokeExpr(vmp, env, t.Expr)
	if err != nil {
		return f, err
	}
	callExprInst := &ast.CallExpr{Func: f, SubExprs: t.SubExprs, VarArg: t.VarArg}
	return runDeferStmtMakeDefer(vmp, f, env, callExprInst)
}

func runDeferStmtCallExpr(vmp *VmParams, t *ast.CallExpr, env envPkg.IEnv) (reflect.Value, error) {
	f := t.Func
	if !f.IsValid() {
		var err error
		f, err = env.GetValue(t.Name)
		if err != nil {
			return f, err
		}
	}
	callExprInst := t
	return runDeferStmtMakeDefer(vmp, f, env, callExprInst)
}

func runDeferStmtMakeDefer(vmp *VmParams, f reflect.Value, env envPkg.IEnv, callExprInst *ast.CallExpr) (reflect.Value, error) {
	injectCtx := false
	if val, ok := f.Interface().(*envPkg.InjectCtx); ok {
		injectCtx = true
		f = reflect.ValueOf(val.Value)
	}
	fType := f.Type()
	isRunVmFunction := checkIfRunVMFunction(fType)
	args, _, useCallSlice, err := makeCallArgs(vmp, env, fType, isRunVmFunction, callExprInst, injectCtx)
	if err != nil {
		return f, err
	}
	env.Defers().Append(envPkg.CapturedFunc{
		Func:      f,
		Args:      args,
		CallSlice: useCallSlice,
	})
	return f, nil
}
