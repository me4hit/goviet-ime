package engine

import (
	"unicode"
)

// TelexMethod implements the Telex input method.
type TelexMethod struct{}

// NewTelexMethod creates a new Telex input method.
func NewTelexMethod() *TelexMethod {
	return &TelexMethod{}
}

// Name returns the method name.
func (t *TelexMethod) Name() string {
	return "Telex"
}

// Telex tone key mappings
var telexToneKeys = map[rune]ToneMark{
	's': ToneSac,   // á
	'f': ToneHuyen, // à
	'r': ToneHoi,   // ả
	'x': ToneNga,   // ã
	'j': ToneNang,  // ạ
	'z': ToneNone,  // Remove tone (thanh ngang)
}

// Telex vowel modifier mappings
var telexVowelModifiers = map[rune]struct {
	targets []rune
	mark    VowelMark
}{
	'a': {targets: []rune{'a', 'A'}, mark: VowelBreve}, // aw -> ă
	'w': {targets: []rune{'a', 'A', 'o', 'O', 'u', 'U'}, mark: VowelHorn},
	'e': {targets: []rune{'e', 'E'}, mark: VowelHat}, // ee -> ê
	'o': {targets: []rune{'o', 'O'}, mark: VowelHat}, // oo -> ô
}

// Special double-letter patterns for vowel marks
var telexDoublePatterns = map[string]struct {
	result rune
	mark   VowelMark
}{
	"aa": {result: 'â', mark: VowelHat},
	"AA": {result: 'Â', mark: VowelHat},
	"Aa": {result: 'Â', mark: VowelHat},
	"aA": {result: 'â', mark: VowelHat},
	"ee": {result: 'ê', mark: VowelHat},
	"EE": {result: 'Ê', mark: VowelHat},
	"Ee": {result: 'Ê', mark: VowelHat},
	"eE": {result: 'ê', mark: VowelHat},
	"oo": {result: 'ô', mark: VowelHat},
	"OO": {result: 'Ô', mark: VowelHat},
	"Oo": {result: 'Ô', mark: VowelHat},
	"oO": {result: 'ô', mark: VowelHat},
	"dd": {result: 'đ', mark: VowelDBar},
	"DD": {result: 'Đ', mark: VowelDBar},
	"Dd": {result: 'Đ', mark: VowelDBar},
	"dD": {result: 'đ', mark: VowelDBar},
}

// Horn patterns with 'w'
var telexHornPatterns = map[rune]rune{
	'o': 'ơ',
	'O': 'Ơ',
	'u': 'ư',
	'U': 'Ư',
	'a': 'ă',
	'A': 'Ă',
}

// IsToneKey checks if the character is a Telex tone key.
func (t *TelexMethod) IsToneKey(char rune) bool {
	_, ok := telexToneKeys[unicode.ToLower(char)]
	return ok
}

// GetToneMark returns the tone mark for a Telex character.
func (t *TelexMethod) GetToneMark(char rune) ToneMark {
	if tone, ok := telexToneKeys[unicode.ToLower(char)]; ok {
		return tone
	}
	return ToneNone
}

// IsVowelModifier checks if the character modifies a vowel in Telex.
func (t *TelexMethod) IsVowelModifier(char rune) bool {
	lower := unicode.ToLower(char)
	return lower == 'w' || lower == 'a' || lower == 'e' || lower == 'o' || lower == 'd'
}

// GetVowelMark returns the vowel mark for a character.
func (t *TelexMethod) GetVowelMark(char rune) VowelMark {
	switch unicode.ToLower(char) {
	case 'w':
		return VowelHorn // or VowelBreve for 'aw'
	case 'd':
		return VowelDBar
	default:
		return VowelHat // For double letters
	}
}

// ProcessChar processes a character according to Telex rules.
func (t *TelexMethod) ProcessChar(char rune, current *Syllable) (string, ToneMark, VowelMark, bool) {
	if current == nil {
		return string(char), ToneNone, VowelNone, false
	}

	lower := unicode.ToLower(char)

	// Check for tone keys
	if t.IsToneKey(char) {
		tone := t.GetToneMark(char)
		// Only apply tone if we have a vowel
		if current.Nucleus != "" {
			return "", tone, VowelNone, true
		}
	}

	// Check for 'w' vowel modifier (ư, ơ, ă)
	if lower == 'w' {
		// 'w' can modify the previous vowel
		if current.Nucleus != "" {
			nucleus := []rune(current.Nucleus)
			lastVowel := nucleus[len(nucleus)-1]
			if result, ok := telexHornPatterns[lastVowel]; ok {
				return string(result), ToneNone, VowelHorn, true
			}
		}
		// 'uw' for 'ư' at the beginning
		if current.Raw != "" && len(current.Raw) > 0 {
			lastRaw := []rune(current.Raw)
			last := lastRaw[len(lastRaw)-1]
			if result, ok := telexHornPatterns[last]; ok {
				return string(result), ToneNone, VowelHorn, true
			}
		}
	}

	// Check for double-letter patterns (aa -> â, etc.)
	if current.Raw != "" {
		rawRunes := []rune(current.Raw)
		if len(rawRunes) > 0 {
			lastRaw := rawRunes[len(rawRunes)-1]
			pattern := string(lastRaw) + string(char)
			if p, ok := telexDoublePatterns[pattern]; ok {
				return string(p.result), ToneNone, p.mark, true
			}
		}
	}

	// Regular character - just pass through
	return string(char), ToneNone, VowelNone, false
}

// CanStartWord checks if a character can start a Vietnamese word.
func (t *TelexMethod) CanStartWord(char rune) bool {
	lower := unicode.ToLower(char)
	// Vietnamese words can start with vowels or consonants
	return unicode.IsLetter(char) ||
		lower == 'a' || lower == 'e' || lower == 'i' ||
		lower == 'o' || lower == 'u' || lower == 'y'
}

// IsWordBreaker checks if a character should break the current word.
func (t *TelexMethod) IsWordBreaker(char rune) bool {
	// Space, punctuation, numbers break words
	return unicode.IsSpace(char) || unicode.IsPunct(char) || unicode.IsDigit(char)
}
