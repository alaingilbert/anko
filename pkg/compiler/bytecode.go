package compiler

import "fmt"

type bytecode byte

const (
	magic   = "anko bytecode"
	version = 1

	NilBytecode            bytecode = 50
	StmtsStmtBytecode      bytecode = 51 // Stmts
	ExprStmtBytecode       bytecode = 52
	VarStmtBytecode        bytecode = 53
	LetsStmtBytecode       bytecode = 54
	LetMapItemStmtBytecode bytecode = 55
	IfStmtBytecode         bytecode = 56
	TryStmtBytecode        bytecode = 57
	LoopStmtBytecode       bytecode = 58
	ForStmtBytecode        bytecode = 59
	CForStmtBytecode       bytecode = 60
	ThrowStmtBytecode      bytecode = 61
	ModuleStmtBytecode     bytecode = 62
	SelectStmtBytecode     bytecode = 63
	SelectBodyStmtBytecode bytecode = 64
	SelectCaseStmtBytecode bytecode = 65
	SwitchStmtBytecode     bytecode = 66
	SwitchCaseStmtBytecode bytecode = 67
	//SwitchBodyStmtBytecode      bytecode = 105
	GoroutineStmtBytecode       bytecode = 68
	DeferStmtBytecode           bytecode = 107
	DeleteStmtBytecode          bytecode = 69
	CloseStmtBytecode           bytecode = 70
	BreakStmtBytecode           bytecode = 71
	ContinueStmtBytecode        bytecode = 72
	ReturnStmtBytecode          bytecode = 73
	OpExprBytecode              bytecode = 74 // Exprs
	StringExprBytecode          bytecode = 75
	NumberExprBytecode          bytecode = 76
	CloseExprBytecode           bytecode = 106
	DeleteExprBytecode          bytecode = 77
	IdentExprBytecode           bytecode = 78
	LiteralExprBytecode         bytecode = 79
	ArrayExprBytecode           bytecode = 80
	MapExprBytecode             bytecode = 81
	DerefExprBytecode           bytecode = 82
	AddrExprBytecode            bytecode = 83
	UnaryExprBytecode           bytecode = 84
	ParenExprBytecode           bytecode = 85
	MemberExprBytecode          bytecode = 86
	ItemExprBytecode            bytecode = 87
	SliceExprBytecode           bytecode = 88
	AssocExprBytecode           bytecode = 89 // "+=" "-=" "*=" "/=" "&=" "|=" "++" "--"
	LetsExprBytecode            bytecode = 90
	TernaryOpExprBytecode       bytecode = 91 // "? :"
	NilCoalescingOpExprBytecode bytecode = 92 // "??"
	LenExprBytecode             bytecode = 93
	//NewExprBytecode             bytecode = 94
	MakeExprBytecode     bytecode = 95
	MakeTypeExprBytecode bytecode = 96
	//MakeChanExprBytecode   bytecode = 97
	ChanExprBytecode       bytecode = 98
	FuncExprBytecode       bytecode = 99
	AnonCallExprBytecode   bytecode = 100
	CallExprBytecode       bytecode = 101
	IncludeExprBytecode    bytecode = 102
	BinaryOperatorBytecode bytecode = 103 // "+" "-" "*" "/" "%" "**" "<<" ">>" "==" "!=" ">" ">=" "<" "<=" "|" "||" "&" "&&"
	ConstExprBytecode      bytecode = 104
)

// String ...
func (b bytecode) String() string {
	switch b {
	case StringExprBytecode:
		return "StringExprBytecode"
	case StmtsStmtBytecode:
		return "StmtsStmtBytecode"
	case ExprStmtBytecode:
		return "ExprStmtBytecode"
	case VarStmtBytecode:
		return "VarStmtBytecode"
	case LetsStmtBytecode:
		return "LetsStmtBytecode"
	case LetMapItemStmtBytecode:
		return "LetMapItemStmtBytecode"
	case IfStmtBytecode:
		return "IfStmtBytecode"
	case TryStmtBytecode:
		return "TryStmtBytecode"
	case LoopStmtBytecode:
		return "LoopStmtBytecode"
	case ForStmtBytecode:
		return "ForStmtBytecode"
	case CForStmtBytecode:
		return "CForStmtBytecode"
	case ThrowStmtBytecode:
		return "ThrowStmtBytecode"
	case ModuleStmtBytecode:
		return "ModuleStmtBytecode"
	case SelectStmtBytecode:
		return "SelectStmtBytecode"
	case SwitchStmtBytecode:
		return "SwitchStmtBytecode"
	case GoroutineStmtBytecode:
		return "GoroutineStmtBytecode"
	case DeleteStmtBytecode:
		return "DeleteStmtBytecode"
	case CloseStmtBytecode:
		return "CloseStmtBytecode"
	case BreakStmtBytecode:
		return "BreakStmtBytecode"
	case ContinueStmtBytecode:
		return "ContinueStmtBytecode"
	case ReturnStmtBytecode:
		return "ReturnStmtBytecode"
	case OpExprBytecode:
		return "OpExprBytecode"
	case IdentExprBytecode:
		return "IdentExprBytecode"
	case LiteralExprBytecode:
		return "LiteralExprBytecode"
	case ArrayExprBytecode:
		return "ArrayExprBytecode"
	case MapExprBytecode:
		return "MapExprBytecode"
	case DerefExprBytecode:
		return "DerefExprBytecode"
	case AddrExprBytecode:
		return "AddrExprBytecode"
	case UnaryExprBytecode:
		return "UnaryExprBytecode"
	case ParenExprBytecode:
		return "ParenExprBytecode"
	case MemberExprBytecode:
		return "MemberExprBytecode"
	case ItemExprBytecode:
		return "ItemExprBytecode"
	case SliceExprBytecode:
		return "SliceExprBytecode"
	case LetsExprBytecode:
		return "LetsExprBytecode"
	case TernaryOpExprBytecode:
		return "TernaryOpExprBytecode"
	case NilCoalescingOpExprBytecode:
		return "NilCoalescingOpExprBytecode"
	case LenExprBytecode:
		return "LenExprBytecode"
	case MakeExprBytecode:
		return "MakeExprBytecode"
	case MakeTypeExprBytecode:
		return "MakeTypeExprBytecode"
	case ChanExprBytecode:
		return "ChanExprBytecode"
	case FuncExprBytecode:
		return "FuncExprBytecode"
	case AnonCallExprBytecode:
		return "AnonCallExprBytecode"
	case CallExprBytecode:
		return "CallExprBytecode"
	default:
		return fmt.Sprintf("INVALID(%d)", byte(b))
	}
}
