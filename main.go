package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/perlmonger42/tdop/scan"
)

func writer() *tabwriter.Writer {
	minWidth := 0
	tabWidth := 8
	padding := 0
	padchar := byte(' ')
	flags := uint(0) // tabwriter.TabIndent | tabwriter.Debug
	return tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, padchar, flags)
}

func main() {
	stringReader := strings.NewReader("Hello, world!\n")
	scanner := scan.NewScanner("<string>", stringReader)
	token := scan.Token{Type: scan.Error}
	w := writer()
	for i := 1; token.Type != scan.EOF; i++ {
		token = scanner.Next()
		fmt.Fprintf(w, "%d:\t %s \t %q \t\n", i, token.Type, token.Text)
	}
	w.Flush()
}

//OUTPUT:
// 1: Symbol   "Hello"
// 2: Unquote  ","
// 3: Symbol   "world!"
// 4: EOF      "<EOF>"
