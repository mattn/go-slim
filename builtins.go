package slim

import (
	"fmt"
	"strings"
)

func Trim(v Value) (Value, error) {
	return strings.TrimSpace(fmt.Sprint(v)), nil
}
