package scan

import "fmt"

type Scope struct {
	def    map[string]*Token
	parent *Scope
}

func thisName(name *Token) AST {
	if name.TkType != Name {
		panic(fmt.Sprintf("expected a Name, got %#v", name))
	}
	return NewNameAST(name.TkValue)
}

func (s *Scope) define(name *Token) {
	if name.TkType != Name {
		panic(fmt.Sprintf("expected a Name, got %#v", name))
	}
	if t, ok := s.def[name.TkValue]; ok {
		if t.TkReserved {
			name.Error("Already reserved")
		} else {
			name.Error("Already defined")
		}
	}
	s.def[name.TkValue] = name
	name.TkReserved = false
	name.parsel = &Parsel{
		TkNud: thisName,
		TkLed: nil,
		TkStd: nil,
		TkLbp: 0,
	}
}

func (s *Scope) find(name string) *Token {
	e := s
	for e != nil {
		if tok, ok := e.def[name]; ok {
			return tok
		}
		e = e.parent
	}
	return nil
}

func (s *Scope) reserve(n *Token) {
	if n.TkType != Name || n.TkReserved {
		return
	}
	if t, ok := s.def[n.TkValue]; ok {
		if t.TkReserved {
			return
		}
		if t.TkType == Name {
			n.Error("Already defined")
		}
	}
	s.def[n.TkValue] = n
	n.TkReserved = true
}
