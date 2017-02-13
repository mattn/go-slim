package slim

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/mattn/go-slim/vm"
)

type state int

const (
	sNeutral state = iota
	sTag
	sId
	sClass
	sAttrKey
	sAttrValue
	sEq
	sText
	sComment
	sExpr
)

var SPACE []byte = []byte(" ")

var NEW_LINE []byte = []byte("\n")
var LESS_THAN_SLASH []byte = []byte("</")
var GREATER_THAN_NEW_LINE []byte = []byte(">\n")
var SLASH_GREATER_THAN_NEW_LINE []byte = []byte ("/>\n")
var DOUBLE_QUOTE []byte = []byte ("\"")
var GREATER_THAN []byte = []byte (">")
var LESS_THAN []byte = []byte ("<")

type Empty struct {}
var emptyElement = map[string]struct{}{
	"doctype":Empty{},
	"area":Empty{},
	"base":Empty{},
	"basefont":Empty{},
	"br":Empty{},
	"col":Empty{},
	"frame":Empty{},
	"hr":Empty{},
	"img":Empty{},
	"input":Empty{},
	"isindex":Empty{},
	"link":Empty{},
	"meta":Empty{},
	"param":Empty{},
	"embed":Empty{},
	"keygen":Empty{},
	"command":Empty{},
}

type Value interface{}

type Func func(...Value) (Value, error)

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
	Text     string
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
	_, ok := emptyElement[n]
	return ok
}

var rubyInlinePattern = regexp.MustCompile(`#{[^}]*}`)

func rubyInline(v *vm.VM, s string) (string, error) {
	var fail error
	text := rubyInlinePattern.ReplaceAllStringFunc(s, func(s string) string {
		expr, err := v.Compile(s[2 : len(s)-1])
		if err != nil {
			fail = err
			return ""
		}
		iv, err := v.Eval(expr)
		if err != nil {
			fail = err
			return ""
		}
		return fmt.Sprint(iv)
	})
	if fail != nil {
		return "", fail
	}
	return text, nil
}

// byteRepeat same as bytes.Repeat but Write to the io.Writer
func bytesRepeat (out io.Writer, b []byte, count int) {
	for i :=0; i < count; i++ {
		out.Write(b)
	}
}

func printNode(out io.Writer, v *vm.VM, n *Node, indent int) error {
	if n.Name == "" && n.Expr == "" {
		for _, c := range n.Children {
			if err := printNode(out, v, c, indent); err != nil {
				return err
			}
		}
	} else if n.Name == "/" {
		return nil
	} else if n.Name == "/!" {
		bytesRepeat(out, SPACE, indent * 2)
		out.Write([]byte("<!-- "))
		out.Write([]byte(n.Text))
		out.Write([]byte(" -->\n"))
	} else {
		// FIXME
		doctype := n.Name == "doctype"
		if doctype {
			bytesRepeat(out, SPACE, indent*2)
			out.Write([]byte("<!"))
			out.Write([]byte(n.Name))
			out.Write([]byte(" html"))
		} else if n.Name != "" {
			bytesRepeat(out, SPACE, indent*2)
			out.Write(LESS_THAN)
			if n.Name[len(n.Name)-1] == ':' {
				name := n.Name[:len(n.Name)-1]
				if name == "javascript" {
					name = "script"
				}
				out.Write([]byte(name))
			} else {
				out.Write([]byte(n.Name))
			}
		}
		if n.Id != "" {
			out.Write([]byte(" id=\""))
			out.Write([]byte(n.Id))
			out.Write(DOUBLE_QUOTE)
		}
		if len(n.Class) > 0 {
			out.Write([]byte(" class=\""))
			for i, c := range n.Class {
				if i > 0 {
					out.Write(SPACE)
				}
				out.Write([]byte(c))
			}
			out.Write(DOUBLE_QUOTE)
		}
		if len(n.Attr) > 0 && !doctype {
			for _, a := range n.Attr {
				if a.Value == "" {
					out.Write(SPACE)
					out.Write([]byte(a.Name))
				} else {
					value, err := rubyInline(v, a.Value)
					if err != nil {
						return err
					}
					fmt.Fprintf(out, " %s=%q", a.Name, value)
				}
			}
		}
		if !isEmptyElement(n.Name) {
			if n.Name != "" {
				out.Write(GREATER_THAN)
			}
			cr := true
			if n.Expr != "" {
				expr, err := v.Compile(n.Expr)
				if err != nil {
					return err
				}
				fe, ok := expr.(*vm.ForExpr)
				if ok {
					rhs, err := v.Eval(fe.Rhs)
					if err != nil {
						return err
					}
					ra := reflect.ValueOf(rhs)
					typ := ra.Type().Kind()
					switch typ {
					case reflect.Array, reflect.Slice, reflect.Chan:
					default:
						println(typ)
						return errors.New("can't iterate: " + n.Expr)
					}
					if n.Name != "" {
						out.Write(NEW_LINE)
					}
					if typ == reflect.Chan {
						i := 0
						for {
							rr, ok := ra.Recv()
							if !ok {
								break
							}
							x := rr.Interface()
							i++
							if fe.Lhs2 != "" {
								v.Set(fe.Lhs1, i)
								v.Set(fe.Lhs2, x)
							} else {
								v.Set(fe.Lhs1, x)
							}
							for _, c := range n.Children {
								if err := printNode(out, v, c, indent); err != nil {
									return err
								}
							}
						}
					} else {
						l := ra.Len()
						for i := 0; i < l; i++ {
							x := ra.Index(i).Interface()
							if fe.Lhs2 != "" {
								v.Set(fe.Lhs1, i)
								v.Set(fe.Lhs2, x)
							} else {
								v.Set(fe.Lhs1, x)
							}
							for _, c := range n.Children {
								if err := printNode(out, v, c, indent); err != nil {
									return err
								}
							}
						}
					}
				} else {
					r, err := v.Eval(expr)
					if err != nil {
						return err
					}
					fmt.Fprint(out, r)
					cr = false
				}
				text, err := rubyInline(v, n.Text)
				if err != nil {
					return err
				}
				out.Write([]byte(text))
			} else if len(n.Children) > 0 {
				out.Write(NEW_LINE)
				for _, c := range n.Children {
					if err := printNode(out, v, c, indent+1); err != nil {
						return err
					}
				}
				text, err := rubyInline(v, n.Text)
				if err != nil {
					return err
				}
				out.Write([]byte(text))
			} else if n.Text != "" {
				text, err := rubyInline(v, n.Text)
				if err != nil {
					return err
				}
				out.Write([]byte(text))
				cr = false
			} else if cr {
				out.Write(NEW_LINE)
			}
			if n.Name != "" {
				if cr {
					bytesRepeat(out, SPACE, indent*2)
				}
				out.Write(LESS_THAN_SLASH)
				out.Write([]byte(n.Name))
				out.Write(GREATER_THAN_NEW_LINE)
			}
		} else if doctype {
			out.Write(GREATER_THAN_NEW_LINE)
		} else {
			out.Write(SLASH_GREATER_THAN_NEW_LINE)
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
		aname := ""
		avalue := ""
		for n := 0; n < len(rs); n++ {
			eol := n == len(rs)-1
			r := rs[n]
		break_st:
			switch st {
			case sNeutral:
				if unicode.IsSpace(r) {
					break
				}
				st = sTag
				tag += string(r)

				if n > last {
					last = n
					if node.Name != "" && node.Name[len(node.Name)-1] == ':' {
						node.Text = tag
						st = sText
						break break_st
					}
					node = node.NewChild()
					stk = append(stk, stack{n: n, node: node})
				} else if n == last {
					last = n
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
					if cur.Name != "" && cur.Name[len(cur.Name)-1] == ':' {
						node.Text = tag
						st = sText
						break break_st
					}
					node = cur.NewChild()
					stk[len(stk)-1].node = node
				} else if n < last {
					last = n
					node = nil
					for i := 0; i < len(stk); i++ {
						if stk[i].n >= n {
							node = stk[i-1].node
							stk = stk[:i]
							break
						}
					}
					if node == nil {
						node = root.NewChild()
						stk = stk[:1]
					} else {
						if node.Name != "" && node.Name[len(node.Name)-1] == ':' {
							node.Text = tag
							st = sText
							break break_st
						}
						node = node.NewChild()
					}
				}
				switch r {
				case '=':
					node.Name = "div"
					st = sExpr
					break break_st
				case '|':
					st = sText
					break break_st
				case '-':
					st = sExpr
					break break_st
				case '#':
					node.Name = "div"
					st = sId
					break break_st
				case '.':
					node.Name = "div"
					st = sClass
					break break_st
				}
				if r > 255 {
					node.Text += string(r)
					st = sText
					break break_st
				}

				node.Name = tag
			case sTag:
				if eol {
					tag += string(r)
					node.Name = tag
					break
				}
				switch r {
				case '=':
					if tag == "" {
						node.Name = "div"
					} else {
						node.Name = tag
					}
					st = sExpr
				case '#':
					if tag == "" {
						node.Name = "div"
					} else {
						node.Name = tag
					}
					st = sId
				case '.':
					if tag == "" {
						node.Name = "div"
					} else {
						node.Name = tag
					}
					st = sClass
				default:
					if tag == "" && unicode.IsLetter(r) {
						node.Text = string(r)
						st = sText
						break break_st
					}
					if unicode.IsSpace(r) {
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
					aname += string(r)
					if avalue != "" {
						node.Attr = append(node.Attr, Attr{Name: strings.TrimSpace(aname), Value: ""})
					} else {
						node.Text = strings.TrimSpace(aname)
					}
					break
				}
				switch r {
				case '=':
					if aname == "" {
						st = sExpr
					} else {
						st = sAttrValue
					}
				default:
					aname += string(r)
				}
			case sAttrValue:
				if eol {
					if unicode.IsLetter(r) || r == '"' {
						avalue += string(r)
						if avalue[0] == '"' && avalue[len(avalue)-1] == '"' {
							avalue = avalue[1 : len(avalue)-1]
						}
						node.Attr = append(node.Attr, Attr{Name: aname, Value: strings.TrimSpace(avalue)})
					}
					break
				}
				if avalue != "" && unicode.IsSpace(r) {
					if avalue[0] == '"' {
						if avalue[len(avalue)-1] == '"' {
							avalue = avalue[1 : len(avalue)-1]
						} else {
							avalue += string(r)
							break
						}
					}
					node.Attr = append(node.Attr, Attr{Name: aname, Value: strings.TrimSpace(avalue)})
					aname = ""
					avalue = ""
					st = sAttrKey
				} else {
					avalue += string(r)
				}
			case sEq:
				if r != '=' && !unicode.IsSpace(r) {
					node.Expr += string(r)
					st = sExpr
				}
			case sExpr:
				node.Expr += string(r)
			case sText:
				if node.Text != "" || !unicode.IsSpace(r) {
					node.Text += string(r)
				}
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
