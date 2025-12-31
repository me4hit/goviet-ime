package main

import (
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
)

// InputEngine is the object that will receive messages from Fcitx5
type InputEngine struct{}

// ProcessKey is the function that will be called from outside
// Input: keysym (keycode), modifiers (Shift/Ctrl keys)
// Output: handled (is it processed), commitText, preeditText
func (e *InputEngine) ProcessKey(keysym uint32, modifiers uint32) (bool, string, string, *dbus.Error) {
	// Print to the screen to let you know it has BEEN RECEIVED
	fmt.Printf(">>> [Go App] Received Key: %d (Mods: %d)\n", keysym, modifiers)

	// Simulate logic: If 'a' key (code 97) is received
	if keysym == 97 {
		fmt.Println("    -> Detected 'a', processing...")
		return true, "", "a", nil // Return: Handled=true, Preedit="a"
	}

	// Ignore other keys
	return false, "", "", nil
}

func main() {
	// 1. Connect to Session Bus (User Bus)
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// 2. Register Service Name (must be unique)
	// This name must match what you call from C++ later
	serviceName := "com.github.goviet.ime"
	reply, err := conn.RequestName(serviceName, dbus.NameFlagDoNotQueue)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to request name:", err)
		os.Exit(1)
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "Name already taken")
		os.Exit(1)
	}

	// 3. Export object so others can call it
	engine := &InputEngine{}
	// Export engine at path "/Engine", with interface name as serviceName
	err = conn.Export(engine, "/Engine", serviceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to export object:", err)
		os.Exit(1)
	}

	fmt.Println("------------------------------------------------")
	fmt.Printf("✅ GoViet-IME Backend is running!\n")
	fmt.Printf("Listening on Bus: %s\n", serviceName)
	fmt.Printf("Object Path:      /Engine\n")
	fmt.Println("Waiting for keys...")
	fmt.Println("------------------------------------------------")

	// 4. Keep the program running forever
	select {}
}
