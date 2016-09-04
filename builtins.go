package slim

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Trim is builtin function provide trim(s).
func Trim(args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, errors.New("trim require 1 argument")
	}
	return strings.TrimSpace(fmt.Sprint(args[0])), nil
}

// ToUpper is builtin function provide to_upper(s).
func ToUpper(args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, errors.New("to_upper require 1 argument")
	}
	return strings.ToUpper(fmt.Sprint(args[0])), nil
}

// ToLower is builtin function provide to_lower(s).
func ToLower(args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, errors.New("to_lower require 1 argument")
	}
	return strings.ToLower(fmt.Sprint(args[0])), nil
}

// Repeat is builtin function provide repeat(s, n).
func Repeat(args ...Value) (Value, error) {
	if len(args) != 2 {
		return nil, errors.New("repeat require 2 arguments")
	}
	i, err := strconv.ParseInt(fmt.Sprint(args[1]), 10, 64)
	if err != nil {
		return nil, err
	}
	return strings.Repeat(fmt.Sprint(args[0]), int(i)), nil
}
