package rewrite

import "github.com/n0madic/txtfmt/internal/ast"

func normalizeEllipsisDocument(doc *ast.Document) {
	applyToAllInlines(doc, normalizeEllipsisList)
}

func normalizeEllipsisList(in []ast.Inline) []ast.Inline {
	in = normalizeChildren(in, normalizeEllipsisList)

	out := make([]ast.Inline, 0, len(in))
	for i := 0; i < len(in); i++ {
		if isTripleDotPunct(in, i) {
			out = append(out, ast.Ellipsis{})
			i += 2
			continue
		}

		switch it := in[i].(type) {
		case ast.Word:
			out = append(out, splitWordEllipsis(it.S)...)
		default:
			out = append(out, in[i])
		}
	}
	return out
}

func isTripleDotPunct(in []ast.Inline, i int) bool {
	if i+2 >= len(in) {
		return false
	}
	p1, ok1 := in[i].(ast.Punct)
	p2, ok2 := in[i+1].(ast.Punct)
	p3, ok3 := in[i+2].(ast.Punct)
	return ok1 && ok2 && ok3 && p1.Ch == '.' && p2.Ch == '.' && p3.Ch == '.'
}

func splitWordEllipsis(s string) []ast.Inline {
	runes := []rune(s)
	out := make([]ast.Inline, 0, 2)
	buf := make([]rune, 0, len(runes))

	flush := func() {
		if len(buf) == 0 {
			return
		}
		out = append(out, ast.Word{S: string(buf)})
		buf = buf[:0]
	}

	for i := 0; i < len(runes); {
		if runes[i] == '…' {
			flush()
			out = append(out, ast.Ellipsis{})
			i++
			continue
		}
		if runes[i] == '.' && i+2 < len(runes) && runes[i+1] == '.' && runes[i+2] == '.' {
			flush()
			out = append(out, ast.Ellipsis{})
			i += 3
			continue
		}
		buf = append(buf, runes[i])
		i++
	}
	flush()

	if len(out) == 0 {
		return []ast.Inline{ast.Word{S: s}}
	}
	return out
}
