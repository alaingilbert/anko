package compiler

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	"github.com/alaingilbert/anko/pkg/ast"
)

// Decoder ...
type Decoder struct {
	*bytes.Reader
}

func NewDecoder(in []byte) *Decoder {
	d := new(Decoder)
	d.Reader = bytes.NewReader(in)
	return d
}

func (d *Decoder) readMagic() string {
	str := make([]byte, 13)
	_, _ = d.Read(str)
	if string(str) != magic {
		panic("invalid file")
	}
	return string(str)
}

func (d *Decoder) readVersion() {
	versionBytes := make([]byte, 2)
	_, err := d.Read(versionBytes)
	if err != nil {
		panic(err)
	}
	_ = binary.BigEndian.Uint16(versionBytes)
}

func (d *Decoder) readBytecode() bytecode {
	by, err := d.ReadByte()
	if err != nil {
		panic(err)
	}
	return bytecode(by)
}

func (d *Decoder) readString() string {
	var nbChars int32
	_ = gob.NewDecoder(d).Decode(&nbChars)
	str := make([]byte, nbChars)
	_, _ = d.Read(str)
	return string(str)
}

func (d *Decoder) readStringArray() []string {
	out := make([]string, 0)
	nbElems := d.readInt32()
	for i := 0; int32(i) < nbElems; i++ {
		out = append(out, d.readString())
	}
	return out
}

func (d *Decoder) readStmtArray() []ast.Stmt {
	isNil := d.readBool()
	if isNil {
		return nil
	}
	nbElems := d.readInt32()
	out := make([]ast.Stmt, 0)
	for i := 0; int32(i) < nbElems; i++ {
		stmt := decodeSingleStmt(d)
		out = append(out, stmt)
	}
	return out
}

func (d *Decoder) readExprArray() []ast.Expr {
	nbElems := d.readInt32()
	out := make([]ast.Expr, 0)
	for i := 0; int32(i) < nbElems; i++ {
		expr := decodeExpr(d)
		out = append(out, expr)
	}
	return out
}

func (d *Decoder) readInt32() (val int32) {
	if err := gob.NewDecoder(d).Decode(&val); err != nil {
		panic(err)
	}
	return val
}

func (d *Decoder) readFloat64() (val float64) {
	if err := gob.NewDecoder(d).Decode(&val); err != nil {
		panic(err)
	}
	return val
}

func (d *Decoder) readBool() bool {
	by, _ := d.ReadByte()
	if by == 0 {
		return false
	} else if by == 1 {
		return true
	}
	panic("failed")
}

func decodePosImpl(r *Decoder) ast.PosImpl {
	out := ast.PosImpl{}
	line := r.readInt32()
	column := r.readInt32()
	out.SetPosition(ast.Position{Line: int(line), Column: int(column)})
	return out
}

func decodeExprImpl(r *Decoder) ast.ExprImpl {
	out := ast.ExprImpl{}
	out.PosImpl = decodePosImpl(r)
	return out
}

func decodeStmtImpl(r *Decoder) ast.StmtImpl {
	out := ast.StmtImpl{}
	out.PosImpl = decodePosImpl(r)
	return out
}

func Decode(in []byte) []ast.Stmt {
	r := NewDecoder(in)
	r.readMagic()
	r.readVersion()
	return r.readStmtArray()
}

func decodeSingleStmt(r *Decoder) ast.Stmt {
	b := r.readBytecode()
	switch b {
	case NilBytecode:
		return nil
	case ExprStmtBytecode:
		return decodeExprStmt(r)
	case VarStmtBytecode:
		return decodeVarStmt(r)
	case LetsStmtBytecode:
		return decodeLetsStmt(r)
	case LetMapItemStmtBytecode:
		return decodeLetMapItemStmt(r)
	case IfStmtBytecode:
		return decodeIfStmt(r)
	case TryStmtBytecode:
		return decodeTryStmt(r)
	case LoopStmtBytecode:
		return decodeLoopStmt(r)
	case ForStmtBytecode:
		return decodeForStmt(r)
	case CForStmtBytecode:
		return decodeCForStmt(r)
	case ThrowStmtBytecode:
		return decodeThrowStmt(r)
	case ModuleStmtBytecode:
		return decodeModuleStmt(r)
	case SelectStmtBytecode:
		return decodeSelectStmt(r)
	case SelectBodyStmtBytecode:
		return decodeSelectBodyStmt(r)
	case SelectCaseStmtBytecode:
		return decodeSelectCaseStmt(r)
	case SwitchStmtBytecode:
		return decodeSwitchStmt(r)
	case SwitchCaseStmtBytecode:
		return decodeSwitchCaseStmt(r)
	case SwitchBodyStmtBytecode:
		return decodeSwitchBodyStmt(r)
	case GoroutineStmtBytecode:
		return decodeGoroutineStmt(r)
	case DeferStmtBytecode:
		return decodeDeferStmt(r)
	case BreakStmtBytecode:
		return decodeBreakStmt(r)
	case ContinueStmtBytecode:
		return decodeContinueStmt(r)
	case ReturnStmtBytecode:
		return decodeReturnStmt(r)
	default:
		panic(fmt.Sprintf("invalid (%d)", b))
	}
	return &ast.ExprStmt{}
}

func decodeExprStmt(r *Decoder) *ast.ExprStmt {
	out := &ast.ExprStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Expr = decodeExpr(r)
	return out
}

func decodeVarStmt(r *Decoder) *ast.VarStmt {
	out := &ast.VarStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Names = r.readStringArray()
	out.Exprs = r.readExprArray()
	return out
}

func decodeLetsStmt(r *Decoder) *ast.LetsStmt {
	out := &ast.LetsStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Lhss = r.readExprArray()
	out.Operator = r.readString()
	out.Rhss = r.readExprArray()
	return out
}

func decodeLetMapItemStmt(r *Decoder) *ast.LetMapItemStmt {
	out := &ast.LetMapItemStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Rhs = decodeExpr(r)
	out.Lhss = r.readExprArray()
	return out
}

func decodeIfStmt(r *Decoder) *ast.IfStmt {
	out := &ast.IfStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.If = decodeExpr(r)
	out.Then = r.readStmtArray()
	out.ElseIf = r.readStmtArray()
	out.Else = r.readStmtArray()
	return out
}

func decodeTryStmt(r *Decoder) *ast.TryStmt {
	out := &ast.TryStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Var = r.readString()
	out.Try = r.readStmtArray()
	out.Catch = r.readStmtArray()
	out.Finally = r.readStmtArray()
	return out
}

func decodeLoopStmt(r *Decoder) *ast.LoopStmt {
	out := &ast.LoopStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Expr = decodeExpr(r)
	out.Stmts = r.readStmtArray()
	return out
}

func decodeForStmt(r *Decoder) *ast.ForStmt {
	out := &ast.ForStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Value = decodeExpr(r)
	out.Stmts = r.readStmtArray()
	out.Vars = r.readStringArray()
	return out
}

func decodeCForStmt(r *Decoder) *ast.CForStmt {
	out := &ast.CForStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Stmts = r.readStmtArray()
	out.Stmt1 = decodeSingleStmt(r)
	out.Expr2 = decodeExpr(r)
	out.Expr3 = decodeExpr(r)
	return out
}

func decodeThrowStmt(r *Decoder) *ast.ThrowStmt {
	out := &ast.ThrowStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Expr = decodeExpr(r)
	return out
}

func decodeModuleStmt(r *Decoder) *ast.ModuleStmt {
	out := &ast.ModuleStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Name = r.readString()
	out.Stmts = r.readStmtArray()
	return out
}

func decodeSelectStmt(r *Decoder) *ast.SelectStmt {
	out := &ast.SelectStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Body = decodeSingleStmt(r)
	return out
}

func decodeSelectBodyStmt(r *Decoder) *ast.SelectBodyStmt {
	out := &ast.SelectBodyStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Default = r.readStmtArray()
	out.Cases = r.readStmtArray()
	return out
}

func decodeSelectCaseStmt(r *Decoder) *ast.SelectCaseStmt {
	out := &ast.SelectCaseStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Expr = decodeSingleStmt(r)
	out.Stmts = r.readStmtArray()
	return out
}

func decodeSwitchStmt(r *Decoder) *ast.SwitchStmt {
	out := &ast.SwitchStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Expr = decodeExpr(r)
	out.Body = decodeSingleStmt(r)
	return out
}

func decodeSwitchCaseStmt(r *Decoder) *ast.SwitchCaseStmt {
	out := &ast.SwitchCaseStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Stmts = r.readStmtArray()
	out.Exprs = r.readExprArray()
	return out
}

func decodeSwitchBodyStmt(r *Decoder) *ast.SwitchBodyStmt {
	out := &ast.SwitchBodyStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Cases = r.readStmtArray()
	out.Default = r.readStmtArray()
	return out
}

func decodeGoroutineStmt(r *Decoder) *ast.GoroutineStmt {
	out := &ast.GoroutineStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Expr = decodeExpr(r)
	return out
}

func decodeDeferStmt(r *Decoder) *ast.DeferStmt {
	out := &ast.DeferStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Expr = decodeExpr(r)
	return out
}

func decodeBreakStmt(r *Decoder) *ast.BreakStmt {
	out := &ast.BreakStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	return out
}

func decodeContinueStmt(r *Decoder) *ast.ContinueStmt {
	out := &ast.ContinueStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	return out
}

func decodeReturnStmt(r *Decoder) *ast.ReturnStmt {
	out := &ast.ReturnStmt{}
	out.StmtImpl = decodeStmtImpl(r)
	out.Exprs = r.readExprArray()
	return out
}

func decodeIncludeExpr(r *Decoder) *ast.IncludeExpr {
	out := &ast.IncludeExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.ItemExpr = decodeExpr(r)
	out.ListExpr = decodeExpr(r)
	return out
}

func decodeExpr(r *Decoder) ast.Expr {
	b := r.readBytecode()
	switch b {
	case NilBytecode:
		return nil
	case NumberExprBytecode:
		return decodeNumberExpr(r)
	case IdentExprBytecode:
		return decodeIdentExpr(r)
	case StringExprBytecode:
		return decodeStringExpr(r)
	case ArrayExprBytecode:
		return decodeArrayExpr(r)
	case MapExprBytecode:
		return decodeMapExpr(r)
	case DerefExprBytecode:
		return decodeDerefExpr(r)
	case AddrExprBytecode:
		return decodeAddrExpr(r)
	case UnaryExprBytecode:
		return decodeUnaryExpr(r)
	case ParenExprBytecode:
		return decodeParenExpr(r)
	case MemberExprBytecode:
		return decodeMemberExpr(r)
	case ItemExprBytecode:
		return decodeItemExpr(r)
	case SliceExprBytecode:
		return decodeSliceExpr(r)
	case AssocExprBytecode:
		return decodeAssocExpr(r)
	case LetsExprBytecode:
		return decodeLetsExpr(r)
	case BinaryOperatorBytecode:
		return decodeBinOpExpr(r)
	case ConstExprBytecode:
		return decodeConstExprExpr(r)
	case TernaryOpExprBytecode:
		return decodeTernaryOpExpr(r)
	case NilCoalescingOpExprBytecode:
		return decodeNilCoalescingOpExpr(r)
	case LenExprBytecode:
		return decodeLenExpr(r)
	case NewExprBytecode:
		return decodeNewExpr(r)
	case MakeExprBytecode:
		return decodeMakeExpr(r)
	case MakeTypeExprBytecode:
		return decodeMakeTypeExpr(r)
	case MakeChanExprBytecode:
		return decodeMakeChanExpr(r)
	case ChanExprBytecode:
		return decodeChanExpr(r)
	case FuncExprBytecode:
		return decodeFuncExpr(r)
	case CloseExprBytecode:
		return decodeCloseExpr(r)
	case DeleteExprBytecode:
		return decodeDeleteExpr(r)
	case AnonCallExprBytecode:
		return decodeAnonCallExpr(r)
	case CallExprBytecode:
		return decodeCallExpr(r)
	case IncludeExprBytecode:
		return decodeIncludeExpr(r)
	default:
		panic(fmt.Sprintf("invalid (%d)", b))
	}
}

func decodeNumberExpr(r *Decoder) *ast.NumberExpr {
	out := &ast.NumberExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Lit = r.readString()
	return out
}

func decodeIdentExpr(r *Decoder) *ast.IdentExpr {
	out := &ast.IdentExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Lit = r.readString()
	return out
}

func decodeStringExpr(r *Decoder) *ast.StringExpr {
	out := &ast.StringExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Lit = r.readString()
	return out
}

func decodeArrayExpr(r *Decoder) *ast.ArrayExpr {
	out := &ast.ArrayExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Exprs = r.readExprArray()
	return out
}

func decodeMapExpr(r *Decoder) *ast.MapExpr {
	out := &ast.MapExpr{}
	out.ExprImpl = decodeExprImpl(r)
	nb := r.readInt32()
	out.MapExpr = make(map[ast.Expr]ast.Expr, nb)
	var i int32
	for i = 0; i < nb; i++ {
		key := decodeExpr(r)
		val := decodeExpr(r)
		out.MapExpr[key] = val
	}
	return out
}

func decodeDerefExpr(r *Decoder) *ast.DerefExpr {
	out := &ast.DerefExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Expr = decodeExpr(r)
	return out
}

func decodeAddrExpr(r *Decoder) *ast.AddrExpr {
	out := &ast.AddrExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Expr = decodeExpr(r)
	return out
}

func decodeUnaryExpr(r *Decoder) *ast.UnaryExpr {
	out := &ast.UnaryExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Operator = r.readString()
	out.Expr = decodeExpr(r)
	return out
}

func decodeParenExpr(r *Decoder) *ast.ParenExpr {
	out := &ast.ParenExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.SubExpr = decodeExpr(r)
	return out
}

func decodeMemberExpr(r *Decoder) *ast.MemberExpr {
	out := &ast.MemberExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Name = r.readString()
	out.Expr = decodeExpr(r)
	return out
}

func decodeItemExpr(r *Decoder) *ast.ItemExpr {
	out := &ast.ItemExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Index = decodeExpr(r)
	out.Value = decodeExpr(r)
	return out
}

func decodeSliceExpr(r *Decoder) *ast.SliceExpr {
	out := &ast.SliceExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Value = decodeExpr(r)
	out.Begin = decodeExpr(r)
	out.End = decodeExpr(r)
	return out
}

func decodeAssocExpr(r *Decoder) *ast.AssocExpr {
	out := &ast.AssocExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Operator = r.readString()
	out.Lhs = decodeExpr(r)
	out.Rhs = decodeExpr(r)
	return out
}

func decodeLetsExpr(r *Decoder) *ast.LetsExpr {
	out := &ast.LetsExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Lhss = r.readExprArray()
	out.Rhss = r.readExprArray()
	return out
}

func decodeBinOpExpr(r *Decoder) *ast.BinOpExpr {
	out := &ast.BinOpExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Operator = r.readString()
	out.Lhs = decodeExpr(r)
	out.Rhs = decodeExpr(r)
	return out
}

func decodeConstExprExpr(r *Decoder) *ast.ConstExpr {
	out := &ast.ConstExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Value = r.readString()
	return out
}

func decodeTernaryOpExpr(r *Decoder) *ast.TernaryOpExpr {
	out := &ast.TernaryOpExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Expr = decodeExpr(r)
	out.Lhs = decodeExpr(r)
	out.Rhs = decodeExpr(r)
	return out
}

func decodeNilCoalescingOpExpr(r *Decoder) *ast.NilCoalescingOpExpr {
	out := &ast.NilCoalescingOpExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Lhs = decodeExpr(r)
	out.Rhs = decodeExpr(r)
	return out
}

func decodeLenExpr(r *Decoder) *ast.LenExpr {
	out := &ast.LenExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Expr = decodeExpr(r)
	return out
}

func decodeNewExpr(r *Decoder) *ast.NewExpr {
	out := &ast.NewExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Type = r.readString()
	return out
}

func decodeMakeExpr(r *Decoder) *ast.MakeExpr {
	out := &ast.MakeExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Dimensions = int(r.readInt32())
	out.Type = r.readString()
	out.LenExpr = decodeExpr(r)
	out.CapExpr = decodeExpr(r)
	return out
}

func decodeMakeTypeExpr(r *Decoder) *ast.MakeTypeExpr {
	out := &ast.MakeTypeExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Name = r.readString()
	out.Type = decodeExpr(r)
	return out
}

func decodeMakeChanExpr(r *Decoder) *ast.MakeChanExpr {
	out := &ast.MakeChanExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Type = r.readString()
	out.SizeExpr = decodeExpr(r)
	return out
}

func decodeChanExpr(r *Decoder) *ast.ChanExpr {
	out := &ast.ChanExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Lhs = decodeExpr(r)
	out.Rhs = decodeExpr(r)
	return out
}

func decodeFuncExpr(r *Decoder) *ast.FuncExpr {
	out := &ast.FuncExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Name = r.readString()
	out.VarArg = r.readBool()
	out.Params = r.readStringArray()
	out.Stmts = r.readStmtArray()
	return out
}

func decodeCloseExpr(r *Decoder) *ast.CloseExpr {
	out := &ast.CloseExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.WhatExpr = decodeExpr(r)
	return out
}

func decodeDeleteExpr(r *Decoder) *ast.DeleteExpr {
	out := &ast.DeleteExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.KeyExpr = decodeExpr(r)
	out.WhatExpr = decodeExpr(r)
	return out
}

func decodeAnonCallExpr(r *Decoder) *ast.AnonCallExpr {
	out := &ast.AnonCallExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.VarArg = r.readBool()
	out.Go = r.readBool()
	out.Expr = decodeExpr(r)
	out.SubExprs = r.readExprArray()
	return out
}

func decodeCallExpr(r *Decoder) *ast.CallExpr {
	out := &ast.CallExpr{}
	out.ExprImpl = decodeExprImpl(r)
	out.Name = r.readString()
	out.VarArg = r.readBool()
	out.SubExprs = r.readExprArray()
	out.Go = r.readBool()
	out.Defer = r.readBool()
	return out
}
