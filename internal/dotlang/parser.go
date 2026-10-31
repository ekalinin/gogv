// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html
// Provenance: observed dot 16.1.0

package dotlang

// Parse reads the graphs of a DOT input. As the reference does, the rest of
// the line after the closing brace of a graph is ignored, and line numbers
// start again from 1 on the next line. On a syntax error, Parse returns the
// graphs read before the one that contains the error, with the error.
func Parse(src string) (*File, []string, error) {
	p := &parser{lex: NewLexer(src)}
	file := &File{}
	for {
		tok, err := p.peek()
		if err != nil {
			return file, p.lex.Warnings, err
		}
		if tok.Kind == EOF {
			return file, p.lex.Warnings, nil
		}
		g, err := p.graph()
		if err != nil {
			return file, p.lex.Warnings, err
		}
		file.Graphs = append(file.Graphs, g)
		p.lex.endGraph()
	}
}

type parser struct {
	lex    *Lexer
	tok    Token
	peeked bool
}

func (p *parser) peek() (Token, error) {
	if !p.peeked {
		tok, err := p.lex.Next()
		if err != nil {
			return tok, err
		}
		p.tok, p.peeked = tok, true
	}
	return p.tok, nil
}

func (p *parser) next() (Token, error) {
	tok, err := p.peek()
	p.peeked = false
	return tok, err
}

// errorAt returns the syntax error for an unexpected token.
func errorAt(tok Token) error {
	if tok.Kind == EOF {
		return &SyntaxError{Line: tok.Line, AtEOF: true}
	}
	return &SyntaxError{Line: tok.Line, Near: tok.Text}
}

func (p *parser) expect(k Kind) (Token, error) {
	tok, err := p.next()
	if err != nil {
		return tok, err
	}
	if tok.Kind != k {
		return tok, errorAt(tok)
	}
	return tok, nil
}

func isIDToken(k Kind) bool { return k == ID || k == QuotedID || k == HTMLID }

// id reads an ID; double-quoted strings may be joined with '+'.
func (p *parser) id() (IDValue, error) {
	tok, err := p.next()
	if err != nil {
		return IDValue{}, err
	}
	switch tok.Kind {
	case ID:
		return IDValue{Text: tok.Text}, nil
	case HTMLID:
		return IDValue{Text: tok.Text, HTML: true}, nil
	case QuotedID:
		text := tok.Text
		for {
			next, err := p.peek()
			if err != nil {
				return IDValue{}, err
			}
			if next.Kind != Plus {
				return IDValue{Text: text}, nil
			}
			p.next()
			part, err := p.expect(QuotedID)
			if err != nil {
				return IDValue{}, err
			}
			text += part.Text
		}
	}
	return IDValue{}, errorAt(tok)
}

func (p *parser) graph() (*GraphDecl, error) {
	g := &GraphDecl{}
	tok, err := p.next()
	if err != nil {
		return nil, err
	}
	if tok.Kind == KwStrict {
		g.Strict = true
		if tok, err = p.next(); err != nil {
			return nil, err
		}
	}
	switch tok.Kind {
	case KwGraph:
	case KwDigraph:
		g.Directed = true
	default:
		return nil, errorAt(tok)
	}
	if tok, err = p.peek(); err != nil {
		return nil, err
	}
	if isIDToken(tok.Kind) {
		id, err := p.id()
		if err != nil {
			return nil, err
		}
		g.ID = &id
	}
	if _, err := p.expect(LBrace); err != nil {
		return nil, err
	}
	if g.Stmts, err = p.stmts(g.Directed); err != nil {
		return nil, err
	}
	// The closing brace is consumed without looking past it, so that the
	// rest of its line can be discarded.
	if _, err := p.expect(RBrace); err != nil {
		return nil, err
	}
	return g, nil
}

// stmts reads statements up to, but not including, the closing brace.
func (p *parser) stmts(directed bool) ([]Stmt, error) {
	var stmts []Stmt
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == RBrace {
			return stmts, nil
		}
		s, err := p.stmt(directed)
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
		if tok, err = p.peek(); err != nil {
			return nil, err
		}
		if tok.Kind == Semi {
			p.next()
		}
	}
}

func (p *parser) stmt(directed bool) (Stmt, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	switch tok.Kind {
	case KwGraph, KwNode, KwEdge:
		p.next()
		target := map[Kind]AttrTarget{KwGraph: TargetGraph, KwNode: TargetNode, KwEdge: TargetEdge}[tok.Kind]
		attrs, err := p.attrLists(true)
		if err != nil {
			return nil, err
		}
		return &AttrStmt{Target: target, Attrs: attrs}, nil
	case KwSubgraph, LBrace:
		sub, err := p.subgraph(directed)
		if err != nil {
			return nil, err
		}
		return p.edgeOrSubgraph(EdgeOperand{Subgraph: sub}, directed)
	}
	if !isIDToken(tok.Kind) {
		p.next()
		return nil, errorAt(tok)
	}
	id, err := p.id()
	if err != nil {
		return nil, err
	}
	if tok, err = p.peek(); err != nil {
		return nil, err
	}
	if tok.Kind == Equal {
		p.next()
		value, err := p.id()
		if err != nil {
			return nil, err
		}
		return &AssignStmt{Attr: Attr{Key: id, Value: value}}, nil
	}
	nodes, err := p.nodeList(id)
	if err != nil {
		return nil, err
	}
	if tok, err = p.peek(); err != nil {
		return nil, err
	}
	if tok.Kind == DirectedOp || tok.Kind == UndirectOp {
		return p.edgeOrSubgraph(EdgeOperand{Nodes: nodes}, directed)
	}
	attrs, err := p.attrLists(false)
	if err != nil {
		return nil, err
	}
	return &NodeStmt{Nodes: nodes, Attrs: attrs}, nil
}

// nodeList reads a node ID, whose ID has been read, and the node IDs that
// follow it after commas.
func (p *parser) nodeList(id IDValue) ([]NodeID, error) {
	var nodes []NodeID
	for {
		node, err := p.port(id)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind != Comma {
			return nodes, nil
		}
		p.next()
		if tok, err = p.peek(); err != nil {
			return nil, err
		}
		if !isIDToken(tok.Kind) {
			p.next()
			return nil, errorAt(tok)
		}
		if id, err = p.id(); err != nil {
			return nil, err
		}
	}
}

// port reads the optional port of a node ID.
func (p *parser) port(id IDValue) (NodeID, error) {
	node := NodeID{ID: id}
	for range 2 {
		tok, err := p.peek()
		if err != nil {
			return node, err
		}
		if tok.Kind != Colon {
			return node, nil
		}
		p.next()
		part, err := p.id()
		if err != nil {
			return node, err
		}
		if node.Port != "" {
			node.Port += ":"
		}
		node.Port += part.Text
	}
	return node, nil
}

// edgeOrSubgraph continues a statement that starts with an operand: an edge
// statement when an edge operator follows, otherwise the subgraph itself.
func (p *parser) edgeOrSubgraph(first EdgeOperand, directed bool) (Stmt, error) {
	operands := []EdgeOperand{first}
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind != DirectedOp && tok.Kind != UndirectOp {
			break
		}
		p.next()
		if (tok.Kind == DirectedOp) != directed {
			return nil, errorAt(tok)
		}
		if tok, err = p.peek(); err != nil {
			return nil, err
		}
		switch {
		case tok.Kind == KwSubgraph || tok.Kind == LBrace:
			sub, err := p.subgraph(directed)
			if err != nil {
				return nil, err
			}
			operands = append(operands, EdgeOperand{Subgraph: sub})
		case isIDToken(tok.Kind):
			id, err := p.id()
			if err != nil {
				return nil, err
			}
			nodes, err := p.nodeList(id)
			if err != nil {
				return nil, err
			}
			operands = append(operands, EdgeOperand{Nodes: nodes})
		default:
			p.next()
			return nil, errorAt(tok)
		}
	}
	if len(operands) == 1 {
		return first.Subgraph, nil
	}
	attrs, err := p.attrLists(false)
	if err != nil {
		return nil, err
	}
	return &EdgeStmt{Operands: operands, Attrs: attrs}, nil
}

func (p *parser) subgraph(directed bool) (*Subgraph, error) {
	sub := &Subgraph{}
	tok, err := p.next()
	if err != nil {
		return nil, err
	}
	if tok.Kind == KwSubgraph {
		if tok, err = p.peek(); err != nil {
			return nil, err
		}
		if isIDToken(tok.Kind) {
			id, err := p.id()
			if err != nil {
				return nil, err
			}
			sub.ID = &id
		}
		if tok, err = p.next(); err != nil {
			return nil, err
		}
	}
	if tok.Kind != LBrace {
		return nil, errorAt(tok)
	}
	if sub.Stmts, err = p.stmts(directed); err != nil {
		return nil, err
	}
	if _, err := p.expect(RBrace); err != nil {
		return nil, err
	}
	return sub, nil
}

// attrLists reads zero or more attribute lists, or at least one when
// required.
func (p *parser) attrLists(required bool) ([]Attr, error) {
	var attrs []Attr
	for n := 0; ; n++ {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind != LBracket {
			if required && n == 0 {
				p.next()
				return nil, errorAt(tok)
			}
			return attrs, nil
		}
		p.next()
		for {
			if tok, err = p.peek(); err != nil {
				return nil, err
			}
			if tok.Kind == RBracket {
				p.next()
				break
			}
			if !isIDToken(tok.Kind) {
				p.next()
				return nil, errorAt(tok)
			}
			key, err := p.id()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(Equal); err != nil {
				return nil, err
			}
			value, err := p.id()
			if err != nil {
				return nil, err
			}
			attrs = append(attrs, Attr{Key: key, Value: value})
			if tok, err = p.peek(); err != nil {
				return nil, err
			}
			if tok.Kind == Comma || tok.Kind == Semi {
				p.next()
			}
		}
	}
}
