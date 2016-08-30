%{
package vm
%}

%union {
  stmt Expr
  str string
  expr Expr
  lit interface{}
}

%type<expr> expr
%type<stmt> stmt
%type<expr> rhs
%token<str> IDENT
%token<lit> LIT FOR RANGE

%%

stmt :  FOR IDENT ',' IDENT ':' '=' RANGE IDENT
     {
       yylex.(*Lexer).e = &RangeExpr{$2, $4, $8}
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
