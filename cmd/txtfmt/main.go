package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/abadojack/whatlanggo"
	"github.com/n0madic/txtfmt/internal/charset"
	"github.com/n0madic/txtfmt/internal/config"
	"github.com/n0madic/txtfmt/internal/diag"
	"github.com/n0madic/txtfmt/internal/parser"
	"github.com/n0madic/txtfmt/internal/printer"
	"github.com/n0madic/txtfmt/internal/rewrite"
)

const langAuto = "auto"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("txtfmt", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: txtfmt -input <file|-> [options]")
		fmt.Fprintln(stderr)
		fs.PrintDefaults()
	}

	lang := fs.String("lang", langAuto, "language: auto|en|ru|ua")
	inner := fs.String("inner-quotes", "", "inner quote style: german|english|guillemets")
	nbsp := fs.Bool("nbsp", false, "enable NBSP transformations")
	dumpAST := fs.Bool("dump-ast", false, "print parsed AST to stderr as JSON")
	format := fs.String("format", string(printer.FormatPlain), "output format: plain|markdown|html|xml")
	inputPath := fs.String("input", "", "input source: file path or '-' for stdin")
	outputPath := fs.String("output", "", "output destination: file path or '-' for stdout")
	inputCharset := fs.String("input-charset", "utf-8", "input charset (utf-8|cp1251|koi8-r|koi8-u|cp866|iso-8859-5|mac-cyrillic)")
	outputCharset := fs.String("output-charset", "utf-8", "output charset (utf-8|cp1251|koi8-r|koi8-u|cp866|iso-8859-5|mac-cyrillic)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() > 0 {
		fmt.Fprintln(stderr, "positional arguments are not supported; use -input")
		fs.Usage()
		return 2
	}

	if *inputPath == "" {
		fs.Usage()
		return 0
	}

	if err := charset.Validate(*inputCharset); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	if err := charset.Validate(*outputCharset); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	outputFormat, err := printer.ParseFormat(*format)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}

	inputRaw, err := readInput(*inputPath, stdin)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	input, err := charset.Decode(inputRaw, *inputCharset)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	resolvedLang, err := resolveLang(*lang, input)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	cfg, err := config.New(resolvedLang, *inner, *nbsp)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}

	doc := parser.Parse(input, cfg)
	if *dumpAST {
		if err := diag.WriteAST(stderr, doc); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
	}
	rewrite.Apply(&doc, cfg)

	output, err := charset.Encode(printer.PrintWithFormat(doc, outputFormat), *outputCharset)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	if err := writeOutput(*outputPath, output, stdout); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}

	if len(doc.Diags) > 0 {
		if err := diag.Write(stderr, doc.Diags); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
	}

	return 0
}

func readInput(inputPath string, stdin io.Reader) ([]byte, error) {
	if inputPath == "-" {
		return io.ReadAll(stdin)
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}
	return data, nil
}

func writeOutput(outputPath string, data []byte, stdout io.Writer) error {
	path := strings.TrimSpace(outputPath)
	if path == "" || path == "-" {
		if _, err := stdout.Write(data); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		return nil
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func resolveLang(langRaw, input string) (string, error) {
	lang := strings.ToLower(strings.TrimSpace(langRaw))
	switch lang {
	case "", langAuto:
		return detectAutoLang(input), nil
	case string(config.LangEN), string(config.LangRU), string(config.LangUA):
		return lang, nil
	default:
		return "", fmt.Errorf("unsupported --lang value %q (expected auto|en|ru|ua)", langRaw)
	}
}

func detectAutoLang(input string) string {
	info := whatlanggo.DetectWithOptions(input, whatlanggo.Options{
		Whitelist: map[whatlanggo.Lang]bool{
			whatlanggo.Eng: true,
			whatlanggo.Rus: true,
			whatlanggo.Ukr: true,
		},
	})

	switch info.Lang {
	case whatlanggo.Eng:
		return string(config.LangEN)
	case whatlanggo.Ukr:
		return string(config.LangUA)
	case whatlanggo.Rus:
		return string(config.LangRU)
	default:
		return string(config.LangEN)
	}
}
