package tray

import (
	"fmt"
	"htb-tool/internal/api"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func (t *TrayApp) downloadVPN(serverID int, protocol, name string) {
	notify("Downloading", fmt.Sprintf("Downloading %s VPN config (%s)...", name, protocol))

	data, err := t.client.DownloadVPN(serverID, protocol)
	if err != nil {
		notify("Error", fmt.Sprintf("Failed to download VPN: %v", err))
		return
	}

	// Ensure directory exists
	vpnDir := t.config.VPNDirectory
	if err := os.MkdirAll(vpnDir, 0755); err != nil {
		notify("Error", fmt.Sprintf("Failed to create directory: %v", err))
		return
	}

	// Save file
	filename := fmt.Sprintf("htb_vpn_%d_%s.ovpn", serverID, protocol)
	filepath := filepath.Join(vpnDir, filename)

	if err := os.WriteFile(filepath, data, 0600); err != nil {
		notify("Error", fmt.Sprintf("Failed to save file: %v", err))
		return
	}

	notify("Success", fmt.Sprintf("Downloaded %s to:\n%s", name, filepath))
}

func (t *TrayApp) downloadAllVPNs(protocol string, servers []api.VPNServer) {
	notify("Downloading", fmt.Sprintf("Downloading all %d VPN configs (%s)...", len(servers), protocol))

	successCount := 0
	for _, server := range servers {
		data, err := t.client.DownloadVPN(server.ID, protocol)
		if err != nil {
			log.Printf("Failed to download VPN %s: %v\n", server.FriendlyName, err)
			continue
		}

		vpnDir := t.config.VPNDirectory
		os.MkdirAll(vpnDir, 0755)

		filename := fmt.Sprintf("htb_vpn_%d_%s.ovpn", server.ID, protocol)
		filepath := filepath.Join(vpnDir, filename)

		if err := os.WriteFile(filepath, data, 0600); err != nil {
			log.Printf("Failed to save VPN %s: %v\n", server.FriendlyName, err)
			continue
		}

		successCount++
	}

	notify("Success", fmt.Sprintf("Downloaded %d/%d VPN configs to:\n%s", successCount, len(servers), t.config.VPNDirectory))
}

func (t *TrayApp) spawnMachine(m api.Machine) {
	notify("Spawning", fmt.Sprintf("Spawning machine: %s", m.Name))

	err := t.client.SpawnMachine(m.ID)
	if err != nil {
		notify("Error", fmt.Sprintf("Failed to spawn %s:\n%v", m.Name, err))
		return
	}

	notify("Success", fmt.Sprintf("%s is spawning!\nIt will be ready in ~2-3 minutes", m.Name))
}

func (t *TrayApp) stopMachine(m api.Machine) {
	notify("Stopping", fmt.Sprintf("Stopping machine: %s", m.Name))

	err := t.client.TerminateMachine(m.ID)
	if err != nil {
		notify("Error", fmt.Sprintf("Failed to stop %s:\n%v", m.Name, err))
		return
	}

	notify("Success", fmt.Sprintf("%s has been stopped", m.Name))
}

func (t *TrayApp) resetMachine(m api.Machine) {
	notify("Resetting", fmt.Sprintf("Resetting machine: %s", m.Name))

	err := t.client.ResetMachine(m.ID)
	if err != nil {
		notify("Error", fmt.Sprintf("Failed to reset %s:\n%v", m.Name, err))
		return
	}

	notify("Success", fmt.Sprintf("%s has been reset", m.Name))
}

func (t *TrayApp) showMachineInfo(m api.Machine) {
	status := "Retired"
	if m.Active {
		status = "Active"
	}

	spawned := "Not spawned"
	ip := "N/A"
	if m.PlayInfo != nil && m.PlayInfo.IsSpawned {
		spawned = "Spawned ✓"
		if m.IP != "" {
			ip = m.IP
		}
	}

	info := fmt.Sprintf(`%s

OS: %s
Difficulty: %s
Rating: %.1f/5.0
Status: %s

User Owns: %d
Root Owns: %d

IP: %s
%s

Release: %s`,
		m.Name,
		m.OS,
		m.Difficulty,
		m.Star,
		status,
		m.UserOwns,
		m.RootOwns,
		ip,
		spawned,
		m.Release,
	)

	notify("Machine Info", info)
}

func (t *TrayApp) showChallengeInfo(c api.Challenge) {
	solved := "Not solved"
	if c.IsSolved {
		solved = "Solved ✓"
	}

	info := fmt.Sprintf(`%s

Category: %s
Difficulty: %s
Points: %d

Solves: %d
Likes: %d
%s`,
		c.Name,
		c.Category,
		c.Difficulty,
		c.Points,
		c.Solves,
		c.Likes,
		solved,
	)

	notify("Challenge Info", info)
}

func (t *TrayApp) promptSubmitFlag(m api.Machine) {
	// Use zenity to prompt for flag
	flag, err := promptInput("Submit Flag", fmt.Sprintf("Enter flag for %s:", m.Name))
	if err != nil || flag == "" {
		return
	}

	// Ask for flag type
	flagType, err := promptChoice("Flag Type", "Select flag type:", []string{"User Flag", "Root Flag"})
	if err != nil {
		return
	}

	difficulty := 0
	if flagType == "Root Flag" {
		difficulty = 1
	}

	// Submit flag
	notify("Submitting", fmt.Sprintf("Submitting %s for %s...", flagType, m.Name))

	resp, err := t.client.SubmitMachineFlag(m.ID, flag, difficulty)
	if err != nil {
		notify("Error", fmt.Sprintf("Failed to submit flag:\n%v", err))
		return
	}

	if resp.Success {
		notify("Success!", fmt.Sprintf("🎉 Correct %s!\n\n%s", flagType, resp.Message))
	} else {
		notify("Incorrect", fmt.Sprintf("❌ Wrong flag\n\n%s", resp.Message))
	}
}

func (t *TrayApp) promptSubmitChallengeFlag(c api.Challenge) {
	// Use zenity to prompt for flag
	flag, err := promptInput("Submit Flag", fmt.Sprintf("Enter flag for %s:", c.Name))
	if err != nil || flag == "" {
		return
	}

	// Submit flag
	notify("Submitting", fmt.Sprintf("Submitting flag for %s...", c.Name))

	resp, err := t.client.SubmitChallengeFlag(c.ID, flag)
	if err != nil {
		notify("Error", fmt.Sprintf("Failed to submit flag:\n%v", err))
		return
	}

	if resp.Success {
		notify("Success!", fmt.Sprintf("🎉 Correct flag!\n\n%s", resp.Message))
	} else {
		notify("Incorrect", fmt.Sprintf("❌ Wrong flag\n\n%s", resp.Message))
	}
}

// Helper functions

func notify(title, message string) {
	log.Printf("[%s] %s\n", title, message)

	// Use notify-send on Linux
	cmd := exec.Command("notify-send", "-a", "HTB Tool", title, message)
	cmd.Run()
}

func promptInput(title, message string) (string, error) {
	cmd := exec.Command("zenity", "--entry", "--title="+title, "--text="+message)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func promptChoice(title, message string, options []string) (string, error) {
	args := []string{"--list", "--title=" + title, "--text=" + message, "--column=Option"}
	args = append(args, options...)

	cmd := exec.Command("zenity", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
