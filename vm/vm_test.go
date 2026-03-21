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

func TestIsTrue(t *testing.T) {
	tests := []struct {
		val    any
		expect bool
	}{
		{nil, false},
		{true, true},
		{false, false},
		{0, false},
		{1, true},
		{int32(0), false},
		{int32(1), true},
		{int64(0), false},
		{int64(1), true},
		{float32(0), false},
		{float32(1.5), true},
		{float64(0), false},
		{float64(1.5), true},
		{"", false},
		{"hello", true},
		{[]int{}, false},
		{[]int{1}, true},
		{map[string]int{}, false},
		{map[string]int{"a": 1}, true},
	}
	for _, tt := range tests {
		got := IsTrue(tt.val)
		if got != tt.expect {
			t.Fatalf("IsTrue(%v): expected %v, got %v", tt.val, tt.expect, got)
		}
	}
}

func TestFloatArithmetic(t *testing.T) {
	v := New()
	v.Set("a", 3.0)
	v.Set("b", 1.5)
	tests := []struct {
		expr   string
		expect float64
	}{
		{"a + b", 4.5},
		{"a - b", 1.5},
		{"a * b", 4.5},
		{"a / b", 2.0},
	}
	for _, tt := range tests {
		expr, err := v.Compile(tt.expr)
		if err != nil {
			t.Fatalf("compile %q: %v", tt.expr, err)
		}
		r, err := v.Eval(expr)
		if err != nil {
			t.Fatalf("eval %q: %v", tt.expr, err)
		}
		f, ok := r.(float64)
		if !ok || f != tt.expect {
			t.Fatalf("%q: expected %v, got %v", tt.expr, tt.expect, r)
		}
	}
}

func TestFloatComparison(t *testing.T) {
	v := New()
	v.Set("a", 3.0)
	tests := []struct {
		expr   string
		expect bool
	}{
		{"a < 4.0", true},
		{"a > 2.0", true},
		{"a <= 3.0", true},
		{"a >= 3.0", true},
	}
	for _, tt := range tests {
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

func TestStringComparison(t *testing.T) {
	v := New()
	v.Set("s", "hello")
	tests := []struct {
		expr   string
		expect bool
	}{
		{`s == "hello"`, true},
		{`s != "world"`, true},
		{`s < "world"`, true},
		{`s > "abc"`, true},
		{`s <= "hello"`, true},
		{`s >= "hello"`, true},
	}
	for _, tt := range tests {
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

func TestStringConcat(t *testing.T) {
	v := New()
	v.Set("a", "hello")
	v.Set("b", " world")
	expr, err := v.Compile(`a + b`)
	if err != nil {
		t.Fatal(err)
	}
	r, err := v.Eval(expr)
	if err != nil {
		t.Fatal(err)
	}
	if r != "hello world" {
		t.Fatalf("expected %q, got %v", "hello world", r)
	}
}

func TestUnknownOperatorError(t *testing.T) {
	v := New()
	v.Set("a", true)
	v.Set("b", true)
	expr, err := v.Compile(`a + b`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = v.Eval(expr)
	if err == nil {
		t.Fatal("expected error for bool addition")
	}
}

func TestMemberMap(t *testing.T) {
	v := New()
	v.Set("m", map[string]int{"x": 42})
	expr, err := v.Compile(`m.x`)
	if err != nil {
		t.Fatal(err)
	}
	r, err := v.Eval(expr)
	if err != nil {
		t.Fatal(err)
	}
	if r != 42 {
		t.Fatalf("expected 42, got %v", r)
	}
}

func TestItemSlice(t *testing.T) {
	v := New()
	v.Set("arr", []string{"a", "b", "c"})
	expr, err := v.Compile(`arr[1]`)
	if err != nil {
		t.Fatal(err)
	}
	r, err := v.Eval(expr)
	if err != nil {
		t.Fatal(err)
	}
	if r != "b" {
		t.Fatalf("expected b, got %v", r)
	}
}

func TestItemMap(t *testing.T) {
	v := New()
	v.Set("m", map[string]int{"key": 99})
	expr, err := v.Compile(`m["key"]`)
	if err != nil {
		t.Fatal(err)
	}
	r, err := v.Eval(expr)
	if err != nil {
		t.Fatal(err)
	}
	if r != 99 {
		t.Fatalf("expected 99, got %v", r)
	}
}

func TestNilLiteral(t *testing.T) {
	v := New()
	expr, err := v.Compile(`nil`)
	if err != nil {
		t.Fatal(err)
	}
	r, err := v.Eval(expr)
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Fatalf("expected nil, got %v", r)
	}
}

func TestUnlessExpr(t *testing.T) {
	v := New()
	v.Set("x", false)
	expr, err := v.Compile(`unless x`)
	if err != nil {
		t.Fatal(err)
	}
	ie, ok := expr.(*IfExpr)
	if !ok {
		t.Fatal("expected IfExpr")
	}
	if !ie.Neg {
		t.Fatal("expected Neg=true for unless")
	}
}

func TestCompileError(t *testing.T) {
	v := New()
	_, err := v.Compile(`+ +`)
	if err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestUndefinedIdent(t *testing.T) {
	v := New()
	expr, err := v.Compile(`unknown`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = v.Eval(expr)
	if err == nil {
		t.Fatal("expected error for undefined identifier")
	}
}

func TestUndefinedFunc(t *testing.T) {
	v := New()
	expr, err := v.Compile(`nope()`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = v.Eval(expr)
	if err == nil {
		t.Fatal("expected error for undefined function")
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
