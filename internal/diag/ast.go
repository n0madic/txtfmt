package diag

import (
	"encoding/json"
	"io"

	"github.com/n0madic/txtfmt/internal/ast"
)

type debugDocument struct {
	Lang   string       `json:"lang"`
	Style  debugStyle   `json:"style"`
	Blocks []debugBlock `json:"blocks"`
	Diags  []debugDiag  `json:"diags,omitempty"`
}

type debugStyle struct {
	Outer debugPair `json:"outer"`
	Inner debugPair `json:"inner"`
}

type debugPair struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

type debugDiag struct {
	Line    int    `json:"line,omitempty"`
	Col     int    `json:"col,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type debugBlock struct {
	Kind    string         `json:"kind"`
	Level   int            `json:"level,omitempty"`
	Marker  string         `json:"marker,omitempty"`
	Key     string         `json:"key,omitempty"`
	In      []debugInline  `json:"in,omitempty"`
	Entries []debugEntry   `json:"entries,omitempty"`
	Turns   []debugTurn    `json:"turns,omitempty"`
	Extra   map[string]any `json:"extra,omitempty"`
}

type debugTurn struct {
	In []debugInline `json:"in"`
}

type debugEntry struct {
	Level int           `json:"level"`
	In    []debugInline `json:"in"`
}

type debugInline struct {
	Kind  string        `json:"kind"`
	Text  string        `json:"text,omitempty"`
	Ch    string        `json:"ch,omitempty"`
	Space string        `json:"space,omitempty"`
	Dash  string        `json:"dash,omitempty"`
	Level string        `json:"level,omitempty"`
	Open  string        `json:"open,omitempty"`
	Close string        `json:"close,omitempty"`
	In    []debugInline `json:"in,omitempty"`
}

func WriteAST(w io.Writer, doc ast.Document) error {
	payload := debugDocument{
		Lang: string(doc.Lang),
		Style: debugStyle{
			Outer: debugPair{Open: string(doc.Style.Outer.Open), Close: string(doc.Style.Outer.Close)},
			Inner: debugPair{Open: string(doc.Style.Inner.Open), Close: string(doc.Style.Inner.Close)},
		},
		Blocks: make([]debugBlock, 0, len(doc.Blocks)),
		Diags:  make([]debugDiag, 0, len(doc.Diags)),
	}

	for _, d := range doc.Diags {
		payload.Diags = append(payload.Diags, debugDiag{
			Line:    d.Pos.Line,
			Col:     d.Pos.Col,
			Code:    d.Code,
			Message: d.Message,
		})
	}

	for _, blk := range doc.Blocks {
		payload.Blocks = append(payload.Blocks, mapBlock(blk))
	}

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	return err
}

func mapBlock(blk ast.Block) debugBlock {
	switch b := blk.(type) {
	case ast.Paragraph:
		return debugBlock{
			Kind: "Paragraph",
			In:   mapInlines(b.In),
		}
	case ast.DialogueBlock:
		turns := make([]debugTurn, 0, len(b.Turns))
		for _, turn := range b.Turns {
			turns = append(turns, debugTurn{In: mapInlines(turn.In)})
		}
		return debugBlock{
			Kind:  "DialogueBlock",
			Turns: turns,
		}
	case ast.Heading:
		return debugBlock{
			Kind:  "Heading",
			Level: b.Level,
			In:    mapInlines(b.In),
		}
	case ast.ContentsBlock:
		entries := make([]debugEntry, 0, len(b.Entries))
		for _, entry := range b.Entries {
			entries = append(entries, debugEntry{
				Level: entry.Level,
				In:    mapInlines(entry.In),
			})
		}
		return debugBlock{
			Kind:    "ContentsBlock",
			In:      mapInlines(b.In),
			Entries: entries,
		}
	case ast.MetaLineBlock:
		return debugBlock{
			Kind: "MetaLineBlock",
			Key:  b.Key,
			In:   mapInlines(b.In),
		}
	case ast.SceneBreak:
		return debugBlock{
			Kind:   "SceneBreak",
			Marker: b.Marker,
		}
	default:
		return debugBlock{
			Kind: "UnknownBlock",
		}
	}
}

func mapInlines(in []ast.Inline) []debugInline {
	out := make([]debugInline, 0, len(in))
	for _, item := range in {
		switch it := item.(type) {
		case ast.Word:
			out = append(out, debugInline{Kind: "Word", Text: it.S})
		case ast.Space:
			out = append(out, debugInline{Kind: "Space", Space: spaceKindString(it.Kind)})
		case ast.Punct:
			out = append(out, debugInline{Kind: "Punct", Ch: string(it.Ch)})
		case ast.Dash:
			out = append(out, debugInline{Kind: "Dash", Dash: dashKindString(it.Kind)})
		case ast.Ellipsis:
			out = append(out, debugInline{Kind: "Ellipsis"})
		case ast.QuoteSpan:
			out = append(out, debugInline{
				Kind:  "QuoteSpan",
				Level: quoteLevelString(it.Level),
				In:    mapInlines(it.In),
			})
		case ast.ParenSpan:
			out = append(out, debugInline{
				Kind:  "ParenSpan",
				Open:  string(it.Open),
				Close: string(it.Close),
				In:    mapInlines(it.In),
			})
		default:
			out = append(out, debugInline{Kind: "UnknownInline"})
		}
	}
	return out
}

func spaceKindString(k ast.SpaceKind) string {
	switch k {
	case ast.SpaceNBSP:
		return "NBSP"
	case ast.SpaceThin:
		return "Thin"
	default:
		return "Normal"
	}
}

func dashKindString(k ast.DashKind) string {
	switch k {
	case ast.DashNDash:
		return "NDash"
	case ast.DashEmDash:
		return "EmDash"
	default:
		return "Hyphen"
	}
}

func quoteLevelString(level ast.QuoteLevel) string {
	switch level {
	case ast.QuoteSecondary:
		return "Secondary"
	default:
		return "Primary"
	}
}
