# GoViet-IME Engine Backend

## Overview

This is the Go backend engine for GoViet-IME, a Vietnamese input method for Linux using Fcitx5. The engine communicates with the Fcitx5 frontend via D-Bus.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Engine Package                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────┐   ┌─────────────────────┐              │
│  │    InputMethod      │   │   OutputFormat      │              │
│  │    Interface        │   │   Interface         │              │
│  ├─────────────────────┤   ├─────────────────────┤              │
│  │ - TelexMethod       │   │ - UnicodeFormat     │              │
│  │ - VNIMethod (TODO)  │   │ - VNIFormat (TODO)  │              │
│  │ - VIQRMethod (TODO) │   │ - TCVN3 (TODO)      │              │
│  └─────────────────────┘   └─────────────────────┘              │
│            │                        │                           │
│            └────────┬───────────────┘                           │
│                     ▼                                           │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              CompositionEngine                          │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │ - ProcessKey(KeyEvent) ProcessResult                    │    │
│  │ - Reset()                                               │    │
│  │ - GetPreedit()                                          │    │
│  │ - SetInputMethod(InputMethod)                           │    │
│  │ - SetOutputFormat(OutputFormat)                         │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
backend/
├── cmd/
│   └── daemon/
│       └── main.go          # D-Bus daemon entry point
├── internal/
│   └── engine/
│       ├── types.go         # Core types and interfaces
│       ├── composition.go   # Main composition engine
│       ├── telex.go         # Telex input method
│       ├── unicode.go       # Unicode output format
│       ├── *_test.go        # Unit tests
├── go.mod
└── go.sum
```

## Key Components

### Types (`types.go`)

- **KeyEvent**: Represents keyboard input (keysym, modifiers)
- **ProcessResult**: Output from processing a key (handled, commitText, preedit)
- **Syllable**: Vietnamese syllable structure (onset, nucleus, coda, tones)
- **ToneMark**: Vietnamese tones (sắc, huyền, hỏi, ngã, nặng)
- **VowelMark**: Vowel modifications (ă, â, ê, ô, ơ, ư, đ)

### Interfaces

- **Engine**: Main interface for input method engines
- **InputMethod**: Interface for typing methods (Telex, VNI, etc.)
- **OutputFormat**: Interface for output encodings (Unicode, VNI, etc.)

### CompositionEngine (`composition.go`)

The main engine that:
- Processes keyboard input
- Manages composition buffer
- Coordinates between InputMethod and OutputFormat
- Handles special keys (backspace, space, enter, escape)

### TelexMethod (`telex.go`)

Implements the Telex typing method:
- Tone keys: `s` (sắc), `f` (huyền), `r` (hỏi), `x` (ngã), `j` (nặng), `z` (remove tone)
- Double letters: `aa` → â, `ee` → ê, `oo` → ô, `dd` → đ
- Horn modifier: `w` → ư/ơ/ă

### UnicodeFormat (`unicode.go`)

Implements Unicode (UTF-8) output:
- Maps base vowels + tone marks to Unicode characters
- Applies vowel marks (circumflex, breve, horn)
- Composes syllables with correct tone placement

## Usage

### Running the Daemon

```bash
cd backend
go build -o goviet-daemon ./cmd/daemon/
./goviet-daemon
```

### Running Tests

```bash
cd backend
go test -v ./internal/engine/...
```

## Extending the Engine

### Adding a New Input Method

1. Create a new file (e.g., `vni.go`)
2. Implement the `InputMethod` interface
3. Add unit tests

```go
type VNIMethod struct{}

func (v *VNIMethod) Name() string { return "VNI" }

func (v *VNIMethod) ProcessChar(char rune, current *Syllable) (string, ToneMark, VowelMark, bool) {
    // VNI uses numbers for tones: 1=sắc, 2=huyền, etc.
    // Implement transformation logic here
}

// Implement other interface methods...
```

### Adding a New Output Format

1. Create a new file (e.g., `vni_output.go`)
2. Implement the `OutputFormat` interface
3. Add character mapping tables

```go
type VNIOutputFormat struct{}

func (v *VNIOutputFormat) Name() string { return "VNI Windows" }

func (v *VNIOutputFormat) Compose(syllable *Syllable) string {
    // Convert to VNI encoding
}

// Implement other interface methods...
```

## D-Bus Interface

The daemon exposes the following D-Bus interface:

- **Service Name**: `com.github.goviet.ime`
- **Object Path**: `/Engine`

### Methods

- `ProcessKey(keysym uint32, modifiers uint32) → (handled bool, commitText string, preedit string)`
- `Reset()`
- `SetEnabled(enabled bool)`
- `GetPreedit() → (preedit string)`

## TODO

- [ ] Add VNI input method
- [ ] Add VIQR input method
- [ ] Add output format options (VNI, TCVN3)
- [ ] Add dictionary-based word prediction
- [ ] Add configuration support (via D-Bus)
- [ ] Add undo/redo support
