package charset

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type codec struct {
	name string
	enc  encoding.Encoding
}

var (
	registryOnce sync.Once
	registry     map[string]codec
)

func Validate(name string) error {
	_, err := resolve(name)
	return err
}

func Decode(input []byte, charsetName string) (string, error) {
	c, err := resolve(charsetName)
	if err != nil {
		return "", err
	}

	if c.enc == nil {
		if !utf8.Valid(input) {
			return "", fmt.Errorf("decode %s: invalid UTF-8 input", c.name)
		}
		return string(input), nil
	}

	out, err := io.ReadAll(transform.NewReader(bytes.NewReader(input), c.enc.NewDecoder()))
	if err != nil {
		return "", fmt.Errorf("decode %s: %w", c.name, err)
	}
	return string(out), nil
}

func Encode(input string, charsetName string) ([]byte, error) {
	c, err := resolve(charsetName)
	if err != nil {
		return nil, err
	}

	if c.enc == nil {
		return []byte(input), nil
	}

	out, _, err := transform.String(c.enc.NewEncoder(), input)
	if err != nil {
		return nil, fmt.Errorf("encode %s: %w", c.name, err)
	}
	return []byte(out), nil
}

func resolve(name string) (codec, error) {
	buildRegistry()
	key := normalizeName(name)
	if c, ok := registry[key]; ok {
		return c, nil
	}
	return codec{}, fmt.Errorf("unsupported charset %q (examples: utf-8, cp1251, koi8-r, koi8-u, cp866, iso-8859-5, mac-cyrillic)", name)
}

func buildRegistry() {
	registryOnce.Do(func() {
		registry = map[string]codec{
			"":     {name: "utf-8"},
			"utf":  {name: "utf-8"},
			"utf8": {name: "utf-8"},
		}

		for _, enc := range charmap.All {
			name := encodingName(enc)
			if name == "" {
				continue
			}
			registry[normalizeName(name)] = codec{name: name, enc: enc}
		}

		aliasTargets := map[string]string{
			"cp1251":        "windows1251",
			"win1251":       "windows1251",
			"1251":          "windows1251",
			"cp866":         "ibmcodepage866",
			"ibm866":        "ibmcodepage866",
			"866":           "ibmcodepage866",
			"dos866":        "ibmcodepage866",
			"latincyrillic": "iso88595",
			"maccyrillic":   "macintoshcyrillic",
			"xmaccyrillic":  "macintoshcyrillic",
			"msmaccyrillic": "macintoshcyrillic",
			"cp10007":       "macintoshcyrillic",
		}
		for alias, target := range aliasTargets {
			if c, ok := registry[target]; ok {
				registry[alias] = c
			}
		}
	})
}

func encodingName(enc encoding.Encoding) string {
	if s, ok := enc.(fmt.Stringer); ok {
		return strings.TrimSpace(s.String())
	}
	return ""
}

func normalizeName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
