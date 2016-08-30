package slim

import (
	"bytes"
	"testing"
)

func TestSimple(t *testing.T) {
	tmpl, err := ParseFile("testdir/test.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		t.Fatal(err)
	}
	expect := `<!doctype>
<html lang=ja>
  <head>
    <meta charset=UTF-8/>
    <title>
    </title>
  </head>
  <body>
    <p>Hello</p>
  </body>
</html>
`
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}
