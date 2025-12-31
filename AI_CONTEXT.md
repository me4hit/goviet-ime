# Project Context: Fcitx5 GoViet Engine

## 1. Project Overview
This project is a custom input method engine for **Fcitx5** on Linux.
**Goal:** To wrap a Go-based Vietnamese input logic (`goviet`) into a C++ shared library (`go-viet-engine.so`) that Fcitx5 can load as an addon.

**Technology Stack:**
* **Core:** C++ (Fcitx5 Module API).
* **Logic:** Go (Golang) compiled via `cgo` or linked as a static library.
* **Build System:** CMake.

## 2. Directory Structure & Key Files
```text
project_root/
├── CMakeLists.txt          # Build configuration
├── src/
│   ├── engine.cpp          # Main Fcitx5 C++ implementation
│   ├── engine.h            # Header file
│   └── (Go interface files)
└── addon/
    └── goviet.conf         # Fcitx5 addon registration