package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-slim"
)

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	t, err := slim.Parse(os.Stdin)
	fatalIf(err)
	m := make(map[string]interface{})
	for _, arg := range os.Args[1:] {
		token := strings.SplitN(arg, "=", 2)
		if len(token) == 2 {
			if v, ok := m[token[0]]; ok {
				if a, ok := v.([]string); ok {
					m[token[0]] = append(a, token[1])
				} else {
					m[token[0]] = []string{v.(string), token[1]}
				}
			} else {
				m[token[0]] = token[1]
			}
		}
	}
	err = t.Execute(os.Stdout, m)
	fatalIf(err)
}
