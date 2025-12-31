# GoViet-IME

Vietnamese Input Method Engine for Fcitx5 on Linux.

## Overview

GoViet-IME is a modern Vietnamese input method that combines a **Go backend** for processing with a **C++ frontend** for Fcitx5 integration. Communication between components uses D-Bus.

## Features

- ✅ **Telex input method** - Full support for Vietnamese typing
- ✅ **Unicode output** - Modern UTF-8 encoding
- ✅ **Real-time composition** - See characters as you type
- ✅ **Proper tone placement** - Following Vietnamese typography rules

## Quick Start

### Prerequisites

- Go 1.20+
- CMake 3.16+
- Fcitx5 development libraries
- D-Bus development libraries

### Build & Install

```bash
# Build backend
cd backend
go build -o goviet-daemon ./cmd/daemon/

# Build frontend
cd ../frontend
mkdir -p build && cd build
cmake ..
make
sudo make install

# Restart Fcitx5
fcitx5 -r
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

1. Words requiring multiple vowel marks (người, lươn) have limited support
2. Undo functionality not yet implemented

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
