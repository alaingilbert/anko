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
%type<stmt_typed_lets> stmt_typed_lets
%type<stmt_try> stmt_try
%type<stmt_defer> stmt_defer
%type<stmt_go> stmt_go
%type<stmt_if> stmt_if
%type<stmt_for> stmt_for
%type<stmt_switch> stmt_switch
%type<stmt_switch_cases> stmt_switch_cases
%type<stmt_switch_cases_helper> stmt_switch_cases_helper
%type<stmt_switch_case> stmt_switch_case
%type<stmt_switch_default> stmt_switch_default
%type<stmt_select> stmt_select
%type<stmt_select_body> stmt_select_body
%type<stmt_select_cases> stmt_select_cases
%type<stmt_select_cases_helper> stmt_select_cases_helper
%type<stmt_select_case> stmt_select_case
%type<stmt_select_default> stmt_select_default
%type<exprs> exprs
%type<opt_exprs> opt_exprs
%type<expr> expr
%type<expr_member_or_ident> expr_member_or_ident
%type<expr_call> expr_call
%type<expr_anon_call> expr_anon_call
%type<expr_func> expr_func
%type<expr_make> expr_make
%type<expr_dbg> expr_dbg
%type<expr_literals> expr_literals
%type<expr_unary> expr_unary
%type<expr_binary> expr_binary
%type<expr_idents> expr_idents
%type<func_expr_idents> func_expr_idents
%type<func_expr_idents_not_empty> func_expr_idents_not_empty
%type<func_expr_untyped_ident> func_expr_untyped_ident
%type<func_expr_typed_ident> func_expr_typed_ident
%type<func_expr_idents_last_untyped> func_expr_idents_last_untyped
%type<func_expr_typed_idents> func_expr_typed_idents
%type<opt_func_return_expr_idents> opt_func_return_expr_idents
%type<opt_func_return_expr_idents1> opt_func_return_expr_idents1
%type<opt_func_return_expr_idents2> opt_func_return_expr_idents2
%type<type_data> type_data
%type<type_data_struct> type_data_struct
%type<slice_count> slice_count
%type<expr_map> expr_map
%type<expr_map_content> expr_map_content
%type<expr_map_content_helper> expr_map_content_helper
%type<expr_map_key_value> expr_map_key_value
%type<expr_slice> expr_slice
%type<expr_ident> expr_ident

%union{
	compstmt                        ast.Stmt
	stmts                           ast.Stmt
	stmt_var_or_lets                ast.Stmt
	stmt_var                        ast.Stmt
	stmt_lets                       ast.Stmt
	stmt_typed_lets                 ast.Stmt
	stmt_try                        ast.Stmt
	stmt_defer                      ast.Stmt
	stmt_go                         ast.Stmt
	stmt_if                         ast.Stmt
	stmt_for                        ast.Stmt
	stmt_switch                     ast.Stmt
	stmt_switch_cases               *ast.SwitchStmt
	stmt_switch_cases_helper        *ast.SwitchStmt
	stmt_switch_case                ast.Stmt
	stmt_switch_default             ast.Stmt
	stmt_select                     ast.Stmt
	stmt_select_body                *ast.SelectBodyStmt
	stmt_select_cases               *ast.SelectBodyStmt
	stmt_select_cases_helper        *ast.SelectBodyStmt
	stmt_select_case                ast.Stmt
	stmt_select_default             ast.Stmt
	stmt                            ast.Stmt
	expr                            ast.Expr
	expr_dbg                        ast.Expr
	expr_literals                   ast.Expr
	expr_unary                      ast.Expr
	expr_binary                     ast.Expr
	expr_member_or_ident            ast.Expr
	expr_call                       *ast.CallExpr
	expr_anon_call                  *ast.AnonCallExpr
	expr_func                       ast.Expr
	expr_make                       ast.Expr
	exprs                           []ast.Expr
	opt_exprs                       []ast.Expr
	expr_idents                     []string
	func_expr_idents                []*ast.ParamExpr
	func_expr_idents_not_empty      []*ast.ParamExpr
	func_expr_untyped_ident         *ast.ParamExpr
	func_expr_typed_ident           *ast.ParamExpr
	func_expr_idents_last_untyped   []*ast.ParamExpr
	func_expr_typed_idents          []*ast.ParamExpr
	opt_func_return_expr_idents     []*ast.FuncReturnValuesExpr
	opt_func_return_expr_idents1    []*ast.FuncReturnValuesExpr
	opt_func_return_expr_idents2    []*ast.FuncReturnValuesExpr
	expr_map                        *ast.MapExpr
	expr_map_content                *ast.MapExpr
	expr_map_content_helper         *ast.MapExpr
	expr_map_key_value              []ast.Expr
	type_data                       *ast.TypeStruct
        type_data_struct                *ast.TypeStruct
        slice_count                     int
	tok                             ast.Token
	expr_slice                      ast.Expr
	expr_ident                      ast.Expr
}

%token<tok> IDENT NUMBER STRING ARRAY VARARG FUNC RETURN VAR THROW IF ELSE FOR IN EQEQ NEQ GE LE OROR ANDAND NEW
            TRUE FALSE NIL NILCOALESCE MODULE TRY CATCH FINALLY PLUSEQ MINUSEQ MULEQ DIVEQ ANDEQ OREQ BREAK
            CONTINUE PLUSPLUS MINUSMINUS POW SHIFTLEFT SHIFTRIGHT SWITCH SELECT CASE DEFAULT GO DEFER CHAN MAKE
            OPCHAN TYPE LEN DELETE CLOSE MAP STRUCT DBG WALRUS

%right '=' PLUSEQ MINUSEQ MULEQ DIVEQ ANDEQ OREQ
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

%%
start : compstmt

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
		if l, ok := yylex.(*Lexer); ok {
			l.stmt = $$
		}
	}
	| stmts term stmt
	{
		stmts := $1.(*ast.StmtsStmt)
		stmts.Stmts = append(stmts.Stmts, $3)
		if l, ok := yylex.(*Lexer); ok {
			l.stmt = $$
		}
	}

stmt :
	stmt_var_or_lets { $$ = $1 }
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
	| RETURN opt_exprs
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
		$$ = &ast.ModuleStmt{Name: $2.Lit, Stmt: $4}
		$$.SetPosition($1.Position())
	}
	| stmt_if
	{
		$$ = $1
		$$.SetPosition($1.Position())
	}
	| stmt_for    { $$ = $1 }
	| stmt_try    { $$ = $1 }
	| stmt_switch { $$ = $1 }
	| stmt_select { $$ = $1 }
	| stmt_go     { $$ = $1 }
	| stmt_defer  { $$ = $1 }
	| expr
	{
		$$ = &ast.ExprStmt{Expr: $1}
		$$.SetPosition($1.Position())
	}

stmt_go :
	GO expr_call
	{
		callExpr := $2
		callExpr.Go = true
		$$ = &ast.GoroutineStmt{Expr: callExpr}
		$$.SetPosition($1.Position())
	}
	| GO expr_anon_call
	{
		anonCallExpr := $2
		anonCallExpr.Go = true
		$$ = &ast.GoroutineStmt{Expr: anonCallExpr}
		$$.SetPosition($1.Position())
	}

stmt_defer :
	DEFER expr_call
	{
		callExpr := $2
        	callExpr.Defer = true
		$$ = &ast.DeferStmt{Expr: callExpr}
		$$.SetPosition($2.Position())
	}
	| DEFER expr_anon_call
	{
		anonCallExpr := $2
		anonCallExpr.Defer = true
		$$ = &ast.DeferStmt{Expr: anonCallExpr}
		$$.SetPosition($2.Position())
	}

stmt_try :
	TRY '{' compstmt '}' CATCH IDENT '{' compstmt '}' FINALLY '{' compstmt '}'
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

stmt_var_or_lets :
	stmt_var          { $$ = $1 }
	| stmt_typed_lets { $$ = $1 }
	| stmt_lets       { $$ = $1 }

stmt_var :
	VAR expr_idents '=' exprs
	{
		$$ = &ast.VarStmt{Names: $2, Exprs: $4}
		$$.SetPosition($1.Position())
	}

stmt_typed_lets :
	expr WALRUS expr
	{
		$$ = &ast.LetsStmt{Lhss: []ast.Expr{$1}, Operator: "=", Rhss: []ast.Expr{$3}, Typed: true}
	}
	| exprs WALRUS exprs
	{
		if len($1) == 2 && len($3) == 1 {
			if _, ok := $3[0].(*ast.ItemExpr); ok {
				$$ = &ast.LetMapItemStmt{Lhss: $1, Rhs: $3[0]}
			} else {
				$$ = &ast.LetsStmt{Lhss: $1, Operator: "=", Rhss: $3, Typed: true}
			}
		} else {
			$$ = &ast.LetsStmt{Lhss: $1, Operator: "=", Rhss: $3, Typed: true}
		}
		if len($1) > 0 {
			$$.SetPosition($1[0].Position())
		}
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
		$$ = &ast.LoopStmt{Stmt: $3}
		$$.SetPosition($1.Position())
	}
	| FOR expr_idents IN expr '{' compstmt '}'
	{
		if len($2) < 1 {
			yylex.Error("missing identifier")
		} else if len($2) > 2 {
			yylex.Error("too many identifiers")
		} else {
			$$ = &ast.ForStmt{Vars: $2, Value: $4, Stmt: $6}
			$$.SetPosition($1.Position())
		}
	}
	| FOR expr '{' compstmt '}'
	{
		$$ = &ast.LoopStmt{Expr: $2, Stmt: $4}
		$$.SetPosition($1.Position())
	}
	| FOR ';' ';' '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Stmt: $5}
		$$.SetPosition($1.Position())
	}
	| FOR ';' ';' expr '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Expr3: $4, Stmt: $6}
		$$.SetPosition($1.Position())
	}
	| FOR ';' expr ';' '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Expr2: $3, Stmt: $6}
		$$.SetPosition($1.Position())
	}
	| FOR ';' expr ';' expr '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Expr2: $3, Expr3: $5, Stmt: $7}
		$$.SetPosition($1.Position())
	}
	| FOR stmt_var_or_lets ';' ';' '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Stmt1: $2, Stmt: $6}
		$$.SetPosition($1.Position())
	}
	| FOR stmt_var_or_lets ';' ';' expr '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Stmt1: $2, Expr3: $5, Stmt: $7}
		$$.SetPosition($1.Position())
	}
	| FOR stmt_var_or_lets ';' expr ';' '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Stmt1: $2, Expr2: $4, Stmt: $7}
		$$.SetPosition($1.Position())
	}
	| FOR stmt_var_or_lets ';' expr ';' expr '{' compstmt '}'
	{
		$$ = &ast.CForStmt{Stmt1: $2, Expr2: $4, Expr3: $6, Stmt: $8}
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
	| stmt_select_cases_helper
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

expr_idents :
	IDENT
	{
		$$ = []string{$1.Lit}
	}
	| expr_idents comma_newlines IDENT
	{
		$$ = append($1, $3.Lit)
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
	| opt_func_return_expr_idents2 comma_newlines type_data
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
	| func_expr_idents_not_empty comma_newlines func_expr_untyped_ident
	{
		$$ = append($1, $3)
	}

func_expr_typed_idents :
	func_expr_typed_ident
	{
		$$ = []*ast.ParamExpr{$1}
	}
	| func_expr_idents_not_empty comma_newlines func_expr_typed_ident
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
	| exprs comma_newlines expr
	{
		if len($1) == 0 {
			yylex.Error("syntax error: unexpected ','")
		}
		$$ = append($1, $3)
	}
	| exprs comma_newlines expr_ident
	{
		if len($1) == 0 {
			yylex.Error("syntax error: unexpected ','")
		}
		$$ = append($1, $3)
	}

expr :
	expr_member_or_ident { $$ = $1 }
	| expr_literals { $$ = $1 }
	| expr_unary { $$ = $1 }
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
	| expr_func { $$ = $1 }
	| '[' ']'
	{
		$$ = &ast.ArrayExpr{}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| '[' opt_newlines opt_exprs opt_comma_newlines ']'
	{
		$$ = &ast.ArrayExpr{Exprs: $3}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| slice_count type_data '{' opt_newlines opt_exprs opt_comma_newlines '}'
	{
		$$ = &ast.ArrayExpr{Exprs: $5, TypeData: &ast.TypeStruct{Kind: ast.TypeSlice, SubType: $2, Dimensions: $1}}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| '(' expr ')'
	{
		$$ = &ast.ParenExpr{SubExpr: $2}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| expr_binary { $$ = $1 }
	| expr_call { $$ = $1 }
	| expr_anon_call { $$ = $1 }
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
	| expr_dbg { $$ = $1 }
	| NEW '(' type_data ')'
	{
		if $3.Kind == ast.TypeDefault {
			$3.Kind = ast.TypePtr
			$$ = &ast.MakeExpr{TypeData: $3}
		} else {
			$$ = &ast.MakeExpr{TypeData: &ast.TypeStruct{Kind: ast.TypePtr, SubType: $3}}
		}
		$$.SetPosition($1.Position())
	}
	| expr_make { $$ = $1 }
	| expr_map { $$ = $1 }
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

expr_literals :
	NUMBER
	{
		$$ = &ast.NumberExpr{Lit: $1.Lit}
		$$.SetPosition($1.Position())
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

expr_member_or_ident :
	expr_ident { $$ = $1 }
	| expr '.' IDENT
	{
		$$ = &ast.MemberExpr{Expr: $1, Name: $3.Lit}
		$$.SetPosition($1.Position())
	}

expr_call :
	IDENT '(' exprs VARARG ')'
	{
		$$ = &ast.CallExpr{Name: $1.Lit, SubExprs: $3, VarArg: true}
		$$.SetPosition($1.Position())
	}
	| IDENT '(' opt_exprs ')'
	{
		$$ = &ast.CallExpr{Name: $1.Lit, SubExprs: $3}
		$$.SetPosition($1.Position())
	}

expr_anon_call :
	expr '(' exprs VARARG ')'
	{
		$$ = &ast.AnonCallExpr{Expr: $1, SubExprs: $3, VarArg: true}
		$$.SetPosition($1.Position())
	}
	| expr '(' opt_exprs ')'
	{
		$$ = &ast.AnonCallExpr{Expr: $1, SubExprs: $3}
		$$.SetPosition($1.Position())
	}

expr_unary :
	'-' expr %prec UNARY
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

expr_binary :
	expr '+' expr
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

expr_func :
	FUNC '(' func_expr_idents ')' opt_func_return_expr_idents '{' compstmt '}'
	{
		$$ = &ast.FuncExpr{Params: $3, Returns: $5, Stmt: $7}
		$$.SetPosition($1.Position())
	}
	| FUNC '(' func_expr_idents_last_untyped VARARG ')' opt_func_return_expr_idents '{' compstmt '}'
	{
		$$ = &ast.FuncExpr{Params: $3, Returns: $6, Stmt: $8, VarArg: true}
		$$.SetPosition($1.Position())
	}
	| FUNC '(' func_expr_idents_last_untyped VARARG type_data ')' opt_func_return_expr_idents '{' compstmt '}'
	{
		$3[len($3)-1].TypeData = $5
		$$ = &ast.FuncExpr{Params: $3, Returns: $7, Stmt: $9, VarArg: true}
		$$.SetPosition($1.Position())
	}
	| FUNC IDENT '(' func_expr_idents ')' opt_func_return_expr_idents '{' compstmt '}'
	{
		$$ = &ast.FuncExpr{Name: $2.Lit, Params: $4, Returns: $6, Stmt: $8}
		$$.SetPosition($1.Position())
	}
	| FUNC IDENT '(' func_expr_idents_last_untyped VARARG ')' opt_func_return_expr_idents '{' compstmt '}'
	{
		$$ = &ast.FuncExpr{Name: $2.Lit, Params: $4, Returns: $7, Stmt: $9, VarArg: true}
		$$.SetPosition($1.Position())
	}
	| FUNC IDENT '(' func_expr_idents_last_untyped VARARG type_data ')' opt_func_return_expr_idents '{' compstmt '}'
	{
		$4[len($4)-1].TypeData = $6
		$$ = &ast.FuncExpr{Name: $2.Lit, Params: $4, Returns: $8, Stmt: $10, VarArg: true}
		$$.SetPosition($1.Position())
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
	| slice_count type_data
	{
		if $2.Kind == ast.TypeDefault {
			$2.Kind = ast.TypeSlice
			$2.Dimensions = $1
			$$ = $2
		} else {
			$$ = &ast.TypeStruct{Kind: ast.TypeSlice, SubType: $2, Dimensions: $1}
		}
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
	| type_data_struct comma_newlines IDENT type_data
	{
		if $1 == nil {
			yylex.Error("syntax error: unexpected ','")
		}
		$$.StructNames = append($$.StructNames, $3.Lit)
		$$.StructTypes = append($$.StructTypes, $4)
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
	| opt_newlines expr_map_content_helper opt_comma_newlines
	{
		$$ = $2
	}

expr_map_content_helper :
	expr_map_key_value
	{
		$$ = &ast.MapExpr{Keys: []ast.Expr{$1[0]}, Values: []ast.Expr{$1[1]}}
	}
	| expr_map_content_helper comma_newlines expr_map_key_value
	{
		if $1.Keys == nil {
			yylex.Error("syntax error: unexpected ','")
		}
		$$.Keys = append($$.Keys, $3[0])
		$$.Values = append($$.Values, $3[1])
	}

expr_map_key_value :
	expr ':' expr
	{
		$$ = []ast.Expr{$1, $3}
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

comma_newlines :
	',' opt_newlines

opt_comma_newlines :
	comma_newlines
	| opt_newlines

%%
