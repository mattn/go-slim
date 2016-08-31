package slim

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"unicode"

	"github.com/mattn/go-slim/vm"
)

type state int

const (
	sNeutral state = iota
	sName
	sId
	sClass
	sAttrKey
	sAttrValue
	sEq
	sExpr
)

var emptyElement = []string{
	"doctype",
	"area",
	"base",
	"basefont",
	"br",
	"col",
	"frame",
	"hr",
	"img",
	"input",
	"isindex",
	"link",
	"meta",
	"param",
	"embed",
	"keygen",
	"command",
}

type Value interface{}

type Func func(Value) (Value, error)

type Funcs map[string]Func
type Values map[string]Value

type Attr struct {
	Name  string
	Value string
}

type Node struct {
	Name     string
	Id       string
	Class    []string
	Attr     []Attr
	Expr     string
	Children []*Node
}

func (n *Node) NewChild() *Node {
	n.Children = append(n.Children, new(Node))
	return n.Children[len(n.Children)-1]
}

type stack struct {
	n    int
	node *Node
}

func isEmptyElement(n string) bool {
	for _, e := range emptyElement {
		if e == n {
			return true
		}
	}
	return false
}

func printNode(out io.Writer, v *vm.VM, n *Node, indent int) error {
	if n.Name == "" {
		for _, c := range n.Children {
			if err := printNode(out, v, c, indent); err != nil {
				return err
			}
		}
	} else {
		// FIXME
		doctype := n.Name == "doctype"
		if doctype {
			out.Write([]byte(strings.Repeat(" ", indent*2) + "<!" + n.Name + " html"))
			n.Attr = nil
		} else {
			out.Write([]byte(strings.Repeat(" ", indent*2) + "<" + n.Name))
		}
		if n.Id != "" {
			out.Write([]byte(" id=\"" + n.Id + "\""))
		}
		if len(n.Class) > 0 {
			out.Write([]byte(" class="))
			for i, c := range n.Class {
				if i > 0 {
					out.Write([]byte(" "))
				}
				out.Write([]byte(c))
			}
		}
		if len(n.Attr) > 0 {
			for i, a := range n.Attr {
				if i > 0 {
					out.Write([]byte(" "))
				}
				if a.Value == "" {
					out.Write([]byte(" " + a.Name))
				} else {
					fmt.Fprintf(out, " %s=%q", a.Name, a.Value)
				}
			}
		}
		if !isEmptyElement(n.Name) {
			out.Write([]byte(">"))
			cr := true
			if n.Expr != "" {
				expr, err := v.Compile(n.Expr)
				if err != nil {
					return err
				}
				fe, ok := expr.(*vm.ForExpr)
				if ok {
					rhs, ok := v.Get(fe.Rhs)
					if !ok {
						return errors.New("invalid token: " + fe.Rhs)
					}
					ra := reflect.ValueOf(rhs)
					switch ra.Type().Kind() {
					case reflect.Array, reflect.Slice:
					default:
						return errors.New("can't iterate: " + fe.Rhs)
					}
					out.Write([]byte("\n"))
					l := ra.Len()
					for i := 0; i < l; i++ {
						x := ra.Index(i).Interface()
						v.Set(fe.Lhs1, x)
						for _, c := range n.Children {
							if err := printNode(out, v, c, indent+1); err != nil {
								return err
							}
						}
					}
				} else {
					r, err := v.Eval(expr)
					if err != nil {
						return err
					}
					out.Write([]byte(fmt.Sprint(r)))
					cr = false
				}
			} else if len(n.Children) > 0 {
				out.Write([]byte("\n"))
				for _, c := range n.Children {
					if err := printNode(out, v, c, indent+1); err != nil {
						return err
					}
				}
			} else if cr {
				out.Write([]byte("\n"))
			}
			if cr {
				out.Write([]byte(strings.Repeat(" ", indent*2)))
			}
			out.Write([]byte("</" + n.Name + ">\n"))
		} else if doctype {
			out.Write([]byte(">\n"))
		} else {
			out.Write([]byte("/>\n"))
		}
	}
	return nil
}

type Template struct {
	root *Node
	fm   Funcs
}

func ParseFile(name string) (*Template, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}

func Parse(in io.Reader) (*Template, error) {
	scanner := bufio.NewScanner(in)
	root := new(Node)
	node := root
	stk := []stack{}
	last := -1
	for scanner.Scan() {
		l := scanner.Text()
		rs := []rune(l)
		st := sNeutral
		tag := ""
		id := ""
		class := ""
		name := ""
		value := ""
		for n := 0; n < len(rs); n++ {
			eol := n == len(rs)-1
			r := rs[n]
			switch st {
			case sNeutral:
				if unicode.IsSpace(r) {
					break
				}
				if r == '-' {
					st = sExpr
					break
				}
				st = sName
				tag += string(r)
				if n > last {
					node = node.NewChild()
					last = n
					stk = append(stk, stack{n: n, node: node})
				} else if n == last {
					cur := root
					for cur != nil {
						var tmp *Node
						if len(cur.Children) == 0 {
							break
						}
						tmp = cur.Children[len(cur.Children)-1]
						if tmp == nil || tmp == node {
							break
						}
						cur = tmp
					}
					node = cur.NewChild()
					last = n
				} else if n < last {
					node = nil
					for i := 0; i < len(stk)-1; i++ {
						if stk[i+1].n >= n {
							node = stk[i].node
							stk = stk[:i+1]
							break
						}
					}
					if node == nil {
						node = root.NewChild()
						stk = stk[:1]
					} else {
						node = node.NewChild()
					}
					last = n
				}
				node.Name = tag
			case sName:
				if eol {
					tag += string(r)
					node.Name = tag
					break
				}
				switch r {
				case '#':
					node.Name = tag
					st = sId
				case '.':
					node.Name = tag
					st = sClass
				default:
					if !unicode.IsLetter(r) {
						node.Name = tag
						st = sAttrKey
					} else {
						tag += string(r)
					}
				}
			case sId:
				if eol {
					if unicode.IsLetter(r) {
						id += string(r)
						node.Id = id
					}
					break
				}
				switch r {
				case '.':
					node.Id = id
					st = sClass
				default:
					if !unicode.IsLetter(r) {
						node.Id = id
						st = sEq
					} else {
						id += string(r)
					}
				}
			case sClass:
				if eol {
					if unicode.IsLetter(r) {
						class += string(r)
						node.Class = append(node.Class, class)
					}
					break
				}
				switch r {
				case '.':
					if class != "" {
						node.Class = append(node.Class, class)
						class = ""
					}
				default:
					if !unicode.IsLetter(r) {
						if class != "" {
							node.Class = append(node.Class, class)
						}
						st = sEq
					} else {
						class += string(r)
					}
				}
			case sAttrKey:
				if eol {
					if !unicode.IsSpace(r) {
						name += string(r)
						node.Attr = append(node.Attr, Attr{Name: name, Value: ""})
					}
					break
				}
				switch r {
				case '=':
					if name == "" {
						st = sExpr
					} else {
						st = sAttrValue
					}
				default:
					if !unicode.IsLetter(r) {
						node.Attr = append(node.Attr, Attr{Name: name, Value: ""})
						st = sEq
					} else {
						name += string(r)
					}
				}
			case sAttrValue:
				if eol {
					if unicode.IsLetter(r) || r == '"' {
						value += string(r)
						if value[0] == '"' && value[len(value)-1] == '"' {
							value = value[1 : len(value)-1]
						}
						node.Attr = append(node.Attr, Attr{Name: name, Value: value})
					}
					break
				}
				if unicode.IsSpace(r) {
					node.Attr = append(node.Attr, Attr{Name: name, Value: value})
					name = ""
					value = ""
				} else {
					value += string(r)
				}
			case sEq:
				if r == '=' {
					st = sExpr
				} else if !unicode.IsSpace(r) {
					return nil, errors.New("invalid token: " + l[n:])
				}
			case sExpr:
				node.Expr += string(r)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &Template{root, nil}, nil
}

func (t *Template) FuncMap(m Funcs) {
	t.fm = m
}

func (t *Template) Execute(out io.Writer, value interface{}) error {
	v := vm.New()

	if t.fm != nil {
		for key, val := range t.fm {
			v.Set(key, val)
		}
	}
	if value != nil {
		rv := reflect.ValueOf(value)
		rt := rv.Type()
		if rt.Kind() == reflect.Map {
			for _, rk := range rv.MapKeys() {
				v.Set(rk.String(), rv.MapIndex(rk).Interface())
			}
		} else if rt.Kind() == reflect.Struct {
			for i := 0; i < rt.NumField(); i++ {
				v.Set(rt.Field(i).Name, rv.Field(i).Interface())
			}
		}
	}
	return printNode(out, v, t.root, 0)
}
