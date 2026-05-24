# HTB Tray

A fast, native Linux system-tray client for Hack The Box. Browse and control machines, challenges, sherlocks, and VPN servers from a tray menu, without opening a browser.

Built with the Go standard library plus a single tray dependency (`getlantern/systray`). No web view, no Electron, no heavy GUI toolkit. The binary is about 10 MB.

<img width="270" height="304" alt="image" src="https://github.com/user-attachments/assets/55c706e3-ac46-4a0c-aadc-e12dc9c540e6" />

## Features

- **Machines**: active and retired lists with difficulty, OS, and owned markers; spawn, stop, reset; submit user/root flags from the clipboard. The status line at the top shows the spawned machine and its IP (for example `Reactor · 10.129.x.x`); click it to copy the IP.
- **Challenges**: active and retired, with difficulty, category, points, and solved markers; submit flags from the clipboard.
- **Sherlocks**: browse, view info, and play. "Download" creates `~/Downloads/<sherlock name>/` and saves the artifact archive, the logo image, and an `info.txt`.
- **VPN**: servers grouped by product (Machines, Competitive, Starting Point, Fortresses) and region, each showing live client load; download TCP or UDP configs; switch server; the assigned server is marked.
- **Search**: copy a query, click Search, get matching machines, challenges, sherlocks, users, and teams.
- **Instant refresh**: menus are built once and updated in place (no process restart). Data loads in the background and is cached on disk, so the menu appears immediately on launch.
- **Lean integration**: desktop notifications via `notify-send` and clipboard input via `xclip`, so no extra dialog dependency is required. API requests are paced and retried so refreshing never trips HTB rate limits.

## Menu structure

```
HTB (tray icon)
- Status: active machine and IP (click to copy)
- Machines
  - Active:  Info, Spawn, Stop, Reset, Submit Flag (clipboard)
  - Retired
- Challenges
  - Active:  Info, Submit Flag (clipboard)
  - Retired
- Sherlocks: Info, Play, Download (files + logo + info)
- VPN
  - Product (Machines / Competitive / Starting Point / Fortresses)
    - Region (EU / US / AU / SG)
      - Server: Download TCP, Download UDP, Switch
- Search (clipboard)
- Set API Token (clipboard)
- Refresh
- Quit
```

## Requirements

- Linux with a system tray host (StatusNotifier): XFCE, KDE, GNOME with an AppIndicator extension, etc.
- Runtime tools: `notify-send` (libnotify) for notifications, `xclip` for clipboard access.
- To build: Go 1.24 or newer.

## Install

```sh
git clone <repo-url> HTBSystemTry
cd HTBSystemTry
go build -o ~/.local/bin/htb-tray ./cmd
~/.local/bin/htb-tray
```

## Autostart on login

An XDG autostart entry launches the tray automatically after login (it runs inside the graphical session, so it gets the display and D-Bus it needs):

`~/.config/autostart/htb-tray.desktop`

```ini
[Desktop Entry]
Type=Application
Name=HTB Tray
Comment=Hack The Box system tray (machines, challenges, sherlocks, VPN)
Exec=/home/<user>/.local/bin/htb-tray
Terminal=false
X-GNOME-Autostart-enabled=true
Categories=Network;Security;
```

To disable autostart, delete that file or untick it in your desktop's startup settings (for example XFCE: Settings, Session and Startup, Application Autostart).

## Configuration

Provide your HTB API token (from your HTB profile settings, App Tokens) in any one of these ways:

1. Click "Set API Token (clipboard)" in the tray after copying the token.
2. Set `HTB_TOKEN` in the environment.
3. Put it in `~/.config/htb-tool/config.json`:

```json
{ "api_token": "your-token", "vpn_directory": "/home/<user>/Downloads/htb-vpn" }
```

VPN configs save to `~/Downloads/htb-vpn` by default.

## Usage notes

- **Submitting flags**: copy the flag to the clipboard, then click "Submit Flag (clipboard)" on the machine or challenge.
- **VPN downloads**: HTB only allows downloading the config for the server you are assigned to. For a different server, click "Switch to this server" first, then download.
- **VPN latency**: per-server ping is intentionally not shown. HTB does not expose pingable hostnames for arbitrary servers, so (like the HTB web platform) only client load is displayed.
- **Pro Labs VPN**: not yet listed; Pro Labs use a separate per-prolab endpoint.
- **Active machine IP**: shown in the menu status line and in the tray icon's tooltip (hover). It is also set as the icon label, which appears next to the icon on desktops that render labels (GNOME, KDE). XFCE's tray ignores icon labels (and scales icons to a square), so on XFCE read the IP from the tooltip or the menu.

## Architecture

- `internal/api`: standard-library HTB API client (v4 and v5) with request pacing and rate-limit retry.
- `internal/cache`: generic, disk-persisted TTL cache for stale-while-revalidate loading.
- `internal/tray`: the systray menu, pre-allocated and updated in place.
- `cmd`: entry point.
