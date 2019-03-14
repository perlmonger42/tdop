package scan

import (
	"fmt"
	"testing"
)

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
		input: `([]){}?.,:;~*/`,
		output: []wanted{
			{Punctuator, "("},
			{Punctuator, "["},
			{Punctuator, "]"},
			{Punctuator, ")"},
			{Punctuator, "{"},
			{Punctuator, "}"},
			{Punctuator, "?"},
			{Punctuator, "."},
			{Punctuator, ","},
			{Punctuator, ":"},
			{Punctuator, ";"},
			{Punctuator, "~"},
			{Punctuator, "*"},
			{Punctuator, "/"},
			{EOF, ""},
		},
	},
	{
		input: "+-<>+=-=<=>====!==!,=!=",
		output: []wanted{
			{Punctuator, "+"},
			{Punctuator, "-"},
			{Punctuator, "<"},
			{Punctuator, ">"},
			{Punctuator, "+="},
			{Punctuator, "-="},
			{Punctuator, "<="},
			{Punctuator, ">="},
			{Punctuator, "==="},
			{Punctuator, "!=="},
			{Punctuator, "!"},
			{Punctuator, ","},
			{Punctuator, "="},
			{Punctuator, "!"},
			{Punctuator, "="},
			{EOF, ""},
		},
	},
	{
		input: "000 1 42\n 3.1415926 1.2\n 3. .4",
		output: []wanted{
			{Number, "000"},
			{Number, "1"},
			{Number, "42"},
			{Number, "3.1415926"},
			{Number, "1.2"},
			{Number, "3"},
			{Punctuator, "."},
			{Punctuator, "."},
			{Number, "4"},
			{EOF, ""},
		},
	},
	{
		input: `"\a" "" "?" "howdy" "\"\n" "unfinished business`,
		output: []wanted{
			{String, `"\a"`},
			{String, `""`},
			{String, `"?"`},
			{String, `"howdy"`},
			{String, `"\"\n"`},
			{String, `"unfinished business`},
			{EOF, ""},
		},
	},
	{
		input: "  // comment\nX15",
		output: []wanted{
			{Name, "X15"},
			{EOF, ""},
		},
	},
}

func checkTestcase(t *testing.T, c *testcase) {
	reportInput := fmt.Sprintf("input: %#q\n", c.input)
	var i int
	var token Token
	tokens := TokenizeString(c.input)

	for i, token = range tokens {
		if i >= len(c.output) {
			if len(reportInput) > 0 {
				t.Errorf("%s", reportInput)
			}
			t.Errorf("  token %d: unexpected token %v",
				i, token)
			continue
		}
		w := c.output[i]
		if token.Type != w.Type {
			if len(reportInput) > 0 {
				t.Errorf("%s", reportInput)
			}
			t.Errorf("  token %d: wanted type %v, got %v",
				i, w.Type, token.Type)
			reportInput = ""
		}
		if token.Text != w.Text {
			if len(reportInput) > 0 {
				t.Errorf("%s", reportInput)
			}
			t.Errorf("  token %d: wanted text %q, got %q",
				i, w.Text, token.Text)
			reportInput = ""
		}
	}
	for i = len(tokens); i < len(c.output); i++ {
		if len(reportInput) > 0 {
			t.Errorf("%s", reportInput)
		}
		t.Errorf("  token %d: missing expected token %v\n",
			i, c.output[i])
		reportInput = ""
	}
}

func TestTokenizer(t *testing.T) {
	for i, _ := range testcases {
		checkTestcase(t, &testcases[i])
	}
}
