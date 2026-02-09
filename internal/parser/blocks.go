package parser

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/n0madic/txtfmt/internal/ast"
	"github.com/n0madic/txtfmt/internal/config"
)

var (
	headingRe         = regexp.MustCompile(`^(#{1,6})\s+(\S.*)$`)
	bookHeadingRe     = regexp.MustCompile(`(?i)^(глава|chapter|часть|частина|part|раздел|section|книга|book|том|volume|розділ)\s+(.+)$`)
	contentsHeadingRe = regexp.MustCompile(`(?i)^(содержание|оглавление|contents|зміст)$`)
	xSceneBreakRe     = regexp.MustCompile(`^[xXхХ](?:\s+[xXхХ]){2,}$`)
	metaLineRe        = regexp.MustCompile(`^(\p{L}[\p{L}\p{N}_-]{1,30}):\s+(\S.*)$`)
)

type candidate struct {
	lines    []string
	lineNums []int
}

type sourceLine struct {
	text string
	line int
}

func Parse(input string, cfg config.Config) ast.Document {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	rawLines := strings.Split(input, "\n")
	lines := prepareSourceLines(rawLines)
	candidates := splitCandidates(lines)

	doc := ast.Document{
		Lang:   cfg.Lang,
		Style:  cfg.Style,
		Blocks: make([]ast.Block, 0, len(lines)),
		Diags:  nil,
	}

	for i := 0; i < len(candidates); i++ {
		c := candidates[i]

		if contents, diags, ok := parseContentsCandidate(c); ok {
			j := i + 1
			for j < len(candidates) {
				entry, entryDiags, entryOK := parseContentsEntryCandidate(candidates[j])
				if !entryOK {
					break
				}
				contents.Entries = append(contents.Entries, entry)
				diags = append(diags, entryDiags...)
				j++
			}

			doc.Blocks = append(doc.Blocks, contents)
			doc.Diags = append(doc.Diags, diags...)
			i = j - 1
			continue
		}

		if blocks, diags, ok := parseCandidate(c); ok {
			doc.Blocks = append(doc.Blocks, blocks...)
			doc.Diags = append(doc.Diags, diags...)
		}
	}

	return doc
}

func prepareSourceLines(raw []string) []sourceLine {
	out := make([]sourceLine, 0, len(raw))
	for i, line := range raw {
		cleaned := stripInvisibleRunes(line)
		if lead, sep, ok := splitTrailingSceneBreak(cleaned); ok {
			out = append(out, sourceLine{text: lead, line: i + 1})
			out = append(out, sourceLine{text: sep, line: i + 1})
			continue
		}
		out = append(out, sourceLine{text: cleaned, line: i + 1})
	}
	return out
}

func splitTrailingSceneBreak(line string) (string, string, bool) {
	r := []rune(strings.TrimRightFunc(line, unicode.IsSpace))
	if len(r) == 0 {
		return "", "", false
	}

	i := len(r) - 1
	for i >= 0 && (unicode.IsSpace(r[i]) || isSceneBreakRune(r[i])) {
		i--
	}

	suffix := strings.TrimSpace(string(r[i+1:]))
	prefix := strings.TrimSpace(string(r[:i+1]))
	if prefix == "" || suffix == "" {
		return "", "", false
	}
	if !isSceneBreak(suffix) || sceneBreakRuneCount(suffix) < 6 {
		return "", "", false
	}
	return prefix, suffix, true
}

func stripInvisibleRunes(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (unicode.IsControl(r) && !unicode.IsSpace(r)) || unicode.In(r, unicode.Cf) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func splitCandidates(lines []sourceLine) []candidate {
	out := make([]candidate, 0)
	cur := candidate{}

	flush := func() {
		if len(cur.lines) == 0 {
			return
		}
		out = append(out, cur)
		cur = candidate{}
	}

	for _, line := range lines {
		if strings.TrimSpace(line.text) == "" {
			flush()
			continue
		}

		if isStandaloneStructuralLine(line.text) {
			flush()
			out = append(out, candidate{lines: []string{line.text}, lineNums: []int{line.line}})
			continue
		}

		cur.lines = append(cur.lines, line.text)
		cur.lineNums = append(cur.lineNums, line.line)
	}
	flush()

	return out
}

func isStandaloneStructuralLine(line string) bool {
	if strings.TrimSpace(line) == "" {
		return false
	}
	if isSceneBreak(line) {
		return true
	}
	if _, _, ok := parseMetaLine(line); ok {
		return true
	}
	if _, ok := parseContentsLine(line); ok {
		return true
	}
	_, _, ok := parseHeading(line)
	return ok
}

func parseCandidate(c candidate) ([]ast.Block, []ast.Diag, bool) {
	if len(c.lines) == 0 {
		return nil, nil, false
	}

	if len(c.lines) == 1 && isSceneBreak(c.lines[0]) {
		return []ast.Block{ast.SceneBreak{Marker: "***"}}, nil, true
	}

	if len(c.lines) == 1 {
		if key, value, ok := parseMetaLine(c.lines[0]); ok {
			in, diags := parseInlineLines([]string{value}, c.lineNums, false)
			return []ast.Block{ast.MetaLineBlock{Key: key, In: in}}, diags, true
		}
		if level, body, ok := parseHeading(c.lines[0]); ok {
			in, diags := parseInlineLines([]string{body}, c.lineNums, false)
			return []ast.Block{ast.Heading{Level: level, In: in}}, diags, true
		}
	}

	if isDialogueCandidate(c.lines) {
		turns, allDiags := parseDialogueTurns(c.lines, c.lineNums)
		return []ast.Block{ast.DialogueBlock{Turns: turns}}, allDiags, true
	}

	parts := splitParagraphByIndent(c.lines, c.lineNums)
	blocks := make([]ast.Block, 0, len(parts))
	allDiags := make([]ast.Diag, 0)
	for _, part := range parts {
		if isDialogueCandidate(part.lines) {
			turns, diags := parseDialogueTurns(part.lines, part.lineNums)
			blocks = append(blocks, ast.DialogueBlock{Turns: turns})
			allDiags = append(allDiags, diags...)
			continue
		}
		in, diags := parseInlineLines(part.lines, part.lineNums, true)
		blocks = append(blocks, ast.Paragraph{In: in})
		allDiags = append(allDiags, diags...)
	}

	return blocks, allDiags, true
}

func parseDialogueTurns(lines []string, lineNums []int) ([]ast.DialogueTurn, []ast.Diag) {
	turns := make([]ast.DialogueTurn, 0, len(lines))
	var allDiags []ast.Diag
	for i, line := range lines {
		trimmed, col := trimLeftWithCol(line)
		in, diags := parseInlineLinesWithCols([]string{trimmed}, []int{lineNums[i]}, []int{col}, false)
		turns = append(turns, ast.DialogueTurn{In: in})
		allDiags = append(allDiags, diags...)
	}
	return turns, allDiags
}

func parseInlineLines(lines []string, lineNums []int, joinWithSpace bool) ([]ast.Inline, []ast.Diag) {
	cols := make([]int, len(lines))
	for i := range cols {
		cols[i] = 1
	}
	return parseInlineLinesWithCols(lines, lineNums, cols, joinWithSpace)
}

func parseInlineLinesWithCols(lines []string, lineNums []int, cols []int, joinWithSpace bool) ([]ast.Inline, []ast.Diag) {
	var toks []token
	for i, line := range lines {
		toks = append(toks, tokenizeInline(line, lineNums[i], cols[i])...)
		if joinWithSpace && i+1 < len(lines) {
			toks = append(toks, token{kind: tokenSpace, text: " ", pos: ast.Pos{Line: lineNums[i], Col: len([]rune(lines[i])) + 1}})
		}
	}
	return buildSpans(toks)
}

func isDialogueCandidate(lines []string) bool {
	if len(lines) == 0 {
		return false
	}
	for _, line := range lines {
		if !isDialogueLine(line) {
			return false
		}
	}
	return true
}

func isDialogueLine(s string) bool {
	t := strings.TrimLeft(s, " \t")
	if len([]rune(t)) < 2 {
		return false
	}
	r := []rune(t)
	if (r[0] == '—' || r[0] == '-') && unicode.IsSpace(r[1]) {
		return true
	}
	return false
}

func isSceneBreak(line string) bool {
	t := normalizeStructureLine(line)
	if t == "" {
		return false
	}

	if xSceneBreakRe.MatchString(t) {
		return true
	}

	compact := strings.ReplaceAll(t, " ", "")
	if utf8.RuneCountInString(compact) < 3 {
		return false
	}

	onlyXLike := true
	for _, r := range compact {
		if !isSceneBreakRune(r) {
			return false
		}
		if !isXLikeRune(r) {
			onlyXLike = false
		}
	}
	if onlyXLike {
		return xSceneBreakRe.MatchString(t)
	}
	return true
}

func isSceneBreakRune(r rune) bool {
	switch r {
	case '*', '-', '—', '–', '_', '=', '~', '·', '•', 'x', 'X', 'х', 'Х':
		return true
	default:
		return false
	}
}

func isXLikeRune(r rune) bool {
	switch r {
	case 'x', 'X', 'х', 'Х':
		return true
	default:
		return false
	}
}

func sceneBreakRuneCount(line string) int {
	n := 0
	for _, r := range line {
		if unicode.IsSpace(r) {
			continue
		}
		n++
	}
	return n
}

func parseHeading(line string) (int, string, bool) {
	t := normalizeStructureLine(line)
	if t == "" {
		return 0, "", false
	}
	if m := headingRe.FindStringSubmatch(t); m != nil {
		level := len(m[1])
		body := strings.TrimSpace(m[2])
		if body == "" {
			return 0, "", false
		}
		return level, body, true
	}

	t = trimHeadingDecorations(t)
	if t == "" {
		return 0, "", false
	}

	if m := bookHeadingRe.FindStringSubmatch(t); m != nil {
		kw := strings.ToLower(m[1])
		body := strings.TrimSpace(m[2])
		if body == "" {
			return 0, "", false
		}
		return headingLevelForKeyword(kw), t, true
	}

	return 0, "", false
}

func parseContentsCandidate(c candidate) (ast.ContentsBlock, []ast.Diag, bool) {
	if len(c.lines) != 1 {
		return ast.ContentsBlock{}, nil, false
	}
	body, ok := parseContentsLine(c.lines[0])
	if !ok {
		return ast.ContentsBlock{}, nil, false
	}
	in, diags := parseInlineLines([]string{body}, c.lineNums, false)
	return ast.ContentsBlock{In: in}, diags, true
}

func parseContentsEntryCandidate(c candidate) (ast.ContentsEntry, []ast.Diag, bool) {
	if len(c.lines) != 1 {
		return ast.ContentsEntry{}, nil, false
	}
	level, body, ok := parseContentsEntryLine(c.lines[0])
	if !ok {
		return ast.ContentsEntry{}, nil, false
	}
	in, diags := parseInlineLines([]string{body}, c.lineNums, false)
	return ast.ContentsEntry{Level: level, In: in}, diags, true
}

func parseContentsLine(line string) (string, bool) {
	t := normalizeStructureLine(line)
	if t == "" {
		return "", false
	}
	if contentsHeadingRe.MatchString(t) {
		return t, true
	}

	t = trimHeadingDecorations(t)
	if t == "" {
		return "", false
	}
	if contentsHeadingRe.MatchString(t) {
		return t, true
	}
	return "", false
}

func parseContentsEntryLine(line string) (int, string, bool) {
	t := normalizeStructureLine(line)
	if t == "" {
		return 0, "", false
	}

	t = trimHeadingDecorations(t)
	if t == "" {
		return 0, "", false
	}

	m := bookHeadingRe.FindStringSubmatch(t)
	if m == nil {
		return 0, "", false
	}
	kw := strings.ToLower(m[1])
	body := strings.TrimSpace(m[2])
	if body == "" {
		return 0, "", false
	}
	return headingLevelForKeyword(kw), t, true
}

func parseMetaLine(line string) (string, string, bool) {
	t := normalizeStructureLine(line)
	if t == "" {
		return "", "", false
	}
	m := metaLineRe.FindStringSubmatch(t)
	if m == nil {
		return "", "", false
	}
	key := strings.TrimSpace(m[1])
	value := strings.TrimSpace(m[2])
	if key == "" || value == "" {
		return "", "", false
	}
	return key, value, true
}

func normalizeStructureLine(line string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(line)), " ")
}

func trimHeadingDecorations(s string) string {
	for {
		r := []rune(strings.TrimSpace(s))
		if len(r) < 2 {
			return strings.TrimSpace(s)
		}
		if !isHeadingDecorationRune(r[0]) || !isHeadingDecorationRune(r[len(r)-1]) {
			return strings.TrimSpace(string(r))
		}
		s = string(r[1 : len(r)-1])
	}
}

func isHeadingDecorationRune(r rune) bool {
	switch r {
	case '*', '-', '—', '–', '_', '=', '~', '·', '•':
		return true
	default:
		return false
	}
}

func headingLevelForKeyword(kw string) int {
	switch kw {
	case "часть", "частина", "part", "книга", "book", "том", "volume":
		return 1
	default:
		return 2
	}
}

func trimLeftWithCol(s string) (string, int) {
	col := 1
	for _, r := range s {
		if r == ' ' || r == '\t' {
			col++
			continue
		}
		break
	}
	return strings.TrimLeft(s, " \t"), col
}

func splitParagraphByIndent(lines []string, lineNums []int) []candidate {
	if len(lines) == 0 {
		return nil
	}

	if shouldSplitLongIndentedLines(lines) {
		return splitLongIndentedParagraphs(lines, lineNums)
	}

	out := make([]candidate, 0, 1)
	start := 0
	for i := 1; i < len(lines); i++ {
		if shouldSplitBetweenLines(lines[i-1], lines[i]) {
			out = append(out, candidate{
				lines:    lines[start:i],
				lineNums: lineNums[start:i],
			})
			start = i
		}
	}

	out = append(out, candidate{
		lines:    lines[start:],
		lineNums: lineNums[start:],
	})
	return out
}

func shouldSplitBetweenLines(prev, cur string) bool {
	prevDialogue := isDialogueLine(prev)
	curDialogue := isDialogueLine(cur)
	if prevDialogue != curDialogue {
		return true
	}
	return hasLeadingIndent(cur) && !hasLeadingIndent(prev)
}

func shouldSplitLongIndentedLines(lines []string) bool {
	if len(lines) < 2 {
		return false
	}

	indented := 0
	totalLen := 0
	longLines := 0
	for _, line := range lines {
		if hasLeadingIndent(line) {
			indented++
		}
		l := lineRuneLen(strings.TrimSpace(line))
		totalLen += l
		if l >= 120 {
			longLines++
		}
	}

	avgLen := totalLen / len(lines)
	return indented == len(lines) && (avgLen >= 140 || longLines*100/len(lines) >= 60)
}

func splitLongIndentedParagraphs(lines []string, lineNums []int) []candidate {
	out := make([]candidate, 0, len(lines))
	for i := 0; i < len(lines); {
		if isDialogueLine(lines[i]) {
			j := i + 1
			for j < len(lines) && isDialogueLine(lines[j]) {
				j++
			}
			out = append(out, candidate{
				lines:    lines[i:j],
				lineNums: lineNums[i:j],
			})
			i = j
			continue
		}

		j := i + 1
		for j < len(lines) && !isDialogueLine(lines[j]) && startsLikelyContinuation(lines[j]) {
			j++
		}
		out = append(out, candidate{
			lines:    lines[i:j],
			lineNums: lineNums[i:j],
		})
		i = j
	}
	return out
}

func startsLikelyContinuation(line string) bool {
	t := strings.TrimSpace(line)
	if t == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(t)
	if unicode.IsLower(r) || unicode.IsDigit(r) {
		return true
	}
	switch r {
	case ',', '.', ';', ':', '!', '?', '…', ')', ']', '}', '»':
		return true
	default:
		return false
	}
}

func lineRuneLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func hasLeadingIndent(s string) bool {
	if s == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return r == ' ' || r == '\t'
}
