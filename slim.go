package slim

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
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
	sID
	sClass
	sAttrKey
	sAttrValue
	sEq
	sText
	sComment
	sExpr
	sWrapAttr
	sWrapValue
)

var (
	cSpace                   = []byte(" ")
	cNewLine                 = []byte("\n")
	cLessThanSlash           = []byte("</")
	cGreaterThanNewLine      = []byte(">\n")
	cSlashGreaterThanNewLine = []byte("/>\n")
	cDoubleQuote             = []byte("\"")
	cGreaterThan             = []byte(">")
	cLessThan                = []byte("<")
)

var emptyElements = []string{
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

// Value is a type for indicating values for expression.
type Value any

// Values is a collection of Value
type Values map[string]Value

// Func is a type for indicating function for expression.
type Func func(...Value) (Value, error)

// Funcs is a type for indicating function map to pass FuncMap().
type Funcs map[string]Func

// Attr is a type for indiacating attribute of tag.
type Attr struct {
	Name  string
	Value string
}

// Node is a type for indicating tag.
type Node struct {
	Name       string
	ID         string
	Class      []string
	Attr       []Attr
	Text       string
	Expr       string
	Children   []*Node
	Raw        bool
	TrimOuter  bool // > remove outer whitespace
	TrimInner  bool // < remove inner whitespace
	SelfClose  bool // / explicit self-closing tag
	Indent   int
}

// NewChild create child node.
func (n *Node) NewChild() *Node {
	n.Children = append(n.Children, new(Node))
	return n.Children[len(n.Children)-1]
}

type stack struct {
	n    int
	node *Node
}

func isEmptyElement(n string) bool {
	for _, s := range emptyElements {
		if s == n {
			return true
		}
	}
	return false
}

var rubyInlinePattern = regexp.MustCompile(`#\{\{[^}]*\}\}|#\{[^}]*\}`)
var jsVarPattern = regexp.MustCompile(`{{[a-zA-Z$_]+[a-zA-Z0-9$_]*}}`)

// rubyInline expands #{expr} (HTML-escaped) and #{{expr}} (raw) in s.
func rubyInline(v *vm.VM, s string) (string, error) {
	var fail error
	text := rubyInlinePattern.ReplaceAllStringFunc(s, func(s string) string {
		raw := strings.HasPrefix(s, "#{{")
		src := s[2 : len(s)-1]
		if raw {
			src = s[3 : len(s)-2]
		}
		expr, err := v.Compile(src)
		if err != nil {
			fail = err
			return ""
		}
		iv, err := v.Eval(expr)
		if err != nil {
			fail = err
			return ""
		}
		if raw {
			return fmt.Sprint(iv)
		}
		return html.EscapeString(fmt.Sprint(iv))
	})
	if fail != nil {
		return "", fail
	}
	return text, nil
}

// byteRepeat same as bytes.Repeat but Write to the io.Writer
func bytesRepeat(out io.Writer, b []byte, count int) {
	for i := 0; i < count; i++ {
		out.Write(b)
	}
}

func isIfExpr(v *vm.VM, n *Node) (*vm.IfExpr, bool) {
	if n.Expr == "" {
		return nil, false
	}
	expr, err := v.Compile(n.Expr)
	if err != nil {
		return nil, false
	}
	ie, ok := expr.(*vm.IfExpr)
	return ie, ok
}

func collectIfChain(n *Node, v *vm.VM) (branches []struct {
	ie   *vm.IfExpr
	body []*Node
}) {
	ie, ok := isIfExpr(v, n)
	if !ok {
		return nil
	}
	var body []*Node
	var chain []*Node
	for _, c := range n.Children {
		if _, ok := isIfExpr(v, c); ok {
			chain = append(chain, c)
		} else {
			body = append(body, c)
		}
	}
	branches = append(branches, struct {
		ie   *vm.IfExpr
		body []*Node
	}{ie, body})
	for _, c := range chain {
		branches = append(branches, collectIfChain(c, v)...)
	}
	return
}

func evalIfChain(t *Template, out io.Writer, v *vm.VM, n *Node, indent int) error {
	branches := collectIfChain(n, v)
	for _, br := range branches {
		// else: no condition
		if br.ie.Cond == nil {
			return printChildren(t, out, v, br.body, indent)
		}
		result, err := v.Eval(br.ie.Cond)
		if err != nil {
			return err
		}
		cond := vm.IsTrue(result)
		if br.ie.Neg {
			cond = !cond
		}
		if cond {
			return printChildren(t, out, v, br.body, indent)
		}
	}
	return nil
}

func printChildren(t *Template, out io.Writer, v *vm.VM, children []*Node, indent int) error {
	for _, c := range children {
		if _, ok := isIfExpr(v, c); ok {
			if err := evalIfChain(t, out, v, c, indent); err != nil {
				return err
			}
			continue
		}
		if err := printNode(t, out, v, c, indent); err != nil {
			return err
		}
	}
	return nil
}

func printNode(t *Template, out io.Writer, v *vm.VM, n *Node, indent int) error {
	// Handle yield before any tag processing
	if n.Expr != "" {
		expr, err := v.Compile(n.Expr)
		if err == nil {
			if ye, ok := expr.(*vm.YieldExpr); ok {
				block, ok := t.blocks[ye.Name]
				if !ok {
					if ye.Name == "" {
						return nil
					}
					return fmt.Errorf("undefined block: %s", ye.Name)
				}
				for _, c := range block.Children {
					if err := printNode(t, out, v, c, indent); err != nil {
						return err
					}
				}
				return nil
			}
		}
	}
	if n.Name == "" && n.Expr == "" {
		return printChildren(t, out, v, n.Children, indent)
	} else if n.Name == "/" {
		return nil
	} else if n.Name == "/!" {
		bytesRepeat(out, cSpace, indent*2)
		out.Write([]byte("<!-- "))
		out.Write([]byte(n.Text))
		out.Write([]byte(" -->\n"))
	} else {
		// FIXME
		doctype := n.Name == "doctype"
		if doctype {
			bytesRepeat(out, cSpace, indent*2)
			out.Write([]byte("<!"))
			out.Write([]byte(n.Name))
			out.Write([]byte(" html"))
		} else if n.Name != "" {
			if !n.TrimOuter {
				bytesRepeat(out, cSpace, indent*2)
			}
			if strings.HasSuffix(n.Name, ":") {
				name := n.Name[:len(n.Name)-1]
				if en, ok := t.renderer[name]; ok {
					return en(out, n, v)
				}
				out.Write(cLessThan)
				out.Write([]byte(name))
			} else {
				out.Write(cLessThan)
				out.Write([]byte(n.Name))
			}
		}
		if n.ID != "" {
			out.Write([]byte(" id=\""))
			out.Write([]byte(n.ID))
			out.Write(cDoubleQuote)
		}
		if len(n.Class) > 0 {
			out.Write([]byte(" class=\""))
			for i, c := range n.Class {
				if i > 0 {
					out.Write(cSpace)
				}
				out.Write([]byte(c))
			}
			out.Write(cDoubleQuote)
		}
		if len(n.Attr) > 0 && !doctype {
			for _, a := range n.Attr {
				if a.Value == "" {
					out.Write(cSpace)
					out.Write([]byte(a.Name))
				} else {
					value, err := rubyInline(v, a.Value)
					if err != nil {
						return err
					}
					fmt.Fprintf(out, " %s=\"%s\"", a.Name, strings.ReplaceAll(value, `"`, "&#34;"))
				}
			}
		}
		if !isEmptyElement(n.Name) && !n.SelfClose {
			if n.Name != "" {
				out.Write(cGreaterThan)
			}
			cr := true
			if n.Expr != "" {
				expr, err := v.Compile(n.Expr)
				if err != nil {
					return err
				}
				fe, ok := expr.(*vm.ForExpr)
				if ok {
					rhs, err := v.Eval(fe.RHS)
					if err != nil {
						return err
					}
					ra := reflect.ValueOf(rhs)
					typ := ra.Type().Kind()
					switch typ {
					case reflect.Array, reflect.Slice, reflect.Chan:
					default:
						return errors.New("can't iterate: " + n.Expr)
					}
					if n.Name != "" {
						out.Write(cNewLine)
					}
					if typ == reflect.Chan {
						i := 0
						for {
							rr, ok := ra.Recv()
							if !ok {
								break
							}
							x := rr.Interface()
							if fe.LHS2 != "" {
								v.Set(fe.LHS1, i)
								v.Set(fe.LHS2, x)
							} else {
								v.Set(fe.LHS1, x)
							}
							if err := printChildren(t, out, v, n.Children, indent); err != nil {
								return err
							}
							i++
						}
					} else {
						l := ra.Len()
						for i := 0; i < l; i++ {
							x := ra.Index(i).Interface()
							if fe.LHS2 != "" {
								v.Set(fe.LHS1, i)
								v.Set(fe.LHS2, x)
							} else {
								v.Set(fe.LHS1, x)
							}
							if err := printChildren(t, out, v, n.Children, indent); err != nil {
								return err
							}
						}
					}
				} else {
					r, err := v.Eval(expr)
					if err != nil {
						return err
					}
					if r != nil {
						text := fmt.Sprint(r)
						if !n.Raw {
							text = html.EscapeString(text)
						}
						out.Write([]byte(text))
					}
					cr = false
				}
				text, err := rubyInline(v, n.Text)
				if err != nil {
					return err
				}
				out.Write([]byte(text))
			} else if len(n.Children) > 0 {
				if !n.TrimInner {
					out.Write(cNewLine)
				}
				if err := printChildren(t, out, v, n.Children, indent+1); err != nil {
					return err
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
				out.Write(cNewLine)
			}
			if n.Name != "" {
				name := n.Name
				if strings.HasSuffix(n.Name, ":") {
					name = n.Name[:len(n.Name)-1]
				}

				if cr && !n.TrimInner {
					bytesRepeat(out, cSpace, indent*2)
				}
				out.Write(cLessThanSlash)
				out.Write([]byte(name))
				if n.TrimOuter {
					out.Write(cGreaterThan)
				} else {
					out.Write(cGreaterThanNewLine)
				}
			}
		} else if doctype {
			out.Write(cGreaterThanNewLine)
		} else {
			out.Write(cSlashGreaterThanNewLine)
		}
	}
	return nil
}

// Template is the representation of a parsed template.
type Template struct {
	root     *Node
	renderer map[string]Renderer
	inner    map[string]*Template
	blocks   map[string]*Node
	fm       Funcs
	dir      string
	fsys     fs.FS
	layout   string
}

// ParseFile parse content of fname.
func ParseFile(fname string) (*Template, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}

// ParseFS parse content of name from fsys.
func ParseFS(fsys fs.FS, name string) (*Template, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	t, err := Parse(f)
	if err != nil {
		return nil, err
	}
	t.fsys = fsys
	t.dir = path.Dir(name)
	return t, nil
}

// Renderer is a type for indicating custom function for renderer.
type Renderer func(out io.Writer, n *Node, v *vm.VM) error

var defaultRenderers = map[string]Renderer{
	"javascript": javascriptRenderer,
	"css":        cssRenderer,
}

var contentForPattern = regexp.MustCompile(`^(\s*)-\s+content_for\s+:?(\w+)\s*$`)

func splitDefines(in io.Reader) (string, map[string]string) {
	scanner := bufio.NewScanner(in)
	blocks := map[string]string{}
	var main []string
	var defName string
	defIndent := -1
	for scanner.Scan() {
		line := scanner.Text()
		if m := contentForPattern.FindStringSubmatch(line); m != nil {
			defName = m[2]
			defIndent = len(m[1])
			blocks[defName] = ""
			continue
		}
		if defName != "" {
			// Lines indented deeper than the define belong to the block
			trimmed := strings.TrimLeft(line, " \t")
			indent := len(line) - len(trimmed)
			if trimmed == "" || indent > defIndent {
				blocks[defName] += line[defIndent:] + "\n"
				continue
			}
			defName = ""
			defIndent = -1
		}
		main = append(main, line)
	}
	return strings.Join(main, "\n"), blocks
}

// Parse parse content with reading from reader.
func Parse(in io.Reader) (*Template, error) {
	if in == nil {
		return nil, errors.New("invalid input")
	}
	body, defineBlocks := splitDefines(in)
	scanner := bufio.NewScanner(strings.NewReader(body))
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
					if strings.HasSuffix(node.Name, ":") {
						node.Text += "\n" + strings.Repeat(" ", n) + tag
						st = sText
						break break_st
					}
					node = node.NewChild()
					stk = append(stk, stack{n: n, node: node})
				} else if n == last {
					last = n
					if strings.HasSuffix(node.Name, ":") {
						node.Text += "\n" + strings.Repeat(" ", n) + tag
						st = sText
						break break_st
					}
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
					stk[len(stk)-1].node = node
				} else if n < last {
					last = n
					found := (*Node)(nil)
					for i := 0; i < len(stk); i++ {
						if i > 0 && stk[i].n >= n {
							found = stk[i-1].node
							stk = stk[:i]
							break
						}
					}
					if found == nil && strings.HasSuffix(node.Name, ":") {
						node.Text += "\n" + strings.Repeat(" ", n) + tag
						st = sText
						break break_st
					}
					node = found
					if node == nil {
						node = root.NewChild()
						stk = stk[:1]
					} else {
						node = node.NewChild()
						stk = append(stk, stack{n: n, node: node})
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
					st = sID
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
				if r == '/' {
					node.Name = tag
					node.SelfClose = true
					break
				}
				if r == '>' {
					if tag != "" {
						node.Name = tag
					}
					node.TrimOuter = true
					break
				}
				if r == '<' {
					if tag != "" {
						node.Name = tag
					}
					node.TrimInner = true
					break
				}
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
					st = sID
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
					if r == '(' {
						node.Name = tag
						st = sWrapAttr
					} else if unicode.IsSpace(r) {
						node.Name = tag
						st = sAttrKey
					} else {
						tag += string(r)
					}
				}
			case sID:
				if eol {
					if isUnquotedAttributeValue(r) { // FIXME
						id += string(r)
						node.ID = id
					}
					break
				}
				switch r {
				case '.':
					node.ID = id
					st = sClass
				case '(':
					node.ID = id
					st = sWrapAttr
				default:
					if !isUnquotedAttributeValue(r) { // FIXME
						node.ID = id
						if unicode.IsSpace(r) {
							st = sAttrKey
						} else {
							st = sEq
						}
					} else {
						id += string(r)
					}
				}
			case sClass:
				if eol {
					if isUnquotedAttributeValue(r) { // FIXME
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
				case '(':
					if class != "" {
						node.Class = append(node.Class, class)
					}
					st = sWrapAttr
				default:
					if !isUnquotedAttributeValue(r) { // FIXME
						if class != "" {
							node.Class = append(node.Class, class)
						}
						if unicode.IsSpace(r) {
							st = sAttrKey
						} else {
							st = sEq
						}
					} else {
						class += string(r)
					}
				}
			case sAttrKey:
				if eol {
					aname += string(r)
					if avalue == "" {
						node.Text = strings.TrimSpace(aname)
					} else {
						node.Attr = append(node.Attr, Attr{Name: strings.TrimSpace(aname), Value: ""})
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
					if isUnquotedAttributeValue(r) || r == '"' { // FIXME
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
				}
				st = sExpr
			case sExpr:
				if node.Expr == "" && r == '=' {
					node.Raw = true
				} else {
					node.Expr += string(r)
				}
			case sText:
				if node.Text != "" || !unicode.IsSpace(r) {
					node.Text += string(r)
				}
			case sWrapAttr:
				if r == ')' {
					if strings.TrimSpace(aname) != "" {
						node.Attr = append(node.Attr, Attr{Name: strings.TrimSpace(aname), Value: ""})
						aname = ""
					}
					st = sAttrKey
					break
				}
				if unicode.IsSpace(r) {
					break
				}
				if r == '=' {
					st = sWrapValue
				} else {
					aname += string(r)
				}
			case sWrapValue:
				if r == ')' {
					if avalue != "" {
						if len(avalue) >= 2 && avalue[0] == '"' && avalue[len(avalue)-1] == '"' {
							avalue = avalue[1 : len(avalue)-1]
						}
						node.Attr = append(node.Attr, Attr{Name: strings.TrimSpace(aname), Value: avalue})
					}
					aname = ""
					avalue = ""
					st = sAttrKey
					break
				}
				if avalue != "" && unicode.IsSpace(r) && (avalue[0] != '"' || avalue[len(avalue)-1] == '"') {
					if len(avalue) >= 2 && avalue[0] == '"' && avalue[len(avalue)-1] == '"' {
						avalue = avalue[1 : len(avalue)-1]
					}
					node.Attr = append(node.Attr, Attr{Name: strings.TrimSpace(aname), Value: avalue})
					aname = ""
					avalue = ""
					st = sWrapAttr
				} else {
					avalue += string(r)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	newrenderer := make(map[string]Renderer)
	for n, k := range defaultRenderers {
		newrenderer[n] = k
	}

	dir, _ := os.Getwd()
	if ff, ok := in.(*os.File); ok {
		dir, _ = filepath.Abs(filepath.Dir(ff.Name()))
	}
	blocks := map[string]*Node{}
	for name, body := range defineBlocks {
		bt, err := Parse(strings.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("content_for %s: %w", name, err)
		}
		blocks[name] = bt.root
	}
	// Store the main body (non-content_for parts) as the unnamed block
	if len(root.Children) > 0 {
		blocks[""] = root
	}

	return &Template{
		root:     root,
		renderer: newrenderer,
		inner:    map[string]*Template{},
		blocks:   blocks,
		fm:       nil,
		dir:      dir,
	}, nil
}

// FuncMap set the template's function map.
func (t *Template) FuncMap(m Funcs) {
	t.fm = m
}

// SetLayout set the layout template name.
func (t *Template) SetLayout(name string) {
	t.layout = name
}

// RegisterRenderer register custom render named with the name.
func (t *Template) RegisterRenderer(name string, r Renderer) {
	t.renderer[name] = r
}

// Execute applies a parsed template to the specified value object,
// and writes the output to out.
func (t *Template) execute(v *vm.VM, out io.Writer, value any) error {
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
	return printNode(t, out, v, t.root, 0)
}

// Execute applies a parsed template to the specified value object,
// and writes the output to out.
func (t *Template) Execute(out io.Writer, value any) error {
	v := vm.New()

	v.Set("render", func(name string) error {
		var tt *Template
		if t.fsys != nil {
			name = path.Join(t.dir, name)
			if cached, ok := t.inner[name]; ok {
				tt = cached
			} else {
				var err error
				tt, err = ParseFS(t.fsys, name)
				if err != nil {
					return err
				}
				t.inner[name] = tt
			}
		} else {
			if !filepath.IsAbs(name) {
				name = filepath.Join(t.dir, name)
			}
			if cached, ok := t.inner[name]; ok {
				tt = cached
			} else {
				var err error
				tt, err = ParseFile(name)
				if err != nil {
					return err
				}
				t.inner[name] = tt
			}
			tt.dir = filepath.Dir(name)
		}
		return tt.execute(v, out, value)
	})

	if t.layout != "" {
		var lt *Template
		var err error
		if t.fsys != nil {
			name := path.Join(t.dir, t.layout)
			lt, err = ParseFS(t.fsys, name)
		} else {
			name := t.layout
			if !filepath.IsAbs(name) {
				name = filepath.Join(t.dir, name)
			}
			lt, err = ParseFile(name)
		}
		if err != nil {
			return err
		}
		for k, b := range t.blocks {
			lt.blocks[k] = b
		}
		lt.fm = t.fm
		return lt.execute(v, out, value)
	}

	return t.execute(v, out, value)
}

func javascriptRenderer(out io.Writer, n *Node, v *vm.VM) error {
	re := jsVarPattern
	var err error
	s := re.ReplaceAllStringFunc(n.Text, func(s string) string {
		if err == nil {
			vv, ok := v.Get(s[2 : len(s)-2])
			if !ok {
				err = fmt.Errorf("invalid variable name: %v", s)
				return ""
			}
			var buf bytes.Buffer
			err = json.NewEncoder(&buf).Encode(vv)
			if err != nil {
				return ""
			}
			return strings.TrimSpace(buf.String())
		}
		return ""
	})
	if err != nil {
		return err
	}
	indent := 0
	for _, r := range s {
		if !unicode.IsSpace(r) {
			break
		}
		indent++
	}
	if indent > 2 {
		indent -= 2
	}
	_, err = fmt.Fprintf(out, "<script>%s%s</script>\n", s, s[:indent])
	return err
}

func cssRenderer(out io.Writer, n *Node, v *vm.VM) error {
	s := n.Text
	indent := 0
	for _, r := range s {
		if !unicode.IsSpace(r) {
			break
		}
		indent++
	}
	if indent > 2 {
		indent -= 2
	}
	_, err := fmt.Fprintf(out, "<style type=\"text/css\">%s%s</style>\n", s, s[:indent])
	return err
}

// ..in addition to the requirements given above for attribute values, must not
//   contain any literal ASCII whitespace, any U+0022 QUOTATION MARK characters ("),
//   U+0027 APOSTROPHE characters ('), U+003D EQUALS SIGN characters (=),
//   U+003C LESS-THAN SIGN characters (<), U+003E GREATER-THAN SIGN characters (>),
//   or U+0060 GRAVE ACCENT characters (`), and must not be the empty string.
func isUnquotedAttributeValue(r rune) bool {
	return !(unicode.IsSpace(r) ||
		r == '"' || r == '\'' || r == '=' ||
		r == '<' || r == '>' || r == '`')
}
