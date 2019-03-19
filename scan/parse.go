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
		t.arity = nameArity
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
		if bp >= s.lbp {
			s.lbp = bp
		}
	} else {
		s = &Token{
			id:    name,
			Value: name,
			lbp:   bp,
			nud:   func(this *Token) *Token { this.Error("Undefined"); return this },
			led: func(this, left *Token) *Token {
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
	x.nud = func(this *Token) *Token {
		//DEBUG fmt.Printf("constant %s: %v\n", v, this)
		p.reserveInScope(this)
		this.Value = p.symbol_table[this.id].Value
		this.arity = literalArity
		return this
	}
	x.Value = v
	return x
}

func (p *Parser) infix(id string, bp int, led BinaryDenotation) *Token {
	s := p.symbol(id, bp)
	s.led = led
	if led == nil {
		s.led = func(this, left *Token) *Token {
			this.first = left
			//DEBUG fmt.Printf("infix after first: %v\n", this)
			this.second = p.expression(bp)
			this.arity = binaryArity
			//DEBUG fmt.Printf("infix after second: %v\n", this)
			return this
		}
	}
	return s
}

func (p *Parser) infixr(id string, bp int, led BinaryDenotation) *Token {
	s := p.symbol(id, bp)
	s.led = led
	if led == nil {
		s.led = func(this, left *Token) *Token {
			this.first = left
			//DEBUG fmt.Printf("infixr after first: %v\n", this)
			this.second = p.expression(bp - 1)
			this.arity = binaryArity
			//DEBUG fmt.Printf("infixr after second: %v\n", this)
			return this
		}
	}
	return s
}

func possibleLvalue(x *Token) bool {
	return x.id == "." || x.id == "[" || x.arity == nameArity
}

func (p *Parser) assignment(id string) *Token {
	return p.infixr(id, 10, func(this, left *Token) *Token {
		//DEBUG fmt.Printf("assignment after first: %v\n", this)
		if !possibleLvalue(left) {
			left.Error("Bad lvalue.")
		}
		this.first = left
		this.second = p.expression(9)
		this.assignment = true
		this.arity = binaryArity
		//DEBUG fmt.Printf("assignment after second: %v\n", this)
		return this
	})
}

func (p *Parser) prefix(id string, nud UnaryDenotation) *Token {
	s := p.symbol(id, -1)
	s.nud = nud
	if nud == nil {
		s.nud = func(this *Token) *Token {
			//DEBUG fmt.Printf("prefix before first: %v\n", this)
			p.reserveInScope(this)
			this.first = p.expression(70)
			this.arity = unaryArity
			//DEBUG fmt.Printf("prefix after first: %v\n", this)
			return this
		}
	}
	return s
}

func (p *Parser) stmt(id string, std UnaryDenotation) *Token {
	x := p.symbol(id, -1)
	x.std = std
	return x
}

func (p *Parser) skip(id string) {
	if p.token.id != id {
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
	v := t.Value
	a := t.Type
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
	p.token.Line = t.Line
	p.token.Column = t.Column
	p.token.Value = v
	p.token.Type = a
	//fmt.Printf("next token: %v\n", p.token)
}

func (p *Parser) expression(rbp int) *Token {
	t := p.token
	p.advance()
	if t == nil {
		panic("expression: initial token is nil")
	}
	if t.nud == nil {
		panic(fmt.Sprintf("expression: nil nud for %s", t))
		return t
	}
	left := t.nud(t)
	for rbp < p.token.lbp {
		t = p.token
		p.advance()
		left = t.led(t, left)
	}
	return left
}

func (p *Parser) statement() *Token {
	n := p.token

	if n.std != nil {
		p.advance()
		p.reserveInScope(n)
		return n.std(n)
	}
	v := p.expression(0)
	if !v.assignment && v.id != "(" {
		v.Error(fmt.Sprintf("Bad expression statement (toplevel is %v).", v))
	}
	p.skip(";")
	return v
}

func (p *Parser) statements() *Token {
	a := []*Token{}
	var s *Token
	for {
		if p.token.id == "}" || p.token.id == "(end)" {
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
			id:    "statements",
			arity: listArity,
			list:  a,
		}
	}
}

func (p *Parser) block() *Token {
	t := p.token
	p.skip("{")
	return t.std(t)
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

	p.symbol("(literal)", -1).nud = itself

	p.symbol("this", -1).nud = func(this *Token) *Token {
		//DEBUG fmt.Printf("this: %v\n", this)
		p.reserveInScope(this)
		this.arity = thisArity
		return this
	}

	p.assignment("=")
	p.assignment("+=")
	p.assignment("-=")

	p.infix("?", 20, func(this, left *Token) *Token {
		this.first = left
		this.second = p.expression(0)
		p.skip(":")
		this.third = p.expression(0)
		this.arity = ternaryArity
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
		this.first = left
		if p.token.arity != nameArity {
			p.token.Error("Expected a property name.")
		}
		p.token.arity = literalArity
		this.second = p.token
		this.arity = binaryArity
		p.advance()
		return this
	})

	p.infix("[", 80, func(this, left *Token) *Token {
		this.first = left
		this.second = p.expression(0)
		this.arity = binaryArity
		p.skip("]")
		return this
	})

	p.infix("(", 80, func(this, left *Token) *Token {
		if left.id == "." || left.id == "[" {
			this.arity = ternaryArity
			this.first = left.first
			this.second = left.second
		} else {
			this.arity = binaryArity
			this.first = left
			if (left.arity != unaryArity || left.id != "function") &&
				left.arity != nameArity && left.id != "(" &&
				left.id != "&&" && left.id != "||" &&
				left.id != "?" {
				left.Error("Expected a variable name.")
			}
		}
		this.list = []*Token{}
		if p.token.id != ")" {
			for {
				this.list = append(this.list, p.expression(0))
				if p.token.id != "," {
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
		if p.token.arity == nameArity {
			p.scope.define(p.token)
			this.name = p.token.Value
			// fmt.Printf("after `function`, consumed name %v\n", p.token)
			p.advance()
		}
		// fmt.Printf("after `function [name]`, looking for `('; current token is  %v\n", p.token)
		p.skip("(")
		if p.token.id != ")" {
			for {
				if p.token.arity != nameArity {
					p.token.Error("Expected a parameter name.")
				}
				p.scope.define(p.token)
				a = append(a, p.token)
				p.advance()
				if p.token.id != "," {
					break
				}
				p.skip(",")
			}
		}
		this.list = a
		p.skip(")")
		p.skip("{")
		this.second = p.statements()
		p.skip("}")
		this.arity = functionArity
		p.popScope()
		return this
	})

	p.prefix("[", func(this *Token) *Token {
		a := []*Token{}
		if p.token.id != "]" {
			for {
				a = append(a, p.expression(0))
				if p.token.id != "," {
					break
				}
				p.skip(",")
			}
		}
		p.skip("]")
		this.list = a
		this.arity = unaryArity
		return this
	})

	p.prefix("{", func(this *Token) *Token {
		a := []*Token{}
		var n, v *Token
		if p.token.id != "}" {
			for {
				n = p.token
				if n.arity != nameArity && n.arity != literalArity {
					p.token.Error("Bad property name.")
				}
				p.advance()
				p.skip(":")
				v = p.expression(0)
				v.key = n.Value
				a = append(a, v)
				if p.token.id != "," {
					break
				}
				p.skip(",")
			}
		}
		p.skip("}")
		this.list = a
		this.arity = unaryArity
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
			if n.arity != nameArity {
				n.Error("Expected a new variable name.")
			}
			p.scope.define(n)
			p.advance()
			if p.token.id == "=" {
				t = p.token
				p.skip("=")
				t.first = n
				t.second = p.expression(0)
				t.arity = binaryArity
				a = append(a, t)
			}
			if p.token.id != "," {
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
				id:    "let",
				arity: listArity,
				list:  a,
			}
		}
	})

	p.stmt("if", func(this *Token) *Token {
		p.skip("(")
		this.first = p.expression(0)
		p.skip(")")
		this.second = p.block()
		if p.token.id == "else" {
			p.reserveInScope(p.token)
			p.skip("else")
			if p.token.id == "if" {
				this.third = p.statement()
			} else {
				this.third = p.block()
			}
		} else {
			this.third = nil
		}
		this.arity = statementArity
		return this
	})

	p.stmt("return", func(this *Token) *Token {
		if p.token.id != ";" {
			this.first = p.expression(0)
		}
		p.skip(";")
		if p.token.id != "}" {
			p.token.Error("Unreachable statement.")
		}
		this.arity = statementArity
		return this
	})

	p.stmt("break", func(this *Token) *Token {
		p.skip(";")
		if p.token.id != "}" {
			p.token.Error("Unreachable statement.")
		}
		this.arity = statementArity
		return this
	})

	p.stmt("while", func(this *Token) *Token {
		p.skip("(")
		this.first = p.expression(0)
		p.skip(")")
		this.second = p.block()
		this.arity = statementArity
		return this
	})
}
