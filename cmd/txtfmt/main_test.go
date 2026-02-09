package main

import (
	"testing"

	"github.com/n0madic/txtfmt/internal/config"
)

func TestResolveLangAutoFallbackToEnglish(t *testing.T) {
	got, err := resolveLang(langAuto, "12345 !!!")
	if err != nil {
		t.Fatalf("resolveLang returned error: %v", err)
	}
	if got != string(config.LangEN) {
		t.Fatalf("expected fallback %q, got %q", config.LangEN, got)
	}
}

func TestResolveLangExplicitUA(t *testing.T) {
	got, err := resolveLang("ua", "будь-який текст")
	if err != nil {
		t.Fatalf("resolveLang returned error: %v", err)
	}
	if got != string(config.LangUA) {
		t.Fatalf("expected %q, got %q", config.LangUA, got)
	}
}

func TestResolveLangRejectsLegacyUKCode(t *testing.T) {
	if _, err := resolveLang("uk", "будь-який текст"); err == nil {
		t.Fatalf("expected error for legacy uk code, got nil")
	}
}
