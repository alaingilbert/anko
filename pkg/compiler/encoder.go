package compiler

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/alaingilbert/anko/pkg/parser"
	"github.com/alaingilbert/anko/pkg/utils"
	"math/rand"
	"reflect"

	"github.com/alaingilbert/anko/pkg/ast"
)

func Compile(src string, obfuscate bool) ([]byte, error) {
	stmt, err := parser.ParseSrc(src)
	if err != nil {
		return nil, err
	}
	return EncodeStmts(stmt, obfuscate)
}

type Encoder struct {
	*bytes.Buffer
	obfuscate  bool
	strings    map[string]int
	stringsArr []string
	stringsIdx int
	idents     map[string]string
}

func NewEncoder(obfuscate bool) *Encoder {
	e := new(Encoder)
	e.obfuscate = obfuscate
	e.Buffer = new(bytes.Buffer)
	e.strings = make(map[string]int)
	e.idents = make(map[string]string)
	return e
}

func EncodeStmts(stmt ast.Stmt, obfuscate bool) ([]byte, error) {
	b := NewEncoder(obfuscate)
	encodeSingleStmt(b, stmt)

	e := NewEncoder(obfuscate)
	writeMagic(e, magic)
	writeVersion(e, version)
	encode(e, b.stringsIdx)
	for _, s := range b.stringsArr {
		encode(e, []byte(s))
	}
	encode(e, b.Bytes())

	return e.Bytes(), nil
}

func writeMagic(w *Encoder, magic string) {
	encode(w, magic)
}

func writeVersion(w *Encoder, version uint16) {
	versionBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(versionBytes, version)
	encode(w, versionBytes)
}

// Encode ...
func encode(w *Encoder, args ...any) {
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

func encodeString(w *Encoder, str string) {
	strIdx := w.stringsIdx
	if el, ok := w.strings[str]; !ok {
		w.strings[str] = w.stringsIdx
		w.stringsIdx += len(str)
		w.stringsArr = append(w.stringsArr, str)
	} else {
		strIdx = el
	}
	encode(w, strIdx, int32(len(str)))
}

func encodeStringArray(w *Encoder, strs []string) {
	encode(w, int32(len(strs)))
	for _, str := range strs {
		encodeString(w, str)
	}
}

func encodeBool(w *Encoder, v bool) {
	encode(w, v)
}

func encodeStmtArray(w *Encoder, stmts []ast.Stmt) {
	isNil := stmts == nil
	encodeBool(w, isNil)
	if !isNil {
		encode(w, int32(len(stmts)))
		for _, e := range stmts {
			encodeSingleStmt(w, e)
		}
	}
}

func encodeExprArray(w *Encoder, exprs []ast.Expr) {
	encode(w, int32(len(exprs)))
	for _, e := range exprs {
		encodeExpr(w, e)
	}
}

func encodeExprMap(w *Encoder, expr map[ast.Expr]ast.Expr) {
	encode(w, int32(len(expr)))
	for k, v := range expr {
		encodeExpr(w, k)
		encodeExpr(w, v)
	}
}

func encodePosImpl(w *Encoder, p ast.PosImpl) {
	pos := p.Position()
	encode(w, int32(pos.Line))
	encode(w, int32(pos.Column))
}

func encodeExprImpl(w *Encoder, i ast.ExprImpl) {
	encodePosImpl(w, i.PosImpl)
}

func encodeStmtImpl(w *Encoder, i ast.StmtImpl) {
	encodePosImpl(w, i.PosImpl)
}

func encodeSingleStmt(w *Encoder, stmt ast.Stmt) {
	//fmt.Println("encodeSingleStmt", reflect.ValueOf(stmt).String())
	switch stmt := stmt.(type) {
	case nil:
		encode(w, NilBytecode)
	case *ast.StmtsStmt:
		encodeStmtsStmt(w, stmt)
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

func encodeStmtsStmt(w *Encoder, stmt *ast.StmtsStmt) {
	encode(w, StmtsStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeStmtArray(w, stmt.Stmts)
}

func encodeExprStmt(w *Encoder, stmt *ast.ExprStmt) {
	encode(w, ExprStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
}

func encodeVarStmt(w *Encoder, stmt *ast.VarStmt) {
	encode(w, VarStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeStringArray(w, stmt.Names)
	encodeExprArray(w, stmt.Exprs)
}

func encodeLetsStmt(w *Encoder, stmt *ast.LetsStmt) {
	encode(w, LetsStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExprArray(w, stmt.Lhss)
	encodeString(w, stmt.Operator)
	encodeExprArray(w, stmt.Rhss)
}

func encodeLetMapItemStmt(w *Encoder, stmt *ast.LetMapItemStmt) {
	encode(w, LetMapItemStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Rhs)
	encodeExprArray(w, stmt.Lhss)
}

func encodeIfStmt(w *Encoder, stmt *ast.IfStmt) {
	encode(w, IfStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.If)
	encodeSingleStmt(w, stmt.Then)
	encodeStmtArray(w, stmt.ElseIf)
	encodeSingleStmt(w, stmt.Else)
}

func encodeTryStmt(w *Encoder, stmt *ast.TryStmt) {
	encode(w, TryStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeString(w, stmt.Var)
	encodeSingleStmt(w, stmt.Try)
	encodeSingleStmt(w, stmt.Catch)
	encodeSingleStmt(w, stmt.Finally)
}

func encodeLoopStmt(w *Encoder, stmt *ast.LoopStmt) {
	encode(w, LoopStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
	encodeSingleStmt(w, stmt.Stmt)
}

func encodeForStmt(w *Encoder, stmt *ast.ForStmt) {
	encode(w, ForStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Value)
	encodeSingleStmt(w, stmt.Stmt)
	encodeStringArray(w, stmt.Vars)
}

func encodeCForStmt(w *Encoder, stmt *ast.CForStmt) {
	encode(w, CForStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeSingleStmt(w, stmt.Stmt)
	encodeSingleStmt(w, stmt.Stmt1)
	encodeExpr(w, stmt.Expr2)
	encodeExpr(w, stmt.Expr3)
}

func encodeThrowStmt(w *Encoder, stmt *ast.ThrowStmt) {
	encode(w, ThrowStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
}

func encodeModuleStmt(w *Encoder, stmt *ast.ModuleStmt) {
	encode(w, ModuleStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeString(w, stmt.Name)
	encodeSingleStmt(w, stmt.Stmt)
}

func encodeSelectStmt(w *Encoder, stmt *ast.SelectStmt) {
	encode(w, SelectStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeSingleStmt(w, stmt.Body)
}

func encodeSelectBodyStmt(w *Encoder, stmt *ast.SelectBodyStmt) {
	encode(w, SelectBodyStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeSingleStmt(w, stmt.Default)
	encodeStmtArray(w, stmt.Cases)
}

func encodeSelectCaseStmt(w *Encoder, stmt *ast.SelectCaseStmt) {
	encode(w, SelectCaseStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeSingleStmt(w, stmt.Expr)
	encodeSingleStmt(w, stmt.Stmt)
}

func encodeSwitchStmt(w *Encoder, stmt *ast.SwitchStmt) {
	encode(w, SwitchStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
	encodeStmtArray(w, stmt.Cases)
	encodeSingleStmt(w, stmt.Default)
}

func encodeSwitchCaseStmt(w *Encoder, stmt *ast.SwitchCaseStmt) {
	encode(w, SwitchCaseStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeSingleStmt(w, stmt.Stmt)
	encodeExprArray(w, stmt.Exprs)
}

func encodeGoroutineStmt(w *Encoder, stmt *ast.GoroutineStmt) {
	encode(w, GoroutineStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
}

func encodeDeferStmt(w *Encoder, stmt *ast.DeferStmt) {
	encode(w, DeferStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExpr(w, stmt.Expr)
}

func encodeBreakStmt(w *Encoder, stmt *ast.BreakStmt) {
	encode(w, BreakStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
}

func encodeContinueStmt(w *Encoder, stmt *ast.ContinueStmt) {
	encode(w, ContinueStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
}

func encodeReturnStmt(w *Encoder, stmt *ast.ReturnStmt) {
	encode(w, ReturnStmtBytecode)
	encodeStmtImpl(w, stmt.StmtImpl)
	encodeExprArray(w, stmt.Exprs)
}

func encodeExpr(w *Encoder, expr ast.Expr) {
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
	case *ast.MakeExpr:
		encodeMakeExpr(w, expr)
	case *ast.MakeTypeExpr:
		encodeMakeTypeExpr(w, expr)
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

func encodeNumberExpr(w *Encoder, expr *ast.NumberExpr) {
	encode(w, NumberExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Lit)
}

func generateToken1() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func encodeIdentExpr(w *Encoder, expr *ast.IdentExpr) {
	lit := expr.Lit
	if w.obfuscate {
		if id, ok := w.idents[expr.Lit]; ok {
			lit = id
		} else {
			lit = "id_" + generateToken1()
			w.idents[expr.Lit] = lit
		}
	}
	encode(w, IdentExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, lit)
}

func encodeStringExpr(w *Encoder, expr *ast.StringExpr) {
	encode(w, StringExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Lit)
}

func encodeArrayExpr(w *Encoder, expr *ast.ArrayExpr) {
	encode(w, ArrayExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExprArray(w, expr.Exprs)
}

func encodeMapExpr(w *Encoder, expr *ast.MapExpr) {
	encode(w, MapExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExprArray(w, expr.Keys)
	encodeExprArray(w, expr.Values)
	encodeTypeStruct(w, expr.TypeData)
}

func encodeDerefExpr(w *Encoder, expr *ast.DerefExpr) {
	encode(w, DerefExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Expr)
}

func encodeAddrExpr(w *Encoder, expr *ast.AddrExpr) {
	encode(w, AddrExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Expr)
}

func encodeUnaryExpr(w *Encoder, expr *ast.UnaryExpr) {
	encode(w, UnaryExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Operator)
	encodeExpr(w, expr.Expr)
}

func encodeParenExpr(w *Encoder, expr *ast.ParenExpr) {
	encode(w, ParenExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.SubExpr)
}

func encodeMemberExpr(w *Encoder, expr *ast.MemberExpr) {
	encode(w, MemberExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Name)
	encodeExpr(w, expr.Expr)
}

func encodeItemExpr(w *Encoder, expr *ast.ItemExpr) {
	encode(w, ItemExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Index)
	encodeExpr(w, expr.Value)
}

func encodeSliceExpr(w *Encoder, expr *ast.SliceExpr) {
	encode(w, SliceExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Value)
	encodeExpr(w, expr.Begin)
	encodeExpr(w, expr.End)
}

func encodeAssocExpr(w *Encoder, expr *ast.AssocExpr) {
	encode(w, AssocExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Operator)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeLetsExpr(w *Encoder, expr *ast.LetsExpr) {
	encode(w, LetsExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExprArray(w, expr.Lhss)
	encodeExprArray(w, expr.Rhss)
}

func encodeBinOpExpr(w *Encoder, expr *ast.BinOpExpr) {
	encode(w, BinaryOperatorBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Operator)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeConstExpr(w *Encoder, expr *ast.ConstExpr) {
	encode(w, ConstExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Value)
}

func encodeTernaryOpExpr(w *Encoder, expr *ast.TernaryOpExpr) {
	encode(w, TernaryOpExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Expr)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeNilCoalescingOpExpr(w *Encoder, expr *ast.NilCoalescingOpExpr) {
	encode(w, NilCoalescingOpExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeLenExpr(w *Encoder, expr *ast.LenExpr) {
	encode(w, LenExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Expr)
}

func encodeMakeExpr(w *Encoder, expr *ast.MakeExpr) {
	encode(w, MakeExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeTypeStruct(w, expr.TypeData)
	encodeExpr(w, expr.LenExpr)
	encodeExpr(w, expr.CapExpr)
}

func encodeTypeStruct(w *Encoder, expr *ast.TypeStruct) {
	isNil := expr == nil
	encodeBool(w, isNil)
	if !isNil {
		encode(w, int(expr.Kind))
		encodeStringArray(w, expr.Env)
		encodeString(w, expr.Name)
		encode(w, expr.Dimensions)
		encodeTypeStruct(w, expr.SubType)
		encodeTypeStruct(w, expr.Key)
		encodeStringArray(w, expr.StructNames)
		encodeTypeStructArray(w, expr.StructTypes)
	}
}

func encodeTypeStructArray(w *Encoder, exprs []*ast.TypeStruct) {
	encode(w, int32(len(exprs)))
	for _, expr := range exprs {
		encodeTypeStruct(w, expr)
	}
}

func encodeMakeTypeExpr(w *Encoder, expr *ast.MakeTypeExpr) {
	encode(w, MakeTypeExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Name)
	encodeExpr(w, expr.Type)
}

func encodeChanExpr(w *Encoder, expr *ast.ChanExpr) {
	encode(w, ChanExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.Lhs)
	encodeExpr(w, expr.Rhs)
}

func encodeFuncExpr(w *Encoder, expr *ast.FuncExpr) {
	encode(w, FuncExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Name)
	encodeBool(w, expr.VarArg)
	encodeStringArray(w, expr.Params)
	encodeSingleStmt(w, expr.Stmt)
}

func encodeAnonCallExpr(w *Encoder, expr *ast.AnonCallExpr) {
	encode(w, AnonCallExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encode(w, expr.VarArg)
	encode(w, expr.Go)
	encodeExpr(w, expr.Expr)
	encodeExprArray(w, expr.SubExprs)
}

func encodeCallExpr(w *Encoder, expr *ast.CallExpr) {
	encode(w, CallExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeString(w, expr.Name)
	encodeBool(w, expr.VarArg)
	encodeExprArray(w, expr.SubExprs)
	encodeBool(w, expr.Go)
	encodeBool(w, expr.Defer)
}

func encodeCloseExpr(w *Encoder, expr *ast.CloseExpr) {
	encode(w, CloseExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.WhatExpr)
}

func encodeDeleteExpr(w *Encoder, expr *ast.DeleteExpr) {
	encode(w, DeleteExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.KeyExpr)
	encodeExpr(w, expr.WhatExpr)
}

func encodeIncludeExpr(w *Encoder, expr *ast.IncludeExpr) {
	encode(w, IncludeExprBytecode)
	encodeExprImpl(w, expr.ExprImpl)
	encodeExpr(w, expr.ItemExpr)
	encodeExpr(w, expr.ListExpr)
}
