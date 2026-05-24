package tray

import (
	"fmt"
	"log"
	"sync"
	"time"

	"htb-tool/internal/api"

	"github.com/getlantern/systray"
)

type challengeRow struct {
	parent *systray.MenuItem
	info   *systray.MenuItem
	submit *systray.MenuItem

	mu sync.Mutex
	c  api.Challenge
}

func (r *challengeRow) bind(c api.Challenge) {
	r.mu.Lock()
	r.c = c
	r.mu.Unlock()
}

func (r *challengeRow) current() api.Challenge {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.c
}

type challengeSection struct {
	parent     *systray.MenuItem
	activeHdr  *systray.MenuItem
	retiredHdr *systray.MenuItem
	active     []*challengeRow
	retired    []*challengeRow
}

func (t *TrayApp) buildChallengeSection() *challengeSection {
	s := &challengeSection{}
	s.parent = systray.AddMenuItem("🎯 Challenges", "HTB challenges")

	activeParent := s.parent.AddSubMenuItem("🟢 Active", "Active challenges")
	s.activeHdr = activeParent.AddSubMenuItem("⏳ Loading…", "")
	s.activeHdr.Disable()
	s.active = t.newChallengeRows(activeParent, maxActiveChalls)

	retiredParent := s.parent.AddSubMenuItem("🔴 Retired", "Retired challenges")
	s.retiredHdr = retiredParent.AddSubMenuItem("⏳ Loading…", "")
	s.retiredHdr.Disable()
	s.retired = t.newChallengeRows(retiredParent, maxRetiredChalls)

	return s
}

func (t *TrayApp) newChallengeRows(parent *systray.MenuItem, n int) []*challengeRow {
	rows := make([]*challengeRow, n)
	for i := 0; i < n; i++ {
		r := &challengeRow{}
		r.parent = parent.AddSubMenuItem("", "")
		r.info = r.parent.AddSubMenuItem("ℹ️ Info", "")
		r.submit = r.parent.AddSubMenuItem("🚩 Submit Flag (clipboard)", "")
		r.parent.Hide()
		rows[i] = r
		go t.runChallengeRow(r)
	}
	return rows
}

func (t *TrayApp) runChallengeRow(r *challengeRow) {
	for {
		select {
		case <-r.info.ClickedCh:
			t.challengeInfo(r.current())
		case <-r.submit.ClickedCh:
			t.challengeSubmit(r.current())
		}
	}
}

func (t *TrayApp) refreshChallenges() {
	go func() {
		d, err := t.client.ListChallenges(false)
		if err != nil {
			t.challenges.activeHdr.SetTitle("⚠️ active: " + truncate(err.Error(), 40))
			log.Println("challenges active:", err)
			return
		}
		_ = t.challActive.Set(d)
		t.renderChallengeRows(t.challenges.active, t.challenges.activeHdr, "active", capN(d, maxActiveChalls))
	}()
	go func() {
		d, err := t.client.ListChallenges(true)
		if err != nil {
			t.challenges.retiredHdr.SetTitle("⚠️ retired: " + truncate(err.Error(), 40))
			log.Println("challenges retired:", err)
			return
		}
		_ = t.challRetired.Set(d)
		t.renderChallengeRows(t.challenges.retired, t.challenges.retiredHdr, "retired", capN(d, maxRetiredChalls))
	}()
}

func (t *TrayApp) renderChallengeRows(rows []*challengeRow, hdr *systray.MenuItem, label string, data []api.Challenge) {
	hdr.SetTitle(fmt.Sprintf("📊 %d %s", len(data), label))
	for i, r := range rows {
		if i < len(data) {
			c := data[i]
			r.bind(c)
			r.parent.SetTitle(challengeTitle(c))
			r.parent.SetTooltip(fmt.Sprintf("%s · %s · %s", c.Name, c.Category, c.Difficulty))
			r.parent.Show()
		} else {
			r.parent.Hide()
		}
	}
}

func challengeTitle(c api.Challenge) string {
	solved := "⚪"
	if c.Solved {
		solved = "✅"
	}
	pts := ""
	if c.Points > 0 {
		pts = fmt.Sprintf(" (%dp)", c.Points)
	}
	return fmt.Sprintf("%s %s %s %s%s", solved, getDifficultyEmoji(c.Difficulty), getCategoryEmoji(c.Category), truncate(c.Name, 24), pts)
}

func (t *TrayApp) challengeInfo(c api.Challenge) {
	solved := "no"
	if c.Solved {
		solved = "yes ✓"
	}
	msg := fmt.Sprintf("Category: %s   Difficulty: %s\nPoints: %d   Solves: %d   Likes: %d\nSolved: %s",
		c.Category, c.Difficulty, c.Points, c.Solves, c.Likes, solved)
	notify("🎯 "+c.Name, msg)
}

func (t *TrayApp) challengeSubmit(c api.Challenge) {
	flag, err := clipboardRead()
	if err != nil || flag == "" {
		notifyErr("Submit Flag", "Copy the flag to your clipboard first")
		return
	}
	notify("Submitting", "Flag for "+c.Name+"…")
	resp, err := t.client.SubmitChallengeFlag(c.ID, flag)
	if err != nil {
		notifyErr("❌ Wrong flag / error", truncate(err.Error(), 120))
		return
	}
	if resp != nil && resp.Success {
		notify("✅ Correct!", c.Name+": "+resp.Message)
	} else if resp != nil {
		notifyErr("❌ Incorrect", resp.Message)
	}
	go func() {
		time.Sleep(1 * time.Second)
		t.refreshChallenges()
	}()
}
