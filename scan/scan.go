// This lexer is based on Douglas Crockford's Simplified JavaScript scanner,
// found at http://crockford.com/javascript/tdop/tokens.js

//go:generate stringer -type Type
////go:generate stringer -type Type/*Arity*/

package scan

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Type identifies the type of lex items.
type Type int

const (
	Unknown Type = iota
	Error        // error occurred; value is text of error

	Name       // alphanumeric identifier
	Punctuator // ( ) { } [ ] ? . , : ; ~ * /
	Fixnum
	Flonum
	String
	UnterminatedString

	Literal
	//)

	//type Type/*Arity*/ int

	//const (
	//	unknownArity Type/*Arity*/ = iota
	nameArity
	literalArity
	thisArity
	functionArity
	unaryArity
	binaryArity
	ternaryArity
	statementArity
	listArity
)

const reWhitespace = `(\s+)`
const reCommentToEol = `(\/\/.*)`
const reName = `([a-zA-Z][a-zA-Z_0-9]*)`
const reFixnum = `(\d+)`
const reFloat1 = `\d+[eE][+\-]?\d+`
const reFloat2 = `\d+\.\d+[eE][+\-]?\d+`
const reFloat3 = `\d+\.\d+`
const reFlonum = `(` + reFloat1 + `|` + reFloat2 + `|` + reFloat3 + `)`
const reString = `("(?:[^"\\]|\\(?:.|u[0-9a-fA-F]{4}))*")`
const rePunctuator = `([(){}\[\]?.,:;~*\/]|&&?|\|\|?|[+\-<>]=?|[!=](?:==)?)`
const reUnterminatedString = `("(?:[^"\\]|\\(?:.|u[0-9a-fA-F]{4}))*)`
const reError = `(.)`

var tokenRegex = regexp.MustCompile(
	strings.Join([]string{
		reWhitespace,
		reCommentToEol,
		reName,
		reFlonum,
		reFixnum,
		reString,
		rePunctuator,
		reUnterminatedString,
		reError,
	},
		"|"),
)

type UnaryDenotation func(this *Token) *Token
type BinaryDenotation func(this, left *Token) *Token

// Token represents a lexical unit returned from the scanner.
type Token struct {
	Type   Type   // The type of this item.
	Value  string // The text of this item.
	Line   int    // The line number on which this token appears
	Column int    // The column number at which this token appears

	id       string
	arity    Type /*Arity*/
	reserved bool
	nud      UnaryDenotation
	led      BinaryDenotation
	std      UnaryDenotation
	lbp      int

	assignment bool
	first      *Token
	second     *Token
	third      *Token
	list       []*Token
	name       string
	key        string
}

func (t *Token) Error(message string) {
	panic(fmt.Sprintf("SyntaxError;  %s while processing %v", message, t))
}

func (t *Token) PrettyPrint(b io.Writer, indent string) {
	fmt.Fprintf(b, "%s%s(%s %q)@%d:%d",
		indent, t.Type, t.arity, t.Value, t.Line, t.Column)
	if t.id != "" {
		fmt.Fprintf(b, "[id %s]", t.id)
	}
	if t.reserved {
		fmt.Fprintf(b, " reserved")
	}
	if t.assignment {
		fmt.Fprintf(b, " assignment")
	}
	if t.name != "" {
		fmt.Fprintf(b, " name:%q", t.name)
	}
	if t.key != "" {
		fmt.Fprintf(b, " key:%q", t.key)
	}
	//	if t.list == nil || len(t.list) == 0 {
	//		fmt.Fprintf(b, " list:empty")
	//	}
	fmt.Fprintf(b, "\n")
	indented := indent + "  "
	if t.first != nil {
		t.first.PrettyPrint(b, indented)
	}
	if t.second != nil {
		t.second.PrettyPrint(b, indented)
	}
	if t.third != nil {
		t.third.PrettyPrint(b, indented)
	}
	if t.list != nil && len(t.list) > 0 {
		fmt.Fprintf(b, "%slist:\n", indented)
		indented += "  "
		for _, item := range t.list {
			item.PrettyPrint(b, indented)
		}
	}
}

func (t *Token) String() string {
	var s strings.Builder
	t.PrettyPrint(&s, "")
	return s.String()
}

// TokenizeString analyzes the source string and returns it as an array of
// Tokens.
func TokenizeString(source string) []*Token {
	return TokenizeLines(
		strings.Split(strings.Replace(source, "\r\n", "\n", -1), "\n"),
	)
}

// TokenizeLines analyzes the array of source strings and returns it as an array
// of Tokens.
func TokenizeLines(sourceLines []string) []*Token {
	var lineNumber int = 0
	var linevalue string = ""
	var loc []int
	var result = []*Token{}

	emit := func(t Type, locIndex int) {
		first := loc[locIndex]
		after := loc[locIndex+1]
		result = append(result, &Token{
			Type:   t,
			Value:  linevalue[first:after],
			Line:   lineNumber + 1,
			Column: first,
			id:     linevalue[first:after],
		})
	}
	for lineNumber, linevalue = range sourceLines {
		allIndexes := tokenRegex.FindAllStringSubmatchIndex(linevalue, -1)
		for _, loc = range allIndexes {
			if loc[2] >= 0 {
				// skip whitespace
			} else if loc[4] >= 0 {
				// skip comment
			} else if loc[6] >= 0 {
				emit(Name, 6)
			} else if loc[8] >= 0 {
				emit(Flonum, 8)
			} else if loc[10] >= 0 {
				emit(Fixnum, 10)
			} else if loc[12] >= 0 {
				emit(String, 12)
			} else if loc[14] >= 0 {
				emit(Punctuator, 14)
			} else if loc[16] >= 0 {
				emit(UnterminatedString, 16)
			} else if loc[18] >= 0 {
				emit(Error, 18)
			} else {
				panic(`token regex didn't match *anything*`)
			}
		}
	}
	loc = []int{len(linevalue), len(linevalue)}
	return result
}
