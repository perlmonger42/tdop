// This lexer is based on Douglas Crockford's Simplified JavaScript scanner,
// found at http://crockford.com/javascript/tdop/tokens.js

//go:generate stringer -type Type

package scan

import (
	"regexp"
	"strings"
)

// Type identifies the type of lex items.
type Type int

const (
	EOF   Type = iota // zero value so closed channel delivers EOF
	Error             // error occurred; value is text of error

	Name       // alphanumeric identifier
	Punctuator // ( ) { } [ ] ? . , : ; ~ * /
	Number
	String
)

var tokenRegex = regexp.MustCompile(`(\s+)|(\/\/.*)|([a-zA-Z][a-zA-Z_0-9]*)|(\d+(?:\.\d+)?(?:[eE][+\-]?\d+)?)|("(?:[^"\\]|\\(?:.|u[0-9a-fA-F]{4}))*"?)|([(){}\[\]?.,:;~*\/]|&&?|\|\|?|[+\-<>]=?|[!=](?:==)?)|(.)`)

// Token represents a lexical unit returned from the scanner.
type Token struct {
	Type   Type   // The type of this item.
	Text   string // The text of this item.
	Line   int    // The line number on which this token appears
	Column int    // The column number at which this token appears
}

func NewToken(kind Type, value string, line int, col int) Token {
	return Token{
		Type:   kind,
		Text:   value,
		Line:   line,
		Column: col,
	}
}

// TokenizeString analyzes the source string and returns it as an array of
// Tokens.
func TokenizeString(source string) []Token {
	return TokenizeLines(
		strings.Split(strings.Replace(source, "\r\n", "\n", -1), "\n"),
	)
}

// TokenizeLines analyzes the array of source strings and returns it as an array
// of Tokens.
func TokenizeLines(sourceLines []string) []Token {
	var lineNumber int = 0
	var lineText string = ""
	var loc []int
	var result = []Token{}

	emit := func(t Type, locIndex int) {
		first := loc[locIndex]
		after := loc[locIndex+1]
		result = append(result, Token{
			Type:   t,
			Text:   lineText[first:after],
			Line:   lineNumber + 1,
			Column: first,
		})
	}
	for lineNumber, lineText = range sourceLines {
		allIndexes := tokenRegex.FindAllStringSubmatchIndex(lineText, -1)
		for _, loc = range allIndexes {
			if loc[2] >= 0 {
				// skip whitespace
			} else if loc[4] >= 0 {
				// skip comment
			} else if loc[6] >= 0 {
				emit(Name, 6)
			} else if loc[8] >= 0 {
				emit(Number, 8)
			} else if loc[10] >= 0 {
				emit(String, 10)
			} else if loc[12] >= 0 {
				emit(Punctuator, 12)
			} else if loc[14] >= 0 {
				emit(Error, 14)
			} else {
				panic(`token regex didn't match *anything*`)
			}
		}
	}
	loc = []int{len(lineText), len(lineText)}
	emit(EOF, 0)
	return result
}
