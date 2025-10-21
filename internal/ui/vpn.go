package ui

import (
	"fmt"
	"htb-tool/internal/api"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (a *App) createVPNTab() fyne.CanvasObject {
	// VPN Server list
	var vpnServers []api.VPNServer
	var selectedServers []int

	serverList := widget.NewList(
		func() int {
			return len(vpnServers)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewCheck("", func(bool) {}),
				widget.NewLabel("Template"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(vpnServers) {
				return
			}

			server := vpnServers[id]
			box := item.(*fyne.Container)

			// Update checkbox
			check := box.Objects[0].(*widget.Check)
			check.OnChanged = func(checked bool) {
				if checked {
					// Add to selection
					found := false
					for _, s := range selectedServers {
						if s == server.ID {
							found = true
							break
						}
					}
					if !found {
						selectedServers = append(selectedServers, server.ID)
					}
				} else {
					// Remove from selection
					for i, s := range selectedServers {
						if s == server.ID {
							selectedServers = append(selectedServers[:i], selectedServers[i+1:]...)
							break
						}
					}
				}
			}

			// Update label
			label := box.Objects[1].(*widget.Label)
			label.SetText(fmt.Sprintf("%s (%s) - %d clients",
				server.FriendlyName, server.Location, server.CurrentClients))
		},
	)

	// Protocol selection
	protocolSelect := widget.NewRadioGroup([]string{"TCP", "UDP"}, func(value string) {})
	protocolSelect.SetSelected("TCP")
	protocolSelect.Horizontal = true

	// Download directory
	downloadDirEntry := widget.NewEntry()
	downloadDirEntry.SetText(a.config.VPNDirectory)

	browseDirBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			downloadDirEntry.SetText(uri.Path())
			a.config.VPNDirectory = uri.Path()
			a.config.Save()
		}, a.mainWindow)
	})

	downloadDirBox := container.NewBorder(nil, nil, nil, browseDirBtn, downloadDirEntry)

	// Select all / Deselect all buttons
	selectAllBtn := widget.NewButton("Select All", func() {
		selectedServers = []int{}
		for _, server := range vpnServers {
			selectedServers = append(selectedServers, server.ID)
		}
		serverList.Refresh()
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		selectedServers = []int{}
		serverList.Refresh()
	})

	// Download button
	downloadBtn := widget.NewButton("Download Selected VPNs", func() {
		if len(selectedServers) == 0 {
			dialog.ShowError(fmt.Errorf("no servers selected"), a.mainWindow)
			return
		}

		protocol := strings.ToLower(protocolSelect.Selected)
		a.downloadVPNConfigs(selectedServers, protocol, downloadDirEntry.Text)
	})

	// Refresh button
	refreshBtn := widget.NewButton("Refresh Servers", func() {
		a.refreshVPNServers(&vpnServers, serverList)
	})

	// Layout
	controls := container.NewVBox(
		widget.NewLabel("VPN Download Directory:"),
		downloadDirBox,
		widget.NewSeparator(),
		widget.NewLabel("Protocol:"),
		protocolSelect,
		widget.NewSeparator(),
		container.NewHBox(selectAllBtn, deselectAllBtn, refreshBtn),
		downloadBtn,
	)

	content := container.NewBorder(controls, nil, nil, nil, serverList)

	// Auto-load VPN servers
	go a.refreshVPNServers(&vpnServers, serverList)

	return content
}

func (a *App) refreshVPNServers(vpnServers *[]api.VPNServer, serverList *widget.List) {
	a.setStatus("Loading VPN servers...")

	servers, err := a.client.ListVPNServers()
	if err != nil {
		a.setStatus("Error: " + err.Error())
		dialog.ShowError(err, a.mainWindow)
		return
	}

	*vpnServers = servers
	serverList.Refresh()
	a.setStatus(fmt.Sprintf("Loaded %d VPN servers", len(servers)))
}

func (a *App) downloadVPNConfigs(serverIDs []int, protocol, directory string) {
	// Ensure directory exists
	if err := os.MkdirAll(directory, 0755); err != nil {
		dialog.ShowError(fmt.Errorf("failed to create directory: %w", err), a.mainWindow)
		return
	}

	successCount := 0
	failCount := 0

	progress := dialog.NewProgressInfinite("Downloading VPN configs...",
		fmt.Sprintf("Downloading %d VPN configuration files", len(serverIDs)),
		a.mainWindow)
	progress.Show()

	go func() {
		for i, serverID := range serverIDs {
			a.setStatus(fmt.Sprintf("Downloading VPN config %d/%d...", i+1, len(serverIDs)))

			data, err := a.client.DownloadVPN(serverID, protocol)
			if err != nil {
				failCount++
				continue
			}

			// Save to file
			filename := fmt.Sprintf("htb_vpn_%d_%s.ovpn", serverID, protocol)
			filepath := filepath.Join(directory, filename)

			if err := os.WriteFile(filepath, data, 0600); err != nil {
				failCount++
				continue
			}

			successCount++
		}

		progress.Hide()

		if failCount > 0 {
			dialog.ShowInformation("Download Complete",
				fmt.Sprintf("Downloaded %d/%d VPN configs\n%d failed",
					successCount, len(serverIDs), failCount),
				a.mainWindow)
		} else {
			dialog.ShowInformation("Success!",
				fmt.Sprintf("Successfully downloaded all %d VPN configs to:\n%s",
					successCount, directory),
				a.mainWindow)
		}

		a.setStatus(fmt.Sprintf("Downloaded %d VPN configs", successCount))
	}()
}
