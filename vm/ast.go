package vm

// Expr is a type for indicating expression.
type Expr interface {
}

// BinOpExpr is a type for indicating binary operator.
type BinOpExpr struct {
	Op  string
	LHS Expr
	RHS Expr
}

// IdentExpr is a type for indicating ident.
type IdentExpr struct {
	Name string
}

// LitExpr is a type for indicating literals.
type LitExpr struct {
	Value any
}

// NotExpr is a type for indicating logical not.
type NotExpr struct {
	Expr Expr
}

// IfExpr is a type for indicating if/elsif/else.
type IfExpr struct {
	Cond Expr
	Neg  bool // true for unless
}

// ForExpr is a type for indicating expression.
type ForExpr struct {
	LHS1 string
	LHS2 string
	RHS  Expr
}

// YieldExpr is a type for indicating yield block.
type YieldExpr struct {
	Name string
}

// CallExpr is a type for indicating calling functions.
type CallExpr struct {
	Name  string
	Exprs []Expr
}

// MethodCallExpr is a type for indicating calling methods.
type MethodCallExpr struct {
	LHS   Expr
	Name  string
	Exprs []Expr
}

// MemberExpr is a type for indicating reference member or fields.
type MemberExpr struct {
	LHS  Expr
	Name string
}

// ItemExpr is a type for indicating reference items in map.
type ItemExpr struct {
	LHS   Expr
	Index Expr
}
