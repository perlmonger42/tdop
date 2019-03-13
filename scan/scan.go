// This lexer is based on Rob Pike's Ivy scanner, found at
// https://github.com/robpike/ivy/blob/master/scan/scan.go

//go:generate stringer -type Type

package scan

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	// "github.com/perlmonger42/LiSP/config" //// config not yet supported
)

// Token represents a token or text string returned from the scanner.
type Token struct {
	Type Type   // The type of this item.
	Line int    // The line number on which this token appears
	Text string // The text of this item.
}

// Type identifies the type of lex items.
type Type int

const (
	EOF   Type = iota // zero value so closed channel delivers EOF
	Error             // error occurred; value is text of error

	Punctuator // ( ) { } [ ] ? . , : ; ~ * /
	Identifier // alphanumeric identifier

	// Scheme tokens
	LeftParen       // '('
	LeftBrack       // '['
	LeftBrace       // '{'
	Quote           // '\''
	QuasiQuote      // '`'
	Unquote         // ','
	UnquoteSplicing // ",@"
	False           // "#f"
	True            // "#t"
	Dot             // "."
	Ellipsis        // "..."
	Fixnum          // a number with no fractional component
	Flonum          // a number with a fractional component
	String          // quoted string (includes quotes)
	RightParen      // ')'
	RightBrack      // ']'
	RightBrace      // '}'
	CharLiteral     // '#\space', e.g.

	// Ivy tokens
	Assign         // '='
	Char           // printable ASCII character; grab bag for comma etc.
	GreaterOrEqual // '>='
	Number         // simple number
	Operator       // known operator
	Op             // "op", operator definition keyword
	Rational       // rational number like 2/3
	Semicolon      // ';'
	Space          // run of spaces separating
)

func (i Token) String() string {
	switch {
	case i.Type == EOF:
		return "<EOF>"
	case i.Type == Error:
		return "error: " + i.Text
	case len(i.Text) > 10:
		return fmt.Sprintf("%#v: %.10q...", i.Type, i.Text)
	}
	return fmt.Sprintf("%#v: %q", i.Type, i.Text)
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*Scanner) stateFn

// Scanner holds the state of the scanner.
type Scanner struct {
	// conf   *config.T  //// config not yet supported
	tokens chan Token // channel of scanned items
	r      io.ByteReader
	done   bool
	name   string // the name of the input; used only for error reports
	buf    []byte
	input  string  // the line of text being scanned.
	state  stateFn // the next lexing function to enter
	line   int     // line number in input
	pos    int     // current position in the input
	start  int     // start position of this item
	width  int     // width of last rune read from input

	lookahead bool  // Peek is usable
	Lookahead Token // The lookahead token
}

// loadLine reads the next line of input and stores it in (appends it to) the
// input.  (l.input may have data left over when we are called.)
// It strips carriage returns to make subsequent processing simpler.
func (l *Scanner) loadLine() {
	l.buf = l.buf[:0]
	for {
		c, err := l.r.ReadByte()
		if err != nil {
			l.done = true
			break
		}
		if c != '\r' {
			l.buf = append(l.buf, c)
		}
		if c == '\n' {
			break
		}
	}
	l.input = l.tokenText() + string(l.buf)
	l.pos -= l.start
	l.start = 0
}

//var quiet_next bool = false //DEBUG

// next returns the next rune in the input.
func (l *Scanner) next() (r rune) {
	if !l.done && int(l.pos) == len(l.input) {
		l.loadLine()
	}
	if int(l.pos) == len(l.input) {
		l.width = 0
		//if !quiet_next { //DEBUG
		//fmt.Printf("  next > eof\n") //DEBUG
		//} //DEBUG
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	//if !quiet_next { //DEBUG
	//fmt.Printf("  next > %q\n", string(r)) //DEBUG
	//} //DEBUG
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *Scanner) peek() rune {
	//quiet_next = true //DEBUG
	r := l.next()
	//quiet_next = false //DEBUG
	l.backup()
	//fmt.Printf("  peek > %q\n", string(r)) //DEBUG
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Scanner) backup() {
	//fmt.Printf("  bkup   %d\n", l.width)//DEBUG
	l.pos -= l.width
}

func (l *Scanner) incLine() {
	l.line++
}

func (l *Scanner) tokenText() string {
	return l.input[l.start:l.pos]
}

// passes an item back to the client.
func (l *Scanner) emit(t Type) {
	s := l.tokenText()
	//// config not yet supported
	//config := l.context.Config()
	//if config.Debug("tokens") {
	//	fmt.Fprintf(config.Output(), "%s:%d: emit %s\n", l.name, l.line, Token{t, l.line, s})
	//}
	//fmt.Printf("%s:%d: emit %s\n", l.name, l.line, Token{t, l.line, s})
	token := Token{t, l.line, s}
	//fmt.Printf("    emit %s:%d: emit %s\n", l.name, l.line, token) //DEBUG
	l.tokens <- token
	l.start = l.pos
	l.width = 0
}

// ignore skips over the pending input before this point.
func (l *Scanner) ignore() {
	//fmt.Printf("    ignore text\n") //DEBUG
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *Scanner) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Scanner) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// acceptRunOf consumes a run of runes from the valid set.
func (l *Scanner) acceptRunOf(isAcceptable func(rune) bool) {
	for isAcceptable(l.next()) {
	}
	l.backup()
}

// acceptLimitedRunOf consumes a run of up to maxCount runes from the valid set,
// but will not accept more than maxCount of input.
func (l *Scanner) acceptLimitedRunOf(isAcceptable func(rune) bool, maxCount int64) {
	for isAcceptable(l.next()) && maxCount > 0 {
		maxCount -= 1
	}
	l.backup()
}

// isLineSeparator reports whether the argument is a line separator.
// If r is '\r' and l.peek() is '\n', consumes the '\n' and returns true.
// Otherwise, returns true iff r is a Unicode line terminator.
//
// These are the Unicode line terminators, according to Wikipedia's [Newline
// article](https://en.wikipedia.org/wiki/Newline#Unicode):
//     LF:    Line Feed, U+000A
//     VT:    Vertical Tab, U+000B
//     FF:    Form Feed, U+000C
//     CR:    Carriage Return, U+000D
//     CR+LF: CR (U+000D) followed by LF (U+000A)
//     NEL:   Next Line, U+0085
//     LS:    Line Separator, U+2028
//     PS:    Paragraph Separator, U+2029
func (l *Scanner) isLineSeparator(r rune) bool {
	if r == '\r' && l.peek() == '\n' {
		l.next()
		return true
	}
	return r == '\n' || r == '\v' || r == '\f' || r == '\r' ||
		r == '\x85' || r == '\u2028' || r == '\u2029'
}

// error returns an error token and continues to scan.
func (l *Scanner) error(msg string) stateFn {
	return l.errorf("%s `%s`", msg, l.tokenText())
}

// errorf returns an error token and continues to scan.
func (l *Scanner) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- Token{Error, l.start, fmt.Sprintf(format, args...)}
	return lexAny
}

// New creates a new scanner for the input string.
func NewScanner( /* conf *config.T // config not yet supported, */ name string, r io.ByteReader) *Scanner {
	l := &Scanner{
		r:    r,
		name: name,
		//// conf:   conf, // config not yet supported
		line:   1,
		tokens: make(chan Token, 2), // We need a little room to save tokens.
		state:  lexAny,
	}
	return l
}

// Next returns the next token.
func (l *Scanner) Next() (result Token) {
	// We have up to one token of lookahead.
	if l.lookahead {
		l.lookahead = false
		return l.Lookahead
	}
	// The lexer is concurrent but we don't want it to run in parallel
	// with the rest of the interpreter, so we only run the state machine
	// when we need a token.
	for l.state != nil {
		select {
		case tok := <-l.tokens:
			return tok
		default:
			// Run the machine
			l.state = l.state(l)
		}
	}
	if l.tokens != nil {
		close(l.tokens)
		l.tokens = nil
	}
	return Token{EOF, l.pos, "<EOF>"}
}

func (l *Scanner) Peek() (result Token) {
	if l.lookahead {
		return l.Lookahead
	}
	l.Lookahead = l.Next()
	l.lookahead = true
	return l.Lookahead
}

// state functions

// lexLineComment scans a //-to-eol comment.
// The `//` comment marker has been consumed.
// The eol is either "\r\n" or a Unicode Line Terminator (see isLineSeparator).
//
// TODO: pass comments to parser?
func lexLineComment(l *Scanner) stateFn {
	//	fmt.Printf("lexLineComment\n")//DEBUG
	for r := l.next(); !l.isLineSeparator(r); r = l.next() {
		if r == eof {
			l.ignore()
			return lexAny
		}
	}
	l.incLine()
	l.ignore()
	return lexAny
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *Scanner) stateFn {
	// fmt.Printf("lexSpace\n")//DEBUG
	for unicode.IsSpace(l.peek()) {
		r := l.next()
		// fmt.Printf("lexSpace: consuming '%c'\n", r)//DEBUG
		if l.isLineSeparator(r) {
			l.line++
		}
	}
	l.ignore()
	return lexAny
}

// lexName scans an identifier. The leading letter has already been consumed.
//
// An identifier is composed of a leading alphabetic character, followed by
// zero or more alphanumeric characters.
//
// Inspired by Perl 6, the definitions of "alphabetic" and "alphanumeric"
// include appropriate Unicode characters.  Specifically:
// - Alphabetic characters are (1) the underscore (`_`), and (2) any character
//   with the Unicode General Category value `Letter (L)`.
// - Alphanumeric characters include all alphabetic characters, plus characters
//   with the Unicode General Category value `Number, Decimal Digit (Nd)`.
//
// For Perl6 details, see
// https://docs.perl6.org/language/syntax#Ordinary_identifiers and
// https://stackoverflow.com/questions/34689850/whats-allowed-in-a-perl-6-identifier#answer-34693397
func lexName(l *Scanner) stateFn {
	//	fmt.Printf("lexName\n")//DEBUG
	for isAlphanumeric(l.peek()) {
		l.next()
	}
	l.emit(Identifier)
	return lexAny
}

func isAlphabetic(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func isAlphanumeric(r rune) bool {
	return isAlphabetic(r) || isDigit(r)
}

// lexDigits scans /\d+(\.\d*)?/
// I.e., an integer or floating-point literal.
// The leading digit has already been consumed.
func lexDigits(l *Scanner) stateFn {
	//	fmt.Printf("lexDigits\n")//DEBUG
	l.acceptRunOf(isDigit)
	if l.next() == '.' {
		return lexDigitsDot
	}
	l.backup()
	return emitNumber(l, l.tokenText())
}

func lexDigitsDot(l *Scanner) stateFn {
	// one or more digits and a single '.' have been consumed
	if isDigit(l.peek()) {
		l.acceptRunOf(isDigit)
	}
	return emitNumber(l, l.tokenText())
}

func lexDot(l *Scanner) stateFn {
	// a single '.' has been consumed
	if isDigit(l.next()) {
		l.acceptRunOf(isDigit)
		text := "0" + l.tokenText() // strconv demands leading digit
		return emitNumber(l, text)
	}
	l.backup()
	l.emit(Dot)
	return lexAny
}

func emitNumber(l *Scanner, text string) stateFn {
	_, err := strconv.ParseInt(text, 0, 64)
	if err == nil {
		l.emit(Fixnum)
	} else if err.(*strconv.NumError).Err == strconv.ErrRange {
		return l.error("Bigints not supported")
	} else if _, err = strconv.ParseFloat(text, 64); err == nil {
		l.emit(Flonum)
	} else if err.(*strconv.NumError).Err == strconv.ErrRange {
		return l.error("Bigfloats not supported")
	} else {
		panic(fmt.Sprintf("unexpected strconv error on %q: %v", text, err))
	}
	return lexAny
}

// lexChar scans a character constant. The leading #\ is already
// scanned.
func lexChar(l *Scanner) stateFn {
	//	fmt.Printf("lexChar\n")//DEBUG
	switch r := l.next(); {
	case r == 'u' && isHexDigit(l.peek()):
		//fmt.Printf("4-digit unicode character\n")
		l.acceptLimitedRunOf(isHexDigit, 4)
	case r == 'U' && isHexDigit(l.peek()):
		//fmt.Printf("6-digit unicode character\n")
		l.acceptLimitedRunOf(isHexDigit, 6)
	case unicode.IsLetter(r) && unicode.IsLetter(l.peek()):
		//fmt.Printf("named character\n")
		l.acceptRunOf(unicode.IsLetter)
		if namedCharacter(l.input[l.start+2:l.pos]) < 0 {
			return l.error("unrecognized character name")
		}
	case isOctDigit(r):
		//fmt.Printf("octal character\n")
		l.acceptRunOf(isOctDigit)
		// there must be exactly 1 or exactly 3 octal digits
		runes := len([]rune(l.tokenText()))
		if runes != 3 && runes != 5 {
			return l.error("bad octal character syntax")
		}
	case unicode.IsLetter(r) != unicode.IsLetter(l.peek()):
		// This is either <letter followed by nonletter> or <nonletter
		// followed by letter>, so accept just the first character.
		//fmt.Printf("letter followed by nonletter, or vice versa\n")
	default:
		// <nonletter followed by nonletter>
		//fmt.Printf("nonletter followed by nonletter\n")
		if l.peek() != eof {
			return l.error("bad character syntax")
		}
	}
	l.emit(Char)
	return lexAny
}

func isHexDigit(r rune) bool {
	return '0' <= r && r <= '9' ||
		'a' <= r && r <= 'f' ||
		'A' <= r && r <= 'F'
}

func isOctDigit(r rune) bool {
	return '0' <= r && r <= '7'
}

func namedCharacter(s string) rune {
	switch s {
	case "nul", "null":
		return 0
	case "backspace":
	case "tab":
		return '\011'
	case "newline", "linefeed":
		return '\012'
	case "vtab":
		return '\013'
	case "page":
		return '\014'
	case "return":
		return '\015'
	case "space":
		return '\040'
	case "rubout":
		return '\177'
	}
	return -1
}

func CharLiteralToRune(s string) rune {
	runes := len([]rune(s))
	switch {
	case runes < 3 || s[0] != '#' || s[1] != '\\':
		// not even close to looking like a char literal
	case runes == 3:
		if r, size := utf8.DecodeRuneInString(s[2:]); size > 0 {
			return r
		}
	case s[2] == 'u' || s[2] == 'U': // runes > 3
		n, err := strconv.ParseInt(s[3:], 16, 64)
		if err == nil && n <= unicode.MaxRune {
			return rune(n)
		}
	default:
		if r := namedCharacter(s[2:]); r >= 0 {
			return r
		}
	}
	panic(fmt.Sprintf("invalid char literal %q", s))
}

// lexString scans a quoted string.
func lexString(l *Scanner) stateFn {
	//	fmt.Printf("lexString\n")//DEBUG
	for {
		switch r := l.next(); {
		case r == '\\':
			if r := l.next(); r != eof && !l.isLineSeparator(r) {
				if !strings.ContainsRune(`nr"\`, r) {
					l.errorf(`unrecognized escape sequence: \\%c`, r)
				}
				break // switch
			}
			fallthrough
		case r == eof || l.isLineSeparator(r):
			return l.errorf("unterminated quoted string")
			if r != eof {
				l.incLine()
			}
		case r == '"':
			l.emit(String)
			return lexAny
		}
	}
}

// lexAny scans non-space items.
func lexAny(l *Scanner) stateFn {
	// fmt.Printf("lexAny: switch on '%c'\n", l.peek()) //DEBUG
	switch r := l.next(); {
	case r == eof:
		return nil
	case l.isLineSeparator(r):
		l.incLine()
		l.ignore()
		return lexAny
	case unicode.IsSpace(r):
		return lexSpace

	case isAlphabetic(r):
		return lexName
	case isDigit(r):
		return lexDigits
	case r == '/':
		if l.peek() == '/' {
			l.next()
			return lexLineComment
		} else {
			l.emit(Punctuator)
		}
		return lexAny
	case r == '(':
		l.emit(LeftParen)
		return lexAny
	case r == ')':
		l.emit(RightParen)
		return lexAny
	case r == '[':
		l.emit(LeftBrack)
		return lexAny
	case r == ']':
		l.emit(RightBrack)
		return lexAny
	case r == '{':
		l.emit(LeftBrace)
		return lexAny
	case r == '}':
		l.emit(RightBrace)
		return lexAny
	case r == '"':
		return lexString
	case r == '.':
		return lexDot
	case r == ',':
		if l.peek() == '@' {
			l.next()
			l.emit(UnquoteSplicing)
		} else {
			l.emit(Unquote)
		}
		return lexAny
	case r == '\'':
		l.emit(Quote)
		return lexAny
	case r == '`':
		l.emit(QuasiQuote)
		return lexAny
	}
	//   anything else not listed above
	return lexAny
}
