package liquid

import "strconv"

// parser parses a Liquid template into an AST.
type parser struct {
	l        *lexer
	curToken token
}

func newParser(input string) *parser {
	p := &parser{l: newLexer(input)}
	p.nextToken()
	return p
}

func (p *parser) nextToken() {
	p.curToken = p.l.nextToken()
}

func (p *parser) parse() (*templateAST, error) {
	nodes, err := p.parseNodes(func() bool { return p.curToken.typ == tokenEOF })
	if err != nil {
		return nil, err
	}
	return &templateAST{nodes: nodes}, nil
}

func (p *parser) parseNodes(endCondition func() bool) ([]Node, error) {
	var nodes []Node

	for !endCondition() && p.curToken.typ != tokenEOF {
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func (p *parser) parseNode() (node Node, err error) {
	switch p.curToken.typ {
	case tokenEOF:
		return // returns nil node and nil error - signals end of parsing
	case tokenText:
		node = &TextNode{
			Text:   p.curToken.literal,
			Line:   p.curToken.line,
			Column: p.curToken.column,
		}
		p.nextToken()
		return node, nil
	case tokenOutputOpen, tokenOutputTrim:
		return p.parseOutput()
	case tokenTagOpen, tokenTagTrim:
		return p.parseTag()
	default:
		err = newParseError(p.curToken.line, p.curToken.column,
			"unexpected token: %v", p.curToken.literal)
		p.nextToken()
		return nil, err
	}
}

func (p *parser) parseOutput() (Node, error) {
	trimLeft := p.curToken.typ == tokenOutputTrim
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken()

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	trimRight := p.curToken.typ == tokenOutputTrimR
	if p.curToken.typ != tokenOutputClose && p.curToken.typ != tokenOutputTrimR {
		return nil, newParseError(p.curToken.line, p.curToken.column,
			"expected }}, got %v", p.curToken.literal)
	}
	p.nextToken()

	return &OutputNode{
		Expr:      expr,
		Line:      line,
		Column:    column,
		TrimLeft:  trimLeft,
		TrimRight: trimRight,
	}, nil
}

func (p *parser) parseTag() (Node, error) {
	p.nextToken() // consume {% or {%-

	switch p.curToken.typ {
	case tokenIf:
		return p.parseIfTag()
	case tokenUnless:
		return p.parseUnlessTag()
	case tokenCase:
		return p.parseCaseTag()
	case tokenFor:
		return p.parseForTag()
	case tokenBreak:
		return p.parseBreakTag()
	case tokenContinue:
		return p.parseContinueTag()
	case tokenAssignTag:
		return p.parseAssignTag()
	case tokenCapture:
		return p.parseCaptureTag()
	case tokenComment:
		return p.parseCommentTag()
	case tokenRaw:
		return p.parseRawTag()
	default:
		return nil, newParseError(p.curToken.line, p.curToken.column,
			"unknown tag: %s", p.curToken.literal)
	}
}

func (p *parser) parseIfTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'if'

	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expectTagClose()
	if err != nil {
		return nil, err
	}

	tag := &IfTag{
		Condition: condition,
		Line:      line,
		Column:    column,
	}

	// Parse then branch
	tag.ThenBranch, err = p.parseNodes(func() bool {
		return p.isTagKeyword(tokenElsif) || p.isTagKeyword(tokenElse) || p.isTagKeyword(tokenEndif)
	})
	if err != nil {
		return nil, err
	}

	// Parse elsif branches
	for p.isTagKeyword(tokenElsif) {
		p.nextToken() // consume {%
		p.nextToken() // consume elsif

		var elsifCond Expression
		elsifCond, err = p.parseExpression()
		if err != nil {
			return nil, err
		}

		err = p.expectTagClose()
		if err != nil {
			return nil, err
		}

		var elsifBody []Node
		elsifBody, err = p.parseNodes(func() bool {
			return p.isTagKeyword(tokenElsif) || p.isTagKeyword(tokenElse) || p.isTagKeyword(tokenEndif)
		})
		if err != nil {
			return nil, err
		}

		tag.ElsifBranches = append(tag.ElsifBranches, struct {
			Condition Expression
			Body      []Node
		}{elsifCond, elsifBody})
	}

	// Parse else branch
	if p.isTagKeyword(tokenElse) {
		p.nextToken() // consume {%
		p.nextToken() // consume else

		err = p.expectTagClose()
		if err != nil {
			return nil, err
		}

		tag.ElseBranch, err = p.parseNodes(func() bool {
			return p.isTagKeyword(tokenEndif)
		})
		if err != nil {
			return nil, err
		}
	}

	// Consume endif
	if !p.isTagKeyword(tokenEndif) {
		return nil, newParseError(p.curToken.line, p.curToken.column, "expected endif")
	}
	p.nextToken() // consume {%
	p.nextToken() // consume endif
	if err := p.expectTagClose(); err != nil {
		return nil, err
	}

	return tag, nil
}

func (p *parser) parseUnlessTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'unless'

	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expectTagClose()
	if err != nil {
		return nil, err
	}

	tag := &UnlessTag{
		Condition: condition,
		Line:      line,
		Column:    column,
	}

	// Parse body
	tag.Body, err = p.parseNodes(func() bool {
		return p.isTagKeyword(tokenElse) || p.isTagKeyword(tokenEndunless)
	})
	if err != nil {
		return nil, err
	}

	// Parse else branch
	if p.isTagKeyword(tokenElse) {
		p.nextToken() // consume {%
		p.nextToken() // consume else

		err = p.expectTagClose()
		if err != nil {
			return nil, err
		}

		tag.ElseBranch, err = p.parseNodes(func() bool {
			return p.isTagKeyword(tokenEndunless)
		})
		if err != nil {
			return nil, err
		}
	}

	// Consume endunless
	if !p.isTagKeyword(tokenEndunless) {
		return nil, newParseError(p.curToken.line, p.curToken.column, "expected endunless")
	}
	p.nextToken() // consume {%
	p.nextToken() // consume endunless
	err = p.expectTagClose()
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (p *parser) parseCaseTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'case'

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expectTagClose()
	if err != nil {
		return nil, err
	}

	tag := &CaseTag{
		Value:  value,
		Line:   line,
		Column: column,
	}

	// Skip any whitespace/text between case and first when
	for p.curToken.typ == tokenText {
		p.nextToken()
	}

	// Parse when clauses
	for p.isTagKeyword(tokenWhen) {
		p.nextToken() // consume {%
		p.nextToken() // consume when

		var values []Expression
		for {
			var val Expression
			val, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
			values = append(values, val)

			if p.curToken.typ != tokenComma {
				break
			}
			p.nextToken() // consume comma
		}

		err = p.expectTagClose()
		if err != nil {
			return nil, err
		}

		var body []Node
		body, err = p.parseNodes(func() bool {
			return p.isTagKeyword(tokenWhen) || p.isTagKeyword(tokenElse) || p.isTagKeyword(tokenEndcase)
		})
		if err != nil {
			return nil, err
		}

		tag.Whens = append(tag.Whens, WhenClause{Values: values, Body: body})
	}

	// Parse else
	if p.isTagKeyword(tokenElse) {
		p.nextToken() // consume {%
		p.nextToken() // consume else

		err = p.expectTagClose()
		if err != nil {
			return nil, err
		}

		tag.Else, err = p.parseNodes(func() bool {
			return p.isTagKeyword(tokenEndcase)
		})
		if err != nil {
			return nil, err
		}
	}

	// Consume endcase
	if !p.isTagKeyword(tokenEndcase) {
		return nil, newParseError(p.curToken.line, p.curToken.column, "expected endcase")
	}
	p.nextToken() // consume {%
	p.nextToken() // consume endcase
	if err := p.expectTagClose(); err != nil {
		return nil, err
	}

	return tag, nil
}

func (p *parser) parseForTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'for'

	if p.curToken.typ != tokenIdent {
		return nil, newParseError(p.curToken.line, p.curToken.column,
			"expected variable name, got %v", p.curToken.literal)
	}
	varName := p.curToken.literal
	p.nextToken()

	if p.curToken.typ != tokenIn {
		return nil, newParseError(p.curToken.line, p.curToken.column,
			"expected 'in', got %v", p.curToken.literal)
	}
	p.nextToken()

	// Check for range (start..end)
	var collection Expression
	if p.curToken.typ == tokenLParen {
		p.nextToken() // consume (
		start, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		if p.curToken.typ != tokenRange {
			return nil, newParseError(p.curToken.line, p.curToken.column,
				"expected '..' in range, got %v", p.curToken.literal)
		}
		p.nextToken() // consume ..

		end, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		if p.curToken.typ != tokenRParen {
			return nil, newParseError(p.curToken.line, p.curToken.column,
				"expected ')', got %v", p.curToken.literal)
		}
		p.nextToken() // consume )

		collection = &RangeExpr{
			Start:  start,
			End:    end,
			Line:   line,
			Column: column,
		}
	} else {
		var err error
		collection, err = p.parsePrimary()
		if err != nil {
			return nil, err
		}
	}

	tag := &ForTag{
		Variable:   varName,
		Collection: collection,
		Line:       line,
		Column:     column,
	}

	// Parse optional parameters: limit, offset, reversed
	for {
		switch p.curToken.typ {
		case tokenLimit:
			p.nextToken() // consume 'limit'
			if p.curToken.typ != tokenColon {
				return nil, newParseError(p.curToken.line, p.curToken.column,
					"expected ':' after limit")
			}
			p.nextToken() // consume ':'
			limit, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			tag.Limit = limit
		case tokenOffset:
			p.nextToken() // consume 'offset'
			if p.curToken.typ != tokenColon {
				return nil, newParseError(p.curToken.line, p.curToken.column,
					"expected ':' after offset")
			}
			p.nextToken() // consume ':'
			offset, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			tag.Offset = offset
		case tokenReversed:
			p.nextToken() // consume 'reversed'
			tag.Reversed = true
		default:
			goto doneParams
		}
	}
doneParams:

	if err := p.expectTagClose(); err != nil {
		return nil, err
	}

	var err error
	tag.Body, err = p.parseNodes(func() bool {
		return p.isTagKeyword(tokenElse) || p.isTagKeyword(tokenEndfor)
	})
	if err != nil {
		return nil, err
	}

	// Parse else branch (for empty collection)
	if p.isTagKeyword(tokenElse) {
		p.nextToken() // consume {%
		p.nextToken() // consume else

		err = p.expectTagClose()
		if err != nil {
			return nil, err
		}

		tag.ElseBody, err = p.parseNodes(func() bool {
			return p.isTagKeyword(tokenEndfor)
		})
		if err != nil {
			return nil, err
		}
	}

	// Consume endfor
	if !p.isTagKeyword(tokenEndfor) {
		return nil, newParseError(p.curToken.line, p.curToken.column, "expected endfor")
	}
	p.nextToken() // consume {%
	p.nextToken() // consume endfor
	err = p.expectTagClose()
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (p *parser) parseBreakTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'break'

	if err := p.expectTagClose(); err != nil {
		return nil, err
	}

	return &BreakTag{Line: line, Column: column}, nil
}

func (p *parser) parseContinueTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'continue'

	if err := p.expectTagClose(); err != nil {
		return nil, err
	}

	return &ContinueTag{Line: line, Column: column}, nil
}

func (p *parser) parseAssignTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'assign'

	if p.curToken.typ != tokenIdent {
		return nil, newParseError(p.curToken.line, p.curToken.column,
			"expected variable name, got %v", p.curToken.literal)
	}
	varName := p.curToken.literal
	p.nextToken()

	if p.curToken.typ != tokenAssign {
		return nil, newParseError(p.curToken.line, p.curToken.column,
			"expected '=', got %v", p.curToken.literal)
	}
	p.nextToken()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectTagClose(); err != nil {
		return nil, err
	}

	return &AssignTag{
		Variable: varName,
		Value:    value,
		Line:     line,
		Column:   column,
	}, nil
}

func (p *parser) parseCaptureTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'capture'

	if p.curToken.typ != tokenIdent {
		return nil, newParseError(p.curToken.line, p.curToken.column,
			"expected variable name, got %v", p.curToken.literal)
	}
	varName := p.curToken.literal
	p.nextToken()

	if err := p.expectTagClose(); err != nil {
		return nil, err
	}

	body, err := p.parseNodes(func() bool {
		return p.isTagKeyword(tokenEndcapture)
	})
	if err != nil {
		return nil, err
	}

	// Consume endcapture
	if !p.isTagKeyword(tokenEndcapture) {
		return nil, newParseError(p.curToken.line, p.curToken.column, "expected endcapture")
	}
	p.nextToken() // consume {%
	p.nextToken() // consume endcapture
	if err := p.expectTagClose(); err != nil {
		return nil, err
	}

	return &CaptureTag{
		Variable: varName,
		Body:     body,
		Line:     line,
		Column:   column,
	}, nil
}

func (p *parser) parseCommentTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'comment'

	if err := p.consumeTagClose(); err != nil {
		return nil, err
	}

	// Set lexer to text mode and scan to endcomment
	p.l.mode = modeText
	content, _, _ := p.l.scanCommentBlock()
	p.nextToken() // refresh cur token

	return &CommentTag{
		Content: content,
		Line:    line,
		Column:  column,
	}, nil
}

func (p *parser) parseRawTag() (Node, error) {
	line := p.curToken.line
	column := p.curToken.column
	p.nextToken() // consume 'raw'

	if err := p.consumeTagClose(); err != nil {
		return nil, err
	}

	// Set lexer to text mode and scan to endraw
	p.l.mode = modeText
	content, _, _ := p.l.scanRawBlock()
	p.nextToken() // refresh cur token

	return &RawTag{
		Content: content,
		Line:    line,
		Column:  column,
	}, nil
}

// parseExpression parses an expression with filters.
func (p *parser) parseExpression() (Expression, error) {
	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	// Parse filter chain
	for p.curToken.typ == tokenPipe {
		line := p.curToken.line
		column := p.curToken.column
		p.nextToken() // consume |

		if p.curToken.typ != tokenIdent {
			return nil, newParseError(p.curToken.line, p.curToken.column,
				"expected filter name, got %v", p.curToken.literal)
		}
		filterName := p.curToken.literal
		p.nextToken()

		var args []Expression
		if p.curToken.typ == tokenColon {
			p.nextToken() // consume :
			for {
				arg, err := p.parseOr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.curToken.typ != tokenComma {
					break
				}
				p.nextToken() // consume comma
			}
		}

		expr = &FilterExpr{
			Input:  expr,
			Name:   filterName,
			Args:   args,
			Line:   line,
			Column: column,
		}
	}

	return expr, nil
}

// parseOr parses "or" expressions.
func (p *parser) parseOr() (Expression, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.curToken.typ == tokenOr {
		line := p.curToken.line
		column := p.curToken.column
		p.nextToken()

		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Operator: "or", Right: right, Line: line, Column: column}
	}

	return left, nil
}

// parseAnd parses "and" expressions.
func (p *parser) parseAnd() (Expression, error) {
	left, err := p.parseContains()
	if err != nil {
		return nil, err
	}

	for p.curToken.typ == tokenAnd {
		line := p.curToken.line
		column := p.curToken.column
		p.nextToken()

		right, err := p.parseContains()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Operator: "and", Right: right, Line: line, Column: column}
	}

	return left, nil
}

// parseContains parses "contains" expressions.
func (p *parser) parseContains() (Expression, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	if p.curToken.typ == tokenContains {
		line := p.curToken.line
		column := p.curToken.column
		p.nextToken()

		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Operator: "contains", Right: right, Line: line, Column: column}
	}

	return left, nil
}

// parseComparison parses comparison expressions.
func (p *parser) parseComparison() (Expression, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		var op string
		switch p.curToken.typ {
		case tokenEq:
			op = "=="
		case tokenNe:
			op = "!="
		case tokenLt:
			op = "<"
		case tokenGt:
			op = ">"
		case tokenLe:
			op = "<="
		case tokenGe:
			op = ">="
		default:
			return left, nil
		}

		line := p.curToken.line
		column := p.curToken.column
		p.nextToken()

		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Operator: op, Right: right, Line: line, Column: column}
	}
}

// parsePrimary parses primary expressions with member access.
func (p *parser) parsePrimary() (Expression, error) {
	expr, err := p.parseAtom()
	if err != nil {
		return nil, err
	}

	for {
		switch p.curToken.typ {
		case tokenDot:
			line := p.curToken.line
			column := p.curToken.column
			p.nextToken()

			if p.curToken.typ != tokenIdent {
				return nil, newParseError(p.curToken.line, p.curToken.column,
					"expected property name, got %v", p.curToken.literal)
			}
			prop := p.curToken.literal
			p.nextToken()
			expr = &DotExpr{Object: expr, Property: prop, Line: line, Column: column}

		case tokenLBracket:
			line := p.curToken.line
			column := p.curToken.column
			p.nextToken()

			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			if p.curToken.typ != tokenRBracket {
				return nil, newParseError(p.curToken.line, p.curToken.column,
					"expected ']', got %v", p.curToken.literal)
			}
			p.nextToken()
			expr = &IndexExpr{Object: expr, Index: index, Line: line, Column: column}

		default:
			return expr, nil
		}
	}
}

// parseAtom parses atomic expressions (literals, identifiers, parenthesized expressions).
func (p *parser) parseAtom() (Expression, error) {
	line := p.curToken.line
	column := p.curToken.column

	switch p.curToken.typ {
	case tokenIdent:
		name := p.curToken.literal
		p.nextToken()
		return &IdentExpr{Name: name, Line: line, Column: column}, nil

	case tokenString:
		value := p.curToken.literal
		p.nextToken()
		return &LiteralExpr{Value: value, Line: line, Column: column}, nil

	case tokenInt:
		value := parseInt(p.curToken.literal)
		p.nextToken()
		return &LiteralExpr{Value: value, Line: line, Column: column}, nil

	case tokenFloat:
		value := parseFloat(p.curToken.literal)
		p.nextToken()
		return &LiteralExpr{Value: value, Line: line, Column: column}, nil

	case tokenTrue:
		p.nextToken()
		return &LiteralExpr{Value: true, Line: line, Column: column}, nil

	case tokenFalse:
		p.nextToken()
		return &LiteralExpr{Value: false, Line: line, Column: column}, nil

	case tokenNil:
		p.nextToken()
		return &LiteralExpr{Value: nil, Line: line, Column: column}, nil

	case tokenEmpty:
		p.nextToken()
		return &LiteralExpr{Value: emptyValue{}, Line: line, Column: column}, nil

	case tokenBlank:
		p.nextToken()
		return &LiteralExpr{Value: blankValue{}, Line: line, Column: column}, nil

	case tokenLParen:
		p.nextToken()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.curToken.typ != tokenRParen {
			return nil, newParseError(p.curToken.line, p.curToken.column,
				"expected ')', got %v", p.curToken.literal)
		}
		p.nextToken()
		return expr, nil

	case tokenMinus:
		// Unary minus for negative numbers
		p.nextToken()
		if p.curToken.typ == tokenInt {
			value := -parseInt(p.curToken.literal)
			p.nextToken()
			return &LiteralExpr{Value: value, Line: line, Column: column}, nil
		}
		if p.curToken.typ == tokenFloat {
			value := -parseFloat(p.curToken.literal)
			p.nextToken()
			return &LiteralExpr{Value: value, Line: line, Column: column}, nil
		}
		return nil, newParseError(p.curToken.line, p.curToken.column,
			"expected number after '-', got %v", p.curToken.literal)

	default:
		return nil, newParseError(line, column,
			"unexpected token in expression: %v", p.curToken.literal)
	}
}

// isTagKeyword checks if the current position is at a tag with the given keyword.
func (p *parser) isTagKeyword(keyword TokenType) bool {
	if p.curToken.typ != tokenTagOpen && p.curToken.typ != tokenTagTrim {
		return false
	}

	// Save state
	savedToken := p.curToken
	savedPos := p.l.pos
	savedReadPos := p.l.readPos
	savedCh := p.l.ch
	savedLine := p.l.line
	savedCol := p.l.column
	savedMode := p.l.mode

	p.nextToken()
	result := p.curToken.typ == keyword

	// Restore state
	p.curToken = savedToken
	p.l.pos = savedPos
	p.l.readPos = savedReadPos
	p.l.ch = savedCh
	p.l.line = savedLine
	p.l.column = savedCol
	p.l.mode = savedMode

	return result
}

// expectTagClose expects and consumes a tag close token (%} or -%}).
func (p *parser) expectTagClose() error {
	if p.curToken.typ != tokenTagClose && p.curToken.typ != tokenTagTrimR {
		return newParseError(p.curToken.line, p.curToken.column,
			"expected %%}, got %v", p.curToken.literal)
	}
	p.nextToken()
	return nil
}

// consumeTagClose consumes a tag close token without advancing to the next token.
// Used for raw and comment tags where we need to directly scan for the end tag.
func (p *parser) consumeTagClose() error {
	if p.curToken.typ != tokenTagClose && p.curToken.typ != tokenTagTrimR {
		return newParseError(p.curToken.line, p.curToken.column,
			"expected %%}, got %v", p.curToken.literal)
	}
	// Don't call nextToken() - we'll scan raw content directly
	return nil
}

// parseInt parses an integer literal.
func parseInt(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// parseFloat parses a float literal.
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// Special values for empty and blank comparisons.
type emptyValue struct{}
type blankValue struct{}

func (e emptyValue) String() string { return "empty" }
func (b blankValue) String() string { return "blank" }
