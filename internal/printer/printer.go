package printer

import (
	"fmt"
	"html"
	"strings"

	"github.com/n0madic/txtfmt/internal/ast"
	"github.com/n0madic/txtfmt/internal/config"
)

type Format string

const (
	FormatPlain    Format = "plain"
	FormatMarkdown Format = "markdown"
	FormatHTML     Format = "html"
	FormatXML      Format = "xml"
)

func ParseFormat(raw string) (Format, error) {
	f := Format(strings.ToLower(strings.TrimSpace(raw)))
	if f == "" {
		f = FormatPlain
	}

	switch f {
	case FormatPlain, FormatMarkdown, FormatHTML, FormatXML:
		return f, nil
	default:
		return "", fmt.Errorf("unsupported -format value %q (expected plain|markdown|html|xml)", raw)
	}
}

func Print(doc ast.Document) string {
	return PrintWithFormat(doc, FormatPlain)
}

func PrintWithFormat(doc ast.Document, format Format) string {
	switch format {
	case FormatMarkdown:
		return printMarkdown(doc)
	case FormatHTML:
		return printHTML(doc)
	case FormatXML:
		return printXML(doc)
	case FormatPlain:
		fallthrough
	default:
		return printPlain(doc)
	}
}

func printPlain(doc ast.Document) string {
	parts := make([]string, 0, len(doc.Blocks))
	for _, blk := range doc.Blocks {
		switch b := blk.(type) {
		case ast.TitleBlock:
			parts = append(parts, printInlines(b.In, doc.Style))
		case ast.Paragraph:
			parts = append(parts, printInlines(b.In, doc.Style))
		case ast.Heading:
			parts = append(parts, strings.Repeat("#", b.Level)+" "+printInlines(b.In, doc.Style))
		case ast.ContentsBlock:
			parts = append(parts, printContentsBlock(b, doc.Style))
		case ast.MetaLineBlock:
			value := printInlines(b.In, doc.Style)
			if value == "" {
				parts = append(parts, b.Key+":")
			} else {
				parts = append(parts, b.Key+": "+value)
			}
		case ast.DialogueBlock:
			lines := make([]string, 0, len(b.Turns))
			for _, turn := range b.Turns {
				lines = append(lines, printInlines(turn.In, doc.Style))
			}
			parts = append(parts, strings.Join(lines, "\n"))
		case ast.SceneBreak:
			parts = append(parts, "***")
		}
	}
	return strings.Join(parts, "\n\n")
}

func printMarkdown(doc ast.Document) string {
	parts := make([]string, 0, len(doc.Blocks))
	for _, blk := range doc.Blocks {
		switch b := blk.(type) {
		case ast.TitleBlock:
			parts = append(parts, "# "+printInlines(b.In, doc.Style))
		case ast.Paragraph:
			parts = append(parts, printInlines(b.In, doc.Style))
		case ast.Heading:
			level := max(b.Level, 1)
			if level > 6 {
				level = 6
			}
			parts = append(parts, strings.Repeat("#", level)+" "+printInlines(b.In, doc.Style))
		case ast.ContentsBlock:
			parts = append(parts, printMarkdownContentsBlock(b, doc.Style))
		case ast.MetaLineBlock:
			value := printInlines(b.In, doc.Style)
			if value == "" {
				parts = append(parts, "- **"+b.Key+":**")
			} else {
				parts = append(parts, "- **"+b.Key+":** "+value)
			}
		case ast.DialogueBlock:
			lines := make([]string, 0, len(b.Turns))
			for _, turn := range b.Turns {
				lines = append(lines, printInlines(turn.In, doc.Style))
			}
			parts = append(parts, strings.Join(lines, "\n\n"))
		case ast.SceneBreak:
			parts = append(parts, "---")
		}
	}
	return strings.Join(parts, "\n\n")
}

func printHTML(doc ast.Document) string {
	lines := make([]string, 0, len(doc.Blocks)+2)
	lines = append(lines, "<article>")
	for _, blk := range doc.Blocks {
		switch b := blk.(type) {
		case ast.TitleBlock:
			lines = append(lines, "  <h1>"+escapeHTMLText(printInlines(b.In, doc.Style))+"</h1>")
		case ast.Paragraph:
			lines = append(lines, "  <p>"+escapeHTMLText(printInlines(b.In, doc.Style))+"</p>")
		case ast.Heading:
			level := max(b.Level, 1)
			if level > 6 {
				level = 6
			}
			lines = append(lines, fmt.Sprintf("  <h%d>%s</h%d>", level, escapeHTMLText(printInlines(b.In, doc.Style)), level))
		case ast.ContentsBlock:
			lines = append(lines, printHTMLContentsBlock(b, doc.Style)...)
		case ast.MetaLineBlock:
			value := printInlines(b.In, doc.Style)
			lines = append(lines, "  <p class=\"meta\"><span class=\"key\">"+escapeHTMLText(b.Key)+":</span> "+escapeHTMLText(value)+"</p>")
		case ast.DialogueBlock:
			lines = append(lines, "  <div class=\"dialogue\">")
			for _, turn := range b.Turns {
				lines = append(lines, "    <p>"+escapeHTMLText(printInlines(turn.In, doc.Style))+"</p>")
			}
			lines = append(lines, "  </div>")
		case ast.SceneBreak:
			lines = append(lines, "  <hr />")
		}
	}
	lines = append(lines, "</article>")
	return strings.Join(lines, "\n")
}

func printXML(doc ast.Document) string {
	lines := make([]string, 0, len(doc.Blocks)+2)
	lines = append(lines, "<document lang=\""+escapeXMLAttr(string(doc.Lang))+"\">")
	for _, blk := range doc.Blocks {
		switch b := blk.(type) {
		case ast.TitleBlock:
			lines = append(lines, "  <title>"+escapeXMLText(printInlines(b.In, doc.Style))+"</title>")
		case ast.Paragraph:
			lines = append(lines, "  <paragraph>"+escapeXMLText(printInlines(b.In, doc.Style))+"</paragraph>")
		case ast.Heading:
			level := max(b.Level, 1)
			lines = append(lines, fmt.Sprintf("  <heading level=\"%d\">%s</heading>", level, escapeXMLText(printInlines(b.In, doc.Style))))
		case ast.ContentsBlock:
			lines = append(lines, "  <contents>")
			lines = append(lines, "    <title>"+escapeXMLText(printInlines(b.In, doc.Style))+"</title>")
			for _, entry := range b.Entries {
				level := max(entry.Level, 1)
				lines = append(lines, fmt.Sprintf("    <entry level=\"%d\">%s</entry>", level, escapeXMLText(printInlines(entry.In, doc.Style))))
			}
			lines = append(lines, "  </contents>")
		case ast.MetaLineBlock:
			lines = append(lines, "  <meta key=\""+escapeXMLAttr(b.Key)+"\">"+escapeXMLText(printInlines(b.In, doc.Style))+"</meta>")
		case ast.DialogueBlock:
			lines = append(lines, "  <dialogue>")
			for _, turn := range b.Turns {
				lines = append(lines, "    <turn>"+escapeXMLText(printInlines(turn.In, doc.Style))+"</turn>")
			}
			lines = append(lines, "  </dialogue>")
		case ast.SceneBreak:
			lines = append(lines, "  <scene-break marker=\""+escapeXMLAttr(b.Marker)+"\" />")
		}
	}
	lines = append(lines, "</document>")
	return strings.Join(lines, "\n")
}

func printContentsBlock(b ast.ContentsBlock, style config.Style) string {
	lines := make([]string, 0, len(b.Entries)+1)
	lines = append(lines, printInlines(b.In, style))
	for _, entry := range b.Entries {
		lines = append(lines, printInlines(entry.In, style))
	}
	return strings.Join(lines, "\n")
}

func printMarkdownContentsBlock(b ast.ContentsBlock, style config.Style) string {
	lines := make([]string, 0, len(b.Entries)+1)
	lines = append(lines, "## "+printInlines(b.In, style))
	for _, entry := range b.Entries {
		indent := strings.Repeat("  ", maxInt(entry.Level-1, 0))
		lines = append(lines, indent+"- "+printInlines(entry.In, style))
	}
	return strings.Join(lines, "\n")
}

func printHTMLContentsBlock(b ast.ContentsBlock, style config.Style) []string {
	lines := make([]string, 0, len(b.Entries)+3)
	lines = append(lines, "  <section class=\"contents\">")
	lines = append(lines, "    <h2>"+escapeHTMLText(printInlines(b.In, style))+"</h2>")
	if len(b.Entries) > 0 {
		tree := buildContentsTree(b.Entries, style)
		lines = append(lines, renderContentsTreeHTML(tree, "    ")...)
	}
	lines = append(lines, "  </section>")
	return lines
}

type contentsTreeNode struct {
	Text     string
	Children []*contentsTreeNode
}

func buildContentsTree(entries []ast.ContentsEntry, style config.Style) []*contentsTreeNode {
	if len(entries) == 0 {
		return nil
	}

	minLevel := 0
	for _, entry := range entries {
		lvl := maxInt(entry.Level, 1)
		if minLevel == 0 || lvl < minLevel {
			minLevel = lvl
		}
	}
	if minLevel == 0 {
		minLevel = 1
	}

	root := &contentsTreeNode{}
	stack := []*contentsTreeNode{root}
	for _, entry := range entries {
		lvl := max(maxInt(entry.Level, 1)-minLevel+1, 1)

		for len(stack)-1 >= lvl {
			stack = stack[:len(stack)-1]
		}

		node := &contentsTreeNode{
			Text: escapeHTMLText(printInlines(entry.In, style)),
		}
		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, node)
		stack = append(stack, node)
	}

	return root.Children
}

func renderContentsTreeHTML(nodes []*contentsTreeNode, indent string) []string {
	if len(nodes) == 0 {
		return nil
	}

	lines := make([]string, 0, len(nodes)*3+2)
	lines = append(lines, indent+"<ul>")
	for _, node := range nodes {
		lines = append(lines, indent+"  <li>"+node.Text)
		if len(node.Children) > 0 {
			lines = append(lines, renderContentsTreeHTML(node.Children, indent+"    ")...)
		}
		lines = append(lines, indent+"  </li>")
	}
	lines = append(lines, indent+"</ul>")
	return lines
}

func escapeHTMLText(s string) string {
	return html.EscapeString(s)
}

func escapeXMLText(s string) string {
	repl := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return repl.Replace(s)
}

func escapeXMLAttr(s string) string {
	repl := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return repl.Replace(s)
}

func maxInt(v, min int) int {
	if v < min {
		return min
	}
	return v
}

func printInlines(in []ast.Inline, style config.Style) string {
	var b strings.Builder
	for _, item := range in {
		switch it := item.(type) {
		case ast.Word:
			b.WriteString(it.S)
		case ast.Space:
			switch it.Kind {
			case ast.SpaceNBSP:
				b.WriteRune('\u00A0')
			case ast.SpaceThin:
				b.WriteRune('\u2009')
			default:
				b.WriteRune(' ')
			}
		case ast.Punct:
			b.WriteRune(it.Ch)
		case ast.Dash:
			switch it.Kind {
			case ast.DashNDash:
				b.WriteRune('–')
			case ast.DashEmDash:
				b.WriteRune('—')
			default:
				b.WriteRune('-')
			}
		case ast.Ellipsis:
			b.WriteRune('…')
		case ast.ParenSpan:
			b.WriteRune(it.Open)
			b.WriteString(printInlines(it.In, style))
			b.WriteRune(it.Close)
		case ast.QuoteSpan:
			pair := pairForLevel(style, it.Level)
			b.WriteRune(pair.Open)
			b.WriteString(printInlines(it.In, style))
			b.WriteRune(pair.Close)
		}
	}
	return b.String()
}

func pairForLevel(style config.Style, level ast.QuoteLevel) config.QuotePair {
	if level == ast.QuoteSecondary {
		return style.Inner
	}
	return style.Outer
}
