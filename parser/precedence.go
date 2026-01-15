// Package parser implements the Rugby parser.
package parser

import "github.com/nchapman/rugby/token"

// Precedence levels for expression parsing (Pratt parser).
// Higher values bind more tightly.
const (
	_ int = iota
	lowest
	rescuePrec  // call() rescue default (very low precedence)
	ternaryPrec // cond ? then : else (ternary conditional)
	rangePrec   // .., ... (ranges)
	nilCoalesce // ?? (nil coalescing)
	orPrec      // or
	andPrec     // and
	equals      // ==, !=
	lessGreater // <, >, <=, >=
	shift       // << (lower than arithmetic, like Ruby)
	sum         // +, -
	product     // *, /, %
	prefix      // -x, not x
	call        // f(x)
	bang        // call()! (error propagation)
	safeNav     // &. (safe navigation)
	member      // x.y (highest)
)

// precedences maps token types to their precedence levels.
var precedences = map[token.TokenType]int{
	token.RESCUE:           rescuePrec,
	token.QUESTION:         ternaryPrec, // ? (ternary conditional)
	token.DOTDOT:           rangePrec,
	token.TRIPLEDOT:        rangePrec,
	token.QUESTIONQUESTION: nilCoalesce, // ?? (nil coalescing)
	token.OR:               orPrec,
	token.AND:              andPrec,
	token.EQ:               equals,
	token.NE:               equals,
	token.MATCH:            equals, // =~ (regex match)
	token.NOTMATCH:         equals, // !~ (regex not match)
	token.LT:               lessGreater,
	token.GT:               lessGreater,
	token.LE:               lessGreater,
	token.GE:               lessGreater,
	token.PLUS:             sum,
	token.MINUS:            sum,
	token.SHOVELLEFT:       shift, // << (array append / channel send)
	token.STAR:             product,
	token.SLASH:            product,
	token.PERCENT:          product,
	token.LPAREN:           call,
	token.LBRACKET:         call,    // array indexing has same precedence as function calls
	token.BANG:             bang,    // postfix ! for error propagation
	token.AMPDOT:           safeNav, // &. (safe navigation)
	token.DOT:              member,
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
