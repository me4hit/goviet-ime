package engine

import (
	"testing"
)

// Tests for real-world Vietnamese typing scenarios discovered during testing.

func TestRealWorld_TonePosition(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint32 // keysyms
		expected string
	}{
		// "chào" = c-h-a-o-f (tone on 'a', not 'o')
		{"chao with huyen -> chào", []uint32{0x63, 0x68, 0x61, 0x6f, 0x66}, "chào"},
		// "xoá" = x-o-a-s (tone on 'a')
		{"xoa with sac -> xoá", []uint32{0x78, 0x6f, 0x61, 0x73}, "xoá"},
		// "hoà" = h-o-a-f
		{"hoa with huyen -> hoà", []uint32{0x68, 0x6f, 0x61, 0x66}, "hoà"},
		// "nghĩa" = n-g-h-i-a-x (tone on 'i', not 'a')
		{"nghia with nga -> nghĩa", []uint32{0x6e, 0x67, 0x68, 0x69, 0x61, 0x78}, "nghĩa"},
		// "thoả" = t-h-o-a-r (tone on 'a')
		{"thoa with hoi -> thoả", []uint32{0x74, 0x68, 0x6f, 0x61, 0x72}, "thoả"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewCompositionEngine()
			var result ProcessResult
			for _, ks := range tt.input {
				result = engine.ProcessKey(KeyEvent{KeySym: ks})
			}
			if result.Preedit != tt.expected {
				t.Errorf("got '%s', want '%s'", result.Preedit, tt.expected)
			}
		})
	}
}

func TestRealWorld_DoubleVowelWithSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint32 // keysyms
		expected string
	}{
		// "tôi" = t-o-o-i (oo->ô, then add i)
		{"tooi -> tôi", []uint32{0x74, 0x6f, 0x6f, 0x69}, "tôi"},
		// "mưa" = m-u-w-a (uw->ư)
		{"muwa -> mưa", []uint32{0x6d, 0x75, 0x77, 0x61}, "mưa"},
		// "lươn" = l-u-o-w-n (but this requires multiple marks - current limitation)
		// For now, test simpler case: "bơi" = b-o-w-i
		{"bowi -> bơi", []uint32{0x62, 0x6f, 0x77, 0x69}, "bơi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewCompositionEngine()
			var result ProcessResult
			for _, ks := range tt.input {
				result = engine.ProcessKey(KeyEvent{KeySym: ks})
			}
			if result.Preedit != tt.expected {
				t.Errorf("got '%s', want '%s'", result.Preedit, tt.expected)
			}
		})
	}
}

func TestRealWorld_CompleteWords(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint32 // keysyms
		expected string
	}{
		// "việt" = v-i-e-e-j-t
		{"vieetjt -> việt", []uint32{0x76, 0x69, 0x65, 0x65, 0x6a, 0x74}, "việt"},
		// "tiếng" = t-i-e-e-s-n-g
		{"tiếng", []uint32{0x74, 0x69, 0x65, 0x65, 0x73, 0x6e, 0x67}, "tiếng"},
		// "các" = c-a-c-s
		{"cacs -> các", []uint32{0x63, 0x61, 0x63, 0x73}, "các"},
		// "bạn" = b-a-n-j
		{"banj -> bạn", []uint32{0x62, 0x61, 0x6e, 0x6a}, "bạn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewCompositionEngine()
			var result ProcessResult
			for _, ks := range tt.input {
				result = engine.ProcessKey(KeyEvent{KeySym: ks})
			}
			if result.Preedit != tt.expected {
				t.Errorf("got '%s', want '%s'", result.Preedit, tt.expected)
			}
		})
	}
}

func TestRealWorld_ToneAfterCoda(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint32 // keysyms
		expected string
	}{
		// In Telex, you can type tone after the coda
		// "bạn" = b-a-n-j (tone after 'n')
		{"ban then j -> bạn", []uint32{0x62, 0x61, 0x6e, 0x6a}, "bạn"},
		// "các" = c-a-c-s (tone after 'c')
		{"cac then s -> các", []uint32{0x63, 0x61, 0x63, 0x73}, "các"},
		// "mát" = m-a-t-s
		{"mat then s -> mát", []uint32{0x6d, 0x61, 0x74, 0x73}, "mát"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewCompositionEngine()
			var result ProcessResult
			for _, ks := range tt.input {
				result = engine.ProcessKey(KeyEvent{KeySym: ks})
			}
			if result.Preedit != tt.expected {
				t.Errorf("got '%s', want '%s'", result.Preedit, tt.expected)
			}
		})
	}
}
