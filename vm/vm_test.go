package vm

import (
	"testing"
)

func TestSimple(t *testing.T) {
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
