package main_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	os.MkdirAll("tmp", os.ModeDir|os.ModePerm)
	exe := "tmp/goslim"
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	err := exec.Command("go", "build", "-o", exe).Run()
	if err != nil {
		fmt.Println("Failed to build goslim binary:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestFileInput(t *testing.T) {
	want, err := ioutil.ReadFile("../../testdir/test_each.html")
	if err != nil {
		t.Fatal(err)
	}
	input, err := ioutil.ReadFile("../../testdir/test_each.slim")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("tmp/goslim", "foo=foo", "foo=bar", "foo=baz")
	cmd.Stdin = bytes.NewReader(input)
	got, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fatal: %v: %v", err, string(got))
	}

	if bytes.Compare(got, want) != 0 {
		t.Errorf("want %v, but %v", string(want), string(got))
	}
}
