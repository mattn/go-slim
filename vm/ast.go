package vm

type Expr interface {
}

type IdentExpr struct {
	name string
}

type LitExpr struct {
	value interface{}
}

type RangeExpr struct {
	lhs1 string
	lhs2 string
	rhs  string
}
