package rewrite

import (
	"github.com/n0madic/txtfmt/internal/ast"
	"github.com/n0madic/txtfmt/internal/config"
)

func Apply(doc *ast.Document, cfg config.Config) {
	normalizeEllipsisDocument(doc)
	classifyDashesDocument(doc)
	emitCanonicalPairsDocument(doc)
	normalizeSpacingDocument(doc)
	normalizeDialogueBlocks(doc)
	if cfg.UseNBSP {
		applyNBSPDocument(doc, cfg)
	}
}

func applyToAllInlines(doc *ast.Document, fn func([]ast.Inline) []ast.Inline) {
	for i, blk := range doc.Blocks {
		switch b := blk.(type) {
		case ast.Paragraph:
			b.In = fn(b.In)
			doc.Blocks[i] = b
		case ast.Heading:
			b.In = fn(b.In)
			doc.Blocks[i] = b
		case ast.ContentsBlock:
			b.In = fn(b.In)
			for j := range b.Entries {
				b.Entries[j].In = fn(b.Entries[j].In)
			}
			doc.Blocks[i] = b
		case ast.MetaLineBlock:
			b.In = fn(b.In)
			doc.Blocks[i] = b
		case ast.DialogueBlock:
			for j := range b.Turns {
				b.Turns[j].In = fn(b.Turns[j].In)
			}
			doc.Blocks[i] = b
		}
	}
}

func normalizeChildren(in []ast.Inline, fn func([]ast.Inline) []ast.Inline) []ast.Inline {
	out := make([]ast.Inline, 0, len(in))
	for _, item := range in {
		switch it := item.(type) {
		case ast.QuoteSpan:
			it.In = fn(it.In)
			out = append(out, it)
		case ast.ParenSpan:
			it.In = fn(it.In)
			out = append(out, it)
		default:
			out = append(out, item)
		}
	}
	return out
}
