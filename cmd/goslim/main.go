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
	m := make(map[string]string)
	for _, arg := range os.Args[1:] {
		token := strings.SplitN(arg, "=", 2)
		if len(token) == 2 {
			m[token[0]] = token[1]
		}
	}
	err = t.Execute(os.Stdout, m)
	fatalIf(err)
}
