package diag

import (
	"fmt"
	"io"

	"github.com/n0madic/txtfmt/internal/ast"
)

func Format(d ast.Diag) string {
	line := d.Pos.Line
	col := d.Pos.Col
	if line <= 0 {
		line = 0
	}
	if col <= 0 {
		col = 0
	}
	return fmt.Sprintf("%d:%d %s %s", line, col, d.Code, d.Message)
}

func Write(w io.Writer, diags []ast.Diag) error {
	for _, d := range diags {
		if _, err := fmt.Fprintln(w, Format(d)); err != nil {
			return err
		}
	}
	return nil
}
