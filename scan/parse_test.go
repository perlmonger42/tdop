package scan

import (
	"testing"
)

func recoverFromPanic(t *testing.T) {
	if r := recover(); r != nil {
		t.Error(r)
	}
}

func parseString(source string) AST {
	var parser = NewParser()
	return parser.Parse(TokenizeString(source))
}

func TestLet(t *testing.T) {
	defer recoverFromPanic(t)
	source := "let answer = 42;"
	tree := parseString(source)
	// fmt.Printf("\n>>> %s\n%s", source, tree)
	if !tree.IsAssignment() {
		t.Errorf("expected assignment; got %#v\n", tree)
	}
	if varname, ok := tree.First().GetName(); !ok || varname != "answer" {
		t.Errorf("expected lhs to be `answer`; got %#v", tree.First())
	}
	if val, ok := tree.Second().GetLiteral(); !ok || val != "42" {
		t.Errorf("expected lhs to be `42`; got %#v", tree.Second())
	}
}

func TestExpression(t *testing.T) {
	defer recoverFromPanic(t)
	source := "let x;\nx = 1+2*3/(4-5);"
	assign := parseString(source)
	// fmt.Printf("\n>>> %s\n%s", source, assign)
	if !assign.IsAssignment() {
		t.Errorf("expected assignment; got %#v\n", assign)
	}
	if lhs, ok := assign.First().GetName(); !ok || lhs != "x" {
		t.Errorf("expected `x` on lhs of assignment; got %#v\n", assign.First())
	}
	plus := assign.Second()
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
	//source := "let f = function (){};"
	//tree := parseString(source)
	////fmt.Printf("\n>>> %s\n%s", source, tree)
	//if tree.TkType != Punctuator || tree.TkValue != "=" || tree.NdAssignment {
	//	t.Errorf("expected initialization; got %#v\n", tree)
	//}
	//if tree.NdFirst.TkType != Name {
	//	t.Errorf("expected Name `f`; got %#v\n", tree.NdFirst.TkType)
	//}
	//if tree.NdSecond.NdId != "function" {
	//	t.Errorf("expected `function` after `f`; got %#v\n", tree.NdSecond.TkType)
	//}
}

func TestAssignment(t *testing.T) {
	//defer recoverFromPanic(t)
	//source := "{\nlet answer;\nanswer = 42;\n}"
	//tree := parseString(source)
	////fmt.Printf("\n>>> %s\n%s", source, tree)
	//if tree.TkType != Punctuator || tree.TkValue != "=" || !tree.NdAssignment {
	//	t.Errorf("expected assignment; got %#v\n", tree)
	//}
}

func TestBlock(t *testing.T) {
	//	defer recoverFromPanic(t)
	//	source := `let token, advance, block = function () {
	//        let t = token;
	//        advance("{");
	//        return t.TkStd();
	//    };`
	//	tree := parseString(source)
	//	//fmt.Printf("\n>>> %s\n%s", source, tree)
	//	if tree.TkType != Punctuator || tree.TkValue != "=" || tree.NdAssignment {
	//		t.Errorf("expected assignment; got %#v\n", tree)
	//	}
}

func TestDoubleDeclaration(t *testing.T) {
	//	var tree AST
	//	defer func() {
	//		if tree != nil {
	//			t.Errorf("expected no tree returned; got %#v", tree)
	//		}
	//		if r := recover(); r != nil {
	//			s, ok := r.(string)
	//			if !ok || !strings.Contains(s, "Already defined") {
	//				t.Errorf("Expected `Already defined`; got %s", r)
	//			}
	//		} else {
	//			t.Errorf("expected a panic to occur")
	//		}
	//	}()
	//	source := `let x; let x; }`
	//	tree = parseString(source)
	//	///fmt.Printf("\n>>> %s\n%s", source, tree)
	//	//
	//	//
	//	//
}

func TestIf(t *testing.T) {
	//	defer recoverFromPanic(t)
	//	source := `let x, y, z; if (x) { y(); } else { z(); }`
	//	tree := parseString(source)
	//	///fmt.Printf("\n>>> %s\n%s", source, tree)
	//	//if tree.TkType != Name || tree.NdArity != statementArity || tree.TkValue != "if" {
	//	//	t.Errorf("expected if-stmt; got %#v\n", tree)
	//	//}
	//
	//	//a := tree.NdFirst
	//	//if a.TkType != Name || a.NdArity != nameArity || a.TkValue != "x" {
	//	//	t.Errorf("expected if's test to be `x`; got %#v", a)
	//	//}
	//
	//	//a = tree.NdSecond
	//	//if a.NdArity != binaryArity || a.TkValue != "(" {
	//	//	t.Errorf("expected if's consequent to be funcall; got %#v", a)
	//	//}
	//	//b := a.NdFirst
	//	//if b.NdArity != nameArity || b.TkValue != "y" {
	//	//	t.Errorf("expected if's consequent to be 'y()'; got %#v", a)
	//	//}
	//
	//	//a = tree.NdThird
	//	//if a.NdArity != binaryArity || a.TkValue != "(" {
	//	//	t.Errorf("expected if's alternative to be funcall; got
	//	%#v", a)
	//	//}
	//	//b = a.NdFirst
	//	//if b.NdArity != nameArity || b.TkValue != "z" {
	//	//	t.Errorf("expected if's alternative to be 'z()'; got %#v", a)
	//	//}
}
