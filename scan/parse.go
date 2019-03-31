package scan

import "fmt"

type Parser struct {
	parsels map[string]*Token
	scope   *Scope

	token       *Token
	tokenNumber int
	tokens      []*Token
}

func NewParser() (p *Parser) {
	p = &Parser{}
	p.initializeSymbolTable()
	return
}

func (p *Parser) Parse(array_of_tokens []*Token) AST {
	p.tokens = array_of_tokens
	p.tokenNumber = 0
	p.newScope()

	p.advance()
	s := p.statements()
	p.skip("(end)")
	p.popScope()
	return s
}

func (p *Parser) popScope() {
	p.scope = p.scope.parent
}

func (p *Parser) newScope() {
	p.scope = &Scope{def: map[string]*Token{}, parent: p.scope}
}

func (p *Parser) findInScope(identifier string) *Token {
	if t := p.scope.find(identifier); t != nil {
		return t
	} else if tok, ok := p.parsels[identifier]; ok {
		return tok
	} else {
		t := p.parsels["(name)"]
		return t
	}
}

func (p *Parser) reserveInScope(t *Token) {
	p.scope.reserve(t)
}

func (p *Parser) symbol(name string, bp int) *Token {
	var s *Token
	var ok bool
	if s, ok = p.parsels[name]; ok {
		if bp >= s.parsel.TkLbp {
			s.parsel.TkLbp = bp
		}
	} else {
		s = &Token{
			TkValue: name,
			parsel: &Parsel{
				TkLbp: bp,
				TkNud: func(this *Token) AST {
					this.Error("Undefined")
					return NewNameAST(name)
				},
				TkLed: func(this *Token, left AST) AST {
					this.Error(fmt.Sprintf("Missing operator; left: %v", left))
					return nil // notreached
				}},
		}
		p.parsels[name] = s
	}
	return s
}

func (p *Parser) constant(s string, v string) *Token {
	x := p.symbol(s, -1)
	x.parsel.TkNud = func(this *Token) AST {
		//DEBUG fmt.Printf("constant %s: %v\n", v, this)
		p.reserveInScope(this)
		this.TkValue = p.parsels[this.TkValue].TkValue
		return NewAST1(s, NewNameAST(v))
	}
	x.TkValue = v
	return x
}

func (p *Parser) infix(id string, bp int, led BinaryDenotation) *Token {
	s := p.symbol(id, bp)
	s.parsel.TkLed = led
	if led == nil {
		s.parsel.TkLed = func(this *Token, left AST) AST {
			first := left
			//DEBUG fmt.Printf("infix after first: %v\n", this)
			second := p.expression(bp)
			//DEBUG fmt.Printf("infix after second: %v\n", this)
			return NewBinaryOperatorAST(id, first, second)
		}
	}
	return s
}

func (p *Parser) infixr(id string, bp int, led BinaryDenotation) *Token {
	s := p.symbol(id, bp)
	s.parsel.TkLed = led
	if led == nil {
		s.parsel.TkLed = func(this *Token, left AST) AST {
			first := left
			//DEBUG fmt.Printf("infixr after first: %v\n", this)
			second := p.expression(bp - 1)
			//DEBUG fmt.Printf("infixr after second: %v\n", this)
			return NewRightAssocBinaryOperatorAST(id, first, second)
		}
	}
	return s
}

func possibleLvalue(x *Token) bool {
	return x.TkValue == "." || x.TkValue == "[" || x.TkType == Name
}

func (p *Parser) assignment(id string) *Token {
	return p.infixr(id, 10, func(this *Token, left AST) AST {
		// if !possibleLvalue(left) {
		// 	left.Error("Bad lvalue.")
		// }
		// TODO: re-instate the code above once Token/AST changeover is complete
		first := left
		second := p.expression(9)
		//DEBUG fmt.Printf("assignment after second: %v\n", this)
		return &ASTimpl{
			kind:         id,
			first:        first,
			second:       second,
			isAssignment: true,
		}
	})
}

func (p *Parser) prefix(id string, nud UnaryDenotation) *Token {
	s := p.symbol(id, -1)
	s.parsel.TkNud = nud
	if nud == nil {
		s.parsel.TkNud = func(this *Token) AST {
			//DEBUG fmt.Printf("prefix before NdFirst: %v\n", this)
			p.reserveInScope(this)
			first := p.expression(70)
			//DEBUG fmt.Printf("prefix after first: %v\n", this)
			return NewAST1(id+" prefix", first)
		}
	}
	return s
}

func (p *Parser) stmt(id string, std UnaryDenotation) *Token {
	x := p.symbol(id, -1)
	x.parsel.TkStd = std
	return x
}

func (p *Parser) skip(id string) {
	if p.token.TkValue != id {
		p.token.Error(fmt.Sprintf("Expected '%s'.", id))
	}
	p.advance()
}

func (p *Parser) advance() {
	if p.tokenNumber >= len(p.tokens) {
		p.token = p.parsels["(end)"]
		return
	}
	t := p.tokens[p.tokenNumber]
	p.tokenNumber += 1
	v := t.TkValue
	a := t.TkType
	var o *Token
	var ok bool
	if a == Name {
		o = p.findInScope(v)
	} else if a == Punctuator {
		if o, ok = p.parsels[v]; !ok {
			t.Error("Unknown operator.")
		}
	} else if a == String || a == Fixnum || a == Flonum {
		o = p.parsels["(literal)"]
		a = Literal
	} else {
		t.Error("Unexpected token.")
	}
	p.token = &Token{}
	*p.token = *o
	p.token.TkLine = t.TkLine
	p.token.TkColumn = t.TkColumn
	p.token.TkValue = v
	p.token.TkType = a
	//fmt.Printf("next token: %v\n", p.token)
}

func (p *Parser) expression(rbp int) AST {
	t := p.token
	p.advance()
	if t == nil {
		panic("expression: initial token is nil")
	}
	if t.parsel.TkNud == nil {
		panic(fmt.Sprintf("expression: nil nud for %s", t))
		return nil // notreached
	}
	left := t.parsel.TkNud(t)
	for rbp < p.token.parsel.TkLbp {
		t = p.token
		p.advance()
		left = t.parsel.TkLed(t, left)
	}
	return left
}

func (p *Parser) statement() AST {
	n := p.token

	if n.parsel.TkStd != nil {
		p.advance()
		p.reserveInScope(n)
		return n.parsel.TkStd(n)
	}
	v := p.expression(0)
	if !v.IsAssignment() && v.IsFuncall() {
		v.Error(fmt.Sprintf("Bad expression statement (toplevel is %v).", v))
	}
	p.skip(";")
	return v
}

func (p *Parser) statements() AST {
	a := []AST{}
	var s AST
	for {
		if p.token.TkValue == "}" || p.token.TkValue == "(end)" { // '⌘' for "(end)"?
			break
		}
		s = p.statement()
		if s != nil {
			a = append(a, s)
		}
	}
	if len(a) == 0 {
		return nil
	} else if len(a) == 1 {
		return a[0]
	} else {
		return &ASTimpl{kind: "statements", list: a}
		//return &Token{
		//	NdId:    "statements",
		//	NdArity: listArity,
		//	NdList:  a,
		//}
	}
}

func (p *Parser) block() AST {
	t := p.token
	p.skip("{")
	return t.parsel.TkStd(t)
}

func thisLiteral(literal *Token) AST {
	//DEBUG fmt.Printf("thisLiteral: %v\n", literal)
	if literal.TkType != Literal {
		panic(fmt.Sprintf("expected literal, got %#v", literal))
	}
	return NewLiteralAST(literal.TkValue)
}

func (p *Parser) initializeSymbolTable() {
	p.parsels = map[string]*Token{}
	p.symbol("(end)", -1)
	p.symbol("(name)", -1)
	p.symbol(":", -1)
	p.symbol(";", -1)
	p.symbol(")", -1)
	p.symbol("]", -1)
	p.symbol("}", -1)
	p.symbol(",", -1)
	p.symbol("else", -1)

	p.constant("true", "#t")
	p.constant("false", "#f")
	p.constant("null", "null")
	p.constant("pi", "3.141592653589793")
	//p.constant("Object", {});
	//p.constant("Array", []);

	p.symbol("(literal)", -1).parsel.TkNud = thisLiteral

	p.symbol("this", -1).parsel.TkNud = func(this *Token) AST {
		//DEBUG fmt.Printf("this: %v\n", this)
		p.reserveInScope(this)
		return NewNameAST("this")
	}

	p.assignment("=")
	p.assignment("+=")
	p.assignment("-=")

	p.infix("?", 20, func(this *Token, left AST) AST {
		first := left
		second := p.expression(0)
		p.skip(":")
		third := p.expression(0)
		return NewAST3("?:", first, second, third)
	})

	p.infixr("&&", 30, nil)
	p.infixr("||", 30, nil)

	p.infixr("===", 40, nil)
	p.infixr("!==", 40, nil)
	p.infixr("<", 40, nil)
	p.infixr("<=", 40, nil)
	p.infixr(">", 40, nil)
	p.infixr(">=", 40, nil)

	p.infix("+", 50, nil)
	p.infix("-", 50, nil)

	p.infix("*", 60, nil)
	p.infix("/", 60, nil)

	p.infix(".", 80, func(this *Token, left AST) AST {
		first := left
		if p.token.TkType != Name {
			p.token.Error("Expected a property name.")
		}
		second := NewAST1("fieldname", NewNameAST(p.token.TkValue))
		p.advance()
		return NewAST2(".", first, second)
	})

	p.infix("[", 80, func(this *Token, left AST) AST {
		first := left
		second := p.expression(0)
		p.skip("]")
		return NewAST2("a[i]", first, second)
	})

	p.infix("(", 80, func(this *Token, left AST) AST {
		//if left.TkValue == "." || left.TkValue == "[" {
		//	this.NdArity = ternaryArity
		//	this.NdFirst = left.first
		//	this.NdSecond = left.second
		//} else {
		//  this.NdArity = binaryArity
		first := left
		//  this.NdFirst = first
		//if (left.NdArity != unaryArity || left.TkValue != "function") && //  'ƒ' for "function"?
		//	left.NdArity != nameArity && left.TkValue != "(" &&
		//	left.TkValue != "&&" && left.TkValue != "||" && // '∧' for "&&" and '∨' for "||"?
		//	left.TkValue != "?" {
		//	left.Error("Expected a variable name.")
		//} TODO: re-instate this check
		//} TODO: re-instate this binary/ternary distinction
		list := []AST{}
		if p.token.TkValue != ")" {
			for {
				list = append(list, p.expression(0))
				if p.token.TkValue != "," {
					break
				}
				p.skip(",")
			}
		}
		p.skip(")")
		return &ASTimpl{
			kind:      "funcall",
			first:     first,
			list:      list,
			isFuncall: true,
		}
	})

	p.prefix("!", nil)
	p.prefix("-", nil)
	p.prefix("typeof", nil)

	p.prefix("(", func(this *Token) AST {
		e := p.expression(0)
		p.skip(")")
		return e
	})

	p.prefix("function", func(this *Token) AST {
		// fmt.Printf("consumed `function`; current token is %v\n", p.token)
		formals := []AST{}
		p.newScope()
		var funcName string
		if p.token.TkType == Name {
			p.scope.define(p.token)
			funcName = p.token.TkValue
			// fmt.Printf("after `function`, consumed name %v\n", p.token)
			p.advance()
		}
		// fmt.Printf("after `function [name]`, looking for `('; current token is  %v\n", p.token)
		p.skip("(")
		if p.token.TkValue != ")" {
			for {
				if p.token.TkType != Name {
					p.token.Error("Expected a parameter name.")
				}
				p.scope.define(p.token)
				formals = append(formals, NewNameAST(p.token.TkValue))
				p.advance()
				if p.token.TkValue != "," {
					break
				}
				p.skip(",")
			}
		}
		p.skip(")")
		p.skip("{")
		second := p.statements()
		p.skip("}")
		p.popScope()
		return &ASTimpl{
			kind:   "defun",
			first:  NewNameAST(funcName),
			second: second,
			list:   formals,
		}
	})

	p.prefix("[", func(this *Token) AST {
		arrayContents := []AST{}
		if p.token.TkValue != "]" {
			for {
				arrayContents = append(arrayContents, p.expression(0))
				if p.token.TkValue != "," {
					break
				}
				p.skip(",")
			}
		}
		p.skip("]")
		return &ASTimpl{kind: "[...]", list: arrayContents}
	})

	//p.prefix("{", func(this *Token) AST {
	//	a := []AST{}
	//	var n *Token
	//	var v AST
	//	if p.token.TkValue != "}" {
	//		for {
	//			n = p.token
	//			if n.NdArity != nameArity && n.NdArity != literalArity {
	//				p.token.Error("Bad property name.")
	//			}
	//			p.advance()
	//			p.skip(":")
	//			v = p.expression(0)
	//			v.NdKey = n.TkValue
	//			a = append(a, v)
	//			if p.token.TkValue != "," {
	//				break
	//			}
	//			p.skip(",")
	//		}
	//	}
	//	p.skip("}")
	//	this.NdList = a
	//	this.NdArity = unaryArity
	//	return this
	//}) TODO: re-instate {...} constructor

	p.stmt("{", func(this *Token) AST {
		p.newScope()
		a := p.statements()
		p.skip("}")
		p.popScope()
		return a
	})

	p.stmt("let", func(this *Token) AST {
		a := []AST{}
		var n *Token
		for {
			n = p.token
			if n.TkType != Name {
				n.Error("Expected a new variable name.")
			}
			p.scope.define(n)
			p.advance()
			if p.token.TkValue == "=" {
				p.skip("=")
				first := NewNameAST(n.TkValue)
				second := p.expression(0)
				a = append(a, NewAST2("=", first, second))
			}
			if p.token.TkValue != "," {
				break
			}
			p.skip(",")
		}
		p.skip(";")
		if len(a) == 0 {
			return nil
		} else if len(a) == 1 {
			return a[0]
		} else {
			return &ASTimpl{kind: "let", list: a}
		}
	})

	p.stmt("if", func(this *Token) AST {
		p.skip("(")
		first := p.expression(0)
		p.skip(")")
		second := p.block()
		var third AST
		if p.token.TkValue == "else" { //  '𝕖' for "else"?
			p.reserveInScope(p.token)
			p.skip("else")
			if p.token.TkValue == "if" { // '𝕚' for "if"?
				third = p.statement()
			} else {
				third = p.block()
			}
		}
		return NewAST3("if", first, second, third)
	})

	p.stmt("return", func(this *Token) AST {
		var first AST
		if p.token.TkValue != ";" {
			first = p.expression(0)
		}
		p.skip(";")
		if p.token.TkValue != "}" {
			p.token.Error("Unreachable statement.")
		}
		return NewAST1("return", first)
	})

	p.stmt("break", func(this *Token) AST {
		p.skip(";")
		if p.token.TkValue != "}" {
			p.token.Error("Unreachable statement.")
		}
		return NewAST0("break")
	})

	p.stmt("while", func(this *Token) AST {
		p.skip("(")
		first := p.expression(0)
		p.skip(")")
		second := p.block()
		return NewAST2("while", first, second)
	})
}
