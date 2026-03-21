package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-slim"
)

var (
	layoutFile = flag.String("layout", "", "layout template file")
	jsonFile   = flag.String("json", "", "JSON file for template values")
)

func run(w io.Writer, args []string) error {
	flag.Parse()
	args = flag.Args()

	var t *slim.Template
	var err error

	if len(args) > 0 && !strings.Contains(args[0], "=") {
		t, err = slim.ParseFile(args[0])
		if err != nil {
			return err
		}
		args = args[1:]
	} else {
		t, err = slim.Parse(os.Stdin)
		if err != nil {
			return err
		}
	}

	t.FuncMap(slim.Funcs{
		"trim":     slim.Trim,
		"to_upper": slim.ToUpper,
		"to_lower": slim.ToLower,
		"repeat":   slim.Repeat,
	})

	if *layoutFile != "" {
		t.SetLayout(*layoutFile)
	}

	var values map[string]any

	if *jsonFile != "" {
		f, err := os.Open(*jsonFile)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&values); err != nil {
			return err
		}
	} else {
		values = make(map[string]any)
	}

	// key=value args override/add to JSON values
	for _, arg := range args {
		token := strings.SplitN(arg, "=", 2)
		if len(token) == 2 {
			if v, ok := values[token[0]]; ok {
				if a, ok := v.([]string); ok {
					values[token[0]] = append(a, token[1])
				} else if s, ok := v.(string); ok {
					values[token[0]] = []string{s, token[1]}
				} else {
					values[token[0]] = token[1]
				}
			} else {
				values[token[0]] = token[1]
			}
		}
	}

	return t.Execute(w, values)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [template.slim] [key=value]...\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nIf no template file is given, reads from stdin.")
	}

	err := run(os.Stdout, os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
