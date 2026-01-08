package engine

import (
	"unicode"
)

// VNIMethod implements the VNI input method.
// VNI uses number keys 1-9 for tone/vowel marks.
type VNIMethod struct{}

// NewVNIMethod creates a new VNI input method.
func NewVNIMethod() *VNIMethod {
	return &VNIMethod{}
}

// Name returns the method name.
func (v *VNIMethod) Name() string {
	return "VNI"
}

// VNI key mappings for tone marks
// 1: sắc    2: huyền   3: hỏi   4: ngã   5: nặng   0: remove
var vniToneKeys = map[rune]ToneMark{
	'1': ToneSac,   // á
	'2': ToneHuyen, // à
	'3': ToneHoi,   // ả
	'4': ToneNga,   // ã
	'5': ToneNang,  // ạ
	'0': ToneNone,  // Remove tone
}

// VNI key mappings for vowel marks
// 6: circumflex (â, ê, ô)   7: horn (ơ, ư)   8: breve (ă)   9: stroke (đ)
var vniVowelKeys = map[rune]VowelMark{
	'6': VowelHat,   // Circumflex: â, ê, ô
	'7': VowelHorn,  // Horn: ơ, ư
	'8': VowelBreve, // Breve: ă
	'9': VowelDBar,  // Stroke: đ
}

// vowelTargetsForMark defines which vowels can receive each mark type
var vniVowelTargets = map[VowelMark][]rune{
	VowelHat:   {'a', 'A', 'e', 'E', 'o', 'O'}, // â, ê, ô
	VowelHorn:  {'o', 'O', 'u', 'U'},           // ơ, ư
	VowelBreve: {'a', 'A'},                     // ă
	VowelDBar:  {'d', 'D'},                     // đ
}

// vniTransformations maps (vowel + mark) -> result
var vniTransformations = map[rune]map[VowelMark]rune{
	'a': {VowelHat: 'â', VowelBreve: 'ă'},
	'A': {VowelHat: 'Â', VowelBreve: 'Ă'},
	'e': {VowelHat: 'ê'},
	'E': {VowelHat: 'Ê'},
	'o': {VowelHat: 'ô', VowelHorn: 'ơ'},
	'O': {VowelHat: 'Ô', VowelHorn: 'Ơ'},
	'u': {VowelHorn: 'ư'},
	'U': {VowelHorn: 'Ư'},
	'd': {VowelDBar: 'đ'},
	'D': {VowelDBar: 'Đ'},
}

// IsToneKey checks if the character is a VNI tone key (1-5, 0).
func (v *VNIMethod) IsToneKey(char rune) bool {
	_, ok := vniToneKeys[char]
	return ok
}

// GetToneMark returns the tone mark for a VNI character.
func (v *VNIMethod) GetToneMark(char rune) ToneMark {
	if tone, ok := vniToneKeys[char]; ok {
		return tone
	}
	return ToneNone
}

// IsVowelModifier checks if the character modifies a vowel in VNI (6-9).
func (v *VNIMethod) IsVowelModifier(char rune) bool {
	_, ok := vniVowelKeys[char]
	return ok
}

// GetVowelMark returns the vowel mark for a VNI key.
func (v *VNIMethod) GetVowelMark(char rune) VowelMark {
	if mark, ok := vniVowelKeys[char]; ok {
		return mark
	}
	return VowelNone
}

// ProcessChar processes a character according to VNI rules.
// Returns (transformed string, tone mark, vowel mark, consumed)
func (v *VNIMethod) ProcessChar(char rune, current *Syllable) (string, ToneMark, VowelMark, bool) {
	if current == nil {
		return string(char), ToneNone, VowelNone, false
	}

	// Check for tone keys (1-5, 0)
	if v.IsToneKey(char) {
		tone := v.GetToneMark(char)
		// Only apply tone if we have a vowel
		if current.Nucleus != "" {
			return "", tone, VowelNone, true
		}
		// No vowel - treat as literal number
		return string(char), ToneNone, VowelNone, false
	}

	// Check for vowel modifier keys (6-9)
	if v.IsVowelModifier(char) {
		mark := v.GetVowelMark(char)

		// Handle đ (9 key)
		if mark == VowelDBar {
			// Find 'd' in onset or nucleus
			if current.Onset != "" {
				onset := []rune(current.Onset)
				for i, r := range onset {
					if r == 'd' || r == 'D' {
						result := 'đ'
						if unicode.IsUpper(r) {
							result = 'Đ'
						}
						onset[i] = result
						return string(result), ToneNone, VowelDBar, true
					}
				}
			}
			// No 'd' found - treat as literal
			return string(char), ToneNone, VowelNone, false
		}

		// Handle vowel marks (6, 7, 8)
		if current.Nucleus != "" {
			nucleus := []rune(current.Nucleus)

			// For VNI key 7 (horn), handle UO compound: uo -> ươ
			if mark == VowelHorn && len(nucleus) >= 2 {
				last := nucleus[len(nucleus)-1]
				secondLast := nucleus[len(nucleus)-2]
				if (unicode.ToLower(secondLast) == 'u') && (unicode.ToLower(last) == 'o') {
					// Transform both: u -> ư, o -> ơ
					var transformedResult string
					for i, r := range nucleus {
						if i == len(nucleus)-2 {
							if unicode.IsUpper(r) {
								transformedResult += "Ư"
							} else {
								transformedResult += "ư"
							}
						} else if i == len(nucleus)-1 {
							if unicode.IsUpper(r) {
								transformedResult += "Ơ"
							} else {
								transformedResult += "ơ"
							}
						} else {
							transformedResult += string(r)
						}
					}
					return transformedResult, ToneNone, VowelHorn, true
				}
			}

			// Find last vowel that can accept this mark
			for i := len(nucleus) - 1; i >= 0; i-- {
				r := nucleus[i]
				if transforms, ok := vniTransformations[r]; ok {
					if result, found := transforms[mark]; found {
						return string(result), ToneNone, mark, true
					}
				}
			}
		}

		// Also check raw buffer for vowels not yet parsed
		if current.Raw != "" {
			rawRunes := []rune(current.Raw)
			for i := len(rawRunes) - 1; i >= 0; i-- {
				r := rawRunes[i]
				if transforms, ok := vniTransformations[r]; ok {
					if result, found := transforms[mark]; found {
						return string(result), ToneNone, mark, true
					}
				}
			}
		}

		// No suitable target - treat as literal number
		return string(char), ToneNone, VowelNone, false
	}

	// Regular character - just pass through
	return string(char), ToneNone, VowelNone, false
}

// CanStartWord checks if a character can start a Vietnamese word.
func (v *VNIMethod) CanStartWord(char rune) bool {
	lower := unicode.ToLower(char)
	// Vietnamese words can start with vowels or consonants
	return unicode.IsLetter(char) ||
		lower == 'a' || lower == 'e' || lower == 'i' ||
		lower == 'o' || lower == 'u' || lower == 'y'
}

// IsWordBreaker checks if a character should break the current word.
func (v *VNIMethod) IsWordBreaker(char rune) bool {
	// Space, punctuation break words
	// Numbers in VNI are modifiers, not word breakers (except when not applicable)
	return unicode.IsSpace(char) || unicode.IsPunct(char)
}

// IsVNIModifier checks if a rune is a VNI modifier (number key used for transformation)
func IsVNIModifier(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}
