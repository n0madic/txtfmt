package rewrite

import (
	"testing"

	"github.com/n0madic/txtfmt/internal/ast"
	"github.com/n0madic/txtfmt/internal/config"
)

func TestNBSPRules(t *testing.T) {
	t.Run("ru short words", func(t *testing.T) {
		in := []ast.Inline{ast.Word{S: "в"}, ast.Space{Kind: ast.SpaceNormal}, ast.Word{S: "доме"}}
		out := applyNBSPList(in, config.LangRU)
		sp, ok := out[1].(ast.Space)
		if !ok || sp.Kind != ast.SpaceNBSP {
			t.Fatalf("expected NBSP, got %#v", out[1])
		}
	})

	t.Run("ua short words", func(t *testing.T) {
		in := []ast.Inline{ast.Word{S: "із"}, ast.Space{Kind: ast.SpaceNormal}, ast.Word{S: "міста"}}
		out := applyNBSPList(in, config.LangUA)
		sp, ok := out[1].(ast.Space)
		if !ok || sp.Kind != ast.SpaceNBSP {
			t.Fatalf("expected NBSP, got %#v", out[1])
		}
	})

	t.Run("numero", func(t *testing.T) {
		in := []ast.Inline{ast.Word{S: "№"}, ast.Space{Kind: ast.SpaceNormal}, ast.Word{S: "12"}}
		out := applyNBSPList(in, config.LangRU)
		sp, ok := out[1].(ast.Space)
		if !ok || sp.Kind != ast.SpaceNBSP {
			t.Fatalf("expected NBSP, got %#v", out[1])
		}
	})

	t.Run("initials", func(t *testing.T) {
		in := []ast.Inline{
			ast.Word{S: "А"}, ast.Punct{Ch: '.'}, ast.Space{Kind: ast.SpaceNormal},
			ast.Word{S: "Б"}, ast.Punct{Ch: '.'}, ast.Space{Kind: ast.SpaceNormal},
			ast.Word{S: "Иванов"},
		}
		out := applyNBSPList(in, config.LangRU)
		sp1, ok1 := out[2].(ast.Space)
		sp2, ok2 := out[5].(ast.Space)
		if !ok1 || sp1.Kind != ast.SpaceNBSP {
			t.Fatalf("expected first initials space NBSP, got %#v", out[2])
		}
		if !ok2 || sp2.Kind != ast.SpaceNBSP {
			t.Fatalf("expected surname space NBSP, got %#v", out[5])
		}
	})
}
