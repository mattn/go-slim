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

func BenchmarkTemplate_Execute(b *testing.B) {
	b.StopTimer()

	tmpl, err := ParseFile("testdir/test_simple.slim")
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	// Increase the initial buffer size impacts the number of malloc
	//buf.Grow(2048)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err = tmpl.Execute(&buf, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestValue(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_value.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
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
	err = tmpl.Execute(&buf, Values{
		"bar": "baz",
	})
	if err == nil {
		t.Fatal("should be fail")
	}
}

func TestEachArray(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_each.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
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

func TestEachChan(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_each.slim")
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan string)
	go func() {
		for _, a := range []string{"foo", "bar", "baz"} {
			ch <- a
		}
		close(ch)
	}()

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
		"foo": ch,
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
	tmpl.FuncMap(Funcs{
		"greet": func(args ...Value) (Value, error) {
			return fmt.Sprintf("Hello %v", args[0]), nil
		},
	})
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
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

func TestBuiltins(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_builtins.slim")
	if err != nil {
		t.Fatal(err)
	}
	tmpl.FuncMap(Funcs{
		"trim":     Trim,
		"to_upper": ToUpper,
		"to_lower": ToLower,
		"repeat":   Repeat,
	})
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
		"name": "golang",
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_builtins.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestBuiltinsError(t *testing.T) {
	_, err := Trim("foo", "bar")
	if err == nil {
		t.Fatal("should be fail")
	}
	_, err = ToUpper("foo", "bar")
	if err == nil {
		t.Fatal("should be fail")
	}
	_, err = ToLower("foo", "bar")
	if err == nil {
		t.Fatal("should be fail")
	}
	_, err = Repeat("foo", "bar", "baz")
	if err == nil {
		t.Fatal("should be fail")
	}
}

func TestOp(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_op.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
		"name": "golang",
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_op.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestInline(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_inline.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
		"name": "golang",
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_inline.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestMember(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_member.slim")
	if err != nil {
		t.Fatal(err)
	}

	m := make(map[string]string)
	m["baz"] = "Baz!"

	type Baz struct {
		Fuga string
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
		"foo": struct {
			Baz []Baz
		}{
			Baz: []Baz{
				{Fuga: "hello"},
				{Fuga: "world"},
				{Fuga: "golang"},
			},
		},
		"bar": m,
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_member.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestItem(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_item.slim")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
		"foo": []interface{}{1, struct {
			Bar []int
		}{[]int{2}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_item.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestComment(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_comment.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_comment.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestRaw(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_raw.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
		"foo": "<script>alert(1)</script>",
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_raw.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestNoRaw(t *testing.T) {
	tmpl, err := ParseFile("testdir/test_noraw.slim")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, Values{
		"foo": "<script>alert(1)</script>",
	})
	if err != nil {
		t.Fatal(err)
	}
	expect := readFile("testdir/test_noraw.html")
	got := buf.String()
	if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
	}
}

func TestAttr(t *testing.T) {
    tmpl, err := ParseFile("testdir/test_attr.slim")
    if err != nil {
        t.Fatal(err)
    }
    var buf bytes.Buffer
    err = tmpl.Execute(&buf, nil)
	if err != nil {
		t.Fatal(err)
	}
    expect := readFile("testdir/test_attr.html")
    got := buf.String()
    if expect != got {
		t.Fatalf("expected %v but %v", expect, got)
    }
}
