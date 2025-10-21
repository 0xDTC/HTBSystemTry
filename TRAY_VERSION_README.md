# HTB Tool - System Tray Version

A lightweight system tray application for managing HackTheBox machines, challenges, and VPN connections - **directly from your panel dropdown menu!**

## 🎯 What It Looks Like

HTB Tool appears as an icon in your system tray (panel). When you click it, you get a dropdown menu hierarchy:

```
[HTB] (System Tray Icon)
  │
  ├─ 👤 User Info
  ├─ 🔑 API Key
  │
  ├─ 🔐 VPN
  │   ├─ 🟢 US Free 1 (45 clients)
  │   │   ├─ 📥 Download TCP
  │   │   └─ 📥 Download UDP
  │   ├─ 🟡 US VIP 1 (78 clients)
  │   │   ├─ 📥 Download TCP
  │   │   └─ 📥 Download UDP
  │   ├─ ...
  │   ├─ ────────────
  │   ├─ 📥 Download All (TCP)
  │   └─ 📥 Download All (UDP)
  │
  ├─ 🖥️  Machines
  │   ├─ 🟢 Active Machines
  │   │   ├─ 📊 125 Active machines
  │   │   ├─ ────────────
  │   │   ├─ 🟢 🐧 Keeper
  │   │   │   ├─ ℹ️  Info
  │   │   │   ├─ ▶️  Spawn
  │   │   │   └─ 🚩 Submit Flag
  │   │   ├─ 🟡 🪟 Manager
  │   │   │   ├─ ℹ️  Info
  │   │   │   ├─ ⏹️  Stop
  │   │   │   ├─ 🔄 Reset
  │   │   │   └─ 🚩 Submit Flag
  │   │   └─ ... (20 most recent)
  │   │
  │   └─ 🔴 Retired Machines
  │       ├─ 📊 380 Retired machines
  │       └─ ... (20 most recent)
  │
  ├─ 🎯 Challenges
  │   ├─ 🟢 Active Challenges
  │   │   ├─ 📊 45 Active challenges
  │   │   ├─ ────────────
  │   │   ├─ 🌐 Web (12)
  │   │   │   ├─ 🟢 🟢 Flag Command (10pts)
  │   │   │   │   ├─ ℹ️  Info
  │   │   │   │   └─ 🚩 Submit Flag
  │   │   │   └─ ...
  │   │   ├─ 🔐 Crypto (8)
  │   │   ├─ 💥 Pwn (10)
  │   │   └─ ...
  │   │
  │   └─ 🔴 Retired Challenges
  │       └─ ... (20 most recent)
  │
  ├─ 🔄 Refresh All
  └─ ❌ Quit
```

## 🚀 Installation

### Quick Install

```bash
cd /media/kali/3315E7784CD31C71/Scripts/htb-tool
./install.sh
```

### Manual Build

```bash
# Install dependencies
sudo apt install -y libxxf86vm-dev libxcursor-dev libxinerama-dev \
                     libxi-dev libxrandr-dev libgl1-mesa-dev \
                     libayatana-appindicator3-dev zenity notify-osd

# Build
go mod tidy
go build -o htb-tool ./cmd

# Install
sudo install -m 755 htb-tool /usr/local/bin/htb-tool

# Create desktop entry for autostart
cat > ~/.local/share/applications/htb-tool.desktop <<'EOF'
[Desktop Entry]
Name=HTB Tool
Comment=HackTheBox System Tray Tool
Exec=/usr/local/bin/htb-tool
Icon=security-high
Terminal=false
Type=Application
Categories=Network;Security;
StartupNotify=false
X-GNOME-Autostart-enabled=true
EOF

# Enable autostart
cp ~/.local/share/applications/htb-tool.desktop ~/.config/autostart/
```

## ⚙️ Configuration

### First Time Setup

1. Create config directory and file:

```bash
mkdir -p ~/.config/htb-tool

cat > ~/.config/htb-tool/config.json <<'EOF'
{
  "api_token": "YOUR_HTB_API_TOKEN_HERE",
  "vpn_directory": "/home/$USER/Downloads/htb-vpn",
  "last_protocol": "tcp"
}
EOF

chmod 600 ~/.config/htb-tool/config.json
```

2. Get your HTB API token from: https://www.hackthebox.com/home/settings
   - Scroll to "Create App Token"
   - Click "Generate"
   - Copy the token (starts with `eyJ...`)
   - Replace `YOUR_HTB_API_TOKEN_HERE` in the config file

3. Launch HTB Tool:

```bash
htb-tool &
```

The icon will appear in your system tray!

## 📖 Usage

### System Tray Icon

Look for the **HTB** icon in your system tray (usually top-right corner of your panel, near network/volume icons).

- **Left-click**: Open dropdown menu
- **Menu stays open**: Navigate through submenus
- **Click outside**: Close menu

### Menu Structure

#### 👤 User Info
- Displays your HTB username and stats
- Click to view detailed profile info

#### 🔑 API Key
- Click to get instructions for updating API token
- Edit `~/.config/htb-tool/config.json` to update

#### 🔐 VPN
- **Server List**: All available VPN servers with real-time client count
- **Color Indicators**:
  - 🟢 Low load (< 50 clients)
  - 🟡 Medium load (50-100 clients)
  - 🔴 High load (> 100 clients)
- **Per-Server**: Hover → Click server → Choose TCP or UDP
- **Bulk Download**: Click "Download All (TCP)" or "Download All (UDP)"
- **Download Location**: `~/Downloads/htb-vpn/` (configurable)

####  🖥️ Machines

**Active Machines** (sorted newest → oldest):
- 🟢 Easy | 🟡 Medium | 🟠 Hard | 🔴 Insane
- 🐧 Linux | 🪟 Windows | 😈 FreeBSD

Per-machine actions:
- **ℹ️ Info**: View machine details (OS, difficulty, IP, owns, rating)
- **▶️ Spawn**: Start the machine (~2-3 min wait)
- **⏹️ Stop**: Terminate running machine
- **🔄 Reset**: Reset machine to clean state
- **🚩 Submit Flag**: Dialog to enter and submit user/root flags

**Retired Machines**: Same structure, historical machines

#### 🎯 Challenges

**Active Challenges** (grouped by category):
- 🌐 Web | 🔐 Crypto | 💥 Pwn | 🔄 Reversing
- 🔍 Forensics | 🕵️ OSINT | 📱 Mobile | 🔧 Hardware

Per-challenge actions:
- **ℹ️ Info**: View challenge details (points, solves, difficulty)
- **🚩 Submit Flag**: Dialog to enter flag

**Retired Challenges**: All retired challenges (not categorized)

#### 🔄 Refresh All
- Reloads all data from HTB API
- Rebuilds menus with latest info
- **Note**: Currently restarts the app (will fix in next version)

#### ❌ Quit
- Exits the application
- Removes icon from system tray

## 🎮 Workflows

### Example 1: Quick Machine Spawn

```
1. Click HTB tray icon
2. Hover: Machines → Active Machines
3. Find machine (e.g., "Keeper")
4. Hover: 🟢 🐧 Keeper
5. Click: ▶️ Spawn
6. Notification: "Keeper is spawning!"
7. Wait ~2 minutes
8. Hover: Keeper → Click: ℹ️ Info
9. See IP address in notification
10. SSH/pwn the machine
```

### Example 2: Submit Flag

```
1. After pwning machine...
2. Click HTB tray icon
3. Machines → Active Machines → [Machine Name]
4. Click: 🚩 Submit Flag
5. Dialog appears
6. Enter flag
7. Select: User Flag or Root Flag
8. Notification: ✓ Correct! or ❌ Incorrect
```

### Example 3: Download VPN (Single Region)

```
1. Click HTB tray icon
2. Hover: 🔐 VPN
3. Find region (e.g., "US Free 1")
4. Hover: 🟢 US Free 1 (45 clients)
5. Click: 📥 Download TCP
6. Notification: "Downloaded to ~/Downloads/htb-vpn/"
7. Use: sudo openvpn ~/Downloads/htb-vpn/htb_vpn_123_tcp.ovpn
```

### Example 4: Download All VPNs

```
1. Click HTB tray icon
2. Hover: 🔐 VPN
3. Scroll to bottom
4. Click: 📥 Download All (TCP)
5. Notification: "Downloading all 20 VPN configs..."
6. Wait ~10 seconds
7. Notification: "Downloaded 20/20 configs"
8. All configs saved to ~/Downloads/htb-vpn/
```

### Example 5: Challenge Hunting

```
1. Click HTB tray icon
2. Hover: 🎯 Challenges → 🟢 Active Challenges
3. Browse categories (e.g., 🌐 Web)
4. Find challenge
5. Click challenge name
6. Click: ℹ️ Info (view details)
7. Solve challenge offline
8. Return to menu
9. Click: 🚩 Submit Flag
10. Enter flag
11. Instant feedback!
```

## 🔔 Notifications

HTB Tool uses desktop notifications for all actions:

- **Success**: ✓ Green notifications
- **Errors**: ❌ Red notifications
- **Progress**: ℹ️ Blue notifications
- **Info**: 📊 Gray notifications

Requirements:
- `notify-send` (usually pre-installed)
- Notification daemon (GNOME, KDE, XFCE, etc.)

## 📊 Data Hierarchy

```
HTB Tool
├─ User Info
│   └─ Profile data from /api/v4/profile/progress/overview
│
├─ VPN
│   ├─ Server list from /api/v4/access/servers
│   └─ Downloads via /api/v4/access/ovpnfile/{id}/{protocol}
│
├─ Machines
│   ├─ List from /api/v4/machine/list
│   ├─ Spawn via /api/v4/machine/play/{id}
│   ├─ Stop via /api/v4/machine/stop/{id}
│   ├─ Reset via /api/v4/machine/reset/{id}
│   └─ Submit via /api/v4/machine/own
│
└─ Challenges
    ├─ List from /api/v4/challenge/list
    └─ Submit via /api/v4/challenge/own
```

## 🎨 Menu Features

### Emoji Indicators

**Difficulty**:
- 🟢 Easy
- 🟡 Medium
- 🟠 Hard
- 🔴 Insane

**Operating System**:
- 🐧 Linux
- 🪟 Windows
- 😈 FreeBSD
- 💻 Other

**Challenge Categories**:
- 🌐 Web
- 🔐 Crypto
- 💥 Pwn
- 🔄 Reversing
- 🔍 Forensics
- 🕵️ OSINT
- 📱 Mobile
- 🔧 Hardware
- 🎲 Misc

**Status**:
- ✅ Solved
- ⚪ Not solved
- 🟢 Active
- 🔴 Retired

### Smart Sorting

- **Machines**: Newest releases first (most recent → oldest)
- **Challenges**: Grouped by category, then by difficulty
- **VPN Servers**: Sorted by current load (lowest → highest)

### Menu Limits

To keep menus manageable:
- Active Machines: Shows 20 most recent
- Retired Machines: Shows 20 most recent
- Challenges per category: Shows 10 per category
- Full list available via HTB website

## 🔧 Troubleshooting

### Icon doesn't appear in tray

**Check system tray support**:
```bash
# Install tray support (if missing)
sudo apt install libayatana-appindicator3-1

# Restart app
killall htb-tool
htb-tool &
```

**Enable system tray** (GNOME):
```bash
# Install extension
sudo apt install gnome-shell-extension-appindicator

# Enable it
gnome-extensions enable appindicatorsupport@rgcjonas.gmail.com

# Restart GNOME Shell: Alt+F2 → type 'r' → Enter
```

### "No API Key" error

Edit config:
```bash
nano ~/.config/htb-tool/config.json
```

Add your token:
```json
{
  "api_token": "eyJhbGc...YOUR_REAL_TOKEN_HERE",
  "vpn_directory": "/home/kali/Downloads/htb-vpn"
}
```

Restart:
```bash
killall htb-tool && htb-tool &
```

### VPN download fails

Check permissions:
```bash
mkdir -p ~/Downloads/htb-vpn
chmod 755 ~/Downloads/htb-vpn
```

Verify subscription:
- Free users: Limited VPN access
- VIP/VIP+: Full access

### Flag submission dialog doesn't appear

Install zenity:
```bash
sudo apt install zenity
```

### Notifications don't work

Install notification daemon:
```bash
sudo apt install notification-daemon
# OR
sudo apt install dunst
```

## 📁 File Locations

```
Binary:          /usr/local/bin/htb-tool
Config:          ~/.config/htb-tool/config.json
VPN Downloads:   ~/Downloads/htb-vpn/*.ovpn
Desktop Entry:   ~/.local/share/applications/htb-tool.desktop
Autostart:       ~/.config/autostart/htb-tool.desktop
```

## 🔐 Security

- API token stored in `~/.config/htb-tool/config.json` (permissions: 0600)
- VPN configs saved with 0600 permissions (readable only by you)
- No caching of flags or sensitive data
- All API calls over HTTPS
- Token never logged or displayed

## 🚀 Autostart

HTB Tool is configured to start automatically on login.

**Check autostart status**:
```bash
ls ~/.config/autostart/htb-tool.desktop
```

**Disable autostart**:
```bash
rm ~/.config/autostart/htb-tool.desktop
```

**Re-enable autostart**:
```bash
cp ~/.local/share/applications/htb-tool.desktop ~/.config/autostart/
```

## 🎯 Keyboard Shortcuts

System tray apps don't typically support keyboard shortcuts, but you can:

**Quick Launch** (set custom keyboard shortcut):
```
Command: htb-tool
Shortcut: Super+H (or your choice)
```

**Restart App**:
```bash
killall htb-tool && htb-tool &
```

## 💡 Pro Tips

1. **Pin to favorites**: Some DEs allow pinning tray icons
2. **Low latency VPN**: Use 🟢 (low load) servers
3. **Download all VPNs**: Download once, use any time
4. **Quick machine info**: Click Info to get IP before spawning
5. **Menu persistence**: Menu stays open until you click away
6. **Recent machines**: Active list shows newest first
7. **Category browsing**: Challenges grouped for easy discovery

## 📊 Resource Usage

- **Binary size**: 8.9MB (lightweight!)
- **Memory**: ~15-20MB (minimal)
- **CPU**: <1% when idle
- **Network**: Only on menu refresh/actions

Much lighter than the full GUI version!

## 🛠️ Development

**Project structure**:
```
htb-tool/
├── cmd/main.go                 # Entry point
├── internal/
│   ├── api/client.go           # HTB API
│   ├── config/config.go        # Config management
│   └── tray/
│       ├── tray.go             # Tray app & menus
│       └── actions.go          # Action handlers
└── go.mod
```

**Rebuild after changes**:
```bash
go build -o htb-tool ./cmd
sudo install -m 755 htb-tool /usr/local/bin/htb-tool
killall htb-tool
htb-tool &
```

## 🎓 HTB API Token

Get your token:
1. Login to https://www.hackthebox.com
2. Go to Settings → Profile
3. Scroll to "Create App Token"
4. Click "Generate"
5. Copy token (500+ characters starting with `eyJ...`)

## 🗑️ Uninstall

```bash
# Stop running instance
killall htb-tool

# Remove binary
sudo rm /usr/local/bin/htb-tool

# Remove desktop files
rm ~/.local/share/applications/htb-tool.desktop
rm ~/.config/autostart/htb-tool.desktop

# Remove config (optional - contains API token)
rm -rf ~/.config/htb-tool

# Remove VPN files (optional)
rm -rf ~/Downloads/htb-vpn
```

## 🆚 Tray vs GUI Version

| Feature | Tray Version | GUI Version |
|---------|-------------|-------------|
| Size | 8.9MB | 31MB |
| Memory | ~20MB | ~100MB |
| UI Type | Dropdown menu | Full window |
| Search | Limited (menu) | Full search |
| Filters | Categories only | Advanced filters |
| Visibility | System tray icon | Window + tray |
| Speed | Instant access | Need to open |
| Best for | Quick actions | Detailed browsing |

**Use tray version when**:
- You want minimal resource usage
- You prefer quick dropdown access
- You don't need advanced search/filters
- You like panel-based tools (like Flameshot)

## 🎉 You're Ready!

Your HTB Tool is now in the system tray. Just click the icon and start hacking!

```bash
# If not running:
htb-tool &

# Check if running:
ps aux | grep htb-tool

# View logs:
# (App logs to stdout, run in terminal to see)
htb-tool
```

Happy Hacking! 🚀🎯
