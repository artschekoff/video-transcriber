package whisper

import (
	"errors"
	"testing"
)

func TestValidateLanguage_AcceptsAuto(t *testing.T) {
	if err := ValidateLanguage("auto"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateLanguage_AcceptsKnownCode(t *testing.T) {
	for _, code := range []string{"en", "ru", "zh", "de"} {
		if err := ValidateLanguage(code); err != nil {
			t.Errorf("code %q: expected nil, got %v", code, err)
		}
	}
}

func TestValidateLanguage_RejectsUnknown(t *testing.T) {
	err := ValidateLanguage("xx")
	if err == nil {
		t.Fatal("expected error for unknown code, got nil")
	}
	var langErr *UnsupportedLanguageError
	if !errors.As(err, &langErr) {
		t.Fatalf("expected UnsupportedLanguageError, got %T", err)
	}
	if langErr.Code != "xx" {
		t.Errorf("expected Code=xx, got %q", langErr.Code)
	}
}

func TestWHISPER_LANGUAGES_AllTwoLetter(t *testing.T) {
	for code, name := range WHISPER_LANGUAGES {
		if len(code) != 2 {
			t.Errorf("code %q has length %d, want 2", code, len(code))
		}
		if name == "" {
			t.Errorf("empty name for code %q", code)
		}
	}
}
