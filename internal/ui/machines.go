package ui

import (
	"fmt"
	"htb-tool/internal/api"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (a *App) createMachinesTab() fyne.CanvasObject {
	// Search and filter controls
	a.searchEntry = widget.NewEntry()
	a.searchEntry.SetPlaceHolder("Search machines...")
	a.searchEntry.OnChanged = func(query string) {
		a.filterMachines()
	}

	a.filterSelect = widget.NewSelect([]string{
		"All",
		"Active",
		"Retired",
		"Easy",
		"Medium",
		"Hard",
		"Insane",
		"Linux",
		"Windows",
		"Free Only",
	}, func(value string) {
		a.filterMachines()
	})
	a.filterSelect.SetSelected("All")

	refreshBtn := widget.NewButton("Refresh", func() {
		a.refreshMachines()
	})

	controls := container.NewBorder(
		nil, nil,
		container.NewHBox(widget.NewLabel("Filter:"), a.filterSelect),
		refreshBtn,
		a.searchEntry,
	)

	// Machine list
	a.machineList = widget.NewList(
		func() int {
			return len(a.filteredMachines)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Template"),
				layout.NewSpacer(),
				widget.NewButton("Info", func() {}),
				widget.NewButton("Spawn", func() {}),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(a.filteredMachines) {
				return
			}

			machine := a.filteredMachines[id]
			box := item.(*fyne.Container)

			// Update label
			label := box.Objects[0].(*widget.Label)
			status := "⚪"
			if machine.Active {
				status = "🟢"
			} else if machine.Retired {
				status = "🔴"
			}

			diffEmoji := a.getDifficultyEmoji(machine.Difficulty)
			osEmoji := a.getOSEmoji(machine.OS)

			label.SetText(fmt.Sprintf("%s %s %s %s - %s",
				status, diffEmoji, osEmoji, machine.Name, machine.Difficulty))

			// Update info button
			infoBtn := box.Objects[2].(*widget.Button)
			infoBtn.OnTapped = func() {
				a.showMachineInfo(machine)
			}

			// Update spawn/stop button
			actionBtn := box.Objects[3].(*widget.Button)
			if machine.PlayInfo != nil && machine.PlayInfo.IsSpawned {
				actionBtn.SetText("Stop")
				actionBtn.OnTapped = func() {
					a.terminateMachine(machine.ID, machine.Name)
				}
			} else {
				actionBtn.SetText("Spawn")
				actionBtn.OnTapped = func() {
					a.spawnMachine(machine.ID, machine.Name)
				}
			}
		},
	)

	content := container.NewBorder(controls, nil, nil, nil, a.machineList)

	// Auto-load machines
	go a.refreshMachines()

	return content
}

func (a *App) refreshMachines() {
	a.setStatus("Loading machines...")

	machines, err := a.client.ListMachines()
	if err != nil {
		a.setStatus("Error: " + err.Error())
		dialog.ShowError(err, a.mainWindow)
		return
	}

	a.allMachines = machines
	a.filterMachines()
	a.setStatus(fmt.Sprintf("Loaded %d machines", len(machines)))
}

func (a *App) filterMachines() {
	query := strings.ToLower(strings.TrimSpace(a.searchEntry.Text))
	filter := a.filterSelect.Selected

	a.filteredMachines = []api.Machine{}

	for _, machine := range a.allMachines {
		// Apply search filter
		if query != "" {
			if !strings.Contains(strings.ToLower(machine.Name), query) &&
				!strings.Contains(strings.ToLower(machine.OS), query) {
				continue
			}
		}

		// Apply category filter
		switch filter {
		case "Active":
			if !machine.Active {
				continue
			}
		case "Retired":
			if !machine.Retired {
				continue
			}
		case "Easy", "Medium", "Hard", "Insane":
			if !strings.EqualFold(machine.Difficulty, filter) {
				continue
			}
		case "Linux":
			if !strings.Contains(strings.ToLower(machine.OS), "linux") {
				continue
			}
		case "Windows":
			if !strings.Contains(strings.ToLower(machine.OS), "windows") {
				continue
			}
		case "Free Only":
			if !machine.Free {
				continue
			}
		}

		a.filteredMachines = append(a.filteredMachines, machine)
	}

	if a.machineList != nil {
		a.machineList.Refresh()
	}
}

func (a *App) showMachineInfo(machine api.Machine) {
	info := fmt.Sprintf(`Name: %s
OS: %s
Difficulty: %s
Rating: %.1f/5.0
User Owns: %d
Root Owns: %d
IP: %s
Release: %s
Status: %s
Free: %v`,
		machine.Name,
		machine.OS,
		machine.Difficulty,
		machine.Star,
		machine.UserOwns,
		machine.RootOwns,
		machine.IP,
		machine.Release,
		func() string {
			if machine.Active {
				return "Active"
			}
			return "Retired"
		}(),
		machine.Free,
	)

	// Create dialog with flag submission
	content := container.NewVBox(
		widget.NewLabel(info),
		widget.NewSeparator(),
	)

	// Add flag submission if machine is spawned
	if machine.PlayInfo != nil && machine.PlayInfo.IsSpawned {
		content.Add(widget.NewLabel("Submit Flag:"))

		flagEntry := widget.NewEntry()
		flagEntry.SetPlaceHolder("Enter flag...")

		userFlagBtn := widget.NewButton("Submit User Flag", func() {
			a.submitFlag(machine.ID, machine.Name, flagEntry.Text, 0)
		})

		rootFlagBtn := widget.NewButton("Submit Root Flag", func() {
			a.submitFlag(machine.ID, machine.Name, flagEntry.Text, 1)
		})

		content.Add(flagEntry)
		content.Add(container.NewHBox(userFlagBtn, rootFlagBtn))
	}

	dialog.ShowCustom("Machine Info: "+machine.Name, "Close", content, a.mainWindow)
}

func (a *App) spawnMachine(machineID int, name string) {
	a.setStatus(fmt.Sprintf("Spawning %s...", name))

	go func() {
		err := a.client.SpawnMachine(machineID)
		if err != nil {
			a.setStatus("Error: " + err.Error())
			dialog.ShowError(fmt.Errorf("failed to spawn %s: %w", name, err), a.mainWindow)
			return
		}

		a.setStatus(fmt.Sprintf("Successfully spawned %s", name))
		dialog.ShowInformation("Success", fmt.Sprintf("Machine %s is spawning!", name), a.mainWindow)
		a.refreshMachines()
	}()
}

func (a *App) terminateMachine(machineID int, name string) {
	a.setStatus(fmt.Sprintf("Terminating %s...", name))

	go func() {
		err := a.client.TerminateMachine(machineID)
		if err != nil {
			a.setStatus("Error: " + err.Error())
			dialog.ShowError(fmt.Errorf("failed to terminate %s: %w", name, err), a.mainWindow)
			return
		}

		a.setStatus(fmt.Sprintf("Successfully terminated %s", name))
		dialog.ShowInformation("Success", fmt.Sprintf("Machine %s terminated!", name), a.mainWindow)
		a.refreshMachines()
	}()
}

func (a *App) submitFlag(machineID int, name, flag string, difficulty int) {
	if flag == "" {
		dialog.ShowError(fmt.Errorf("flag cannot be empty"), a.mainWindow)
		return
	}

	flagType := "User"
	if difficulty == 1 {
		flagType = "Root"
	}

	a.setStatus(fmt.Sprintf("Submitting %s flag for %s...", flagType, name))

	go func() {
		resp, err := a.client.SubmitMachineFlag(machineID, flag, difficulty)
		if err != nil {
			a.setStatus("Error: " + err.Error())
			dialog.ShowError(fmt.Errorf("failed to submit flag: %w", err), a.mainWindow)
			return
		}

		if resp.Success {
			a.setStatus(fmt.Sprintf("✓ Correct flag for %s!", name))
			dialog.ShowInformation("Success!", fmt.Sprintf("🎉 Correct %s flag!\n\n%s", flagType, resp.Message), a.mainWindow)
		} else {
			a.setStatus(fmt.Sprintf("✗ Incorrect flag for %s", name))
			dialog.ShowError(fmt.Errorf("incorrect flag: %s", resp.Message), a.mainWindow)
		}
	}()
}

func (a *App) getDifficultyEmoji(difficulty string) string {
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

func (a *App) getOSEmoji(os string) string {
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
