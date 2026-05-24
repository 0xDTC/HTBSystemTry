// Package tray implements the Hack The Box system-tray application: a native
// getlantern/systray menu backed by the stdlib-only HTB API client. Menus are
// built once and refreshed in place (no process restart).
package tray

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"htb-tool/internal/api"
	"htb-tool/internal/cache"
	"htb-tool/internal/config"

	"github.com/getlantern/systray"
)

const (
	maxActiveMachines  = 40
	maxRetiredMachines = 25
	maxActiveChalls    = 45
	maxRetiredChalls   = 25
	maxSherlocks       = 40
	cacheTTL           = 5 * time.Minute
)

// TrayApp owns the menu state, API client, and per-dataset caches.
type TrayApp struct {
	config *config.Config
	client *api.Client

	machinesActive  *cache.Store[[]api.Machine]
	machinesRetired *cache.Store[[]api.Machine]
	challActive     *cache.Store[[]api.Challenge]
	challRetired    *cache.Store[[]api.Challenge]
	sherlocksStore  *cache.Store[[]api.Sherlock]
	vpnStore        *cache.Store[[]api.VPNServer]

	statusItem  *systray.MenuItem
	searchItem  *systray.MenuItem
	tokenItem   *systray.MenuItem
	refreshItem *systray.MenuItem
	quitItem    *systray.MenuItem

	machines   *machineSection
	challenges *challengeSection
	sherlocks  *sherlockSection
	vpn        *vpnSection
}

// New returns a ready-to-run tray application.
func New() *TrayApp { return &TrayApp{} }

// Run starts the tray event loop and blocks until the user quits.
func (t *TrayApp) Run() { systray.Run(t.onReady, t.onExit) }

func (t *TrayApp) onExit() {}

func (t *TrayApp) onReady() {
	systray.SetIcon(generateIcon())
	systray.SetTitle("")
	systray.SetTooltip("HackTheBox")

	cfg, err := config.Load()
	if err != nil {
		log.Println("config load:", err)
		cfg = &config.Config{}
	}
	t.config = cfg

	cacheDir := filepath.Join(filepath.Dir(config.GetConfigPath()), "cache")
	t.machinesActive = cache.New[[]api.Machine](cacheTTL, filepath.Join(cacheDir, "machines_active.json"))
	t.machinesRetired = cache.New[[]api.Machine](cacheTTL, filepath.Join(cacheDir, "machines_retired.json"))
	t.challActive = cache.New[[]api.Challenge](cacheTTL, filepath.Join(cacheDir, "challenges_active.json"))
	t.challRetired = cache.New[[]api.Challenge](cacheTTL, filepath.Join(cacheDir, "challenges_retired.json"))
	t.vpnStore = cache.New[[]api.VPNServer](cacheTTL, filepath.Join(cacheDir, "vpn.json"))
	t.sherlocksStore = cache.New[[]api.Sherlock](cacheTTL, filepath.Join(cacheDir, "sherlocks.json"))
	t.machinesActive.Load()
	t.machinesRetired.Load()
	t.challActive.Load()
	t.challRetired.Load()
	t.sherlocksStore.Load()
	t.vpnStore.Load()

	if tok := readToken(cfg); tok != "" {
		t.client = api.NewClient(tok)
	}

	t.buildSkeleton()
	go t.handleTopLevel()

	if t.client != nil {
		t.renderFromCache()
		t.refreshAll()
	} else {
		t.showNoToken()
	}
}

// buildSkeleton creates every menu item exactly once. Dynamic lists are
// pre-allocated pools of reusable rows that later get updated in place.
func (t *TrayApp) buildSkeleton() {
	t.statusItem = systray.AddMenuItem("⏳ Loading…", "HTB - click to copy active machine IP")
	systray.AddSeparator()

	t.machines = t.buildMachineSection()
	t.challenges = t.buildChallengeSection()
	t.sherlocks = t.buildSherlockSection()
	t.vpn = t.buildVPNSection()

	systray.AddSeparator()
	t.searchItem = systray.AddMenuItem("🔎 Search (clipboard)", "Copy a query, then click to search HTB")
	t.tokenItem = systray.AddMenuItem("🔑 Set API Token (clipboard)", "Copy your HTB token, then click")
	t.refreshItem = systray.AddMenuItem("🔄 Refresh", "Reload everything now")
	systray.AddSeparator()
	t.quitItem = systray.AddMenuItem("❌ Quit", "Exit HTB Tool")
}

func (t *TrayApp) handleTopLevel() {
	for {
		select {
		case <-t.statusItem.ClickedCh:
			t.onStatusClick()
		case <-t.searchItem.ClickedCh:
			t.onSearch()
		case <-t.tokenItem.ClickedCh:
			t.onSetToken()
		case <-t.refreshItem.ClickedCh:
			if t.client == nil {
				notifyErr("HTB", "Set your API token first")
				continue
			}
			notify("HTB", "Refreshing…")
			t.refreshAll()
		case <-t.quitItem.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (t *TrayApp) refreshAll() {
	if t.client == nil {
		return
	}
	t.refreshMachines()
	t.refreshChallenges()
	t.refreshSherlocks()
	t.refreshVPN()
	t.refreshStatus()
}

// renderFromCache paints any disk-persisted data instantly at startup so the
// menu is populated before the first network round-trip completes.
func (t *TrayApp) renderFromCache() {
	if d, ok := t.machinesActive.Get(); ok {
		t.renderMachineRows(t.machines.active, t.machines.activeHdr, "active", capN(d, maxActiveMachines))
	}
	if d, ok := t.machinesRetired.Get(); ok {
		t.renderMachineRows(t.machines.retired, t.machines.retiredHdr, "retired", capN(d, maxRetiredMachines))
	}
	if d, ok := t.challActive.Get(); ok {
		t.renderChallengeRows(t.challenges.active, t.challenges.activeHdr, "active", capN(d, maxActiveChalls))
	}
	if d, ok := t.challRetired.Get(); ok {
		t.renderChallengeRows(t.challenges.retired, t.challenges.retiredHdr, "retired", capN(d, maxRetiredChalls))
	}
	if d, ok := t.sherlocksStore.Get(); ok {
		t.renderSherlockRows(capN(d, maxSherlocks))
	}
	if d, ok := t.vpnStore.Get(); ok {
		t.renderVPN(d)
	}
}

func (t *TrayApp) refreshStatus() {
	go func() {
		m, err := t.client.ActiveMachine()
		if err != nil {
			t.statusItem.SetTitle("🟢 Connected")
			systray.SetTitle("")
			systray.SetTooltip("HackTheBox")
			return
		}
		if m == nil {
			t.statusItem.SetTitle("🟢 Connected · no active machine")
			systray.SetTitle("")
			systray.SetTooltip("HackTheBox · no active machine")
			return
		}
		if m.IP == "" {
			t.statusItem.SetTitle("🟢 " + m.Name + " · spawning…")
			systray.SetTitle("▸ spawning…")
			systray.SetTooltip(m.Name + " · spawning…")
			return
		}
		t.statusItem.SetTitle(fmt.Sprintf("🟢 %s · %s", m.Name, m.IP))
		t.statusItem.SetTooltip("Click to copy the active machine's IP")
		// The active IP is shown three ways: the menu status line above, the
		// tray tooltip on hover (works on XFCE), and the icon label
		// (rendered next to the icon on GNOME/KDE; XFCE ignores SNI labels).
		systray.SetTitle("▸ " + m.IP)
		systray.SetTooltip(m.Name + " · " + m.IP)
	}()
}

func (t *TrayApp) onStatusClick() {
	if t.client == nil {
		notifyErr("HTB", "Set your API token first")
		return
	}
	go func() {
		m, err := t.client.ActiveMachine()
		if err != nil {
			notifyErr("HTB", truncate(err.Error(), 120))
			return
		}
		if m == nil {
			notify("HTB", "No active machine")
			return
		}
		if m.IP != "" {
			clipboardWrite(m.IP)
			notify("HTB", fmt.Sprintf("%s - %s (copied)", m.Name, m.IP))
		} else {
			notify("HTB", m.Name+" (no IP yet)")
		}
	}()
}

func (t *TrayApp) onSetToken() {
	tok, err := clipboardRead()
	if err != nil || tok == "" {
		notifyErr("HTB", "Clipboard empty - copy your API token first")
		return
	}
	t.config.APIToken = tok
	if err := t.config.Save(); err != nil {
		notifyErr("HTB", "Save failed: "+err.Error())
		return
	}
	t.client = api.NewClient(tok)
	t.enableSections()
	notify("HTB", "Token saved - loading data…")
	t.refreshAll()
}

func (t *TrayApp) onSearch() {
	if t.client == nil {
		notifyErr("HTB", "Set your API token first")
		return
	}
	q, err := clipboardRead()
	if err != nil || q == "" {
		notifyErr("HTB Search", "Copy a search term to your clipboard first")
		return
	}
	go func() {
		res, err := t.client.Search(q, nil)
		if err != nil {
			notifyErr("HTB Search", truncate(err.Error(), 120))
			return
		}
		var b strings.Builder
		add := func(label string, hits []api.SearchHit) {
			if len(hits) == 0 {
				return
			}
			names := make([]string, 0, 5)
			for i, h := range hits {
				if i >= 5 {
					break
				}
				names = append(names, h.Name)
			}
			fmt.Fprintf(&b, "%s: %s\n", label, strings.Join(names, ", "))
		}
		add("Machines", res.Machines)
		add("Challenges", res.Challenges)
		add("Sherlocks", res.Sherlocks)
		add("Users", res.Users)
		add("Teams", res.Teams)
		out := strings.TrimSpace(b.String())
		if out == "" {
			out = "No results for \"" + q + "\""
		}
		notify("🔎 "+q, out)
	}()
}

func (t *TrayApp) showNoToken() {
	t.statusItem.SetTitle("🔑 No API token - click ‘Set API Token’")
	t.machines.parent.SetTitle("🖥️ Machines (no token)")
	t.machines.parent.Disable()
	t.challenges.parent.SetTitle("🎯 Challenges (no token)")
	t.challenges.parent.Disable()
	t.sherlocks.parent.SetTitle("🔍 Sherlocks (no token)")
	t.sherlocks.parent.Disable()
	t.vpn.parent.SetTitle("🔐 VPN (no token)")
	t.vpn.parent.Disable()
}

func (t *TrayApp) enableSections() {
	t.statusItem.SetTitle("⏳ Loading…")
	t.machines.parent.SetTitle("🖥️ Machines")
	t.machines.parent.Enable()
	t.challenges.parent.SetTitle("🎯 Challenges")
	t.challenges.parent.Enable()
	t.sherlocks.parent.SetTitle("🔍 Sherlocks")
	t.sherlocks.parent.Enable()
	t.vpn.parent.SetTitle("🔐 VPN")
	t.vpn.parent.Enable()
}
