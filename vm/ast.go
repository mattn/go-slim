package vm

type Expr interface {
}

type BinOpExpr struct {
	Op  string
	Lhs Expr
	Rhs Expr
}

type IdentExpr struct {
	Name string
}

type LitExpr struct {
	Value interface{}
}

type ForExpr struct {
	Lhs1 string
	Lhs2 string
	Rhs  Expr
}

type CallExpr struct {
	Name  string
	Exprs []Expr
}

type MemberExpr struct {
	Lhs  Expr
	Name string
}
