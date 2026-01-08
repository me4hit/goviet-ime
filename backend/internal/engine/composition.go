package engine

import (
	"strings"
	"unicode"
)

// TransformType indicates the type of transformation
type TransformType int

const (
	TransformNone      TransformType = iota
	TransformTone                    // Tone mark (s, f, r, x, j)
	TransformVowelMark               // Vowel mark (aa, ee, oo, w)
	TransformStroke                  // D-bar (dd)
	TransformWAsVowel                // W as ư
)

const breakMarker = '\u200b' // Zero-width space to break patterns (e.g. aa -> aa literal)

// LastTransform tracks the last transformation for double-key revert
type LastTransform struct {
	Key      rune          // The key that triggered the transform
	Type     TransformType // What kind of transform
	Position int           // Position in nucleus (for vowel marks)
	Original string        // Original value before transform
}

// CompositionEngine is the main engine that processes keyboard input.
type CompositionEngine struct {
	inputMethod   InputMethod
	outputFormat  OutputFormat
	buffer        *CompositionBuffer
	enabled       bool
	lastTransform LastTransform // For double-key revert
	config        *EngineConfig // Engine configuration
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
		inputMethod:   NewTelexMethod(), // Default to Telex
		outputFormat:  NewUnicodeFormat(),
		buffer:        NewCompositionBuffer(),
		enabled:       true,
		lastTransform: LastTransform{},
		config:        DefaultConfig(),
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
	e.lastTransform = LastTransform{}
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
	// but skip characters that are input method modifiers if they were likely already consumed.
	runes := []rune(raw)
	if syllable.Consumed < len(runes) && syllable.Consumed >= 0 {
		modifierIdx := 0
		for _, r := range runes[syllable.Consumed:] {
			// If it's a modifier, only skip it if it was "consumed" by the engine
			if e.isInputModifier(r) {
				if modifierIdx < syllable.ConsumedModifiers {
					modifierIdx++
					continue
				}
			}
			composed += string(r)
		}
	}

	if composed != "" {
		// Filter out pattern breakers
		return strings.ReplaceAll(composed, string(breakMarker), "")
	}

	return strings.ReplaceAll(raw, string(breakMarker), "")
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
	char := KeysymToRune(event.KeySym)
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
	// Check for double-key revert FIRST (before any processing)
	if e.config.EnableDoubleKeyRevert && e.checkDoubleKeyRevert(char) {
		return
	}

	transformed, tone, vowelMark, consumed := e.inputMethod.ProcessChar(char, e.buffer.syllable)

	if consumed {
		// VALIDATION FIRST: Check if we should apply transformation
		// Only validate if validation is enabled and we have a vowel (potential Vietnamese)
		if e.config.EnableValidation {
			// Check if the buffer forms valid Vietnamese before transformation
			if !e.shouldTransform(char) {
				// Not valid Vietnamese - treat as literal character
				e.buffer.raw.WriteRune(char)
				e.updateSyllableStructure()
				e.lastTransform = LastTransform{} // Clear last transform
				return
			}
		}

		if e.inputMethod.IsToneKey(char) {

			if e.buffer.syllable.ToneMark == tone && tone != ToneNone {
				e.buffer.syllable.ToneMark = ToneNone
			} else {
				e.buffer.syllable.ToneMark = tone
			}

			// Track this transformation for double-key revert
			e.lastTransform = LastTransform{
				Key:      char,
				Type:     TransformTone,
				Original: "", // We don't need to track original tone value for revert
			}
		} else if vowelMark != VowelNone || len(transformed) > 0 {
			// Save original for potential revert
			originalNucleus := e.buffer.syllable.Nucleus
			originalOnset := e.buffer.syllable.Onset

			// For VNI input method: we need to modify the raw buffer directly
			// because VNI modifiers (6,7,8,9) transform vowels but shouldn't appear in output
			if e.inputMethod.Name() == "VNI" && len(transformed) > 0 {
				// VNI: Replace the target character in raw buffer with transformed result
				e.applyVNITransformation(vowelMark, transformed, char)
			} else {
				e.applyVowelMark(vowelMark, transformed)
				e.buffer.raw.WriteRune(char)
			}

			// Track this transformation
			if vowelMark == VowelDBar {
				e.lastTransform = LastTransform{
					Key:      char,
					Type:     TransformStroke,
					Original: originalOnset,
				}
			} else {
				e.lastTransform = LastTransform{
					Key:      char,
					Type:     TransformVowelMark,
					Original: originalNucleus,
				}
			}
			e.updateSyllableStructure()
			return
		}
		e.buffer.raw.WriteRune(char)
		e.updateSyllableStructure()
	} else {
		// Handle W-as-Vowel feature
		if e.config.EnableWAsVowel && unicode.ToLower(char) == 'w' {
			if e.tryWAsVowel(char) {
				return
			}
		}

		e.processCharInternal(char)
		e.lastTransform = LastTransform{} // Clear since no transform happened
	}
}

// shouldTransform checks if the current buffer + modifier would form valid Vietnamese
func (e *CompositionEngine) shouldTransform(modifierKey rune) bool {
	syllable := e.buffer.syllable
	if syllable == nil {
		return false
	}

	// Must have at least a nucleus (vowel) to transform
	if syllable.Nucleus == "" {
		return false
	}

	// Validate the syllable structure
	result := ValidateVietnamese(syllable.Onset, syllable.Nucleus, syllable.Coda)
	return result.Valid
}

// checkDoubleKeyRevert checks if this key should revert the last transformation
// Returns true if revert was performed
func (e *CompositionEngine) checkDoubleKeyRevert(char rune) bool {
	last := e.lastTransform
	if last.Type == TransformNone {
		return false
	}

	// Check if same key pressed again
	if unicode.ToLower(last.Key) != unicode.ToLower(char) {
		return false
	}

	raw := e.buffer.raw.String()
	runes := []rune(raw)
	if len(runes) == 0 {
		return false
	}

	// Remove the last modifier key from raw buffer
	// (It's always the last one for these transforms)
	newRaw := string(runes[:len(runes)-1])

	switch last.Type {
	case TransformTone:
		// Revert tone: remove tone mark and restore literal
		e.buffer.syllable.ToneMark = ToneNone
		e.buffer.raw.Reset()
		e.buffer.raw.WriteString(newRaw)
		e.buffer.raw.WriteRune(char)
		e.updateSyllableStructure()
		e.lastTransform = LastTransform{}
		return true

	case TransformVowelMark, TransformStroke:
		// Revert vowel mark or stroke: break the pattern
		e.buffer.syllable.VowelMark = VowelNone
		e.buffer.raw.Reset()
		e.buffer.raw.WriteString(newRaw)
		e.buffer.raw.WriteRune(breakMarker)
		e.buffer.raw.WriteRune(char)
		e.updateSyllableStructure()
		e.lastTransform = LastTransform{}
		return true

	case TransformWAsVowel:
		// Revert ư -> w + literal w (already handled mostly)
		e.revertWAsVowel(char)
		e.lastTransform = LastTransform{}
		return true
	}

	return false
}

// revertVowelTransform reverts a vowel transformation (aa->â back to aa)
func (e *CompositionEngine) revertVowelTransform(char rune) {
	raw := e.buffer.raw.String()
	runes := []rune(raw)

	// Find and revert the transformed vowel
	// For patterns like aa -> â, we need to change â back to a
	for i := len(runes) - 1; i >= 0; i-- {
		r := runes[i]
		lower := unicode.ToLower(r)

		// Check if this is a transformed vowel that matches our key
		var baseVowel rune
		switch lower {
		case 'â', 'Â':
			if unicode.ToLower(char) == 'a' {
				baseVowel = 'a'
			}
		case 'ê', 'Ê':
			if unicode.ToLower(char) == 'e' {
				baseVowel = 'e'
			}
		case 'ô', 'Ô':
			if unicode.ToLower(char) == 'o' {
				baseVowel = 'o'
			}
		case 'ơ', 'Ơ', 'ư', 'Ư', 'ă', 'Ă':
			if unicode.ToLower(char) == 'w' {
				// For horn/breve, revert to base
				switch lower {
				case 'ơ', 'Ơ':
					baseVowel = 'o'
				case 'ư', 'Ư':
					baseVowel = 'u'
				case 'ă', 'Ă':
					baseVowel = 'a'
				}
			}
		}

		if baseVowel != 0 {
			// Find this vowel in raw (for Telex) or already transformed in runes
			// In Telex, "â" is "aa". We want "aa" -> "aaa"
			// Actually, if it's already "â" in runes (VNI), we revert it.
			// If it's Telex, runes still contains "aa".

			if r == 'â' || r == 'Â' || r == 'ê' || r == 'Ê' || r == 'ô' || r == 'Ô' ||
				r == 'ơ' || r == 'Ơ' || r == 'ư' || r == 'Ư' || r == 'ă' || r == 'Ă' {
				// VNI Case: already transformed character in raw
				if unicode.IsUpper(r) {
					runes[i] = unicode.ToUpper(baseVowel)
				} else {
					runes[i] = baseVowel
				}
			} else {
				// Telex Case: the transformation is implicit.
				// We don't need to change anything in runes, just append the new char
				// so "aa" + "a" -> "aaa"
			}
			break
		}
	}

	// Rebuild buffer and add the new character
	e.Reset()
	for _, r := range runes {
		e.buffer.raw.WriteRune(r)
	}
	e.buffer.raw.WriteRune(char)
	e.updateSyllableStructure()
}

// revertStrokeTransform reverts đ -> dd
func (e *CompositionEngine) revertStrokeTransform(char rune) {
	raw := e.buffer.raw.String()
	runes := []rune(raw)

	for i := len(runes) - 1; i >= 0; i-- {
		r := runes[i]
		if r == 'đ' {
			runes[i] = 'd'
			break
		} else if r == 'Đ' {
			runes[i] = 'D'
			break
		}
	}

	e.Reset()
	for _, r := range runes {
		e.buffer.raw.WriteRune(r)
	}
	e.buffer.raw.WriteRune(char)
	e.updateSyllableStructure()
}

// tryWAsVowel attempts to use 'w' as the vowel 'ư'
// This is for cases like single 'w' at start of syllable
func (e *CompositionEngine) tryWAsVowel(char rune) bool {
	syllable := e.buffer.syllable

	// Only works if no nucleus yet
	if syllable.Nucleus != "" {
		return false
	}

	// Try adding 'ư' as nucleus
	testNucleus := "ư"
	if unicode.IsUpper(char) {
		testNucleus = "Ư"
	}

	// Validate if this would be valid Vietnamese
	result := ValidateVietnamese(syllable.Onset, testNucleus, syllable.Coda)
	if !result.Valid {
		return false
	}

	// Apply the W as vowel
	e.buffer.raw.WriteRune(char)
	e.buffer.syllable.Nucleus = testNucleus
	e.buffer.syllable.VowelMark = VowelHorn
	e.updateSyllableStructure()

	// Track for potential revert
	e.lastTransform = LastTransform{
		Key:  char,
		Type: TransformWAsVowel,
	}

	return true
}

// revertWAsVowel reverts ư back to w
func (e *CompositionEngine) revertWAsVowel(char rune) {
	// Simply treat w as literal
	e.buffer.raw.WriteRune(char)
	e.buffer.syllable.Nucleus = ""
	e.buffer.syllable.VowelMark = VowelNone
	e.updateSyllableStructure()
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

// applyVNITransformation handles VNI vowel mark transformation
// It modifies the raw buffer directly, replacing the target character with the transformed result
func (e *CompositionEngine) applyVNITransformation(mark VowelMark, transformed string, modifierKey rune) {
	e.buffer.syllable.VowelMark = mark

	if len(transformed) == 0 {
		return
	}

	raw := e.buffer.raw.String()
	runes := []rune(raw)
	transformedRunes := []rune(transformed)

	// Handle đ (stroke) - find 'd' and replace
	if mark == VowelDBar {
		for i := len(runes) - 1; i >= 0; i-- {
			r := runes[i]
			if r == 'd' || r == 'D' {
				runes[i] = transformedRunes[0]
				break
			}
		}
	} else if len(transformedRunes) == 1 {
		// Single character transformation (â, ê, ô, ơ, ư, ă)
		// Find the last vowel that matches and replace it
		targetTransformed := transformedRunes[0]

		// Determine what base vowel we're looking for based on the transformed result
		var targetBases []rune
		switch unicode.ToLower(targetTransformed) {
		case 'â':
			targetBases = []rune{'a', 'A'}
		case 'ê':
			targetBases = []rune{'e', 'E'}
		case 'ô':
			targetBases = []rune{'o', 'O'}
		case 'ơ':
			targetBases = []rune{'o', 'O'}
		case 'ư':
			targetBases = []rune{'u', 'U'}
		case 'ă':
			targetBases = []rune{'a', 'A'}
		}

		// Find and replace the last matching vowel
		for i := len(runes) - 1; i >= 0; i-- {
			r := runes[i]
			for _, target := range targetBases {
				if r == target {
					if unicode.IsUpper(r) {
						runes[i] = unicode.ToUpper(targetTransformed)
					} else {
						runes[i] = targetTransformed
					}
					goto done
				}
			}
		}
	done:
	} else if len(transformedRunes) > 1 {
		// Multi-character transformation (ươ for UO compound)
		// Find the UO pattern and replace both
		for i := 0; i < len(runes)-1; i++ {
			if unicode.ToLower(runes[i]) == 'u' && unicode.ToLower(runes[i+1]) == 'o' {
				if unicode.IsUpper(runes[i]) {
					runes[i] = 'Ư'
				} else {
					runes[i] = 'ư'
				}
				if unicode.IsUpper(runes[i+1]) {
					runes[i+1] = 'Ơ'
				} else {
					runes[i+1] = 'ơ'
				}
				break
			}
		}
	}

	// Rebuild the raw buffer with transformed content
	e.buffer.raw.Reset()
	for _, r := range runes {
		e.buffer.raw.WriteRune(r)
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
	vMark := e.buffer.syllable.VowelMark
	e.buffer.syllable = &Syllable{Raw: raw, ToneMark: tone, VowelMark: vMark}

	consumedModifiers := 0
	if tone != ToneNone {
		consumedModifiers++
	}
	// Note: vowel mark modifiers are also tracked below

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

		if r == breakMarker {
			i++
			continue
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
		if r == 'f' || r == 'F' || r == 'j' || r == 'J' || r == 'z' || r == 'Z' {
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
		if r == breakMarker {
			i++
			continue
		}
		if isVietnameseVowelRune(r) {
			// Double-vowel transformations
			if i+1 < len(runes) && unicode.ToLower(runes[i+1]) == unicode.ToLower(r) && runes[i+1] != breakMarker {
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
			// Handle 'w' as vowel 'ư'
			if len(nucleus) == 0 && e.config.EnableWAsVowel {
				if unicode.IsUpper(r) {
					nucleus += "Ư"
				} else {
					nucleus += "ư"
				}
				i++
				consumedModifiers++
				continue
			}

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
					consumedModifiers++
				}
			}
			i++
		} else {
			break
		}
	}

	// Parse coda
	for i < len(runes) {
		r := runes[i]
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
	nRunes := []rune(nucleus)
	if coda != "" && len(nRunes) >= 2 {
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

	// Important: Skip as many modifiers as were actually consumed.
	consumedSoFar := 0
	for i < len(runes) {
		if e.isInputModifier(runes[i]) {
			if consumedSoFar < consumedModifiers {
				i++
				consumedSoFar++
			} else {
				break
			}
		} else {
			break
		}
	}

	e.buffer.syllable.Onset = onset
	e.buffer.syllable.Nucleus = nucleus
	e.buffer.syllable.Coda = coda
	e.buffer.syllable.Consumed = i
	e.buffer.syllable.ConsumedModifiers = consumedModifiers

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

// isVNIModifier checks if a character is a VNI modifier key (number 0-9)
func isVNIModifier(r rune) bool {
	return r >= '0' && r <= '9'
}

// isInputModifier checks if a character is a modifier for the given input method
func (e *CompositionEngine) isInputModifier(r rune) bool {
	if e.inputMethod == nil {
		return isTelexModifier(r)
	}

	switch e.inputMethod.Name() {
	case "VNI":
		return isVNIModifier(r)
	default:
		return isTelexModifier(r)
	}
}

// KeysymToRune converts an X11 keysym to a rune.
func KeysymToRune(keysym uint32) rune {
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
