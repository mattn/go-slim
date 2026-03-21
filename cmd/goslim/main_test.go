package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"testing"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	layoutFile = flag.String("layout", "", "layout template file")
	jsonFile = flag.String("json", "", "JSON file for template values")
}

func TestStdinInput(t *testing.T) {
	resetFlags()
	old := os.Stdin
	defer func() { os.Stdin = old }()

	f, err := os.Open("../../testdata/test_each.slim")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	os.Stdin = f

	os.Args = []string{"goslim", "foo=foo", "foo=bar", "foo=baz"}
	var buf bytes.Buffer
	err = run(&buf, os.Args[1:])
	if err != nil {
		t.Fatal(err)
	}

	want, _ := os.ReadFile("../../testdata/test_each.html")
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("want %q, got %q", want, buf.Bytes())
	}
}

func TestFileInput(t *testing.T) {
	resetFlags()
	os.Args = []string{"goslim", "../../testdata/test_each.slim", "foo=foo", "foo=bar", "foo=baz"}
	var buf bytes.Buffer
	err := run(&buf, os.Args[1:])
	if err != nil {
		t.Fatal(err)
	}

	want, _ := os.ReadFile("../../testdata/test_each.html")
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("want %q, got %q", want, buf.Bytes())
	}
}

func TestJSONInput(t *testing.T) {
	resetFlags()

	// Create temp JSON file
	data := map[string]any{
		"foo": []string{"foo", "bar", "baz"},
	}
	tmp, err := os.CreateTemp("", "goslim-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	json.NewEncoder(tmp).Encode(data)
	tmp.Close()

	os.Args = []string{"goslim", "-json", tmp.Name(), "../../testdata/test_each.slim"}
	var buf bytes.Buffer
	err = run(&buf, os.Args[1:])
	if err != nil {
		t.Fatal(err)
	}

	want, _ := os.ReadFile("../../testdata/test_each.html")
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("want %q, got %q", want, buf.Bytes())
	}
}

func TestLayoutFlag(t *testing.T) {
	resetFlags()
	os.Args = []string{"goslim", "-layout", "test_define_layout.slim", "../../testdata/test_define_render.slim"}
	var buf bytes.Buffer
	err := run(&buf, os.Args[1:])
	if err != nil {
		t.Fatal(err)
	}

	want, _ := os.ReadFile("../../testdata/test_define_render.html")
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("want %q, got %q", want, buf.Bytes())
	}
}
