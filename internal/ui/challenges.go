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

func (a *App) createChallengesTab() fyne.CanvasObject {
	// Search and filter controls
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search challenges...")

	categoryFilter := widget.NewSelect([]string{
		"All",
		"Web",
		"Crypto",
		"Pwn",
		"Reversing",
		"Forensics",
		"Misc",
		"OSINT",
		"Mobile",
		"Hardware",
	}, func(value string) {
		a.filterChallenges(searchEntry.Text, value)
	})
	categoryFilter.SetSelected("All")

	searchEntry.OnChanged = func(query string) {
		a.filterChallenges(query, categoryFilter.Selected)
	}

	refreshBtn := widget.NewButton("Refresh", func() {
		a.refreshChallenges()
	})

	controls := container.NewBorder(
		nil, nil,
		container.NewHBox(widget.NewLabel("Category:"), categoryFilter),
		refreshBtn,
		searchEntry,
	)

	// Challenge list
	var filteredChallenges []api.Challenge

	challengeList := widget.NewList(
		func() int {
			return len(filteredChallenges)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Template"),
				layout.NewSpacer(),
				widget.NewButton("Submit Flag", func() {}),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(filteredChallenges) {
				return
			}

			challenge := filteredChallenges[id]
			box := item.(*fyne.Container)

			// Update label
			label := box.Objects[0].(*widget.Label)
			solvedEmoji := "⚪"
			if challenge.IsSolved {
				solvedEmoji = "✅"
			}

			diffEmoji := a.getDifficultyEmoji(challenge.Difficulty)

			label.SetText(fmt.Sprintf("%s %s [%s] %s - %d pts (%d solves)",
				solvedEmoji, diffEmoji, challenge.Category, challenge.Name, challenge.Points, challenge.Solves))

			// Update submit button
			submitBtn := box.Objects[2].(*widget.Button)
			submitBtn.OnTapped = func() {
				a.showChallengeDialog(challenge)
			}
		},
	)

	a.filterChallenges = func(query, category string) {
		query = strings.ToLower(strings.TrimSpace(query))
		filteredChallenges = []api.Challenge{}

		for _, chal := range a.allChallenges {
			// Apply search filter
			if query != "" {
				if !strings.Contains(strings.ToLower(chal.Name), query) &&
					!strings.Contains(strings.ToLower(chal.Category), query) {
					continue
				}
			}

			// Apply category filter
			if category != "All" && !strings.EqualFold(chal.Category, category) {
				continue
			}

			filteredChallenges = append(filteredChallenges, chal)
		}

		challengeList.Refresh()
	}

	content := container.NewBorder(controls, nil, nil, nil, challengeList)

	// Auto-load challenges
	go a.refreshChallenges()

	return content
}

func (a *App) refreshChallenges() {
	a.setStatus("Loading challenges...")

	challenges, err := a.client.ListChallenges()
	if err != nil {
		a.setStatus("Error: " + err.Error())
		dialog.ShowError(err, a.mainWindow)
		return
	}

	a.allChallenges = challenges
	a.setStatus(fmt.Sprintf("Loaded %d challenges", len(challenges)))
}

func (a *App) showChallengeDialog(challenge api.Challenge) {
	info := fmt.Sprintf(`Name: %s
Category: %s
Difficulty: %s
Points: %d
Solves: %d
Likes: %d
Dislikes: %d
Status: %s`,
		challenge.Name,
		challenge.Category,
		challenge.Difficulty,
		challenge.Points,
		challenge.Solves,
		challenge.Likes,
		challenge.Dislikes,
		func() string {
			if challenge.IsSolved {
				return "Solved ✅"
			}
			return "Not Solved"
		}(),
	)

	flagEntry := widget.NewEntry()
	flagEntry.SetPlaceHolder("Enter flag...")

	submitBtn := widget.NewButton("Submit Flag", func() {
		a.submitChallengeFlag(challenge.ID, challenge.Name, flagEntry.Text)
	})

	content := container.NewVBox(
		widget.NewLabel(info),
		widget.NewSeparator(),
		widget.NewLabel("Submit Flag:"),
		flagEntry,
		submitBtn,
	)

	dialog.ShowCustom("Challenge: "+challenge.Name, "Close", content, a.mainWindow)
}

func (a *App) submitChallengeFlag(challengeID int, name, flag string) {
	if flag == "" {
		dialog.ShowError(fmt.Errorf("flag cannot be empty"), a.mainWindow)
		return
	}

	a.setStatus(fmt.Sprintf("Submitting flag for %s...", name))

	go func() {
		resp, err := a.client.SubmitChallengeFlag(challengeID, flag)
		if err != nil {
			a.setStatus("Error: " + err.Error())
			dialog.ShowError(fmt.Errorf("failed to submit flag: %w", err), a.mainWindow)
			return
		}

		if resp.Success {
			a.setStatus(fmt.Sprintf("✓ Correct flag for %s!", name))
			dialog.ShowInformation("Success!", fmt.Sprintf("🎉 Correct flag!\n\n%s", resp.Message), a.mainWindow)
			a.refreshChallenges()
		} else {
			a.setStatus(fmt.Sprintf("✗ Incorrect flag for %s", name))
			dialog.ShowError(fmt.Errorf("incorrect flag: %s", resp.Message), a.mainWindow)
		}
	}()
}
