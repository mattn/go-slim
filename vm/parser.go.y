%{
package vm
%}

%union {
  expr Expr
  str string
  lit interface{}
}

%type<expr> stmt
%type<expr> expr
%type<expr> rhs
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

expr : rhs
     {
       $$ = $1
     }
     | IDENT '(' expr ')'
     {
       $$ = &CallExpr{$1, $3}
     }
     ;

rhs : IDENT
    {
      $$ = &IdentExpr{$1}
    }
    | LIT
    {
      $$ = &LitExpr{$1}
    }
    ;
%%
