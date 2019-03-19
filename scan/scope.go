package scan

import ()

type Scope struct {
	def    map[string]*Token
	parent *Scope
}

func (s *Scope) define(n *Token) {
	if t, ok := s.def[n.Value]; ok {
		if t.reserved {
			n.Error("Already reserved")
		} else {
			n.Error("Already defined")
		}
	}
	s.def[n.Value] = n
	n.reserved = false
	n.nud = itself
	n.led = nil
	n.std = nil
	n.lbp = 0
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
	if n.arity != nameArity || n.reserved {
		return
	}
	if t, ok := s.def[n.Value]; ok {
		if t.reserved {
			return
		}
		if t.arity == nameArity {
			n.Error("Already defined")
		}
	}
	s.def[n.Value] = n
	n.reserved = true
}
