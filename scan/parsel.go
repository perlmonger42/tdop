package scan

type UnaryDenotation func(this *Token) AST
type BinaryDenotation func(this *Token, left AST) AST

type Parsel struct {
	TkNud UnaryDenotation
	TkLed BinaryDenotation
	TkStd UnaryDenotation
	TkLbp int
}
