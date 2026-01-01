# Project Context: GoViet-IME - Vietnamese Input Method Engine

## 1. Project Overview

A Vietnamese input method engine for **Fcitx5** on Linux, using **Go backend** with **C++ frontend**.

**Architecture:**
```
┌─────────────────────┐     D-Bus      ┌─────────────────────┐
│   Fcitx5 Frontend   │ ◄────────────► │   Go Backend        │
│   (C++ Plugin)      │                │   (goviet-daemon)   │
└─────────────────────┘                └─────────────────────┘
```

**Technology Stack:**
- **Frontend:** C++ (Fcitx5 Module API) - communicates via D-Bus
- **Backend:** Go (composition engine, Telex input method)
- **Build System:** CMake (frontend), Go modules (backend)

## 2. Current Status (2026-01-01)

### ✅ Completed
- [x] D-Bus communication between frontend and backend
- [x] Telex input method implementation
- [x] Unicode output format
- [x] Tone marks (sắc, huyền, hỏi, ngã, nặng)
- [x] Vowel marks (ă, â, ê, ô, ơ, ư, đ)
- [x] Double-letter patterns (aa→â, ee→ê, oo→ô, dd→đ)
- [x] Horn modifier (ow→ơ, uw→ư, aw→ă)
- [x] **Multiple vowel marks** - Words like "người" (ươ) now work correctly
- [x] **Multi-character coda** - Words like "càng", "tương" (ng coda) work correctly
- [x] Tone placement algorithm (quy tắc cũ)
- [x] Special key handling (backspace, space, enter, escape)
- [x] **Improved Special Keys** - Proper handling for `Ctrl+A`, `Delete`, and `Tab`
- [x] **Focus & Reset handling** - Auto-reset on window switch/focus change via D-Bus `Reset`
- [x] Comprehensive unit tests (75+ test cases)

- [x] **Undo tone** - Typing 'z' removes tone, double modifier toggles tone
- [x] **Improved Preedit Fallback** - Correctly handles mixed input (Vietnamese + unparsed English)
- [x] **Deterministic Re-parsing** - Syllable structure is rebuilt precisely from raw buffer
- [ ] **Undo vowel marks** - Repeating modifier key should undo transformation (e.g. `aaa` -> `aa`)
- [ ] **Word boundary detection** - Better handling of punctuation and numbers

### ❌ Not Started
- [ ] VNI input method
- [ ] VIQR input method
- [ ] Output format options (VNI Windows, TCVN3)
- [ ] Configuration via D-Bus
- [ ] Dictionary-based word prediction

## 3. Directory Structure

```
goviet-ime/
├── AI_CONTEXT.md           # This file
├── README.md               # Project overview
├── backend/                # Go backend
│   ├── cmd/daemon/
│   │   └── main.go         # D-Bus daemon entry point
│   ├── internal/engine/
│   │   ├── types.go        # Core types and interfaces
│   │   ├── composition.go  # Main composition engine (complex!)
│   │   ├── telex.go        # Telex input method
│   │   ├── unicode.go      # Unicode output format + tone rules
│   │   ├── *_test.go       # Unit tests
│   │   └── realworld_test.go # Real-world typing tests
│   ├── go.mod
│   └── goviet-daemon       # Compiled binary
├── frontend/               # C++ Fcitx5 addon
│   ├── src/
│   │   ├── engine.cpp      # Fcitx5 integration
│   │   ├── engine.h
│   │   └── main.cpp
│   ├── CMakeLists.txt
│   └── *.conf              # Fcitx5 addon config
└── protocol/               # (empty, for future protobuf/etc)
```

## 4. Key Components

### Backend Engine (`backend/internal/engine/`)

| File | Purpose |
|------|---------|
| `types.go` | Core types: KeyEvent, ProcessResult, Syllable, interfaces |
| `composition.go` | **MAIN FILE** - CompositionEngine, buffer management, syllable parsing |
| `telex.go` | Telex input method: tone keys (s,f,r,x,j,z), vowel modifiers |
| `unicode.go` | Unicode output: tone/vowel mappings, `findTonePosition` algorithm |

### Critical Functions to Understand

1. **`CompositionEngine.ProcessKey()`** - Entry point for all key events
2. **`updateSyllableStructure()`** - Parses raw buffer into onset/nucleus/coda
3. **`GetPreedit()`** - Composes final display string from syllable
4. **`findTonePosition()`** - Determines where to place tone mark (complex rules!)

## 5. Known Issues & Technical Debt

### Issue 1: VNI Input Method
**Status:** Not started.
**Plan:** Implement `InputMethod` interface for VNI. The `CompositionEngine` is now robust enough to support different methods by correctly interpreting the raw buffer.

### Issue 2: Tone Position Rules Configuration
**Status:** Old rule implemented.
**Plan:** Add a configuration option to switch between "old rule" and "new rule" for tone placement in `unicode.go`.

### Issue 3: Undo Vowel Marks
**Status:** Partially implemented.
**Plan:** Refactor `updateSyllableStructure` or the modifier logic to support "toggling" vowel marks when the same key is pressed multiple times (e.g., `aaa` -> `aa`).

## 6. Running the Project

### Backend Only (for testing)
```bash
cd backend
go test -v ./internal/engine/...  # Run all tests
./goviet-daemon                    # Start D-Bus daemon
```

### With Fcitx5 (full integration)
```bash
# Build frontend
cd frontend/build
cmake ..
make

# Install (needs sudo)
sudo make install

# Restart Fcitx5
fcitx5 -r
```

## 7. Test Commands

```bash
# All tests
go test ./internal/engine/...

# Specific test groups
go test -v -run TestRealWorld ./internal/engine/...    # Real-world scenarios
go test -v -run TestVietnamese ./internal/engine/...   # Vietnamese word tests
go test -v -run TestTelex ./internal/engine/...        # Telex method tests
go test -v -run TestUnicode ./internal/engine/...      # Unicode format tests
```

## 8. D-Bus Interface

**Service:** `com.github.goviet.ime`
**Object Path:** `/Engine`

### Methods
| Method | Input | Output | Notes |
|--------|-------|--------|-------|
| `ProcessKey` | (keysym uint32, modifiers uint32) | (handled, commit, preedit) | Now commits on Ctrl/Alt |
| `Reset` | () | () | Clears internal buffer immediately |
| `SetEnabled` | (enabled bool) | () | |
| `GetPreedit` | () | (preedit string) | |

## 9. Next Steps (Priority Order)

1. **Add VNI input method** - Popular alternative to Telex
2. **Implement Vowel Mark Undo** - Toggle marks on/off with repeating keys
3. **Add configuration** - Allow switching input methods, tone rules

## 10. Useful References

- [Fcitx5 Developer Docs](https://fcitx-im.org/wiki/Develop)
- [Vietnamese Tone Placement Rules](https://vi.wikipedia.org/wiki/Quy_tắc_đặt_dấu_thanh_trong_tiếng_Việt)
- [Telex Input Method](https://vi.wikipedia.org/wiki/Telex_(kiểu_gõ))