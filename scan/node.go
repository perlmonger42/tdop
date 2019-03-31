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

	First() AST
	Second() AST
	Third() AST
	List() []AST

	IsAssignment() bool
	IsFuncall() bool
	//GetKind() string
	GetLiteral() (value string, ok bool)
	GetKind() (value string, ok bool)
	GetName() (value string, ok bool)
	GetOp() (op string, ok bool)
}

type ASTimpl struct {
	token        *Token
	kind         string
	value        string // only used when kind == "(literal")
	name         string // only used when kind == "(name)"
	op           string // only used when kind == "binary operator left associative" or "binary operator right associative"
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

func NewTokenAST(t *Token) AST {
	return &ASTimpl{token: t,
		isAssignment: t.IsAssignment(),
		isFuncall:    t.IsFuncall(),
	}
}

func (t *ASTimpl) First() AST  { return t.first }
func (t *ASTimpl) Second() AST { return t.second }
func (t *ASTimpl) Third() AST  { return t.third }
func (t *ASTimpl) List() []AST { return t.list }

func NewLiteralAST(value string) AST {
	return &ASTimpl{kind: "(literal)", value: value}
}

func NewNameAST(name string) AST {
	return &ASTimpl{kind: "(name)", name: name}
}

func NewBinaryOperatorAST(op string, left, right AST) AST {
	return &ASTimpl{kind: "binary operator left associative", op: op, first: left, second: right}
}

func NewRightAssocBinaryOperatorAST(op string, left, right AST) AST {
	return &ASTimpl{kind: "binary operator right associative", op: op, first: left, second: right}
}

func NewAST0(kind string) AST {
	return &ASTimpl{kind: kind}
}

func NewAST1(kind string, p AST) AST {
	return &ASTimpl{kind: kind, first: p}
}

func NewAST2(kind string, p, q AST) AST {
	return &ASTimpl{kind: kind, first: p, second: q}
}
func NewAST3(kind string, p, q, r AST) AST {
	return &ASTimpl{kind: kind, first: p, second: q, third: r}
}

func (t *ASTimpl) IsAssignment() bool {
	return t.isAssignment || t.token != nil && t.IsAssignment() ||
		t.kind == "="
}

func (t *ASTimpl) IsFuncall() bool {
	return t.isFuncall || t.token != nil && t.IsFuncall()
}

func (t *ASTimpl) GetLiteral() (value string, ok bool) {
	if t.token != nil {
		value, ok = t.token.GetLiteral()
		return
	}
	if t.kind == "(literal)" {
		return t.value, true
	}
	return "", false
}

func (t *ASTimpl) GetKind() (value string, ok bool) {
	if t.token != nil {
		value, ok = t.GetOp()
		return
	}
	return "", false
}

func (t *ASTimpl) GetName() (value string, ok bool) {
	if t.kind == "(name)" {
		return t.name, true
	}
	return "", false
}

func (t *ASTimpl) GetOp() (op string, ok bool) {
	if t.kind == "binary operator left associative" || t.kind == "binary operator right associative" {
		return t.op, true
	}
	return "", false
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
		fmt.Fprintf(b, "%#q%s", t.kind, assignment)
		return
	}

	indented := indent + "  "
	fmt.Fprintf(b, "(%#q%s ", t.kind, assignment)
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
