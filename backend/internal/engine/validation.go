package engine

import (
	"strings"
	"unicode"
)

// ValidInitials are valid Vietnamese initial consonants (phụ âm đầu)
var validInitials = map[string]bool{
	// Single consonants (16)
	"b": true, "c": true, "d": true, "đ": true, "g": true, "h": true,
	"k": true, "l": true, "m": true, "n": true, "p": true, "q": true,
	"r": true, "s": true, "t": true, "v": true, "x": true,
	// Double consonants (11)
	"ch": true, "gh": true, "gi": true, "kh": true, "ng": true,
	"nh": true, "ph": true, "qu": true, "th": true, "tr": true,
	// Triple consonant
	"ngh": true,
}

// ValidFinals are valid Vietnamese final consonants (phụ âm cuối)
var validFinals = map[string]bool{
	// Single consonants
	"c": true, "m": true, "n": true, "p": true, "t": true,
	// Double consonants
	"ch": true, "ng": true, "nh": true,
	// Semi-vowels (bán nguyên âm cuối)
	"i": true, "y": true, "o": true, "u": true,
}

// SpellingRules defines invalid combinations that need correction
// key: invalid pattern, value: what should be used instead
var spellingRules = map[string]string{
	// c + (e,i,y) → should use k
	"ce": "ke", "ci": "ki", "cy": "ky",
	// k + (a,o,u) → should use c
	"ka": "ca", "ko": "co", "ku": "cu",
	// g + e → should use gh
	"ge": "ghe",
	// ng + (e,i) → should use ngh
	"nge": "nghe", "ngi": "nghi",
	// gh + (a,o,u) → should use g
	"gha": "ga", "gho": "go", "ghu": "gu",
	// ngh + (a,o,u) → should use ng
	"ngha": "nga", "ngho": "ngo", "nghu": "ngu",
}

// ValidationResult contains the result of syllable validation
type ValidationResult struct {
	Valid        bool
	Reason       string
	HasVowel     bool
	InitialValid bool
	FinalValid   bool
	SpellingOK   bool
}

// ValidateVietnamese checks if the current buffer forms a valid Vietnamese syllable
// This is called BEFORE any transformation to prevent modifying non-Vietnamese text
func ValidateVietnamese(onset, nucleus, coda string) ValidationResult {
	result := ValidationResult{Valid: true}

	// Rule 1: Must have at least one vowel
	if nucleus == "" {
		result.Valid = false
		result.Reason = "no_vowel"
		result.HasVowel = false
		return result
	}
	result.HasVowel = true

	// Rule 2: Check initial consonant (if present)
	if onset != "" {
		onsetLower := strings.ToLower(onset)
		// Remove đ for checking (it's valid)
		onsetLower = strings.ReplaceAll(onsetLower, "đ", "d")

		if !isValidInitial(onsetLower) {
			result.Valid = false
			result.Reason = "invalid_initial"
			result.InitialValid = false
			return result
		}
	}
	result.InitialValid = true

	// Rule 3: Check final consonant (if present)
	if coda != "" {
		codaLower := strings.ToLower(coda)
		if !validFinals[codaLower] {
			result.Valid = false
			result.Reason = "invalid_final"
			result.FinalValid = false
			return result
		}
	}
	result.FinalValid = true

	// Rule 4: Check spelling rules
	if onset != "" && nucleus != "" {
		combined := strings.ToLower(onset) + string(unicode.ToLower([]rune(nucleus)[0]))
		if _, invalid := spellingRules[combined]; invalid {
			result.Valid = false
			result.Reason = "spelling_rule_violation"
			result.SpellingOK = false
			return result
		}
	}
	result.SpellingOK = true

	return result
}

// isValidInitial checks if a string is a valid Vietnamese initial
func isValidInitial(s string) bool {
	if s == "" {
		return true
	}

	// Direct lookup
	if validInitials[s] {
		return true
	}

	// For single characters, check if it's a consonant
	if len(s) == 1 {
		r := []rune(s)[0]
		switch r {
		case 'b', 'c', 'd', 'g', 'h', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'x':
			return true
		}
	}

	return false
}

// ValidateForModifier checks if the buffer should accept a modifier key
// Returns true if the modifier should transform the text, false if it should be literal
func (e *CompositionEngine) ValidateForModifier() bool {
	syllable := e.buffer.syllable
	if syllable == nil {
		return false
	}

	// Check if we have a valid Vietnamese structure
	result := ValidateVietnamese(syllable.Onset, syllable.Nucleus, syllable.Coda)
	return result.Valid
}

// QuickValidate does a fast check if a raw string could be Vietnamese
// This is used before heavy parsing to quickly reject obvious non-Vietnamese
func QuickValidate(raw string) bool {
	if raw == "" {
		return false
	}

	runes := []rune(strings.ToLower(raw))

	// Check for invalid characters
	for _, r := range runes {
		// Skip modifier keys
		if r == 's' || r == 'f' || r == 'r' || r == 'x' || r == 'j' || r == 'z' || r == 'w' {
			continue
		}

		// Must be valid Vietnamese letter
		if !isValidVietnameseLetter(r) {
			return false
		}
	}

	// Check for at least one vowel (or potential vowel from modifier)
	hasVowel := false
	for _, r := range runes {
		if isVietnameseVowelRune(r) || r == 'w' { // w can become ư
			hasVowel = true
			break
		}
	}

	return hasVowel
}

// isValidVietnameseLetter checks if a rune is valid in Vietnamese
func isValidVietnameseLetter(r rune) bool {
	// Vowels
	switch r {
	case 'a', 'ă', 'â', 'e', 'ê', 'i', 'o', 'ô', 'ơ', 'u', 'ư', 'y':
		return true
	}

	// Consonants
	switch r {
	case 'b', 'c', 'd', 'đ', 'g', 'h', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'x':
		return true
	}

	// Common modifiers
	switch r {
	case 'f', 'j', 'w', 'z':
		return true
	}

	return false
}
