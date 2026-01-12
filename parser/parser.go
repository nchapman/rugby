// Package parser implements the Rugby parser.
package parser

import (
	"fmt"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/token"
)

// Parser parses Rugby source code into an AST.
type Parser struct {
	l            *lexer.Lexer
	curToken     token.Token
	peekToken    token.Token
	errors       []ParseError
	commentIndex int // current position in lexer.Comments slice
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	// Note: comments are collected during lexing (as tokens are skipped)
	// and accessed via p.l.CollectComments() after parsing completes.
	// The p.leadingComments/p.trailingComment methods consume from p.comments,
	// which will be populated when we need it. For now, we defer collecting.
	return p
}

// Errors returns parse errors as strings for backwards compatibility.
func (p *Parser) Errors() []string {
	result := make([]string, len(p.errors))
	for i, e := range p.errors {
		result[i] = e.String()
	}
	return result
}

// ParseErrors returns structured parse errors.
func (p *Parser) ParseErrors() []ParseError {
	return p.errors
}

// nextToken advances the parser to the next token.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs checks if the current token is of the given type.
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if the peek token is of the given type.
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// peekTokenAfterIs checks if the token immediately after peekToken matches the given type.
// Used for three-token lookahead patterns like @name = expr.
func (p *Parser) peekTokenAfterIs(t token.TokenType) bool {
	// Save lexer state and parser tokens
	lexerState := p.l.SaveState()
	savedCur := p.curToken
	savedPeek := p.peekToken

	// Advance to see the token after peekToken
	p.nextToken()
	p.nextToken()
	result := p.curToken.Type == t

	// Restore lexer state and parser tokens
	p.l.RestoreState(lexerState)
	p.curToken = savedCur
	p.peekToken = savedPeek
	return result
}

// skipNewlines skips over consecutive newline tokens.
func (p *Parser) skipNewlines() {
	for p.curTokenIs(token.NEWLINE) {
		p.nextToken()
	}
}

// leadingComments returns a group of consecutive comments that end immediately
// before the given line. Multi-line doc comments are collected as a group.
// Returns nil if no leading comments.
func (p *Parser) leadingComments(line int) *ast.CommentGroup {
	comments := p.l.Comments
	if p.commentIndex >= len(comments) {
		return nil
	}

	// Skip comments that are too early (free-floating)
	for p.commentIndex < len(comments) && comments[p.commentIndex].Line < line-1 {
		// Check if this could be the start of a consecutive group ending at line-1
		startIdx := p.commentIndex
		endIdx := startIdx

		// Find the end of consecutive comments
		for endIdx+1 < len(comments) && comments[endIdx+1].Line == comments[endIdx].Line+1 {
			endIdx++
		}

		// If this group ends at line-1, collect it
		if comments[endIdx].Line == line-1 {
			var result []*ast.Comment
			for i := startIdx; i <= endIdx; i++ {
				result = append(result, &ast.Comment{
					Text: comments[i].Text,
					Line: comments[i].Line,
				})
			}
			p.commentIndex = endIdx + 1
			return &ast.CommentGroup{List: result}
		}

		// Not the right group, skip it
		p.commentIndex = endIdx + 1
	}

	// Check for comments at exactly line-1
	if p.commentIndex < len(comments) && comments[p.commentIndex].Line == line-1 {
		// Collect consecutive comments ending at line-1
		startIdx := p.commentIndex
		endIdx := startIdx
		for endIdx+1 < len(comments) && comments[endIdx+1].Line == comments[endIdx].Line+1 &&
			comments[endIdx+1].Line <= line-1 {
			endIdx++
		}

		var result []*ast.Comment
		for i := startIdx; i <= endIdx; i++ {
			result = append(result, &ast.Comment{
				Text: comments[i].Text,
				Line: comments[i].Line,
			})
		}
		p.commentIndex = endIdx + 1
		return &ast.CommentGroup{List: result}
	}

	return nil
}

// ParseProgram parses the entire Rugby program.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	for !p.curTokenIs(token.EOF) {
		p.skipNewlines()
		if p.curTokenIs(token.EOF) {
			break
		}

		switch p.curToken.Type {
		case token.IMPORT:
			if imp := p.parseImport(); imp != nil {
				program.Imports = append(program.Imports, imp)
			}
		case token.PUB:
			pubLine := p.curToken.Line
			doc := p.leadingComments(pubLine)
			p.nextToken() // consume 'pub'
			switch p.curToken.Type {
			case token.DEF:
				if fn := p.parseFuncDeclWithDoc(doc); fn != nil {
					fn.Pub = true
					program.Declarations = append(program.Declarations, fn)
				}
			case token.CLASS:
				if cls := p.parseClassDeclWithDoc(doc); cls != nil {
					cls.Pub = true
					program.Declarations = append(program.Declarations, cls)
				}
			case token.INTERFACE:
				if iface := p.parseInterfaceDeclWithDoc(doc); iface != nil {
					iface.Pub = true
					program.Declarations = append(program.Declarations, iface)
				}
			case token.MODULE:
				if mod := p.parseModuleDeclWithDoc(doc); mod != nil {
					mod.Pub = true
					program.Declarations = append(program.Declarations, mod)
				}
			default:
				p.errorAt(p.curToken.Line, p.curToken.Column, "'pub' must be followed by 'def', 'class', 'interface', or 'module'")
				p.nextToken()
			}
		case token.DEF:
			if fn := p.parseFuncDecl(); fn != nil {
				program.Declarations = append(program.Declarations, fn)
			}
		case token.CLASS:
			if cls := p.parseClassDecl(); cls != nil {
				program.Declarations = append(program.Declarations, cls)

				// Validate pub def is only used in pub class
				if !cls.Pub {
					for _, method := range cls.Methods {
						if method.Pub {
							p.error(fmt.Sprintf("'pub def %s' not allowed in non-pub class '%s'", method.Name, cls.Name))
						}
					}
				}
			}
		case token.INTERFACE:
			if iface := p.parseInterfaceDecl(); iface != nil {
				program.Declarations = append(program.Declarations, iface)
			}
		case token.MODULE:
			if mod := p.parseModuleDecl(); mod != nil {
				program.Declarations = append(program.Declarations, mod)
			}
		case token.DESCRIBE:
			if desc := p.parseDescribeStmt(); desc != nil {
				program.Declarations = append(program.Declarations, desc)
			}
		case token.TEST:
			if test := p.parseTestStmt(); test != nil {
				program.Declarations = append(program.Declarations, test)
			}
		case token.TABLE:
			if table := p.parseTableStmt(); table != nil {
				program.Declarations = append(program.Declarations, table)
			}
		default:
			// Top-level statement
			if stmt := p.parseStatement(); stmt != nil {
				program.Declarations = append(program.Declarations, stmt)
			} else {
				p.errorAt(p.curToken.Line, p.curToken.Column, fmt.Sprintf("unexpected token %s at top level", p.curToken.Type))
				p.nextToken()
			}
		}
	}

	// Collect all comments after parsing is complete
	program.Comments = p.l.CollectComments()

	return program
}
