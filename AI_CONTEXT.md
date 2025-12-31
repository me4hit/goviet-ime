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
- [x] Comprehensive unit tests (75+ test cases)

### 🔄 In Progress / Known Issues
- [ ] **Undo tone/mark** - Typing 'z' should remove tone, double modifier should undo
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

### Issue 1: Single VowelMark Limitation
**Problem:** Current `Syllable` struct only supports one `VowelMark`:
```go
type Syllable struct {
    VowelMark VowelMark  // Only one!
}
```

**Impact:** Can't handle words like:
- "người" = ng + ư + ơ + i (needs BOTH ư and ơ)
- "lươn" = l + ư + ơ + n

**Proposed Fix:** Change to store marks per vowel position, or use a different representation:
```go
type Syllable struct {
    VowelMarks map[int]VowelMark  // Mark per position
    // OR
    Nucleus string  // Already transformed: "ươ" instead of "uo" + marks
}
```

### Issue 2: Preserved Nucleus Logic Complexity
**Location:** `composition.go:updateSyllableStructure()`

The logic for preserving transformed vowels (like oo→ô) when adding new characters is complex and fragile. Consider refactoring to:
1. Store transformed nucleus directly
2. Don't re-parse from raw buffer each time

### Issue 3: isTelexModifier Hardcoded
**Location:** `composition.go:isTelexModifier()`

This function is Telex-specific but lives in the generic composition engine. Should be moved to `InputMethod` interface.

### Issue 4: Tone Position Rules
**Location:** `unicode.go:findTonePosition()`

The Vietnamese tone placement rules are implemented for "quy tắc cũ" (old rule). Some users prefer "quy tắc mới" (new rule). This should be configurable.

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
| Method | Input | Output |
|--------|-------|--------|
| `ProcessKey` | (keysym uint32, modifiers uint32) | (handled bool, commit string, preedit string) |
| `Reset` | () | () |
| `SetEnabled` | (enabled bool) | () |
| `GetPreedit` | () | (preedit string) |

## 9. Next Steps (Priority Order)

1. **Fix multiple vowel marks** - Critical for words like "người", "lươn"
2. **Add undo functionality** - 'z' to remove tone, double modifier to undo
3. **Add VNI input method** - Popular alternative to Telex
4. **Add configuration** - Allow switching input methods, tone rules

## 10. Useful References

- [Fcitx5 Developer Docs](https://fcitx-im.org/wiki/Develop)
- [Vietnamese Tone Placement Rules](https://vi.wikipedia.org/wiki/Quy_tắc_đặt_dấu_thanh_trong_tiếng_Việt)
- [Telex Input Method](https://vi.wikipedia.org/wiki/Telex_(kiểu_gõ))