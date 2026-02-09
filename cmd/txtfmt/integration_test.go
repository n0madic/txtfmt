package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIFromStdin(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"--lang", "ru", "-input", "-"}, "- Привет!")
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if stdout != "— Привет!" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCLIAutoLangByDefaultDetectsEnglish(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"-input", "-"}, `"Where there is a will there is a way"`)
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if stdout != "“Where there is a will there is a way”" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCLIExplicitLangDisablesAutoDetection(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"--lang", "ru", "-input", "-"}, `"Where there is a will there is a way"`)
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if stdout != "«Where there is a will there is a way»" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCLIFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "in.txt")
	if err := os.WriteFile(path, []byte("Ну... ладно."), 0o644); err != nil {
		t.Fatalf("write temp input: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{"--lang", "ru", "-input", path}, "")
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if stdout != "Ну… ладно." {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
}

func TestCLIOutputToFile(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "out.txt")

	stdout, stderr, code := runCLI(t, []string{
		"--lang", "ru",
		"-input", "-",
		"-output", outPath,
	}, "- Привет!")
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected empty stdout when -output is set, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(got) != "— Привет!" {
		t.Fatalf("unexpected output file content: %q", string(got))
	}
}

func TestCLIOutputFormatMarkdown(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{
		"--lang", "ru",
		"-input", "-",
		"-format", "markdown",
	}, "- Привет!\n- Пока!")
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	want := "— Привет!\n\n— Пока!"
	if stdout != want {
		t.Fatalf("unexpected markdown output:\nwant: %q\ngot:  %q", want, stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCLIOutputFormatHTML(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{
		"--lang", "ru",
		"-input", "-",
		"-format", "html",
	}, "- Привет!")
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "<article>") || !strings.Contains(stdout, "<div class=\"dialogue\">") {
		t.Fatalf("unexpected html output: %q", stdout)
	}
	if !strings.Contains(stdout, "<p>— Привет!</p>") {
		t.Fatalf("expected normalized dialogue line in html, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCLIOutputFormatXML(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{
		"--lang", "ru",
		"-input", "-",
		"-format", "xml",
	}, "- Привет!")
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "<document lang=\"ru\">") || !strings.Contains(stdout, "<dialogue>") {
		t.Fatalf("unexpected xml output: %q", stdout)
	}
	if !strings.Contains(stdout, "<turn>— Привет!</turn>") {
		t.Fatalf("expected normalized dialogue turn in xml, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCLIOutputFormatXMLLeadingTitleBlock(t *testing.T) {
	input := strings.Join([]string{
		"\x14Ирина Ветрова. Сигнал с ледяной орбиты\x15",
		"---------------------------------------------------------------",
		"OCR: Алексей Протасов",
		"Spellcheck: Марина Логинова ---------------------------------------------------------------",
		"",
		"\x14 * ЧАСТЬ ПЕРВАЯ * \x15",
		"",
		"Первый абзац основного текста.",
	}, "\n")

	stdout, stderr, code := runCLI(t, []string{
		"--lang", "ru",
		"-input", "-",
		"-format", "xml",
	}, input)
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "<title>Ирина Ветрова. Сигнал с ледяной орбиты</title>") {
		t.Fatalf("expected title block in xml output, got %q", stdout)
	}
	if strings.Contains(stdout, "<paragraph>Ирина Ветрова. Сигнал с ледяной орбиты</paragraph>") {
		t.Fatalf("leading title must not be printed as paragraph, got %q", stdout)
	}
	if !strings.Contains(stdout, "<paragraph>Первый абзац основного текста.</paragraph>") {
		t.Fatalf("expected body paragraph to remain paragraph, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestCLIUnsupportedFormatReturns2(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{
		"--lang", "ru",
		"-input", "-",
		"-format", "pdf",
	}, "- Привет!")
	if code != 1 {
		t.Fatalf("expected go run exit code 1 for app code 2, got %d", code)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "unsupported -format value") {
		t.Fatalf("expected format validation error, got %q", stderr)
	}
}

func TestCLIOutputWriteErrorReturns1(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "missing", "out.txt")

	stdout, stderr, code := runCLI(t, []string{
		"--lang", "ru",
		"-input", "-",
		"-output", outPath,
	}, "- Привет!")
	if code != 1 {
		t.Fatalf("expected code 1, got %d stderr=%q", code, stderr)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected empty stdout on write error, got %q", stdout)
	}
	if !strings.Contains(stderr, "write output") {
		t.Fatalf("expected write output error, got %q", stderr)
	}
}

func TestCLIDiagnosticsGoToStderr(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"--lang", "ru", "-input", "-"}, "Привет )")
	if code != 0 {
		t.Fatalf("expected code 0 with diagnostics, got %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Привет") {
		t.Fatalf("expected formatted text in stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "PAREN_MISMATCH") {
		t.Fatalf("expected diagnostic code in stderr, got %q", stderr)
	}
}

func TestCLIDumpASTToStderr(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"--lang", "ru", "--dump-ast", "-input", "-"}, "- Привет!")
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if stdout != "— Привет!" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if !strings.Contains(stderr, "\"kind\": \"DialogueBlock\"") {
		t.Fatalf("expected AST JSON in stderr, got %q", stderr)
	}
	if !strings.Contains(stderr, "\"dash\": \"Hyphen\"") {
		t.Fatalf("expected pre-rewrite dash kind in AST dump, got %q", stderr)
	}
}

func TestCLIWithoutInputPrintsHelp(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"--lang", "ru"}, "")
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, stderr)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "Usage: txtfmt -input <file|-> [options]") {
		t.Fatalf("expected usage in stderr, got %q", stderr)
	}
	if !strings.Contains(stderr, "-input") {
		t.Fatalf("expected -input flag description in usage, got %q", stderr)
	}
}

func TestCLIInputCharsetCP1251(t *testing.T) {
	stdin := []byte{0x2d, 0x20, 0xcf, 0xf0, 0xe8, 0xe2, 0xe5, 0xf2, 0x21}
	stdout, stderr, code := runCLIBytes(t, []string{
		"--lang", "ru",
		"-input", "-",
		"-input-charset", "cp1251",
		"-output-charset", "utf-8",
	}, stdin)
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%q", code, string(stderr))
	}
	if string(stdout) != "— Привет!" {
		t.Fatalf("unexpected stdout: %q", string(stdout))
	}
	if strings.TrimSpace(string(stderr)) != "" {
		t.Fatalf("expected empty stderr, got %q", string(stderr))
	}
}

func TestCLIUnsupportedCharsetReturns2(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{
		"--lang", "ru",
		"-input", "-",
		"-input-charset", "x-unknown",
	}, "- Привет!")
	if code != 1 {
		t.Fatalf("expected go run exit code 1 for app code 2, got %d", code)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "unsupported charset") {
		t.Fatalf("expected unsupported charset error, got %q", stderr)
	}
}

func runCLI(t *testing.T, args []string, stdin string) (string, string, int) {
	t.Helper()
	stdout, stderr, code := runCLIBytes(t, args, []byte(stdin))
	return string(stdout), string(stderr), code
}

func runCLIBytes(t *testing.T, args []string, stdin []byte) ([]byte, []byte, int) {
	t.Helper()
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	err := cmd.Run()
	if err == nil {
		return outBuf.Bytes(), errBuf.Bytes(), 0
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run go command: %v", err)
	}
	return outBuf.Bytes(), errBuf.Bytes(), exitErr.ExitCode()
}
