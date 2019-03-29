package scan

import (
	"fmt"
	"io"
	"strings"
)

type AST interface {
	Error(message string)

	String() string
	PrettyPrint(b io.Writer, indent string)

	IsFuncall() bool
	IsAssignment() bool
}

type ASTimpl struct {
	token        *Token
	op           string
	first        AST
	second       AST
	third        AST
	list         []AST
	isAssignment bool
	isFuncall    bool
}

func (t *ASTimpl) Error(message string) {
	panic(fmt.Sprintf("SyntaxError;  %s while processing %v", message, t))
}

func TokenAST(t *Token) AST {
	return &ASTimpl{token: t,
		isAssignment: t.IsAssignment(),
		isFuncall:    t.IsFuncall(),
	}
}

func NewAST0(op string) AST {
	return &ASTimpl{op: op}
}

func NewAST1(op string, p AST) AST {
	return &ASTimpl{op: op, first: p}
}

func NewAST2(op string, p, q AST) AST {
	return &ASTimpl{op: op, first: p, second: q}
}
func NewAST3(op string, p, q, r AST) AST {
	return &ASTimpl{op: op, first: p, second: q, third: r}
}

func (t *ASTimpl) IsAssignment() bool {
	return t.isAssignment || t.token != nil && t.IsAssignment()
}

func (t *ASTimpl) IsFuncall() bool {
	return t.isFuncall || t.token != nil && t.IsFuncall()
}

func (t *ASTimpl) PrettyPrint(b io.Writer, indent string) {
	if t.token != nil {
		t.token.PrettyPrint(b, indent)
		return
	}

	assignment := ""
	if t.isAssignment {
		assignment = " assignment"
	}
	if t.first == nil && t.second == nil && t.third == nil && t.list == nil {
		fmt.Fprintf(b, "%#q%s", t.op, assignment)
		return
	}

	indented := indent + "  "
	fmt.Fprintf(b, "(%#q%s ", t.op, assignment)
	if t.first != nil {
		t.first.PrettyPrint(b, indented)
	}
	if t.second != nil {
		fmt.Fprintf(b, ",\n%s", indented)
		t.second.PrettyPrint(b, indented)
	}
	if t.third != nil {
		fmt.Fprintf(b, ",\n%s", indented)
		t.third.PrettyPrint(b, indented)
	}
	if t.list != nil {
		fmt.Fprintf(b, ", [")
		indented2 := indented + "  "
		comma := ""
		for _, v := range t.list {
			fmt.Fprintf(b, "%s\n%s", comma, indented2)
			v.PrettyPrint(b, indented2)
			comma = ","
		}
		fmt.Fprintf(b, "]")
	}
	fmt.Fprintf(b, ")")
}

func (t *ASTimpl) String() string {
	var s strings.Builder
	t.PrettyPrint(&s, "")
	return s.String()
}
