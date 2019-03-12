package scan

import (
	"fmt"
	"strings"
	"testing"
)

func expectToken(t *testing.T, scanner *Scanner, tokenType Type, tokenText string) {
	token := scanner.Next()
	if token.Type != tokenType {
		t.Errorf("line %d: wanted type %v, got %v", token.Line, tokenType, token.Type)
	}
	if token.Text != tokenText {
		t.Errorf("line %d: wanted text %q, got %q", token.Line, tokenText, token.Text)
	}
}

func TestPunctuation(t *testing.T) {
	stringReader := strings.NewReader("( [ ] )")
	scanner := NewScanner("<string>", stringReader)
	want := func(tokenType Type, tokenText string) {
		expectToken(t, scanner, tokenType, tokenText)
	}
	want(LeftParen, "(")
	want(LeftBrack, "[")
	want(RightBrack, "]")
	want(RightParen, ")")
	want(EOF, "<EOF>")
}

type wanted struct {
	Type
	Text string
}

type testcase struct {
	input  string
	output []wanted
}

var testcases []testcase = []testcase{
	{
		input: "([])",
		output: []wanted{
			{LeftParen, "("},
			{LeftBrack, "["},
			{RightBrack, "]"},
			{RightParen, ")"},
			{EOF, "<EOF>"},
		},
	},
	{
		input: "  // comment\nX15",
		output: []wanted{
			{Identifier, "X15"},
			{EOF, "<EOF>"},
		},
	},
	{
		input: "000 1 42\n 3.1415926 1.2\n 3. .4",
		output: []wanted{
			{Fixnum, "000"},
			{Fixnum, "1"},
			{Fixnum, "42"},
			{Flonum, "3.1415926"},
			{Flonum, "1.2"},
			{Flonum, "3."},
			{Flonum, ".4"},
			{EOF, "<EOF>"},
		},
	},
	{
		input: `"" "?" "howdy" "\"\x" "unfinished business`,
		output: []wanted{
			{String, `""`},
			{String, `"?"`},
			{String, `"howdy"`},
			{String, `"\"\x"`},
			{Error, "unterminated quoted string"},
			{EOF, "<EOF>"},
		},
	},
}

func checkTestcase(t *testing.T, c *testcase) {
	stringReader := strings.NewReader(c.input)
	reportInput := fmt.Sprintf("input: %q\n", c.input)
	scanner := NewScanner("<string>", stringReader)
	for i, w := range c.output {
		token := scanner.Next()
		if token.Type != w.Type {
			t.Errorf("%s  token %d: wanted type %v, got %v",
				reportInput, i, w.Type, token.Type)
			reportInput = ""
		}
		if token.Text != w.Text {
			t.Errorf("%s  token %d: wanted text %q, got %q",
				reportInput, i, w.Text, token.Text)
		}
	}
}

func TestTokenizer(t *testing.T) {
	for i, _ := range testcases {
		checkTestcase(t, &testcases[i])
	}
}
