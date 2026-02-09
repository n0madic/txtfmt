package rewrite

import "github.com/n0madic/txtfmt/internal/ast"

func normalizeDialogueBlocks(doc *ast.Document) {
	for i, blk := range doc.Blocks {
		db, ok := blk.(ast.DialogueBlock)
		if !ok {
			continue
		}

		for j := range db.Turns {
			normalized, emptyAfterMarker := normalizeDialogueTurn(db.Turns[j].In)
			db.Turns[j].In = normalized
			if emptyAfterMarker {
				doc.Diags = append(doc.Diags, ast.Diag{
					Code:    "DIALOGUE_EMPTY",
					Message: "empty dialogue turn after dash marker",
				})
			}
		}

		doc.Blocks[i] = db
	}
}

func normalizeDialogueTurn(in []ast.Inline) ([]ast.Inline, bool) {
	in = trimLeadingSpaces(in)
	if len(in) == 0 {
		return []ast.Inline{ast.Dash{Kind: ast.DashEmDash}}, true
	}

	if d, ok := in[0].(ast.Dash); ok {
		d.Kind = ast.DashEmDash
		rest := trimLeadingSpaces(in[1:])
		if len(rest) == 0 {
			return []ast.Inline{d}, true
		}
		out := make([]ast.Inline, 0, len(rest)+2)
		out = append(out, d, ast.Space{Kind: ast.SpaceNormal})
		out = append(out, rest...)
		return out, false
	}

	rest := trimLeadingSpaces(in)
	if len(rest) == 0 {
		return []ast.Inline{ast.Dash{Kind: ast.DashEmDash}}, true
	}

	out := make([]ast.Inline, 0, len(rest)+2)
	out = append(out, ast.Dash{Kind: ast.DashEmDash}, ast.Space{Kind: ast.SpaceNormal})
	out = append(out, rest...)
	return out, false
}

func trimLeadingSpaces(in []ast.Inline) []ast.Inline {
	idx := 0
	for idx < len(in) {
		if _, ok := in[idx].(ast.Space); ok {
			idx++
			continue
		}
		break
	}
	return in[idx:]
}
