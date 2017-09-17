package vm

import (
	"errors"
	"testing"
)

type testStruct1 struct {
	Foo int
}

func (v *testStruct1) SomeFunction(a int, b int) int {
	return v.Foo + a + b
}

func (v *testStruct1) SomeFunctionWithError(a int) (int, error) {
	if a == 0 {
		return 0, errors.New("zero")
	}
	return v.Foo / a, nil
}

func (v *testStruct1) Itself() *testStruct1 {
	return v
}

func TestInt(t *testing.T) {
	v := New()
	v.Set("foo", 1)
	expr, err := v.Compile(`foo`)
	if err != nil {
		t.Fatal(err)
	}
	r, err := v.Eval(expr)
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
	expr, err := v.Compile(`"foo"`)
	if err != nil {
		t.Fatal(err)
	}
	r, err := v.Eval(expr)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := r.(string)
	if !ok || s != "foo" {
		t.Fatalf("Expected %v, but %v:", 1, r)
	}
}

func TestMethodCall(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		v := New()
		v.Set("test", testStruct1{
			Foo: 133,
		})
		v.Set("x", 12)
		v.Set("y", 3)
		expr, err := v.Compile(`test.SomeFunction(x, y)`)
		if err != nil {
			t.Fatal(err)
		}
		r, err := v.Eval(expr)
		if err != nil {
			t.Fatal(err)
		}
		s, ok := r.(int)
		if !ok || s != 148 {
			t.Fatalf("Expected %v, but %v:", 148, r)
		}
	})
	t.Run("calling undefined method", func(t *testing.T) {
		v := New()
		v.Set("test", testStruct1{
			Foo: 133,
		})
		v.Set("x", 12)
		v.Set("y", 3)
		expr, err := v.Compile(`test.SomeUndefinedFunction(x, y)`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = v.Eval(expr)
		if err == nil {
			t.Fatalf("Expected to error, but not")
		}
	})
	t.Run("chained", func(t *testing.T) {
		v := New()
		v.Set("test", testStruct1{
			Foo: 133,
		})
		v.Set("x", 12)
		v.Set("y", 3)
		expr, err := v.Compile(`test.Itself().SomeFunction(x, y)`)
		if err != nil {
			t.Fatal(err)
		}
		r, err := v.Eval(expr)
		if err != nil {
			t.Fatal(err)
		}
		s, ok := r.(int)
		if !ok || s != 148 {
			t.Fatalf("Expected %v, but %v:", 148, r)
		}
	})
}

func TestMethodCallMayReturnError(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		v := New()
		v.Set("test", testStruct1{
			Foo: 12,
		})
		v.Set("x", 2)
		expr, err := v.Compile(`test.SomeFunctionWithError(x)`)
		if err != nil {
			t.Fatal(err)
		}
		r, err := v.Eval(expr)
		if err != nil {
			t.Fatal(err)
		}
		s, ok := r.(int)
		if !ok || s != 6 {
			t.Fatalf("Expected %v, but %v:", 6, r)
		}
	})
	t.Run("with error", func(t *testing.T) {
		v := New()
		v.Set("test", testStruct1{
			Foo: 10,
		})
		v.Set("x", 0)
		expr, err := v.Compile(`test.SomeFunctionWithError(x)`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = v.Eval(expr)
		if err == nil {
			t.Fatalf("Expected to error, but not")
		}
	})
}
