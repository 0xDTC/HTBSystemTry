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

// vpnSlot is a reusable menu entry under a product submenu. It acts either as a
// region header (disabled, no actions) or a server row (Download TCP/UDP and
// Switch).
type vpnSlot struct {
	parent *systray.MenuItem
	tcp    *systray.MenuItem
	udp    *systray.MenuItem
	sw     *systray.MenuItem

	mu       sync.Mutex
	s        api.VPNServer
	isServer bool
}

func (sl *vpnSlot) current() (api.VPNServer, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.s, sl.isServer
}

func (sl *vpnSlot) setHeader(label string) {
	sl.mu.Lock()
	sl.isServer = false
	sl.mu.Unlock()
	sl.parent.SetTitle(label)
	sl.parent.Disable()
	sl.tcp.Hide()
	sl.udp.Hide()
	sl.sw.Hide()
	sl.parent.Show()
}

func (sl *vpnSlot) setServer(s api.VPNServer, title string) {
	sl.mu.Lock()
	sl.s = s
	sl.isServer = true
	sl.mu.Unlock()
	sl.parent.SetTitle(title)
	sl.parent.Enable()
	sl.tcp.Show()
	sl.udp.Show()
	sl.sw.Show()
	sl.parent.Show()
}

type vpnProductMenu struct {
	parent *systray.MenuItem
	hdr    *systray.MenuItem
	slots  []*vpnSlot
}

type vpnSection struct {
	parent   *systray.MenuItem
	products []*vpnProductMenu
}

// vpnPoolSizes defines pre-allocated product submenus (descending). The product
// with the most servers (Machines) is matched to the largest pool.
var vpnPoolSizes = []int{36, 20, 16, 14, 12, 10, 10, 10}

// vpnProductName maps API product keys to a friendly, icon-prefixed label.
var vpnProductName = map[string]string{
	"lab": "🖥️ Machines", "labs": "🖥️ Machines",
	"starting_point": "🚩 Starting Point",
	"release_arena":  "🆕 Release Arena", "arena": "🆕 Release Arena", "season": "🆕 Release Arena",
	"competitive": "⚔️ Competitive",
	"prolab":      "🧪 Pro Labs", "prolabs": "🧪 Pro Labs", "pro_labs": "🧪 Pro Labs",
	"fortresses": "🏰 Fortresses", "fortress": "🏰 Fortresses",
	"endgames": "🛡️ Endgame", "endgame": "🛡️ Endgame",
}

func (t *TrayApp) buildVPNSection() *vpnSection {
	s := &vpnSection{}
	s.parent = systray.AddMenuItem("🔐 VPN", "HTB VPN servers")

	for _, n := range vpnPoolSizes {
		pm := &vpnProductMenu{}
		pm.parent = s.parent.AddSubMenuItem("", "")
		pm.hdr = pm.parent.AddSubMenuItem("⏳ Loading…", "")
		pm.hdr.Disable()
		pm.slots = make([]*vpnSlot, n)
		for i := 0; i < n; i++ {
			sl := &vpnSlot{}
			sl.parent = pm.parent.AddSubMenuItem("", "")
			sl.tcp = sl.parent.AddSubMenuItem("📥 Download TCP", "")
			sl.udp = sl.parent.AddSubMenuItem("📥 Download UDP", "")
			sl.sw = sl.parent.AddSubMenuItem("🔀 Switch to this server", "")
			sl.parent.Hide()
			pm.slots[i] = sl
			go t.runVPNSlot(sl)
		}
		pm.parent.Hide()
		s.products = append(s.products, pm)
	}
	return s
}

func (t *TrayApp) runVPNSlot(sl *vpnSlot) {
	for {
		select {
		case <-sl.tcp.ClickedCh:
			if s, ok := sl.current(); ok {
				go t.downloadVPNFile(s, true)
			}
		case <-sl.udp.ClickedCh:
			if s, ok := sl.current(); ok {
				go t.downloadVPNFile(s, false)
			}
		case <-sl.sw.ClickedCh:
			if s, ok := sl.current(); ok {
				go t.vpnSwitch(s)
			}
		}
	}
}

func (t *TrayApp) refreshVPN() {
	go func() {
		d, err := t.client.ListVPNServers()
		if err != nil {
			t.vpn.products[0].hdr.SetTitle("⚠️ " + truncate(err.Error(), 40))
			log.Println("vpn servers:", err)
			return
		}
		_ = t.vpnStore.Set(d)
		t.renderVPN(d)
	}()
}

// renderVPN groups servers by product, matches each product to a pre-allocated
// submenu (largest first), and lays the servers out by region. Per-server
// latency is intentionally not shown: HTB does not expose pingable hostnames
// for arbitrary servers, so (like the web platform) only client load is shown.
func (t *TrayApp) renderVPN(servers []api.VPNServer) {
	byProduct := map[string][]api.VPNServer{}
	for _, s := range servers {
		byProduct[s.Product] = append(byProduct[s.Product], s)
	}
	keys := make([]string, 0, len(byProduct))
	for k := range byProduct {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(byProduct[keys[i]]) != len(byProduct[keys[j]]) {
			return len(byProduct[keys[i]]) > len(byProduct[keys[j]])
		}
		return keys[i] < keys[j]
	})

	for idx, pm := range t.vpn.products {
		if idx >= len(keys) {
			pm.parent.Hide()
			continue
		}
		key := keys[idx]
		list := byProduct[key]
		pm.parent.SetTitle(fmt.Sprintf("%s (%d)", vpnProductLabel(key), len(list)))
		pm.hdr.SetTitle(fmt.Sprintf("📊 %d servers", len(list)))
		pm.parent.Show()

		rows := vpnDisplayRows(list)
		for i, sl := range pm.slots {
			if i < len(rows) {
				r := rows[i]
				if r.header {
					sl.setHeader(r.label)
				} else {
					sl.setServer(r.server, vpnServerLabel(r.server))
				}
			} else {
				sl.parent.Hide()
			}
		}
	}
}

type vpnDisplayRow struct {
	header bool
	label  string
	server api.VPNServer
}

func vpnDisplayRows(list []api.VPNServer) []vpnDisplayRow {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Location != list[j].Location {
			return list[i].Location < list[j].Location
		}
		if list[i].Tier != list[j].Tier {
			return list[i].Tier < list[j].Tier
		}
		return list[i].FriendlyName < list[j].FriendlyName
	})
	var rows []vpnDisplayRow
	lastRegion := "\x00"
	for _, s := range list {
		if s.Location != lastRegion {
			rows = append(rows, vpnDisplayRow{header: true, label: regionHeader(s.Location)})
			lastRegion = s.Location
		}
		rows = append(rows, vpnDisplayRow{server: s})
	}
	return rows
}

func vpnServerLabel(s api.VPNServer) string {
	label := fmt.Sprintf("%s %s · %d clients", loadEmoji(s), truncate(s.FriendlyName, 22), s.CurrentClients)
	if s.Assigned {
		label += " ✅"
	}
	return label
}

func vpnProductLabel(key string) string {
	if n, ok := vpnProductName[strings.ToLower(key)]; ok {
		return n
	}
	parts := strings.Fields(strings.ReplaceAll(key, "_", " "))
	if len(parts) == 0 {
		return "🔐 VPN"
	}
	for i, p := range parts {
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return "🔐 " + strings.Join(parts, " ")
}

func regionHeader(loc string) string {
	if loc == "" {
		loc = "Other"
	}
	return fmt.Sprintf("%s  %s", regionEmoji(loc), loc)
}

func regionEmoji(loc string) string {
	switch strings.ToUpper(loc) {
	case "EU":
		return "🇪🇺"
	case "US":
		return "🇺🇸"
	case "AU":
		return "🇦🇺"
	case "SG":
		return "🇸🇬"
	default:
		return "🌍"
	}
}

func loadEmoji(s api.VPNServer) string {
	switch {
	case s.Full || s.CurrentClients > 100:
		return "🔴"
	case s.CurrentClients > 50:
		return "🟡"
	default:
		return "🟢"
	}
}

func (t *TrayApp) vpnDir() string {
	if d := t.config.VPNDirectory; d != "" {
		return d
	}
	return filepath.Join(os.Getenv("HOME"), "Downloads", "htb-vpn")
}

// downloadVPNFile downloads a server's .ovpn. HTB only allows this for the
// server you are assigned to, so a "not assigned" error is turned into a hint
// to switch first.
func (t *TrayApp) downloadVPNFile(s api.VPNServer, tcp bool) {
	proto := "udp"
	if tcp {
		proto = "tcp"
	}
	notify("VPN", fmt.Sprintf("Downloading %s (%s)…", s.FriendlyName, proto))
	data, err := t.client.DownloadVPN(s.ID, tcp)
	if err != nil {
		msg := truncate(err.Error(), 120)
		if strings.Contains(err.Error(), "not assigned") || strings.Contains(err.Error(), "404") {
			msg = "Not assigned to this server. Use 'Switch to this server' first, then download."
		}
		notifyErr("VPN download", msg)
		return
	}
	dir := t.vpnDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		notifyErr("VPN", "mkdir: "+err.Error())
		return
	}
	fn := filepath.Join(dir, fmt.Sprintf("htb_%d_%s.ovpn", s.ID, proto))
	if err := os.WriteFile(fn, data, 0600); err != nil {
		notifyErr("VPN", "write: "+err.Error())
		return
	}
	notify("VPN saved", fn)
}

func (t *TrayApp) vpnSwitch(s api.VPNServer) {
	if err := t.client.SwitchVPN(s.ID); err != nil {
		notifyErr("VPN switch failed", truncate(err.Error(), 120))
		return
	}
	notify("VPN", "Switched to "+s.FriendlyName+". You can now download its config.")
	go t.refreshVPN()
}
