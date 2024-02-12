package compiler

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/alaingilbert/anko/pkg/parser"
	"github.com/alaingilbert/anko/pkg/utils"
	"reflect"

	"github.com/alaingilbert/anko/pkg/ast"
)

func Compile(src string) ([]byte, error) {
	stmts, err := parser.ParseSrc(src)
	if err != nil {
		return nil, err
	}
	return EncodeStmts(stmts)
}

func EncodeStmts(stmts []ast.Stmt) ([]byte, error) {
	e := new(bytes.Buffer)
	writeMagic(e, magic)
	writeVersion(e, version)
	encodeStmtArray(e, stmts)
	return e.Bytes(), nil
}

func writeMagic(w *bytes.Buffer, magic string) {
	encode(w, magic)
}

func writeVersion(w *bytes.Buffer, version uint16) {
	versionBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(versionBytes, version)
	encode(w, versionBytes)
}

// Encode ...
func encode(w *bytes.Buffer, args ...any) {
	var err error
	for _, v := range args {
		switch v := v.(type) {
		case byte: // uint8
			err = w.WriteByte(v)
		case bytecode:
			err = w.WriteByte(byte(v))
		case string:
			_, err = w.WriteString(v)
		case []byte:
			_, err = w.Write(v)
		case bool:
			err = w.WriteByte(utils.Ternary[uint8](v, 1, 0))
		case int:
			err = gob.NewEncoder(w).Encode(v)
		case int32:
			err = gob.NewEncoder(w).Encode(v)
		case int64:
			err = gob.NewEncoder(w).Encode(v)
		case float64:
			err = gob.NewEncoder(w).Encode(v)
		default:
			fmt.Println(reflect.ValueOf(v).String())
			panic("failed to encode")
		}
		if err != nil {
			panic(err)
		}
	}
}

func encodeString(w *bytes.Buffer, str string) {
	encode(w, int32(len(str)), str)
}

func encodeStringArray(w *bytes.Buffer, strs []string) {
	encode(w, int32(len(strs)))
	for _, str := range strs {
		encodeString(w, str)
	}
}

func encodeBool(w *bytes.Buffer, v bool) {
	encode(w, v)
}

func encodeStmtArray(w *bytes.Buffer, stmts []ast.Stmt) {
	isNil := stmts == nil
	encodeBool(w, isNil)
	if !isNil {
		encode(w, int32(len(stmts)))
		for _, e := range stmts {
			encodeSingleStmt(w, e)
		}
	}
}

func encodeExprArray(w *bytes.Buffer, exprs []ast.Expr) {
	encode(w, int32(len(exprs)))
	for _, e := range exprs {
		encodeExpr(w, e)
	}
}

func encodeExprMap(w *bytes.Buffer, expr map[ast.Expr]ast.Expr) {
	encode(w, int32(len(expr)))
	for k, v := range expr {
		encodeExpr(w, k)
		encodeExpr(w, v)
	}
}

func encodePosImpl(w *bytes.Buffer, p ast.PosImpl) {
	pos := p.Position()
	encode(w, int32(pos.Line))
	encode(w, int32(pos.Column))
}

func encodeExprImpl(w *bytes.Buffer, i ast.ExprImpl) {
	encodePosImpl(w, i.PosImpl)
}

func encodeStmtImpl(w *bytes.Buffer, i ast.StmtImpl) {
	encodePosImpl(w, i.PosImpl)
}

func encodeSingleStmt(w *bytes.Buffer, stmt ast.Stmt) {
	//fmt.Println("encodeSingleStmt", reflect.ValueOf(stmt).String())
	switch stmt := stmt.(type) {
	case nil:
		encode(w, NilBytecode)
	case *ast.ExprStmt:
		encodeExprStmt(w, stmt)
	case *ast.VarStmt:
		encodeVarStmt(w, stmt)
	case *ast.LetsStmt:
		encodeLetsStmt(w, stmt)
	case *ast.LetMapItemStmt:
		encodeLetMapItemStmt(w, stmt)
	case *ast.IfStmt:
		encodeIfStmt(w, stmt)
	case *ast.TryStmt:
		encodeTryStmt(w, stmt)
	case *ast.LoopStmt:
		encodeLoopStmt(w, stmt)
	case *ast.ForStmt:
		encodeForStmt(w, stmt)
	case *ast.CForStmt:
		encodeCForStmt(w, stmt)
	case *ast.ThrowStmt:
		encodeThrowStmt(w, stmt)
	case *ast.ModuleStmt:
		encodeModuleStmt(w, stmt)
	case *ast.SelectStmt:
		encodeSelectStmt(w, stmt)
	case *ast.SelectBodyStmt:
		encodeSelectBodyStmt(w, stmt)
	case *ast.SelectCaseStmt:
		encodeSelectCaseStmt(w, stmt)
	case *ast.SwitchStmt:
		encodeSwitchStmt(w, stmt)
	case *ast.SwitchCaseStmt:
		encodeSwitchCaseStmt(w, stmt)
	case *ast.SwitchBodyStmt:
		encodeSwitchBodyStmt(w, stmt)
	case *ast.GoroutineStmt:
		encodeGoroutineStmt(w, stmt)
	case *ast.DeferStmt:
		encodeDeferStmt(w, stmt)
	case *ast.BreakStmt:
		encodeBreakStmt(w, stmt)
	case *ast.ContinueStmt:
		encodeContinueStmt(w, stmt)
	case *ast.ReturnStmt:
		encodeReturnStmt(w, stmt)
	default:
		panic("failed")
	}
}

func encodeExprStmt(w *bytes.Buffer, stmt *ast.ExprStmt) {
	encode(w, ExprStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
}

func encodeVarStmt(w *bytes.Buffer, stmt *ast.VarStmt) {
	encode(w, VarStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeStringArray(w, stmt.Names)
	encodeExprArray(w, stmt.Exprs)
}

func encodeLetsStmt(w *bytes.Buffer, stmt *ast.LetsStmt) {
	encode(w, LetsStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExprArray(w, stmt.Lhss)
	encodeString(w, stmt.Operator)
	encodeExprArray(w, stmt.Rhss)
}

func encodeLetMapItemStmt(w *bytes.Buffer, stmt *ast.LetMapItemStmt) {
	encode(w, LetMapItemStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Rhs)
	encodeExprArray(w, stmt.Lhss)
}

func encodeIfStmt(w *bytes.Buffer, stmt *ast.IfStmt) {
	encode(w, IfStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.If)
	encodeStmtArray(w, stmt.Then)
	encodeStmtArray(w, stmt.ElseIf)
	encodeStmtArray(w, stmt.Else)
}

func encodeTryStmt(w *bytes.Buffer, stmt *ast.TryStmt) {
	encode(w, TryStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeString(w, stmt.Var)
	encodeStmtArray(w, stmt.Try)
	encodeStmtArray(w, stmt.Catch)
	encodeStmtArray(w, stmt.Finally)
}

func encodeLoopStmt(w *bytes.Buffer, stmt *ast.LoopStmt) {
	encode(w, LoopStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
	encodeStmtArray(w, stmt.Stmts)
}

func encodeForStmt(w *bytes.Buffer, stmt *ast.ForStmt) {
	encode(w, ForStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Value)
	encodeStmtArray(w, stmt.Stmts)
	encodeStringArray(w, stmt.Vars)
}

func encodeCForStmt(w *bytes.Buffer, stmt *ast.CForStmt) {
	encode(w, CForStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeStmtArray(w, stmt.Stmts)
	encodeSingleStmt(w, stmt.Stmt1)
	encodeExpr(w, stmt.Expr2)
	encodeExpr(w, stmt.Expr3)
}

func encodeThrowStmt(w *bytes.Buffer, stmt *ast.ThrowStmt) {
	encode(w, ThrowStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
}

func encodeModuleStmt(w *bytes.Buffer, stmt *ast.ModuleStmt) {
	encode(w, ModuleStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeString(w, stmt.Name)
	encodeStmtArray(w, stmt.Stmts)
}

func encodeSelectStmt(w *bytes.Buffer, stmt *ast.SelectStmt) {
	encode(w, SelectStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeSingleStmt(w, stmt.Body)
}

func encodeSelectBodyStmt(w *bytes.Buffer, stmt *ast.SelectBodyStmt) {
	encode(w, SelectBodyStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeStmtArray(w, stmt.Default)
	encodeStmtArray(w, stmt.Cases)
}

func encodeSelectCaseStmt(w *bytes.Buffer, stmt *ast.SelectCaseStmt) {
	encode(w, SelectCaseStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeSingleStmt(w, stmt.Expr)
	encodeStmtArray(w, stmt.Stmts)
}

func encodeSwitchStmt(w *bytes.Buffer, stmt *ast.SwitchStmt) {
	encode(w, SwitchStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
	encodeSingleStmt(w, stmt.Body)
}

func encodeSwitchCaseStmt(w *bytes.Buffer, stmt *ast.SwitchCaseStmt) {
	encode(w, SwitchCaseStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeStmtArray(w, stmt.Stmts)
	encodeExprArray(w, stmt.Exprs)
}

func encodeSwitchBodyStmt(w *bytes.Buffer, stmt *ast.SwitchBodyStmt) {
	encode(w, SwitchBodyStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeStmtArray(w, stmt.Cases)
	encodeStmtArray(w, stmt.Default)
}

func encodeGoroutineStmt(w *bytes.Buffer, stmt *ast.GoroutineStmt) {
	encode(w, GoroutineStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
}

func encodeDeferStmt(w *bytes.Buffer, stmt *ast.DeferStmt) {
	encode(w, DeferStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
}

func encodeBreakStmt(w *bytes.Buffer, stmt *ast.BreakStmt) {
	encode(w, BreakStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
}

func encodeContinueStmt(w *bytes.Buffer, stmt *ast.ContinueStmt) {
	encode(w, ContinueStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
}

func encodeReturnStmt(w *bytes.Buffer, stmt *ast.ReturnStmt) {
	encode(w, ReturnStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExprArray(w, stmt.Exprs)
}

func encodeExpr(w *bytes.Buffer, expr ast.Expr) {
	//fmt.Println("encodeExpr", reflect.ValueOf(expr).String())
	switch expr := expr.(type) {
	case nil:
		encode(w, NilBytecode)
	case *ast.NumberExpr:
		encodeNumberExpr(w, expr)
	case *ast.IdentExpr:
		encodeIdentExpr(w, expr)
	case *ast.StringExpr:
		encodeStringExpr(w, expr)
	case *ast.ArrayExpr:
		encodeArrayExpr(w, expr)
	case *ast.MapExpr:
		encodeMapExpr(w, expr)
	case *ast.DerefExpr:
		encodeDerefExpr(w, expr)
	case *ast.AddrExpr:
		encodeAddrExpr(w, expr)
	case *ast.UnaryExpr:
		encodeUnaryExpr(w, expr)
	case *ast.ParenExpr:
		encodeParenExpr(w, expr)
	case *ast.MemberExpr:
		encodeMemberExpr(w, expr)
	case *ast.ItemExpr:
		encodeItemExpr(w, expr)
	case *ast.SliceExpr:
		encodeSliceExpr(w, expr)
	case *ast.AssocExpr:
		encodeAssocExpr(w, expr)
	case *ast.LetsExpr:
		encodeLetsExpr(w, expr)
	case *ast.BinOpExpr:
		encodeBinOpExpr(w, expr)
	case *ast.ConstExpr:
		encodeConstExpr(w, expr)
	case *ast.TernaryOpExpr:
		encodeTernaryOpExpr(w, expr)
	case *ast.NilCoalescingOpExpr:
		encodeNilCoalescingOpExpr(w, expr)
	case *ast.LenExpr:
		encodeLenExpr(w, expr)
	case *ast.NewExpr:
		encodeNewExpr(w, expr)
	case *ast.MakeExpr:
		encodeMakeExpr(w, expr)
	case *ast.MakeTypeExpr:
		encodeMakeTypeExpr(w, expr)
	case *ast.MakeChanExpr:
		encodeMakeChanExpr(w, expr)
	case *ast.ChanExpr:
		encodeChanExpr(w, expr)
	case *ast.FuncExpr:
		encodeFuncExpr(w, expr)
	case *ast.AnonCallExpr:
		encodeAnonCallExpr(w, expr)
	case *ast.CallExpr:
		encodeCallExpr(w, expr)
	case *ast.CloseExpr:
		encodeCloseExpr(w, expr)
	case *ast.DeleteExpr:
		encodeDeleteExpr(w, expr)
	case *ast.IncludeExpr:
		encodeIncludeExpr(w, expr)
	default:
		fmt.Println("encodeExpr", reflect.ValueOf(expr).String())
		panic("failed")
	}
}

func encodeNumberExpr(w *bytes.Buffer, expr *ast.NumberExpr) {
	encode(w, NumberExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Lit)
}

func encodeIdentExpr(w *bytes.Buffer, expr *ast.IdentExpr) {
	encode(w, IdentExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Lit)
}

func encodeStringExpr(w *bytes.Buffer, expr *ast.StringExpr) {
	encode(w, StringExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Lit)
}

func encodeArrayExpr(w *bytes.Buffer, expr *ast.ArrayExpr) {
	encode(w, ArrayExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExprArray(w, expr.Exprs)
}

func encodeMapExpr(w *bytes.Buffer, expr *ast.MapExpr) {
	encode(w, MapExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExprMap(w, expr.MapExpr)
}

func encodeDerefExpr(w *bytes.Buffer, expr *ast.DerefExpr) {
	encode(w, DerefExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Expr)
}

func encodeAddrExpr(w *bytes.Buffer, expr *ast.AddrExpr) {
	encode(w, AddrExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Expr)
}

func encodeUnaryExpr(w *bytes.Buffer, expr *ast.UnaryExpr) {
	encode(w, UnaryExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Operator)
	encodeExpr(w, expr.Expr)
}

func encodeParenExpr(w *bytes.Buffer, expr *ast.ParenExpr) {
	encode(w, ParenExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.SubExpr)
}

func encodeMemberExpr(w *bytes.Buffer, expr *ast.MemberExpr) {
	encode(w, MemberExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Name)
	encodeExpr(w, expr.Expr)
}

func encodeItemExpr(w *bytes.Buffer, expr *ast.ItemExpr) {
	encode(w, ItemExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Index)
	encodeExpr(w, expr.Value)
}

func encodeSliceExpr(w *bytes.Buffer, expr *ast.SliceExpr) {
	encode(w, SliceExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Value)
	encodeExpr(w, expr.Begin)
	encodeExpr(w, expr.End)
}

func encodeAssocExpr(w *bytes.Buffer, expr *ast.AssocExpr) {
	encode(w, AssocExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Operator)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeLetsExpr(w *bytes.Buffer, expr *ast.LetsExpr) {
	encode(w, LetsExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExprArray(w, expr.Lhss)
	encodeExprArray(w, expr.Rhss)
}

func encodeBinOpExpr(w *bytes.Buffer, expr *ast.BinOpExpr) {
	encode(w, BinaryOperatorBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Operator)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeConstExpr(w *bytes.Buffer, expr *ast.ConstExpr) {
	encode(w, ConstExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Value)
}

func encodeTernaryOpExpr(w *bytes.Buffer, expr *ast.TernaryOpExpr) {
	encode(w, TernaryOpExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Expr)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeNilCoalescingOpExpr(w *bytes.Buffer, expr *ast.NilCoalescingOpExpr) {
	encode(w, NilCoalescingOpExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeLenExpr(w *bytes.Buffer, expr *ast.LenExpr) {
	encode(w, LenExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Expr)
}

func encodeNewExpr(w *bytes.Buffer, expr *ast.NewExpr) {
	encode(w, NewExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Type)
}

func encodeMakeExpr(w *bytes.Buffer, expr *ast.MakeExpr) {
	encode(w, MakeExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encode(w, int32(expr.Dimensions))
	encodeString(w, expr.Type)
	encodeExpr(w, expr.LenExpr)
	encodeExpr(w, expr.CapExpr)
}

func encodeMakeTypeExpr(w *bytes.Buffer, expr *ast.MakeTypeExpr) {
	encode(w, MakeTypeExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Name)
	encodeExpr(w, expr.Type)
}

func encodeMakeChanExpr(w *bytes.Buffer, expr *ast.MakeChanExpr) {
	encode(w, MakeChanExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Type)
	encodeExpr(w, expr.SizeExpr)
}

func encodeChanExpr(w *bytes.Buffer, expr *ast.ChanExpr) {
	encode(w, ChanExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeFuncExpr(w *bytes.Buffer, expr *ast.FuncExpr) {
	encode(w, FuncExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Name)
	encodeBool(w, expr.VarArg)
	encodeStringArray(w, expr.Params)
	encodeStmtArray(w, expr.Stmts)
}

func encodeAnonCallExpr(w *bytes.Buffer, expr *ast.AnonCallExpr) {
	encode(w, AnonCallExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encode(w, expr.VarArg)
	encode(w, expr.Go)
	encodeExpr(w, expr.Expr)
	encodeExprArray(w, expr.SubExprs)
}

func encodeCallExpr(w *bytes.Buffer, expr *ast.CallExpr) {
	encode(w, CallExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Name)
	encodeBool(w, expr.VarArg)
	encodeExprArray(w, expr.SubExprs)
	encodeBool(w, expr.Go)
	encodeBool(w, expr.Defer)
}

func encodeCloseExpr(w *bytes.Buffer, expr *ast.CloseExpr) {
	encode(w, CloseExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.WhatExpr)
}

func encodeDeleteExpr(w *bytes.Buffer, expr *ast.DeleteExpr) {
	encode(w, DeleteExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.KeyExpr)
	encodeExpr(w, expr.WhatExpr)
}

func encodeIncludeExpr(w *bytes.Buffer, expr *ast.IncludeExpr) {
	encode(w, IncludeExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.ItemExpr)
	encodeExpr(w, expr.ListExpr)
}
