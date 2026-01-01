package engine

import (
	"strings"
	"unicode"
)

// CompositionEngine is the main engine that processes keyboard input.
type CompositionEngine struct {
	inputMethod  InputMethod
	outputFormat OutputFormat
	buffer       *CompositionBuffer
	enabled      bool
}

// CompositionBuffer holds the current composition state.
type CompositionBuffer struct {
	raw           strings.Builder // Raw input characters
	syllable      *Syllable       // Parsed syllable structure
	committed     string          // Text to commit
	modifierCount int             // Number of modifier characters consumed (tones, vowel marks)
}

// NewCompositionBuffer creates a new empty buffer.
func NewCompositionBuffer() *CompositionBuffer {
	return &CompositionBuffer{
		syllable: &Syllable{},
	}
}

// NewCompositionEngine creates a new composition engine with default settings.
func NewCompositionEngine() *CompositionEngine {
	return &CompositionEngine{
		inputMethod:  NewTelexMethod(), // Default to Telex
		outputFormat: NewUnicodeFormat(),
		buffer:       NewCompositionBuffer(),
		enabled:      true,
	}
}

// SetInputMethod sets the typing method (e.g., Telex, VNI).
func (e *CompositionEngine) SetInputMethod(method InputMethod) {
	e.inputMethod = method
}

// SetOutputFormat sets the output encoding format.
func (e *CompositionEngine) SetOutputFormat(format OutputFormat) {
	e.outputFormat = format
}

// SetEnabled enables or disables the engine.
func (e *CompositionEngine) SetEnabled(enabled bool) {
	e.enabled = enabled
	if !enabled {
		e.Reset()
	}
}

// IsEnabled returns whether the engine is enabled.
func (e *CompositionEngine) IsEnabled() bool {
	return e.enabled
}

// Reset clears the current composition state.
func (e *CompositionEngine) Reset() {
	e.buffer = NewCompositionBuffer()
}

// GetPreedit returns the current preedit string.
func (e *CompositionEngine) GetPreedit() string {
	raw := e.buffer.raw.String()
	if raw == "" {
		return ""
	}

	syllable := e.buffer.syllable
	if syllable == nil {
		return raw
	}

	// Always try to compose from structure first
	composed := e.outputFormat.Compose(syllable)

	// Append any unparsed characters from the raw buffer,
	// but skip characters that are Telex modifiers if they were likely already consumed.
	runes := []rune(raw)
	if syllable.Consumed < len(runes) && syllable.Consumed >= 0 {
		for _, r := range runes[syllable.Consumed:] {
			if !isTelexModifier(r) {
				composed += string(r)
			}
		}
	}

	if composed != "" {
		return composed
	}

	return raw
}

// ProcessKey handles a key event and returns the result.
func (e *CompositionEngine) ProcessKey(event KeyEvent) ProcessResult {
	result := ProcessResult{
		Handled:    false,
		CommitText: "",
		Preedit:    "",
	}

	// If disabled, don't process
	if !e.enabled {
		return result
	}

	// Handle special keys
	if specialResult, handled := e.handleSpecialKey(event); handled {
		return specialResult
	}

	// Check for modifiers (Ctrl, Alt) - commit and don't process these
	if event.Modifiers&(ModControl|ModMod1) != 0 {
		if e.buffer.raw.Len() > 0 {
			preedit := e.GetPreedit()
			e.Reset()
			result.Handled = false
			result.CommitText = preedit
			result.Preedit = ""
			return result
		}
		return result
	}

	// Convert keysym to character
	char := keysymToRune(event.KeySym)
	if char == 0 {
		return result
	}

	// Process the character
	return e.processChar(char)
}

// handleSpecialKey handles special keys like Backspace, Space, Enter.
func (e *CompositionEngine) handleSpecialKey(event KeyEvent) (ProcessResult, bool) {
	result := ProcessResult{}

	switch event.KeySym {
	case KeyBackspace:
		return e.handleBackspace(), true

	case KeySpace:
		// Commit current composition and add space
		preedit := e.GetPreedit()
		e.Reset()
		result.Handled = true
		result.CommitText = preedit + " "
		result.Preedit = ""
		return result, true

	case KeyReturn:
		// If we have preedit, commit it and let Enter pass through to app
		// If no preedit, don't handle - let app receive the Enter key
		preedit := e.GetPreedit()
		if preedit != "" {
			e.Reset()
			result.Handled = true
			result.CommitText = preedit // Don't add \n, let app handle Enter
			result.Preedit = ""
			return result, true
		}
		// No preedit - don't handle, let Enter pass through
		return result, false

	case KeyEscape:
		// Cancel composition
		e.Reset()
		result.Handled = true
		result.CommitText = ""
		result.Preedit = ""
		return result, true

	case KeyTab:
		// Commit current composition and pass through tab
		if e.buffer.raw.Len() > 0 {
			preedit := e.GetPreedit()
			e.Reset()
			result.Handled = true
			result.CommitText = preedit
			result.Preedit = ""
			return result, true
		}
		return result, false // Let tab pass through

	case KeyDelete:
		// If we have preedit, commit it and then let Delete pass to app
		if e.buffer.raw.Len() > 0 {
			preedit := e.GetPreedit()
			e.Reset()
			result.Handled = false
			result.CommitText = preedit
			result.Preedit = ""
			return result, true
		}
		return result, false
	}

	return result, false
}

// handleBackspace handles the backspace key.
func (e *CompositionEngine) handleBackspace() ProcessResult {
	result := ProcessResult{Handled: true}

	raw := e.buffer.raw.String()
	if len(raw) == 0 {
		// Nothing to delete, pass through
		result.Handled = false
		return result
	}

	// Remove the last character
	runes := []rune(raw)
	newRaw := string(runes[:len(runes)-1])

	// Re-parse the syllable using full processKeyInternal logic
	e.Reset()
	// OPTIMIZATION: processKeyInternal will call updateSyllableStructure
	// which is now faster due to global map.
	for _, r := range newRaw {
		e.processKeyInternal(r)
	}

	result.Preedit = e.GetPreedit()
	return result
}

// processChar processes a regular character input.
func (e *CompositionEngine) processChar(char rune) ProcessResult {
	e.processKeyInternal(char)
	return ProcessResult{
		Handled: true,
		Preedit: e.GetPreedit(),
	}
}

// processKeyInternal is a helper for re-parsing
func (e *CompositionEngine) processKeyInternal(char rune) {
	transformed, tone, vowelMark, consumed := e.inputMethod.ProcessChar(char, e.buffer.syllable)

	if consumed {
		if e.inputMethod.IsToneKey(char) {
			if e.buffer.syllable.ToneMark == tone && tone != ToneNone {
				e.buffer.syllable.ToneMark = ToneNone
			} else {
				e.buffer.syllable.ToneMark = tone
			}
		} else if vowelMark != VowelNone || len(transformed) > 0 {
			e.applyVowelMark(vowelMark, transformed)
		}
		e.buffer.raw.WriteRune(char)
		e.updateSyllableStructure()
	} else {
		e.processCharInternal(char)
	}
}

// processCharInternal adds a character to the buffer and updates the syllable.
func (e *CompositionEngine) processCharInternal(char rune) {
	e.buffer.raw.WriteRune(char)

	// Update syllable structure
	// The raw string is now the source of truth for updateSyllableStructure
	e.updateSyllableStructure()
}

// applyVowelMark applies a vowel mark to the current syllable.
func (e *CompositionEngine) applyVowelMark(mark VowelMark, transformed string) {
	e.buffer.syllable.VowelMark = mark

	// Special case for đ - it modifies the onset consonant 'd', not a vowel
	if mark == VowelDBar && len(transformed) > 0 {
		// Check if onset ends with 'd' and replace with 'đ"
		if len(e.buffer.syllable.Onset) > 0 {
			onset := []rune(e.buffer.syllable.Onset)
			last := onset[len(onset)-1]
			if last == 'd' || last == 'D' {
				onset[len(onset)-1] = []rune(transformed)[0]
				e.buffer.syllable.Onset = string(onset)
			}
		} else if e.buffer.syllable.Raw != "" {
			// If no onset parsed yet, but we have raw 'd', update raw
			rawRunes := []rune(e.buffer.syllable.Raw)
			if len(rawRunes) > 0 {
				last := rawRunes[len(rawRunes)-1]
				if last == 'd' || last == 'D' {
					rawRunes[len(rawRunes)-1] = []rune(transformed)[0]
					e.buffer.syllable.Raw = string(rawRunes)
				}
			}
		}
		return
	}

	// If transformed contains the result, update nucleus
	if len(transformed) > 0 && len(e.buffer.syllable.Nucleus) > 0 {
		// Replace the last vowel in the nucleus
		nucleus := []rune(e.buffer.syllable.Nucleus)
		nucleus[len(nucleus)-1] = []rune(transformed)[0]
		e.buffer.syllable.Nucleus = string(nucleus)
	}
}

// updateSyllableStructure parses the raw input into onset, nucleus, coda.
func (e *CompositionEngine) updateSyllableStructure() {
	raw := e.buffer.raw.String()
	if raw == "" {
		e.buffer.syllable = &Syllable{}
		return
	}

	// Preserve ToneMark and reset structure
	tone := e.buffer.syllable.ToneMark
	e.buffer.syllable = &Syllable{Raw: raw, ToneMark: tone}

	runes := []rune(raw)
	onset := ""
	nucleus := ""
	coda := ""
	i := 0

	// Parse onset
	for i < len(runes) {
		r := runes[i]
		if isVietnameseVowelRune(r) {
			break
		}

		// Double-consonant 'dd' -> 'đ'
		if (r == 'd' || r == 'D') && i+1 < len(runes) && (runes[i+1] == 'd' || runes[i+1] == 'D') {
			if r == 'd' {
				onset += "đ"
			} else {
				onset += "Đ"
			}
			i += 2
			continue
		}

		// Only skip keys that are DEFINITELY not consonants in onset
		if r == 'f' || r == 'F' || r == 'j' || r == 'J' || r == 'z' || r == 'Z' || r == 'w' || r == 'W' {
			i++
			continue
		}

		if isVietnameseConsonantRune(r) {
			onset += string(r)
			i++
		} else {
			break
		}
	}

	// Parse nucleus
	for i < len(runes) {
		r := runes[i]
		if isVietnameseVowelRune(r) {
			// Double-vowel transformations
			if i+1 < len(runes) && unicode.ToLower(runes[i+1]) == unicode.ToLower(r) {
				var transformed rune
				switch unicode.ToLower(r) {
				case 'a':
					transformed = 'â'
				case 'e':
					transformed = 'ê'
				case 'o':
					transformed = 'ô'
				}
				if transformed != 0 {
					if unicode.IsUpper(r) {
						nucleus += string(unicode.ToUpper(transformed))
					} else {
						nucleus += string(transformed)
					}
					i += 2
					continue
				}
			}
			nucleus += string(r)
			i++
		} else if unicode.ToLower(r) == 'w' {
			// Horn/Breve modifier
			if len(nucleus) > 0 {
				nucleusRunes := []rune(nucleus)
				lastIdx := len(nucleusRunes) - 1
				last := nucleusRunes[lastIdx]

				var transformed rune
				switch unicode.ToLower(last) {
				case 'a':
					transformed = 'ă'
				case 'o':
					// Special case: uo + w -> ươ
					if len(nucleusRunes) >= 2 && unicode.ToLower(nucleusRunes[len(nucleusRunes)-2]) == 'u' {
						u := nucleusRunes[len(nucleusRunes)-2]
						transformedU := 'ư'
						if unicode.IsUpper(u) {
							transformedU = 'Ư'
						}
						nucleusRunes[len(nucleusRunes)-2] = transformedU
					}
					transformed = 'ơ'
				case 'u':
					transformed = 'ư'
				}

				if transformed != 0 {
					if unicode.IsUpper(last) {
						nucleusRunes[lastIdx] = unicode.ToUpper(transformed)
					} else {
						nucleusRunes[lastIdx] = transformed
					}
					nucleus = string(nucleusRunes)
				}
			}
			i++
		} else if isTelexModifier(r) {
			i++
		} else {
			break
		}
	}

	// Parse coda
	for i < len(runes) {
		r := runes[i]
		if isTelexModifier(r) {
			i++
			continue
		}
		if isVietnameseConsonantRune(r) {
			// 2-character coda
			if i+1 < len(runes) {
				nextR := runes[i+1]
				if isVietnameseConsonantRune(nextR) && isValidCoda(string(r)+string(nextR)) {
					coda += string(r) + string(nextR)
					i += 2
					continue
				}
			}
			if isValidCoda(string(r)) {
				coda += string(r)
				i++
			} else {
				break
			}
		} else {
			break
		}
	}

	// Rule: Automatic vowel mark transformation for ia/ua/ia patterns followed by a coda.
	// E.g. i + e + n -> iên, u + o + n -> uôn
	if coda != "" && len(nucleus) >= 2 {
		nRunes := []rune(nucleus)
		first := unicode.ToLower(nRunes[0])
		second := unicode.ToLower(nRunes[1])

		// 'ia' + coda -> iê (tiền, tiếng)
		if first == 'i' && second == 'e' {
			if unicode.IsUpper(nRunes[1]) {
				nRunes[1] = 'Ê'
			} else {
				nRunes[1] = 'ê'
			}
			nucleus = string(nRunes)
		}
		// 'ua' + coda -> uô (buồn, muốn)
		if first == 'u' && second == 'o' {
			if unicode.IsUpper(nRunes[1]) {
				nRunes[1] = 'Ô'
			} else {
				nRunes[1] = 'ô'
			}
			nucleus = string(nRunes)
		}
	}

	// Important: Skip ANY remaining Telex modifiers in the raw buffer.
	// This prevents modifiers that were consumed but didn't fit the
	// strict syllable structure from being shown as unparsed suffixes.
	for i < len(runes) {
		if isTelexModifier(runes[i]) {
			i++
		} else {
			break
		}
	}

	e.buffer.syllable.Onset = onset
	e.buffer.syllable.Nucleus = nucleus
	e.buffer.syllable.Coda = coda
	e.buffer.syllable.Consumed = i

	// Tone and VowelMark are already set in processChar
}

// isMarkedVowelRune checks if a rune is a vowel with diacritic mark
func isMarkedVowelRune(r rune) bool {
	switch r {
	case 'ă', 'Ă', 'â', 'Â', 'ê', 'Ê', 'ô', 'Ô', 'ơ', 'Ơ', 'ư', 'Ư':
		return true
	}
	return false
}

// isVietnameseVowelRune checks if a rune is a Vietnamese vowel.
func isVietnameseVowelRune(r rune) bool {
	lower := unicode.ToLower(r)
	switch lower {
	case 'a', 'ă', 'â', 'e', 'ê', 'i', 'o', 'ô', 'ơ', 'u', 'ư', 'y':
		return true
	}
	return false
}

// isVietnameseConsonantRune checks if a rune is a consonant.
func isVietnameseConsonantRune(r rune) bool {
	lower := unicode.ToLower(r)
	switch lower {
	case 'b', 'c', 'd', 'đ', 'g', 'h', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'x':
		return true
	}
	return false
}

var validCodas = map[string]bool{
	"c": true, "ch": true, "m": true, "n": true,
	"ng": true, "nh": true, "p": true, "t": true,
}

// isValidCoda checks if a consonant can be a valid coda in Vietnamese.
func isValidCoda(s string) bool {
	lower := strings.ToLower(s)
	return validCodas[lower]
}

// isTelexModifier checks if a character is a Telex modifier key.
// These characters are consumed as modifiers and should be skipped when parsing raw buffer.
func isTelexModifier(r rune) bool {
	switch r {
	case 's', 'S', 'f', 'F', 'r', 'R', 'x', 'X', 'j', 'J', 'z', 'Z', 'w', 'W':
		return true
	}
	return false
}

// keysymToRune converts an X11 keysym to a rune.
func keysymToRune(keysym uint32) rune {
	// ASCII printable characters (0x20 - 0x7E)
	if keysym >= 0x0020 && keysym <= 0x007e {
		return rune(keysym)
	}

	// Latin-1 supplement (0xA0 - 0xFF)
	if keysym >= 0x00a0 && keysym <= 0x00ff {
		return rune(keysym)
	}

	// Unicode keysyms (0x01000000 + unicode codepoint)
	if keysym >= 0x01000000 {
		return rune(keysym - 0x01000000)
	}

	return 0
}
