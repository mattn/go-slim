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
			rf := reflect.ValueOf(f)
			args := []reflect.Value{}
			for _, arg := range t.Exprs {
				arg, err := v.Eval(arg)
				if err != nil {
					return nil, err
				}
				args = append(args, reflect.ValueOf(arg))
			}
			rets := rf.Call(args)
			if len(rets) == 0 {
				return nil, nil
			}
			vals := []interface{}{}
			for _, ret := range rets {
				vals = append(vals, ret.Interface())
			}
			if len(rets) == 1 {
				return vals[0], nil
			}
			if err, ok := vals[1].(error); ok {
				return vals[0], err
			}
			return vals[0], nil
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
