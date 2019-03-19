package scan

import (
	"strings"
	"testing"
)

func recoverFromPanic(t *testing.T) {
	if r := recover(); r != nil {
		t.Error(r)
	}
}

func parseString(source string) *Token {
	var parser = NewParser()
	return parser.Parse(TokenizeString(source))
}

func TestLet(t *testing.T) {
	defer recoverFromPanic(t)
	source := "let answer = 42;"
	tree := parseString(source)
	if tree.Type != Punctuator || tree.Value != "=" || tree.assignment {
		t.Errorf("expected assignment; got %v\n", tree)
	}
}

func TestExpression(t *testing.T) {
	defer recoverFromPanic(t)
	source := "let x;\nx = 1+2*3/(4-5);"
	tree := parseString(source)
	if tree.Type != Punctuator || tree.Value != "=" || !tree.assignment {
		t.Errorf("expected assignment; got %v\n", tree)
	}
}

func TestFuncDef(t *testing.T) {
	defer recoverFromPanic(t)
	source := "let f = function (){};"
	tree := parseString(source)
	if tree.Type != Punctuator || tree.Value != "=" || tree.assignment {
		t.Errorf("expected initialization; got %v\n", tree)
	}
	if tree.first.Type != Name {
		t.Errorf("expected Name `f`; got %v\n", tree.first.Type)
	}
	if tree.second.id != "function" {
		t.Errorf("expected `function` after `f`; got %v\n", tree.second.Type)
	}
}

func TestAssignment(t *testing.T) {
	defer recoverFromPanic(t)
	source := "{\nlet answer;\nanswer = 42;\n}"
	tree := parseString(source)
	if tree.Type != Punctuator || tree.Value != "=" || !tree.assignment {
		t.Errorf("expected assignment; got %v\n", tree)
	}
}

func TestBlock(t *testing.T) {
	defer recoverFromPanic(t)
	source := `let token, advance, block = function () {
        let t = token;
        advance("{");
        return t.std();
    };`
	tree := parseString(source)
	if tree.Type != Punctuator || tree.Value != "=" || tree.assignment {
		t.Errorf("expected assignment; got %v\n", tree)
	}

}

func TestDoubleDeclaration(t *testing.T) {
	var tree *Token
	defer func() {
		if tree != nil {
			t.Errorf("expected no tree returned; got %v", tree)
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
	tree = parseString(source)
}

func TestIf(t *testing.T) {
	defer recoverFromPanic(t)
	source := `let x, y, z; if (x) { y(); } else { z(); }`
	tree := parseString(source)
	if tree.Type != Name || tree.arity != statementArity || tree.Value != "if" {
		t.Errorf("expected if-stmt; got %v\n", tree)
	}

	a := tree.first
	if a.Type != Name || a.arity != nameArity || a.Value != "x" {
		t.Errorf("expected if's test to be `x`; got %v", a)
	}

	a = tree.second
	if a.arity != binaryArity || a.Value != "(" {
		t.Errorf("expected if's consequent to be funcall; got %v", a)
	}
	b := a.first
	if b.arity != nameArity || b.Value != "y" {
		t.Errorf("expected if's consequent to be 'y()'; got %v", a)
	}

	a = tree.third
	if a.arity != binaryArity || a.Value != "(" {
		t.Errorf("expected if's alternative to be funcall; got %v", a)
	}
	b = a.first
	if b.arity != nameArity || b.Value != "z" {
		t.Errorf("expected if's alternative to be 'z()'; got %v", a)
	}
}
