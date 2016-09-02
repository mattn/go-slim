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
		case int, int32, int64:
			li, err := strconv.ParseInt(fmt.Sprint(lhs), 10, 64)
			if err != nil {
				return nil, err
			}
			ri, err := strconv.ParseInt(fmt.Sprint(rhs), 10, 64)
			if err != nil {
				return nil, err
			}
			switch t.Op {
			case "+":
				return li + ri, nil
			case "-":
				return li - ri, nil
			case "*":
				return li * ri, nil
			case "/":
				return li / ri, nil
			}
			return nil, errors.New("unknown operator")
		case float32, float64:
			lf, err := strconv.ParseFloat(fmt.Sprint(lhs), 64)
			if err != nil {
				return nil, err
			}
			rf, err := strconv.ParseFloat(fmt.Sprint(rhs), 64)
			if err != nil {
				return nil, err
			}
			switch t.Op {
			case "+":
				return lf + rf, nil
			case "-":
				return lf - rf, nil
			case "*":
				return lf * rf, nil
			case "/":
				return lf / rf, nil
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
	case *ItemExpr:
		lhs, err := v.Eval(t.Lhs)
		if err != nil {
			return nil, err
		}
		rv := reflect.ValueOf(lhs)

	deref_item:
		for {
			switch rv.Kind() {
			case reflect.Interface, reflect.Ptr:
				rv = rv.Elem()
			default:
				break deref_item
			}
		}

		if !rv.IsValid() {
			return nil, errors.New("cannot reference item")
		}

		rhs, err := v.Eval(t.Index)
		if err != nil {
			return nil, err
		}

		if rv.Kind() == reflect.Struct {
			rv = rv.FieldByName(fmt.Sprint(rhs))
			if !rv.IsValid() {
				return nil, errors.New("cannot reference item")
			}
			return rv.Interface(), nil
		} else if rv.Kind() == reflect.Map {
			rv = rv.MapIndex(reflect.ValueOf(fmt.Sprint(rhs)))
			if !rv.IsValid() {
				return nil, errors.New("cannot reference item")
			}
			return rv.Interface(), nil
		} else if rv.Kind() == reflect.Slice && reflect.TypeOf(rhs).Kind() == reflect.Int64 {
			rv = rv.Index(int(rhs.(int64)))
			if !rv.IsValid() {
				return nil, errors.New("cannot reference item")
			}
			return rv.Interface(), nil
		}
		return nil, errors.New("cannot reference item")
	case *MemberExpr:
		lhs, err := v.Eval(t.Lhs)
		if err != nil {
			return nil, err
		}
		rv := reflect.ValueOf(lhs)

	deref_member:
		for {
			switch rv.Kind() {
			case reflect.Interface, reflect.Ptr:
				rv = rv.Elem()
			default:
				break deref_member
			}
		}

		if !rv.IsValid() {
			return nil, errors.New("cannot reference member")
		}

		if rv.Kind() == reflect.Struct {
			rv = rv.FieldByName(t.Name)
			if !rv.IsValid() {
				return nil, errors.New("cannot reference member")
			}
			return rv.Interface(), nil
		} else if rv.Kind() == reflect.Map {
			rv = rv.MapIndex(reflect.ValueOf(t.Name))
			if !rv.IsValid() {
				return nil, errors.New("cannot reference member")
			}
			return rv.Interface(), nil
		}
		return nil, errors.New("cannot reference member")

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
