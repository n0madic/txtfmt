package parser

import (
	"unicode"
	"unicode/utf8"

	"github.com/n0madic/txtfmt/internal/ast"
)

type nodeKind int

const (
	nodeWord nodeKind = iota
	nodeSpace
	nodePunct
	nodeDash
	nodeEllipsis
	nodeQuoteMark
	nodeParenSpan
	nodeQuoteSpan
)

type node struct {
	kind     nodeKind
	text     string
	ch       rune
	dashKind ast.DashKind
	pos      ast.Pos

	level ast.QuoteLevel
	open  rune
	close rune

	children []node
}

type parenFrame struct {
	open  rune
	pos   ast.Pos
	nodes []node
}

type quoteFrame struct {
	quote rune
	pos   ast.Pos
	level ast.QuoteLevel
	nodes []node
}

func buildSpans(tokens []token) ([]ast.Inline, []ast.Diag) {
	nodes, diags := buildParenNodes(tokens)
	nodes, quoteDiags := buildQuoteNodes(nodes)
	diags = append(diags, quoteDiags...)
	return nodesToInlines(nodes), diags
}

func buildParenNodes(tokens []token) ([]node, []ast.Diag) {
	root := make([]node, 0, len(tokens))
	stack := make([]parenFrame, 0)
	diags := make([]ast.Diag, 0)

	appendCurrent := func(n node) {
		if len(stack) == 0 {
			root = append(root, n)
			return
		}
		i := len(stack) - 1
		stack[i].nodes = append(stack[i].nodes, n)
	}

	for idx, tok := range tokens {
		switch tok.kind {
		case tokenParenOpen:
			stack = append(stack, parenFrame{open: tok.ch, pos: tok.pos})
		case tokenParenClose:
			if len(stack) == 0 {
				if isEnumeratorClose(tokens, idx) {
					appendCurrent(node{kind: nodeWord, text: tok.text, pos: tok.pos})
					continue
				}
				diags = append(diags, ast.Diag{
					Pos:     tok.pos,
					Code:    "PAREN_MISMATCH",
					Message: "unexpected closing parenthesis",
				})
				appendCurrent(node{kind: nodeWord, text: tok.text, pos: tok.pos})
				continue
			}

			top := stack[len(stack)-1]
			if matchingClose(top.open) != tok.ch {
				diags = append(diags, ast.Diag{
					Pos:     tok.pos,
					Code:    "PAREN_MISMATCH",
					Message: "mismatched closing parenthesis",
				})
				appendCurrent(node{kind: nodeWord, text: tok.text, pos: tok.pos})
				continue
			}

			stack = stack[:len(stack)-1]
			span := node{
				kind:     nodeParenSpan,
				open:     top.open,
				close:    tok.ch,
				pos:      top.pos,
				children: top.nodes,
			}
			appendCurrent(span)
		default:
			appendCurrent(tokenToNode(tok))
		}
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		diags = append(diags, ast.Diag{
			Pos:     top.pos,
			Code:    "PAREN_UNCLOSED",
			Message: "unclosed parenthesis",
		})

		material := make([]node, 0, len(top.nodes)+1)
		material = append(material, node{kind: nodeWord, text: string(top.open), pos: top.pos})
		material = append(material, top.nodes...)

		if len(stack) == 0 {
			root = append(root, material...)
		} else {
			i := len(stack) - 1
			stack[i].nodes = append(stack[i].nodes, material...)
		}
	}

	return root, diags
}

func buildQuoteNodes(nodes []node) ([]node, []ast.Diag) {
	diags := make([]ast.Diag, 0)
	prepared := make([]node, 0, len(nodes))

	for _, n := range nodes {
		if n.kind == nodeParenSpan {
			inner, innerDiags := buildQuoteNodes(n.children)
			n.children = inner
			diags = append(diags, innerDiags...)
		}
		prepared = append(prepared, n)
	}

	root := make([]node, 0, len(prepared))
	stack := make([]quoteFrame, 0)

	appendCurrent := func(n node) {
		if len(stack) == 0 {
			root = append(root, n)
			return
		}
		i := len(stack) - 1
		stack[i].nodes = append(stack[i].nodes, n)
	}

	for i, n := range prepared {
		if n.kind != nodeQuoteMark {
			appendCurrent(n)
			continue
		}

		ch := n.ch
		switch {
		case isExplicitOpenQuote(ch):
			stack = append(stack, quoteFrame{quote: ch, pos: n.pos, level: quoteLevelForDepth(len(stack) + 1)})
		case isExplicitCloseQuote(ch):
			if len(stack) == 0 {
				diags = append(diags, ast.Diag{Pos: n.pos, Code: "QUOTE_MISMATCH", Message: "unexpected closing quote"})
				appendCurrent(node{kind: nodeWord, text: string(ch), pos: n.pos})
				continue
			}
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			appendCurrent(node{kind: nodeQuoteSpan, level: top.level, pos: top.pos, children: top.nodes})
		case isSymmetricQuote(ch):
			if shouldCloseSymmetric(prepared, i, len(stack)) {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				appendCurrent(node{kind: nodeQuoteSpan, level: top.level, pos: top.pos, children: top.nodes})
				continue
			}
			if shouldOpenSymmetric(prepared, i, len(stack)) {
				stack = append(stack, quoteFrame{quote: ch, pos: n.pos, level: quoteLevelForDepth(len(stack) + 1)})
				continue
			}
			appendCurrent(node{kind: nodeWord, text: string(ch), pos: n.pos})
		default:
			appendCurrent(node{kind: nodeWord, text: string(ch), pos: n.pos})
		}
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if shouldReportUnclosedQuote(top) {
			diags = append(diags, ast.Diag{
				Pos:     top.pos,
				Code:    "QUOTE_UNCLOSED",
				Message: "unclosed quote",
			})
		}

		material := make([]node, 0, len(top.nodes)+1)
		material = append(material, node{kind: nodeWord, text: string(top.quote), pos: top.pos})
		material = append(material, top.nodes...)

		if len(stack) == 0 {
			root = append(root, material...)
		} else {
			i := len(stack) - 1
			stack[i].nodes = append(stack[i].nodes, material...)
		}
	}

	return root, diags
}

func shouldReportUnclosedQuote(f quoteFrame) bool {
	// Symmetric quotes are ambiguous in noisy OCR text; keep literal text without warning.
	return !isSymmetricQuote(f.quote)
}

func tokenToNode(tok token) node {
	switch tok.kind {
	case tokenWord:
		return node{kind: nodeWord, text: tok.text, pos: tok.pos}
	case tokenSpace:
		return node{kind: nodeSpace, text: " ", pos: tok.pos}
	case tokenPunct:
		return node{kind: nodePunct, ch: tok.ch, text: tok.text, pos: tok.pos}
	case tokenDash:
		return node{kind: nodeDash, ch: tok.ch, text: tok.text, dashKind: tok.dashKind, pos: tok.pos}
	case tokenEllipsis:
		return node{kind: nodeEllipsis, text: tok.text, pos: tok.pos}
	case tokenQuote:
		return node{kind: nodeQuoteMark, ch: tok.ch, text: tok.text, pos: tok.pos}
	default:
		return node{kind: nodeWord, text: tok.text, pos: tok.pos}
	}
}

func nodesToInlines(nodes []node) []ast.Inline {
	out := make([]ast.Inline, 0, len(nodes))
	for _, n := range nodes {
		switch n.kind {
		case nodeWord:
			if n.text != "" {
				out = append(out, ast.Word{S: n.text})
			}
		case nodeSpace:
			out = append(out, ast.Space{Kind: ast.SpaceNormal})
		case nodePunct:
			out = append(out, ast.Punct{Ch: n.ch})
		case nodeDash:
			out = append(out, ast.Dash{Kind: n.dashKind})
		case nodeEllipsis:
			out = append(out, ast.Ellipsis{})
		case nodeParenSpan:
			out = append(out, ast.ParenSpan{Open: n.open, Close: n.close, In: nodesToInlines(n.children)})
		case nodeQuoteSpan:
			out = append(out, ast.QuoteSpan{Level: n.level, In: nodesToInlines(n.children)})
		case nodeQuoteMark:
			out = append(out, ast.Word{S: string(n.ch)})
		}
	}
	return out
}

func isEnumeratorClose(tokens []token, idx int) bool {
	if idx < 0 || idx >= len(tokens) || tokens[idx].kind != tokenParenClose || tokens[idx].ch != ')' {
		return false
	}

	prev, okPrev := prevNonSpaceToken(tokens, idx)
	if !okPrev || prev.kind != tokenWord {
		return false
	}
	if !isEnumerationWord(prev.text) {
		return false
	}

	next, okNext := nextNonSpaceToken(tokens, idx)
	if !okNext {
		return true
	}
	switch next.kind {
	case tokenWord, tokenQuote, tokenParenOpen:
		return idx+1 < len(tokens) && tokens[idx+1].kind == tokenSpace
	default:
		return true
	}
}

func prevNonSpaceToken(tokens []token, idx int) (token, bool) {
	for i := idx - 1; i >= 0; i-- {
		if tokens[i].kind == tokenSpace {
			continue
		}
		return tokens[i], true
	}
	return token{}, false
}

func nextNonSpaceToken(tokens []token, idx int) (token, bool) {
	for i := idx + 1; i < len(tokens); i++ {
		if tokens[i].kind == tokenSpace {
			continue
		}
		return tokens[i], true
	}
	return token{}, false
}

func isEnumerationWord(s string) bool {
	if s == "" {
		return false
	}
	allDigits := true
	allRoman := true
	for _, r := range s {
		if !unicode.IsDigit(r) {
			allDigits = false
		}
		if !isRomanRune(r) {
			allRoman = false
		}
	}
	return allDigits || allRoman
}

func isRomanRune(r rune) bool {
	switch unicode.ToUpper(r) {
	case 'I', 'V', 'X', 'L', 'C', 'D', 'M':
		return true
	default:
		return false
	}
}

func matchingClose(open rune) rune {
	switch open {
	case '(':
		return ')'
	case '[':
		return ']'
	default:
		return 0
	}
}

func quoteLevelForDepth(depth int) ast.QuoteLevel {
	if depth%2 == 1 {
		return ast.QuotePrimary
	}
	return ast.QuoteSecondary
}

func isExplicitOpenQuote(ch rune) bool {
	switch ch {
	case '«', '„', '‘':
		return true
	default:
		return false
	}
}

func isExplicitCloseQuote(ch rune) bool {
	switch ch {
	case '»', '”', '’':
		return true
	default:
		return false
	}
}

func isSymmetricQuote(ch rune) bool {
	switch ch {
	case '"', '“', '‟':
		return true
	default:
		return false
	}
}

func shouldCloseSymmetric(nodes []node, idx int, stackLen int) bool {
	if stackLen == 0 {
		return false
	}

	quote := nodes[idx]
	hasSpaceBefore := idx > 0 && nodes[idx-1].kind == nodeSpace
	hasSpaceAfter := idx+1 < len(nodes) && nodes[idx+1].kind == nodeSpace

	prev, okPrev := prevSignificant(nodes, idx)
	next, okNext := nextSignificant(nodes, idx)

	if !okPrev {
		return false
	}
	if !okNext {
		return true
	}

	if hasSpaceBefore && !hasSpaceAfter && nextSuggestsQuoteOpen(next) {
		return false
	}
	if !hasSpaceBefore && hasSpaceAfter {
		return true
	}

	if prevSuggestsQuoteClose(prev) {
		if hasSpaceBefore && nextSuggestsQuoteOpen(next) {
			return false
		}
		return true
	}

	if nextSuggestsQuoteClose(next) {
		return true
	}

	if next.pos.Line > quote.pos.Line && prevSuggestsQuoteClose(prev) {
		return true
	}

	return false
}

func shouldOpenSymmetric(nodes []node, idx int, stackLen int) bool {
	next, okNext := nextSignificant(nodes, idx)
	if !okNext || !nextSuggestsQuoteOpen(next) {
		return false
	}

	hasSpaceBefore := idx > 0 && nodes[idx-1].kind == nodeSpace
	hasSpaceAfter := idx+1 < len(nodes) && nodes[idx+1].kind == nodeSpace

	prev, okPrev := prevSignificant(nodes, idx)
	if !okPrev {
		return true
	}

	if !hasSpaceBefore && hasSpaceAfter {
		return false
	}

	if prevSuggestsQuoteOpen(prev) {
		return true
	}

	if stackLen > 0 && hasSpaceBefore && !hasSpaceAfter {
		return true
	}

	if prevSuggestsQuoteClose(prev) && !hasSpaceBefore {
		return false
	}

	if hasSpaceBefore {
		return true
	}

	return false
}

func prevSignificant(nodes []node, idx int) (node, bool) {
	for i := idx - 1; i >= 0; i-- {
		if nodes[i].kind == nodeSpace {
			continue
		}
		return nodes[i], true
	}
	return node{}, false
}

func nextSignificant(nodes []node, idx int) (node, bool) {
	for i := idx + 1; i < len(nodes); i++ {
		if nodes[i].kind == nodeSpace {
			continue
		}
		return nodes[i], true
	}
	return node{}, false
}

func prevSuggestsQuoteClose(n node) bool {
	switch n.kind {
	case nodeParenSpan, nodeQuoteSpan:
		return true
	case nodeEllipsis:
		return true
	case nodePunct:
		switch n.ch {
		case ',', '.', ';', ':', '!', '?':
			return true
		}
	case nodeWord:
		r, _ := utf8.DecodeLastRuneInString(n.text)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
		switch r {
		case ')', ']', '»', '”', '’':
			return true
		}
	}
	return false
}

func prevSuggestsQuoteOpen(n node) bool {
	switch n.kind {
	case nodeDash:
		return true
	case nodePunct:
		switch n.ch {
		case ':', ';', ',', '(', '[', '-', '—':
			return true
		}
	case nodeWord:
		r, _ := utf8.DecodeLastRuneInString(n.text)
		switch r {
		case ':', ';', ',', '(', '[', '-', '—':
			return true
		}
	}
	return false
}

func nextSuggestsQuoteClose(n node) bool {
	switch n.kind {
	case nodePunct, nodeEllipsis:
		return true
	case nodeWord:
		r, _ := utf8.DecodeRuneInString(n.text)
		switch r {
		case ')', ']', '»', '”', '’', ',', '.', ';', ':', '!', '?':
			return true
		}
	}
	return false
}

func nextSuggestsQuoteOpen(n node) bool {
	switch n.kind {
	case nodeParenSpan, nodeQuoteSpan:
		return true
	case nodeWord:
		r, _ := utf8.DecodeRuneInString(n.text)
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	}
	return false
}
