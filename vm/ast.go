package vm

type Expr interface {
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
	Rhs  string
}

type CallExpr struct {
	Name  string
	Exprs []Expr
}
