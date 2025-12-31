package engine

import (
	"testing"
)

// Integration tests for Vietnamese word typing with Telex input method.

func TestVietnameseWords_Simple(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint32 // keysyms
		expected string
	}{
		{"xin", []uint32{0x78, 0x69, 0x6e}, "xin"},         // x-i-n
		{"chao", []uint32{0x63, 0x68, 0x61, 0x6f}, "chao"}, // c-h-a-o
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewCompositionEngine()
			var result ProcessResult
			for _, ks := range tt.input {
				result = engine.ProcessKey(KeyEvent{KeySym: ks})
			}
			if result.Preedit != tt.expected {
				t.Errorf("Typing %s: got '%s', want '%s'", tt.name, result.Preedit, tt.expected)
			}
		})
	}
}

func TestVietnameseWords_WithTones(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint32 // keysyms
		expected string
	}{
		// Sắc tone (s)
		{"xin chào -> xin (á with sac)", []uint32{0x61, 0x73}, "á"}, // a-s
		// Huyền tone (f)
		{"huyền (à)", []uint32{0x61, 0x66}, "à"}, // a-f
		// Hỏi tone (r)
		{"hỏi (ả)", []uint32{0x61, 0x72}, "ả"}, // a-r
		// Ngã tone (x)
		{"ngã (ã)", []uint32{0x61, 0x78}, "ã"}, // a-x
		// Nặng tone (j)
		{"nặng (ạ)", []uint32{0x61, 0x6a}, "ạ"}, // a-j
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewCompositionEngine()
			var result ProcessResult
			for _, ks := range tt.input {
				result = engine.ProcessKey(KeyEvent{KeySym: ks})
			}
			if result.Preedit != tt.expected {
				t.Errorf("Typing %s: got '%s', want '%s'", tt.name, result.Preedit, tt.expected)
			}
		})
	}
}

func TestVietnameseWords_WithVowelMarks(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint32 // keysyms
		expected string
	}{
		// Double letter patterns
		{"aa -> â", []uint32{0x61, 0x61}, "â"},
		{"ee -> ê", []uint32{0x65, 0x65}, "ê"},
		{"oo -> ô", []uint32{0x6f, 0x6f}, "ô"},
		{"dd -> đ", []uint32{0x64, 0x64}, "đ"},
		// Horn with w
		{"ow -> ơ", []uint32{0x6f, 0x77}, "ơ"},
		{"uw -> ư", []uint32{0x75, 0x77}, "ư"},
		{"aw -> ă", []uint32{0x61, 0x77}, "ă"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewCompositionEngine()
			var result ProcessResult
			for _, ks := range tt.input {
				result = engine.ProcessKey(KeyEvent{KeySym: ks})
			}
			if result.Preedit != tt.expected {
				t.Errorf("Typing %s: got '%s', want '%s'", tt.name, result.Preedit, tt.expected)
			}
		})
	}
}

func TestVietnameseWords_ComplexWords(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint32 // keysyms
		expected string
	}{
		// "việt" = v-i-e-e-t-j (viêt with nặng)
		// But the tone is applied to ê, so: viêj -> viêt̤
		// Actually in Telex: việt = v-i-ee-j-t
		// "việt" = v-i-ệ-t
		{"viet", []uint32{0x76, 0x69, 0x65, 0x74}, "viet"},

		// "đẹp" = d-d-e-j-p
		{"ddejp (đẹp)", []uint32{0x64, 0x64, 0x65, 0x6a, 0x70}, "đẹp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewCompositionEngine()
			var result ProcessResult
			for _, ks := range tt.input {
				result = engine.ProcessKey(KeyEvent{KeySym: ks})
			}
			if result.Preedit != tt.expected {
				t.Errorf("Typing %s: got '%s', want '%s'", tt.name, result.Preedit, tt.expected)
			}
		})
	}
}

func TestVietnameseWords_SpaceCommits(t *testing.T) {
	engine := NewCompositionEngine()

	// Type "xin"
	engine.ProcessKey(KeyEvent{KeySym: 0x78}) // x
	engine.ProcessKey(KeyEvent{KeySym: 0x69}) // i
	engine.ProcessKey(KeyEvent{KeySym: 0x6e}) // n

	// Press space to commit
	result := engine.ProcessKey(KeyEvent{KeySym: KeySpace})

	if result.CommitText != "xin " {
		t.Errorf("CommitText = '%s', want 'xin '", result.CommitText)
	}
	if result.Preedit != "" {
		t.Errorf("Preedit after space = '%s', want ''", result.Preedit)
	}

	// Type "chào" = c-h-a-f-o
	engine.ProcessKey(KeyEvent{KeySym: 0x63})          // c
	engine.ProcessKey(KeyEvent{KeySym: 0x68})          // h
	engine.ProcessKey(KeyEvent{KeySym: 0x61})          // a
	engine.ProcessKey(KeyEvent{KeySym: 0x66})          // f (huyền tone)
	result = engine.ProcessKey(KeyEvent{KeySym: 0x6f}) // o

	// Preedit should be "chào"
	if result.Preedit != "chào" {
		t.Errorf("Preedit = '%s', want 'chào'", result.Preedit)
	}
}
