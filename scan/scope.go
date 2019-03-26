package scan

import ()

type Scope struct {
	def    map[string]*Token
	parent *Scope
}

func (s *Scope) define(n *Token) {
	if t, ok := s.def[n.TkValue]; ok {
		if t.TkReserved {
			n.Error("Already reserved")
		} else {
			n.Error("Already defined")
		}
	}
	s.def[n.TkValue] = n
	n.TkReserved = false
	n.TkNud = itself
	n.TkLed = nil
	n.TkStd = nil
	n.TkLbp = 0
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
	if n.NdArity != nameArity || n.TkReserved {
		return
	}
	if t, ok := s.def[n.TkValue]; ok {
		if t.TkReserved {
			return
		}
		if t.NdArity == nameArity {
			n.Error("Already defined")
		}
	}
	s.def[n.TkValue] = n
	n.TkReserved = true
}
