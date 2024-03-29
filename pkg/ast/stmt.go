package ast

// Stmt provides all of interfaces for statement.
type Stmt interface {
	Pos
	stmt()
	SetLabel(string)
	GetLabel() string
}

// StmtImpl provide commonly implementations for Stmt..
type StmtImpl struct {
	PosImpl // StmtImpl provide Pos() function.
	Label   string
}

// GetLabel ...
func (x *StmtImpl) GetLabel() string {
	return x.Label
}

// SetLabel ...
func (x *StmtImpl) SetLabel(name string) {
	x.Label = name
}

// stmt provide restraint interface.
func (x *StmtImpl) stmt() {}

// StmtsStmt provides statements.
type StmtsStmt struct {
	StmtImpl
	Stmts []Stmt
}

// ExprStmt provide expression statement.
type ExprStmt struct {
	StmtImpl
	Expr Expr
}

// DbgStmt provide statement to debug anything
type DbgStmt struct {
	StmtImpl
	Expr     Expr
	TypeData *TypeStruct
}

// LabelStmt provide label statement.
type LabelStmt struct {
	StmtImpl
	Name string
	Stmt Stmt
}

// IfStmt provide "if/else" statement.
type IfStmt struct {
	StmtImpl
	If   Expr
	Then Stmt
	Else Stmt
}

// TryStmt provide "try/catch/finally" statement.
type TryStmt struct {
	StmtImpl
	Try     Stmt
	Var     string
	Catch   Stmt
	Finally Stmt
}

// ForStmt provide "for in" expression statement.
type ForStmt struct {
	StmtImpl
	Vars  []string
	Value Expr
	Stmt  Stmt
}

// CForStmt provide C-style "for (;;)" expression statement.
type CForStmt struct {
	StmtImpl
	Stmt1 Stmt
	Expr2 Expr
	Expr3 Expr
	Stmt  Stmt
}

// LoopStmt provide "for expr" expression statement.
type LoopStmt struct {
	StmtImpl
	Expr Expr
	Stmt Stmt
}

// BreakStmt provide "break" expression statement.
type BreakStmt struct {
	StmtImpl
	Label string
}

// ContinueStmt provide "continue" expression statement.
type ContinueStmt struct {
	StmtImpl
	Label string
}

// ReturnStmt provide "return" expression statement.
type ReturnStmt struct {
	StmtImpl
	Exprs Expr
}

// ThrowStmt provide "throw" expression statement.
type ThrowStmt struct {
	StmtImpl
	Expr Expr
}

// ModuleStmt provide "module" expression statement.
type ModuleStmt struct {
	StmtImpl
	Name string
	Stmt Stmt
}

// SelectStmt provide switch statement.
type SelectStmt struct {
	StmtImpl
	Body Stmt
}

// SelectBodyStmt provide switch case statements and default statement.
type SelectBodyStmt struct {
	StmtImpl
	Cases   []Stmt
	Default Stmt
}

// SelectCaseStmt provide switch case statement.
type SelectCaseStmt struct {
	StmtImpl
	Expr Stmt
	Stmt Stmt
}

// SwitchStmt provide switch statement.
type SwitchStmt struct {
	StmtImpl
	Expr    Expr
	Cases   []Stmt
	Default Stmt
}

// SwitchCaseStmt provide switch case statement.
type SwitchCaseStmt struct {
	StmtImpl
	Exprs *ExprsExpr
	Stmt  Stmt
}

// VarStmt provide statement to let variables in current scope.
type VarStmt struct {
	StmtImpl
	Names []string
	Exprs []Expr
}

// LetsStmt provide multiple statement of let.
type LetsStmt struct {
	StmtImpl
	Lhss     Expr
	Operator string
	Rhss     Expr
	Typed    bool
	Mutable  bool
}

// LetMapItemStmt provide statement of let for map item.
type LetMapItemStmt struct {
	StmtImpl
	Lhss Expr
	Rhs  Expr
}

// GoroutineStmt provide statement of groutine.
type GoroutineStmt struct {
	StmtImpl
	Expr Expr
}

// DeferStmt provide statement of defer.
type DeferStmt struct {
	StmtImpl
	Expr Expr
}
