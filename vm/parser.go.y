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
