package scan

import "fmt"

type Parser struct {
	// actually, this is more of a parse_table,
	// because its contents are tokens that direct the parsing
	symbol_table map[string]*Token
	scope        *Scope

	token       *Token
	tokenNumber int
	tokens      []*Token
}

func NewParser() (p *Parser) {
	p = &Parser{}
	p.initializeSymbolTable()
	return
}

func (p *Parser) Parse(array_of_tokens []*Token) *Token {
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

func (p *Parser) findInScope(name string) *Token {
	if t := p.scope.find(name); t != nil {
		return t
	} else if tok, ok := p.symbol_table[name]; ok {
		return tok
	} else {
		t := p.symbol_table["(name)"]
		t.NdArity = nameArity
		return t
	}
}

func (p *Parser) reserveInScope(t *Token) {
	p.scope.reserve(t)
}

func (p *Parser) symbol(name string, bp int) *Token {
	var s *Token
	var ok bool
	if s, ok = p.symbol_table[name]; ok {
		if bp >= s.TkLbp {
			s.TkLbp = bp
		}
	} else {
		s = &Token{
			NdId:    name,
			TkValue: name,
			TkLbp:   bp,
			TkNud:   func(this *Token) *Token { this.Error("Undefined"); return this },
			TkLed: func(this, left *Token) *Token {
				this.Error(fmt.Sprintf("Missing operator; left: %v", left))
				return this
			},
		}
		p.symbol_table[name] = s
	}
	return s
}

func (p *Parser) constant(s string, v string) *Token {
	x := p.symbol(s, -1)
	x.TkNud = func(this *Token) *Token {
		//DEBUG fmt.Printf("constant %s: %v\n", v, this)
		p.reserveInScope(this)
		this.TkValue = p.symbol_table[this.NdId].TkValue
		this.NdArity = literalArity
		return this
	}
	x.TkValue = v
	return x
}

func (p *Parser) infix(id string, bp int, led BinaryDenotation) *Token {
	s := p.symbol(id, bp)
	s.TkLed = led
	if led == nil {
		s.TkLed = func(this, left *Token) *Token {
			this.NdFirst = left
			//DEBUG fmt.Printf("infix after NdFirst: %v\n", this)
			this.NdSecond = p.expression(bp)
			this.NdArity = binaryArity
			//DEBUG fmt.Printf("infix after NdSecond: %v\n", this)
			return this
		}
	}
	return s
}

func (p *Parser) infixr(id string, bp int, led BinaryDenotation) *Token {
	s := p.symbol(id, bp)
	s.TkLed = led
	if led == nil {
		s.TkLed = func(this, left *Token) *Token {
			this.NdFirst = left
			//DEBUG fmt.Printf("infixr after NdFirst: %v\n", this)
			this.NdSecond = p.expression(bp - 1)
			this.NdArity = binaryArity
			//DEBUG fmt.Printf("infixr after NdSecond: %v\n", this)
			return this
		}
	}
	return s
}

func possibleLvalue(x *Token) bool {
	return x.NdId == "." || x.NdId == "[" || x.NdArity == nameArity
}

func (p *Parser) assignment(id string) *Token {
	return p.infixr(id, 10, func(this, left *Token) *Token {
		//DEBUG fmt.Printf("assignment after NdFirst: %v\n", this)
		if !possibleLvalue(left) {
			left.Error("Bad lvalue.")
		}
		this.NdFirst = left
		this.NdSecond = p.expression(9)
		this.NdAssignment = true
		this.NdArity = binaryArity
		//DEBUG fmt.Printf("assignment after NdSecond: %v\n", this)
		return this
	})
}

func (p *Parser) prefix(id string, nud UnaryDenotation) *Token {
	s := p.symbol(id, -1)
	s.TkNud = nud
	if nud == nil {
		s.TkNud = func(this *Token) *Token {
			//DEBUG fmt.Printf("prefix before NdFirst: %v\n", this)
			p.reserveInScope(this)
			this.NdFirst = p.expression(70)
			this.NdArity = unaryArity
			//DEBUG fmt.Printf("prefix after NdFirst: %v\n", this)
			return this
		}
	}
	return s
}

func (p *Parser) stmt(id string, std UnaryDenotation) *Token {
	x := p.symbol(id, -1)
	x.TkStd = std
	return x
}

func (p *Parser) skip(id string) {
	if p.token.NdId != id {
		p.token.Error(fmt.Sprintf("Expected '%s'.", id))
	}
	p.advance()
}

func (p *Parser) advance() {
	if p.tokenNumber >= len(p.tokens) {
		p.token = p.symbol_table["(end)"]
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
		if o, ok = p.symbol_table[v]; !ok {
			t.Error("Unknown operator.")
		}
	} else if a == String || a == Fixnum || a == Flonum {
		o = p.symbol_table["(literal)"]
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

func (p *Parser) expression(rbp int) *Token {
	t := p.token
	p.advance()
	if t == nil {
		panic("expression: initial token is nil")
	}
	if t.TkNud == nil {
		panic(fmt.Sprintf("expression: nil nud for %s", t))
		return t
	}
	left := t.TkNud(t)
	for rbp < p.token.TkLbp {
		t = p.token
		p.advance()
		left = t.TkLed(t, left)
	}
	return left
}

func (p *Parser) statement() *Token {
	n := p.token

	if n.TkStd != nil {
		p.advance()
		p.reserveInScope(n)
		return n.TkStd(n)
	}
	v := p.expression(0)
	if !v.NdAssignment && v.NdId != "(" {
		v.Error(fmt.Sprintf("Bad expression statement (toplevel is %v).", v))
	}
	p.skip(";")
	return v
}

func (p *Parser) statements() *Token {
	a := []*Token{}
	var s *Token
	for {
		if p.token.NdId == "}" || p.token.NdId == "(end)" { // '⌘' for "(end)"?
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
		return &Token{
			NdId:    "statements",
			NdArity: listArity,
			NdList:  a,
		}
	}
}

func (p *Parser) block() *Token {
	t := p.token
	p.skip("{")
	return t.TkStd(t)
}

func itself(this *Token) *Token {
	//DEBUG fmt.Printf("itself: %v\n", this)
	return this
}

func (p *Parser) initializeSymbolTable() {
	p.symbol_table = map[string]*Token{}
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

	p.symbol("(literal)", -1).TkNud = itself

	p.symbol("this", -1).TkNud = func(this *Token) *Token {
		//DEBUG fmt.Printf("this: %v\n", this)
		p.reserveInScope(this)
		this.NdArity = thisArity
		return this
	}

	p.assignment("=")
	p.assignment("+=")
	p.assignment("-=")

	p.infix("?", 20, func(this, left *Token) *Token {
		this.NdFirst = left
		this.NdSecond = p.expression(0)
		p.skip(":")
		this.NdThird = p.expression(0)
		this.NdArity = ternaryArity
		return this
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

	p.infix(".", 80, func(this, left *Token) *Token {
		this.NdFirst = left
		if p.token.NdArity != nameArity {
			p.token.Error("Expected a property name.")
		}
		p.token.NdArity = literalArity
		this.NdSecond = p.token
		this.NdArity = binaryArity
		p.advance()
		return this
	})

	p.infix("[", 80, func(this, left *Token) *Token {
		this.NdFirst = left
		this.NdSecond = p.expression(0)
		this.NdArity = binaryArity
		p.skip("]")
		return this
	})

	p.infix("(", 80, func(this, left *Token) *Token {
		if left.NdId == "." || left.NdId == "[" {
			this.NdArity = ternaryArity
			this.NdFirst = left.NdFirst
			this.NdSecond = left.NdSecond
		} else {
			this.NdArity = binaryArity
			this.NdFirst = left
			if (left.NdArity != unaryArity || left.NdId != "function") && //  'ƒ' for "function"?
				left.NdArity != nameArity && left.NdId != "(" &&
				left.NdId != "&&" && left.NdId != "||" && // '∧' for "&&" and '∨' for "||"?
				left.NdId != "?" {
				left.Error("Expected a variable name.")
			}
		}
		this.NdList = []*Token{}
		if p.token.NdId != ")" {
			for {
				this.NdList = append(this.NdList, p.expression(0))
				if p.token.NdId != "," {
					break
				}
				p.skip(",")
			}
		}
		p.skip(")")
		return this
	})

	p.prefix("!", nil)
	p.prefix("-", nil)
	p.prefix("typeof", nil)

	p.prefix("(", func(this *Token) *Token {
		e := p.expression(0)
		p.skip(")")
		return e
	})

	p.prefix("function", func(this *Token) *Token {
		// fmt.Printf("consumed `function`; current token is %v\n", p.token)
		a := []*Token{}
		p.newScope()
		if p.token.NdArity == nameArity {
			p.scope.define(p.token)
			this.NdName = p.token.TkValue
			// fmt.Printf("after `function`, consumed name %v\n", p.token)
			p.advance()
		}
		// fmt.Printf("after `function [name]`, looking for `('; current token is  %v\n", p.token)
		p.skip("(")
		if p.token.NdId != ")" {
			for {
				if p.token.NdArity != nameArity {
					p.token.Error("Expected a parameter name.")
				}
				p.scope.define(p.token)
				a = append(a, p.token)
				p.advance()
				if p.token.NdId != "," {
					break
				}
				p.skip(",")
			}
		}
		this.NdList = a
		p.skip(")")
		p.skip("{")
		this.NdSecond = p.statements()
		p.skip("}")
		this.NdArity = functionArity
		p.popScope()
		return this
	})

	p.prefix("[", func(this *Token) *Token {
		a := []*Token{}
		if p.token.NdId != "]" {
			for {
				a = append(a, p.expression(0))
				if p.token.NdId != "," {
					break
				}
				p.skip(",")
			}
		}
		p.skip("]")
		this.NdList = a
		this.NdArity = unaryArity
		return this
	})

	p.prefix("{", func(this *Token) *Token {
		a := []*Token{}
		var n, v *Token
		if p.token.NdId != "}" {
			for {
				n = p.token
				if n.NdArity != nameArity && n.NdArity != literalArity {
					p.token.Error("Bad property name.")
				}
				p.advance()
				p.skip(":")
				v = p.expression(0)
				v.NdKey = n.TkValue
				a = append(a, v)
				if p.token.NdId != "," {
					break
				}
				p.skip(",")
			}
		}
		p.skip("}")
		this.NdList = a
		this.NdArity = unaryArity
		return this
	})

	p.stmt("{", func(this *Token) *Token {
		p.newScope()
		a := p.statements()
		p.skip("}")
		p.popScope()
		return a
	})

	p.stmt("let", func(this *Token) *Token {
		a := []*Token{}
		var n, t *Token
		for {
			n = p.token
			if n.NdArity != nameArity {
				n.Error("Expected a new variable name.")
			}
			p.scope.define(n)
			p.advance()
			if p.token.NdId == "=" {
				t = p.token
				p.skip("=")
				t.NdFirst = n
				t.NdSecond = p.expression(0)
				t.NdArity = binaryArity
				a = append(a, t)
			}
			if p.token.NdId != "," {
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
			return &Token{
				NdId:    "let",
				NdArity: listArity,
				NdList:  a,
			}
		}
	})

	p.stmt("if", func(this *Token) *Token {
		p.skip("(")
		this.NdFirst = p.expression(0)
		p.skip(")")
		this.NdSecond = p.block()
		if p.token.NdId == "else" { //  '𝕖' for "else"?
			p.reserveInScope(p.token)
			p.skip("else")
			if p.token.NdId == "if" { // '𝕚' for "if"?
				this.NdThird = p.statement()
			} else {
				this.NdThird = p.block()
			}
		} else {
			this.NdThird = nil
		}
		this.NdArity = statementArity
		return this
	})

	p.stmt("return", func(this *Token) *Token {
		if p.token.NdId != ";" {
			this.NdFirst = p.expression(0)
		}
		p.skip(";")
		if p.token.NdId != "}" {
			p.token.Error("Unreachable statement.")
		}
		this.NdArity = statementArity
		return this
	})

	p.stmt("break", func(this *Token) *Token {
		p.skip(";")
		if p.token.NdId != "}" {
			p.token.Error("Unreachable statement.")
		}
		this.NdArity = statementArity
		return this
	})

	p.stmt("while", func(this *Token) *Token {
		p.skip("(")
		this.NdFirst = p.expression(0)
		p.skip(")")
		this.NdSecond = p.block()
		this.NdArity = statementArity
		return this
	})
}
