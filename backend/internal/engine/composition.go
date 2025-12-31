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

	// Check if we have any composed content
	hasComposedContent := syllable.Nucleus != "" || syllable.Onset != "" || syllable.Coda != ""

	// Try to compose the syllable
	if hasComposedContent {
		// Check if all raw characters are accounted for in the parsed structure
		parsedContent := syllable.Onset + syllable.Nucleus + syllable.Coda
		parsedLen := len([]rune(parsedContent))
		rawLen := len([]rune(raw))

		// Account for modifier characters that were consumed (tones, vowel marks)
		totalAccountedFor := parsedLen + e.buffer.modifierCount

		// Only use composed output if all characters are accounted for
		// OR if we have Telex modifiers applied
		hasModifiers := syllable.ToneMark != ToneNone || syllable.VowelMark != VowelNone

		if totalAccountedFor >= rawLen || hasModifiers {
			// Compose the syllable
			composed := e.outputFormat.Compose(syllable)

			// If Compose returns raw (no nucleus), manually build from onset
			if composed == syllable.Raw && syllable.Nucleus == "" && syllable.Onset != "" {
				composed = syllable.Onset
			}

			if composed != "" {
				// If we have unparsed trailing characters (not consumed as modifiers), append them
				if totalAccountedFor < rawLen {
					rawRunes := []rune(raw)
					composed += string(rawRunes[totalAccountedFor:])
				}
				return composed
			}
		}
	}

	// Fallback to raw input
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

	// Check for modifiers (Ctrl, Alt) - don't process these
	if event.Modifiers&(ModControl|ModMod1) != 0 {
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
		// Commit current composition and add newline
		preedit := e.GetPreedit()
		e.Reset()
		result.Handled = true
		result.CommitText = preedit + "\n"
		result.Preedit = ""
		return result, true

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

	// Re-parse the syllable
	e.buffer = NewCompositionBuffer()
	for _, r := range newRaw {
		e.processCharInternal(r)
	}

	result.Preedit = e.GetPreedit()
	return result
}

// processChar processes a regular character input.
func (e *CompositionEngine) processChar(char rune) ProcessResult {
	result := ProcessResult{Handled: true}

	// Use the input method to process the character
	transformed, tone, vowelMark, consumed := e.inputMethod.ProcessChar(char, e.buffer.syllable)

	if consumed {
		// The character was consumed as a modifier
		if tone != ToneNone {
			e.buffer.syllable.ToneMark = tone
			// Add the raw character for backspace support
			e.buffer.raw.WriteRune(char)
			e.buffer.modifierCount++
		}
		if vowelMark != VowelNone {
			e.applyVowelMark(vowelMark, transformed)
			// Add the raw character for backspace support
			e.buffer.raw.WriteRune(char)
			e.buffer.modifierCount++
		}
	} else {
		// Regular character - add to buffer
		e.processCharInternal(char)
	}

	result.Preedit = e.GetPreedit()
	return result
}

// processCharInternal adds a character to the buffer and updates the syllable.
func (e *CompositionEngine) processCharInternal(char rune) {
	e.buffer.raw.WriteRune(char)

	// Update syllable structure
	if e.buffer.syllable.Raw == "" {
		e.buffer.syllable = &Syllable{Raw: string(char)}
	} else {
		e.buffer.syllable.Raw += string(char)
	}

	// Determine if this is onset, nucleus, or coda
	e.updateSyllableStructure()
}

// applyVowelMark applies a vowel mark to the current syllable.
func (e *CompositionEngine) applyVowelMark(mark VowelMark, transformed string) {
	e.buffer.syllable.VowelMark = mark

	// Special case for đ - it modifies the onset consonant 'd', not a vowel
	if mark == VowelDBar && transformed != "" {
		// Check if onset ends with 'd' and replace with 'đ'
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
	if transformed != "" && len(e.buffer.syllable.Nucleus) > 0 {
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
		return
	}

	runes := []rune(raw)

	// Preserve transformed onset if we have VowelDBar applied
	preservedOnset := ""
	if e.buffer.syllable.VowelMark == VowelDBar && e.buffer.syllable.Onset != "" {
		// Check if the onset has been transformed to đ
		onsetRunes := []rune(e.buffer.syllable.Onset)
		if len(onsetRunes) > 0 {
			last := onsetRunes[len(onsetRunes)-1]
			if last == 'đ' || last == 'Đ' {
				preservedOnset = e.buffer.syllable.Onset
			}
		}
	}

	// Preserve transformed nucleus if we have VowelMark applied (like ô from oo)
	preservedNucleus := ""
	if e.buffer.syllable.VowelMark != VowelNone && e.buffer.syllable.VowelMark != VowelDBar {
		if e.buffer.syllable.Nucleus != "" {
			// Check if nucleus contains a marked vowel
			for _, r := range e.buffer.syllable.Nucleus {
				if isMarkedVowelRune(r) {
					preservedNucleus = e.buffer.syllable.Nucleus
					break
				}
			}
		}
	}

	onset := ""
	nucleus := ""
	coda := ""

	i := 0

	// Parse onset (initial consonants)
	for i < len(runes) {
		r := runes[i]
		if isVietnameseVowelRune(r) {
			break
		}
		if isVietnameseConsonantRune(r) {
			onset += string(r)
			i++
		} else {
			break
		}
	}

	// Parse nucleus (vowels) - skip consumed modifiers
	for i < len(runes) {
		r := runes[i]
		if isVietnameseVowelRune(r) {
			nucleus += string(r)
			i++
		} else if isTelexModifier(r) {
			// Skip consumed modifiers in the raw buffer
			i++
		} else {
			break
		}
	}

	// Parse coda (final consonants) - skip consumed modifiers
	for i < len(runes) {
		r := runes[i]
		if isTelexModifier(r) {
			// Skip consumed modifiers
			i++
			continue
		}
		if isVietnameseConsonantRune(r) && isValidCoda(string(r)) {
			coda += string(r)
			i++
		} else {
			break
		}
	}

	// Use preserved onset if we had a đ transformation
	if preservedOnset != "" {
		e.buffer.syllable.Onset = preservedOnset
	} else {
		e.buffer.syllable.Onset = onset
	}

	// Use preserved nucleus if we had a vowel mark transformation,
	// but add any new vowels that came after
	if preservedNucleus != "" {
		// The preserved nucleus has the marked vowel (like 'ê' from 'ee')
		// Raw nucleus contains all vowels including the consumed one
		// We need to figure out how many vowels were in the original pattern

		// Count non-modifier vowels in raw
		rawNucleusRunes := []rune(nucleus)

		// The modifier consumed one vowel char (e.g., 'ee' -> 'ê' means 2 raw chars -> 1 marked)
		// preserved has 1 or more marked vowels, need to add vowels that came AFTER the pattern

		preservedLen := len([]rune(preservedNucleus))

		// For patterns like 'oo'->ô, 'ee'->ê, 'aa'->â:
		// - Raw has 2 vowels, preserved has 1 marked vowel
		// - We consumed 1 extra vowel as modifier
		// For patterns like 'ow'->ơ, 'uw'->ư:
		// - Raw has 1 vowel + 1 modifier, preserved has 1 marked vowel

		// Calculate: how many raw vowels correspond to preserved?
		// If modifierCount > 0 for vowel marks, one vowel was consumed
		consumedVowelChars := 1 // The repeated vowel like 'oo', 'ee', 'aa'

		startIdx := preservedLen + consumedVowelChars
		if startIdx < len(rawNucleusRunes) {
			e.buffer.syllable.Nucleus = preservedNucleus + string(rawNucleusRunes[startIdx:])
		} else {
			e.buffer.syllable.Nucleus = preservedNucleus
		}
	} else {
		e.buffer.syllable.Nucleus = nucleus
	}

	e.buffer.syllable.Coda = coda
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

// isValidCoda checks if a consonant can be a valid coda in Vietnamese.
func isValidCoda(s string) bool {
	lower := strings.ToLower(s)
	validCodas := map[string]bool{
		"c": true, "ch": true, "m": true, "n": true,
		"ng": true, "nh": true, "p": true, "t": true,
	}
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
