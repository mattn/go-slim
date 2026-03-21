# go-slim

[![Build Status](https://github.com/mattn/go-slim/workflows/test/badge.svg?branch=master)](https://github.com/mattn/go-slim/actions?query=workflow%3Atest)
[![Codecov](https://codecov.io/gh/mattn/go-slim/branch/master/graph/badge.svg)](https://codecov.io/gh/mattn/go-slim)
[![Go Reference](https://pkg.go.dev/badge/github.com/mattn/go-slim.svg)](https://pkg.go.dev/github.com/mattn/go-slim)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattn/go-slim)](https://goreportcard.com/report/github.com/mattn/go-slim)

Slim template engine for Go.

## Features

* **Slim-compatible syntax** — indentation-based HTML templating
* **Expressions** — arithmetic, comparison (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical (`&&`, `||`, `!`)
* **Conditionals** — `if` / `elsif` / `else` / `unless`
* **Iteration** — `for` loops over slices, arrays, and channels
* **Layouts** — `content_for` / `yield` with `SetLayout`, matching Ruby Slim conventions
* **Embedded filesystem** — `ParseFS` for use with `embed.FS`
* **String interpolation** — `Hello #{name}`
* **Custom renderers** — built-in `javascript:` and `css:` blocks, extensible via `RegisterRenderer`
* **Whitespace control** — `>` (outer) and `<` (inner)
* **Wrapped attributes** — `div(id="x" class="y")`
* **Self-closing tags** — `div/`

## Usage

### Basic

```slim
doctype 5
html lang="ja"
  head
    meta charset="UTF-8"
    title
  body
    ul
      - for x in foo
        li = x
```

```go
tmpl, err := slim.ParseFile("template.slim")
if err != nil {
    log.Fatal(err)
}
err = tmpl.Execute(os.Stdout, slim.Values{
    "foo": []string{"foo", "bar", "baz"},
})
```

Output:

```html
<!doctype html>
<html lang="ja">
  <head>
    <meta charset="UTF-8"/>
    <title>
    </title>
  </head>
  <body>
    <ul>
      <li>foo</li>
      <li>bar</li>
      <li>baz</li>
    </ul>
  </body>
</html>
```

### Conditionals

```slim
- if user.Role == "admin"
  p Welcome, admin!
- elsif user.Role == "mod"
  p Welcome, moderator!
- else
  p Hello!

- unless hidden
  p This is visible
```

### Layouts

Define blocks in pages with `content_for`, render them in layouts with `yield`.

**layout.slim:**

```slim
doctype 5
html
  head
    title My App
  body
    = yield
    = yield sidebar
```

**index.slim:**

```slim
h1 Hello
p Main content

- content_for sidebar
  nav
    a href="/" Home
```

**Go code:**

```go
tmpl, _ := slim.ParseFile("index.slim")
tmpl.SetLayout("layout.slim")
tmpl.Execute(w, values)
```

### Embedded Filesystem

```go
//go:embed templates/*
var templates embed.FS

tmpl, _ := slim.ParseFS(templates, "templates/index.slim")
tmpl.SetLayout("templates/layout.slim")
tmpl.Execute(w, values)
```

### Tags

```slim
/ Code comment (not rendered)
/! HTML comment

div#main.container
div(id="main" class="container")

img/
custom-element/

p> Removes outer whitespace
p< Removes inner whitespace

| Verbatim text
= expression (HTML escaped, wrapped in div)
== expression (raw output)
- code (control flow, no output)
```

### Expressions

```slim
= name
= user.Name
= items[0]
= score + 10
= price * 1.1
= count > 0 && active
= !hidden
= greet("world")
= item.Method(arg)
```

### String Interpolation

```slim
p Hello #{name}
a href="#{base_url}/path" Link
```

### Custom Renderers

```slim
javascript:
  var x = 1;

css:
  body { color: red; }
```

Register custom renderers:

```go
tmpl.RegisterRenderer("my-lang", func(out io.Writer, n *slim.Node, v *vm.VM) error {
    fmt.Fprint(out, n.Text)
    return nil
})
```

## Built-in Functions

| Function | Description |
|----------|-------------|
| `trim(s)` | Trim whitespace |
| `to_upper(s)` | Convert to uppercase |
| `to_lower(s)` | Convert to lowercase |
| `repeat(s, n)` | Repeat string n times |

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
