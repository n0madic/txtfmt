package config

import (
	"fmt"
	"strings"
)

type Lang string

const (
	LangEN Lang = "en"
	LangRU Lang = "ru"
	LangUA Lang = "ua"
)

type InnerQuotes string

const (
	InnerQuotesGerman    InnerQuotes = "german"
	InnerQuotesEnglish   InnerQuotes = "english"
	InnerQuotesGuillemet InnerQuotes = "guillemets"
)

type QuotePair struct {
	Open  rune
	Close rune
}

type Style struct {
	Outer QuotePair
	Inner QuotePair
}

type Config struct {
	Lang        Lang
	InnerQuotes InnerQuotes
	UseNBSP     bool
	Style       Style
}

func DefaultConfig() Config {
	cfg := Config{
		Lang:        LangRU,
		InnerQuotes: InnerQuotesGerman,
		UseNBSP:     false,
	}
	cfg.Style = defaultStyleForLang(cfg.Lang)
	return cfg
}

func New(langRaw, innerRaw string, useNBSP bool) (Config, error) {
	cfg := DefaultConfig()
	cfg.UseNBSP = useNBSP

	if langRaw != "" {
		cfg.Lang = Lang(strings.ToLower(langRaw))
	}
	if err := validateLang(cfg.Lang); err != nil {
		return Config{}, err
	}

	cfg.Style = defaultStyleForLang(cfg.Lang)

	if innerRaw != "" {
		inner := InnerQuotes(strings.ToLower(innerRaw))
		if err := validateInnerQuotes(inner); err != nil {
			return Config{}, err
		}
		cfg.InnerQuotes = inner
		cfg.Style.Inner = innerPair(inner)
	}

	return cfg, nil
}

func validateLang(lang Lang) error {
	switch lang {
	case LangEN, LangRU, LangUA:
		return nil
	default:
		return fmt.Errorf("unsupported --lang value %q (expected en|ru|ua)", string(lang))
	}
}

func validateInnerQuotes(inner InnerQuotes) error {
	switch inner {
	case InnerQuotesGerman, InnerQuotesEnglish, InnerQuotesGuillemet:
		return nil
	default:
		return fmt.Errorf("unsupported --inner-quotes value %q (expected german|english|guillemets)", string(inner))
	}
}

func defaultStyleForLang(lang Lang) Style {
	switch lang {
	case LangEN:
		return Style{
			Outer: QuotePair{Open: '“', Close: '”'},
			Inner: QuotePair{Open: '‘', Close: '’'},
		}
	case LangRU, LangUA:
		fallthrough
	default:
		return Style{
			Outer: QuotePair{Open: '«', Close: '»'},
			Inner: QuotePair{Open: '„', Close: '“'},
		}
	}
}

func innerPair(inner InnerQuotes) QuotePair {
	switch inner {
	case InnerQuotesEnglish:
		return QuotePair{Open: '‘', Close: '’'}
	case InnerQuotesGuillemet:
		return QuotePair{Open: '«', Close: '»'}
	case InnerQuotesGerman:
		fallthrough
	default:
		return QuotePair{Open: '„', Close: '“'}
	}
}
