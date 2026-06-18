package whisper

import "fmt"

// WHISPER_LANGUAGES maps ISO 639-1 codes to display names for all languages
// supported by the Whisper model.
var WHISPER_LANGUAGES = map[string]string{
	"af": "Afrikaans", "ar": "Arabic", "hy": "Armenian",
	"az": "Azerbaijani", "be": "Belarusian", "bs": "Bosnian",
	"bg": "Bulgarian", "ca": "Catalan", "zh": "Chinese",
	"hr": "Croatian", "cs": "Czech", "da": "Danish",
	"nl": "Dutch", "en": "English", "et": "Estonian",
	"fi": "Finnish", "fr": "French", "gl": "Galician",
	"de": "German", "el": "Greek", "he": "Hebrew",
	"hi": "Hindi", "hu": "Hungarian", "is": "Icelandic",
	"id": "Indonesian", "it": "Italian", "ja": "Japanese",
	"kn": "Kannada", "kk": "Kazakh", "ko": "Korean",
	"lv": "Latvian", "lt": "Lithuanian", "mk": "Macedonian",
	"ms": "Malay", "mr": "Marathi", "mi": "Maori",
	"ne": "Nepali", "no": "Norwegian", "fa": "Persian",
	"pl": "Polish", "pt": "Portuguese", "ro": "Romanian",
	"ru": "Russian", "sr": "Serbian", "sk": "Slovak",
	"sl": "Slovenian", "es": "Spanish", "sw": "Swahili",
	"sv": "Swedish", "tl": "Tagalog", "ta": "Tamil",
	"th": "Thai", "tr": "Turkish", "uk": "Ukrainian",
	"ur": "Urdu", "vi": "Vietnamese", "cy": "Welsh",
}

type UnsupportedLanguageError struct {
	Code string
}

func (e *UnsupportedLanguageError) Error() string {
	return fmt.Sprintf("language %q is not supported by Whisper", e.Code)
}

func ValidateLanguage(code string) error {
	if code == "auto" {
		return nil
	}
	if _, ok := WHISPER_LANGUAGES[code]; !ok {
		return &UnsupportedLanguageError{Code: code}
	}
	return nil
}
