package ui

import (
	"fmt"
	"htb-tool/internal/api"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (a *App) createSettingsTab() fyne.CanvasObject {
	// API Token
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetText(a.config.APIToken)

	saveTokenBtn := widget.NewButton("Update API Token", func() {
		newToken := tokenEntry.Text
		if newToken == "" {
			dialog.ShowError(fmt.Errorf("API token cannot be empty"), a.mainWindow)
			return
		}

		a.config.APIToken = newToken
		if err := a.config.Save(); err != nil {
			dialog.ShowError(err, a.mainWindow)
			return
		}

		// Update client
		a.client = api.NewClient(newToken)
		dialog.ShowInformation("Success", "API token updated successfully!", a.mainWindow)
	})

	tokenBox := container.NewVBox(
		widget.NewLabel("HTB API Token:"),
		widget.NewLabel("Get your token from: https://www.hackthebox.com/home/settings"),
		tokenEntry,
		saveTokenBtn,
		widget.NewSeparator(),
	)

	// VPN Directory
	vpnDirEntry := widget.NewEntry()
	vpnDirEntry.SetText(a.config.VPNDirectory)

	browseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			vpnDirEntry.SetText(uri.Path())
		}, a.mainWindow)
	})

	saveVPNDirBtn := widget.NewButton("Save", func() {
		a.config.VPNDirectory = vpnDirEntry.Text
		if err := a.config.Save(); err != nil {
			dialog.ShowError(err, a.mainWindow)
			return
		}
		dialog.ShowInformation("Success", "VPN directory updated!", a.mainWindow)
	})

	vpnDirBox := container.NewVBox(
		widget.NewLabel("Default VPN Download Directory:"),
		container.NewBorder(nil, nil, nil, browseBtn, vpnDirEntry),
		saveVPNDirBtn,
		widget.NewSeparator(),
	)

	// About
	aboutBox := container.NewVBox(
		widget.NewLabel("HTB Tool v1.0"),
		widget.NewLabel("A HackTheBox management tool"),
		widget.NewLabel(""),
		widget.NewLabel("Features:"),
		widget.NewLabel("• Browse and search machines"),
		widget.NewLabel("• Spawn/terminate machines"),
		widget.NewLabel("• Submit flags directly"),
		widget.NewLabel("• Download VPN configs"),
		widget.NewLabel("• System tray integration"),
	)

	content := container.NewVBox(
		tokenBox,
		vpnDirBox,
		aboutBox,
	)

	return container.NewScroll(content)
}
