# GoViet-IME

Vietnamese Input Method Engine for Fcitx5 on Linux.

## Overview

GoViet-IME is a modern Vietnamese input method that combines a **Go backend** for processing with a **C++ frontend** for Fcitx5 integration. Communication between components uses D-Bus.

## Features

- ✅ **Telex input method** - Full support for tone replacement, toggling, and removal ('z')
- ✅ **Proper tone placement** - Automatic placement based on Vietnamese grammar
- ✅ **Complex vowel support** - Handles words with multiple vowel marks like "người", "lươn"
- ✅ **Robust backspace** - Correctly reapplies diacritics when deleting characters

## Quick Start

### Prerequisites

- Go 1.20+
- CMake 3.16+
- **Fcitx5 development libraries** (`fcitx5`)
- **Extra CMake Modules** (`extra-cmake-modules`)
- **D-Bus development libraries** (`dbus`)
- **Base development tools** (`base-devel`, `pkgconf`)

On **Arch Linux**:
```bash
sudo pacman -S cmake extra-cmake-modules fcitx5 pkgconf go base-devel
```

### Build & Install

```bash
# Build backend
cd backend
go build -o goviet-daemon ./cmd/daemon/

# Build frontend
cd ../frontend
# If you moved the project, clear old cache: rm -rf build
mkdir -p build && cd build
cmake ..
make
sudo make install

# Restart Fcitx5 to load changes
fcitx5 -r &
```

### Run Backend (for development)

```bash
cd backend
./goviet-daemon
```

## Project Structure

```
goviet-ime/
├── AI_CONTEXT.md       # Detailed technical context for AI/developers
├── README.md           # This file
├── backend/            # Go composition engine
│   ├── cmd/daemon/     # D-Bus daemon
│   └── internal/engine # Core engine code
├── frontend/           # C++ Fcitx5 addon
│   └── src/            # Engine integration
└── protocol/           # Future: shared protocol definitions
```

## Documentation

- **[AI_CONTEXT.md](./AI_CONTEXT.md)** - Comprehensive technical documentation
- **[backend/README.md](./backend/README.md)** - Backend engine details
- **[frontend/README.md](./frontend/README.md)** - Fcitx5 integration details

## Current Status

| Component | Status |
|-----------|--------|
| Telex input | ✅ Working |
| Unicode output | ✅ Working |
| D-Bus communication | ✅ Working |
| Tone placement | ✅ Working |
| VNI input | ❌ Not started |
| Configuration | ❌ Not started |

## Known Issues

1. **Undo vowel marks** - Toggling vowel marks by repeating the modifier key (e.g., `aaa` -> `aa`)
2. **Dictionary support** - No built-in dictionary or word suggestions yet

See [AI_CONTEXT.md](./AI_CONTEXT.md) for detailed issue descriptions and proposed fixes.

## Testing

```bash
cd backend
go test -v ./internal/engine/...
```

## Contributing

This project uses AI-assisted development. The `AI_CONTEXT.md` file contains all necessary context for continuing development in a new AI session.

## License

MIT License
