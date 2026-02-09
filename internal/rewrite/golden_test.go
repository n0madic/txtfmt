package rewrite_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/n0madic/txtfmt/internal/config"
	"github.com/n0madic/txtfmt/internal/parser"
	"github.com/n0madic/txtfmt/internal/printer"
	"github.com/n0madic/txtfmt/internal/rewrite"
)

func TestGoldenCasesRU(t *testing.T) {
	cases := []string{"dialogue", "dashes", "quotes", "nested_quotes", "ellipsis", "spacing", "indented_paragraphs", "contents_meta", "contents_chapters"}
	cfg, err := config.New("ru", "", false)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			in := mustReadFixture(t, tc+".in")
			expected := mustReadFixture(t, tc+".golden")

			out := formatText(in, cfg)
			if out != expected {
				t.Fatalf("unexpected output\nexpected: %q\nactual:   %q", expected, out)
			}

			again := formatText(out, cfg)
			if again != out {
				t.Fatalf("output is not idempotent\nfirst:  %q\nsecond: %q", out, again)
			}
		})
	}
}

func formatText(input string, cfg config.Config) string {
	doc := parser.Parse(input, cfg)
	rewrite.Apply(&doc, cfg)
	return printer.Print(doc)
}

func mustReadFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.TrimSuffix(string(b), "\n")
}
