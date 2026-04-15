package rewrite

import (
	"strings"

	"github.com/n0madic/txtfmt/internal/ast"
)

func normalizeEllipsisDocument(doc *ast.Document) {
	applyToAllInlines(doc, normalizeEllipsisList)
}

func normalizeEllipsisList(in []ast.Inline) []ast.Inline {
	in = normalizeChildren(in, normalizeEllipsisList)

	out := make([]ast.Inline, 0, len(in))
	for _, item := range in {
		if w, ok := item.(ast.Word); ok {
			out = append(out, splitWordEllipsis(w.S)...)
			continue
		}
		out = append(out, item)
	}
	return out
}

func splitWordEllipsis(s string) []ast.Inline {
	if !strings.ContainsAny(s, ".…") {
		return []ast.Inline{ast.Word{S: s}}
	}

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
