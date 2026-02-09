package parser

import (
	"testing"

	"github.com/n0madic/txtfmt/internal/ast"
)

func TestSymmetricQuotesCloseAtEnd(t *testing.T) {
	in, diags := parseInlineLines([]string{"\"abc\""}, []int{1}, false)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(in) != 1 {
		t.Fatalf("expected one top-level inline, got %d", len(in))
	}
	q, ok := in[0].(ast.QuoteSpan)
	if !ok {
		t.Fatalf("expected QuoteSpan, got %T", in[0])
	}
	if len(q.In) != 1 {
		t.Fatalf("expected quote content length 1, got %d", len(q.In))
	}
	if w, ok := q.In[0].(ast.Word); !ok || w.S != "abc" {
		t.Fatalf("expected quote content word abc, got %#v", q.In[0])
	}
}

func TestSymmetricQuoteBeforePunctCloses(t *testing.T) {
	in, diags := parseInlineLines([]string{"\"abc\","}, []int{1}, false)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(in) != 2 {
		t.Fatalf("expected quote span + punct, got %d", len(in))
	}
	if _, ok := in[0].(ast.QuoteSpan); !ok {
		t.Fatalf("expected first token QuoteSpan, got %T", in[0])
	}
	if p, ok := in[1].(ast.Punct); !ok || p.Ch != ',' {
		t.Fatalf("expected comma punct, got %#v", in[1])
	}
}

func TestSymmetricQuoteWithEllipsisCloses(t *testing.T) {
	in, diags := parseInlineLines([]string{"\"Обнять...\""}, []int{1}, false)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(in) != 1 {
		t.Fatalf("expected one top-level inline, got %d", len(in))
	}
	if _, ok := in[0].(ast.QuoteSpan); !ok {
		t.Fatalf("expected QuoteSpan, got %T", in[0])
	}
}

func TestSymmetricNestedQuotesAfterSpaceRemainOpen(t *testing.T) {
	in, diags := parseInlineLines(
		[]string{"\"А что такое \"координаты\"? - подумал он.\""},
		[]int{1},
		false,
	)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(in) != 1 {
		t.Fatalf("expected one top-level inline, got %d", len(in))
	}
	outer, ok := in[0].(ast.QuoteSpan)
	if !ok {
		t.Fatalf("expected outer QuoteSpan, got %T", in[0])
	}
	foundNested := false
	for _, child := range outer.In {
		if _, ok := child.(ast.QuoteSpan); ok {
			foundNested = true
			break
		}
	}
	if !foundNested {
		t.Fatalf("expected nested QuoteSpan inside outer quote")
	}
}

func TestSymmetricQuoteAtLineEndClosesBeforeNextLineWord(t *testing.T) {
	in, diags := parseInlineLines(
		[]string{"Он подумал: \"Обнять...\"", "Потом отвернулся."},
		[]int{10, 11},
		true,
	)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(in) < 1 {
		t.Fatalf("expected non-empty inlines")
	}
}

func TestEnumeratorParenDoesNotReportMismatch(t *testing.T) {
	_, diags := parseInlineLines(
		[]string{`"I) 5 лет назад, 2) 2 года назад, 3) год назад"`},
		[]int{1},
		false,
	)
	for _, d := range diags {
		if d.Code == "PAREN_MISMATCH" {
			t.Fatalf("unexpected PAREN_MISMATCH: %+v", diags)
		}
	}
}

func TestStrayQuoteBetweenWordsStaysLiteral(t *testing.T) {
	_, diags := parseInlineLines(
		[]string{`чуть пологи" склон`},
		[]int{1},
		false,
	)
	for _, d := range diags {
		if d.Code == "QUOTE_UNCLOSED" {
			t.Fatalf("unexpected QUOTE_UNCLOSED: %+v", diags)
		}
	}
}

func TestUnclosedSymmetricQuoteNoDiag(t *testing.T) {
	_, diags := parseInlineLines([]string{`"незакрытая цитата`}, []int{1}, false)
	for _, d := range diags {
		if d.Code == "QUOTE_UNCLOSED" {
			t.Fatalf("unexpected QUOTE_UNCLOSED: %+v", diags)
		}
	}
}

func TestLongRadiogramStyleSymmetricQuotesNoDiag(t *testing.T) {
	_, diags := parseInlineLines(
		[]string{
			`Он записал: "Радиограмма. Срочно. Командир судна "Вектор". Проверьте линию и подтвердите прием.`,
		},
		[]int{1},
		false,
	)
	for _, d := range diags {
		if d.Code == "QUOTE_UNCLOSED" {
			t.Fatalf("unexpected QUOTE_UNCLOSED: %+v", diags)
		}
	}
}

func TestResponseLineWithInnerNameQuoteNoDiag(t *testing.T) {
	_, diags := parseInlineLines(
		[]string{
			`Ответ был: "Ждем. Поздравляем экипаж "Вектора" и научную группу.`,
		},
		[]int{1},
		false,
	)
	for _, d := range diags {
		if d.Code == "QUOTE_UNCLOSED" {
			t.Fatalf("unexpected QUOTE_UNCLOSED: %+v", diags)
		}
	}
}
