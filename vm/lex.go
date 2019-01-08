package vm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/scanner"
)

// Lexer is a lexer.
type Lexer struct {
	s *scanner.Scanner
	e Expr
}

func (l *Lexer) init(reader *strings.Reader) {
	l.s.Init(reader)
}

// Lex parse the token.
func (l *Lexer) Lex(v *yySymType) int {
	var err error
	var tok int
	i := l.s.Scan()
	switch i {
	case scanner.Ident:
		v.str = l.s.TokenText()
		switch v.str {
		case "for":
			tok = cfor
		case "in":
			tok = in
		default:
			tok = ident
		}
	case scanner.Int:
		tok = lit
		v.lit, err = strconv.ParseInt(l.s.TokenText(), 10, 64)
		if err != nil {
			return -1
		}
	case scanner.Float:
		tok = lit
		v.lit, _ = strconv.ParseFloat(l.s.TokenText(), 64)
		if err != nil {
			return -1
		}
	case scanner.String:
		tok = lit
		s := l.s.TokenText()
		if len(s) >= 2 {
			v.lit = s[1 : len(s)-1]
		}
	case scanner.EOF:
		tok = 0
	default:
		tok = int(i)
	}
	return tok
}

func (l *Lexer) Error(e string) {
	fmt.Fprintf(os.Stderr, "syntax error: %s\n", e)
}
