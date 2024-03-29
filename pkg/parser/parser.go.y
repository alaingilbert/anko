// Remaining reduce/reduce conflicts are:
//     label conflicts with expr_member_or_ident
//
// https://gitlab.com/cznic/gc/-/blob/20cf7bee948948b44142f2a04eff623012069407/v3/internal/ebnf/spec.ebnf

%{
package parser

import (
	"github.com/alaingilbert/anko/pkg/ast"
)

%}

%type<stmtsStmt> stmtsStmt

%type<stmts> stmt_select_cases
%type<stmts> opt_stmt_select_cases
%type<stmts> opt_stmt_switch_cases
%type<stmts> stmt_switch_cases

%type<stmt> stmt
%type<stmt> stmt_var_or_lets
%type<stmt> stmt_var
%type<stmt> stmt_lets
%type<stmt> stmt_try
%type<stmt> stmt_defer
%type<stmt> stmt_go
%type<stmt> stmt_if
%type<stmt> stmt_loop
%type<stmt> stmt_for
%type<stmt> for_content
%type<stmt> stmt_switch
%type<stmt> stmt_select
%type<stmt> stmt_expr
%type<stmt> stmt_module
%type<stmt> stmt_break
%type<stmt> labeled_stmt
%type<stmt> stmt_return
%type<stmt> stmt_continue
%type<stmt> stmt_throw
%type<stmt> start
%type<stmt> compstmt
%type<stmt> block
%type<stmt> opt_stmt_var_or_lets
%type<stmt> opt_finally
%type<stmt> maybe_else
%type<stmt> switch_content
%type<stmt> stmt_switch_case
%type<stmt> stmt_switch_opt_default
%type<stmt> stmt_switch_default
%type<stmt_select_content> stmt_select_content
%type<stmt> stmt_select_case
%type<stmt> stmt_select_default
%type<stmt> stmt_select_opt_default
%type<stmt> stmt_dbg
%type<stmt> dbg_content

%type<exprsExpr> exprs
%type<exprsExpr> opt_exprs
%type<exprsExpr> element_list
%type<exprsExpr> opt_element_list
%type<exprs> expr_map_key_value

%type<expr> expr
%type<expr> array_length
%type<expr> element
%type<expr> key
%type<expr> keyed_element
%type<expr> literal_value
%type<expr> composite_lit
%type<expr> expr_member_or_ident
%type<expr> expr_literals
%type<expr> expr_unary
%type<expr> expr_ternary
%type<expr> expr_func
%type<expr> expr_array
%type<expr> expr_paren
%type<expr> expr_binary
%type<expr> expr_call
%type<expr> expr_callable
%type<expr> expr_anon_call
%type<expr> expr_item_or_slice
%type<expr> expr_len
%type<expr> expr_new
%type<expr> expr_make
%type<expr> expr_map
%type<expr> expr_opchan
%type<expr> expr_close
%type<expr> expr_delete
%type<expr> expr_iterable
%type<expr> expr_member
%type<expr> expr_ident
%type<expr> expr_literals_helper
%type<expr> expr_assoc
%type<expr> opt_expr
%type<expr> slice

%type<str> unary_op
%type<str> bin_op
%type<str> op_assoc1

%type<type_data> type
%type<type_data> key_type
%type<type_data> element_type
%type<type_data> array_type
%type<type_data> slice_type
%type<type_data> literal_type
%type<type_data> type_lit
%type<type_data> qualified_ident
%type<type_data> type_name
%type<type_data> channel_type
%type<type_data> struct_type
%type<type_data> pointer_type
%type<type_data> map_type
%type<type_data> type_data_struct
%type<type_data> type_struct_content
%type<type_data> typed_slice_count

%type<expr_idents> expr_idents

%type<expr_call_helper> expr_call_helper

%type<func_expr_idents> func_expr_idents
%type<func_expr_idents> func_expr_idents_not_empty
%type<func_expr_idents> func_expr_typed_idents
%type<func_expr_idents> func_expr_idents_last_untyped

%type<func_expr_typed_ident> func_expr_typed_ident
%type<func_expr_typed_ident> func_expr_untyped_ident

%type<func_expr_args> func_expr_args

%type<opt_func_return_expr_idents> opt_func_return_expr_idents
%type<opt_func_return_expr_idents> opt_func_return_expr_idents1
%type<opt_func_return_expr_idents> opt_func_return_expr_idents2

%type<expr_map> expr_map_container
%type<expr_map> expr_map_content
%type<expr_map> expr_map_content_helper

%type<slice_count> slice_count
%type<expr_typed_ident> expr_typed_ident
%type<opt_ident> opt_ident
%type<op_lets> op_lets
%type<tok> label
%type<tok> package_name
%type<tok> const_expr
%type<stmt_lets_helper> stmt_lets_helper

%union{
	stmtsStmt                       *ast.StmtsStmt
	exprsExpr                      	*ast.ExprsExpr
	stmt                            ast.Stmt
	expr                            ast.Expr
	exprs                           []ast.Expr
	stmts                           []ast.Stmt
	stmt_select_content             *ast.SelectBodyStmt
	expr_call_helper                struct{Exprs *ast.ExprsExpr; VarArg bool}
	expr_idents                     []string
	func_expr_idents                []*ast.ParamExpr
	func_expr_typed_ident           *ast.ParamExpr
	func_expr_args                  struct{Params []*ast.ParamExpr; VarArg bool; TypeData *ast.TypeStruct}
	expr_typed_ident                struct{Name string; TypeData *ast.TypeStruct}
	stmt_lets_helper                struct{Exprs1, Exprs2 *ast.ExprsExpr; Typed, Mutable bool}
	opt_func_return_expr_idents     []*ast.FuncReturnValuesExpr
	expr_map                        *ast.MapExpr
	type_data                       *ast.TypeStruct
        slice_count                     int
	tok                             ast.Token
	opt_ident                       *ast.Token
	str                             string
	op_lets                         bool
}

%token<tok> IDENT NUMBER STRING ARRAY VARARG FUNC RETURN VAR THROW IF ELSE FOR LOOP IN EQEQ NEQ GE LE OROR ANDAND NEW
            TRUE FALSE NIL NILCOALESCE MODULE TRY CATCH FINALLY PLUSEQ MINUSEQ MULEQ DIVEQ ANDEQ OREQ BREAK
            CONTINUE PLUSPLUS MINUSMINUS POW SHIFTLEFT SHIFTRIGHT SWITCH SELECT CASE DEFAULT GO DEFER CHAN MAKE
            OPCHAN TYPE LEN DELETE CLOSE MAP STRUCT DBG WALRUS EMPTYARR MUT

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
	  opt_term                    { $$ = nil }
	| opt_term stmtsStmt opt_term { $$ = $2  }

stmtsStmt :
	  stmt                { $$ = &ast.StmtsStmt{Stmts: []ast.Stmt{$1}} }
	| stmtsStmt term stmt { $1.Stmts = append($1.Stmts, $3)            }

stmt :
	stmt_var_or_lets
	| labeled_stmt
	| stmt_break
	| stmt_continue
	| stmt_return
	| stmt_throw
	| stmt_module
	| stmt_if
	| stmt_loop
	| stmt_for
	| stmt_try
	| stmt_switch
	| stmt_select
	| stmt_go
	| stmt_defer
	| stmt_expr
	| stmt_dbg

expr :
	expr_iterable
	| composite_lit
	| expr_literals
	| expr_unary
	| expr_func
	| expr_binary
	| expr_len
	| expr_new
	| expr_make
	| expr_opchan
	| expr_close
	| expr_delete

expr_iterable :
	expr_map
	| expr_paren
	| expr_array
	| expr_anon_call
	| expr_call
	| expr_member_or_ident
	| expr_item_or_slice
	| expr_ternary

block : '{' compstmt '}' { $$ = $2 }

label : IDENT

labeled_stmt :
	label ':' term stmt
	{
		$4.SetLabel($1.Lit)
		$$ = &ast.LabelStmt{Name: $1.Lit, Stmt: $4}
	}

stmt_break :
	BREAK
	{
		$$ = &ast.BreakStmt{}
		$$.SetPosition($1.Position())
	}
	| BREAK label {
		$$ = &ast.BreakStmt{Label: $2.Lit}
		$$.SetPosition($1.Position())
	}

stmt_continue :
	CONTINUE
	{
		$$ = &ast.ContinueStmt{}
		$$.SetPosition($1.Position())
	}
	| CONTINUE label {
		$$ = &ast.ContinueStmt{Label: $2.Lit}
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
	MODULE IDENT block
	{
		$$ = &ast.ModuleStmt{Name: $2.Lit, Stmt: $3}
		$$.SetPosition($1.Position())
	}

stmt_expr :
	expr
	{
		$$ = &ast.ExprStmt{Expr: $1}
		$$.SetPosition($1.Position())
	}

stmt_go :
	GO expr_callable
	{
		if el, ok := $2.(*ast.CallExpr); ok {
			el.Go = true
		} else if el, ok := $2.(*ast.AnonCallExpr); ok {
			el.Go = true
		}
		$$ = &ast.GoroutineStmt{Expr: $2}
		$$.SetPosition($1.Position())
	}

stmt_defer :
	DEFER expr_callable
	{
		if el, ok := $2.(*ast.CallExpr); ok {
			el.Defer = true
		} else if el, ok := $2.(*ast.AnonCallExpr); ok {
			el.Defer = true
		}
		$$ = &ast.DeferStmt{Expr: $2}
		$$.SetPosition($2.Position())
	}

stmt_try :
	TRY block CATCH opt_ident block opt_finally
	{
		$$ = &ast.TryStmt{Try: $2, Var: $4.Lit, Catch: $5, Finally: $6}
		$$.SetPosition($1.Position())
	}

opt_finally :
	/* nothing */   { $$ = nil }
	| FINALLY block { $$ = $2  }

opt_stmt_var_or_lets :
	/* nothing */      { $$ = nil }
	| stmt_var_or_lets { $$ = $1  }

stmt_var_or_lets :
	stmt_var
	| stmt_lets

stmt_var :
	VAR expr_idents '=' exprs
	{
		isItem := false
		if len($2) == 2 && len($4.Exprs) == 1 {
			if _, ok := $4.Exprs[0].(*ast.ItemExpr); ok {
				isItem = true
				arr := &ast.ExprsExpr{}
				for _, el := range $2 {
					arr.Exprs = append(arr.Exprs, &ast.IdentExpr{Lit: el})
				}
				$$ = &ast.LetMapItemStmt{Lhss: arr, Rhs: $4.Exprs[0]}
			}
		}
		if !isItem {
			$$ = &ast.VarStmt{Names: $2, Exprs: $4.Exprs}
			if len($2) != len($4.Exprs) && !(len($4.Exprs) == 1 && len($2) > len($4.Exprs)) {
				yylex.Error("unexpected ','")
			}
		}
		$$.SetPosition($1.Position())
	}

stmt_lets :
	stmt_lets_helper
	{
		lhs := $1.Exprs1
		rhs := $1.Exprs2
		isItem := false
		if len(lhs.Exprs) == 2 && len(rhs.Exprs) == 1 {
			if _, ok := rhs.Exprs[0].(*ast.ItemExpr); ok {
				isItem = true
				$$ = &ast.LetMapItemStmt{Lhss: lhs, Rhs: rhs.Exprs[0]}
			}
		}
		if !isItem {
			$$ = &ast.LetsStmt{Lhss: lhs, Operator: "=", Rhss: rhs, Typed: $1.Typed, Mutable: $1.Mutable}
			if len(lhs.Exprs) != len(rhs.Exprs) && !(len(rhs.Exprs) == 1 && len(lhs.Exprs) > len(rhs.Exprs)) {
				yylex.Error("unexpected ','")
			}
		}
		$$.SetPosition(lhs.Exprs[0].Position())
	}

stmt_lets_helper :
	exprs op_lets exprs
	{
		$$ = struct{Exprs1, Exprs2 *ast.ExprsExpr; Typed, Mutable bool}{Exprs1: $1, Exprs2: $3, Typed: $2, Mutable: false}
	}
	| MUT exprs WALRUS exprs
	{
		$$ = struct{Exprs1, Exprs2 *ast.ExprsExpr; Typed, Mutable bool}{Exprs1: $2, Exprs2: $4, Typed: true, Mutable: true}
	}

op_lets :
	WALRUS { $$ = true }
	| '='  { $$ = false }

stmt_if :
	IF expr block maybe_else
	{
		$$ = &ast.IfStmt{If: $2, Then: $3, Else: $4}
		$$.SetPosition($1.Position())
	}

maybe_else :
	/* nothing */   { $$ = nil }
	| ELSE stmt_if  { $$ = $2  }
	| ELSE block    { $$ = $2  }

stmt_loop :
	LOOP block
	{
		$$ = &ast.LoopStmt{Stmt: $2}
		$$.SetPosition($1.Position())
	}

stmt_for :
	FOR for_content block
	{
		if el, ok := $2.(*ast.LoopStmt); ok {
			el.Stmt = $3
		} else if el, ok := $2.(*ast.ForStmt); ok {
			el.Stmt = $3
		} else if el, ok := $2.(*ast.CForStmt); ok {
			el.Stmt = $3
		}
		$$ = $2
		$$.SetPosition($1.Position())
	}

for_content :
	expr
	{
		$$ = &ast.LoopStmt{Expr: $1}
	}
	| IDENT IN expr_iterable
	{
		$$ = &ast.ForStmt{Vars: []string{$1.Lit}, Value: $3}
	}
	| IDENT ',' IDENT IN expr_iterable
	{
		$$ = &ast.ForStmt{Vars: []string{$1.Lit, $3.Lit}, Value: $5}
	}
	| opt_stmt_var_or_lets ';' opt_expr ';' opt_expr
	{
		$$ = &ast.CForStmt{Stmt1: $1, Expr2: $3, Expr3: $5}
	}

stmt_select :
	SELECT '{' stmt_select_content '}'
	{
		$$ = &ast.SelectStmt{Body: $3}
		$$.SetPosition($1.Position())
	}

stmt_select_content :
	opt_newlines opt_stmt_select_cases stmt_select_opt_default
	{
		$$ = &ast.SelectBodyStmt{Cases: $2, Default: $3}
	}

opt_stmt_select_cases :
	/* nothing */       { $$ = nil }
	| stmt_select_cases { $$ = $1  }

stmt_select_cases :
	stmt_select_case
	{
		$$ = []ast.Stmt{$1}
	}
	| stmt_select_cases stmt_select_case
	{
		$$ = append($$, $2)
	}

stmt_select_case :
	CASE stmt ':' compstmt
	{
		$$ = &ast.SelectCaseStmt{Expr: $2, Stmt: $4}
		$$.SetPosition($1.Position())
	}

stmt_select_opt_default :
	/* nothing */
	{ $$ = nil }
	| stmt_select_default { $$ = $1 }

stmt_select_default :
	DEFAULT ':' compstmt
	{
		$$ = $3
	}

stmt_switch :
	SWITCH expr '{' switch_content '}'
	{
		$4.(*ast.SwitchStmt).Expr = $2
		$$ = $4
		$$.SetPosition($1.Position())
	}

switch_content :
	opt_newlines opt_stmt_switch_cases stmt_switch_opt_default
	{
		$$ = &ast.SwitchStmt{Cases: $2, Default: $3}
	}

opt_stmt_switch_cases :
	/* nothing */
	{ $$ = nil }
	| stmt_switch_cases
	{
		$$ = $1
	}

stmt_switch_cases :
	stmt_switch_case
	{
		$$ = []ast.Stmt{$1}
	}
	| stmt_switch_cases stmt_switch_case
	{
		$$ = append($$, $2)
	}

stmt_switch_case :
	CASE opt_exprs ':' compstmt
	{
		$$ = &ast.SwitchCaseStmt{Exprs: $2, Stmt: $4}
		$$.SetPosition($1.Position())
	}

stmt_switch_opt_default :
	/* nothing */
	{ $$ = nil }
	| stmt_switch_default { $$ = $1 }

stmt_switch_default :
	DEFAULT ':' compstmt { $$ = $3 }

opt_func_return_expr_idents :
	{ $$ = nil }
	| type
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
	type
	{
		$$ = []*ast.FuncReturnValuesExpr{&ast.FuncReturnValuesExpr{TypeData: $1}}
	}
	| opt_func_return_expr_idents2 comma_opt_newlines type
	{
		$$ = append($1, &ast.FuncReturnValuesExpr{TypeData: $3})
	}

func_expr_idents :
	{
		$$ = []*ast.ParamExpr{}
	}
	| func_expr_idents_not_empty { $$ = $1 }

func_expr_idents_not_empty :
	func_expr_idents_last_untyped
	| func_expr_typed_idents

func_expr_untyped_ident :
	IDENT
	{
		$$ = &ast.ParamExpr{Name: $1.Lit}
	}

func_expr_typed_ident :
	expr_typed_ident
	{
		$$ = &ast.ParamExpr{Name: $1.Name, TypeData: $1.TypeData}
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
	/* nothing */ { $$ = &ast.ExprsExpr{Exprs: []ast.Expr{}} }
	| exprs       { $$ = $1  }

exprs :
	  expr                          { $$ = &ast.ExprsExpr{Exprs: []ast.Expr{$1}}  }
	| exprs comma_opt_newlines expr { $1.Exprs = append($1.Exprs, $3) }

opt_expr :
	/* nothing */ { $$ = nil }
	| expr        { $$ = $1  }

stmt_dbg :
	DBG '(' ')'
	{
		$$ = &ast.DbgStmt{Expr: nil}
		$$.SetPosition($1.Position())
	}
	| DBG '(' dbg_content ')'
	{
		$$ = $3
		$$.SetPosition($1.Position())
	}

dbg_content :
	type
	{
		$$ = &ast.DbgStmt{TypeData: $1}
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

element : expr

element_list :
	keyed_element                    { $$ = &ast.ExprsExpr{Exprs: []ast.Expr{$1}}  }
	| element_list ',' keyed_element { $1.Exprs = append($1.Exprs, $3) }

opt_element_list :
	/* nothing */  { $$ = nil }
	| element_list { $$ = $1  }

key : expr

keyed_element :
	element           { $$ = $1 }
	| key ':' element { $$ = $3 }

composite_lit :
	literal_type literal_value
	{
		if $1.Kind == ast.TypeSlice {
			$$ = &ast.ArrayExpr{TypeData: $1, Exprs: $2.(*ast.ExprsExpr)}
		} else {
			$$ = $2
		}
	}

literal_type : slice_type // array_type

array_length : expr

array_type : '[' array_length ']' element_type { $$ = $4 }

slice_type :
	EMPTYARR element_type
	{
		$$ = &ast.TypeStruct{Kind: ast.TypeSlice, SubType: $2}
	}

literal_value : '{' opt_element_list opt_comma '}' { $$ = $2 }

expr_array :
	EMPTYARR
	{
		$$ = &ast.ArrayExpr{Exprs: &ast.ExprsExpr{Exprs: []ast.Expr{}}}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}
	| '[' element_list ']'
	{
		$$ = &ast.ArrayExpr{Exprs: $2}
		if l, ok := yylex.(*Lexer); ok { $$.SetPosition(l.pos) }
	}

expr_ternary :
	expr '?' expr ':' expr
	{
		$$ = &ast.TernaryOpExpr{Expr: $1, Lhs: $3, Rhs: $5}
		$$.SetPosition($1.Position())
	}

expr_new :
	NEW '(' type ')'
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
	OPCHAN expr
	{
		$$ = &ast.ChanExpr{Rhs: $2}
		$$.SetPosition($2.Position())
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
	  NUMBER     { $$ = &ast.NumberExpr{Lit: $1.Lit} }
	| STRING     { $$ = &ast.StringExpr{Lit: $1.Lit} }
	| const_expr { $$ = &ast.ConstExpr{Value: $1.Lit} }

const_expr : TRUE | FALSE | NIL

expr_member_or_ident :
	expr_ident
	| expr_member

expr_typed_ident :
	expr_ident type
	{
		typeData := $2
		$$ = struct{Name string; TypeData *ast.TypeStruct}{Name: $1.(*ast.IdentExpr).Lit, TypeData: typeData}
	}
	| MUT expr_ident type
	{
		typeData := $3
		typeData.Mutable = true
		$$ = struct{Name string; TypeData *ast.TypeStruct}{Name: $2.(*ast.IdentExpr).Lit, TypeData: typeData}
	}

opt_ident :
	/* nothing */
	{ $$ = nil }
	| IDENT { $$ = &$1 }

expr_member :
	expr '.' IDENT
	{
		$$ = &ast.MemberExpr{Expr: $1, Name: $3.Lit}
		$$.SetPosition($1.Position())
	}

expr_callable :
	expr_call
	| expr_anon_call

expr_call :
	expr_ident expr_call_helper
	{
		$$ = &ast.CallExpr{Name: $1.(*ast.IdentExpr).Lit, Callable: &ast.Callable{SubExprs: $2.Exprs, VarArg: $2.VarArg}}
		$$.SetPosition($1.Position())
	}

expr_anon_call :
	expr expr_call_helper
	{
		$$ = &ast.AnonCallExpr{Expr: $1, Callable: &ast.Callable{SubExprs: $2.Exprs, VarArg: $2.VarArg}}
		$$.SetPosition($1.Position())
	}

expr_call_helper :
	'(' exprs VARARG ')'
	{
		$$ = struct{Exprs *ast.ExprsExpr; VarArg bool}{Exprs: $2, VarArg: true}
	}
	| '(' opt_exprs ')'
	{
		$$ = struct{Exprs *ast.ExprsExpr; VarArg bool}{Exprs: $2}
	}

unary_op :
	  '+' { $$ = "+" }
	| '-' { $$ = "-" }
	| '!' { $$ = "!" }
	| '^' { $$ = "^" }
	| '*' { $$ = "*" }
	| '&' { $$ = "&" }

expr_unary :
	unary_op expr %prec UNARY
	{
		if $1 == "&" {
			if el, ok := $2.(*ast.IdentExpr); ok {
				$$ = &ast.AddrExpr{Expr: el}
			} else if el, ok := $2.(*ast.MemberExpr); ok {
				$$ = el
			}
		} else if $1 == "*" {
			$$ = &ast.DerefExpr{Expr: $2}
		} else {
			$$ = &ast.UnaryExpr{Operator: $1, Expr: $2}
		}
		$$.SetPosition($2.Position())
	}

bin_op :
	  '+'         { $$ = "+"  }
	| '-'         { $$ = "-"  }
	| '*'         { $$ = "*"  }
	| '/'         { $$ = "/"  }
	| POW         { $$ = "**" }
	| '%'         { $$ = "%"  }
	| SHIFTLEFT   { $$ = "<<" }
	| SHIFTRIGHT  { $$ = ">>" }
	| '|'         { $$ = "|"  }
	| OROR        { $$ = "||" }
	| '&'         { $$ = "&"  }
	| ANDAND      { $$ = "&&" }
	| NEQ         { $$ = "!=" }
	| '>'         { $$ = ">"  }
	| GE          { $$ = ">=" }
	| '<'         { $$ = "<"  }
	| LE          { $$ = "<=" }
	| NILCOALESCE { $$ = "??" }
	| PLUSEQ      { $$ = "+=" }
	| MINUSEQ     { $$ = "-=" }
	| MULEQ       { $$ = "*=" }
	| DIVEQ       { $$ = "/=" }
	| ANDEQ       { $$ = "&=" }
	| OREQ        { $$ = "|=" }
	| OPCHAN      { $$ = "<-" }

expr_binary :
	expr bin_op expr
	{
		if $2 == "??" {
			$$ = &ast.NilCoalescingOpExpr{Lhs: $1, Rhs: $3}
		} else if $2 == "<-" {
			$$ = &ast.ChanExpr{Lhs: $1, Rhs: $3}
		} else if $2 == "+=" ||
			  $2 == "-=" ||
			  $2 == "*=" ||
			  $2 == "/=" ||
			  $2 == "&=" ||
			  $2 == "|=" {
			$$ = &ast.AssocExpr{Lhs: $1, Operator: $2, Rhs: $3}
		} else {
			$$ = &ast.BinOpExpr{Lhs: $1, Operator: $2, Rhs: $3}
		}
		$$.SetPosition($1.Position())
	}
	| expr EQEQ expr
	{
		$$ = &ast.BinOpExpr{Lhs: $1, Operator: "==", Rhs: $3}
		$$.SetPosition($1.Position())
	}
	| expr IN expr
	{
		$$ = &ast.IncludeExpr{ItemExpr: $1, ListExpr: &ast.SliceExpr{Value: $3, Begin: nil, End: nil}}
		$$.SetPosition($1.Position())
	}
	| expr_assoc

op_assoc1 :
	PLUSPLUS     { $$ = "++" }
	| MINUSMINUS { $$ = "--" }

expr_assoc :
	expr op_assoc1
	{
		$$ = &ast.AssocExpr{Lhs: $1, Operator: $2}
		$$.SetPosition($1.Position())
	}

expr_func :
	FUNC opt_ident '(' func_expr_args ')' opt_func_return_expr_idents block
	{
		f := &ast.FuncExpr{Params: $4.Params, Returns: $6, Stmt: $7, VarArg: $4.VarArg}
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
	func_expr_idents_last_untyped VARARG type
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
	MAKE '(' type ')'
	{
		$$ = &ast.MakeExpr{TypeData: $3}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' type ',' expr ')'
	{
		$$ = &ast.MakeExpr{TypeData: $3, LenExpr: $5}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' type ',' expr ',' expr ')'
	{
		$$ = &ast.MakeExpr{TypeData: $3, LenExpr: $5, CapExpr: $7}
		$$.SetPosition($1.Position())
	}
	| MAKE '(' TYPE IDENT ',' expr ')'
	{
		$$ = &ast.MakeTypeExpr{Name: $4.Lit, Type: $6}
		$$.SetPosition($1.Position())
	}

type :
	type_name
	| type_lit

type_lit :
	pointer_type
	| array_type
	| slice_type
	| typed_slice_count
	| map_type
	| channel_type
	| struct_type

type_name :
	  IDENT           { $$ = &ast.TypeStruct{Name: $1.Lit} }
	| qualified_ident { $$ = $1 }

package_name: IDENT

qualified_ident :
	package_name '.' IDENT
	{
		$$ =  &ast.TypeStruct{Env: []string{$1.Lit}, Name: $3.Lit}
	}

pointer_type :
	'*' type
	{
		if $2.Kind == ast.TypeDefault {
			$2.Kind = ast.TypePtr
			$$ = $2
		} else {
			$$ = &ast.TypeStruct{Kind: ast.TypePtr, SubType: $2}
		}
	}

struct_type :
	STRUCT '{' type_struct_content '}'
	{
		$$ = $3
	}

channel_type :
	CHAN type
	{
		if $2.Kind == ast.TypeDefault {
			$2.Kind = ast.TypeChan
			$$ = $2
		} else {
			$$ = &ast.TypeStruct{Kind: ast.TypeChan, SubType: $2}
		}
	}

key_type : type

element_type : type

map_type :
	MAP '[' key_type ']' element_type
	{
		$$ = &ast.TypeStruct{Kind: ast.TypeMap, Key: $3, SubType: $5}
	}

type_struct_content :
	opt_newlines type_data_struct opt_newlines
	{
		$$ = $2
	}

type_data_struct :
	expr_typed_ident
	{
		$$ = &ast.TypeStruct{Kind: ast.TypeStructType, StructNames: []string{$1.Name}, StructTypes: []*ast.TypeStruct{$1.TypeData}}
	}
	| type_data_struct comma_opt_newlines expr_typed_ident
	{
		if $1 == nil {
			yylex.Error("syntax error: unexpected ','")
		}
		$$.StructNames = append($$.StructNames, $3.Name)
		$$.StructTypes = append($$.StructTypes, $3.TypeData)
	}

typed_slice_count :
	slice_count type
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
	  EMPTYARR             { $$ = 1      }
	| EMPTYARR slice_count { $$ = $2 + 1 }

expr_map :
	map_type expr_map_container
	{
		$2.TypeData = $1
		$$ = $2
		$$.SetPosition($2.Position())
	}
	| expr_map_container
	{
		$$ = $1
		$$.SetPosition($1.Position())
	}

expr_map_container :
	'{' expr_map_content '}'
	{
		$$ = $2
	}

expr_map_content :
	opt_newlines
	{
		$$ = &ast.MapExpr{Keys: &ast.ExprsExpr{Exprs: []ast.Expr{}}, Values: &ast.ExprsExpr{Exprs: []ast.Expr{}}}
	}
	| opt_newlines expr_map_content_helper opt_comma_opt_newlines
	{
		$$ = $2
	}

expr_map_content_helper :
	expr_map_key_value
	{
		$$ = &ast.MapExpr{Keys: &ast.ExprsExpr{Exprs: []ast.Expr{$1[0]}}, Values:  &ast.ExprsExpr{Exprs: []ast.Expr{$1[1]}}}
	}
	| expr_map_content_helper comma_opt_newlines expr_map_key_value
	{
		$$.Keys.Exprs = append($$.Keys.Exprs, $3[0])
		$$.Values.Exprs = append($$.Values.Exprs, $3[1])
	}

expr_map_key_value :
	expr ':' expr
	{
		$$ = []ast.Expr{$1, $3}
	}

expr_item_or_slice :
	expr '[' slice ']'
	{
		if el, ok := $3.(*ast.SliceExpr); ok {
			el.Value = $1
		} else if el, ok := $3.(*ast.ItemExpr); ok {
			el.Value = $1
		}
		$$ = $3
               	$$.SetPosition($1.Position())
	}

slice :
	  expr ':' expr { $$ = &ast.SliceExpr{Begin: $1, End: $3}  }
	| expr ':'      { $$ = &ast.SliceExpr{Begin: $1, End: nil} }
	|      ':' expr { $$ = &ast.SliceExpr{Begin: nil, End: $2} }
	| expr          { $$ = &ast.ItemExpr{Index: $1}            }

expr_idents :
	expr_ident
	{
		$$ = []string{$1.(*ast.IdentExpr).Lit}
	}
	| expr_idents comma_opt_newlines expr_ident
	{
		$$ = append($1, $3.(*ast.IdentExpr).Lit)
	}

expr_ident :
	IDENT
	{
		$$ = &ast.IdentExpr{Lit: $1.Lit}
		$$.SetPosition($1.Position())
	}

comma : ','

opt_comma:
	/* nothing */
	| comma

opt_term :
	/* nothing */
	| term

term :
	newlines
	| ';' opt_newlines

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
