package decompiler

import (
	"bytes"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"reflect"
	"strings"
)

func Decompile(stmt ast.Stmt) string {
	buf := new(bytes.Buffer)
	decompileStmt(buf, stmt, 0)
	return buf.String()
}

func decompileStmt(w *bytes.Buffer, stmt ast.Stmt, deep int) {
	prefix := strings.Repeat(" ", 4*deep)
	switch s := stmt.(type) {
	case *ast.IfStmt:
		decompileIfStmt(w, prefix, s, deep)
	case *ast.SwitchStmt:
		decompileSwitchStmt(w, prefix, s, deep)
	case *ast.SwitchCaseStmt:
		decompileSwitchCaseStmt(w, prefix, s, deep)
	case *ast.ExprStmt: // Completed
		decompileExpr(w, s.Expr, deep)
	case *ast.LetsStmt: // Completed
		decompileLetsStmt(w, prefix, s)
	case *ast.BreakStmt: // Completed
		decompileBreakStmt(w, prefix)
	case *ast.ReturnStmt:
		decompileReturnStmt(w, s, prefix)
	case *ast.LoopStmt:
		decompileLoopStmt(w, prefix, s, deep)
	case *ast.ForStmt:
		decompileForStmt(w, prefix, s, deep)
	case *ast.CForStmt: // Completed
		decompileCForStmt(w, prefix, s, deep)
	default:
		panic(fmt.Sprintf("stmt: %s %s", s, reflect.TypeOf(s)))
	}
}

func decompileExpr(w *bytes.Buffer, expr ast.Expr, deep int) {
	prefix := strings.Repeat(" ", 4*deep)
	switch e := expr.(type) {
	case nil:
	case *ast.ParenExpr: // Completed
		decompileParenExpr(w, prefix, e)
	case *ast.LenExpr: // Completed
		decompileLenExpr(w, prefix, e)
	case *ast.ItemExpr:
		decompileItemExpr(w, prefix, e)
	case *ast.UnaryExpr:
		decompileUnaryExpr(w, prefix, e)
	case *ast.MemberExpr:
		decompileMemberExpr(w, prefix, e)
	case *ast.AnonCallExpr:
		decompileAnonCallExpr(w, prefix, e)
	case *ast.FuncExpr:
		decompileFuncExpr(w, prefix, e, deep)
	case *ast.ConstExpr:
		decompileConstExpr(w, prefix, e)
	case *ast.MapExpr:
		decompileMapExpr(w, prefix, e)
	case *ast.ArrayExpr:
		decompileArrayExpr(w, prefix, e)
	case *ast.StringExpr: // Completed
		decompileStringExpr(w, prefix, e)
	case *ast.CallExpr: // Completed
		decompileCallExpr(w, prefix, e)
	case *ast.AssocExpr: // Completed
		decompileAssocExpr(w, prefix, e)
	case *ast.BinOpExpr: // Completed
		decompileBinOpExpr(w, prefix, e)
	case *ast.NumberExpr: // Completed
		decompileNumberExpr(w, prefix, e)
	case *ast.IdentExpr: // Completed
		decompileIdentExpr(w, prefix, e)
	default:
		panic(fmt.Sprintf("expr: %s %s", e, reflect.TypeOf(e)))
	}
}

func decompileCForStmt(w *bytes.Buffer, prefix string, s *ast.CForStmt, deep int) {
	w.WriteString(prefix + "for ")
	if el, ok := s.Stmt1.(*ast.LetsStmt); ok {
		decompileLetsStmtHelper(w, el)
	}
	w.WriteString("; ")
	decompileExpr(w, s.Expr2, 0)
	w.WriteString("; ")
	decompileExpr(w, s.Expr3, 0)
	w.WriteString(" {\n")
	decompileStmt(w, s.Stmt, deep+1)
	w.WriteString(prefix + "}\n")
}

func joinStr(w *bytes.Buffer, arr []string) {
	for i := 0; i < len(arr); i++ {
		w.WriteString(arr[i])
		if i < len(arr)-1 {
			w.WriteString(", ")
		}
	}
}
func joinExpr(w *bytes.Buffer, arr []ast.Expr) {
	for i := 0; i < len(arr); i++ {
		decompileExpr(w, arr[i], 0)
		if i < len(arr)-1 {
			w.WriteString(", ")
		}
	}
}

func decompileForStmt(w *bytes.Buffer, prefix string, s *ast.ForStmt, deep int) {
	w.WriteString(prefix + "for ")
	joinStr(w, s.Vars)
	if len(s.Vars) > 0 {
		w.WriteString(" in ")
	}
	decompileExpr(w, s.Value, 0)
	w.WriteString(" {\n")
	decompileStmt(w, s.Stmt, deep+1)
	w.WriteString(prefix + "}\n")
}

func decompileLoopStmt(w *bytes.Buffer, prefix string, s *ast.LoopStmt, deep int) {
	w.WriteString(prefix + "for ")
	decompileExpr(w, s.Expr, 0)
	w.WriteString(" {\n")
	decompileStmt(w, s.Stmt, deep+1)
	w.WriteString(prefix + "}\n")
}

func decompileReturnStmt(w *bytes.Buffer, s *ast.ReturnStmt, prefix string) {
	for _, e := range s.Exprs {
		w.WriteString(prefix + "return ")
		decompileExpr(w, e, 0)
		w.WriteString("\n")
	}
}

func decompileBreakStmt(w *bytes.Buffer, prefix string) {
	_, _ = w.WriteString(prefix + "break\n")
}

func decompileLetsStmtHelper(w *bytes.Buffer, s *ast.LetsStmt) {
	joinExpr(w, s.Lhss)
	w.WriteString(" = ")
	joinExpr(w, s.Rhss)
}

func decompileLetsStmt(w *bytes.Buffer, prefix string, s *ast.LetsStmt) {
	w.WriteString(prefix)
	decompileLetsStmtHelper(w, s)
	w.WriteString("\n")
}

func decompileSwitchCaseStmt(w *bytes.Buffer, prefix string, s *ast.SwitchCaseStmt, deep int) {
	w.WriteString(prefix + "case ")
	joinExpr(w, s.Exprs)
	w.WriteString(":\n")
	decompileStmt(w, s.Stmt, deep+1)
}

func decompileSwitchStmt(w *bytes.Buffer, prefix string, s *ast.SwitchStmt, deep int) {
	w.WriteString(prefix + "switch ")
	decompileExpr(w, s.Expr, 0)
	w.WriteString(" {\n")
	for _, c := range s.Cases {
		decompileStmt(w, c, deep)
		w.WriteString("\n")
	}
	w.WriteString(prefix + "default:\n")
	decompileStmt(w, s.Default, deep+1)
	w.WriteString(prefix + "}\n")
}

func decompileIfStmt(w *bytes.Buffer, prefix string, s *ast.IfStmt, deep int) {
	w.WriteString(prefix + "if ")
	decompileExpr(w, s.If, 0)
	w.WriteString(" {\n")
	decompileStmt(w, s.Then, deep+1)
	if s.Else != nil {
		w.WriteString(prefix + "} else {\n")
		decompileStmt(w, s.Else, deep+1)
	}
	w.WriteString(prefix + "}\n")
}

func decompileParenExpr(w *bytes.Buffer, prefix string, e *ast.ParenExpr) {
	w.WriteString(prefix)
	w.WriteString("(")
	decompileExpr(w, e.SubExpr, 0)
	w.WriteString(")")
}

func decompileLenExpr(w *bytes.Buffer, prefix string, e *ast.LenExpr) {
	w.WriteString(prefix)
	w.WriteString("len(")
	decompileExpr(w, e.Expr, 0)
	w.WriteString(")")
}

func decompileItemExpr(w *bytes.Buffer, prefix string, e *ast.ItemExpr) {
	w.WriteString(prefix)
	decompileExpr(w, e.Value, 0)
	w.WriteString("[")
	decompileExpr(w, e.Index, 0)
	w.WriteString("]")
}

func decompileUnaryExpr(w *bytes.Buffer, prefix string, e *ast.UnaryExpr) {
	w.WriteString(prefix)
	w.WriteString(e.Operator)
	decompileExpr(w, e.Expr, 0)
}

func decompileMemberExpr(w *bytes.Buffer, prefix string, e *ast.MemberExpr) {
	w.WriteString(prefix)
	decompileExpr(w, e.Expr, 0)
	w.WriteString(".")
	w.WriteString(e.Name)
}

func decompileAnonCallExpr(w *bytes.Buffer, prefix string, e *ast.AnonCallExpr) {
	w.WriteString(prefix)
	decompileExpr(w, e.Expr, 0)
	w.WriteString("(")
	joinExpr(w, e.SubExprs)
	w.WriteString(")")
}

func decompileFuncExpr(w *bytes.Buffer, prefix string, e *ast.FuncExpr, deep int) {
	w.WriteString(prefix)
	w.WriteString("func " + e.Name + "(")
	//joinStr(w, e.Params)
	w.WriteString(") {\n")
	decompileStmt(w, e.Stmt, deep+1)
	w.WriteString("}\n")
}

func decompileConstExpr(w *bytes.Buffer, prefix string, e *ast.ConstExpr) {
	w.WriteString(prefix)
	w.WriteString(e.Value)
}

func decompileMapExpr(w *bytes.Buffer, prefix string, e *ast.MapExpr) {
	w.WriteString(prefix)
	w.WriteString(`{`)
	for i, k := range e.Keys {
		v := e.Values[i]
		decompileExpr(w, k, 0)
		w.WriteString(": ")
		decompileExpr(w, v, 0)
		w.WriteString("\n")
	}
	w.WriteString(`}`)
}

func decompileArrayExpr(w *bytes.Buffer, prefix string, e *ast.ArrayExpr) {
	w.WriteString(prefix)
	w.WriteString("[")
	joinExpr(w, e.Exprs)
	w.WriteString("]")
}

func decompileStringExpr(w *bytes.Buffer, prefix string, e *ast.StringExpr) {
	w.WriteString(prefix)
	w.WriteString(`"` + e.Lit + `"`)
}

func decompileCallExpr(w *bytes.Buffer, prefix string, e *ast.CallExpr) {
	w.WriteString(prefix)
	w.WriteString(e.Name + "(")
	joinExpr(w, e.SubExprs)
	w.WriteString(")")
}

func decompileAssocExpr(w *bytes.Buffer, prefix string, e *ast.AssocExpr) {
	w.WriteString(prefix)
	decompileExpr(w, e.Lhs, 0)
	if e.Operator == "++" || e.Operator == "--" {
		w.WriteString(e.Operator)
	} else {
		w.WriteString(" " + e.Operator + " ")
	}
	decompileExpr(w, e.Rhs, 0)
}

func decompileBinOpExpr(w *bytes.Buffer, prefix string, e *ast.BinOpExpr) {
	w.WriteString(prefix)
	decompileExpr(w, e.Lhs, 0)
	w.WriteString(" " + e.Operator + " ")
	decompileExpr(w, e.Rhs, 0)
}

func decompileNumberExpr(w *bytes.Buffer, prefix string, e *ast.NumberExpr) {
	w.WriteString(prefix)
	w.WriteString(e.Lit)
}

func decompileIdentExpr(w *bytes.Buffer, prefix string, e *ast.IdentExpr) {
	w.WriteString(prefix)
	w.WriteString(e.Lit)
}
