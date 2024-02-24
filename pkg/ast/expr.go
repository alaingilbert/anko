package ast

import (
	"reflect"
)

// Expr provides all of interfaces for expression.
type Expr interface {
	Pos
	expr()
}

// ExprImpl provide commonly implementations for Expr.
type ExprImpl struct {
	PosImpl // ExprImpl provide Pos() function.
}

// expr provide restraint interface.
func (x *ExprImpl) expr() {}

// NumberExpr provide Number expression.
type NumberExpr struct {
	ExprImpl
	Lit string
}

// StringExpr provide String expression.
type StringExpr struct {
	ExprImpl
	Lit string
}

// ArrayExpr provide Array expression.
type ArrayExpr struct {
	ExprImpl
	Exprs    []Expr
	TypeData *TypeStruct
}

// MapExpr provide Map expression.
type MapExpr struct {
	ExprImpl
	Keys     []Expr
	Values   []Expr
	TypeData *TypeStruct
}

// IdentExpr provide identity expression.
type IdentExpr struct {
	ExprImpl
	Lit string
}

// UnaryExpr provide unary minus expression. ex: -1, ^1, ~1.
type UnaryExpr struct {
	ExprImpl
	Operator string
	Expr     Expr
}

// AddrExpr provide referencing address expression.
type AddrExpr struct {
	ExprImpl
	Expr Expr
}

// DerefExpr provide dereferencing address expression.
type DerefExpr struct {
	ExprImpl
	Expr Expr
}

// ParenExpr provide parent block expression.
type ParenExpr struct {
	ExprImpl
	SubExpr Expr
}

// BinOpExpr provide binary operator expression.
type BinOpExpr struct {
	ExprImpl
	Lhs      Expr
	Operator string
	Rhs      Expr
}

// NilCoalescingOpExpr provide if invalid operator expression.
type NilCoalescingOpExpr struct {
	ExprImpl
	Lhs Expr
	Rhs Expr
}

// TernaryOpExpr provide ternary operator expression.
type TernaryOpExpr struct {
	ExprImpl
	Expr Expr
	Lhs  Expr
	Rhs  Expr
}

// Callable ...
type Callable struct {
	ExprImpl
	SubExprs []Expr
	VarArg   bool
	Go       bool
	Defer    bool
}

// CallExpr provide calling expression.
type CallExpr struct {
	*Callable
	Func reflect.Value
	Name string
}

// AnonCallExpr provide anonymous calling expression. ex: func(){}().
type AnonCallExpr struct {
	*Callable
	Expr Expr
}

// MemberExpr provide expression to refer member.
type MemberExpr struct {
	ExprImpl
	Expr Expr
	Name string
}

// ItemExpr provide expression to refer Map/Array item.
type ItemExpr struct {
	ExprImpl
	Value Expr
	Index Expr
}

// SliceExpr provide expression to refer slice of Array.
type SliceExpr struct {
	ExprImpl
	Value Expr
	Begin Expr
	End   Expr
}

// FuncReturnValuesExpr ...
type FuncReturnValuesExpr struct {
	TypeData *TypeStruct
}

// ParamExpr ...
type ParamExpr struct {
	Name     string
	TypeData *TypeStruct
}

// FuncExpr provide function expression.
type FuncExpr struct {
	ExprImpl
	Name    string
	Stmt    Stmt
	Params  []*ParamExpr
	Returns []*FuncReturnValuesExpr
	VarArg  bool
}

// LetsExpr provide multiple expression of let.
type LetsExpr struct {
	ExprImpl
	Lhss     []Expr
	Operator string
	Rhss     []Expr
}

// AssocExpr provide expression to assoc operation.
type AssocExpr struct {
	ExprImpl
	Lhs      Expr
	Operator string
	Rhs      Expr
}

// ConstExpr provide expression for constant variable.
type ConstExpr struct {
	ExprImpl
	Value string
}

// ChanExpr provide chan expression.
type ChanExpr struct {
	ExprImpl
	Lhs Expr
	Rhs Expr
}

// ArrayCount is used in MakeExpr to provide Dimensions
type ArrayCount struct {
	Count int
}

// MakeExpr provide expression to make instance.
type MakeExpr struct {
	ExprImpl
	TypeData *TypeStruct
	LenExpr  Expr
	CapExpr  Expr
}

// MakeTypeExpr provide expression to make type.
type MakeTypeExpr struct {
	ExprImpl
	Name string
	Type Expr
}

// LenExpr provide expression to get length of array, map, etc.
type LenExpr struct {
	ExprImpl
	Expr Expr
}

// CloseExpr provide close expression
type CloseExpr struct {
	ExprImpl
	WhatExpr Expr
}

// DeleteExpr provide delete expression
type DeleteExpr struct {
	ExprImpl
	WhatExpr Expr
	KeyExpr  Expr
}

// IncludeExpr provide in expression
type IncludeExpr struct {
	ExprImpl
	ItemExpr Expr
	ListExpr Expr
}
