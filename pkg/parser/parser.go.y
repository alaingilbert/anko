%{
package parser

import (
	"github.com/alaingilbert/anko/pkg/ast"
)

%}

%type<compstmt> compstmt
%type<stmts> stmts
%type<stmt> stmt
%type<stmt_var_or_lets> stmt_var_or_lets
%type<stmt_var> stmt_var
%type<stmt_lets> stmt_lets
%type<stmt_if> stmt_if
%type<stmt_for> stmt_for
%type<stmt_switch> stmt_switch
%type<stmt_switch_cases> stmt_switch_cases
%type<stmt_switch_case> stmt_switch_case
%type<stmt_switch_default> stmt_switch_default
%type<stmt_select> stmt_select
%type<stmt_select_cases> stmt_select_cases
%type<stmt_select_case> stmt_select_case
%type<stmt_select_default> stmt_select_default
%type<expr> expr
%type<exprs> exprs
%type<map_expr> map_expr
%type<expr_idents> expr_idents
%type<expr_type> expr_type
%type<array_count> array_count
%type<expr_slice> expr_slice
%type<expr_ident> expr_ident

%union{
	compstmt               []ast.Stmt
	stmt_var_or_lets       ast.Stmt
	stmt_var               ast.Stmt
	stmt_lets              ast.Stmt
	stmt_if                ast.Stmt
	stmt_for               ast.Stmt
	stmt_switch            ast.Stmt
	stmt_switch_cases      ast.Stmt
	stmt_switch_case       ast.Stmt
	stmt_switch_default    []ast.Stmt
	stmt_select            ast.Stmt
	stmt_select_cases      ast.Stmt
	stmt_select_case       ast.Stmt
	stmt_select_default    []ast.Stmt
	stmts                  []ast.Stmt
	stmt                   ast.Stmt
	expr                   ast.Expr
	exprs                  []ast.Expr
	map_expr               map[ast.Expr]ast.Expr
	expr_idents            []string
	expr_type              string
	tok                    ast.Token
	array_count            ast.ArrayCount
	expr_slice             ast.Expr
	expr_ident             ast.Expr
}

%token<tok> IDENT NUMBER STRING ARRAY VARARG FUNC RETURN VAR THROW IF ELSE FOR IN EQEQ NEQ GE LE OROR ANDAND NEW
            TRUE FALSE NIL NILCOALESCE MODULE TRY CATCH FINALLY PLUSEQ MINUSEQ MULEQ DIVEQ ANDEQ OREQ BREAK
            CONTINUE PLUSPLUS MINUSMINUS POW SHIFTLEFT SHIFTRIGHT SWITCH SELECT CASE DEFAULT GO DEFER CHAN MAKE
            OPCHAN TYPE LEN DELETE CLOSE

%right '='
%right '?' ':'
%right NILCOALESCE
%left OROR
%left ANDAND
%right IN
%left IDENT
%nonassoc EQEQ NEQ ','
%left '>' GE '<' LE SHIFTLEFT SHIFTRIGHT

%left '+' '-' PLUSPLUS MINUSMINUS
%left '*' '/' '%'
%right UNARY

%%

compstmt :
	opt_term
	{
		$$ = nil
	}
	| stmts opt_term
	{
		$$ = $1
	}

stmts :
	opt_term stmt
	{
		if $2 != nil {
			$$ = []ast.Stmt{$2}
		} else {
			$$ = []ast.Stmt{}
		}
		if l, ok := yylex.(*Lexer); ok {
			l.stmts = $$
		}
	}
	| stmts term stmt
	{
		if $3 != nil {
			$$ = append($1, $3)
			if l, ok := yylex.(*Lexer); ok {
				l.stmts = $$
			}
		}
	}

stmt :
	/* nothing */
	{
		$$ = nil
	}
	| stmt_var_or_lets
	{
		$$ = $1
	}
	| BREAK
	{
		$$ = &ast.BreakStmt{}
		$$.SetPosition($1.Position())
	}
	| CONTINUE
	{
		$$ = &ast.ContinueStmt{}
		$$.SetPosition($1.Position())
	}
	| RETURN exprs
	{
		$$ = &ast.ReturnStmt{Exprs: $2}
		$$.SetPosition($1.Position())
	}
	| THROW expr
	{
		$$ = &ast.ThrowStmt{Expr: $2}
		$$.SetPosition($1.Position())
	}
	| MODULE IDENT '{' compstmt '}'
	{
		$$ = &ast.ModuleStmt{Name: $2.Lit, Stmts: $4}
		$$.SetPosition($1.Position())
	}
	| stmt_if
	{
		$$ = $1
		$$.SetPosition($1.Position())
	}
	| stmt_for
	{
		$$ = $1
	}
	| TRY '{' compstmt '}' CATCH IDENT '{' compstmt '}' FINALLY '{' compstmt '}'
	{
		$$ = &ast.TryStmt{Try: $3, Var: $6.Lit, Catch: $8, Finally: $12}
		$$.SetPosition($1.Position())
	}
	| TRY '{' compstmt '}' CATCH '{' compstmt '}' FINALLY '{' compstmt '}'
	{
		$$ = &ast.TryStmt{Try: $3, Catch: $7, Finally: $11}
		$$.SetPosition($1.Position())
	}
	| TRY '{' compstmt '}' CATCH IDENT '{' compstmt '}'
	{
		$$ = &ast.TryStmt{Try: $3, Var: $6.Lit, Catch: $8}
		$$.SetPosition($1.Position())
	}
	| TRY '{' compstmt '}' CATCH '{' compstmt '}'
	{
		$$ = &ast.TryStmt{Try: $3, Catch: $7}
		$$.SetPosition($1.Position())
	}
	| stmt_switch
	{
		$$ = $1
	}
    | stmt_select
    {
        $$ = $1
    }
	| GO IDENT '(' exprs VARARG ')'
	{
		$$ = &ast.GoroutineStmt{Expr: &ast.CallExpr{Name: $2.Lit, SubExprs: $4, VarArg: true, Go: true}}
		$$.SetPosition($2.Position())
	}
	| GO IDENT '(' exprs ')'
	{
		$$ = &ast.GoroutineStmt{Expr: &ast.CallExpr{Name: $2.Lit, SubExprs: $4, Go: true}}
		$$.SetPosition($2.Position())
	}
	| GO expr '(' exprs VARARG ')'
	{
		$$ = &ast.GoroutineStmt{Expr: &ast.AnonCallExpr{Expr: $2, SubExprs: $4, VarArg: true, Go: true}}
		$$.SetPosition($2.Position())
	}
	| GO expr '(' exprs ')'
	{
		$$ = &ast.GoroutineStmt{Expr: &ast.AnonCallExpr{Expr: $2, SubExprs: $4, Go: true}}
		$$.SetPosition($1.Position())
	}
	| DEFER IDENT '(' exprs VARARG ')'
	{
		$$ = &ast.DeferStmt{Expr: &ast.CallExpr{Name: $2.Lit, SubExprs: $4, VarArg: true, Defer: true}}
		$$.SetPosition($2.Position())
	}
	| DEFER IDENT '(' exprs ')'
	{
		$$ = &ast.DeferStmt{Expr: &ast.CallExpr{Name: $2.Lit, SubExprs: $4, Defer: true}}
		$$.SetPosition($2.Position())
	}
	| DEFER expr '(' exprs VARARG ')'
	{
		$$ = &ast.DeferStmt{Expr: &ast.AnonCallExpr{Expr: $2, SubExprs: $4, VarArg: true, Defer: true}}
		$$.SetPosition($2.Position())
	}
	| DEFER expr '(' exprs ')'
	{
		$$ = &ast.DeferStmt{Expr: &ast.AnonCallExpr{Expr: $2, SubExprs: $4, Defer: true}}
		$$.SetPosition($1.Position())
	}
	| expr
	{
		$$ = &ast.ExprStmt{Expr: $1}
		$$.SetPosition($1.Position())
	}

stmt_var_or_lets :
	stmt_var
	{
		$$ = $1
	}
	| stmt_lets
	{
		$$ = $1
	}

stmt_var :
	VAR expr_idents '=' exprs
	{
		$$ = &ast.VarStmt{Names: $2, Exprs: $4}
		$$.SetPosition($1.Position())
	}

stmt_lets :
	expr '=' expr
	{
		$$ = &ast.LetsStmt{Lhss: []ast.Expr{$1}, Operator: "=", Rhss: []ast.Expr{$3}}
	}
	| exprs '=' exprs
	{
		if len($1) == 2 && len($3) == 1 {
			if _, ok := $3[0].(*ast.ItemExpr); ok {
				$$ = &ast.LetMapItemStmt{Lhss: $1, Rhs: $3[0]}
			} else {
				$$ = &ast.LetsStmt{Lhss: $1, Operator: "=", Rhss: $3}
			}
		} else {
			$$ = &ast.LetsStmt{Lhss: $1, Operator: "=", Rhss: $3}
		}
		if len($1) > 0 {
			$$.SetPosition($1[0].Position())
		}
	}

stmt_if :
	IF expr '{' compstmt '}'
	{
		$$ = &ast.IfStmt{If: $2, Then: $4, Else: nil}
		$$.SetPosition($1.Position())
	}
	| stmt_if ELSE IF expr '{' compstmt '}'
	{
		$1.(*ast.IfStmt).ElseIf = append($1.(*ast.IfStmt).ElseIf, &ast.IfStmt{If: $4, Then: $6})
		$$.SetPosition($1.Position())
	}
	| stmt_if ELSE '{' compstmt '}'
	{
		$$.SetPosition($1.Position())
		if $$.(*ast.IfStmt).Else != nil {
			yylex.Error("multiple else statement")
		} else {
			$$.(*ast.IfStmt).Else = $4
		}
	}
stmt_for :
	FOR '{' compstmt '}'
	{
		$$ = &ast.LoopStmt{Stmts: $3}
		$$.SetPosition($1.Position())
	}
	| FOR expr_idents IN expr '{' compstmt '}'
	{
		$$ = &ast.ForStmt{Vars: $2, Value: $4, Stmts: $6}
		$$.SetPosition($1.Position())
	}
	| FOR stmt_var_or_lets ';' expr ';' expr '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Stmt1: $2, Expr2: $4, Expr3: $6, Stmts: $8}
		$$.SetPosition($1.Position())
	}
	| FOR expr '{' compstmt '}'
	{
		$$ = &ast.LoopStmt{Expr: $2, Stmts: $4}
		$$.SetPosition($1.Position())
	}
stmt_select :
	SELECT '{' opt_newlines stmt_select_cases opt_newlines '}'
	{
		$$ = &ast.SelectStmt{Body: $4}
		$$.SetPosition($1.Position())
	}

stmt_select_cases :
    /* nothing */
    {
        $$ = &ast.SelectBodyStmt{}
    }
    | stmt_select_default
    {
        $$ = &ast.SelectBodyStmt{Default: $1}
    }
    | stmt_select_case
    {
        $$ = &ast.SelectBodyStmt{Cases: []ast.Stmt{$1}}
    }
    | stmt_select_cases stmt_select_case
    {
        body := $$.(*ast.SelectBodyStmt)
        body.Cases = append(body.Cases, $2)
    }
    | stmt_select_cases stmt_select_default
    {
        body := $$.(*ast.SelectBodyStmt)
        if body.Default != nil {
            yylex.Error("multiple default statement")
        }
        body.Default = $2
    }


stmt_select_case :
    CASE stmt ':' compstmt
    {
        $$ = &ast.SelectCaseStmt{Expr: $2, Stmts: $4}
        $$.SetPosition($1.Position())
    }

stmt_select_default :
    DEFAULT ':' compstmt
    {
        $$ = $3
    }

stmt_switch :
	SWITCH expr '{' opt_newlines stmt_switch_cases opt_newlines '}'
	{
		$$ = &ast.SwitchStmt{Expr: $2, Body: $5}
		$$.SetPosition($1.Position())
	}

stmt_switch_cases :
	/* nothing */
	{
		$$ = &ast.SwitchBodyStmt{}
	}
	| stmt_switch_default
	{
		$$ = &ast.SwitchBodyStmt{Default: $1}
	}
	| stmt_switch_case
	{
		$$ = &ast.SwitchBodyStmt{Cases: []ast.Stmt{$1}}
	}
	| stmt_switch_cases stmt_switch_case
	{
		body := $$.(*ast.SwitchBodyStmt)
		body.Cases = append(body.Cases, $2)
	}
	| stmt_switch_cases stmt_switch_default
	{
		body := $$.(*ast.SwitchBodyStmt)
		if body.Default != nil {
			yylex.Error("multiple default statement")
		}
		body.Default = $2
	}


stmt_switch_case :
	CASE expr ':' compstmt
	{
		$$ = &ast.SwitchCaseStmt{Exprs: []ast.Expr{$2}, Stmts: $4}
		$$.SetPosition($1.Position())
	}
	| CASE exprs ':' compstmt
	{
		$$ = &ast.SwitchCaseStmt{Exprs: $2, Stmts: $4}
		$$.SetPosition($1.Position())
	}

stmt_switch_default :
	DEFAULT ':' compstmt
	{
		$$ = $3
	}

array_count :
	{
		$$ = ast.ArrayCount{Count: 0}
	}
	| '[' ']'
	{
		$$ = ast.ArrayCount{Count: 1}
	}
	| array_count '[' ']'
	{
		$$.Count = $$.Count + 1
	}

map_expr :
	{
		$$ = make(map[ast.Expr]ast.Expr)
	}
	| expr ':' expr
	{
		mapExpr := make(map[ast.Expr]ast.Expr)
		mapExpr[$1] = $3
		$$ = mapExpr
	}
	| map_expr ',' opt_newlines expr ':' expr
	{
		if len($1) == 0 {
			yylex.Error("syntax error: unexpected ','")
		}
		$1[$4] = $6
	}

expr_idents :
	{
		$$ = []string{}
	}
	| IDENT
	{
		$$ = []string{$1.Lit}
	}
	| expr_idents ',' opt_newlines IDENT
	{
		if len($1) == 0 {
			yylex.Error("syntax error: unexpected ','")
		}
		$$ = append($1, $4.Lit)
	}

expr_type :
	IDENT
	{
		$$ = $1.Lit
	}
	| expr_type '.' IDENT
	{
		$$ = $$ + "." + $3.Lit
	}

exprs :
	{
		$$ = nil
	}
	| expr 
	{
		$$ = []ast.Expr{$1}
	}
	| exprs ',' opt_newlines expr
	{
		if len($1) == 0 {
			yylex.Error("syntax error: unexpected ','")
		}
		$$ = append($1, $4)
	}
	| exprs ',' opt_newlines expr_ident
	{
		if len($1) == 0 {
			yylex.Error("syntax error: unexpected ','")
		}
		$$ = append($1, $4)
	}

expr_slice :
	expr_ident '[' expr ':' expr ']'
	{
		$$ = &ast.SliceExpr{Value: $1, Begin: $3, End: $5}
	}
	| expr_ident '[' expr ':' ']'
	{
		$$ = &ast.SliceExpr{Value: $1, Begin: $3, End: nil}
	}
	| expr_ident '[' ':' expr ']'
	{
		$$ = &ast.SliceExpr{Value: $1, Begin: nil, End: $4}
	}
	| expr '[' expr ':' expr ']'
	{
		$$ = &ast.SliceExpr{Value: $1, Begin: $3, End: $5}
	}
	| expr '[' expr ':' ']'
	{
		$$ = &ast.SliceExpr{Value: $1, Begin: $3, End: nil}
	}
	| expr '[' ':' expr ']'
	{
		$$ = &ast.SliceExpr{Value: $1, Begin: nil, End: $4}
	}

expr :
	expr_ident
	{
		$$ = $1
	}
	| NUMBER
	{
		$$ = &ast.NumberExpr{Lit: $1.Lit}
		$$.SetPosition($1.Position())
	}
	| '-' expr %prec UNARY
	{
		$$ = &ast.UnaryExpr{Operator: "-", Expr: $2}
		$$.SetPosition($2.Position())
	}
	| '!' expr %prec UNARY
	{
		$$ = &ast.UnaryExpr{Operator: "!", Expr: $2}
		$$.SetPosition($2.Position())
	}
	| '^' expr %prec UNARY
	{
		$$ = &ast.UnaryExpr{Operator: "^", Expr: $2}
		$$.SetPosition($2.Position())
	}
	| '&' expr_ident %prec UNARY
	{
		$$ = &ast.AddrExpr{Expr: $2}
		$$.SetPosition($2.Position())
	}
	| '&' expr '.' IDENT %prec UNARY
	{
		$$ = &ast.AddrExpr{Expr: &ast.MemberExpr{Expr: $2, Name: $4.Lit}}
		$$.SetPosition($2.Position())
	}
	| '*' expr_ident %prec UNARY
	{
		$$ = &ast.DerefExpr{Expr: $2}
		$$.SetPosition($2.Position())
	}
	| '*' expr '.' IDENT %prec UNARY
	{
		$$ = &ast.DerefExpr{Expr: &ast.MemberExpr{Expr: $2, Name: $4.Lit}}
		$$.SetPosition($2.Position())
	}
	| STRING
	{
		$$ = &ast.StringExpr{Lit: $1.Lit}
		$$.SetPosition($1.Position())
	}
	| TRUE
	{
		$$ = &ast.ConstExpr{Value: $1.Lit}
		$$.SetPosition($1.Position())
	}
	| FALSE
	{
		$$ = &ast.ConstExpr{Value: $1.Lit}
		$$.SetPosition($1.Position())
	}
	| NIL
	{
		$$ = &ast.ConstExpr{Value: $1.Lit}
		$$.SetPosition($1.Position())
	}
	| expr '?' expr ':' expr
	{
		$$ = &ast.TernaryOpExpr{Expr: $1, Lhs: $3, Rhs: $5}
		$$.SetPosition($1.Position())
	}
	| expr NILCOALESCE expr
	{
		$$ = &ast.NilCoalescingOpExpr{Lhs: $1, Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr '.' IDENT
	{
		$$ = &ast.MemberExpr{Expr: $1, Name: $3.Lit}
		$$.SetPosition($1.Position())
	}
	| FUNC '(' expr_idents ')' '{' compstmt '}'
	{
		$$ = &ast.FuncExpr{Params: $3, Stmts: $6}
		$$.SetPosition($1.Position())
	}
	| FUNC '(' expr_idents VARARG ')' '{' compstmt '}'
	{
		$$ = &ast.FuncExpr{Params: $3, Stmts: $7, VarArg: true}
		$$.SetPosition($1.Position())
	}
	| FUNC IDENT '(' expr_idents ')' '{' compstmt '}'
	{
		$$ = &ast.FuncExpr{Name: $2.Lit, Params: $4, Stmts: $7}
		$$.SetPosition($1.Position())
	}
	| FUNC IDENT '(' expr_idents VARARG ')' '{' compstmt '}'
	{
		$$ = &ast.FuncExpr{Name: $2.Lit, Params: $4, Stmts: $8, VarArg: true}
		$$.SetPosition($1.Position())
	}
	| '[' opt_newlines exprs opt_comma_newlines ']'
	{
		$$ = &ast.ArrayExpr{Exprs: $3}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| '{' opt_newlines map_expr opt_comma_newlines '}'
	{
		$$ = &ast.MapExpr{MapExpr: $3}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| '(' expr ')'
	{
		$$ = &ast.ParenExpr{SubExpr: $2}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| expr '+' expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "+", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr '-' expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "-", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr '*' expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "*", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr '/' expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "/", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr '%' expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "%", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr POW expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "**", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr SHIFTLEFT expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "<<", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr SHIFTRIGHT expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: ">>", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr EQEQ expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "==", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr NEQ expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "!=", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr '>' expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: ">", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr GE expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: ">=", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr '<' expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "<", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr LE expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "<=", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr PLUSEQ expr
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: "+=", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr MINUSEQ expr
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: "-=", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr MULEQ expr
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: "*=", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr DIVEQ expr
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: "/=", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr ANDEQ expr
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: "&=", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr OREQ expr
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: "|=", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr PLUSPLUS
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: "++"}
		$$.SetPosition($1.Position())
	}
	| expr MINUSMINUS
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: "--"}
		$$.SetPosition($1.Position())
	}
	| expr '|' expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "|", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr OROR expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "||", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr '&' expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "&", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr ANDAND expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "&&", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| IDENT '(' exprs VARARG ')'
	{
		$$ = &ast.CallExpr{Name: $1.Lit, SubExprs: $3, VarArg: true}
		$$.SetPosition($1.Position())
	}
	| IDENT '(' exprs ')'
	{
		$$ = &ast.CallExpr{Name: $1.Lit, SubExprs: $3}
		$$.SetPosition($1.Position())
	}
	| expr '(' exprs VARARG ')'
	{
		$$ = &ast.AnonCallExpr{Expr: $1, SubExprs: $3, VarArg: true}
		$$.SetPosition($1.Position())
	}
	| expr '(' exprs ')'
	{
		$$ = &ast.AnonCallExpr{Expr: $1, SubExprs: $3}
		$$.SetPosition($1.Position())
	}
	| expr_ident '[' expr ']'
	{
		$$ = &ast.ItemExpr{Value: $1, Index: $3}
		$$.SetPosition($1.Position())
	}
	| expr '[' expr ']'
	{
		$$ = &ast.ItemExpr{Value: $1, Index: $3}
		$$.SetPosition($1.Position())
	}
	| expr_slice
	{
		$$ = $1
		$$.SetPosition($1.Position())
	}
	| LEN '(' expr ')'
	{
		$$ = &ast.LenExpr{Expr: $3}
		$$.SetPosition($1.Position())
	}
	| NEW '(' expr_type ')'
	{
		$$ = &ast.NewExpr{Type: $3}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' CHAN expr_type ')'
	{
		$$ = &ast.MakeChanExpr{Type: $4, SizeExpr: nil}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' CHAN expr_type ',' expr ')'
	{
		$$ = &ast.MakeChanExpr{Type: $4, SizeExpr: $6}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' array_count expr_type ')'
	{
		$$ = &ast.MakeExpr{Dimensions: $3.Count, Type: $4}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' array_count expr_type ',' expr ')'
	{
		$$ = &ast.MakeExpr{Dimensions: $3.Count,Type: $4, LenExpr: $6}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' array_count expr_type ',' expr ',' expr ')'
	{
		$$ = &ast.MakeExpr{Dimensions: $3.Count,Type: $4, LenExpr: $6, CapExpr: $8}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' TYPE IDENT ',' expr ')'
	{
		$$ = &ast.MakeTypeExpr{Name: $4.Lit, Type: $6}
		$$.SetPosition($1.Position())
	}
	| expr OPCHAN expr
	{
		$$ = &ast.ChanExpr{Lhs: $1, Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| OPCHAN expr
	{
		$$ = &ast.ChanExpr{Rhs: $2}
		$$.SetPosition($2.Position())
	}
	| CLOSE '(' expr ')'
	{
		$$ = &ast.CloseExpr{WhatExpr: $3}
		$$.SetPosition($1.Position())
	}
	| DELETE '(' expr ')'
	{
		$$ = &ast.DeleteExpr{WhatExpr: $3}
		$$.SetPosition($1.Position())
	}
	| DELETE '(' expr ',' expr ')'
	{
		$$ = &ast.DeleteExpr{WhatExpr: $3, KeyExpr: $5}
		$$.SetPosition($1.Position())
	}
	| expr IN expr
	{
		$$ = &ast.IncludeExpr{ItemExpr: $1, ListExpr: &ast.SliceExpr{Value: $3, Begin: nil, End: nil}}
		$$.SetPosition($1.Position())
	}

expr_ident :
	IDENT
	{
		$$ = &ast.IdentExpr{Lit: $1.Lit}
		$$.SetPosition($1.Position())
	}

opt_term :
	/* nothing */
	| term
	
term :
	';' newlines
	| newlines
	| ';'

opt_newlines : 
	/* nothing */
	| newlines

newlines : 
	newline
	| newlines newline

newline : '\n'

opt_comma_newlines : 
	/* nothing */
	| ',' newlines
	| newlines
	| ','

%%
