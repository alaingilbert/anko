//go:build !appengine

package astutil

import (
	"fmt"
	"reflect"

	"github.com/alaingilbert/anko/pkg/ast"
)

// WalkFunc is used in Walk to walk the AST
type WalkFunc func(any, int) error

// Walk walks the ASTs associated with a statement list generated by parser.ParseSrc
// each expression and/or statement is passed to the WalkFunc function.
// If the WalkFunc returns an error the walk is aborted and the error is returned
func Walk(stmt ast.Stmt, f WalkFunc) error {
	return WalkHelper(stmt, f, 0)
}

func WalkHelper(stmt ast.Stmt, f WalkFunc, deep int) error {
	if err := walkStmt(stmt, f, deep); err != nil {
		return err
	}
	return nil
}

func walkExprs(exprs []ast.Expr, f WalkFunc, deep int) error {
	for _, exp := range exprs {
		if err := walkExpr(exp, f, deep); err != nil {
			return err
		}
	}
	return nil
}

func walkStmts(stmts []ast.Stmt, f WalkFunc, deep int) error {
	for _, stmt := range stmts {
		if err := walkStmt(stmt, f, deep); err != nil {
			return err
		}
	}
	return nil
}

func walkStmt(stmt ast.Stmt, f WalkFunc, deep int) error {
	//short circuit out if there are no functions
	if stmt == nil || f == nil {
		return nil
	}
	if err := callFunc(stmt, f, deep); err != nil {
		return err
	}
	deep++
	switch stmt := stmt.(type) {
	case *ast.StmtsStmt:
		if err := walkStmts(stmt.Stmts, f, deep); err != nil {
			return err
		}
	case *ast.BreakStmt:
	case *ast.ContinueStmt:
	case *ast.LetMapItemStmt:
		if err := walkExpr(stmt.Lhss, f, deep); err != nil {
			return err
		}
		if err := walkExpr(stmt.Rhs, f, deep); err != nil {
			return err
		}
	case *ast.ReturnStmt:
		return walkExpr(stmt.Exprs, f, deep)
	case *ast.ExprStmt:
		return walkExpr(stmt.Expr, f, deep)
	case *ast.VarStmt:
		return walkExprs(stmt.Exprs, f, deep)
	case *ast.LetsStmt:
		if err := walkExpr(stmt.Lhss, f, deep); err != nil {
			return err
		}
		if err := walkExpr(stmt.Rhss, f, deep); err != nil {
			return err
		}
		return nil
	case *ast.IfStmt:
		if err := walkExpr(stmt.If, f, deep); err != nil {
			return err
		}
		if err := WalkHelper(stmt.Then, f, deep); err != nil {
			return err
		}
		if err := WalkHelper(stmt.Else, f, deep); err != nil {
			return err
		}
	case *ast.TryStmt:
		if err := WalkHelper(stmt.Try, f, deep); err != nil {
			return err
		}
		if err := WalkHelper(stmt.Catch, f, deep); err != nil {
			return err
		}
		if err := WalkHelper(stmt.Finally, f, deep); err != nil {
			return err
		}
	case *ast.LoopStmt:
		if err := walkExpr(stmt.Expr, f, deep); err != nil {
			return err
		}
		if err := WalkHelper(stmt.Stmt, f, deep); err != nil {
			return err
		}
	case *ast.ForStmt:
		if err := walkExpr(stmt.Value, f, deep); err != nil {
			return err
		}
		if err := WalkHelper(stmt.Stmt, f, deep); err != nil {
			return err
		}
	case *ast.CForStmt:
		if err := walkStmt(stmt.Stmt1, f, deep); err != nil {
			return err
		}
		if err := walkExpr(stmt.Expr2, f, deep); err != nil {
			return err
		}
		if err := walkExpr(stmt.Expr3, f, deep); err != nil {
			return err
		}
		if err := WalkHelper(stmt.Stmt, f, deep); err != nil {
			return err
		}
	case *ast.ThrowStmt:
		if err := walkExpr(stmt.Expr, f, deep); err != nil {
			return err
		}
	case *ast.ModuleStmt:
		if err := WalkHelper(stmt.Stmt, f, deep); err != nil {
			return err
		}
	case *ast.SwitchStmt:
		if err := walkExpr(stmt.Expr, f, deep); err != nil {
			return err
		}
		for _, switchCaseStmt := range stmt.Cases {
			caseStmt := switchCaseStmt.(*ast.SwitchCaseStmt)
			if err := walkStmt(caseStmt.Stmt, f, deep); err != nil {
				return err
			}
		}
		if err := walkStmt(stmt.Default, f, deep); err != nil {
			return err
		}
	case *ast.GoroutineStmt:
		return walkExpr(stmt.Expr, f, deep)
	default:
		return fmt.Errorf("unknown statement %v", reflect.TypeOf(stmt))
	}
	return nil
}

func walkExpr(expr ast.Expr, f WalkFunc, deep int) error {
	//short circuit out if there are no functions
	if expr == nil || f == nil {
		return nil
	}
	if err := callFunc(expr, f, deep); err != nil {
		return err
	}
	switch expr := expr.(type) {
	case *ast.ExprsExpr:
		for i := range expr.Exprs {
			if err := walkExpr(expr.Exprs[i], f, deep); err != nil {
				return err
			}
		}
	case *ast.LenExpr:
	case *ast.NumberExpr:
	case *ast.IdentExpr:
	case *ast.MemberExpr:
		return walkExpr(expr.Expr, f, deep)
	case *ast.StringExpr:
	case *ast.ItemExpr:
		if err := walkExpr(expr.Value, f, deep); err != nil {
			return err
		}
		return walkExpr(expr.Index, f, deep)
	case *ast.SliceExpr:
		if err := walkExpr(expr.Value, f, deep); err != nil {
			return err
		}
		if err := walkExpr(expr.Begin, f, deep); err != nil {
			return err
		}
		return walkExpr(expr.End, f, deep)
	case *ast.ArrayExpr:
		return walkExpr(expr.Exprs, f, deep)
	case *ast.MapExpr:
		if expr.Keys != nil {
			for i := range expr.Keys.Exprs {
				if err := walkExpr(expr.Keys.Exprs[i], f, deep); err != nil {
					return err
				}
				if err := walkExpr(expr.Values.Exprs[i], f, deep); err != nil {
					return err
				}
			}
		}
	case *ast.DerefExpr:
		return walkExpr(expr.Expr, f, deep)
	case *ast.AddrExpr:
		return walkExpr(expr.Expr, f, deep)
	case *ast.UnaryExpr:
		return walkExpr(expr.Expr, f, deep)
	case *ast.ParenExpr:
		return walkExpr(expr.SubExpr, f, deep)
	case *ast.FuncExpr:
		return WalkHelper(expr.Stmt, f, deep)
	case *ast.AssocExpr:
		if err := walkExpr(expr.Lhs, f, deep); err != nil {
			return err
		}
		if err := walkExpr(expr.Rhs, f, deep); err != nil {
			return err
		}
	case *ast.LetsExpr:
		if err := walkExprs(expr.Lhss, f, deep); err != nil {
			return err
		}
		if err := walkExprs(expr.Rhss, f, deep); err != nil {
			return err
		}
	case *ast.BinOpExpr:
		if err := walkExpr(expr.Lhs, f, deep); err != nil {
			return err
		}
		if err := walkExpr(expr.Rhs, f, deep); err != nil {
			return err
		}
	case *ast.ConstExpr:
	case *ast.AnonCallExpr:
		if err := walkExpr(expr.Expr, f, deep); err != nil {
			return err
		}
		return walkExpr(&ast.CallExpr{Func: reflect.Value{}, Callable: &ast.Callable{SubExprs: expr.SubExprs, VarArg: expr.VarArg, Go: expr.Go}}, f, deep)
	case *ast.CallExpr:
		return walkExpr(expr.SubExprs, f, deep)
	case *ast.TernaryOpExpr:
		if err := walkExpr(expr.Expr, f, deep); err != nil {
			return err
		}
		if err := walkExpr(expr.Rhs, f, deep); err != nil {
			return err
		}
		if err := walkExpr(expr.Lhs, f, deep); err != nil {
			return err
		}
	case *ast.MakeExpr:
		if err := walkExpr(expr.LenExpr, f, deep); err != nil {
			return err
		}
		return walkExpr(expr.CapExpr, f, deep)
	case *ast.ChanExpr:
		if err := walkExpr(expr.Rhs, f, deep); err != nil {
			return err
		}
		if err := walkExpr(expr.Lhs, f, deep); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown expression %v", reflect.TypeOf(expr))
	}
	return nil
}

func callFunc(x any, f WalkFunc, deep int) error {
	if x == nil || f == nil {
		return nil
	}
	return f(x, deep)
}
