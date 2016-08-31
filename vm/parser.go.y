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
%type<expr> value
%type<exprs> exprs
%token<str> IDENT
%token<lit> LIT FOR IN

%%

stmt :  FOR IDENT IN IDENT
     {
       yylex.(*Lexer).e = &ForExpr{$2, "", $4}
     }
     | FOR IDENT ',' IDENT IN IDENT
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

expr : value
     {
       $$ = $1
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
     ;

value : IDENT
      {
        $$ = &IdentExpr{$1}
      }
      | LIT
      {
        $$ = &LitExpr{$1}
      }
      ;
%%

/* vim: set et sw=2: */
