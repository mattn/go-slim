package main

import (
	"fmt"
	"os"

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
	err = t.Execute(os.Stdout, nil)
	fatalIf(err)
}
