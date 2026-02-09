package printer

import (
	"strings"
	"testing"

	"github.com/n0madic/txtfmt/internal/ast"
	"github.com/n0madic/txtfmt/internal/config"
)

func TestPrintHTMLContentsUsesNestedLists(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	doc := ast.Document{
		Lang:  cfg.Lang,
		Style: cfg.Style,
		Blocks: []ast.Block{
			ast.ContentsBlock{
				In: []ast.Inline{ast.Word{S: "СОДЕРЖАНИЕ"}},
				Entries: []ast.ContentsEntry{
					{
						Level: 1,
						In: []ast.Inline{
							ast.Word{S: "Часть"},
							ast.Space{Kind: ast.SpaceNormal},
							ast.Word{S: "I"},
						},
					},
					{
						Level: 2,
						In: []ast.Inline{
							ast.Word{S: "Глава"},
							ast.Space{Kind: ast.SpaceNormal},
							ast.Word{S: "1"},
						},
					},
					{
						Level: 2,
						In: []ast.Inline{
							ast.Word{S: "Глава"},
							ast.Space{Kind: ast.SpaceNormal},
							ast.Word{S: "2"},
						},
					},
					{
						Level: 1,
						In: []ast.Inline{
							ast.Word{S: "Часть"},
							ast.Space{Kind: ast.SpaceNormal},
							ast.Word{S: "II"},
						},
					},
				},
			},
		},
	}

	out := PrintWithFormat(doc, FormatHTML)

	if !strings.Contains(out, "<section class=\"contents\">") {
		t.Fatalf("expected contents section in html output, got:\n%s", out)
	}
	if strings.Count(out, "<ul>") < 2 {
		t.Fatalf("expected nested lists, got:\n%s", out)
	}

	partIdx := strings.Index(out, "<li>Часть I")
	chapterIdx := strings.Index(out, "<li>Глава 1")
	if partIdx < 0 || chapterIdx < 0 || chapterIdx < partIdx {
		t.Fatalf("expected chapter entry under first part, got:\n%s", out)
	}
}

func TestPrintMarkdownDialogueNoBlockquoteAndHardBreaks(t *testing.T) {
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	doc := ast.Document{
		Lang:  cfg.Lang,
		Style: cfg.Style,
		Blocks: []ast.Block{
			ast.DialogueBlock{
				Turns: []ast.DialogueTurn{
					{In: []ast.Inline{ast.Dash{Kind: ast.DashEmDash}, ast.Space{Kind: ast.SpaceNormal}, ast.Word{S: "Привет"}}},
					{In: []ast.Inline{ast.Dash{Kind: ast.DashEmDash}, ast.Space{Kind: ast.SpaceNormal}, ast.Word{S: "Пока"}}},
				},
			},
		},
	}

	out := PrintWithFormat(doc, FormatMarkdown)
	if strings.Contains(out, "> ") {
		t.Fatalf("did not expect blockquote markers in markdown dialogue, got:\n%s", out)
	}
	if !strings.Contains(out, "\n\n") {
		t.Fatalf("expected paragraph break between dialogue turns, got:\n%s", out)
	}
}
