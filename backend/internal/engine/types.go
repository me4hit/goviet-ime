// Package engine provides the core input method engine for Vietnamese typing.
package engine

// KeyEvent represents a keyboard event from the frontend.
type KeyEvent struct {
	KeySym    uint32 // X11 keysym value
	Modifiers uint32 // Modifier state (Shift, Ctrl, Alt, etc.)
}

// ProcessResult contains the output from processing a key event.
type ProcessResult struct {
	Handled    bool   // Whether the key was consumed by the engine
	CommitText string // Text to commit to the application
	Preedit    string // Current preedit/composition string
}

// Modifier flags for keyboard state.
const (
	ModNone    uint32 = 0
	ModShift   uint32 = 1 << 0
	ModLock    uint32 = 1 << 1 // Caps Lock
	ModControl uint32 = 1 << 2
	ModMod1    uint32 = 1 << 3 // Alt
	ModMod4    uint32 = 1 << 6 // Super/Windows key
)

// Common keysym values for Vietnamese input.
const (
	KeyBackspace uint32 = 0xff08
	KeyReturn    uint32 = 0xff0d
	KeyEscape    uint32 = 0xff1b
	KeySpace     uint32 = 0x0020
	KeyTab       uint32 = 0xff09
	KeyDelete    uint32 = 0xffff

	// Lowercase letters
	KeyA uint32 = 0x0061
	KeyZ uint32 = 0x007a

	// Uppercase letters
	KeyShiftA uint32 = 0x0041
	KeyShiftZ uint32 = 0x005a

	// Numbers
	Key0 uint32 = 0x0030
	Key9 uint32 = 0x0039
)

// ToneMark represents Vietnamese tone marks.
type ToneMark int

const (
	ToneNone  ToneMark = iota // No tone (thanh ngang)
	ToneSac                   // Sắc (á)
	ToneHuyen                 // Huyền (à)
	ToneHoi                   // Hỏi (ả)
	ToneNga                   // Ngã (ã)
	ToneNang                  // Nặng (ạ)
)

// VowelMark represents Vietnamese vowel modifications.
type VowelMark int

const (
	VowelNone  VowelMark = iota
	VowelHat             // Circumflex (â, ê, ô)
	VowelBreve           // Breve (ă)
	VowelHorn            // Horn (ơ, ư)
	VowelDBar            // D-bar (đ)
)

// Syllable represents a Vietnamese syllable being composed.
type Syllable struct {
	Raw               string    // Raw input characters
	Onset             string    // Initial consonant(s) - phụ âm đầu
	Nucleus           string    // Vowel cluster - nguyên âm
	Coda              string    // Final consonant(s) - phụ âm cuối
	ToneMark          ToneMark  // Tone mark position
	VowelMark         VowelMark // Vowel modification
	Consumed          int       // How many characters from Raw were accounted for
	ConsumedModifiers int       // How many modifier keys were used in transformation
}

// Engine is the main interface for input method engines.
type Engine interface {
	// ProcessKey handles a key event and returns the result.
	ProcessKey(event KeyEvent) ProcessResult

	// Reset clears the current composition state.
	Reset()

	// GetPreedit returns the current preedit string.
	GetPreedit() string

	// SetInputMethod sets the typing method (e.g., Telex, VNI).
	SetInputMethod(method InputMethod)

	// SetOutputFormat sets the output encoding format.
	SetOutputFormat(format OutputFormat)
}

// InputMethod defines the interface for different typing methods.
type InputMethod interface {
	// Name returns the name of the input method (e.g., "Telex", "VNI").
	Name() string

	// ProcessChar processes a character and returns the transformation.
	// Returns (transformed string, tone mark, vowel mark, consumed).
	ProcessChar(char rune, current *Syllable) (string, ToneMark, VowelMark, bool)

	// IsToneKey checks if the character is used for tone marking.
	IsToneKey(char rune) bool

	// GetToneMark returns the tone mark for a given character.
	GetToneMark(char rune) ToneMark

	// IsVowelModifier checks if the character modifies a vowel.
	IsVowelModifier(char rune) bool

	// GetVowelMark returns the vowel mark for a given character.
	GetVowelMark(char rune) VowelMark
}

// OutputFormat defines the interface for different output encodings.
type OutputFormat interface {
	// Name returns the name of the output format.
	Name() string

	// Compose creates the final string from a syllable.
	Compose(syllable *Syllable) string

	// ApplyTone applies a tone mark to a vowel character.
	ApplyTone(vowel rune, tone ToneMark) string

	// ApplyVowelMark applies a vowel mark to a character.
	ApplyVowelMark(char rune, mark VowelMark) string
}
