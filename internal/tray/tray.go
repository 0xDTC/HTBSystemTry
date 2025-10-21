package tray

import (
	"fmt"
	"htb-tool/internal/api"
	"htb-tool/internal/config"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/getlantern/systray"
)

type TrayApp struct {
	config  *config.Config
	client  *api.Client

	// Menu items
	userInfoItem    *systray.MenuItem
	apiKeyItem      *systray.MenuItem
	vpnMenu         *systray.MenuItem
	machinesMenu    *systray.MenuItem
	challengesMenu  *systray.MenuItem
	refreshItem     *systray.MenuItem
	quitItem        *systray.MenuItem
}

func New() *TrayApp {
	return &TrayApp{}
}

func (t *TrayApp) Run() {
	systray.Run(t.onReady, t.onExit)
}

func (t *TrayApp) onReady() {
	// Set icon and tooltip
	systray.SetIcon(getIcon())
	systray.SetTitle("HTB")
	systray.SetTooltip("HackTheBox Tool")

	// Load config
	var err error
	t.config, err = config.Load()
	if err != nil {
		log.Println("Failed to load config:", err)
	}

	// Initialize client if token exists
	if t.config.APIToken != "" {
		t.client = api.NewClient(t.config.APIToken)
	}

	// Build menu structure
	t.buildMenu()

	// Handle menu actions
	go t.handleMenuActions()
}

func (t *TrayApp) buildMenu() {
	// User Info Section
	t.userInfoItem = systray.AddMenuItem("👤 User Info", "View user information")
	systray.AddSeparator()

	// API Key
	t.apiKeyItem = systray.AddMenuItem("🔑 API Key", "Set your HTB API token")
	systray.AddSeparator()

	// VPN Menu
	t.vpnMenu = systray.AddMenuItem("🔐 VPN", "VPN connections and downloads")
	systray.AddSeparator()

	// Machines Menu
	t.machinesMenu = systray.AddMenuItem("🖥️  Machines", "Browse HTB machines")
	systray.AddSeparator()

	// Challenges Menu
	t.challengesMenu = systray.AddMenuItem("🎯 Challenges", "Browse HTB challenges")
	systray.AddSeparator()

	// Refresh
	t.refreshItem = systray.AddMenuItem("🔄 Refresh All", "Refresh all data")
	systray.AddSeparator()

	// Quit
	t.quitItem = systray.AddMenuItem("❌ Quit", "Exit HTB Tool")

	// Build submenus
	if t.client != nil {
		// Enable all menus
		t.userInfoItem.Enable()
		t.vpnMenu.Enable()
		t.machinesMenu.Enable()
		t.challengesMenu.Enable()

		// Load data
		go t.buildUserInfoMenu()
		go t.buildVPNMenu()
		go t.buildMachinesMenu()
		go t.buildChallengesMenu()
	} else {
		t.userInfoItem.SetTitle("👤 User Info (No API Key)")
		t.userInfoItem.Disable()
		t.vpnMenu.SetTitle("🔐 VPN (No API Key)")
		t.vpnMenu.Disable()
		t.machinesMenu.SetTitle("🖥️  Machines (No API Key)")
		t.machinesMenu.Disable()
		t.challengesMenu.SetTitle("🎯 Challenges (No API Key)")
		t.challengesMenu.Disable()
	}
}

func (t *TrayApp) buildUserInfoMenu() {
	// This will show a notification with user info
	// In a real implementation, fetch user profile from API
	t.userInfoItem.SetTitle("👤 User Info - Click to view")
}

func (t *TrayApp) buildVPNMenu() {
	servers, err := t.client.ListVPNServers()
	if err != nil {
		log.Println("Failed to load VPN servers:", err)
		return
	}

	// Group by location
	for _, server := range servers {
		load := "🟢"
		if server.CurrentClients > 50 {
			load = "🟡"
		}
		if server.CurrentClients > 100 {
			load = "🔴"
		}

		itemTitle := fmt.Sprintf("%s %s (%d clients)", load, server.FriendlyName, server.CurrentClients)
		serverItem := t.vpnMenu.AddSubMenuItem(itemTitle, fmt.Sprintf("Download VPN for %s", server.FriendlyName))

		// Add protocol submenu
		tcpItem := serverItem.AddSubMenuItem("📥 Download TCP", fmt.Sprintf("Download TCP config for %s", server.FriendlyName))
		udpItem := serverItem.AddSubMenuItem("📥 Download UDP", fmt.Sprintf("Download UDP config for %s", server.FriendlyName))

		// Handle downloads
		go func(id int, name string, tcp, udp *systray.MenuItem) {
			for {
				select {
				case <-tcp.ClickedCh:
					t.downloadVPN(id, "tcp", name)
				case <-udp.ClickedCh:
					t.downloadVPN(id, "udp", name)
				}
			}
		}(server.ID, server.FriendlyName, tcpItem, udpItem)
	}

	// Add separator and download all option
	t.vpnMenu.AddSubMenuItem("", "").Disable()
	downloadAllTCP := t.vpnMenu.AddSubMenuItem("📥 Download All (TCP)", "Download all VPN configs as TCP")
	downloadAllUDP := t.vpnMenu.AddSubMenuItem("📥 Download All (UDP)", "Download all VPN configs as UDP")

	go func() {
		for {
			select {
			case <-downloadAllTCP.ClickedCh:
				t.downloadAllVPNs("tcp", servers)
			case <-downloadAllUDP.ClickedCh:
				t.downloadAllVPNs("udp", servers)
			}
		}
	}()
}

func (t *TrayApp) buildMachinesMenu() {
	// Active Machines submenu
	activeMenu := t.machinesMenu.AddSubMenuItem("🟢 Active Machines", "Browse active machines")

	// Retired Machines submenu
	retiredMenu := t.machinesMenu.AddSubMenuItem("🔴 Retired Machines", "Browse retired machines")

	// Load machines
	go func() {
		machines, err := t.client.ListMachines()
		if err != nil {
			log.Println("Failed to load machines:", err)
			return
		}

		// Sort by release date (newest first)
		sort.Slice(machines, func(i, j int) bool {
			return machines[i].Release > machines[j].Release
		})

		// Split into active and retired
		var activeMachines, retiredMachines []api.Machine
		for _, m := range machines {
			if m.Active {
				activeMachines = append(activeMachines, m)
			} else if m.Retired {
				retiredMachines = append(retiredMachines, m)
			}
		}

		// Build active machines menu
		t.buildMachinesList(activeMenu, activeMachines, "Active")

		// Build retired machines menu
		t.buildMachinesList(retiredMenu, retiredMachines, "Retired")
	}()
}

func (t *TrayApp) buildMachinesList(parentMenu *systray.MenuItem, machines []api.Machine, status string) {
	// Add search info
	searchItem := parentMenu.AddSubMenuItem(fmt.Sprintf("📊 %d %s machines", len(machines), status), "Total count")
	searchItem.Disable()
	parentMenu.AddSubMenuItem("", "").Disable()

	// Add machines (limit to recent 20 for menu size)
	displayCount := 20
	if len(machines) < displayCount {
		displayCount = len(machines)
	}

	for i := 0; i < displayCount; i++ {
		m := machines[i]

		diffEmoji := getDifficultyEmoji(m.Difficulty)
		osEmoji := getOSEmoji(m.OS)

		title := fmt.Sprintf("%s %s %s", diffEmoji, osEmoji, m.Name)
		machineItem := parentMenu.AddSubMenuItem(title, fmt.Sprintf("%s - %s", m.Name, m.Difficulty))

		// Add machine actions
		infoItem := machineItem.AddSubMenuItem("ℹ️  Info", fmt.Sprintf("View info for %s", m.Name))

		var spawnItem, stopItem, resetItem *systray.MenuItem
		if m.PlayInfo != nil && m.PlayInfo.IsSpawned {
			stopItem = machineItem.AddSubMenuItem("⏹️  Stop", fmt.Sprintf("Stop %s", m.Name))
			resetItem = machineItem.AddSubMenuItem("🔄 Reset", fmt.Sprintf("Reset %s", m.Name))
		} else {
			spawnItem = machineItem.AddSubMenuItem("▶️  Spawn", fmt.Sprintf("Spawn %s", m.Name))
		}

		machineItem.AddSubMenuItem("", "").Disable()
		submitFlagItem := machineItem.AddSubMenuItem("🚩 Submit Flag", fmt.Sprintf("Submit flag for %s", m.Name))

		// Handle actions
		go t.handleMachineActions(m, infoItem, spawnItem, stopItem, resetItem, submitFlagItem)
	}

	if len(machines) > displayCount {
		moreItem := parentMenu.AddSubMenuItem(fmt.Sprintf("... and %d more", len(machines)-displayCount), "")
		moreItem.Disable()
	}
}

func (t *TrayApp) buildChallengesMenu() {
	// Active Challenges submenu
	activeMenu := t.challengesMenu.AddSubMenuItem("🟢 Active Challenges", "Browse active challenges")

	// Retired Challenges submenu
	retiredMenu := t.challengesMenu.AddSubMenuItem("🔴 Retired Challenges", "Browse retired challenges")

	// Load challenges
	go func() {
		challenges, err := t.client.ListChallenges()
		if err != nil {
			log.Println("Failed to load challenges:", err)
			return
		}

		// Split into active and retired
		var activeChallenges, retiredChallenges []api.Challenge
		for _, c := range challenges {
			if !c.Retired {
				activeChallenges = append(activeChallenges, c)
			} else {
				retiredChallenges = append(retiredChallenges, c)
			}
		}

		// Build active challenges menu by category
		t.buildChallengesListByCategory(activeMenu, activeChallenges, "Active")

		// Build retired challenges menu
		t.buildChallengesList(retiredMenu, retiredChallenges, "Retired")
	}()
}

func (t *TrayApp) buildChallengesListByCategory(parentMenu *systray.MenuItem, challenges []api.Challenge, status string) {
	// Group by category
	categoryMap := make(map[string][]api.Challenge)
	for _, c := range challenges {
		categoryMap[c.Category] = append(categoryMap[c.Category], c)
	}

	// Sort categories
	var categories []string
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	// Add count
	countItem := parentMenu.AddSubMenuItem(fmt.Sprintf("📊 %d %s challenges", len(challenges), status), "")
	countItem.Disable()
	parentMenu.AddSubMenuItem("", "").Disable()

	// Build menu for each category
	for _, category := range categories {
		chals := categoryMap[category]
		catEmoji := getCategoryEmoji(category)

		categoryItem := parentMenu.AddSubMenuItem(fmt.Sprintf("%s %s (%d)", catEmoji, category, len(chals)), "")

		// Add challenges in this category (limit to 10)
		displayCount := 10
		if len(chals) < displayCount {
			displayCount = len(chals)
		}

		for i := 0; i < displayCount; i++ {
			c := chals[i]
			diffEmoji := getDifficultyEmoji(c.Difficulty)
			solvedEmoji := "⚪"
			if c.IsSolved {
				solvedEmoji = "✅"
			}

			title := fmt.Sprintf("%s %s %s (%dpts)", solvedEmoji, diffEmoji, c.Name, c.Points)
			chalItem := categoryItem.AddSubMenuItem(title, fmt.Sprintf("%s - %s", c.Name, c.Difficulty))

			// Add actions
			infoItem := chalItem.AddSubMenuItem("ℹ️  Info", fmt.Sprintf("View info for %s", c.Name))
			submitItem := chalItem.AddSubMenuItem("🚩 Submit Flag", fmt.Sprintf("Submit flag for %s", c.Name))

			go t.handleChallengeActions(c, infoItem, submitItem)
		}

		if len(chals) > displayCount {
			moreItem := categoryItem.AddSubMenuItem(fmt.Sprintf("... and %d more", len(chals)-displayCount), "")
			moreItem.Disable()
		}
	}
}

func (t *TrayApp) buildChallengesList(parentMenu *systray.MenuItem, challenges []api.Challenge, status string) {
	countItem := parentMenu.AddSubMenuItem(fmt.Sprintf("📊 %d %s challenges", len(challenges), status), "")
	countItem.Disable()
	parentMenu.AddSubMenuItem("", "").Disable()

	// Show recent 20
	displayCount := 20
	if len(challenges) < displayCount {
		displayCount = len(challenges)
	}

	for i := 0; i < displayCount; i++ {
		c := challenges[i]
		catEmoji := getCategoryEmoji(c.Category)
		diffEmoji := getDifficultyEmoji(c.Difficulty)
		solvedEmoji := "⚪"
		if c.IsSolved {
			solvedEmoji = "✅"
		}

		title := fmt.Sprintf("%s %s %s [%s] (%dpts)", solvedEmoji, diffEmoji, catEmoji, c.Category, c.Points)
		chalItem := parentMenu.AddSubMenuItem(title, c.Name)

		infoItem := chalItem.AddSubMenuItem("ℹ️  Info", "View info")
		submitItem := chalItem.AddSubMenuItem("🚩 Submit Flag", "Submit flag")

		go t.handleChallengeActions(c, infoItem, submitItem)
	}

	if len(challenges) > displayCount {
		moreItem := parentMenu.AddSubMenuItem(fmt.Sprintf("... and %d more", len(challenges)-displayCount), "")
		moreItem.Disable()
	}
}

func (t *TrayApp) handleMenuActions() {
	for {
		select {
		case <-t.userInfoItem.ClickedCh:
			t.showUserInfo()

		case <-t.apiKeyItem.ClickedCh:
			t.promptAPIKey()

		case <-t.refreshItem.ClickedCh:
			t.refreshAllData()

		case <-t.quitItem.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (t *TrayApp) showUserInfo() {
	if t.client == nil {
		notify("User Info", "Please set your API key first!")
		return
	}

	// For now, show basic info
	notify("User Info", "API Key: Configured ✓\nClick 'Refresh All' to load data")
}

func (t *TrayApp) promptAPIKey() {
	apiKey, err := promptInput("HTB API Key", "Enter your HTB API token:\n(Get it from https://hackthebox.com/home/settings)")
	if err != nil || apiKey == "" {
		return
	}

	// Remove any whitespace/newlines
	apiKey = strings.TrimSpace(apiKey)

	// Update config
	t.config.APIToken = apiKey
	if err := t.config.Save(); err != nil {
		notify("Error", fmt.Sprintf("Failed to save API key:\n%v", err))
		return
	}

	// Update client
	t.client = api.NewClient(apiKey)

	notify("Success", "API key saved! Loading data...")

	// Automatically refresh all data
	t.refreshAllData()
}

func (t *TrayApp) refreshAllData() {
	if t.client == nil {
		notify("Error", "Please set your API key first!")
		return
	}

	notify("Refreshing", "Restarting HTB Tool to load fresh data...")

        // Create restart marker
        os.WriteFile("/tmp/htb-tool-restart", []byte("1"), 0644)
	}()
}

func (t *TrayApp) handleMachineActions(m api.Machine, info, spawn, stop, reset, submitFlag *systray.MenuItem) {
	for {
		select {
		case <-info.ClickedCh:
			t.showMachineInfo(m)

		case <-spawn.ClickedCh:
			t.spawnMachine(m)

		case <-stop.ClickedCh:
			t.stopMachine(m)

		case <-reset.ClickedCh:
			t.resetMachine(m)

		case <-submitFlag.ClickedCh:
			t.promptSubmitFlag(m)
		}
	}
}

func (t *TrayApp) handleChallengeActions(c api.Challenge, info, submit *systray.MenuItem) {
	for {
		select {
		case <-info.ClickedCh:
			t.showChallengeInfo(c)

		case <-submit.ClickedCh:
			t.promptSubmitChallengeFlag(c)
		}
	}
}

func (t *TrayApp) onExit() {
	// Cleanup
}

// Helper functions

func getDifficultyEmoji(difficulty string) string {
	switch strings.ToLower(difficulty) {
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
	osLower := strings.ToLower(os)
	if strings.Contains(osLower, "linux") {
		return "🐧"
	} else if strings.Contains(osLower, "windows") {
		return "🪟"
	} else if strings.Contains(osLower, "freebsd") {
		return "😈"
	}
	return "💻"
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
	case "misc", "miscellaneous":
		return "🎲"
	default:
		return "📦"
	}
}

func getIcon() []byte {
	// Try to load HTB logo from assets
	iconPath := "/media/kali/3315E7784CD31C71/Scripts/htb-tool/assets/htb-logo.png"
	data, err := os.ReadFile(iconPath)
	if err == nil && len(data) > 0 {
		return data
	}

	// Fallback: minimal PNG icon (16x16 green square with H)
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x91, 0x68, 0x36, 0x00, 0x00, 0x00,
	}
}

func showNotification(title, message string) {
	// Use notify-send on Linux
	log.Printf("[%s] %s\n", title, message)
}
