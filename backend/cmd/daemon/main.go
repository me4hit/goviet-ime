package main

import (
	"fmt"
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
}

// NewInputEngine creates a new InputEngine with default settings.
func NewInputEngine() *InputEngine {
	return &InputEngine{
		engine: engine.NewCompositionEngine(),
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

	// fmt.Printf(">>> [GoViet] Key: %d (Mods: %d) -> Handled: %v, Commit: %q, Preedit: %q\n",
	// 	keysym, modifiers, result.Handled, result.CommitText, result.Preedit)

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

	// 3. Create and export the engine
	inputEngine := NewInputEngine()

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
