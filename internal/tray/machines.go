package tray

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"htb-tool/internal/api"

	"github.com/getlantern/systray"
)

// machineRow is one reusable menu entry, re-bound to a different machine on
// each refresh (systray cannot remove items, so we update them in place).
type machineRow struct {
	parent *systray.MenuItem
	info   *systray.MenuItem
	spawn  *systray.MenuItem
	stop   *systray.MenuItem
	reset  *systray.MenuItem
	submit *systray.MenuItem

	mu sync.Mutex
	m  api.Machine
}

func (r *machineRow) bind(m api.Machine) {
	r.mu.Lock()
	r.m = m
	r.mu.Unlock()
}

func (r *machineRow) current() api.Machine {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.m
}

type machineSection struct {
	parent     *systray.MenuItem
	activeHdr  *systray.MenuItem
	retiredHdr *systray.MenuItem
	active     []*machineRow
	retired    []*machineRow
}

func (t *TrayApp) buildMachineSection() *machineSection {
	s := &machineSection{}
	s.parent = systray.AddMenuItem("🖥️ Machines", "HTB machines")

	activeParent := s.parent.AddSubMenuItem("🟢 Active", "Active machines")
	s.activeHdr = activeParent.AddSubMenuItem("⏳ Loading…", "")
	s.activeHdr.Disable()
	s.active = t.newMachineRows(activeParent, maxActiveMachines)

	retiredParent := s.parent.AddSubMenuItem("🔴 Retired", "Retired machines")
	s.retiredHdr = retiredParent.AddSubMenuItem("⏳ Loading…", "")
	s.retiredHdr.Disable()
	s.retired = t.newMachineRows(retiredParent, maxRetiredMachines)

	return s
}

func (t *TrayApp) newMachineRows(parent *systray.MenuItem, n int) []*machineRow {
	rows := make([]*machineRow, n)
	for i := 0; i < n; i++ {
		r := &machineRow{}
		r.parent = parent.AddSubMenuItem("", "")
		r.info = r.parent.AddSubMenuItem("ℹ️ Info", "")
		r.spawn = r.parent.AddSubMenuItem("▶️ Spawn", "")
		r.stop = r.parent.AddSubMenuItem("⏹️ Stop", "")
		r.reset = r.parent.AddSubMenuItem("🔄 Reset", "")
		r.submit = r.parent.AddSubMenuItem("🚩 Submit Flag (clipboard)", "")
		r.parent.Hide()
		rows[i] = r
		go t.runMachineRow(r)
	}
	return rows
}

func (t *TrayApp) runMachineRow(r *machineRow) {
	for {
		select {
		case <-r.info.ClickedCh:
			t.machineInfo(r.current())
		case <-r.spawn.ClickedCh:
			t.machineSpawn(r.current())
		case <-r.stop.ClickedCh:
			t.machineStop(r.current())
		case <-r.reset.ClickedCh:
			t.machineReset(r.current())
		case <-r.submit.ClickedCh:
			t.machineSubmit(r.current())
		}
	}
}

// refreshMachines fetches active and retired machines concurrently and updates
// the menu in place - no process restart.
func (t *TrayApp) refreshMachines() {
	go func() {
		d, err := t.client.ListActiveMachines()
		if err != nil {
			t.machines.activeHdr.SetTitle("⚠️ active: " + truncate(err.Error(), 40))
			log.Println("machines active:", err)
			return
		}
		_ = t.machinesActive.Set(d)
		t.renderMachineRows(t.machines.active, t.machines.activeHdr, "active", capN(d, maxActiveMachines))
	}()
	go func() {
		d, err := t.client.ListRetiredMachines()
		if err != nil {
			t.machines.retiredHdr.SetTitle("⚠️ retired: " + truncate(err.Error(), 40))
			log.Println("machines retired:", err)
			return
		}
		_ = t.machinesRetired.Set(d)
		t.renderMachineRows(t.machines.retired, t.machines.retiredHdr, "retired", capN(d, maxRetiredMachines))
	}()
}

func (t *TrayApp) renderMachineRows(rows []*machineRow, hdr *systray.MenuItem, label string, data []api.Machine) {
	hdr.SetTitle(fmt.Sprintf("📊 %d %s", len(data), label))
	for i, r := range rows {
		if i < len(data) {
			m := data[i]
			r.bind(m)
			r.parent.SetTitle(machineTitle(m))
			r.parent.SetTooltip(fmt.Sprintf("%s · %s · %s", m.Name, m.OS, m.Difficulty))
			r.info.Show()
			r.submit.Show()
			if m.Spawned {
				r.spawn.Hide()
				r.stop.Show()
				r.reset.Show()
			} else {
				r.spawn.Show()
				r.stop.Hide()
				r.reset.Hide()
			}
			r.parent.Show()
		} else {
			r.parent.Hide()
		}
	}
}

func machineTitle(m api.Machine) string {
	owned := ""
	switch {
	case m.UserOwns && m.RootOwns:
		owned = " ✓✓"
	case m.UserOwns || m.RootOwns:
		owned = " ✓"
	}
	spawned := ""
	if m.Spawned {
		spawned = " ●"
	}
	return fmt.Sprintf("%s %s %s%s%s", getDifficultyEmoji(m.Difficulty), getOSEmoji(m.OS), truncate(m.Name, 26), owned, spawned)
}

func (t *TrayApp) machineInfo(m api.Machine) {
	owned := "none"
	switch {
	case m.UserOwns && m.RootOwns:
		owned = "user + root ✓"
	case m.UserOwns:
		owned = "user ✓"
	case m.RootOwns:
		owned = "root ✓"
	}
	state := "Retired"
	if m.Active {
		state = "Active"
	}
	ip := m.IP
	if ip == "" {
		ip = "-"
	}
	spawned := "no"
	if m.Spawned {
		spawned = "yes"
	}
	msg := fmt.Sprintf("OS: %s   Difficulty: %s\nStars: %.1f   %s\nOwned: %s   Spawned: %s\nIP: %s",
		m.OS, m.Difficulty, m.Stars, state, owned, spawned, ip)
	notify("🖥️ "+m.Name, msg)
}

// isServerErr reports whether err is an HTB API 5xx (server-side) error.
func isServerErr(err error) bool {
	var ae *api.APIError
	return errors.As(err, &ae) && ae.Status >= 500
}

func (t *TrayApp) machineSpawn(m api.Machine) {
	notify("Spawning", m.Name+"…")
	err := t.client.SpawnMachine(m.ID)
	if err != nil {
		// HTB runs one machine at a time on the dedicated instance. If another
		// machine is already active, that is the cause; say so plainly.
		if active, aerr := t.client.ActiveMachine(); aerr == nil && active != nil && active.ID != m.ID {
			notifyErr("Spawn failed", fmt.Sprintf("%s is already running. Stop it first, then spawn %s.", active.Name, m.Name))
			return
		}
		// Otherwise HTB likely returned a transient 5xx ("please try again").
		for attempt := 0; err != nil && isServerErr(err) && attempt < 2; attempt++ {
			time.Sleep(2 * time.Second)
			err = t.client.SpawnMachine(m.ID)
		}
		if err != nil {
			// The usual cause is no Machines VPN server selected; HTB then cannot
			// place the machine ("Failed to spawn on the Dedicated server").
			notifyErr("Spawn failed", "Pick a Machines VPN server first (VPN menu, then 'Switch to this server'), then retry. ("+truncate(err.Error(), 70)+")")
			return
		}
	}
	notify("Spawning", m.Name+" is starting (~1-2 min)")
	go func() {
		time.Sleep(4 * time.Second)
		t.refreshStatus()
		t.refreshMachines()
	}()
}

func (t *TrayApp) machineStop(m api.Machine) {
	notify("Stopping", m.Name+"…")
	if err := t.client.StopMachine(m.ID); err != nil {
		notifyErr("Stop failed", truncate(err.Error(), 120))
		return
	}
	notify("Stopped", m.Name)
	go func() {
		time.Sleep(1 * time.Second)
		t.refreshStatus()
		t.refreshMachines()
	}()
}

func (t *TrayApp) machineReset(m api.Machine) {
	notify("Resetting", m.Name+"…")
	if err := t.client.ResetMachine(m.ID); err != nil {
		notifyErr("Reset failed", truncate(err.Error(), 120))
		return
	}
	notify("Reset", m.Name+" has been reset")
}

func (t *TrayApp) machineSubmit(m api.Machine) {
	flag, err := clipboardRead()
	if err != nil || flag == "" {
		notifyErr("Submit Flag", "Copy the flag to your clipboard first")
		return
	}
	notify("Submitting", "Flag for "+m.Name+"…")
	resp, err := t.client.SubmitMachineFlag(m.ID, flag, 0)
	if err != nil {
		notifyErr("❌ Wrong flag / error", truncate(err.Error(), 120))
		return
	}
	if resp != nil && resp.Success {
		notify("✅ Correct!", m.Name+": "+resp.Message)
	} else if resp != nil {
		notifyErr("❌ Incorrect", resp.Message)
	}
	go func() {
		time.Sleep(1 * time.Second)
		t.refreshMachines()
	}()
}
