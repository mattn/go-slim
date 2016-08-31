package slim

import (
	"fmt"
	"strconv"
	"strings"
)

func Trim(s Value) (Value, error) {
	return strings.TrimSpace(fmt.Sprint(s)), nil
}

func ToUpper(s Value) (Value, error) {
	return strings.ToUpper(fmt.Sprint(s)), nil
}

func ToLower(s Value) (Value, error) {
	return strings.ToLower(fmt.Sprint(s)), nil
}

func Repeat(s Value, n Value) (Value, error) {
	i, err := strconv.ParseInt(fmt.Sprint(n), 10, 64)
	if err != nil {
		return nil, err
	}
	return strings.Repeat(fmt.Sprint(s), int(i)), nil
}
