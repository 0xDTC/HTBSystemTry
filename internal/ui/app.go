package ui

import (
	"fmt"
	"htb-tool/internal/api"
	"htb-tool/internal/config"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window
	config     *config.Config
	client     *api.Client

	// UI Components
	machineList   *widget.List
	challengeList *widget.List
	searchEntry   *widget.Entry
	filterSelect  *widget.Select
	statusLabel   *widget.Label

	// Data
	allMachines      []api.Machine
	filteredMachines []api.Machine
	allChallenges    []api.Challenge

	// Helper functions
	filterChallenges func(query, category string)
}

func NewApp() *App {
	a := &App{
		fyneApp: app.NewWithID("com.htb.tool"),
	}

	return a
}

func (a *App) Run() {
	// Load config
	var err error
	a.config, err = config.Load()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load config: %w", err), nil)
	}

	// Check if API token is set
	if a.config.APIToken == "" {
		a.showSetupDialog()
	} else {
		a.client = api.NewClient(a.config.APIToken)
		a.showMainWindow()
	}

	a.fyneApp.Run()
}

func (a *App) showSetupDialog() {
	a.mainWindow = a.fyneApp.NewWindow("HTB Tool - Setup")
	a.mainWindow.Resize(fyne.NewSize(500, 200))

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("Enter your HTB API Token")

	saveBtn := widget.NewButton("Save & Continue", func() {
		token := strings.TrimSpace(tokenEntry.Text)
		if token == "" {
			dialog.ShowError(fmt.Errorf("API token cannot be empty"), a.mainWindow)
			return
		}

		a.config.APIToken = token
		if err := a.config.Save(); err != nil {
			dialog.ShowError(err, a.mainWindow)
			return
		}

		a.client = api.NewClient(token)
		a.mainWindow.Close()
		a.showMainWindow()
	})

	helpLabel := widget.NewLabel("Get your API token from: https://www.hackthebox.com/home/settings")
	helpLabel.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		widget.NewLabel("Welcome to HTB Tool!"),
		widget.NewSeparator(),
		helpLabel,
		tokenEntry,
		saveBtn,
	)

	a.mainWindow.SetContent(content)
	a.mainWindow.Show()
}

func (a *App) showMainWindow() {
	a.mainWindow = a.fyneApp.NewWindow("HTB Tool")
	a.mainWindow.Resize(fyne.NewSize(float32(a.config.WindowWidth), float32(a.config.WindowHeight)))

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Machines", a.createMachinesTab()),
		container.NewTabItem("Challenges", a.createChallengesTab()),
		container.NewTabItem("VPN", a.createVPNTab()),
		container.NewTabItem("Settings", a.createSettingsTab()),
	)

	// Status bar
	a.statusLabel = widget.NewLabel("Ready")
	statusBar := container.NewBorder(nil, nil, nil, nil, a.statusLabel)

	// Main content
	content := container.NewBorder(nil, statusBar, nil, nil, tabs)

	a.mainWindow.SetContent(content)

	// Setup system tray
	if desk, ok := a.fyneApp.(desktop.App); ok {
		a.setupSystemTray(desk)
	}

	a.mainWindow.SetCloseIntercept(func() {
		a.mainWindow.Hide()
	})

	a.mainWindow.Show()
}

func (a *App) setupSystemTray(desk desktop.App) {
	menu := fyne.NewMenu("HTB Tool",
		fyne.NewMenuItem("Show", func() {
			a.mainWindow.Show()
		}),
		fyne.NewMenuItem("Refresh Machines", func() {
			a.refreshMachines()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			a.fyneApp.Quit()
		}),
	)

	desk.SetSystemTrayMenu(menu)
}

func (a *App) setStatus(message string) {
	a.statusLabel.SetText(message)
}
