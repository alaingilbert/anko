package runner

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"reflect"
	"strconv"
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
	case *ast.DbgStmt:
		return invokeDbgStmt(vmp, env, stmt)
	case *ast.LabelStmt:
		return invokeLabelStmt(vmp, env, stmt)
	default:
		return nilValue, newError(stmt, ErrUnknownStmt)
	}
}

func runStmtsStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.StmtsStmt) (reflect.Value, error) {
	rv := nilValue
	var err error
	for _, s := range stmt.Stmts {
		switch e := s.(type) {
		case *ast.BreakStmt:
			return nilValue, NewBreakErr(e.Label)
		case *ast.ContinueStmt:
			return nilValue, NewContinueErr(e.Label)
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

func invokeDbgStmt(vmp *VmParams, env envPkg.IEnv, e *ast.DbgStmt) (reflect.Value, error) {
	if !vmp.DbgEnabled {
		return nilValue, nil
	}
	if e.Expr == nil && e.TypeData == nil {
		fmt.Println(env.String())
		return nilValue, nil
	} else if e.Expr != nil {
		val, err := invokeExpr(vmp, env, e.Expr)
		if err != nil {
			return nilValue, err
		}
		fmt.Println(val.String())
		return nilValue, nil
	} else if e.TypeData != nil {
		typeEnv, err := env.GetEnvFromPath(e.TypeData.Env)
		if err != nil {
			return nilValue, err
		}
		if rv, err := typeEnv.GetValue(e.TypeData.Name); err == nil {
			if e, ok := rv.Interface().(*envPkg.Env); ok {
				fmt.Print(e.String())
				return nilValue, nil
			}
			out := vmUtils.FormatValue(rv)
			if rv.Kind() != reflect.Func {
				out += fmt.Sprintf(" | %s", vmUtils.ReplaceInterface(reflect.TypeOf(rv.Interface()).String()))
			}
			fmt.Println(out)
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
			fmt.Println(buf.String())
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
			fmt.Println(buf.String())
			return nilValue, nil
		}
		fmt.Println(rt.String())
		return nilValue, nil
	}
	return nilValue, nil
}

func runVarOrLetStmt[K any](vmp *VmParams, env envPkg.IEnv, exprs1 []K, exprs2 []ast.Expr,
	defineFn func(K, reflect.Value) error) (reflect.Value, error) {
	// get right side expression values
	rvs := make([]reflect.Value, len(exprs2))
	for i, expr := range exprs2 {
		var err error
		rvs[i], err = invokeExpr(vmp, env, expr)
		if err != nil {
			return nilValue, newError(expr, err)
		}
	}

	if len(rvs) == 1 && len(exprs1) > 1 {
		// only one right side value but many left side names
		value := elemIfInterfaceNNil(rvs[0])
		if (value.Kind() == reflect.Slice || value.Kind() == reflect.Array) && value.Len() > 0 {
			// value is slice/array, add each value to left side names
			for i := 0; i < value.Len() && i < len(exprs1); i++ {
				if err := defineFn(exprs1[i], value.Index(i)); err != nil {
					return nilValue, err
				}
			}
			// return last value of slice/array
			return value.Index(value.Len() - 1), nil
		}
	}

	// define all names with right side values
	for i := 0; i < len(rvs) && i < len(exprs1); i++ {
		if err := defineFn(exprs1[i], elemIfInterfaceNNil(rvs[i])); err != nil {
			return nilValue, err
		}
	}

	// return last right side value
	return rvs[len(rvs)-1], nil
}

func runVarStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.VarStmt) (reflect.Value, error) {
	defineFn := func(key string, v reflect.Value) error {
		if err := env.DefineValue(key, v); err != nil {
			return newError(stmt, err)
		}
		return nil
	}
	return runVarOrLetStmt(vmp, env, stmt.Names, stmt.Exprs, defineFn)
}

func runLetsStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.LetsStmt) (reflect.Value, error) {
	defineFn := func(e ast.Expr, value reflect.Value) error {
		if _, err := invokeLetExpr(vmp, env, stmt, e, value); err != nil {
			return newError(e, err)
		}
		return nil
	}
	return runVarOrLetStmt(vmp, env, stmt.Lhss.(*ast.ExprsExpr).Exprs, stmt.Rhss.(*ast.ExprsExpr).Exprs, defineFn)
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
	for i, lhs := range stmt.Lhss.(*ast.ExprsExpr).Exprs {
		v := rvs[i]
		v = elemIfInterfaceNNil(v)
		_, err = invokeLetExpr(vmp, env, &ast.LetsStmt{Typed: false}, lhs, v)
		if err != nil {
			return nilValueL, newError(lhs, err)
		}
	}
	return rvs[0], nil
}

func runIfStmt(vmp *VmParams, env envPkg.IEnv, stmt *ast.IfStmt) (reflect.Value, error) {
	validate := vmp.Validate

	invokeExprFn := func(env envPkg.IEnv, e ast.Expr) (rv reflect.Value, err error) {
		env.WithNewEnv(func(newenv envPkg.IEnv) {
			rv, err = invokeExpr(vmp, newenv, e)
		})
		return rv, err
	}

	runStmtFn := func(env envPkg.IEnv, stmt ast.Stmt) (rv reflect.Value, err error) {
		env.WithNewEnv(func(newenv envPkg.IEnv) {
			rv, err = runSingleStmt(vmp, newenv, stmt)
		})
		return rv, err
	}

	// if
	rv, err := invokeExprFn(env, stmt.If)
	if err != nil {
		return rv, newError(stmt.If, err)
	}

	if toBool(rv) || validate {
		// then
		rv, err = runStmtFn(env, stmt.Then)
		if err != nil {
			return rv, newError(stmt, err)
		}
		if !validate {
			return rv, nil
		}
	}

	if stmt.Else != nil {
		// else
		rv, err = runStmtFn(env, stmt.Else)
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

var (
	errLoopContinue = errors.New("continue")
	errLoopBreak    = errors.New("break")
	errLoopReturn   = errors.New("return")
)

func handleStmtErr(vmp *VmParams, stmt ast.Stmt, err error) error {
	stmtLabel := stmt.GetLabel()
	if err != nil {
		if errors.Is(err, ErrContinue) {
			var cErr *ContinueErr
			if errors.As(err, &cErr) {
				if cErr.label != "" && cErr.label != stmtLabel {
					return errLoopReturn
				}
			}
			if !vmp.Validate {
				return errLoopContinue
			}
		} else if errors.Is(err, ErrBreak) {
			var bErr *BreakErr
			if errors.As(err, &bErr) {
				if bErr.label != "" && bErr.label != stmtLabel {
					return errLoopReturn
				}
			}
			return errLoopBreak
		} else if errors.Is(err, ErrReturn) {
			return errLoopReturn
		} else {
			return errLoopReturn
		}
	}
	return nil
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
		herr := handleStmtErr(vmp, stmt, err)
		if errors.Is(herr, errLoopContinue) {
			continue
		} else if errors.Is(herr, errLoopBreak) {
			break
		} else if errors.Is(herr, errLoopReturn) {
			return rv, err
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
	val = elemIfInterfaceNNil(val)
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
		iv = elemIfInterface(iv)
		_ = newenv.DefineValue(stmt.Vars[0], iv)
		rv, err := runSingleStmt(vmp, newenv, stmt.Stmt)
		herr := handleStmtErr(vmp, stmt, err)
		if errors.Is(herr, errLoopContinue) {
			continue
		} else if errors.Is(herr, errLoopBreak) {
			break
		} else if errors.Is(herr, errLoopReturn) {
			return rv, err
		}
		if vmp.Validate {
			break
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
		herr := handleStmtErr(vmp, stmt, err)
		if errors.Is(herr, errLoopContinue) {
			continue
		} else if errors.Is(herr, errLoopBreak) {
			break
		} else if errors.Is(herr, errLoopReturn) {
			return rv, err
		}
		if vmp.Validate {
			break
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
			cases := []reflect.SelectCase{
				{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(vmp.ctx.Done()), Send: zeroValue},
				{Dir: reflect.SelectRecv, Chan: val, Send: zeroValue}}
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
		iv = elemIfInterface(iv)
		_ = newenv.DefineValue(stmt.Vars[0], iv)
		rv, err := runSingleStmt(vmp, newenv, stmt.Stmt)
		herr := handleStmtErr(vmp, stmt, err)
		if errors.Is(herr, errLoopContinue) {
			continue
		} else if errors.Is(herr, errLoopBreak) {
			break
		} else if errors.Is(herr, errLoopReturn) {
			return rv, err
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
		if err != nil {
			if errors.Is(err, ErrContinue) {
				var cErr *ContinueErr
				if errors.As(err, &cErr) {
					if cErr.label != "" && cErr.label != stmt.Label {
						return nilValue, cErr
					}
				}
			} else if errors.Is(err, ErrBreak) {
				var bErr *BreakErr
				if errors.As(err, &bErr) {
					if bErr.label != "" {
						return nilValueL, bErr
					}
				}
				break
			} else if errors.Is(err, ErrReturn) {
				return rv, err
			} else {
				return nilValueL, newError(stmt, err)
			}
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
	switch len(stmt.Exprs.(*ast.ExprsExpr).Exprs) {
	case 0:
		return rv, nil
	case 1:
		rv, err = invokeExpr(vmp, env, stmt.Exprs.(*ast.ExprsExpr).Exprs[0])
		if err != nil {
			return rv, newError(stmt, err)
		}
		return rv, nil
	}
	rvs := make([]any, len(stmt.Exprs.(*ast.ExprsExpr).Exprs))
	for i, expr := range stmt.Exprs.(*ast.ExprsExpr).Exprs {
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
			pos = e.Rhss.(*ast.ExprsExpr).Exprs[0]
			che, ok = e.Rhss.(*ast.ExprsExpr).Exprs[0].(*ast.ChanExpr)
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
			rv, err = invokeLetExpr(vmp, newenv, letStmt, letStmt.Lhss.(*ast.ExprsExpr).Exprs[0], rv)
			if err != nil {
				return nilValueL, newError(letStmt.Lhss.(*ast.ExprsExpr).Exprs[0], err)
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
		for _, expr := range caseStmt.Exprs.Exprs {
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
	callExprInst := &ast.CallExpr{Func: f, Callable: &ast.Callable{SubExprs: t.SubExprs, VarArg: t.VarArg}}
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

func invokeLabelStmt(vmp *VmParams, env envPkg.IEnv, e *ast.LabelStmt) (reflect.Value, error) {
	rv, err := runSingleStmt(vmp, env, e.Stmt)
	if err != nil {
		var bErr *BreakErr
		if errors.As(err, &bErr) {
			if bErr.label == e.Name {
				return nilValue, nil
			}
		}
		var cErr *ContinueErr
		if errors.As(err, &cErr) {
			if cErr.label == e.Name {
				return nilValue, nil
			}
		}
		return nilValue, err
	}
	return rv, nil
}
