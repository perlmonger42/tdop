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

	IsPossibleStatementExpression() bool
	GetLiteral() (value string, ok bool)
	GetName() (value string, ok bool)
	GetFundef() (name string, formals []string, body []AST, ok bool)

	GetOp() (op string, ok bool)
	GetFuncall() (callee AST, actuals []AST, ok bool)
	GetFieldAccess() (lhs AST, fieldname string, ok bool)

	GetBlock() (statements []AST, ok bool)
	GetAssignment() (lhs, rhs AST, ok bool)
	GetReturn() (value AST, ok bool)
	GetIf() (test, ifBlock, elseBlock AST, ok bool)
}

type ASTimpl struct {
	kind    string
	value   string // only used when kind == "(literal")
	name    string // only used when kind == "(name)" or kind == "(fundef)" or kind == "(fieldaccess)"
	op      string // only used when kind == "binary operator left associative" or "binary operator right associative"
	first   AST
	second  AST
	third   AST
	list    []AST
	formals []string // only used when kind == "(fundef)"
	body    []AST    // only used when kind == "(fundef)"

	callee  AST   // only used when kind == "(funcall)"
	actuals []AST // only used when kind == "(funcall)"
}

func (t *ASTimpl) Error(message string) {
	panic(fmt.Sprintf("SyntaxError;  %s while processing %v", message, t))
}

func (t *ASTimpl) First() AST  { return t.first }
func (t *ASTimpl) Second() AST { return t.second }
func (t *ASTimpl) Third() AST  { return t.third }
func (t *ASTimpl) List() []AST { return t.list }

func NewLiteralAST(value string) AST {
	return &ASTimpl{kind: "(literal)", value: value}
}
func (t *ASTimpl) GetLiteral() (value string, ok bool) {
	if t.kind == "(literal)" {
		return t.value, true
	}
	return "", false
}

func NewNameAST(name string) AST {
	return &ASTimpl{kind: "(name)", name: name}
}
func (t *ASTimpl) GetName() (value string, ok bool) {
	if t.kind == "(name)" {
		return t.name, true
	}
	return "", false
}

func NewAssignmentAST(lhs, rhs AST) AST {
	return &ASTimpl{kind: "(assignment)", first: lhs, second: rhs}
}
func (t *ASTimpl) GetAssignment() (lhs, rhs AST, ok bool) {
	if t.kind == "(assignment)" {
		return t.first, t.second, true
	}
	return nil, nil, false
}
func (t *ASTimpl) IsPossibleStatementExpression() bool {
	return t.kind == "(assignment)" || t.kind == "(funcall)"
}

func NewReturnAST(value AST) AST {
	return &ASTimpl{kind: "(return)", first: value}
}
func (t *ASTimpl) GetReturn() (value AST, ok bool) {
	if t.kind == "(return)" {
		return t.first, true
	}
	return nil, false
}

func NewIfAST(test, bTrue, bFalse AST) AST {
	return &ASTimpl{kind: "(if)", first: test, second: bTrue, third: bFalse}
}
func (t *ASTimpl) GetIf() (test, bTrue, bFalse AST, ok bool) {
	if t.kind == "(if)" {
		return t.first, t.second, t.third, true
	}
	return nil, nil, nil, false
}

func NewBinaryOperatorAST(op string, left, right AST) AST {
	return &ASTimpl{kind: "binary operator left associative", op: op, first: left, second: right}
}
func NewRightAssocBinaryOperatorAST(op string, left, right AST) AST {
	return &ASTimpl{kind: "binary operator right associative", op: op, first: left, second: right}
}
func (t *ASTimpl) GetOp() (op string, ok bool) {
	if t.kind == "binary operator left associative" || t.kind == "binary operator right associative" {
		return t.op, true
	}
	return "", false
}

func NewFundefAST(name string, formals []string, body []AST) AST {
	return &ASTimpl{kind: "(fundef)", name: name, formals: formals, body: body}
}
func (t *ASTimpl) GetFundef() (name string, args []string, body []AST, ok bool) {
	if t.kind == "(fundef)" {
		return t.name, t.formals, t.body, true
	}
	return "", nil, nil, false
}

func NewFuncallAST(callee AST, actuals []AST) AST {
	return &ASTimpl{kind: "(funcall)", callee: callee, actuals: actuals}
}
func (t *ASTimpl) GetFuncall() (callee AST, actuals []AST, ok bool) {
	if t.kind == "(funcall)" {
		return t.callee, t.actuals, true
	}
	return nil, nil, false
}

func NewBlockAST(statements []AST) AST {
	return &ASTimpl{kind: "(block)", list: statements}
}
func (t *ASTimpl) GetBlock() (statements []AST, ok bool) {
	if t.kind == "(block)" {
		return t.list, true
	}
	return nil, false
}

func NewFieldAccessAST(lhs AST, fieldname string) AST {
	return &ASTimpl{kind: "(fieldaccess)", first: lhs, name: fieldname}
}
func (t *ASTimpl) GetFieldAccess() (lhs AST, fieldname string, ok bool) {
	if t.kind == "(fieldaccess)" {
		return t.first, t.name, true
	}
	return nil, "", false
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

func (t *ASTimpl) IsFuncall() bool {
	return t.kind == "(funcall)"
}

func (t *ASTimpl) PrettyPrint(b io.Writer, indent string) {
	if t.first == nil && t.second == nil && t.third == nil && t.list == nil {
		fmt.Fprintf(b, "%#q", t.kind)
		return
	}

	indented := indent + "  "
	fmt.Fprintf(b, "(%#q ", t.kind)
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
