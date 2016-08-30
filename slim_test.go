package slim

import (
	"fmt"
	"testing"
)

func TestSimple(t *testing.T) {
	tmpl, err := ParseFile("testdir/test.slim")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(tmpl)
}
