# txtfmt

`txtfmt` is a Go CLI tool for deterministic text formatting (RU/UA/EN), aimed at prose and technical text.

It reads input from a file or `stdin`, normalizes typography, and prints the rewritten text to `stdout`.

## What does

- Normalizes dashes, spacing, and ellipses.
- Normalizes quotes with nesting support.
- Detects dialogue lines and normalizes the dialogue marker to `—`.
- With `-lang auto` (default), detects input language automatically (`en`/`ru`/`ua`).
- Supports scene breaks (`***`, `-----`, `x x x`, and similar variants).
- Detects contents blocks (`CONTENTS` / `СОДЕРЖАНИЕ` / `ЗМІСТ`) with nested chapter entries.
- Detects metadata lines in `Key: value` form as separate blocks.
- Writes diagnostics to `stderr` without mixing with formatted output.
- Supports UTF-8 and legacy charsets (for example `cp1251`, `koi8-r`).
- Can dump parsed AST as JSON for debugging (`-dump-ast`).

## Requirements

- Go `1.24+`

## Build

```bash
go build -o txtfmt ./cmd/txtfmt
```

Or run without building:

```bash
go run ./cmd/txtfmt -input <file-or-stdin>
```

## Usage

```bash
txtfmt -input <file|-> [options]
```

Important:

- `-input` is required.
- Positional arguments are not supported.
- If `-input` is missing, the program prints help to `stderr` and exits without formatting.

## Flags

- `-input <file|->`
  Input source: file path or `-` for `stdin`.
- `-output <file|->`
  Output destination: file path or `-` for `stdout`.  
  If omitted, output goes to `stdout`.
- `-format plain|markdown|html|xml`
  Output text format (default: `plain`).
- `-lang auto|en|ru|ua`  
  Language rules (default: `auto`).  
  If `auto` is set, language is detected from input text; when detection is uncertain, fallback is `en`.
- `-inner-quotes german|english|guillemets`
  Inner quote style (overrides only nested quote pair).
- `-nbsp`
  Enable non-breaking space (NBSP) transformations.
- `-dump-ast`
  Print AST JSON to `stderr` (before rewrite stage).
- `-input-charset <name>`
  Input charset (default: `utf-8`).
- `-output-charset <name>`
  Output charset (default: `utf-8`).

## Exit codes

- `0` successful formatting (even when diagnostics are present).
- `1` read/write/decode/internal error.
- `2` invalid CLI args or flag values.

## Quick examples

Format a file:

```bash
./txtfmt -input input.txt > output.txt
```

Format directly to an output file:

```bash
./txtfmt -input input.txt -output output.txt
```

Format from `stdin`:

```bash
echo '- Привет!' | ./txtfmt -lang ru -input -
```

Generate Markdown:

```bash
./txtfmt -input input.txt -format markdown
```

Generate HTML:

```bash
./txtfmt -input input.txt -format html
```

Generate XML:

```bash
./txtfmt -input input.txt -format xml
```

Read `cp1251`, write UTF-8:

```bash
./txtfmt -input book.txt -input-charset cp1251 -output-charset utf-8 > book.utf8.txt
```

Read/write `cp1251`:

```bash
./txtfmt -input in.txt -input-charset cp1251 -output-charset cp1251 > out.txt
```

Dump AST + formatted output:

```bash
./txtfmt -input in.txt -dump-ast > out.txt 2> debug.log
```

## Output streams

- `stdout`: formatted text (unless `-output <file>` is used).
- `stderr`: diagnostics and (if `-dump-ast` is enabled) AST JSON.

Diagnostics format:

```text
line:col CODE message
```

Example:

```text
2328:89 PAREN_MISMATCH unexpected closing parenthesis
```

## Supported charsets

`txtfmt` uses `golang.org/x/text/encoding/charmap` and accepts standard names/aliases.

Common options:

- `utf-8`
- `cp1251` (`windows-1251`)
- `koi8-r`
- `koi8-u`
- `cp866`
- `iso-8859-5`
- `mac-cyrillic`

If charset is unsupported, the error looks like:

```text
unsupported charset "<name>" (examples: utf-8, cp1251, koi8-r, ...)
```

## Formatting rules (short)

- `...` -> `…`
- `word-word` (no spaces) stays hyphen `-`
- `1990-2000` (no spaces) -> en dash `–`
- Other dash usage is normalized to em dash `—` with proper spacing
- Quotes are emitted as canonical pairs by language and nesting level
- Dialogue line markers are normalized to `—` with a single space after it
- With `-nbsp`, NBSP rules are applied for RU/UA short function words, initials, and patterns like `№ 12`, `стр. 5`

## Block parser behavior

- Supports:
  - regular paragraphs,
  - dialogue blocks (lines starting with `- ` or `— `),
  - headings (`# Heading` and book-style forms like `Глава ...`, `Часть ...`, `Chapter ...`, `Розділ ...`),
  - dedicated contents blocks with nested chapter entries,
  - metadata lines (`Key: value`) as dedicated blocks,
  - scene breaks.
- Adjacent dialogue lines are grouped into one `DialogueBlock`.
- If dialogue lines are separated by narrative text, they become separate blocks.
- For OCR/book-like line-wrapped text, heuristics are used to avoid collapsing entire chapters into one paragraph.

## Limitations

- No NLP/ML, only local deterministic heuristics.
- Markdown parsing support in input is limited (mainly `#` headings and basic text structure).
- In ambiguous OCR cases, symmetric quotes may remain literal to avoid noisy false-positive warnings.

## Testing

```bash
go test ./...
```
