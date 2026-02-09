package parser

import (
	"unicode"

	"github.com/n0madic/txtfmt/internal/ast"
)

type tokenKind int

const (
	tokenWord tokenKind = iota
	tokenSpace
	tokenPunct
	tokenDash
	tokenEllipsis
	tokenQuote
	tokenParenOpen
	tokenParenClose
)

type token struct {
	kind     tokenKind
	text     string
	ch       rune
	dashKind ast.DashKind
	pos      ast.Pos
}

var punctRunes = map[rune]struct{}{
	',': {},
	'.': {},
	';': {},
	':': {},
	'!': {},
	'?': {},
}

var quoteRunes = map[rune]struct{}{
	'"': {},
	'«': {},
	'»': {},
	'„': {},
	'“': {},
	'”': {},
	'‟': {},
	'‘': {},
	'’': {},
}

func tokenizeInline(s string, line, startCol int) []token {
	runes := []rune(s)
	out := make([]token, 0, len(runes))

	i := 0
	col := startCol
	off := 0
	for i < len(runes) {
		r := runes[i]
		pos := ast.Pos{Line: line, Col: col, Off: off}

		if isIgnorableInlineRune(r) {
			i++
			col++
			off++
			continue
		}

		if unicode.IsSpace(r) {
			for i < len(runes) && unicode.IsSpace(runes[i]) {
				i++
				col++
				off++
			}
			out = append(out, token{kind: tokenSpace, text: " ", pos: pos})
			continue
		}

		if r == '.' && i+2 < len(runes) && runes[i+1] == '.' && runes[i+2] == '.' {
			out = append(out, token{kind: tokenEllipsis, text: "...", pos: pos})
			i += 3
			col += 3
			off += 3
			continue
		}

		if r == '…' {
			out = append(out, token{kind: tokenEllipsis, text: "…", pos: pos})
			i++
			col++
			off++
			continue
		}

		switch r {
		case '-', '–', '—':
			out = append(out, token{kind: tokenDash, ch: r, dashKind: dashKindFromRune(r), text: string(r), pos: pos})
			i++
			col++
			off++
			continue
		case '(':
			out = append(out, token{kind: tokenParenOpen, ch: '(', text: "(", pos: pos})
			i++
			col++
			off++
			continue
		case '[':
			out = append(out, token{kind: tokenParenOpen, ch: '[', text: "[", pos: pos})
			i++
			col++
			off++
			continue
		case ')':
			out = append(out, token{kind: tokenParenClose, ch: ')', text: ")", pos: pos})
			i++
			col++
			off++
			continue
		case ']':
			out = append(out, token{kind: tokenParenClose, ch: ']', text: "]", pos: pos})
			i++
			col++
			off++
			continue
		}

		if _, ok := punctRunes[r]; ok {
			out = append(out, token{kind: tokenPunct, ch: r, text: string(r), pos: pos})
			i++
			col++
			off++
			continue
		}

		if _, ok := quoteRunes[r]; ok {
			out = append(out, token{kind: tokenQuote, ch: r, text: string(r), pos: pos})
			i++
			col++
			off++
			continue
		}

		start := i
		startCol := col
		startOff := off
		for i < len(runes) && isWordRune(runes, i) {
			i++
			col++
			off++
		}
		if i > start {
			out = append(out, token{
				kind: tokenWord,
				text: string(runes[start:i]),
				pos:  ast.Pos{Line: line, Col: startCol, Off: startOff},
			})
			continue
		}

		out = append(out, token{kind: tokenWord, text: string(r), pos: pos})
		i++
		col++
		off++
	}

	return collapseSpaces(out)
}

func collapseSpaces(tokens []token) []token {
	if len(tokens) == 0 {
		return tokens
	}
	out := make([]token, 0, len(tokens))
	for _, tok := range tokens {
		if tok.kind == tokenSpace {
			if len(out) > 0 && out[len(out)-1].kind == tokenSpace {
				continue
			}
		}
		out = append(out, tok)
	}
	return out
}

func isWordRune(runes []rune, i int) bool {
	r := runes[i]
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	if r == '\'' || r == '’' || r == '№' {
		return true
	}
	if r == '-' {
		if i == 0 || i+1 >= len(runes) {
			return false
		}
		return unicode.IsLetter(runes[i-1]) && unicode.IsLetter(runes[i+1])
	}
	return false
}

func dashKindFromRune(r rune) ast.DashKind {
	switch r {
	case '–':
		return ast.DashNDash
	case '—':
		return ast.DashEmDash
	default:
		return ast.DashHyphen
	}
}

func isIgnorableInlineRune(r rune) bool {
	return (unicode.IsControl(r) && !unicode.IsSpace(r)) || unicode.In(r, unicode.Cf)
}
