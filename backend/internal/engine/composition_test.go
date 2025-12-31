package engine

import (
	"testing"
)

func TestCompositionEngine_ProcessKey_BasicLetters(t *testing.T) {
	engine := NewCompositionEngine()

	tests := []struct {
		name            string
		keysym          uint32
		expectedHandled bool
		expectedPreedit string
	}{
		{"type a", 0x0061, true, "a"},
		{"type b after a", 0x0062, true, "ab"},
		{"type c after ab", 0x0063, true, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.ProcessKey(KeyEvent{KeySym: tt.keysym})
			if result.Handled != tt.expectedHandled {
				t.Errorf("ProcessKey(%x) Handled = %v, want %v",
					tt.keysym, result.Handled, tt.expectedHandled)
			}
			if result.Preedit != tt.expectedPreedit {
				t.Errorf("ProcessKey(%x) Preedit = %s, want %s",
					tt.keysym, result.Preedit, tt.expectedPreedit)
			}
		})
	}
}

func TestCompositionEngine_Reset(t *testing.T) {
	engine := NewCompositionEngine()

	// Type something
	engine.ProcessKey(KeyEvent{KeySym: 0x0061}) // a
	engine.ProcessKey(KeyEvent{KeySym: 0x0062}) // b

	if engine.GetPreedit() != "ab" {
		t.Fatalf("Expected preedit 'ab', got '%s'", engine.GetPreedit())
	}

	// Reset
	engine.Reset()

	if engine.GetPreedit() != "" {
		t.Errorf("After Reset(), preedit = '%s', want ''", engine.GetPreedit())
	}
}

func TestCompositionEngine_HandleSpace(t *testing.T) {
	engine := NewCompositionEngine()

	// Type "abc"
	engine.ProcessKey(KeyEvent{KeySym: 0x0061}) // a
	engine.ProcessKey(KeyEvent{KeySym: 0x0062}) // b
	engine.ProcessKey(KeyEvent{KeySym: 0x0063}) // c

	// Press space
	result := engine.ProcessKey(KeyEvent{KeySym: KeySpace})

	if !result.Handled {
		t.Error("Space should be handled")
	}
	if result.CommitText != "abc " {
		t.Errorf("CommitText = '%s', want 'abc '", result.CommitText)
	}
	if result.Preedit != "" {
		t.Errorf("Preedit after space = '%s', want ''", result.Preedit)
	}
}

func TestCompositionEngine_HandleBackspace(t *testing.T) {
	engine := NewCompositionEngine()

	// Type "abc"
	engine.ProcessKey(KeyEvent{KeySym: 0x0061}) // a
	engine.ProcessKey(KeyEvent{KeySym: 0x0062}) // b
	engine.ProcessKey(KeyEvent{KeySym: 0x0063}) // c

	// Press backspace
	result := engine.ProcessKey(KeyEvent{KeySym: KeyBackspace})

	if !result.Handled {
		t.Error("Backspace should be handled")
	}
	if result.Preedit != "ab" {
		t.Errorf("Preedit after backspace = '%s', want 'ab'", result.Preedit)
	}
}

func TestCompositionEngine_HandleEscape(t *testing.T) {
	engine := NewCompositionEngine()

	// Type something
	engine.ProcessKey(KeyEvent{KeySym: 0x0061}) // a
	engine.ProcessKey(KeyEvent{KeySym: 0x0062}) // b

	// Press escape
	result := engine.ProcessKey(KeyEvent{KeySym: KeyEscape})

	if !result.Handled {
		t.Error("Escape should be handled")
	}
	if result.CommitText != "" {
		t.Errorf("CommitText after escape = '%s', want ''", result.CommitText)
	}
	if engine.GetPreedit() != "" {
		t.Errorf("Preedit after escape = '%s', want ''", engine.GetPreedit())
	}
}

func TestCompositionEngine_HandleEnter(t *testing.T) {
	engine := NewCompositionEngine()

	// Type "ab"
	engine.ProcessKey(KeyEvent{KeySym: 0x0061}) // a
	engine.ProcessKey(KeyEvent{KeySym: 0x0062}) // b

	// Press enter - should commit preedit WITHOUT newline (let app handle Enter)
	result := engine.ProcessKey(KeyEvent{KeySym: KeyReturn})

	if !result.Handled {
		t.Error("Enter with preedit should be handled")
	}
	if result.CommitText != "ab" {
		t.Errorf("CommitText = '%s', want 'ab' (no newline)", result.CommitText)
	}
}

func TestCompositionEngine_HandleEnterEmpty(t *testing.T) {
	engine := NewCompositionEngine()

	// Press enter with empty buffer - should NOT be handled (pass through to app)
	result := engine.ProcessKey(KeyEvent{KeySym: KeyReturn})

	if result.Handled {
		t.Error("Enter on empty buffer should pass through to app")
	}
}

func TestCompositionEngine_IgnoreModifiers(t *testing.T) {
	engine := NewCompositionEngine()

	// Ctrl+A should not be handled
	result := engine.ProcessKey(KeyEvent{
		KeySym:    0x0061,
		Modifiers: ModControl,
	})

	if result.Handled {
		t.Error("Ctrl+key should not be handled")
	}

	// Alt+A should not be handled
	result = engine.ProcessKey(KeyEvent{
		KeySym:    0x0061,
		Modifiers: ModMod1,
	})

	if result.Handled {
		t.Error("Alt+key should not be handled")
	}
}

func TestCompositionEngine_Disabled(t *testing.T) {
	engine := NewCompositionEngine()

	// Disable engine
	engine.SetEnabled(false)

	if engine.IsEnabled() {
		t.Error("Engine should be disabled")
	}

	// Keys should not be handled
	result := engine.ProcessKey(KeyEvent{KeySym: 0x0061})

	if result.Handled {
		t.Error("Disabled engine should not handle keys")
	}

	// Re-enable
	engine.SetEnabled(true)

	if !engine.IsEnabled() {
		t.Error("Engine should be enabled")
	}
}

func TestCompositionEngine_SetInputMethod(t *testing.T) {
	engine := NewCompositionEngine()

	// Default is Telex
	telex := NewTelexMethod()
	engine.SetInputMethod(telex)

	// Type 'a' then 's' (should apply sac tone)
	engine.ProcessKey(KeyEvent{KeySym: 0x0061})           // a
	result := engine.ProcessKey(KeyEvent{KeySym: 0x0073}) // s

	if !result.Handled {
		t.Error("'s' after 'a' should be handled in Telex")
	}
	// The preedit should show "á" with sac tone
	preedit := engine.GetPreedit()
	if preedit != "á" {
		t.Errorf("Expected preedit 'á', got '%s'", preedit)
	}
}

func TestCompositionEngine_BackspaceOnEmpty(t *testing.T) {
	engine := NewCompositionEngine()

	// Backspace on empty buffer
	result := engine.ProcessKey(KeyEvent{KeySym: KeyBackspace})

	if result.Handled {
		t.Error("Backspace on empty buffer should not be handled (pass through)")
	}
}

func TestCompositionEngine_TabWithContent(t *testing.T) {
	engine := NewCompositionEngine()

	// Type something
	engine.ProcessKey(KeyEvent{KeySym: 0x0061}) // a

	// Tab should commit
	result := engine.ProcessKey(KeyEvent{KeySym: KeyTab})

	if !result.Handled {
		t.Error("Tab with content should be handled")
	}
	if result.CommitText != "a" {
		t.Errorf("CommitText = '%s', want 'a'", result.CommitText)
	}
}

func TestCompositionEngine_TabOnEmpty(t *testing.T) {
	engine := NewCompositionEngine()

	// Tab on empty buffer
	result := engine.ProcessKey(KeyEvent{KeySym: KeyTab})

	if result.Handled {
		t.Error("Tab on empty buffer should pass through")
	}
}

func TestKeysymToRune(t *testing.T) {
	tests := []struct {
		keysym   uint32
		expected rune
	}{
		{0x0061, 'a'},     // lowercase a
		{0x0041, 'A'},     // uppercase A
		{0x0020, ' '},     // space
		{0x0039, '9'},     // number
		{0x01000061, 'a'}, // Unicode keysym
		{0x01001EA1, 'ạ'}, // Unicode Vietnamese
		{0xff08, 0},       // Backspace (special key)
		{0x00, 0},         // Invalid
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			result := keysymToRune(tt.keysym)
			if result != tt.expected {
				t.Errorf("keysymToRune(%x) = %c (%x), want %c (%x)",
					tt.keysym, result, result, tt.expected, tt.expected)
			}
		})
	}
}
