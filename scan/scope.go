package scan

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
	n.parsel = &Parsel{
		TkNud: itself,
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
