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

func TestFloatLiteral(t *testing.T) {
	v := New()
	expr, err := v.Compile(`3.14`)
	if err != nil {
		t.Fatal(err)
	}
	r, err := v.Eval(expr)
	if err != nil {
		t.Fatal(err)
	}
	f, ok := r.(float64)
	if !ok || f != 3.14 {
		t.Fatalf("Expected %v, but %v", 3.14, r)
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		expr   string
		expect bool
	}{
		{"a == 10", true},
		{"a == 5", false},
		{"a != 5", true},
		{"a != 10", false},
		{"a > 5", true},
		{"a > 10", false},
		{"a < 20", true},
		{"a < 10", false},
		{"a >= 10", true},
		{"a >= 11", false},
		{"a <= 10", true},
		{"a <= 9", false},
	}
	for _, tt := range tests {
		v := New()
		v.Set("a", 10)
		expr, err := v.Compile(tt.expr)
		if err != nil {
			t.Fatalf("compile %q: %v", tt.expr, err)
		}
		r, err := v.Eval(expr)
		if err != nil {
			t.Fatalf("eval %q: %v", tt.expr, err)
		}
		got, ok := r.(bool)
		if !ok || got != tt.expect {
			t.Fatalf("%q: expected %v, got %v", tt.expr, tt.expect, r)
		}
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		expr   string
		expect bool
	}{
		{"a > 5 && a < 20", true},
		{"a > 5 && a < 8", false},
		{"a > 20 || a == 10", true},
		{"a > 20 || a < 5", false},
		{"!flag", true},
		{"!a", false},
	}
	for _, tt := range tests {
		v := New()
		v.Set("a", 10)
		v.Set("flag", false)
		expr, err := v.Compile(tt.expr)
		if err != nil {
			t.Fatalf("compile %q: %v", tt.expr, err)
		}
		r, err := v.Eval(expr)
		if err != nil {
			t.Fatalf("eval %q: %v", tt.expr, err)
		}
		got, ok := r.(bool)
		if !ok || got != tt.expect {
			t.Fatalf("%q: expected %v, got %v", tt.expr, tt.expect, r)
		}
	}
}

func TestBooleanLiterals(t *testing.T) {
	v := New()
	expr, err := v.Compile(`true`)
	if err != nil {
		t.Fatal(err)
	}
	r, err := v.Eval(expr)
	if err != nil {
		t.Fatal(err)
	}
	if r != true {
		t.Fatalf("expected true, got %v", r)
	}
	expr, err = v.Compile(`false`)
	if err != nil {
		t.Fatal(err)
	}
	r, err = v.Eval(expr)
	if err != nil {
		t.Fatal(err)
	}
	if r != false {
		t.Fatalf("expected false, got %v", r)
	}
}

func TestIfExpr(t *testing.T) {
	v := New()
	v.Set("x", 10)
	expr, err := v.Compile(`if x > 5`)
	if err != nil {
		t.Fatal(err)
	}
	ie, ok := expr.(*IfExpr)
	if !ok {
		t.Fatal("expected IfExpr")
	}
	r, err := v.Eval(ie.Cond)
	if err != nil {
		t.Fatal(err)
	}
	if r != true {
		t.Fatalf("expected true, got %v", r)
	}
}

func TestDivisionByZero(t *testing.T) {
	v := New()
	v.Set("a", 10)
	v.Set("b", 0)
	expr, err := v.Compile(`a / b`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = v.Eval(expr)
	if err == nil {
		t.Fatal("expected division by zero error")
	}
}

func TestCallNonFunction(t *testing.T) {
	v := New()
	v.Set("notfunc", 42)
	expr, err := v.Compile(`notfunc()`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = v.Eval(expr)
	if err == nil {
		t.Fatal("expected error when calling non-function")
	}
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
