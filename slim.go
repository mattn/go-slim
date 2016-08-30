package slim

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"unicode"
)

type state int

const (
	sNeutral state = iota
	sTagName
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

type stack struct {
	n    int
	node *Node
}

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

func isEmptyElement(n string) bool {
	for _, e := range emptyElement {
		if e == n {
			return true
		}
	}
	return false
}

func printNode(n *Node, indent int) string {
	s := ""
	if n.Name == "" {
		for _, c := range n.Children {
			s += c.String()
		}
	} else {
		s += strings.Repeat(" ", indent*2) + "<" + n.Name
		if n.Id != "" {
			s += " id=\"" + n.Id + "\""
		}
		if len(n.Class) > 0 {
			s += " class="
			for i, c := range n.Class {
				if i > 0 {
					s += " "
				}
				s += c
			}
		}
		if len(n.Attr) > 0 {
			for i, a := range n.Attr {
				if i > 0 {
					s += " "
				}
				s += " " + a.Name + "=" + a.Value
			}
		}
		if !isEmptyElement(n.Name) {
			s += ">\n"
		} else {
			s += "/>\n"
		}
		if len(n.Children) > 0 {
			for _, c := range n.Children {
				s += printNode(c, indent+1)
			}
		}
		if !isEmptyElement(n.Name) {
			s += strings.Repeat(" ", indent*2) + "</" + n.Name + ">\n"
		}
	}
	return s
}

func (n *Node) String() string {
	return printNode(n, 0)
}

type Template struct {
	Root *Node
}

func ParseFile(name string) (*Template, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	root := new(Node)
	node := root
	stk := []stack{}
	last := 0
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
				if eol {
					last = n
					break
				}
				if !unicode.IsSpace(r) {
					st = sTagName
					tag += string(r)
					if n > last {
						node.Children = append(node.Children, new(Node))
						node = node.Children[len(node.Children)-1]
						stk = append(stk, stack{n: n, node: node})
						last = n
					} else if n <= last {
						node = nil
						for i := len(stk) - 1; i >= 0; i-- {
							if stk[i].n < last {
								last = n
								node = stk[i].node
								stk = stk[:i]
								break
							}
						}
						if node == nil {
							if n == 0 {
								root.Children = append(root.Children, new(Node))
								node = root.Children[len(root.Children)-1]
							} else {
								node = root.Children[len(root.Children)-1]
								node.Children = append(node.Children, new(Node))
								node = node.Children[len(node.Children)-1]
							}
							stk = stk[:0]
							last = 0
						} else {
							node.Children = append(node.Children, new(Node))
							node = node.Children[len(node.Children)-1]
						}
					}
				}
			case sTagName:
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
					if unicode.IsLetter(r) {
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
	return &Template{root}, nil
}

func (t *Template) Execute() {
}
