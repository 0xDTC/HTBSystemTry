package tray

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"htb-tool/internal/config"
)

// notify shows a normal desktop notification (and logs it).
func notify(title, message string) { notifyLevel(title, message, false) }

// notifyErr shows a critical desktop notification (and logs it).
func notifyErr(title, message string) { notifyLevel(title, message, true) }

func notifyLevel(title, message string, critical bool) {
	log.Printf("[%s] %s", title, message)
	urgency, icon := "normal", "dialog-information"
	if critical {
		urgency, icon = "critical", "dialog-error"
	}
	// notify-send is the only external tool we depend on for output.
	_ = exec.Command("notify-send", "-a", "HTB Tool", "-u", urgency, "-i", icon, title, message).Run()
}

// clipboardRead returns the current X clipboard contents (trimmed) via xclip.
func clipboardRead() (string, error) {
	out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// clipboardWrite places s on the X clipboard via xclip (best effort).
func clipboardWrite(s string) {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(s)
	_ = cmd.Run()
}

// readToken resolves the API token from the saved config, falling back to the
// HTB_TOKEN environment variable.
func readToken(cfg *config.Config) string {
	if cfg != nil {
		if tok := strings.TrimSpace(cfg.APIToken); tok != "" {
			return tok
		}
	}
	return strings.TrimSpace(os.Getenv("HTB_TOKEN"))
}

// capN returns at most the first n elements of s.
func capN[T any](s []T, n int) []T {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// truncate shortens s to at most n runes, adding an ellipsis when cut.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

func getDifficultyEmoji(difficulty string) string {
	switch strings.ToLower(strings.TrimSpace(difficulty)) {
	case "very easy":
		return "🟢"
	case "easy":
		return "🟢"
	case "medium":
		return "🟡"
	case "hard":
		return "🟠"
	case "insane":
		return "🔴"
	default:
		return "⚪"
	}
}

func getOSEmoji(os string) string {
	switch o := strings.ToLower(os); {
	case strings.Contains(o, "linux"):
		return "🐧"
	case strings.Contains(o, "windows"):
		return "🪟"
	case strings.Contains(o, "freebsd"):
		return "😈"
	case strings.Contains(o, "android"):
		return "🤖"
	default:
		return "💻"
	}
}

func getCategoryEmoji(category string) string {
	switch strings.ToLower(category) {
	case "web":
		return "🌐"
	case "crypto", "cryptography":
		return "🔐"
	case "pwn", "pwning":
		return "💥"
	case "reversing", "reverse engineering":
		return "🔄"
	case "forensics":
		return "🔍"
	case "osint":
		return "🕵️"
	case "mobile":
		return "📱"
	case "hardware":
		return "🔧"
	case "blockchain":
		return "⛓️"
	case "misc", "miscellaneous":
		return "🎲"
	default:
		return "📦"
	}
}
