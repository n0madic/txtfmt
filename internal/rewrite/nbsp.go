package rewrite

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/n0madic/txtfmt/internal/ast"
	"github.com/n0madic/txtfmt/internal/config"
)

var shortWordsRU = map[string]struct{}{
	"в": {},
	"к": {},
	"с": {},
	"у": {},
	"о": {},
	"и": {},
	"а": {},
}

var shortWordsUA = map[string]struct{}{
	"в":  {},
	"у":  {},
	"з":  {},
	"із": {},
	"й":  {},
	"і":  {},
}

func applyNBSPDocument(doc *ast.Document, cfg config.Config) {
	applyToAllInlines(doc, func(in []ast.Inline) []ast.Inline {
		return applyNBSPList(in, cfg.Lang)
	})
}

func applyNBSPList(in []ast.Inline, lang config.Lang) []ast.Inline {
	out := make([]ast.Inline, 0, len(in))
	for _, item := range in {
		switch it := item.(type) {
		case ast.QuoteSpan:
			it.In = applyNBSPList(it.In, lang)
			out = append(out, it)
		case ast.ParenSpan:
			it.In = applyNBSPList(it.In, lang)
			out = append(out, it)
		default:
			out = append(out, item)
		}
	}

	for i := range out {
		sp, ok := out[i].(ast.Space)
		if !ok || sp.Kind != ast.SpaceNormal {
			continue
		}

		prevIdx := prevNonSpaceIndex(out, i)
		nextIdx := nextNonSpaceIndex(out, i)
		if prevIdx < 0 || nextIdx < 0 {
			continue
		}

		if shouldNBSPShortWord(out[prevIdx], out[nextIdx], lang) ||
			shouldNBSPNumero(out, prevIdx, nextIdx) ||
			shouldNBSPPageAbbr(out, prevIdx, nextIdx) ||
			shouldNBSPInitialPattern(out, prevIdx, nextIdx) {
			sp.Kind = ast.SpaceNBSP
			out[i] = sp
		}
	}

	return out
}

func prevNonSpaceIndex(in []ast.Inline, idx int) int {
	for i := idx - 1; i >= 0; i-- {
		if _, ok := in[i].(ast.Space); ok {
			continue
		}
		return i
	}
	return -1
}

func nextNonSpaceIndex(in []ast.Inline, idx int) int {
	for i := idx + 1; i < len(in); i++ {
		if _, ok := in[i].(ast.Space); ok {
			continue
		}
		return i
	}
	return -1
}

func shouldNBSPShortWord(prev, next ast.Inline, lang config.Lang) bool {
	if lang == config.LangEN {
		return false
	}
	pw, okPrev := prev.(ast.Word)
	nw, okNext := next.(ast.Word)
	if !okPrev || !okNext || nw.S == "" {
		return false
	}
	word := strings.ToLower(pw.S)
	switch lang {
	case config.LangRU:
		_, ok := shortWordsRU[word]
		return ok
	case config.LangUA:
		_, ok := shortWordsUA[word]
		return ok
	default:
		return false
	}
}

func shouldNBSPNumero(in []ast.Inline, prevIdx, nextIdx int) bool {
	pw, ok := in[prevIdx].(ast.Word)
	if !ok || pw.S != "№" {
		return false
	}
	return isNumericWordInline(in[nextIdx])
}

func shouldNBSPPageAbbr(in []ast.Inline, prevIdx, nextIdx int) bool {
	p, ok := in[prevIdx].(ast.Punct)
	if !ok || p.Ch != '.' || prevIdx < 1 {
		return false
	}
	w, ok := in[prevIdx-1].(ast.Word)
	if !ok {
		return false
	}
	if strings.ToLower(w.S) != "стр" {
		return false
	}
	return isNumericWordInline(in[nextIdx])
}

func shouldNBSPInitialPattern(in []ast.Inline, prevIdx, nextIdx int) bool {
	if !isInitialEndingAt(in, prevIdx) {
		return false
	}
	if isInitialStartingAt(in, nextIdx) {
		return true
	}
	if w, ok := in[nextIdx].(ast.Word); ok {
		return utf8.RuneCountInString(w.S) > 1
	}
	return false
}

func isInitialEndingAt(in []ast.Inline, idx int) bool {
	p, ok := in[idx].(ast.Punct)
	if !ok || p.Ch != '.' || idx < 1 {
		return false
	}
	w, ok := in[idx-1].(ast.Word)
	if !ok {
		return false
	}
	return isUpperSingleLetter(w.S)
}

func isInitialStartingAt(in []ast.Inline, idx int) bool {
	w, ok := in[idx].(ast.Word)
	if !ok || !isUpperSingleLetter(w.S) {
		return false
	}
	if idx+1 >= len(in) {
		return false
	}
	p, ok := in[idx+1].(ast.Punct)
	if !ok || p.Ch != '.' {
		return false
	}
	return true
}

func isUpperSingleLetter(s string) bool {
	runes := []rune(s)
	if len(runes) != 1 {
		return false
	}
	r := runes[0]
	return unicode.IsLetter(r) && unicode.IsUpper(r)
}

func isNumericWordInline(in ast.Inline) bool {
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
