package rewrite

import "github.com/n0madic/txtfmt/internal/ast"

func emitCanonicalPairsDocument(doc *ast.Document) {
	for i, blk := range doc.Blocks {
		switch b := blk.(type) {
		case ast.TitleBlock:
			b.In = normalizeQuoteLevels(b.In, 0)
			doc.Blocks[i] = b
		case ast.Paragraph:
			b.In = normalizeQuoteLevels(b.In, 0)
			doc.Blocks[i] = b
		case ast.Heading:
			b.In = normalizeQuoteLevels(b.In, 0)
			doc.Blocks[i] = b
		case ast.ContentsBlock:
			b.In = normalizeQuoteLevels(b.In, 0)
			for k := range b.Entries {
				b.Entries[k].In = normalizeQuoteLevels(b.Entries[k].In, 0)
			}
			doc.Blocks[i] = b
		case ast.MetaLineBlock:
			b.In = normalizeQuoteLevels(b.In, 0)
			doc.Blocks[i] = b
		case ast.DialogueBlock:
			for j := range b.Turns {
				b.Turns[j].In = normalizeQuoteLevels(b.Turns[j].In, 0)
			}
			doc.Blocks[i] = b
		}
	}
}

func normalizeQuoteLevels(in []ast.Inline, depth int) []ast.Inline {
	out := make([]ast.Inline, 0, len(in))
	for _, item := range in {
		switch it := item.(type) {
		case ast.QuoteSpan:
			it.Level = quoteLevelFromDepth(depth + 1)
			it.In = normalizeQuoteLevels(it.In, depth+1)
			out = append(out, it)
		case ast.ParenSpan:
			it.In = normalizeQuoteLevels(it.In, depth)
			out = append(out, it)
		default:
			out = append(out, item)
		}
	}
	return out
}

func quoteLevelFromDepth(depth int) ast.QuoteLevel {
	if depth%2 == 1 {
		return ast.QuotePrimary
	}
	return ast.QuoteSecondary
}
