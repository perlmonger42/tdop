package scan

import ()

type Node struct {
	TkType     Type   // The type of this item.
	TkValue    string // The text of this item.
	TkLine     int    // The line number on which this token appears
	TkColumn   int    // The column number at which this token appears
	NdId       string
	NdArity    Type /*Arity*/
	reserved   bool
	nud        UnaryDenotation
	led        BinaryDenotation
	std        UnaryDenotation
	lbp        int
	assignment bool
	first      *Token
	second     *Token
	third      *Token
	list       []*Token
	name       string
	key        string
}

func NodeToToken(x *Node) (y *Token) {
	y = &Token{}
	return
}
