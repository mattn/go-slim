package slim

import (
	"fmt"
	"strings"
)

func Trim(v Value) (Value, error) {
	return strings.TrimSpace(fmt.Sprint(v)), nil
}

func ToUpper(v Value) (Value, error) {
	return strings.ToUpper(fmt.Sprint(v)), nil
}

func ToLower(v Value) (Value, error) {
	return strings.ToLower(fmt.Sprint(v)), nil
}
