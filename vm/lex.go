package vm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/scanner"
)

type Lexer struct {
	s *scanner.Scanner
	e Expr
}

func (l *Lexer) init(reader *strings.Reader) {
	l.s.Init(reader)
}

func (l *Lexer) Lex(v *yySymType) int {
	var err error
	var tok int
	i := l.s.Scan()
	switch i {
	case scanner.Ident:
		v.str = l.s.TokenText()
		switch v.str {
		case "for":
			tok = FOR
		case "in":
			tok = IN
		default:
			tok = IDENT
		}
	case scanner.Int:
		tok = LIT
		v.lit, err = strconv.ParseInt(l.s.TokenText(), 10, 64)
		if err != nil {
			return -1
		}
	case scanner.Float:
		tok = LIT
		v.lit, _ = strconv.ParseFloat(l.s.TokenText(), 64)
		if err != nil {
			return -1
		}
	case scanner.String:
		tok = LIT
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
