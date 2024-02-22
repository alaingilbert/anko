%{
package parser

import (
	"github.com/alaingilbert/anko/pkg/ast"
)

%}

%type<compstmt> compstmt
%type<stmts> stmts

%type<stmt> stmt
%type<stmt> stmt_var_or_lets
%type<stmt> stmt_var
%type<stmt> stmt_lets
%type<stmt> stmt_try
%type<stmt> stmt_defer
%type<stmt> stmt_go
%type<stmt> stmt_if
%type<stmt> stmt_for
%type<stmt> stmt_switch
%type<stmt> stmt_select
%type<stmt> stmt_expr
%type<stmt> stmt_module
%type<stmt> stmt_break
%type<stmt> stmt_return
%type<stmt> stmt_continue
%type<stmt> stmt_throw

%type<opt_stmt_var_or_lets> opt_stmt_var_or_lets
%type<stmt_typed_lets> stmt_typed_lets
%type<opt_finally> opt_finally
%type<maybe_else> maybe_else
%type<else_if_list> else_if_list
%type<else_if> else_if
%type<stmt_switch_cases> stmt_switch_cases
%type<stmt_switch_cases_helper> stmt_switch_cases_helper
%type<stmt_switch_case> stmt_switch_case
%type<stmt_switch_default> stmt_switch_default
%type<stmt_select_body> stmt_select_body
%type<stmt_select_content> stmt_select_content
%type<stmt_select_cases> stmt_select_cases
%type<stmt_select_cases_helper> stmt_select_cases_helper
%type<stmt_select_case> stmt_select_case
%type<stmt_select_default> stmt_select_default
%type<exprs> exprs
%type<opt_exprs> opt_exprs
%type<comma_separated_exprs> comma_separated_exprs

%type<expr> expr
%type<expr> expr_member_or_ident
%type<expr> expr_literals
%type<expr> expr_unary
%type<expr> expr_ternary
%type<expr> expr_nil_coalesce
%type<expr> expr_func
%type<expr> expr_array
%type<expr> expr_paren
%type<expr> expr_binary
%type<expr> expr_call
%type<expr> expr_anon_call
%type<expr> expr_item_or_slice
%type<expr> expr_len
%type<expr> expr_dbg
%type<expr> expr_new
%type<expr> expr_make
%type<expr> expr_map
%type<expr> expr_opchan
%type<expr> expr_close
%type<expr> expr_delete
%type<expr> expr_in

%type<opt_expr> opt_expr
%type<expr_member> expr_member
%type<expr_call_helper> expr_call_helper
%type<expr_literals_helper> expr_literals_helper
%type<unary_op> unary_op
%type<bin_op> bin_op
%type<op_assoc> op_assoc
%type<op_assoc1> op_assoc1
%type<expr_assoc> expr_assoc
%type<expr_idents> expr_idents
%type<expr_for_idents> expr_for_idents
%type<func_expr_idents> func_expr_idents
%type<func_expr_idents_not_empty> func_expr_idents_not_empty
%type<func_expr_untyped_ident> func_expr_untyped_ident
%type<func_expr_typed_ident> func_expr_typed_ident
%type<func_expr_idents_last_untyped> func_expr_idents_last_untyped
%type<func_expr_args> func_expr_args
%type<func_expr_typed_idents> func_expr_typed_idents
%type<opt_func_return_expr_idents> opt_func_return_expr_idents
%type<opt_func_return_expr_idents1> opt_func_return_expr_idents1
%type<opt_func_return_expr_idents2> opt_func_return_expr_idents2
%type<type_data> type_data
%type<type_data_struct> type_data_struct
%type<slice_count> slice_count
%type<typed_slice_count> typed_slice_count
%type<expr_map_content> expr_map_content
%type<expr_map_content_helper> expr_map_content_helper
%type<expr_map_key_value> expr_map_key_value
%type<expr_slice_helper1> expr_slice_helper1
%type<slice> slice
%type<expr_ident> expr_ident
%type<opt_expr_ident> opt_expr_ident
%type<start> start

%union{
	start                           ast.Stmt
	compstmt                        ast.Stmt
	stmts                           *ast.StmtsStmt
	opt_stmt_var_or_lets            ast.Stmt
	stmt_typed_lets                 ast.Stmt
	opt_finally                     ast.Stmt
	maybe_else                      ast.Stmt
	else_if_list                    []ast.Stmt
	else_if                         ast.Stmt
	stmt_switch_cases               *ast.SwitchStmt
	stmt_switch_cases_helper        *ast.SwitchStmt
	stmt_switch_case                ast.Stmt
	stmt_switch_default             ast.Stmt
	stmt_select_body                *ast.SelectBodyStmt
	stmt_select_content             *ast.SelectBodyStmt
	stmt_select_cases               *ast.SelectBodyStmt
	stmt_select_cases_helper        *ast.SelectBodyStmt
	stmt_select_case                ast.Stmt
	stmt_select_default             ast.Stmt
	stmt                            ast.Stmt
	expr                            ast.Expr
	opt_expr                        ast.Expr
	expr_literals_helper            ast.Expr
	unary_op                        string
	expr_assoc                      ast.Expr
	expr_member                     ast.Expr
	expr_call_helper                struct{Exprs []ast.Expr; VarArg bool}
	exprs                           []ast.Expr
	opt_exprs                       []ast.Expr
	comma_separated_exprs           []ast.Expr
	expr_idents                     []string
	expr_for_idents                 []string
	func_expr_idents                []*ast.ParamExpr
	func_expr_idents_not_empty      []*ast.ParamExpr
	func_expr_untyped_ident         *ast.ParamExpr
	func_expr_typed_ident           *ast.ParamExpr
	func_expr_idents_last_untyped   []*ast.ParamExpr
	func_expr_args                  struct{Params []*ast.ParamExpr; VarArg bool; TypeData *ast.TypeStruct}
	func_expr_typed_idents          []*ast.ParamExpr
	opt_func_return_expr_idents     []*ast.FuncReturnValuesExpr
	opt_func_return_expr_idents1    []*ast.FuncReturnValuesExpr
	opt_func_return_expr_idents2    []*ast.FuncReturnValuesExpr
	expr_map_content                *ast.MapExpr
	expr_map_content_helper         *ast.MapExpr
	expr_map_key_value              []ast.Expr
	type_data                       *ast.TypeStruct
        type_data_struct                *ast.TypeStruct
        typed_slice_count               *ast.TypeStruct
        slice_count                     int
	tok                             ast.Token
	bin_op                          string
	op_assoc                        string
	op_assoc1                       string
	expr_slice_helper1              ast.Expr
	slice                           ast.Expr
	expr_ident                      *ast.IdentExpr
	opt_expr_ident                  *ast.IdentExpr
}

%token<tok> IDENT NUMBER STRING ARRAY VARARG FUNC RETURN VAR THROW IF ELSE FOR IN EQEQ NEQ GE LE OROR ANDAND NEW
            TRUE FALSE NIL NILCOALESCE MODULE TRY CATCH FINALLY PLUSEQ MINUSEQ MULEQ DIVEQ ANDEQ OREQ BREAK
            CONTINUE PLUSPLUS MINUSMINUS POW SHIFTLEFT SHIFTRIGHT SWITCH SELECT CASE DEFAULT GO DEFER CHAN MAKE
            OPCHAN TYPE LEN DELETE CLOSE MAP STRUCT DBG WALRUS

/* lowest precedence */
%left POW
%right '=' PLUSEQ MINUSEQ DIVEQ MULEQ ANDEQ OREQ
%right ':'
%right OPCHAN
%right '?' NILCOALESCE
%left OROR
%left ANDAND
%left EQEQ NEQ '<' LE '>' GE
%left '+' '-' '|' '^'
%left '*' '/' '%' SHIFTLEFT SHIFTRIGHT '&'
%right IN
%right PLUSPLUS MINUSMINUS
%right UNARY
/* highest precedence */
/* https://golang.org/ref/spec#Expression */

%start    start

%%
start :
	compstmt
	{
		$$ = $1
		if l, ok := yylex.(*Lexer); ok {
			l.stmt = $$
		}
	}

compstmt :
	opt_term
	{
		$$ = nil
	}
	| opt_term stmts opt_term
	{
		$$ = $2
	}

stmts :
	stmt
	{
		$$ = &ast.StmtsStmt{Stmts: []ast.Stmt{$1}}
	}
	| stmts term stmt
	{
		$1.Stmts = append($1.Stmts, $3)
	}

stmt :
	stmt_var_or_lets
	| stmt_break
	| stmt_continue
	| stmt_return
	| stmt_throw
	| stmt_module
	| stmt_if
	| stmt_for
	| stmt_try
	| stmt_switch
	| stmt_select
	| stmt_go
	| stmt_defer
	| stmt_expr

stmt_break :
	BREAK
	{
		$$ = &ast.BreakStmt{}
		$$.SetPosition($1.Position())
	}

stmt_continue :
	CONTINUE
	{
		$$ = &ast.ContinueStmt{}
		$$.SetPosition($1.Position())
	}

stmt_return :
	RETURN opt_exprs
	{
		$$ = &ast.ReturnStmt{Exprs: $2}
		$$.SetPosition($1.Position())
	}

stmt_throw :
	THROW expr
	{
		$$ = &ast.ThrowStmt{Expr: $2}
		$$.SetPosition($1.Position())
	}

stmt_module :
	MODULE IDENT '{' compstmt '}'
	{
		$$ = &ast.ModuleStmt{Name: $2.Lit, Stmt: $4}
		$$.SetPosition($1.Position())
	}

stmt_expr :
	expr
	{
		$$ = &ast.ExprStmt{Expr: $1}
		$$.SetPosition($1.Position())
	}

stmt_go :
	GO expr_call
	{
		$2.(*ast.CallExpr).Go = true
		$$ = &ast.GoroutineStmt{Expr: $2}
		$$.SetPosition($1.Position())
	}
	| GO expr_anon_call
	{
		$2.(*ast.AnonCallExpr).Go = true
		$$ = &ast.GoroutineStmt{Expr: $2}
		$$.SetPosition($1.Position())
	}

stmt_defer :
	DEFER expr_call
	{
        	$2.(*ast.CallExpr).Defer = true
		$$ = &ast.DeferStmt{Expr: $2}
		$$.SetPosition($2.Position())
	}
	| DEFER expr_anon_call
	{
		$2.(*ast.AnonCallExpr).Defer = true
		$$ = &ast.DeferStmt{Expr: $2}
		$$.SetPosition($2.Position())
	}

stmt_try :
	TRY '{' compstmt '}' CATCH opt_expr_ident '{' compstmt '}' opt_finally
	{
		$$ = &ast.TryStmt{Try: $3, Var: $6.Lit, Catch: $8, Finally: $10}
		$$.SetPosition($1.Position())
	}

opt_finally :
	/* nothing */
	{ $$ = nil }
	| FINALLY '{' compstmt '}'
	{
		$$ = $3
	}

opt_stmt_var_or_lets :
	/* nothing */
	{ $$ = nil }
	| stmt_var_or_lets { $$ = $1 }

stmt_var_or_lets :
	stmt_var          { $$ = $1 }
	| stmt_typed_lets { $$ = $1 }
	| stmt_lets       { $$ = $1 }

stmt_var :
	VAR expr_idents '=' exprs
	{
		if len($2) == 2 && len($4) == 1 {
			if _, ok := $4[0].(*ast.ItemExpr); ok {
				$$ = &ast.VarStmt{Names: $2, Exprs: $4}
			} else {
				$$ = &ast.VarStmt{Names: $2, Exprs: $4}
			}
		} else {
			$$ = &ast.VarStmt{Names: $2, Exprs: $4}
			if len($2) != len($4) && !(len($4) == 1 && len($2) > len($4)) {
				yylex.Error("unexpected ','")
			}
		}
		$$.SetPosition($1.Position())
	}

stmt_typed_lets :
	exprs WALRUS exprs
	{
		if len($1) == 2 && len($3) == 1 {
			if _, ok := $3[0].(*ast.ItemExpr); ok {
				$$ = &ast.LetMapItemStmt{Lhss: $1, Rhs: $3[0]}
			} else {
				$$ = &ast.LetsStmt{Lhss: $1, Operator: "=", Rhss: $3, Typed: true}
			}
		} else {
			$$ = &ast.LetsStmt{Lhss: $1, Operator: "=", Rhss: $3, Typed: true}
			if len($1) != len($3) && !(len($3) == 1 && len($1) > len($3)) {
				yylex.Error("unexpected ','")
			}
		}
		$$.SetPosition($1[0].Position())
	}

stmt_lets :
	exprs '=' exprs
	{
		if len($1) == 2 && len($3) == 1 {
			if _, ok := $3[0].(*ast.ItemExpr); ok {
				$$ = &ast.LetMapItemStmt{Lhss: $1, Rhs: $3[0]}
			} else {
				$$ = &ast.LetsStmt{Lhss: $1, Operator: "=", Rhss: $3}
			}
		} else {
			$$ = &ast.LetsStmt{Lhss: $1, Operator: "=", Rhss: $3}
			if len($1) != len($3) && !(len($3) == 1 && len($1) > len($3)) {
				yylex.Error("unexpected ','")
			}
		}
		$$.SetPosition($1[0].Position())
	}

stmt_if :
	IF expr '{' compstmt '}' else_if_list maybe_else
	{
		$$ = &ast.IfStmt{If: $2, Then: $4, ElseIf: $6, Else: $7}
		$$.SetPosition($1.Position())
	}

else_if_list :
	/* nothing */
	{ $$ = []ast.Stmt{} }
	| else_if_list else_if
	{
		$1 = append($1, $2)
		$$ = $1
	}

else_if :
	ELSE IF expr '{' compstmt '}'
	{
		$$ = &ast.IfStmt{If: $3, Then: $5}
	}

maybe_else :
	/* nothing */
	{ $$ = nil }
	| ELSE '{' compstmt '}'
	{
		$$ = $3
	}

stmt_for :
	FOR '{' compstmt '}'
	{
		$$ = &ast.LoopStmt{Stmt: $3}
		$$.SetPosition($1.Position())
	}
	| FOR expr '{' compstmt '}'
	{
		$$ = &ast.LoopStmt{Expr: $2, Stmt: $4}
		$$.SetPosition($1.Position())
	}
	| FOR expr_for_idents IN expr '{' compstmt '}'
	{
		$$ = &ast.ForStmt{Vars: $2, Value: $4, Stmt: $6}
		$$.SetPosition($1.Position())
	}
	| FOR opt_stmt_var_or_lets ';' opt_expr ';' opt_expr '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Stmt1: $2, Expr2: $4, Expr3: $6, Stmt: $8}
		$$.SetPosition($1.Position())
	}

expr_for_idents :
	IDENT
	{
		$$ = []string{$1.Lit}
	}
	| IDENT ',' IDENT
	{
		$$ = []string{$1.Lit, $3.Lit}
	}

stmt_select :
	SELECT '{' stmt_select_content '}'
	{
		$$ = &ast.SelectStmt{Body: $3}
		$$.SetPosition($1.Position())
	}

stmt_select_content :
	/* nothing */
	{
		$$ = &ast.SelectBodyStmt{}
	}
	| opt_newlines stmt_select_cases opt_newlines
	{
		$$ = $2
	}

stmt_select_cases :
	stmt_select_cases_helper
	{
		$$ = $1
	}

stmt_select_cases_helper :
	stmt_select_body
	{
		$$ = $1
	}
	| stmt_select_cases_helper stmt_select_case
	{
		$$.Cases = append($$.Cases, $2)
	}
	| stmt_select_cases_helper stmt_select_default
	{
		if $$.Default != nil {
		    yylex.Error("multiple default statement")
		}
		$$.Default = $2
	}

stmt_select_body :
	stmt_select_default
	{
		$$ = &ast.SelectBodyStmt{Default: $1}
	}
	| stmt_select_case
	{
		$$ = &ast.SelectBodyStmt{Cases: []ast.Stmt{$1}}
	}

stmt_select_case :
	CASE stmt ':' compstmt
	{
		$$ = &ast.SelectCaseStmt{Expr: $2, Stmt: $4}
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
		$5.Expr = $2
		$$ = $5
		$$.SetPosition($1.Position())
	}

stmt_switch_cases :
	/* nothing */
	{
		$$ = &ast.SwitchStmt{}
	}
	| stmt_switch_cases_helper
	{
		$$ = $1
	}

stmt_switch_cases_helper :
	stmt_switch_default
	{
		$$ = &ast.SwitchStmt{Default: $1}
	}
	| stmt_switch_case
	{
		$$ = &ast.SwitchStmt{Cases: []ast.Stmt{$1}}
	}
	| stmt_switch_cases_helper stmt_switch_case
	{
		$1.Cases = append($1.Cases, $2)
		$$ = $1
	}
	| stmt_switch_cases_helper stmt_switch_default
	{
		if $1.Default != nil {
			yylex.Error("multiple default statement")
		}
		$1.Default = $2
	}


stmt_switch_case :
	CASE expr ':' compstmt
	{
		$$ = &ast.SwitchCaseStmt{Exprs: []ast.Expr{$2}, Stmt: $4}
		$$.SetPosition($1.Position())
	}
	| CASE opt_exprs ':' compstmt
	{
		$$ = &ast.SwitchCaseStmt{Exprs: $2, Stmt: $4}
		$$.SetPosition($1.Position())
	}

stmt_switch_default :
	DEFAULT ':' compstmt
	{
		$$ = $3
	}

opt_func_return_expr_idents :
	{
		$$ = nil
	}
	| type_data
	{
		$$ = []*ast.FuncReturnValuesExpr{&ast.FuncReturnValuesExpr{TypeData: $1}}
	}
	| '(' opt_func_return_expr_idents1 ')'
	{
		$$ = $2
	}
opt_func_return_expr_idents1 :
	{
		$$ = []*ast.FuncReturnValuesExpr{}
	}
	| opt_func_return_expr_idents2
	{
		$$ = $1
	}

opt_func_return_expr_idents2 :
	type_data
	{
		$$ = []*ast.FuncReturnValuesExpr{&ast.FuncReturnValuesExpr{TypeData: $1}}
	}
	| opt_func_return_expr_idents2 comma_opt_newlines type_data
	{
		$$ = append($1, &ast.FuncReturnValuesExpr{TypeData: $3})
	}

func_expr_idents :
	{
		$$ = []*ast.ParamExpr{}
	}
	| func_expr_idents_not_empty { $$ = $1 }

func_expr_idents_not_empty :
	func_expr_idents_last_untyped { $$ = $1 }
	| func_expr_typed_idents      { $$ = $1 }

func_expr_untyped_ident :
	IDENT
	{
		$$ = &ast.ParamExpr{Name: $1.Lit}
	}

func_expr_typed_ident :
	IDENT type_data
	{
		$$ = &ast.ParamExpr{Name: $1.Lit, TypeData: $2}
	}

func_expr_idents_last_untyped :
	func_expr_untyped_ident
	{
		$$ = []*ast.ParamExpr{$1}
	}
	| func_expr_idents_not_empty comma_opt_newlines func_expr_untyped_ident
	{
		$$ = append($1, $3)
	}

func_expr_typed_idents :
	func_expr_typed_ident
	{
		$$ = []*ast.ParamExpr{$1}
	}
	| func_expr_idents_not_empty comma_opt_newlines func_expr_typed_ident
	{
		$$ = append($1, $3)
	}

opt_exprs :
	{
		$$ = nil
	}
	| exprs { $$ = $1 }

exprs :
	expr
	{
		$$ = []ast.Expr{$1}
	}
	| exprs comma_opt_newlines expr
	{
		if len($1) == 0 {
			yylex.Error("syntax error: unexpected ','")
		}
		$$ = append($1, $3)
	}

opt_expr :
	/* nothing */
	{ $$ = nil }
	| expr { $$ = $1 }

expr :
	expr_member_or_ident
	| expr_literals
	| expr_unary
	| expr_ternary
	| expr_nil_coalesce
	| expr_func
	| expr_array
	| expr_paren
	| expr_binary
	| expr_call
	| expr_anon_call
	| expr_item_or_slice
	| expr_len
	| expr_dbg
	| expr_new
	| expr_make
	| expr_map
	| expr_opchan
	| expr_close
	| expr_delete
	| expr_in

expr_dbg :
	DBG '(' ')'
	{
		$$ = &ast.DbgExpr{}
		$$.SetPosition($1.Position())
	}
	| DBG '(' expr ')'
	{
		$$ = &ast.DbgExpr{Expr: $3}
		$$.SetPosition($1.Position())
	}
	| DBG '(' type_data ')'
	{
		$$ = &ast.DbgExpr{TypeData: $3}
		$$.SetPosition($1.Position())
	}

expr_len :
	LEN '(' expr ')'
	{
		$$ = &ast.LenExpr{Expr: $3}
		$$.SetPosition($1.Position())
	}

expr_paren :
	'(' expr ')'
	{
		$$ = &ast.ParenExpr{SubExpr: $2}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}

expr_array :
	'[' ']'
	{
		$$ = &ast.ArrayExpr{}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| '[' comma_separated_exprs ']'
	{
		$$ = &ast.ArrayExpr{Exprs: $2}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| typed_slice_count '{' comma_separated_exprs '}'
	{
		$$ = &ast.ArrayExpr{Exprs: $3, TypeData: $1}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}

comma_separated_exprs :
	opt_newlines opt_exprs opt_comma_opt_newlines
	{
		$$ = $2
	}

expr_nil_coalesce :
	expr NILCOALESCE expr
	{
		$$ = &ast.NilCoalescingOpExpr{Lhs: $1, Rhs: $3}
		$$.SetPosition($1.Position())
	}

expr_ternary :
	expr '?' expr ':' expr
	{
		$$ = &ast.TernaryOpExpr{Expr: $1, Lhs: $3, Rhs: $5}
		$$.SetPosition($1.Position())
	}

expr_new :
	NEW '(' type_data ')'
	{
		if $3.Kind == ast.TypeDefault {
			$3.Kind = ast.TypePtr
			$$ = &ast.MakeExpr{TypeData: $3}
		} else {
			$$ = &ast.MakeExpr{TypeData: &ast.TypeStruct{Kind: ast.TypePtr, SubType: $3}}
		}
		$$.SetPosition($1.Position())
	}

expr_opchan :
	expr OPCHAN expr
	{
		$$ = &ast.ChanExpr{Lhs: $1, Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| OPCHAN expr
	{
		$$ = &ast.ChanExpr{Rhs: $2}
		$$.SetPosition($2.Position())
	}

expr_in :
	expr IN expr
	{
		$$ = &ast.IncludeExpr{ItemExpr: $1, ListExpr: &ast.SliceExpr{Value: $3, Begin: nil, End: nil}}
		$$.SetPosition($1.Position())
	}

expr_delete :
	DELETE '(' expr ')'
	{
		$$ = &ast.DeleteExpr{WhatExpr: $3}
		$$.SetPosition($1.Position())
	}
	| DELETE '(' expr ',' expr ')'
	{
		$$ = &ast.DeleteExpr{WhatExpr: $3, KeyExpr: $5}
		$$.SetPosition($1.Position())
	}

expr_close :
	CLOSE '(' expr ')'
	{
		$$ = &ast.CloseExpr{WhatExpr: $3}
		$$.SetPosition($1.Position())
	}

expr_literals :
	expr_literals_helper
	{
		$$ = $1
		$$.SetPosition($1.Position())
	}

expr_literals_helper :
	NUMBER   { $$ = &ast.NumberExpr{Lit: $1.Lit} }
	| STRING { $$ = &ast.StringExpr{Lit: $1.Lit} }
	| TRUE   { $$ = &ast.ConstExpr{Value: $1.Lit} }
	| FALSE  { $$ = &ast.ConstExpr{Value: $1.Lit} }
	| NIL    { $$ = &ast.ConstExpr{Value: $1.Lit} }

expr_member_or_ident :
	expr_ident    { $$ = $1 }
	| expr_member { $$ = $1 }

expr_ident :
	IDENT
	{
		$$ = &ast.IdentExpr{Lit: $1.Lit}
		$$.SetPosition($1.Position())
	}

expr_member :
	expr '.' IDENT
	{
		$$ = &ast.MemberExpr{Expr: $1, Name: $3.Lit}
		$$.SetPosition($1.Position())
	}

opt_expr_ident :
	/* nothing */
	{ $$ = nil }
	| expr_ident { $$ = $1 }

expr_call :
	expr_ident expr_call_helper
	{
		$$ = &ast.CallExpr{Name: $1.Lit, SubExprs: $2.Exprs, VarArg: $2.VarArg}
		$$.SetPosition($1.Position())
	}

expr_anon_call :
	expr expr_call_helper
	{
		$$ = &ast.AnonCallExpr{Expr: $1, SubExprs: $2.Exprs, VarArg: $2.VarArg}
		$$.SetPosition($1.Position())
	}

expr_call_helper :
	'(' exprs VARARG ')'
	{
		$$ = struct{Exprs []ast.Expr; VarArg bool}{Exprs: $2, VarArg: true}
	}
	| '(' opt_exprs ')'
	{
		$$ = struct{Exprs []ast.Expr; VarArg bool}{Exprs: $2}
	}

unary_op :
	'-'   { $$ = "-" }
	| '!' { $$ = "!" }
	| '^' { $$ = "^" }

expr_unary :
	unary_op expr %prec UNARY
	{
		$$ = &ast.UnaryExpr{Operator: $1, Expr: $2}
		$$.SetPosition($2.Position())
	}
	| '&' expr_member_or_ident %prec UNARY
	{
		if el, ok := $2.(*ast.IdentExpr); ok {
			$$ = &ast.AddrExpr{Expr: el}
		} else if el, ok := $2.(*ast.MemberExpr); ok {
			$$ = el
		}
		$$.SetPosition($2.Position())
	}
	| '*' expr_member_or_ident %prec UNARY
	{
		$$ = &ast.DerefExpr{Expr: $2}
		$$.SetPosition($2.Position())
	}

bin_op :
	'+'          { $$ = "+"; }
	| '-'        { $$ = "-"; }
	| '*'        { $$ = "*"; }
	| '/'        { $$ = "/"; }
	| POW        { $$ = $1.Lit; }
	| '%'        { $$ = "%" }
	| SHIFTLEFT  { $$ = $1.Lit }
	| SHIFTRIGHT { $$ = $1.Lit }
	| '|'        { $$ = "|" }
	| OROR       { $$ = $1.Lit }
	| '&'        { $$ = "&" }
	| ANDAND     { $$ = $1.Lit }

expr_binary :
	expr bin_op expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: $2, Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr EQEQ expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "==", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr op_assoc1 expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: $2, Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr_assoc { $$ = $1 }

op_assoc :
	PLUSEQ    { $$ = "+=" }
	| MINUSEQ { $$ = "-=" }
	| MULEQ   { $$ = "*=" }
	| DIVEQ   { $$ = "/=" }
	| ANDEQ   { $$ = "&=" }
	| OREQ    { $$ = "|=" }

expr_assoc :
	expr op_assoc expr
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: $2, Rhs: $3}
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

op_assoc1 :
	NEQ     { $$ = "!=" }
	| '>'   { $$ = ">" }
	| GE    { $$ = ">=" }
	| '<'   { $$ = "<" }
	| LE    { $$ = "<=" }

expr_func :
	FUNC opt_expr_ident '(' func_expr_args ')' opt_func_return_expr_idents '{' compstmt '}'
	{
		f := &ast.FuncExpr{Params: $4.Params, Returns: $6, Stmt: $8, VarArg: $4.VarArg}
		if $4.TypeData != nil {
			f.Params[len(f.Params)-1].TypeData = $4.TypeData
		}
		if $2 != nil {
			f.Name = $2.Lit
		}
		$$ = f
		$$.SetPosition($1.Position())
	}

func_expr_args :
	func_expr_idents_last_untyped VARARG type_data
	{
		$$ = struct{Params []*ast.ParamExpr; VarArg bool; TypeData *ast.TypeStruct}{Params: $1, VarArg: true, TypeData: $3}
	}
	| func_expr_idents_last_untyped VARARG
	{
		$$ = struct{Params []*ast.ParamExpr; VarArg bool; TypeData *ast.TypeStruct}{Params: $1, VarArg: true, TypeData: nil}
	}
	| func_expr_idents
	{
		$$ = struct{Params []*ast.ParamExpr; VarArg bool; TypeData *ast.TypeStruct}{Params: $1, VarArg: false, TypeData: nil}
	}

expr_make :
	MAKE '(' type_data ')'
	{
		$$ = &ast.MakeExpr{TypeData: $3}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' type_data ',' expr ')'
	{
		$$ = &ast.MakeExpr{TypeData: $3, LenExpr: $5}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' type_data ',' expr ',' expr ')'
	{
		$$ = &ast.MakeExpr{TypeData: $3, LenExpr: $5, CapExpr: $7}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' TYPE IDENT ',' expr ')'
	{
		$$ = &ast.MakeTypeExpr{Name: $4.Lit, Type: $6}
		$$.SetPosition($1.Position())
	}

type_data :
	IDENT
	{
		$$ = &ast.TypeStruct{Name: $1.Lit}
	}
	| type_data '.' IDENT
	{
		if $1.Kind != ast.TypeDefault {
			yylex.Error("not type default")
		} else {
			$1.Env = append($1.Env, $1.Name)
			$1.Name = $3.Lit
		}
	}
	| '*' type_data
	{
		if $2.Kind == ast.TypeDefault {
			$2.Kind = ast.TypePtr
			$$ = $2
		} else {
			$$ = &ast.TypeStruct{Kind: ast.TypePtr, SubType: $2}
		}
	}
	| typed_slice_count
	{
		$$ = $1
	}
	| MAP '[' type_data ']' type_data
	{
		$$ = &ast.TypeStruct{Kind: ast.TypeMap, Key: $3, SubType: $5}
	}
	| CHAN type_data
	{
		if $2.Kind == ast.TypeDefault {
			$2.Kind = ast.TypeChan
			$$ = $2
		} else {
			$$ = &ast.TypeStruct{Kind: ast.TypeChan, SubType: $2}
		}
	}
	| STRUCT '{' opt_newlines type_data_struct opt_newlines '}'
	{
		$$ = $4
	}

type_data_struct :
	IDENT type_data
	{
		$$ = &ast.TypeStruct{Kind: ast.TypeStructType, StructNames: []string{$1.Lit}, StructTypes: []*ast.TypeStruct{$2}}
	}
	| type_data_struct comma_opt_newlines IDENT type_data
	{
		if $1 == nil {
			yylex.Error("syntax error: unexpected ','")
		}
		$$.StructNames = append($$.StructNames, $3.Lit)
		$$.StructTypes = append($$.StructTypes, $4)
	}

typed_slice_count :
	slice_count type_data
	{
		if $2.Kind == ast.TypeDefault {
			$2.Kind = ast.TypeSlice
			$2.Dimensions = $1
			$$ = $2
		} else {
			$$ = &ast.TypeStruct{Kind: ast.TypeSlice, SubType: $2, Dimensions: $1}
		}
	}

slice_count :
	'[' ']'
	{
		$$ = 1
	}
	| '[' ']' slice_count
	{
		$$ = $3 + 1
	}

expr_map :
	MAP '{' expr_map_content '}'
	{
		$3.TypeData = &ast.TypeStruct{Kind: ast.TypeMap, Key: &ast.TypeStruct{Name: "interface"}, SubType: &ast.TypeStruct{Name: "interface"}}
		$$ = $3
		$$.SetPosition($1.Position())
	}
	| MAP '[' type_data ']' type_data '{' expr_map_content '}'
	{
		$7.TypeData = &ast.TypeStruct{Kind: ast.TypeMap, Key: $3, SubType: $5}
		$$ = $7
		$$.SetPosition($1.Position())
	}
	| '{' expr_map_content '}'
	{
		$$ = $2
		$$.SetPosition($2.Position())
	}

expr_map_content :
	opt_newlines
	{
		$$ = &ast.MapExpr{}
	}
	| opt_newlines expr_map_content_helper opt_comma_opt_newlines
	{
		$$ = $2
	}

expr_map_content_helper :
	expr_map_key_value
	{
		$$ = &ast.MapExpr{Keys: []ast.Expr{$1[0]}, Values: []ast.Expr{$1[1]}}
	}
	| expr_map_content_helper comma_opt_newlines expr_map_key_value
	{
		$$.Keys = append($$.Keys, $3[0])
		$$.Values = append($$.Values, $3[1])
	}

expr_map_key_value :
	expr ':' expr
	{
		$$ = []ast.Expr{$1, $3}
	}

expr_item_or_slice :
	expr expr_slice_helper1
	{
		if el, ok := $2.(*ast.SliceExpr); ok {
			el.Value = $1
		} else if el, ok := $2.(*ast.ItemExpr); ok {
			el.Value = $1
		}
		$$ = $2
               	$$.SetPosition($1.Position())
	}

expr_slice_helper1 :
	'[' slice ']'
	{
		$$ = $2
	}

slice :
	expr ':' expr
	{
		$$ = &ast.SliceExpr{Begin: $1, End: $3}
	}
	| expr ':'
	{
		$$ = &ast.SliceExpr{Begin: $1, End: nil}
	}
	| ':' expr
	{
		$$ = &ast.SliceExpr{Begin: nil, End: $2}
	}
	| expr
	{
		$$ = &ast.ItemExpr{Index: $1}
	}

expr_idents :
	expr_ident
	{
		$$ = []string{$1.Lit}
	}
	| expr_idents comma_opt_newlines expr_ident
	{
		$$ = append($1, $3.Lit)
	}

opt_term :
	/* nothing */
	| term

term :
	';'
	| newlines
	| ';' newlines

opt_newlines :
	/* nothing */
	| newlines

newlines :
	newline
	| newlines newline

newline : '\n'

comma_opt_newlines :
	',' opt_newlines

opt_comma_opt_newlines :
	comma_opt_newlines
	| opt_newlines

%%
