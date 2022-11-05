package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestNilInput(t *testing.T) {
	err := run(nil, nil, []string{"foo=foo", "foo=bar", "foo=baz"})
	if err == nil {
		t.Fatalf("should be fail")
	}
}

func TestFileInput(t *testing.T) {
	want, err := ioutil.ReadFile("../../testdata/test_each.html")
	if err != nil {
		t.Fatal(err)
	}
	input, err := ioutil.ReadFile("../../testdata/test_each.slim")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = run(&buf, bytes.NewReader(input), []string{"foo=foo", "foo=bar", "foo=baz"})
	if err != nil {
		t.Fatalf("fatal: %v", err)
	}
	got := buf.Bytes()

	if bytes.Compare(got, want) != 0 {
		t.Errorf("want %v, but %v", string(want), string(got))
	}
}
