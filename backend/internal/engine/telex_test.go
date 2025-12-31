package engine

import (
	"testing"
)

func TestTelexMethod_IsToneKey(t *testing.T) {
	telex := NewTelexMethod()

	tests := []struct {
		char     rune
		expected bool
	}{
		{'s', true},  // sac
		{'f', true},  // huyen
		{'r', true},  // hoi
		{'x', true},  // nga
		{'j', true},  // nang
		{'z', true},  // remove tone
		{'S', true},  // uppercase also works
		{'a', false}, // not a tone key
		{'b', false},
		{'1', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := telex.IsToneKey(tt.char)
			if result != tt.expected {
				t.Errorf("IsToneKey(%c) = %v, want %v", tt.char, result, tt.expected)
			}
		})
	}
}

func TestTelexMethod_GetToneMark(t *testing.T) {
	telex := NewTelexMethod()

	tests := []struct {
		char     rune
		expected ToneMark
	}{
		{'s', ToneSac},
		{'f', ToneHuyen},
		{'r', ToneHoi},
		{'x', ToneNga},
		{'j', ToneNang},
		{'z', ToneNone},
		{'a', ToneNone}, // not a tone key
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := telex.GetToneMark(tt.char)
			if result != tt.expected {
				t.Errorf("GetToneMark(%c) = %v, want %v", tt.char, result, tt.expected)
			}
		})
	}
}

func TestTelexMethod_IsVowelModifier(t *testing.T) {
	telex := NewTelexMethod()

	tests := []struct {
		char     rune
		expected bool
	}{
		{'w', true},  // horn (ư, ơ) and breve (ă)
		{'a', true},  // double for â
		{'e', true},  // double for ê
		{'o', true},  // double for ô
		{'d', true},  // double for đ
		{'s', false}, // tone key, not vowel modifier
		{'b', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := telex.IsVowelModifier(tt.char)
			if result != tt.expected {
				t.Errorf("IsVowelModifier(%c) = %v, want %v", tt.char, result, tt.expected)
			}
		})
	}
}

func TestTelexMethod_ProcessChar_ToneKeys(t *testing.T) {
	telex := NewTelexMethod()

	// Create a syllable with a vowel
	syllable := &Syllable{
		Raw:     "a",
		Nucleus: "a",
	}

	tests := []struct {
		name         string
		char         rune
		expectedTone ToneMark
		consumed     bool
	}{
		{"s applies sac", 's', ToneSac, true},
		{"f applies huyen", 'f', ToneHuyen, true},
		{"r applies hoi", 'r', ToneHoi, true},
		{"x applies nga", 'x', ToneNga, true},
		{"j applies nang", 'j', ToneNang, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, tone, _, consumed := telex.ProcessChar(tt.char, syllable)
			if consumed != tt.consumed {
				t.Errorf("ProcessChar(%c) consumed = %v, want %v", tt.char, consumed, tt.consumed)
			}
			if tone != tt.expectedTone {
				t.Errorf("ProcessChar(%c) tone = %v, want %v", tt.char, tone, tt.expectedTone)
			}
		})
	}
}

func TestTelexMethod_ProcessChar_DoubleLetters(t *testing.T) {
	telex := NewTelexMethod()

	tests := []struct {
		name     string
		raw      string
		char     rune
		expected string
		consumed bool
	}{
		{"aa -> â", "a", 'a', "â", true},
		{"AA -> Â", "A", 'A', "Â", true},
		{"ee -> ê", "e", 'e', "ê", true},
		{"oo -> ô", "o", 'o', "ô", true},
		{"dd -> đ", "d", 'd', "đ", true},
		{"ab not consumed", "a", 'b', "b", false}, // not a double pattern
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syllable := &Syllable{Raw: tt.raw}
			result, _, _, consumed := telex.ProcessChar(tt.char, syllable)
			if consumed != tt.consumed {
				t.Errorf("ProcessChar(%c) after %s: consumed = %v, want %v",
					tt.char, tt.raw, consumed, tt.consumed)
			}
			if consumed && result != tt.expected {
				t.Errorf("ProcessChar(%c) after %s = %s, want %s",
					tt.char, tt.raw, result, tt.expected)
			}
		})
	}
}

func TestTelexMethod_ProcessChar_HornWithW(t *testing.T) {
	telex := NewTelexMethod()

	tests := []struct {
		name     string
		nucleus  string
		expected string
		consumed bool
	}{
		{"ow -> ơ", "o", "ơ", true},
		{"uw -> ư", "u", "ư", true},
		{"aw -> ă", "a", "ă", true},
		{"Ow -> Ơ", "O", "Ơ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syllable := &Syllable{
				Raw:     tt.nucleus,
				Nucleus: tt.nucleus,
			}
			result, _, _, consumed := telex.ProcessChar('w', syllable)
			if consumed != tt.consumed {
				t.Errorf("ProcessChar('w') with nucleus %s: consumed = %v, want %v",
					tt.nucleus, consumed, tt.consumed)
			}
			if consumed && result != tt.expected {
				t.Errorf("ProcessChar('w') with nucleus %s = %s, want %s",
					tt.nucleus, result, tt.expected)
			}
		})
	}
}

func TestTelexMethod_Name(t *testing.T) {
	telex := NewTelexMethod()
	if telex.Name() != "Telex" {
		t.Errorf("Name() = %s, want Telex", telex.Name())
	}
}
