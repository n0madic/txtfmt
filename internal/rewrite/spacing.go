package rewrite

import "github.com/n0madic/txtfmt/internal/ast"

func normalizeSpacingDocument(doc *ast.Document) {
	applyToAllInlines(doc, normalizeSpacingList)
}

func normalizeSpacingList(in []ast.Inline) []ast.Inline {
	in = normalizeChildren(in, normalizeSpacingList)

	nonSpace := make([]ast.Inline, 0, len(in))
	for _, item := range in {
		if _, ok := item.(ast.Space); ok {
			continue
		}
		nonSpace = append(nonSpace, item)
	}
	if len(nonSpace) == 0 {
		return nil
	}

	out := make([]ast.Inline, 0, len(nonSpace)*2)
	out = append(out, nonSpace[0])
	for i := 1; i < len(nonSpace); i++ {
		prev := out[len(out)-1]
		cur := nonSpace[i]
		if needSpaceBetween(prev, cur) {
			out = append(out, ast.Space{Kind: ast.SpaceNormal})
		}
		out = append(out, cur)
	}
	return out
}

func needSpaceBetween(prev, cur ast.Inline) bool {
	if isTightRight(cur) {
		return false
	}
	if isHyphenOrNDash(prev) || isHyphenOrNDash(cur) {
		return false
	}

	if isEmDash(cur) {
		return true
	}
	if isEmDash(prev) {
		return !isTightRight(cur)
	}

	if p, ok := prev.(ast.Punct); ok {
		switch p.Ch {
		case ',', ';', ':':
			return startsWordLike(cur) || isEmDash(cur)
		case '.', '?', '!':
			return startsWordLike(cur) || isEmDash(cur)
		}
	}
	if _, ok := prev.(ast.Ellipsis); ok {
		return startsWordLike(cur) || isEmDash(cur)
	}

	if startsWordLike(prev) && startsWordLike(cur) {
		return true
	}

	return false
}

func isTightRight(in ast.Inline) bool {
	if _, ok := in.(ast.Ellipsis); ok {
		return true
	}
	if p, ok := in.(ast.Punct); ok {
		switch p.Ch {
		case ',', '.', ';', ':', '?', '!':
			return true
		}
	}
	return false
}

func isHyphenOrNDash(in ast.Inline) bool {
	d, ok := in.(ast.Dash)
	if !ok {
		return false
	}
	return d.Kind == ast.DashHyphen || d.Kind == ast.DashNDash
}

func isEmDash(in ast.Inline) bool {
	d, ok := in.(ast.Dash)
	if !ok {
		return false
	}
	return d.Kind == ast.DashEmDash
}

func startsWordLike(in ast.Inline) bool {
	switch in.(type) {
	case ast.Word, ast.QuoteSpan, ast.ParenSpan:
		return true
	default:
		return false
	}
}
