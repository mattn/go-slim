package main

import (
	"log"
	"os"

	"github.com/mattn/go-slim"
)

func main() {
	t, err := slim.Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(os.Stdout, nil)
}
