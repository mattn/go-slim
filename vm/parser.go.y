// Packer vm is
%{
package vm
%}

%union {
  expr Expr
  exprs []Expr
  str string
  lit interface{}
}

%type<expr> stmt
%type<expr> expr
%type<exprs> exprs
%token<str> ident
%token<lit> lit for in

%%

stmt :  for ident in expr
     {
       yylex.(*Lexer).e = &ForExpr{$2, "", $4}
     }
     | for ident ',' ident in expr
     {
       yylex.(*Lexer).e = &ForExpr{$2, $4, $6}
     }
     | expr
     {
       yylex.(*Lexer).e = $1
     }
     ;

exprs :
      {
          $$ = nil
      }
      | expr 
      {
          $$ = []Expr{$1}
      }
      | exprs ',' expr
      {
          $$ = append($1, $3)
      }
      ;

expr : lit
     {
       $$ = &LitExpr{$1}
     }
     | '(' expr ')'
     {
       $$ = $2
     }
     | expr '+' expr
     {
       $$ = &BinOpExpr{"+", $1, $3}
     }
     | expr '-' expr
     {
       $$ = &BinOpExpr{"-", $1, $3}
     }
     | expr '*' expr
     {
       $$ = &BinOpExpr{"*", $1, $3}
     }
     | expr '/' expr
     {
       $$ = &BinOpExpr{"/", $1, $3}
     }
     | ident '(' exprs ')'
     {
       $$ = &CallExpr{$1, $3}
     }
     | expr '.' ident '(' exprs ')'
     {
       $$ = &MethodCallExpr{LHS: $1, Name: $3, Exprs: $5}
     }
     | expr '.' ident
     {
       $$ = &MemberExpr{LHS: $1, Name: $3}
     }
     | expr '[' expr ']'
     {
       $$ = &ItemExpr{LHS: $1, Index: $3}
     }
     | ident
     {
       $$ = &IdentExpr{$1}
     }
     ;

%%

/* vim: set et sw=2: */
