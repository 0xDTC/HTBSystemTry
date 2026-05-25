package tray

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"htb-tool/internal/api"

	"github.com/getlantern/systray"
)

type sherlockRow struct {
	parent   *systray.MenuItem
	info     *systray.MenuItem
	play     *systray.MenuItem
	download *systray.MenuItem

	mu sync.Mutex
	s  api.Sherlock
}

func (r *sherlockRow) bind(s api.Sherlock) {
	r.mu.Lock()
	r.s = s
	r.mu.Unlock()
}

func (r *sherlockRow) current() api.Sherlock {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.s
}

type sherlockSection struct {
	parent *systray.MenuItem
	hdr    *systray.MenuItem
	rows   []*sherlockRow
}

func (t *TrayApp) buildSherlockSection() *sherlockSection {
	s := &sherlockSection{}
	s.parent = systray.AddMenuItem("🔍 Sherlocks", "HTB Sherlocks (DFIR)")
	s.hdr = s.parent.AddSubMenuItem("⏳ Loading…", "")
	s.hdr.Disable()
	s.rows = t.newSherlockRows(s.parent, maxSherlocks)
	return s
}

func (t *TrayApp) newSherlockRows(parent *systray.MenuItem, n int) []*sherlockRow {
	rows := make([]*sherlockRow, n)
	for i := 0; i < n; i++ {
		r := &sherlockRow{}
		r.parent = parent.AddSubMenuItem("", "")
		r.info = r.parent.AddSubMenuItem("ℹ️ Info", "")
		r.play = r.parent.AddSubMenuItem("▶️ Play", "")
		r.download = r.parent.AddSubMenuItem("⬇️ Download (files + logo + info)", "")
		r.parent.Hide()
		rows[i] = r
		go t.runSherlockRow(r)
	}
	return rows
}

func (t *TrayApp) runSherlockRow(r *sherlockRow) {
	for {
		select {
		case <-r.info.ClickedCh:
			s := r.current()
			go t.sherlockInfo(s)
		case <-r.play.ClickedCh:
			s := r.current()
			go t.sherlockPlay(s)
		case <-r.download.ClickedCh:
			s := r.current()
			go t.sherlockDownload(s)
		}
	}
}

func (t *TrayApp) refreshSherlocks() {
	go func() {
		d, err := t.client.ListSherlocks()
		if err != nil {
			t.sherlocks.hdr.SetTitle("⚠️ " + truncate(err.Error(), 40))
			log.Println("sherlocks:", err)
			return
		}
		sort.Slice(d, func(i, j int) bool {
			if d[i].ReleaseDate != d[j].ReleaseDate {
				return d[i].ReleaseDate > d[j].ReleaseDate // newest first
			}
			return d[i].Name < d[j].Name
		})
		_ = t.sherlocksStore.Set(d)
		t.renderSherlockRows(capN(d, maxSherlocks))
	}()
}

func (t *TrayApp) renderSherlockRows(data []api.Sherlock) {
	t.sherlocks.hdr.SetTitle(fmt.Sprintf("📊 %d sherlocks", len(data)))
	for i, r := range t.sherlocks.rows {
		if i < len(data) {
			s := data[i]
			r.bind(s)
			r.parent.SetTitle(sherlockTitle(s))
			r.parent.SetTooltip(fmt.Sprintf("%s · %s · %s", s.Name, s.Category, s.Difficulty))
			r.parent.Show()
		} else {
			r.parent.Hide()
		}
	}
}

func sherlockTitle(s api.Sherlock) string {
	solved := "⚪"
	if s.Solved {
		solved = "✅"
	}
	return fmt.Sprintf("%s %s %s %s", solved, getDifficultyEmoji(s.Difficulty), getCategoryEmoji(s.Category), truncate(s.Name, 26))
}

func (t *TrayApp) sherlockInfo(s api.Sherlock) {
	desc := s.Description
	if det, err := t.client.GetSherlock(s.ID); err == nil && det.Description != "" {
		desc = det.Description
	}
	solved := "no"
	if s.Solved {
		solved = "yes ✓"
	}
	msg := fmt.Sprintf("Difficulty: %s   Category: %s\nRating: %.1f   Solves: %d   Progress: %d%%\nSolved: %s\n\n%s",
		s.Difficulty, s.Category, s.Rating, s.Solves, s.Progress, solved, truncate(desc, 220))
	notify("🔍 "+s.Name, msg)
}

func (t *TrayApp) sherlockPlay(s api.Sherlock) {
	if err := t.client.PlaySherlock(s.ID); err != nil {
		notifyErr("Sherlock", truncate(err.Error(), 120))
		return
	}
	notify("Sherlock", "Started "+s.Name)
}

// sherlockDownload creates ~/Downloads/<sherlock name>/ and saves the artifact
// archive, the logo image, and an info.txt into it.
func (t *TrayApp) sherlockDownload(s api.Sherlock) {
	notify("Sherlock", "Downloading "+s.Name+"…")

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("HOME")
	}
	dir := filepath.Join(home, "Downloads", sanitizeName(s.Name))
	if err := os.MkdirAll(dir, 0755); err != nil {
		notifyErr("Sherlock", "mkdir: "+err.Error())
		return
	}

	var saved []string

	// 1) artifact archive (temporary pre-signed link)
	if link, err := t.client.SherlockDownloadLink(s.ID); err != nil {
		notifyErr("Sherlock", "download link: "+truncate(err.Error(), 100))
	} else if data, fn, err := t.client.FetchURL(link); err != nil {
		notifyErr("Sherlock", "download: "+truncate(err.Error(), 100))
	} else {
		if fn == "" {
			fn = sanitizeName(s.Name) + ".zip"
		}
		if os.WriteFile(filepath.Join(dir, fn), data, 0644) == nil {
			saved = append(saved, fn)
		}
	}

	// 2) logo / avatar
	if s.Avatar != "" {
		logoURL := s.Avatar
		if strings.HasPrefix(logoURL, "/") {
			logoURL = "https://labs.hackthebox.com" + logoURL
		}
		if img, ifn, err := t.client.FetchURL(logoURL); err == nil && len(img) > 0 {
			ext := filepath.Ext(ifn)
			if ext == "" {
				ext = filepath.Ext(logoURL)
			}
			if ext == "" {
				ext = ".png"
			}
			if os.WriteFile(filepath.Join(dir, "logo"+ext), img, 0644) == nil {
				saved = append(saved, "logo"+ext)
			}
		}
	}

	// 3) info.txt (pull description from the detail endpoint)
	desc := s.Description
	if det, err := t.client.GetSherlock(s.ID); err == nil && det.Description != "" {
		desc = det.Description
	}
	info := fmt.Sprintf("%s\n\nDifficulty: %s\nCategory:   %s\nRating:     %.1f\nSolves:     %d\nProgress:   %d%%\nState:      %s\n\n%s\n",
		s.Name, s.Difficulty, s.Category, s.Rating, s.Solves, s.Progress, s.State, desc)
	if os.WriteFile(filepath.Join(dir, "info.txt"), []byte(info), 0644) == nil {
		saved = append(saved, "info.txt")
	}

	// 4) question.md (tasks / questions) - best-effort
	if tasks, err := t.client.SherlockTasks(s.ID); err == nil && len(tasks) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "# %s - Tasks\n\n", s.Name)
		for i, task := range tasks {
			title := strings.TrimSpace(task.Title)
			if title == "" {
				title = fmt.Sprintf("Task %d", i+1)
			}
			fmt.Fprintf(&b, "## Task %d: %s", i+1, title)
			if task.Completed {
				b.WriteString(" (completed)")
			}
			b.WriteString("\n")
			if task.Description != "" {
				fmt.Fprintf(&b, "%s\n", task.Description)
			}
			b.WriteString("\n")
		}
		if os.WriteFile(filepath.Join(dir, "question.md"), []byte(b.String()), 0644) == nil {
			saved = append(saved, "question.md")
		}
	}

	if len(saved) == 0 {
		notifyErr("Sherlock", "Nothing could be downloaded for "+s.Name)
		return
	}
	notify("✅ Sherlock saved", fmt.Sprintf("%s\n%s\n(%s)", s.Name, dir, strings.Join(saved, ", ")))
}

// sanitizeName makes a Sherlock name safe to use as a directory name.
func sanitizeName(name string) string {
	repl := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_", "\n", " ", "\t", " ",
	)
	s := strings.TrimSpace(repl.Replace(name))
	if s == "" {
		s = "sherlock"
	}
	return s
}
