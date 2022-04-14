# go-slim

[![Build Status](https://github.com/mattn/go-slim/workflows/test/badge.svg?branch=master)](https://github.com/mattn/go-slim/actions?query=workflow%3Atest)
[![Codecov](https://codecov.io/gh/mattn/go-slim/branch/master/graph/badge.svg)](https://codecov.io/gh/mattn/go-slim)
[![Go Reference](https://pkg.go.dev/badge/github.com/mattn/go-slim.svg)](https://pkg.go.dev/github.com/mattn/go-slim)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattn/go-slim)](https://goreportcard.com/report/github.com/mattn/go-slim)

slim template engine for golang

## Features

* Small Virtual Machine

  Enough to manipulate object in template. Support Number/String/Function/Array/Map.

* Ruby Like Text Rendering

  Support `Hello #{"Golang"}`

## Usage

### Template File

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

### Your Code

```go
tmpl, err := slim.ParseFile("template.slim")
if err != nil {
	t.Fatal(err)
}
err = tmpl.Execute(os.Stdout, slim.Values{
	"foo": []string{"foo", "bar", "baz"},
})
```

### Output

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

## Builtin-Functions

* trim(s)
* to_upper(s)
* to_lower(s)
* repeat(s, n)

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
