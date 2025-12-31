# GoViet IME

A Vietnamese Input Method Editor for Linux (Fcitx5), combining the performance of C++ (Frontend) with the flexibility of Go (Backend).

## Project Structure

- **`frontend/`**: C++ source code (Fcitx5 Engine). Handles key events and communicates via DBus.
- **`backend/`**: Go source code. Handles Vietnamese input logic and returns results via DBus.
- **`protocol/`**: Definitions for communication protocols between components.

## Quick Start

### 1. Build and Install Frontend
See detailed instructions in [frontend/README.md](./frontend/README.md).

### 2. Build and Run Backend
```bash
cd backend
go build -o goviet-daemon ./cmd/daemon/main.go
./goviet-daemon
```

### 3. Usage
1. Successfully install the Frontend.
2. Run the Backend (daemon).
3. Add **GoViet** in `fcitx5-configtool`.
4. Switch to GoViet and start typing.
