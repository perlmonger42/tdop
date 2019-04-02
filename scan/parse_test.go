package scan

import (
	"strings"
	"testing"
)

func recoverFromPanic(t *testing.T) {
	if r := recover(); r != nil {
		if t == nil {
			panic("recoverFromPanic: token is nil!")
		}
		t.Error(r)
	}
}

func parseString(source string) []AST {
	var parser = NewParser()
	return parser.Parse(TokenizeString(source))
}

func TestLet(t *testing.T) {
	defer recoverFromPanic(t)
	source := "let answer = 42;"
	stmts := parseString(source)
	if len(stmts) != 1 {
		t.Errorf("expected exactly one statement; got %#v\n", stmts)
		return
	}
	stmt := stmts[0]
	lhs, rhs, ok := stmt.GetAssignment()
	if !ok {
		t.Errorf("expected assignment; got %#v\n", stmt)
		return
	}
	if varname, ok := lhs.GetName(); !ok || varname != "answer" {
		t.Errorf("expected lhs to be `answer`; got %#v", lhs)
	}
	if val, ok := rhs.GetLiteral(); !ok || val != "42" {
		t.Errorf("expected lhs to be `42`; got %#v", rhs)
	}
}

func TestExpression(t *testing.T) {
	defer recoverFromPanic(t)
	source := "let x;\nx = 1+2*3/(4-5);"
	stmts := parseString(source)
	if len(stmts) != 1 {
		t.Errorf("expected exactly one statement; got %#v\n", stmts)
		return
	}
	// fmt.Printf("\n>>> %s\n%s", source, assign)
	stmt := stmts[0]
	lhs, rhs, ok := stmt.GetAssignment()
	if !ok {
		t.Errorf("expected assignment; got %#v\n", stmt)
		return
	}
	if name, ok := lhs.GetName(); !ok || name != "x" {
		t.Errorf("expected `x` on lhs of assignment; got %#v\n", lhs)
	}
	plus := rhs
	if op, ok := plus.GetOp(); !ok || op != "+" {
		t.Errorf("expected `+` as top-level operator on rhs; got op: %s, subtree: %#v\n", op, plus)
	}
	one := plus.First()
	if op, ok := one.GetLiteral(); !ok || op != "1" {
		t.Errorf("expected `1` as left op of `+`; got %#v\n", one)
	}
	div := plus.Second()
	if op, ok := div.GetOp(); !ok || op != "/" {
		t.Errorf("expected `/` as right op of `+`; got %#v\n", div)
	}
	times := div.First()
	if op, ok := times.GetOp(); !ok || op != "*" {
		t.Errorf("expected `*` as left op of `/`; got %#v\n", times)
	}
	two := times.First()
	if val, ok := two.GetLiteral(); !ok || val != "2" {
		t.Errorf("expected `2` as left op of `*`; got %#v\n", two)
	}
	three := times.Second()
	if val, ok := three.GetLiteral(); !ok || val != "3" {
		t.Errorf("expected `3` as right op of `*`; got %#v\n", three)
	}
	minus := div.Second()
	if op, ok := minus.GetOp(); !ok || op != "-" {
		t.Errorf("expected `-` as right op of `/`; got %#v\n", minus)
	}
	four := minus.First()
	if val, ok := four.GetLiteral(); !ok || val != "4" {
		t.Errorf("expected `4` as left op of `-`; got %#v\n", four)
	}
	five := minus.Second()
	if val, ok := five.GetLiteral(); !ok || val != "5" {
		t.Errorf("expected `5` as right op of `-`; got %#v\n", five)
	}
}

func TestFuncDef(t *testing.T) {
	defer recoverFromPanic(t)
	source := "let f = function (){};"
	stmts := parseString(source)
	if len(stmts) != 1 {
		t.Errorf("expected exactly one statement; got %#v\n", stmts)
		return
	}
	stmt := stmts[0]
	lhs, rhs, ok := stmt.GetAssignment()
	if !ok {
		t.Errorf("expected assignment; got %#v\n", stmt)
		return
	}
	if name, ok := lhs.GetName(); !ok || name != "f" {
		t.Errorf("expected `f` on lhs of assignment; got %#v\n", lhs)
	}
	name, args, body, ok := rhs.GetFundef()
	if !ok {
		t.Errorf("expected fundef on rhs of assignment; got %#v\n", rhs)
		return
	}
	if name != "" {
		t.Errorf("expected unnamed function; got name %#v\n", name)
	}
	if len(args) > 0 {
		t.Errorf("expected no formals for function; got %#v\n", args)
	}
	if len(body) != 0 {
		t.Errorf("expected empty body for function; got %#v\n", body)
	}
}

func TestLetAndAssignment(t *testing.T) {
	defer recoverFromPanic(t)
	source := "{\nlet answer;\nanswer = 42;\n}"
	stmts := parseString(source)
	if len(stmts) != 1 {
		t.Errorf("expected exactly one statement; got %#v\n", stmts)
		return
	}
	stmt := stmts[0]
	body, ok := stmt.GetBlock()
	if !ok {
		t.Errorf("expected statement to be a block; got %#v\n", stmt)
	}
	if len(body) != 1 {
		t.Errorf("expected exactly one statement in block; got %#v\n", body)
	}
	stmt = body[0]
	lhs, rhs, ok := stmt.GetAssignment()
	if !ok {
		t.Errorf("expected assignment; got %#v\n", stmt)
		return
	}
	if name, ok := lhs.GetName(); !ok || name != "answer" {
		t.Errorf("expected `f` on lhs of assignment; got %#v\n", lhs)
	}
	if val, ok := rhs.GetLiteral(); !ok || val != "42" {
		t.Errorf("expected lhs to be `42`; got %#v", rhs)
	}
}

func TestBlock(t *testing.T) {
	defer recoverFromPanic(t)
	source := `let token, advance, block = function (x, y) {
	        let t = token;
	        advance("{");
	        return t.TkStd();
	    };`
	stmts := parseString(source)
	if len(stmts) != 1 {
		t.Errorf("expected exactly one statement; got %#v\n", stmts)
		return
	}
	stmt := stmts[0]

	// check parse of:        block = function (x, y) { ... }
	lhs, rhs, ok := stmt.GetAssignment()
	if !ok {
		t.Errorf("expected assignment; got %#v\n", stmt)
	}
	if name, ok := lhs.GetName(); !ok || name != "block" {
		t.Errorf("expected lhs to be `block`; got %#v\n", lhs)
	}
	name, formals, body, ok := rhs.GetFundef()
	if !ok {
		t.Errorf("expected rhs to be fundef; got %#v\n", rhs)
	}
	if name != "" {
		t.Errorf("expected unnamed fundef; got %#v\n", name)
	}
	if len(formals) != 2 || formals[0] != "x" || formals[1] != "y" {
		t.Errorf("expected `(x,y)`; got %#v\n", formals)
	}
	if len(body) != 3 {
		t.Errorf("expected fundef to have 3 statements; got %#v\n", body)
	}

	// check parse of:        let t = token;
	lhs, rhs, ok = body[0].GetAssignment()
	if !ok {
		t.Errorf("expected 1st stmt to be assignment; got %#v\n", body[0])
	}
	if name, ok = lhs.GetName(); !ok || name != "t" {
		t.Errorf("expected 1st stmt to be assignment to `t`; got %#v\n", lhs)
	}
	if name, ok = rhs.GetName(); !ok || name != "token" {
		t.Errorf("expected 1st stmt to be assignment from `token`; got %#v\n", rhs)
	}

	// check parse of:        advance("{");
	callee, actuals, ok := body[1].GetFuncall()
	if !ok {
		t.Errorf("expected 2nd stmt to be funcall; got #%v\n", body[1])
	}
	if name, ok = callee.GetName(); !ok || name != "advance" {
		t.Errorf("expected 2nd stmt callee to be `advance`; got #%v\n", callee)
	}
	if len(actuals) != 1 {
		t.Errorf("expected exactly 1 actual passed to callee; got #%v\n", actuals)
	}
	if literal, ok := actuals[0].GetLiteral(); !ok || literal != `"{"` {
		t.Errorf(`expected actual to be "{"; got #%v\n`, actuals[0])
	}

	// check parse of:        return t.TkStd();
	value, ok := body[2].GetReturn()
	if !ok || value == nil {
		t.Errorf("expected 3rd stmt to be return of value; got #%v\n", body[2])
	}
	if callee, actuals, ok = value.GetFuncall(); !ok || callee == nil {
		t.Errorf("expected return value to be funcall; got #%v\n", value)
	}
	lhs, fieldname, ok := callee.GetFieldAccess()
	if !ok || fieldname != "TkStd" {
		t.Errorf("expected return to call `t.TkStd`; got #%v (fieldname %q)\n", callee, fieldname)
	}
	if name, ok := lhs.GetName(); !ok || name != "t" {
		t.Errorf("expected `t` on lhs of `.TkStd`; got #%v\n", lhs)
	}
	if len(actuals) != 0 {
		t.Errorf("expected 3rd stmt to pass 0 args; got #%v\n", actuals)
	}
}

func TestDoubleDeclaration(t *testing.T) {
	var stmts []AST
	defer func() {
		if len(stmts) != 0 {
			t.Errorf("expected no statements returned; got %#v", stmts)
		}
		if r := recover(); r != nil {
			s, ok := r.(string)
			if !ok || !strings.Contains(s, "Already defined") {
				t.Errorf("Expected `Already defined`; got %s", r)
			}
		} else {
			t.Errorf("expected a panic to occur")
		}
	}()
	source := `let x; let x; }`
	stmts = parseString(source)
}

func TestIf(t *testing.T) {
	defer recoverFromPanic(t)
	source := `let x, y, z; if (x) { y(); } else { z(); }`
	stmts := parseString(source)
	if len(stmts) != 1 {
		t.Errorf("expected exactly one statement; got %#v\n", stmts)
		return
	}
	stmt := stmts[0]
	test, ifBlock, elseBlock, ok := stmt.GetIf()
	if !ok || test == nil || ifBlock == nil || elseBlock == nil {
		t.Errorf("expected if-statement; got %#v\n", stmt)
	}
	if name, ok := test.GetName(); !ok || name != "x" {
		t.Errorf("expected if-test to be `x`; got %#v\n", test)
	}
	// check if-block
	ifStmts, ok := ifBlock.GetBlock()
	if !ok || len(ifStmts) != 1 {
		t.Errorf("expected one stmt in true branch; got %#v\n", ifBlock)
	}
	callee, actuals, ok := ifStmts[0].GetFuncall()
	if !ok {
		t.Errorf("expected true branch to be a funcall; got %#v\n", ifStmts[0])
	}
	if name, ok := callee.GetName(); !ok || name != "y" {
		t.Errorf("expected true branch to call `y`; got %#v\n", callee)
	}
	if len(actuals) != 0 {
		t.Errorf("expected true branch to pass 0 args; %#v\n", actuals)
	}
	// check else-block
	elseStmts, ok := elseBlock.GetBlock()
	if !ok || len(elseStmts) != 1 {
		t.Errorf("expected one stmt in false branch; got %#v\n", elseBlock)
	}
	callee, actuals, ok = elseStmts[0].GetFuncall()
	if !ok {
		t.Errorf("expected false branch to be a funcall; got %#v\n", elseStmts[0])
	}
	if name, ok := callee.GetName(); !ok || name != "z" {
		t.Errorf("expected false branch to call `z`; got %#v\n", callee)
	}
	if len(actuals) != 0 {
		t.Errorf("expected false branch to pass 0 args; got %#v\n", actuals)
	}
}
