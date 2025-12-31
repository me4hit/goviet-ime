package engine

// UnicodeFormat implements OutputFormat for Unicode output.
type UnicodeFormat struct{}

// NewUnicodeFormat creates a new Unicode output format.
func NewUnicodeFormat() *UnicodeFormat {
	return &UnicodeFormat{}
}

// Name returns the format name.
func (u *UnicodeFormat) Name() string {
	return "Unicode"
}

// Vietnamese vowels with all tone combinations.
// Format: [base_vowel][tone] -> unicode_char
var unicodeVowelTones = map[rune]map[ToneMark]rune{
	'a': {ToneNone: 'a', ToneSac: 'á', ToneHuyen: 'à', ToneHoi: 'ả', ToneNga: 'ã', ToneNang: 'ạ'},
	'A': {ToneNone: 'A', ToneSac: 'Á', ToneHuyen: 'À', ToneHoi: 'Ả', ToneNga: 'Ã', ToneNang: 'Ạ'},
	'ă': {ToneNone: 'ă', ToneSac: 'ắ', ToneHuyen: 'ằ', ToneHoi: 'ẳ', ToneNga: 'ẵ', ToneNang: 'ặ'},
	'Ă': {ToneNone: 'Ă', ToneSac: 'Ắ', ToneHuyen: 'Ằ', ToneHoi: 'Ẳ', ToneNga: 'Ẵ', ToneNang: 'Ặ'},
	'â': {ToneNone: 'â', ToneSac: 'ấ', ToneHuyen: 'ầ', ToneHoi: 'ẩ', ToneNga: 'ẫ', ToneNang: 'ậ'},
	'Â': {ToneNone: 'Â', ToneSac: 'Ấ', ToneHuyen: 'Ầ', ToneHoi: 'Ẩ', ToneNga: 'Ẫ', ToneNang: 'Ậ'},
	'e': {ToneNone: 'e', ToneSac: 'é', ToneHuyen: 'è', ToneHoi: 'ẻ', ToneNga: 'ẽ', ToneNang: 'ẹ'},
	'E': {ToneNone: 'E', ToneSac: 'É', ToneHuyen: 'È', ToneHoi: 'Ẻ', ToneNga: 'Ẽ', ToneNang: 'Ẹ'},
	'ê': {ToneNone: 'ê', ToneSac: 'ế', ToneHuyen: 'ề', ToneHoi: 'ể', ToneNga: 'ễ', ToneNang: 'ệ'},
	'Ê': {ToneNone: 'Ê', ToneSac: 'Ế', ToneHuyen: 'Ề', ToneHoi: 'Ể', ToneNga: 'Ễ', ToneNang: 'Ệ'},
	'i': {ToneNone: 'i', ToneSac: 'í', ToneHuyen: 'ì', ToneHoi: 'ỉ', ToneNga: 'ĩ', ToneNang: 'ị'},
	'I': {ToneNone: 'I', ToneSac: 'Í', ToneHuyen: 'Ì', ToneHoi: 'Ỉ', ToneNga: 'Ĩ', ToneNang: 'Ị'},
	'o': {ToneNone: 'o', ToneSac: 'ó', ToneHuyen: 'ò', ToneHoi: 'ỏ', ToneNga: 'õ', ToneNang: 'ọ'},
	'O': {ToneNone: 'O', ToneSac: 'Ó', ToneHuyen: 'Ò', ToneHoi: 'Ỏ', ToneNga: 'Õ', ToneNang: 'Ọ'},
	'ô': {ToneNone: 'ô', ToneSac: 'ố', ToneHuyen: 'ồ', ToneHoi: 'ổ', ToneNga: 'ỗ', ToneNang: 'ộ'},
	'Ô': {ToneNone: 'Ô', ToneSac: 'Ố', ToneHuyen: 'Ồ', ToneHoi: 'Ổ', ToneNga: 'Ỗ', ToneNang: 'Ộ'},
	'ơ': {ToneNone: 'ơ', ToneSac: 'ớ', ToneHuyen: 'ờ', ToneHoi: 'ở', ToneNga: 'ỡ', ToneNang: 'ợ'},
	'Ơ': {ToneNone: 'Ơ', ToneSac: 'Ớ', ToneHuyen: 'Ờ', ToneHoi: 'Ở', ToneNga: 'Ỡ', ToneNang: 'Ợ'},
	'u': {ToneNone: 'u', ToneSac: 'ú', ToneHuyen: 'ù', ToneHoi: 'ủ', ToneNga: 'ũ', ToneNang: 'ụ'},
	'U': {ToneNone: 'U', ToneSac: 'Ú', ToneHuyen: 'Ù', ToneHoi: 'Ủ', ToneNga: 'Ũ', ToneNang: 'Ụ'},
	'ư': {ToneNone: 'ư', ToneSac: 'ứ', ToneHuyen: 'ừ', ToneHoi: 'ử', ToneNga: 'ữ', ToneNang: 'ự'},
	'Ư': {ToneNone: 'Ư', ToneSac: 'Ứ', ToneHuyen: 'Ừ', ToneHoi: 'Ử', ToneNga: 'Ữ', ToneNang: 'Ự'},
	'y': {ToneNone: 'y', ToneSac: 'ý', ToneHuyen: 'ỳ', ToneHoi: 'ỷ', ToneNga: 'ỹ', ToneNang: 'ỵ'},
	'Y': {ToneNone: 'Y', ToneSac: 'Ý', ToneHuyen: 'Ỳ', ToneHoi: 'Ỷ', ToneNga: 'Ỹ', ToneNang: 'Ỵ'},
}

// Vowel mark transformations: base_char -> mark -> result_char
var unicodeVowelMarks = map[rune]map[VowelMark]rune{
	// Breve (ă)
	'a': {VowelBreve: 'ă', VowelHat: 'â'},
	'A': {VowelBreve: 'Ă', VowelHat: 'Â'},
	// Circumflex (ê, ô)
	'e': {VowelHat: 'ê'},
	'E': {VowelHat: 'Ê'},
	'o': {VowelHat: 'ô', VowelHorn: 'ơ'},
	'O': {VowelHat: 'Ô', VowelHorn: 'Ơ'},
	// Horn (ư)
	'u': {VowelHorn: 'ư'},
	'U': {VowelHorn: 'Ư'},
	// D-bar
	'd': {VowelDBar: 'đ'},
	'D': {VowelDBar: 'Đ'},
}

// ApplyTone applies a tone mark to a vowel.
func (u *UnicodeFormat) ApplyTone(vowel rune, tone ToneMark) string {
	if tones, ok := unicodeVowelTones[vowel]; ok {
		if result, ok := tones[tone]; ok {
			return string(result)
		}
	}
	return string(vowel)
}

// ApplyVowelMark applies a vowel mark (hat, breve, horn) to a character.
func (u *UnicodeFormat) ApplyVowelMark(char rune, mark VowelMark) string {
	if marks, ok := unicodeVowelMarks[char]; ok {
		if result, ok := marks[mark]; ok {
			return string(result)
		}
	}
	return string(char)
}

// Compose creates the final Unicode string from a syllable.
func (u *UnicodeFormat) Compose(syllable *Syllable) string {
	if syllable == nil || syllable.Nucleus == "" {
		return syllable.Raw
	}

	result := syllable.Onset

	// Find the position to place the tone mark
	nucleus := []rune(syllable.Nucleus)
	tonePos := findTonePosition(nucleus, syllable.Coda)

	for i, r := range nucleus {
		// Apply vowel mark first
		modified := r
		if marks, ok := unicodeVowelMarks[r]; ok {
			if result, ok := marks[syllable.VowelMark]; ok {
				modified = result
			}
		}

		// Apply tone mark at the correct position
		if i == tonePos {
			result += u.ApplyTone(modified, syllable.ToneMark)
		} else {
			result += string(modified)
		}
	}

	result += syllable.Coda
	return result
}

// findTonePosition determines where to place the tone mark in a vowel cluster.
// Vietnamese tone placement rules (in order of priority):
// 1. If there's a marked vowel (ă, â, ê, ô, ơ, ư), put tone on it
// 2. If syllable has 'oa', 'oe', 'oo', 'uy' -> tone on the second vowel
// 3. If syllable has 'ua', 'uô', 'ưa', 'ươ' -> tone on the second vowel
// 4. If there's a coda -> tone on the vowel before last (penultimate vowel rule)
// 5. If no coda and 2+ vowels -> tone on the second vowel
// 6. Otherwise, tone on the only vowel
func findTonePosition(nucleus []rune, coda string) int {
	n := len(nucleus)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 0
	}

	// Rule 1: Find marked vowels (these always get the tone)
	for i, r := range nucleus {
		if isMarkedVowel(r) {
			return i
		}
	}

	// Rule 2: For 'oa', 'oe', 'uy' patterns without coda -> second vowel
	if n == 2 && coda == "" {
		first := nucleus[0]
		second := nucleus[1]

		// 'oa', 'oă', 'oe' -> second vowel
		if (first == 'o' || first == 'O') &&
			(second == 'a' || second == 'A' || second == 'ă' || second == 'Ă' ||
				second == 'e' || second == 'E') {
			return 1
		}

		// 'uy' -> second vowel
		if (first == 'u' || first == 'U') && (second == 'y' || second == 'Y') {
			return 1
		}
	}

	// Rule 3: For complex vowel pairs without coda
	// Using traditional/old rule (quy tắc cũ):
	// - 'ia' -> tone on 'i' (first vowel): nghĩa, mía, kìa
	// - 'ua', 'ưa' -> tone on 'a' (second vowel): mùa, lừa
	if n >= 2 && coda == "" {
		first := nucleus[0]
		second := nucleus[1]

		// 'ia' without coda -> FIRST vowel (traditional rule)
		if (first == 'i' || first == 'I') && (second == 'a' || second == 'A') {
			return 0 // Traditional: nghĩa, not nghiã
		}

		// 'ua', 'ưa' without coda -> second vowel (a)
		if (first == 'u' || first == 'U' || first == 'ư' || first == 'Ư') &&
			(second == 'a' || second == 'A') {
			return 1
		}

		// 'iê', 'uô', 'ươ' always -> the marked vowel (handled in rule 1)
	}

	// Rule 4: With coda, tone goes on the last vowel that can take a tone
	// For 2 vowels + coda (like 'oat'), put tone on first vowel
	// For 3 vowels + coda (like 'uyen'), put tone on middle vowel
	if coda != "" {
		if n == 2 {
			return 0 // First vowel: oát, oàn, etc.
		}
		if n >= 3 {
			return 1 // Middle vowel: uyến, etc.
		}
	}

	// Rule 5: Without coda, 2 vowels - need to determine which vowel gets the tone
	// 'ao', 'au', 'ay', 'ai', 'eo', 'eu' -> FIRST vowel (not second!)
	// This is the default case after 'oa', 'oe', 'uy', 'ia', 'ua', 'ưa' were handled above
	if n == 2 {
		return 0 // First vowel for 'ao', 'au', 'ay', etc.
	}

	// Rule 6: 3+ vowels without coda -> middle vowel
	if n >= 3 {
		return 1
	}

	return 0
}

// isMarkedVowel checks if a vowel has a diacritic mark (not tone)
func isMarkedVowel(r rune) bool {
	switch r {
	case 'ă', 'Ă', 'â', 'Â', 'ê', 'Ê', 'ô', 'Ô', 'ơ', 'Ơ', 'ư', 'Ư':
		return true
	}
	return false
}

// GetBaseVowel returns the base form of a vowel (without tone marks).
func GetBaseVowel(r rune) (rune, ToneMark) {
	// Check each base vowel's tone map
	for base, tones := range unicodeVowelTones {
		for tone, char := range tones {
			if char == r {
				return base, tone
			}
		}
	}
	return r, ToneNone
}

// IsVietnameseVowel checks if a character is a Vietnamese vowel.
func IsVietnameseVowel(r rune) bool {
	switch r {
	case 'a', 'A', 'ă', 'Ă', 'â', 'Â',
		'e', 'E', 'ê', 'Ê',
		'i', 'I', 'y', 'Y',
		'o', 'O', 'ô', 'Ô', 'ơ', 'Ơ',
		'u', 'U', 'ư', 'Ư':
		return true
	}

	// Also check for vowels with tone marks
	_, tone := GetBaseVowel(r)
	return tone != ToneNone
}

// IsVietnameseConsonant checks if a character is a Vietnamese consonant.
func IsVietnameseConsonant(r rune) bool {
	switch r {
	case 'b', 'B', 'c', 'C', 'd', 'D', 'đ', 'Đ',
		'g', 'G', 'h', 'H', 'k', 'K', 'l', 'L',
		'm', 'M', 'n', 'N', 'p', 'P', 'q', 'Q',
		'r', 'R', 's', 'S', 't', 'T', 'v', 'V',
		'x', 'X':
		return true
	}
	return false
}
