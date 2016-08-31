package vm

import (
	"errors"
	"fmt"
	"reflect"
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

func (v *VM) Get(n string) (interface{}, bool) {
	val, ok := v.env[n]
	return val, ok
}

func (v *VM) Eval(expr Expr) (interface{}, error) {
	switch t := expr.(type) {
	case *IdentExpr:
		if r, ok := v.env[t.Name]; ok {
			return r, nil
		}
		return nil, errors.New("invalid token: " + t.Name)
	case *LitExpr:
		return t.Value, nil
	case *CallExpr:
		if f, ok := v.env[t.Name]; ok {
			arg, err := v.Eval(t.Expr)
			if err != nil {
				return nil, err
			}
			rf := reflect.ValueOf(f)
			ret := rf.Call([]reflect.Value{reflect.ValueOf(arg)})
			if len(ret) == 0 {
				return nil, nil
			}
			return ret[0].Interface(), nil
		}
		return nil, errors.New("invalid token: " + t.Name)
	}
	return nil, nil
}

func (v *VM) Compile(s string) (Expr, error) {
	lex := &Lexer{new(scanner.Scanner), nil}
	lex.s.Init(strings.NewReader(s))
	if yyParse(lex) != 0 {
		return nil, fmt.Errorf("syntax error: %s", s)
	}
	return lex.e, nil
}
