# GoViet IME - Frontend (Fcitx5 Engine)

This is the C++ plugin component for Fcitx5. It acts as the frontend layer that captures key events from the system and sends them to the Backend for processing via the DBus protocol.

## 1. System Requirements

To build and run this plugin, you need to install the following dependencies (example for Arch Linux):
```bash
sudo pacman -S extra-cmake-modules fcitx5 fcitx5-qt fcitx5-gtk dbus pkgconf gcc
```

## 2. Build and Installation Guide

1. **Navigate to the frontend directory:**
   ```bash
   cd frontend
   ```

2. **Create build directory and compile:**
   ```bash
   mkdir -p build && cd build
   cmake ..
   make -j$(nproc)
   ```

3. **Install to the system (requires root privileges):**
   ```bash
   sudo make install
   ```
   *Note: This command installs the .so file to `/usr/lib/fcitx5/` and configuration files to `/usr/share/fcitx5/`.*

## 3. How to Activate the Input Method

After installation, follow these steps to make GoViet appear in your list:

1. **Refresh Fcitx5 data:**
   ```bash
   fcitx5 -rd
   ```
   *(This command forces Fcitx5 to recognize new configuration files without requiring a logout)*

2. **Open Fcitx5 Configuration Tool:**
   ```bash
   fcitx5-configtool
   ```

3. **Add GoViet Input Method:**
   - Click the **Add Input Method** button (+ icon).
   - Search for **GoViet**.
   - Select it and click OK to add it to your active input methods.

## 4. Important Installed Files

- `/usr/lib/fcitx5/libgo-viet-engine.so`: The main engine library.
- `/usr/share/fcitx5/addon/goviet.conf`: Addon registration for Fcitx5.
- `/usr/share/fcitx5/inputmethod/goviet.conf`: Input Method registration (makes it visible in the selection list).
