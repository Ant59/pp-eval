package ppeval

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Item
type item struct {
	typ itemType
	val string
}

type itemType int

// Items
const (
	itemError itemType = iota // 0
	itemEOL
	itemOpPlus
	itemOpMinus
	itemOpTimes
	itemOpDivide
	itemOpPow
	itemIf
	itemThen
	itemElse
	itemCmpOpEq // 10
	itemCmpOpNotEq
	itemCmpOpLtEq
	itemCmpOpGtEq
	itemCmpOpLt
	itemCmpOpGt
	itemLeftBrack
	itemRightBrack
	itemLeftQuote
	itemRightQuote
	itemStringLiteral // 20
	itemNumber
	itemText
	itemLength
	itemConst /* Y, N, y, n */
	itemFuncRight
	itemFuncLeft
	itemFuncHyp
	itemArgSplit
	itemLogiOr
	itemLogiAnd // 30
	itemOpRound
	itemOpRoundDown
	itemOpRoundUp
	itemShortIf
	itemShortElse
)

// Punctuation, operators, etc.
const (
	tokLeftBrack    = "("
	tokRightBrack   = ")"
	tokIf           = "if"
	tokThen         = "then"
	tokElse         = "else"
	tokShortIf      = "?"
	tokShortElse    = ":"
	tokCmpOpEq      = "="
	tokCmpOpNotEq   = "<>"
	tokCmpOpLtEq    = "<="
	tokCmpOpGtEq    = ">="
	tokCmpOpLt      = "<"
	tokCmpOpGt      = ">"
	tokFuncRight    = "right"
	tokFuncLeft     = "left"
	tokFuncHyp      = "hyp"
	tokFuncHypAlt   = "hypot"
	tokLogiOr       = "or"
	tokLogiAnd      = "and"
	tokLogiOrShort  = "||"
	tokLogiAndShort = "&&"
	eol             = '\n'
)

type lexer struct {
	input string
	start int
	pos   int
	width int
	items chan item
	state stateFn
}

func (i item) String() string {
	switch i.typ {
	case itemEOL:
		return "EOL"
	case itemError:
		return i.val
	}
	if len(i.val) > 20 {
		return fmt.Sprintf("VALUE: %.20q... TYPE: %v", i.val, i.typ)
	}
	return fmt.Sprintf("VALUE: %q TYPE: %v", i.val, i.typ)
}

type stateFn func(*lexer) stateFn

func lex(input string) *lexer {
	l := &lexer{
		input: input,
		state: lexExpression,
		items: make(chan item, 2),
	}
	return l
}

func lexExpression(l *lexer) stateFn {
exprLoop:
	for {
		// Conditionals
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokIf) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexIf
		}
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokThen) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexThen
		}
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokElse) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexElse
		}

		// Comparitors
		if strings.HasPrefix(l.input[l.pos:], tokCmpOpEq) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexCmpOpEq
		}
		if strings.HasPrefix(l.input[l.pos:], tokCmpOpNotEq) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexCmpOpNotEq
		}
		if strings.HasPrefix(l.input[l.pos:], tokCmpOpLtEq) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexCmpOpLtEq
		}
		if strings.HasPrefix(l.input[l.pos:], tokCmpOpGtEq) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexCmpOpGtEq
		}
		if strings.HasPrefix(l.input[l.pos:], tokCmpOpLt) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexCmpOpLt
		}
		if strings.HasPrefix(l.input[l.pos:], tokCmpOpGt) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexCmpOpGt
		}

		// Logic
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokLogiOr) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexLogiOr
		}
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokLogiAnd) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexLogiAnd
		}
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokLogiOrShort) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexLogiOrShort
		}
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokLogiAndShort) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexLogiAndShort
		}

		// Functions
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokFuncRight) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexFuncRight
		}
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokFuncLeft) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexFuncLeft
		}
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokFuncHypAlt) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexFuncHypAlt
		}
		if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), tokFuncHyp) {
			if l.pos > l.start {
				l.ignore()
			}
			return lexFuncHyp
		}

		switch r := l.next(); {
		case r == eol || r == '\n':
			break exprLoop
		case isSpace(r):
			l.ignore()
		case r == '"':
			l.ignore()
			return lexQuote
		case r == '(':
			l.stepBack()
			return lexLeftBrack
		case r == ')':
			l.stepBack()
			return lexRightBrack
		case r == '?':
			return lexShortIf
		case r == ':':
			return lexShortElse
		case r == '~':
			return lexOpRound
		case r == '@':
			return lexOpRoundDown
		case r == '#':
			return lexOpRoundUp
		case r == '+':
			return lexOpPlus
		case r == '-':
			return lexOpMinus
		case r == '*':
			return lexOpTimes
		case r == '/':
			return lexOpDivide
		case r == '^':
			return lexOpPow
		case r == ',':
			return lexArgSplit
		case '0' <= r && r <= '9':
			l.stepBack()
			return lexNumber
		case r == 'N' || r == 'Y' || r == 'n' || r == 'y':
			return lexConst
		}
	}

	if l.pos > l.start {
		l.ignore()
	}
	l.emit(itemEOL)
	return nil
}

func lexQuote(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], "\"") {
			if l.pos >= l.start {
				l.emit(itemStringLiteral)
			}
			l.pos++
			return lexExpression
		}
		if l.next() == eol {
			break
		}
	}
	return lexExpression
}

func lexLeftBrack(l *lexer) stateFn {
	l.pos += len(tokLeftBrack)
	l.emit(itemLeftBrack)
	return lexExpression
}

func lexRightBrack(l *lexer) stateFn {
	l.pos += len(tokRightBrack)
	l.emit(itemRightBrack)
	return lexExpression
}

func lexCmpOpEq(l *lexer) stateFn {
	l.pos += len(tokCmpOpEq)
	l.emit(itemCmpOpEq)
	return lexExpression
}

func lexCmpOpNotEq(l *lexer) stateFn {
	l.pos += len(tokCmpOpNotEq)
	l.emit(itemCmpOpNotEq)
	return lexExpression
}

func lexCmpOpLtEq(l *lexer) stateFn {
	l.pos += len(tokCmpOpLtEq)
	l.emit(itemCmpOpLtEq)
	return lexExpression
}

func lexCmpOpGtEq(l *lexer) stateFn {
	l.pos += len(tokCmpOpGtEq)
	l.emit(itemCmpOpGtEq)
	return lexExpression
}

func lexCmpOpLt(l *lexer) stateFn {
	l.pos += len(tokCmpOpLt)
	l.emit(itemCmpOpLt)
	return lexExpression
}

func lexCmpOpGt(l *lexer) stateFn {
	l.pos += len(tokCmpOpGt)
	l.emit(itemCmpOpGt)
	return lexExpression
}

func lexOpRound(l *lexer) stateFn {
	l.emit(itemOpRound)
	return lexExpression
}

func lexOpRoundDown(l *lexer) stateFn {
	l.emit(itemOpRoundDown)
	return lexExpression
}

func lexOpRoundUp(l *lexer) stateFn {
	l.emit(itemOpRoundUp)
	return lexExpression
}

func lexOpPlus(l *lexer) stateFn {
	l.emit(itemOpPlus)
	return lexExpression
}

func lexOpMinus(l *lexer) stateFn {
	l.emit(itemOpMinus)
	return lexExpression
}

func lexOpTimes(l *lexer) stateFn {
	l.emit(itemOpTimes)
	return lexExpression
}

func lexOpDivide(l *lexer) stateFn {
	l.emit(itemOpDivide)
	return lexExpression
}

func lexOpPow(l *lexer) stateFn {
	l.emit(itemOpPow)
	return lexExpression
}

func lexNumber(l *lexer) stateFn {
	// Optional leading sign.
	l.accept("+-")
	digits := "0123456789"
	l.acceptSeries(digits)
	if l.accept(".") {
		l.acceptSeries(digits)
	}
	p := l.pos
	if l.accept("eE") {
		l.accept("+-")
		if l.accept(digits) {
			l.accept(digits)
		} else {
			l.pos = p
		}
	}
	// If "mm" follows, it's a dimensional length
	if l.acceptString("mm") {
		l.emit(itemLength)
		return lexExpression
	}
	// Next thing mustn't be alphanumeric.
	if isNumeric(l.ahead()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(itemNumber)
	return lexExpression
}

func lexIf(l *lexer) stateFn {
	l.pos += len(tokIf)
	l.emit(itemIf)
	return lexExpression
}

func lexThen(l *lexer) stateFn {
	l.pos += len(tokThen)
	l.emit(itemThen)
	return lexExpression
}

func lexElse(l *lexer) stateFn {
	l.pos += len(tokElse)
	l.emit(itemElse)
	return lexExpression
}

func lexShortIf(l *lexer) stateFn {
	l.emit(itemShortIf)
	return lexExpression
}

func lexShortElse(l *lexer) stateFn {
	l.emit(itemShortElse)
	return lexExpression
}

func lexConst(l *lexer) stateFn {
	l.emit(itemConst)
	return lexExpression
}

func lexArgSplit(l *lexer) stateFn {
	l.emit(itemArgSplit)
	return lexExpression
}

func lexFuncRight(l *lexer) stateFn {
	l.pos += len(tokFuncRight)
	l.emit(itemFuncRight)
	return lexExpression
}

func lexFuncLeft(l *lexer) stateFn {
	l.pos += len(tokFuncLeft)
	l.emit(itemFuncLeft)
	return lexExpression
}

func lexFuncHyp(l *lexer) stateFn {
	l.pos += len(tokFuncHyp)
	l.emit(itemFuncHyp)
	return lexExpression
}

func lexFuncHypAlt(l *lexer) stateFn {
	l.pos += len(tokFuncHypAlt)
	l.emit(itemFuncHyp)
	return lexExpression
}

func lexLogiOr(l *lexer) stateFn {
	l.pos += len(tokLogiOr)
	l.emit(itemLogiOr)
	return lexExpression
}

func lexLogiAnd(l *lexer) stateFn {
	l.pos += len(tokLogiAnd)
	l.emit(itemLogiAnd)
	return lexExpression
}
func lexLogiOrShort(l *lexer) stateFn {
	l.pos += len(tokLogiOrShort)
	l.emit(itemLogiOr)
	return lexExpression
}

func lexLogiAndShort(l *lexer) stateFn {
	l.pos += len(tokLogiAndShort)
	l.emit(itemLogiAnd)
	return lexExpression
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eol
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) stepBack() {
	l.pos -= l.width
}

func (l *lexer) ahead() rune {
	r := l.next()
	l.stepBack()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.stepBack()
	return false
}

func (l *lexer) acceptSeries(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.stepBack()
}

func (l *lexer) acceptString(valid string) bool {
	if strings.HasPrefix(l.input[l.pos:], valid) {
		l.pos += len(valid)
		return true
	}
	return false
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isNumeric(r rune) bool {
	return r == '_' || unicode.IsDigit(r)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

func (l *lexer) nextItem() item {
	for {
		select {
		case i := <-l.items:
			return i
		default:
			l.state = l.state(l)
		}
	}
}
