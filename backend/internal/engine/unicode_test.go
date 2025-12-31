package engine

import (
	"testing"
)

func TestUnicodeFormat_ApplyTone(t *testing.T) {
	format := NewUnicodeFormat()

	tests := []struct {
		name     string
		vowel    rune
		tone     ToneMark
		expected string
	}{
		{"a with sac", 'a', ToneSac, "á"},
		{"a with huyen", 'a', ToneHuyen, "à"},
		{"a with hoi", 'a', ToneHoi, "ả"},
		{"a with nga", 'a', ToneNga, "ã"},
		{"a with nang", 'a', ToneNang, "ạ"},
		{"a with none", 'a', ToneNone, "a"},
		{"e with sac", 'e', ToneSac, "é"},
		{"o with huyen", 'o', ToneHuyen, "ò"},
		{"u with hoi", 'u', ToneHoi, "ủ"},
		{"i with nga", 'i', ToneNga, "ĩ"},
		{"uppercase A with sac", 'A', ToneSac, "Á"},
		{"ă with sac", 'ă', ToneSac, "ắ"},
		{"â with huyen", 'â', ToneHuyen, "ầ"},
		{"ê with hoi", 'ê', ToneHoi, "ể"},
		{"ô with nga", 'ô', ToneNga, "ỗ"},
		{"ơ with nang", 'ơ', ToneNang, "ợ"},
		{"ư with sac", 'ư', ToneSac, "ứ"},
		{"y with huyen", 'y', ToneHuyen, "ỳ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := format.ApplyTone(tt.vowel, tt.tone)
			if result != tt.expected {
				t.Errorf("ApplyTone(%c, %v) = %s, want %s", tt.vowel, tt.tone, result, tt.expected)
			}
		})
	}
}

func TestUnicodeFormat_ApplyVowelMark(t *testing.T) {
	format := NewUnicodeFormat()

	tests := []struct {
		name     string
		char     rune
		mark     VowelMark
		expected string
	}{
		{"a with breve", 'a', VowelBreve, "ă"},
		{"a with hat", 'a', VowelHat, "â"},
		{"A with breve", 'A', VowelBreve, "Ă"},
		{"e with hat", 'e', VowelHat, "ê"},
		{"o with hat", 'o', VowelHat, "ô"},
		{"o with horn", 'o', VowelHorn, "ơ"},
		{"u with horn", 'u', VowelHorn, "ư"},
		{"d with dbar", 'd', VowelDBar, "đ"},
		{"D with dbar", 'D', VowelDBar, "Đ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := format.ApplyVowelMark(tt.char, tt.mark)
			if result != tt.expected {
				t.Errorf("ApplyVowelMark(%c, %v) = %s, want %s", tt.char, tt.mark, result, tt.expected)
			}
		})
	}
}

func TestGetBaseVowel(t *testing.T) {
	tests := []struct {
		name         string
		input        rune
		expectedBase rune
		expectedTone ToneMark
	}{
		{"á returns a, sac", 'á', 'a', ToneSac},
		{"à returns a, huyen", 'à', 'a', ToneHuyen},
		{"ả returns a, hoi", 'ả', 'a', ToneHoi},
		{"ã returns a, nga", 'ã', 'a', ToneNga},
		{"ạ returns a, nang", 'ạ', 'a', ToneNang},
		{"a returns a, none", 'a', 'a', ToneNone},
		{"ế returns ê, sac", 'ế', 'ê', ToneSac},
		{"ừ returns ư, huyen", 'ừ', 'ư', ToneHuyen},
		{"b returns b, none", 'b', 'b', ToneNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, tone := GetBaseVowel(tt.input)
			if base != tt.expectedBase || tone != tt.expectedTone {
				t.Errorf("GetBaseVowel(%c) = (%c, %v), want (%c, %v)",
					tt.input, base, tone, tt.expectedBase, tt.expectedTone)
			}
		})
	}
}

func TestIsVietnameseVowel(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{'a', true},
		{'e', true},
		{'i', true},
		{'o', true},
		{'u', true},
		{'y', true},
		{'ă', true},
		{'â', true},
		{'ê', true},
		{'ô', true},
		{'ơ', true},
		{'ư', true},
		{'á', true},
		{'ề', true},
		{'b', false},
		{'c', false},
		{'d', false},
		{'1', false},
		{' ', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := IsVietnameseVowel(tt.char)
			if result != tt.expected {
				t.Errorf("IsVietnameseVowel(%c) = %v, want %v", tt.char, result, tt.expected)
			}
		})
	}
}

func TestIsVietnameseConsonant(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{'b', true},
		{'c', true},
		{'d', true},
		{'đ', true},
		{'g', true},
		{'h', true},
		{'k', true},
		{'l', true},
		{'m', true},
		{'n', true},
		{'p', true},
		{'q', true},
		{'r', true},
		{'s', true},
		{'t', true},
		{'v', true},
		{'x', true},
		{'a', false},
		{'e', false},
		{'1', false},
		{' ', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := IsVietnameseConsonant(tt.char)
			if result != tt.expected {
				t.Errorf("IsVietnameseConsonant(%c) = %v, want %v", tt.char, result, tt.expected)
			}
		})
	}
}
