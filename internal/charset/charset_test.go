package charset

import "testing"

func TestDecodeCP1251(t *testing.T) {
	in := []byte{0xcf, 0xf0, 0xe8, 0xe2, 0xe5, 0xf2}
	got, err := Decode(in, "cp1251")
	if err != nil {
		t.Fatalf("decode cp1251: %v", err)
	}
	if got != "Привет" {
		t.Fatalf("unexpected decode result: %q", got)
	}
}

func TestRoundTripKOI8R(t *testing.T) {
	original := "Привет, мир!"
	encoded, err := Encode(original, "koi8-r")
	if err != nil {
		t.Fatalf("encode koi8-r: %v", err)
	}
	decoded, err := Decode(encoded, "koi8-r")
	if err != nil {
		t.Fatalf("decode koi8-r: %v", err)
	}
	if decoded != original {
		t.Fatalf("round-trip mismatch: got %q want %q", decoded, original)
	}
}

func TestAliases(t *testing.T) {
	aliases := []string{"utf8", "windows-1251", "win1251", "1251", "koi8r", "ibm866", "iso8859-5", "x-mac-cyrillic"}
	for _, alias := range aliases {
		if err := Validate(alias); err != nil {
			t.Fatalf("expected alias %q to be accepted, got error: %v", alias, err)
		}
	}
}

func TestUnsupportedCharset(t *testing.T) {
	if err := Validate("x-unknown"); err == nil {
		t.Fatalf("expected unsupported charset error")
	}
}

func TestInvalidUTF8(t *testing.T) {
	_, err := Decode([]byte{0xff, 0xfe}, "utf-8")
	if err == nil {
		t.Fatalf("expected invalid utf-8 error")
	}
}
