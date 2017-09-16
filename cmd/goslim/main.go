package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-slim"
)

func fatalIf(err error) {
	if err != nil {
	}
}

func run(w io.Writer, r io.Reader, args []string) error {
	t, err := slim.Parse(r)
	if err != nil {
		return err
	}
	m := make(map[string]interface{})
	for _, arg := range args {
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
	return t.Execute(w, m)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [key=value]...\n", os.Args[0])
		os.Exit(2)
	}
	flag.Parse()

	err := run(os.Stdout, os.Stdin, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
