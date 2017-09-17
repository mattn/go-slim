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
%token<str> IDENT
%token<lit> LIT FOR IN

%%

stmt :  FOR IDENT IN expr
     {
       yylex.(*Lexer).e = &ForExpr{$2, "", $4}
     }
     | FOR IDENT ',' IDENT IN expr
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

expr : LIT
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
     | IDENT '(' exprs ')'
     {
       $$ = &CallExpr{$1, $3}
     }
     | expr '.' IDENT '(' exprs ')'
     {
       $$ = &MethodCallExpr{Lhs: $1, Name: $3, Exprs: $5}
     }
     | expr '.' IDENT
     {
       $$ = &MemberExpr{Lhs: $1, Name: $3}
     }
     | expr '[' expr ']'
     {
       $$ = &ItemExpr{Lhs: $1, Index: $3}
     }
     | IDENT
     {
       $$ = &IdentExpr{$1}
     }
     ;

%%

/* vim: set et sw=2: */
