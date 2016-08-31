package slim

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func readFile(fn string) string {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestSimple(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_simple.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_simple.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestValue(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_value.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"foo": "bar",
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_value.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestUnknownIdentifier(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_value.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"bar": "baz",
	})
	if err == nil {
		t.Fatal("should be fail")
	}
}

func TestEach(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_each.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"foo": []string{"foo", "bar", "baz"},
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_each.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestFunc(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_func.slim")
	if err != nil {
		t.Fatal(err)
	}
	tmpl.FuncMap(map[string]Func{
		"greet": func(v Value) (Value, error) {
			return fmt.Sprintf("Hello %v", v), nil
		},
	})
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"name": "golang",
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_func.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}
