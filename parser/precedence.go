// Package parser implements the Rugby parser.
package parser

import "github.com/nchapman/rugby/token"

// Precedence levels for expression parsing (Pratt parser).
// Higher values bind more tightly.
const (
	_ int = iota
	lowest
	rescuePrec  // call() rescue default (very low precedence)
	rangePrec   // .., ... (ranges)
	orPrec      // or
	andPrec     // and
	equals      // ==, !=
	lessGreater // <, >, <=, >=
	sum         // +, -
	product     // *, /, %
	prefix      // -x, not x
	call        // f(x)
	bang        // call()! (error propagation)
	member      // x.y (highest)
)

// precedences maps token types to their precedence levels.
var precedences = map[token.TokenType]int{
	token.RESCUE:    rescuePrec,
	token.DOTDOT:    rangePrec,
	token.TRIPLEDOT: rangePrec,
	token.OR:        orPrec,
	token.AND:       andPrec,
	token.EQ:        equals,
	token.NE:        equals,
	token.LT:        lessGreater,
	token.GT:        lessGreater,
	token.LE:        lessGreater,
	token.GE:        lessGreater,
	token.PLUS:      sum,
	token.MINUS:     sum,
	token.STAR:      product,
	token.SLASH:     product,
	token.PERCENT:   product,
	token.LPAREN:    call,
	token.LBRACKET:  call, // array indexing has same precedence as function calls
	token.BANG:      bang, // postfix ! for error propagation
	token.DOT:       member,
}

// peekPrecedence returns the precedence of the peek token.
func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return lowest
}

// curPrecedence returns the precedence of the current token.
func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return lowest
}
