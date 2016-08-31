package vm

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
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
	case *BinOpExpr:
		lhs, err := v.Eval(t.Lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := v.Eval(t.Rhs)
		if err != nil {
			return nil, err
		}
		switch vt := lhs.(type) {
		case string:
			switch t.Op {
			case "+":
				return vt + fmt.Sprint(rhs), nil
			}
			return nil, errors.New("unknown operator")
		case int64:
			i, err := strconv.ParseInt(fmt.Sprint(rhs), 10, 64)
			if err != nil {
				return nil, err
			}
			switch t.Op {
			case "+":
				return vt + i, nil
			case "-":
				return vt - i, nil
			case "*":
				return vt * i, nil
			case "/":
				return vt / i, nil
			}
			return nil, errors.New("unknown operator")
		case float64:
			f, err := strconv.ParseFloat(fmt.Sprint(rhs), 64)
			if err != nil {
				return nil, err
			}
			switch t.Op {
			case "+":
				return vt + f, nil
			case "-":
				return vt - f, nil
			case "*":
				return vt * f, nil
			case "/":
				return vt / f, nil
			}
			return nil, errors.New("unknown operator")
		default:
			return nil, errors.New("invalid type conversion")
		}
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
