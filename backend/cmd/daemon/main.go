package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/godbus/dbus/v5"
	"github.com/username/goviet-ime/internal/engine"
)

const (
	serviceName = "com.github.goviet.ime"
	objectPath  = "/Engine"
)

// InputEngine is the D-Bus object that receives key events from Fcitx5.
type InputEngine struct {
	engine *engine.CompositionEngine
	logger *log.Logger
}

// NewInputEngine creates a new InputEngine with default settings.
func NewInputEngine(logger *log.Logger) *InputEngine {
	return &InputEngine{
		engine: engine.NewCompositionEngine(),
		logger: logger,
	}
}

// ProcessKey handles key events from Fcitx5 frontend.
// Input: keysym (X11 keycode), modifiers (Shift/Ctrl/Alt state)
// Output: handled (was key consumed), commitText (text to commit), preeditText (composition)
func (e *InputEngine) ProcessKey(keysym uint32, modifiers uint32) (bool, string, string, *dbus.Error) {
	event := engine.KeyEvent{
		KeySym:    keysym,
		Modifiers: modifiers,
	}

	result := e.engine.ProcessKey(event)

	// Log the key event and result
	if e.logger != nil {
		char := engine.KeysymToRune(keysym)
		keyStr := fmt.Sprintf("0x%x", keysym)
		if char != 0 {
			keyStr = fmt.Sprintf("%q", char)
		} else {
			// Handle special keys if they don't have a rune representation
			switch keysym {
			case engine.KeyBackspace:
				keyStr = "Backspace"
			case engine.KeySpace:
				keyStr = "Space"
			case engine.KeyReturn:
				keyStr = "Enter"
			case engine.KeyTab:
				keyStr = "Tab"
			case engine.KeyEscape:
				keyStr = "Esc"
			case engine.KeyDelete:
				keyStr = "Delete"
			case 0xff51:
				keyStr = "Left"
			case 0xff52:
				keyStr = "Up"
			case 0xff53:
				keyStr = "Right"
			case 0xff54:
				keyStr = "Down"
			case 0xff50:
				keyStr = "Home"
			case 0xff57:
				keyStr = "End"
			case 0xff55:
				keyStr = "PgUp"
			case 0xff56:
				keyStr = "PgDn"
			}
		}

		modsStr := ""
		if modifiers&engine.ModShift != 0 {
			modsStr += "Shift+"
		}
		if modifiers&engine.ModControl != 0 {
			modsStr += "Ctrl+"
		}
		if modifiers&engine.ModMod1 != 0 {
			modsStr += "Alt+"
		}

		e.logger.Printf("Type: %-15s | Preedit: %-15q | Commit: %-15q | Handled: %v",
			modsStr+keyStr, result.Preedit, result.CommitText, result.Handled)
	}

	return result.Handled, result.CommitText, result.Preedit, nil
}

// Reset clears the current composition state.
func (e *InputEngine) Reset() *dbus.Error {
	e.engine.Reset()
	fmt.Println(">>> [GoViet] Engine reset")
	return nil
}

// SetEnabled enables or disables the engine.
func (e *InputEngine) SetEnabled(enabled bool) *dbus.Error {
	e.engine.SetEnabled(enabled)
	fmt.Printf(">>> [GoViet] Engine enabled: %v\n", enabled)
	return nil
}

// GetPreedit returns the current preedit string.
func (e *InputEngine) GetPreedit() (string, *dbus.Error) {
	return e.engine.GetPreedit(), nil
}

func main() {
	// 1. Connect to Session Bus
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// 2. Register Service Name
	reply, err := conn.RequestName(serviceName, dbus.NameFlagDoNotQueue)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to request name:", err)
		os.Exit(1)
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "Name already taken - another instance may be running")
		os.Exit(1)
	}

	// 3. Setup Logging
	logFile, err := os.OpenFile("typing.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	var logger *log.Logger
	if err == nil {
		logger = log.New(logFile, "", log.LstdFlags)
		fmt.Println(">>> [GoViet] Logging to typing.log")
	} else {
		fmt.Fprintf(os.Stderr, ">>> [GoViet] Failed to open log file: %v\n", err)
	}
	defer logFile.Close()

	// 4. Create and export the engine
	inputEngine := NewInputEngine(logger)

	err = conn.Export(inputEngine, dbus.ObjectPath(objectPath), serviceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to export object:", err)
		os.Exit(1)
	}

	// 4. Print startup banner
	fmt.Println("================================================")
	fmt.Println("✅ GoViet-IME Backend is running!")
	fmt.Println("================================================")
	fmt.Printf("  Service:     %s\n", serviceName)
	fmt.Printf("  Object Path: %s\n", objectPath)
	fmt.Printf("  Input Method: Telex\n")
	fmt.Printf("  Output Format: Unicode\n")
	fmt.Println("------------------------------------------------")
	fmt.Println("Waiting for key events...")
	fmt.Println()

	// 5. Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n>>> [GoViet] Shutting down...")
}
