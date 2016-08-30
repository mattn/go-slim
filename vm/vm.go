package vm

import (
	"errors"
	"fmt"
	"strings"
	"text/scanner"
)

type VM struct {
	env map[string]interface{}
}

func New() *VM {
	return &VM{make(map[string]interface{})}
}

func (v *VM) Set(n string, vv interface{}) {
	v.env[n] = vv
}

func invoke(v *VM, expr Expr) (interface{}, error) {
	switch t := expr.(type) {
	case *IdentExpr:
		if r, ok := v.env[t.name]; ok {
			return r, nil
		}
		return nil, errors.New("invalid token")
	case *LitExpr:
		return t.value, nil
	}
	return nil, nil
}

func (v *VM) Run(s string) (interface{}, error) {
	lex := &Lexer{new(scanner.Scanner), nil}
	lex.s.Init(strings.NewReader(s))
	if yyParse(lex) != 0 {
		return nil, fmt.Errorf("syntax error: %s", s)
	}
	return invoke(v, lex.e)
}
