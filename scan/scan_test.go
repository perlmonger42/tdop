package scan

import (
	"fmt"
	"testing"
)

type wanted struct {
	Type
	Value string
}

type testcase struct {
	input  string
	output []wanted
	class  string
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
		},
	},
	{
		input: "000 1 42\n 3.1415926 1.2\n 3. .4 5.6e7",
		output: []wanted{
			{Fixnum, "000"},
			{Fixnum, "1"},
			{Fixnum, "42"},
			{Flonum, "3.1415926"},
			{Flonum, "1.2"},
			{Fixnum, "3"},
			{Punctuator, "."},
			{Punctuator, "."},
			{Fixnum, "4"},
			{Flonum, "5.6e7"},
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
			{UnterminatedString, `"unfinished business`},
		},
	},
	{
		input: "  // comment\nX15",
		output: []wanted{
			{Name, "X15"},
		},
	},
}

func checkTestcase(t *testing.T, c *testcase) {
	reportInput := fmt.Sprintf("input: %#q\n", c.input)
	var i int
	var token *Token
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
		if token.TkType != w.Type || token.TkValue != w.Value {
			if len(reportInput) > 0 {
				t.Errorf("%s", reportInput)
			}
			t.Errorf("  token %d: wanted {%v,%q}, got {%v,%q}",
				i, w.Type, w.Value, token.TkType, token.TkValue)
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
