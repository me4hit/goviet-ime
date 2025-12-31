# GoViet-IME Engine Backend

Vietnamese input method engine written in Go for the GoViet-IME project.

## Quick Start

```bash
# Run tests
go test -v ./internal/engine/...

# Build daemon
go build -o goviet-daemon ./cmd/daemon/

# Run daemon
./goviet-daemon
```

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
│  │ ✅ TelexMethod      │   │ ✅ UnicodeFormat    │              │
│  │ ❌ VNIMethod        │   │ ❌ VNIFormat        │              │
│  │ ❌ VIQRMethod       │   │ ❌ TCVN3Format      │              │
│  └─────────────────────┘   └─────────────────────┘              │
│            │                        │                           │
│            └────────┬───────────────┘                           │
│                     ▼                                           │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              CompositionEngine                          │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │ - ProcessKey(KeyEvent) ProcessResult                    │    │
│  │ - Reset()                                               │    │
│  │ - GetPreedit() string                                   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
backend/
├── cmd/daemon/
│   └── main.go              # D-Bus daemon entry point
├── internal/engine/
│   ├── types.go             # Core types & interfaces
│   ├── composition.go       # Main composition engine
│   ├── telex.go             # Telex input method
│   ├── unicode.go           # Unicode output format
│   ├── composition_test.go  # Engine tests
│   ├── telex_test.go        # Telex tests
│   ├── unicode_test.go      # Unicode tests
│   ├── vietnamese_test.go   # Vietnamese word tests
│   └── realworld_test.go    # Real-world scenario tests
├── go.mod
├── go.sum
├── goviet-daemon            # Compiled binary
└── README.md                # This file
```

## Telex Input Method

| Key | Function | Example |
|-----|----------|---------|
| `s` | Sắc (acute) | `as` → á |
| `f` | Huyền (grave) | `af` → à |
| `r` | Hỏi (hook) | `ar` → ả |
| `x` | Ngã (tilde) | `ax` → ã |
| `j` | Nặng (dot) | `aj` → ạ |
| `z` | Remove tone | `ás` + `z` → as |
| `aa` | Circumflex | `aa` → â |
| `ee` | Circumflex | `ee` → ê |
| `oo` | Circumflex | `oo` → ô |
| `dd` | Stroke | `dd` → đ |
| `ow` | Horn | `ow` → ơ |
| `uw` | Horn | `uw` → ư |
| `aw` | Breve | `aw` → ă |

## Tone Placement Rules

Using "quy tắc cũ" (old/traditional rule):

| Pattern | Tone Position | Example |
|---------|---------------|---------|
| Single vowel | On that vowel | `án`, `ồ` |
| `oa`, `oe`, `uy` | Second vowel | `hoá`, `huỷ` |
| `ao`, `au`, `ay`, `ai` | First vowel | `chào`, `màu` |
| `ia` | First vowel | `nghĩa`, `mía` |
| `ua`, `ưa` | Second vowel | `mùa`, `lừa` |
| Marked vowel (ă,â,ê,ô,ơ,ư) | On marked | `việt`, `đường` |
| With coda | See rules | `oán`, `uyển` |

## Testing

```bash
# All tests
go test ./internal/engine/...

# With verbose output
go test -v ./internal/engine/...

# Specific test pattern
go test -v -run TestRealWorld ./internal/engine/...
go test -v -run TestVietnamese ./internal/engine/...
go test -v -run TestTelex ./internal/engine/...

# Coverage
go test -cover ./internal/engine/...
```

## Known Limitations

1. **Single VowelMark per syllable** - Words requiring multiple marks (người, lươn) not fully supported
2. **No undo functionality** - Can't undo tone with 'z' or double-modifier
3. **Telex-only** - VNI and VIQR not implemented yet

## D-Bus Interface

- **Service:** `com.github.goviet.ime`
- **Object Path:** `/Engine`
- **Methods:**
  - `ProcessKey(keysym uint32, modifiers uint32) → (handled bool, commit string, preedit string)`
  - `Reset()`
  - `SetEnabled(enabled bool)`
  - `GetPreedit() → (preedit string)`

## Extending

### Adding New Input Method

1. Create new file (e.g., `vni.go`)
2. Implement `InputMethod` interface
3. Add tests
4. Register in daemon

### Adding New Output Format

1. Create new file (e.g., `vni_output.go`)
2. Implement `OutputFormat` interface
3. Add character mapping tables
4. Add tests

## Contributing

See `AI_CONTEXT.md` in project root for detailed technical documentation and next steps.
