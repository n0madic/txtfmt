package rewrite

import (
	"testing"

	"github.com/n0madic/txtfmt/internal/ast"
)

func TestDashClassification(t *testing.T) {
	t.Run("numeric range", func(t *testing.T) {
		in := []ast.Inline{ast.Word{S: "1990"}, ast.Dash{Kind: ast.DashHyphen}, ast.Word{S: "2000"}}
		out := classifyDashesList(in)
		d, ok := out[1].(ast.Dash)
		if !ok || d.Kind != ast.DashNDash {
			t.Fatalf("expected NDash, got %#v", out[1])
		}
	})

	t.Run("word compound", func(t *testing.T) {
		in := []ast.Inline{ast.Word{S: "из"}, ast.Dash{Kind: ast.DashEmDash}, ast.Word{S: "за"}}
		out := classifyDashesList(in)
		d, ok := out[1].(ast.Dash)
		if !ok || d.Kind != ast.DashHyphen {
			t.Fatalf("expected Hyphen, got %#v", out[1])
		}
	})

	t.Run("spaced dash", func(t *testing.T) {
		in := []ast.Inline{
			ast.Word{S: "слово"}, ast.Space{Kind: ast.SpaceNormal}, ast.Dash{Kind: ast.DashHyphen}, ast.Space{Kind: ast.SpaceNormal}, ast.Word{S: "слово"},
		}
		out := classifyDashesList(in)
		d, ok := out[2].(ast.Dash)
		if !ok || d.Kind != ast.DashEmDash {
			t.Fatalf("expected EmDash, got %#v", out[2])
		}
	})
}
