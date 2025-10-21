# HTB Tool - Installation Complete! ✅

## Installation Summary

HTB Tool has been successfully created and installed on your system!

### What Was Installed

```
✓ Binary: /usr/local/bin/htb-tool (31MB)
✓ Desktop Entry: ~/.local/share/applications/htb-tool.desktop
✓ Autostart: ~/.config/autostart/htb-tool.desktop
✓ Config Directory: ~/.config/htb-tool/
```

### Project Location

```
/media/kali/3315E7784CD31C71/Scripts/htb-tool/
├── cmd/main.go                    # Entry point
├── internal/
│   ├── api/client.go              # HTB API client
│   ├── config/config.go           # Configuration management
│   └── ui/
│       ├── app.go                 # Main application & system tray
│       ├── machines.go            # Machine browser
│       ├── challenges.go          # Challenge browser
│       ├── vpn.go                 # VPN downloader
│       └── settings.go            # Settings page
├── htb-tool                       # Compiled binary
├── install.sh                     # Installation script
├── README.md                      # Full documentation
├── QUICKSTART.md                  # Quick start guide
└── go.mod                         # Go dependencies
```

## 🚀 Quick Start

### Step 1: Get Your HTB API Token

Visit: https://www.hackthebox.com/home/settings

1. Scroll to "Create App Token"
2. Click "Generate"
3. Copy the token (starts with `eyJ...`)

### Step 2: Launch HTB Tool

**Option A: From terminal**
```bash
htb-tool &
```

**Option B: From applications menu**
- Search for "HTB Tool" in your app launcher
- Click to launch

**Option C: It will auto-start on next login**
- Already configured in ~/.config/autostart/

### Step 3: First Run Setup

1. HTB Tool will show a setup dialog
2. Paste your API token
3. Click "Save & Continue"
4. Application opens with all features ready!

The app will minimize to system tray (look for the icon near Flameshot).

## ✨ Features

### 🖥️ Machine Management
- **Browse** all HTB machines (500+)
- **Search** by name or OS
- **Filter** by difficulty (Easy/Medium/Hard/Insane), status (Active/Retired), OS
- **Spawn** machines with one click
- **Terminate** running machines
- **Submit flags** directly (User & Root flags)
- View machine details (IP, rating, owns, release date)
- Color-coded difficulty indicators (🟢🟡🟠🔴)
- OS icons (🐧 Linux, 🪟 Windows, 😈 FreeBSD)

### 🎯 Challenge Browser
- Browse all HTB challenges
- Filter by category:
  - Web, Crypto, Pwn, Reversing
  - Forensics, OSINT, Mobile, Hardware, Misc
- Search by name
- Submit flags directly
- View points, solves, difficulty
- Track solved status ✅

### 🔐 VPN Management
- List all VPN servers with real-time load
- **Select multiple servers** to download at once
- Choose protocol: **TCP** or **UDP**
- Download .ovpn configs for:
  - US servers (multiple regions)
  - EU servers (multiple regions)
  - AU servers
  - SG servers
  - All available regions
- Batch download with progress tracking
- Configurable download directory

### 🎛️ System Tray Integration (Like Flameshot!)
- **Minimizes to system tray** instead of closing
- **Click icon** to show/hide window
- **Right-click menu** for quick actions:
  - Show window
  - Refresh machines
  - Quit application
- **Autostart** on login (like Flameshot)
- Runs as a background service

### ⚙️ Settings
- Update API token anytime
- Change VPN download directory
- Persistent configuration
- About/version info

## 📖 Usage Examples

### Example 1: Quick Machine Pwn
```
1. Launch htb-tool
2. Go to "Machines" tab
3. Filter: "Easy"
4. Search: "keeper"
5. Click "Spawn"
6. Wait ~2 minutes
7. Click "Info" to see IP
8. SSH/scan/pwn the machine
9. Click "Info" → Enter flag → "Submit User Flag"
10. Get instant feedback! 🎉
```

### Example 2: Download VPNs for All Regions
```
1. Go to "VPN" tab
2. Click "Select All"
3. Choose "TCP"
4. Set directory: ~/Downloads/htb-vpn
5. Click "Download Selected VPNs"
6. Progress shows download status
7. All configs saved to directory
```

### Example 3: Hunt Challenges
```
1. Go to "Challenges" tab
2. Filter: "Web"
3. Browse web challenges
4. Click challenge to see details
5. Solve it offline
6. Return to app
7. Enter flag → Submit
8. Instant correctness check ✓
```

## 🎯 System Tray Behavior

HTB Tool behaves like **Flameshot**:

### On Launch:
- Window opens
- Icon appears in system tray
- Can minimize to tray

### On Close (X button):
- Window hides (doesn't quit)
- Continues running in tray
- Click tray icon to restore

### To Quit:
- Right-click tray icon → "Quit"
- OR Ctrl+Q in window
- OR `killall htb-tool`

### Autostart:
- Starts automatically on login
- Minimizes to tray
- Ready when you need it

## 🔧 Configuration

### Config File Location
```
~/.config/htb-tool/config.json
```

### Manual Config (Optional)
If you want to pre-configure before first launch:

```bash
mkdir -p ~/.config/htb-tool

cat > ~/.config/htb-tool/config.json <<'EOF'
{
  "api_token": "YOUR_HTB_API_TOKEN_HERE",
  "vpn_directory": "/home/kali/Downloads/htb-vpn",
  "last_protocol": "tcp",
  "window_width": 1200,
  "window_height": 800
}
EOF

chmod 600 ~/.config/htb-tool/config.json
```

Replace `YOUR_HTB_API_TOKEN_HERE` with your actual token from HTB settings.

## 🛠️ Troubleshooting

### Issue: "API token invalid"
**Solution**: Go to Settings tab → Update API Token

### Issue: "Failed to load machines"
**Solutions**:
- Check internet connection
- Verify https://hackthebox.com is accessible
- Click "Refresh" button
- Restart app

### Issue: System tray icon not showing
**Solutions**:
- Check if your DE supports system tray
- Try: `htb-tool &` and look near Flameshot icon
- Restart desktop environment
- Check if other tray apps work

### Issue: VPN download fails
**Solutions**:
- Verify you have active VIP/VIP+ subscription
- Check download directory permissions
- Try a different server
- Check HTB website status

### Issue: App won't start
**Check**:
```bash
# Run in terminal to see errors
htb-tool

# Or check if already running
ps aux | grep htb-tool
```

## 📁 File Permissions

```
Binary:     /usr/local/bin/htb-tool         (755)
Config:     ~/.config/htb-tool/config.json  (600)
VPN files:  ~/Downloads/htb-vpn/*.ovpn      (600)
```

## 🔐 Security Notes

- API token stored in `~/.config/htb-tool/config.json`
- Config file has 0600 permissions (only you can read)
- Token is never displayed after initial setup
- No caching of flags or sensitive data
- VPN configs downloaded with secure permissions
- All API calls over HTTPS

## 📊 Resource Usage

- **Binary size**: 31MB
- **Memory usage**: ~50-100MB (similar to Flameshot)
- **CPU usage**: Minimal when idle in tray
- **Network**: Only when refreshing data or downloading VPNs

## 🎨 UI Features

- **Modern design** with Fyne toolkit
- **Emoji indicators** for quick visual scanning
- **Color-coded** difficulty levels
- **Responsive** search and filters
- **Real-time** status updates
- **System tray** integration
- **Persistent** window size/position

## 🚦 Next Steps

1. **Get your API token** from HTB settings
2. **Launch** htb-tool
3. **Enter token** on first run
4. **Start hacking!** 🎯

### Pro Tips:
- Pin tray icon for quick access
- Download VPNs for multiple regions (connect to least loaded)
- Use search + filters to find target machines quickly
- Submit flags immediately for instant feedback
- App stays in tray - no need to close/reopen

## 📚 Documentation

- **Full README**: /media/kali/3315E7784CD31C71/Scripts/htb-tool/README.md
- **Quick Start**: /media/kali/3315E7784CD31C71/Scripts/htb-tool/QUICKSTART.md
- **This File**: /media/kali/3315E7784CD31C71/Scripts/htb-tool/INSTALLATION_COMPLETE.md

## 🎓 HTB API Token Guide

### Where to Find It:
https://www.hackthebox.com/home/settings

### How to Generate:
1. Log into HackTheBox
2. Click profile → Settings
3. Scroll to "App Token" section
4. Click "Create App Token"
5. Copy the token (it's long!)
6. Paste into HTB Tool setup dialog

### Token Format:
```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```
(Usually 500+ characters)

## 🔄 Updating

To rebuild after code changes:
```bash
cd /media/kali/3315E7784CD31C71/Scripts/htb-tool
go build -o htb-tool ./cmd
sudo install -m 755 htb-tool /usr/local/bin/htb-tool
killall htb-tool
htb-tool &
```

## 🗑️ Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/htb-tool

# Remove desktop entries
rm ~/.local/share/applications/htb-tool.desktop
rm ~/.config/autostart/htb-tool.desktop

# Remove config (contains API token)
rm -rf ~/.config/htb-tool

# Update desktop database
update-desktop-database ~/.local/share/applications/
```

## 🎉 You're All Set!

HTB Tool is ready to use. Launch it and start managing your HackTheBox machines with ease!

```bash
htb-tool &
```

Happy Hacking! 🚀🎯
