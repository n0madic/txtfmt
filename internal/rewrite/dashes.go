package rewrite

import (
	"unicode"

	"github.com/n0madic/txtfmt/internal/ast"
)

func classifyDashesDocument(doc *ast.Document) {
	applyToAllInlines(doc, classifyDashesList)
}

func classifyDashesList(in []ast.Inline) []ast.Inline {
	in = normalizeChildren(in, classifyDashesList)

	out := make([]ast.Inline, len(in))
	copy(out, in)

	for i := range out {
		d, ok := out[i].(ast.Dash)
		if !ok {
			continue
		}

		leftSpace := i > 0 && isSpace(out[i-1])
		rightSpace := i+1 < len(out) && isSpace(out[i+1])
		left := prevNonSpace(out, i)
		right := nextNonSpace(out, i)

		if !leftSpace && !rightSpace && left != nil && right != nil && isWordLike(*left) && isWordLike(*right) {
			if isNumericInline(*left) && isNumericInline(*right) {
				d.Kind = ast.DashNDash
			} else {
				d.Kind = ast.DashHyphen
			}
			out[i] = d
			continue
		}

		d.Kind = ast.DashEmDash
		out[i] = d
	}

	return out
}

func isWordLike(in ast.Inline) bool {
	switch in.(type) {
	case ast.Word, ast.QuoteSpan, ast.ParenSpan:
		return true
	default:
		return false
	}
}

func isNumericInline(in ast.Inline) bool {
	w, ok := in.(ast.Word)
	if !ok || w.S == "" {
		return false
	}
	for _, r := range []rune(w.S) {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isSpace(in ast.Inline) bool {
	_, ok := in.(ast.Space)
	return ok
}

func prevNonSpace(in []ast.Inline, idx int) *ast.Inline {
	for i := idx - 1; i >= 0; i-- {
		if isSpace(in[i]) {
			continue
		}
		return &in[i]
	}
	return nil
}

func nextNonSpace(in []ast.Inline, idx int) *ast.Inline {
	for i := idx + 1; i < len(in); i++ {
		if isSpace(in[i]) {
			continue
		}
		return &in[i]
	}
	return nil
}
