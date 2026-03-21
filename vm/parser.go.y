// Packer vm is
%{
package vm
%}

%union {
  expr Expr
  exprs []Expr
  str string
  lit any
}

%type<expr> stmt
%type<expr> expr
%type<exprs> exprs
%token<str> ident
%token<lit> lit cfor in cyield cif celsif celse cunless
%token eqeq neq le ge andand oror

%left oror
%left andand
%left eqeq neq
%left '<' '>' le ge
%left '+' '-'
%left '*' '/'
%right '!'

%%

stmt :  cfor ident in expr
     {
       yylex.(*Lexer).e = &ForExpr{$2, "", $4}
     }
     | cfor ident ',' ident in expr
     {
       yylex.(*Lexer).e = &ForExpr{$2, $4, $6}
     }
     | cyield ident
     {
       yylex.(*Lexer).e = &YieldExpr{$2}
     }
     | cyield
     {
       yylex.(*Lexer).e = &YieldExpr{}
     }
     | cif expr
     {
       yylex.(*Lexer).e = &IfExpr{Cond: $2}
     }
     | celsif expr
     {
       yylex.(*Lexer).e = &IfExpr{Cond: $2}
     }
     | celse
     {
       yylex.(*Lexer).e = &IfExpr{}
     }
     | cunless expr
     {
       yylex.(*Lexer).e = &IfExpr{Cond: $2, Neg: true}
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
     | '!' expr
     {
       $$ = &NotExpr{$2}
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
     | expr eqeq expr
     {
       $$ = &BinOpExpr{"==", $1, $3}
     }
     | expr neq expr
     {
       $$ = &BinOpExpr{"!=", $1, $3}
     }
     | expr '<' expr
     {
       $$ = &BinOpExpr{"<", $1, $3}
     }
     | expr '>' expr
     {
       $$ = &BinOpExpr{">", $1, $3}
     }
     | expr le expr
     {
       $$ = &BinOpExpr{"<=", $1, $3}
     }
     | expr ge expr
     {
       $$ = &BinOpExpr{">=", $1, $3}
     }
     | expr andand expr
     {
       $$ = &BinOpExpr{"&&", $1, $3}
     }
     | expr oror expr
     {
       $$ = &BinOpExpr{"||", $1, $3}
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
