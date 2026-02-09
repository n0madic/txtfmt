# AGENTS.md

Instructions for agents working in this repository.

## Project Goal

`txtfmt` is a deterministic Go CLI for typographic/structural text normalization.
Pipeline: `parse -> rewrite -> print`.

## Fixed CLI Contract

- Use only `-input <file|->` as the text source.
- If `-input` is not set, CLI prints help and does not format.
- Positional arguments are not supported.
- `stdout` must contain only formatted text.
- All diagnostics must go to `stderr`.
- Exit codes:
  - `0` success (even with diagnostics),
  - `1` I/O/internal error,
  - `2` invalid args/flags.

## Core Parser Rules

- Blocks:
  - `Paragraph`, `DialogueBlock`, `Heading`, `SceneBreak`.
  - Candidates are split by blank lines and by structural lines.
  - Structural lines include:
    - headings (`# ...`, `Глава ...`, `Часть ...`, `Содержание ...`),
    - scene-breaks (`***`, `-----`, `x x x`, `х х х`),
    - meta lines in `Token: value` form (for example `OCR: ...`, `Spellcheck: ...`).
- Dialogue:
  - A dialogue line starts with `-` or `—` followed by whitespace.
  - Adjacent dialogue lines are grouped into one `DialogueBlock` with multiple `Turns`.
  - If normal prose appears between them, that is a different block.
- OCR artifacts:
  - Control/format runes (`control`/`Cf`) are stripped.
  - For long uniformly-indented OCR chunks, paragraph-splitting heuristics are used (do not collapse a whole chapter into one paragraph).

## Quotes/Parentheses

- Quotes are parsed using a stack plus local heuristics.
- For ambiguous symmetric quotes (`"`, `“`, `‟`), unclosed cases remain literal text and should not spam `QUOTE_UNCLOSED`.
- Explicit parenthesis mismatches are diagnosed, except enumerators like `I)`, `2)`, `3)`.

## Charset

- Use `golang.org/x/text/encoding/charmap` (`internal/charset`).
- Do not add an `iconv` dependency.

## Required Workflow for Changes

1. Keep changes minimal and deterministic (no NLP/ML).
2. Run `gofmt -w` on changed files.
3. Run `go test ./...`.
4. If you modify `parser/spans`, verify on a real large OCR file to avoid reintroducing mass false diagnostics or structural collapsing.

## Test Content Policy

- Never use copyrighted source text in tests (including direct excerpts from books/articles).
- For all test fixtures and inline test strings, create original synthetic text specifically for tests.
- If a real-world sample is needed to reproduce formatting shape, rewrite it into a new invented variant before adding to repo.

## Behavioral Changes

If a change affects user-facing behavior (especially block segmentation, diagnostics, or dialogue formatting), add/update regression tests in `internal/parser/*_test.go` and briefly document the reason in the PR/message.
