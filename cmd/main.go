package main

import (
	"htb-tool/internal/tray"
	"os"
	"os/exec"
	"time"
)

func main() {
	for {
		app := tray.New()
		app.Run()

		// Check if we should restart
		// Small delay to let app fully exit
		time.Sleep(500 * time.Millisecond)

		if shouldRestart() {
			// Restart the application
			restartSelf()
			break
		}

		// Normal exit
		break
	}
}

func shouldRestart() bool {
	// Check if restart marker file exists
	markerFile := "/tmp/htb-tool-restart"
	if _, err := os.Stat(markerFile); err == nil {
		os.Remove(markerFile)
		return true
	}
	return false
}

func restartSelf() {
	// Get the current executable path
	executable, err := os.Executable()
	if err != nil {
		return
	}

	// Start new instance in background
	cmd := exec.Command(executable)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}
