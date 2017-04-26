package ppeval

import (
	"log"
	"math"
)

type parser struct {
	input string
	lexer *lexer
	next  item
	prev  item
}

// Parse given string input
func Parse(input string) interface{} {
	p := &parser{
		input: input,
		lexer: lex(input),
	}

	p.next = p.lexer.nextItem()

	return p.statement()
}

func (p *parser) factor() float64 {
	if p.accept(itemNumber) {
		return parseFloat(p.prev.val)
	} else if p.accept(itemLength) {
		return parseFloat(p.prev.val[:len(p.prev.val)-2])
	} else if p.accept(itemConst) {
		if p.prev.val == "N" || p.prev.val == "n" {
			return 0
		} else if p.prev.val == "Y" || p.prev.val == "y" {
			return 1
		} else {
			log.Printf("Rejected! %s\n", p.next.String())
			panic("factor: not a constant")
		}
	} else if p.accept(itemLeftBrack) {
		v := p.statement().(float64)
		p.expect(itemRightBrack)
		return v
	} else {
		log.Printf("Rejected! %s\n", p.next.String())
		log.Printf("Input: %v\n", p.input)
		panic("factor: syntax error")
	}
}

func (p *parser) round() float64 {
	v := p.factor()
	for p.next.typ == itemOpRound || p.next.typ == itemOpRoundDown || p.next.typ == itemOpRoundUp {
		if p.next.typ == itemOpRound {
			p.getsym()
			w := p.factor()
			if math.Mod(v, w) >= w/2 {
				v = math.Ceil(v/w) * w
			} else {
				v = math.Floor(v/w) * w
			}
		} else if p.next.typ == itemOpRoundDown {
			p.getsym()
			w := p.factor()
			v = math.Floor(v/w) * w
		} else if p.next.typ == itemOpRoundUp {
			p.getsym()
			w := p.factor()
			v = math.Ceil(v/w) * w
		}
	}
	return v
}

func (p *parser) pow() float64 {
	v := p.round()
	for p.next.typ == itemOpPow {
		p.getsym()
		v = math.Pow(v, p.round())
	}
	return v
}

func (p *parser) term() float64 {
	v := p.pow()
	for p.next.typ == itemOpTimes || p.next.typ == itemOpDivide {
		if p.next.typ == itemOpTimes {
			p.getsym()
			v *= p.round()
		} else if p.next.typ == itemOpDivide {
			p.getsym()
			v /= p.round()
		}
	}
	return v
}

func (p *parser) expression() float64 {
	var v float64
	if p.next.typ == itemOpPlus {
		p.getsym()
		v = p.term()
	} else if p.next.typ == itemOpMinus {
		p.getsym()
		v = -p.term()
	} else {
		v = p.term()
	}
	for p.next.typ == itemOpPlus || p.next.typ == itemOpMinus {
		if p.next.typ == itemOpPlus {
			p.getsym()
			v += p.term()
		} else if p.next.typ == itemOpMinus {
			p.getsym()
			v -= p.term()
		}
	}
	return v
}

func (p *parser) string() (string, bool) {
	var v string
	if p.accept(itemFuncLeft) {
		p.expect(itemLeftBrack)
		p.expect(itemStringLiteral)
		s := p.prev.val
		p.expect(itemArgSplit)
		p.expect(itemNumber)

		if int64(len(s)) < parseInt(p.prev.val) {
			v = s
		} else {
			v = s[:parseInt(p.prev.val)]
		}

		p.expect(itemRightBrack)
		return v, true
	} else if p.accept(itemFuncRight) {
		p.expect(itemLeftBrack)
		p.expect(itemStringLiteral)
		s := p.prev.val
		p.expect(itemArgSplit)
		p.expect(itemNumber)

		if int64(len(s)) < parseInt(p.prev.val) {
			v = s
		} else {
			v = s[int64(len(s))-parseInt(p.prev.val):]
		}

		p.expect(itemRightBrack)
		return v, true
	} else if p.accept(itemStringLiteral) {
		return p.prev.val, true
	}
	return "", false
}

// Throws panic if not string
func (p *parser) stringOnly() string {
	t, isString := p.string()
	if isString {
		return t
	}
	panic("Expected string!")
}

func (p *parser) comparison() interface{} {
	s, isString := p.string()
	//log.Printf("'%v' is string? '%v'\n", s, isString)
	var v bool
	if isString {
		if p.accept(itemCmpOpEq) {
			v = s == p.stringOnly()
		} else if p.accept(itemCmpOpNotEq) {
			v = s != p.stringOnly()
		} else {
			return s //panic("condition: string with no comparitor")
		}
	} else {
		i := p.expression()
		if p.accept(itemCmpOpLtEq) {
			v = i <= p.expression()
		} else if p.accept(itemCmpOpGtEq) {
			v = i >= p.expression()
		} else if p.accept(itemCmpOpEq) {
			v = i == p.expression()
		} else if p.accept(itemCmpOpNotEq) {
			v = i != p.expression()
		} else if p.accept(itemCmpOpLt) {
			v = i < p.expression()
		} else if p.accept(itemCmpOpGt) {
			v = i > p.expression()
		} else {
			return i
		}
	}
	return v
}

func (p *parser) condition() interface{} {
	v := p.comparison()
	for p.next.typ == itemLogiOr || p.next.typ == itemLogiAnd {
		if p.accept(itemLogiOr) {
			p.accept(itemIf)
			v = p.comparison().(bool) || v.(bool)
		} else if p.accept(itemLogiAnd) {
			p.accept(itemIf)
			v = p.comparison().(bool) && v.(bool)
		}
	}
	return v
}

func (p *parser) statement() interface{} {
	if p.accept(itemFuncHyp) {
		p.expect(itemLeftBrack)
		a := p.expression()
		p.expect(itemArgSplit)

		v := math.Hypot(a, p.expression())

		p.expect(itemRightBrack)
		return v
	} else if p.accept(itemIf) {
		var funcElse interface{}
		condition := p.condition()
		p.expect(itemThen)
		then := p.statement()
		if p.accept(itemElse) {
			funcElse = p.statement()
		}
		if condition.(bool) {
			return then
		}
		return funcElse
	}
	v := p.condition()
	if p.accept(itemShortIf) {
		var funcElse interface{}
		then := p.statement()
		if p.accept(itemShortElse) {
			funcElse = p.statement()
		}
		if v.(bool) {
			return then
		}
		return funcElse
	}
	return v
}

// Get next symbol from lexer
func (p *parser) getsym() bool {
	if p.lexer.state != nil {
		p.next = p.lexer.nextItem()
		return true
	}
	return false
}

// Accepts the item and gets the next from lexer
func (p *parser) accept(typ itemType) bool {
	if typ == p.next.typ {
		//log.Printf("Accepted! %s\n", p.next.String())
		p.prev = p.next
		if p.getsym() {
			return true
		}
	}
	return false
}

// Requires the item and gets the next from the lexer or error
func (p *parser) expect(typ itemType) bool {
	if p.accept(typ) {
		return true
	}
	log.Printf("Rejected! %s\n", p.next.String())
	log.Printf("Expecting %v\n", typ)
	panic("Did not recieve expected item")
}

// Consume returns the current items and replaces with the next item from the lexer
func (p *parser) consume(typ itemType) item {
	if typ == p.next.typ {
		itemToReturn := p.next
		if p.lexer.state != nil {
			p.next = p.lexer.nextItem()
		}
		return itemToReturn
	}
	panic("Incorrect item type")
}
