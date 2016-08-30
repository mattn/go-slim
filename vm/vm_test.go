package vm

import (
	"testing"
)

func TestInt(t *testing.T) {
	v := New()
	v.Set("foo", 1)
	r, err := v.Run("foo")
	if err != nil {
		t.Fatal(err)
	}
	i, ok := r.(int)
	if !ok || i != 1 {
		t.Fatalf("Expected %v, but %v:", 1, r)
	}
}

func TestString(t *testing.T) {
	v := New()
	v.Set("foo", 2)
	r, err := v.Run(`"foo"`)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := r.(string)
	if !ok || s != "foo" {
		t.Fatalf("Expected %v, but %v:", 1, r)
	}
}
