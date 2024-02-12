package vm

import (
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"reflect"
)

// runSingleStmt executes one statement in the specified environment.
func runSingleStmt(vmp *vmParams, env *envPkg.Env, stmt ast.Stmt) (reflect.Value, error) {
	if err := incrCycle(vmp); err != nil {
		return nilValue, ErrInterrupt
	}
	//fmt.Println("runSingleStmt", reflect.ValueOf(stmt).String())
	switch stmt := stmt.(type) {
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
		return nilValue, newStringError(stmt, "unknown statement")
	}
}

func runExprStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.ExprStmt) (reflect.Value, error) {
	rv, err := invokeExpr(vmp, env, stmt.Expr)
	if err != nil {
		return rv, newError(stmt.Expr, err)
	}
	return rv, nil
}

func runVarStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.VarStmt) (reflect.Value, error) {
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

func runLetsStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.LetsStmt) (reflect.Value, error) {
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
				_, err = invokeLetExpr(vmp, env, stmt.Lhss[i], value.Index(i))
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
		_, err = invokeLetExpr(vmp, env, stmt.Lhss[i], value)
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

func runLetMapItemStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.LetMapItemStmt) (reflect.Value, error) {
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
		_, err = invokeLetExpr(vmp, env, lhs, v)
		if err != nil {
			return nilValueL, newError(lhs, err)
		}
	}
	return rvs[0], nil
}

func runIfStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.IfStmt) (reflect.Value, error) {
	// if
	rv, err := invokeExpr(vmp, env, stmt.If)
	if err != nil {
		return rv, newError(stmt.If, err)
	}

	if toBool(rv) {
		// then
		newenv := env.NewEnv()
		rv, err = run(vmp, newenv, stmt.Then)
		if err != nil {
			return rv, newError(stmt, err)
		}
		return rv, nil
	}

	for _, statement := range stmt.ElseIf {
		elseIf := statement.(*ast.IfStmt)
		// else if - if
		newenv := env.NewEnv()
		rv, err = invokeExpr(vmp, newenv, elseIf.If)
		if err != nil {
			return rv, newError(elseIf.If, err)
		}
		if !toBool(rv) {
			continue
		}

		// else if - then
		newenv = env.NewEnv()
		rv, err = run(vmp, newenv, elseIf.Then)
		if err != nil {
			return rv, newError(elseIf, err)
		}
		return rv, nil
	}

	if len(stmt.Else) > 0 {
		// else
		newenv := env.NewEnv()
		rv, err = run(vmp, newenv, stmt.Else)
		if err != nil {
			return rv, newError(stmt, err)
		}
	}

	return rv, nil
}

func runTryStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.TryStmt) (reflect.Value, error) {
	newenv := env.NewEnv()
	_, err := run(vmp, newenv, stmt.Try)
	if err != nil {
		// Catch
		cenv := env.NewEnv()
		if stmt.Var != "" {
			_ = cenv.DefineValue(stmt.Var, reflect.ValueOf(err))
		}
		_, e1 := run(vmp, cenv, stmt.Catch)
		if e1 != nil {
			err = newError(stmt.Catch[0], e1)
		} else {
			err = nil
		}
	}
	if len(stmt.Finally) > 0 {
		// Finally
		_, e2 := run(vmp, newenv, stmt.Finally)
		if e2 != nil {
			err = newError(stmt.Finally[0], e2)
		}
	}
	return nilValue, newError(stmt, err)
}

func runLoopStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.LoopStmt) (reflect.Value, error) {
	nilValueL := nilValue
	newenv := env.NewEnv()
	for {
		if err := incrCycle(vmp); err != nil {
			return nilValueL, ErrInterrupt
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

		rv, err := run(vmp, newenv, stmt.Stmts)
		if err != nil && !errors.Is(err, ErrContinue) {
			if errors.Is(err, ErrBreak) {
				break
			}
			if errors.Is(err, ErrReturn) {
				return rv, err
			}
			return nilValueL, newError(stmt, err)
		}
		if vmp.validate {
			break
		}
	}
	return nilValueL, nil
}

func runForStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.ForStmt) (reflect.Value, error) {
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

func runForStmtSlice(vmp *vmParams, env *envPkg.Env, stmt *ast.ForStmt, val reflect.Value) (reflect.Value, error) {
	nilValueL := nilValue
	newenv := env.NewEnv()
	for i := 0; i < val.Len(); i++ {
		if err := incrCycle(vmp); err != nil {
			return nilValueL, ErrInterrupt
		}
		iv := val.Index(i)
		if iv.Kind() == reflect.Pointer || (iv.Kind() == reflect.Interface && !iv.IsNil()) {
			iv = iv.Elem()
		}
		_ = newenv.DefineValue(stmt.Vars[0], iv)
		rv, err := run(vmp, newenv, stmt.Stmts)
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

func runForStmtMap(vmp *vmParams, env *envPkg.Env, stmt *ast.ForStmt, val reflect.Value) (reflect.Value, error) {
	nilValueL := nilValue
	newenv := env.NewEnv()
	keys := val.MapKeys()
	for i := 0; i < len(keys); i++ {
		if err := incrCycle(vmp); err != nil {
			return nilValueL, ErrInterrupt
		}
		_ = newenv.DefineValue(stmt.Vars[0], keys[i])
		if len(stmt.Vars) > 1 {
			vmp.mapMutex.Lock()
			m := val.MapIndex(keys[i])
			vmp.mapMutex.Unlock()
			_ = newenv.DefineValue(stmt.Vars[1], m)
		}
		rv, err := run(vmp, newenv, stmt.Stmts)
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

func runForStmtChan(vmp *vmParams, env *envPkg.Env, stmt *ast.ForStmt, val reflect.Value) (reflect.Value, error) {
	newenv := env.NewEnv()
	for {
		iv := nilValue
		if !vmp.validate {
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
				return nilValue, ErrInterrupt
			}
			if !ok {
				break
			}
		}
		if iv.Kind() == reflect.Interface || iv.Kind() == reflect.Pointer {
			iv = iv.Elem()
		}
		_ = newenv.DefineValue(stmt.Vars[0], iv)
		rv, err := run(vmp, newenv, stmt.Stmts)
		if err != nil && !errors.Is(err, ErrContinue) {
			if errors.Is(err, ErrBreak) {
				break
			}
			if errors.Is(err, ErrReturn) {
				return rv, err
			}
			return nilValue, newError(stmt, err)
		}
		if vmp.validate {
			break
		}
	}
	return nilValue, nil
}

func runCForStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.CForStmt) (reflect.Value, error) {
	nilValueL := nilValue
	newenv := env.NewEnv()
	_, err := runSingleStmt(vmp, newenv, stmt.Stmt1)
	if err != nil {
		return nilValueL, err
	}
	for {
		if err := incrCycle(vmp); err != nil {
			return nilValueL, ErrInterrupt
		}
		fb, err := invokeExpr(vmp, newenv, stmt.Expr2)
		if err != nil {
			return nilValueL, err
		}
		if !toBool(fb) {
			break
		}

		rv, err := run(vmp, newenv, stmt.Stmts)
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
		if vmp.validate {
			break
		}
	}
	return nilValueL, nil
}

func runReturnStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.ReturnStmt) (reflect.Value, error) {
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

func runThrowStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.ThrowStmt) (reflect.Value, error) {
	rv, err := invokeExpr(vmp, env, stmt.Expr)
	if err != nil {
		return rv, newError(stmt, err)
	}
	if !rv.IsValid() {
		return nilValue, newError(stmt, err)
	}
	return rv, newStringError(stmt, fmt.Sprint(rv.Interface()))
}

func runModuleStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.ModuleStmt) (reflect.Value, error) {
	newenv := env.NewEnv()
	rv, err := run(vmp, newenv, stmt.Stmts)
	if err != nil {
		return rv, newError(stmt, err)
	}
	_ = env.DefineGlobalValue(stmt.Name, reflect.ValueOf(newenv))
	return rv, nil
}

func runSelectStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.SelectStmt) (reflect.Value, error) {
	nilValueL := nilValue
	var err error
	newenv := env.NewEnv()
	body := stmt.Body.(*ast.SelectBodyStmt)
	letsStmts := []*ast.LetsStmt{nil}
	bodies := [][]ast.Stmt{nil}
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
		if !ok {
			return nilValueL, newStringError(pos, "invalid operation")
		}
		ident, ok := che.Rhs.(*ast.IdentExpr)
		if !ok {
			return nilValueL, newStringError(pos, "invalid operation")
		}
		v, err := newenv.GetValue(ident.Lit)
		if err != nil {
			return nilValueL, newError(che, err)
		}
		letsStmts = append(letsStmts, letStmt)
		bodies = append(bodies, caseStmt.Stmts)
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
			rv, err = invokeLetExpr(vmp, newenv, letStmt.Lhss[0], rv)
			if err != nil {
				return nilValueL, newError(letStmt.Lhss[0], err)
			}
		}
		if statements := bodies[chosen]; statements != nil {
			rv, err = run(vmp, newenv, statements)
			if err != nil {
				return rv, err
			}
		}
		return rv, nil
	}

	if vmp.validate {
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
		return nilValueL, ErrInterrupt
	}
	return tmp(chosen, rv)
}

func runSwitchStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.SwitchStmt) (reflect.Value, error) {
	newenv := env.NewEnv()
	rv, err := invokeExpr(vmp, newenv, stmt.Expr)
	if err != nil {
		return rv, newError(stmt, err)
	}

	var caseValue reflect.Value
	body := stmt.Body.(*ast.SwitchBodyStmt)
	var statements []ast.Stmt
	if body.Default != nil {
		statements = body.Default
	}

Loop:
	for _, switchCaseStmt := range body.Cases {
		caseStmt := switchCaseStmt.(*ast.SwitchCaseStmt)
		for _, expr := range caseStmt.Exprs {
			caseValue, err = invokeExpr(vmp, newenv, expr)
			if err != nil {
				return nilValue, newError(expr, err)
			}
			if equal(rv, caseValue) {
				statements = caseStmt.Stmts
				break Loop
			}
		}
	}

	if statements != nil {
		rv, err = run(vmp, newenv, statements)
		if err != nil {
			return rv, err
		}
	}

	return rv, nil
}

func runGoroutineStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.GoroutineStmt) (reflect.Value, error) {
	return invokeExpr(vmp, env, stmt.Expr)
}

func runDeferStmt(vmp *vmParams, env *envPkg.Env, stmt *ast.DeferStmt) (reflect.Value, error) {
	switch t := stmt.Expr.(type) {
	case *ast.AnonCallExpr:
		return runDeferStmtAnonCallExpr(vmp, env, t)
	case *ast.CallExpr:
		return runDeferStmtCallExpr(vmp, t, env)
	}
	return invokeExpr(vmp, env, stmt.Expr)
}

func runDeferStmtAnonCallExpr(vmp *vmParams, env *envPkg.Env, t *ast.AnonCallExpr) (reflect.Value, error) {
	f, err := invokeExpr(vmp, env, t.Expr)
	if err != nil {
		return f, err
	}
	callExprInst := &ast.CallExpr{Func: f, SubExprs: t.SubExprs, VarArg: t.VarArg}
	return runDeferStmtMakeDefer(vmp, f, env, callExprInst)
}

func runDeferStmtCallExpr(vmp *vmParams, t *ast.CallExpr, env *envPkg.Env) (reflect.Value, error) {
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

func runDeferStmtMakeDefer(vmp *vmParams, f reflect.Value, env *envPkg.Env, callExprInst *ast.CallExpr) (reflect.Value, error) {
	fType := f.Type()
	isRunVmFunction := checkIfRunVMFunction(fType)
	args, useCallSlice, err := makeCallArgs(vmp, env, fType, isRunVmFunction, callExprInst)
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
