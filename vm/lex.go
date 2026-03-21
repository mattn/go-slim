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
		case "yield":
			tok = cyield
		case "if":
			tok = cif
		case "elsif":
			tok = celsif
		case "else":
			tok = celse
		case "unless":
			tok = cunless
		case "true":
			tok = lit
			v.lit = true
		case "false":
			tok = lit
			v.lit = false
		case "nil":
			tok = lit
			v.lit = nil
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
		v.lit, err = strconv.ParseFloat(l.s.TokenText(), 64)
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
	case '=':
		if l.s.Peek() == '=' {
			l.s.Scan()
			tok = eqeq
		} else {
			tok = int(i)
		}
	case '!':
		if l.s.Peek() == '=' {
			l.s.Scan()
			tok = neq
		} else {
			tok = int(i)
		}
	case '<':
		if l.s.Peek() == '=' {
			l.s.Scan()
			tok = le
		} else {
			tok = int(i)
		}
	case '>':
		if l.s.Peek() == '=' {
			l.s.Scan()
			tok = ge
		} else {
			tok = int(i)
		}
	case '&':
		if l.s.Peek() == '&' {
			l.s.Scan()
			tok = andand
		} else {
			tok = int(i)
		}
	case '|':
		if l.s.Peek() == '|' {
			l.s.Scan()
			tok = oror
		} else {
			tok = int(i)
		}
	default:
		tok = int(i)
	}
	return tok
}

func (l *Lexer) Error(e string) {
	fmt.Fprintf(os.Stderr, "syntax error: %s\n", e)
}
