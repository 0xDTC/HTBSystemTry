# HTB Tool - Quick Start Guide

## Installation

```bash
cd /media/kali/3315E7784CD31C71/Scripts/htb-tool
./install.sh
```

## Get Your HTB API Token

### Method 1: Web Browser
1. Go to https://www.hackthebox.com/home/settings
2. Scroll down to "**Create App Token**"
3. Click "**Generate**"
4. Copy the token (starts with `eyJ...`)

### Method 2: Manual Config
If you already have a token, create the config file directly:

```bash
mkdir -p ~/.config/htb-tool

cat > ~/.config/htb-tool/config.json <<'EOF'
{
  "api_token": "YOUR_TOKEN_HERE",
  "vpn_directory": "/home/$USER/Downloads/htb-vpn",
  "last_protocol": "tcp",
  "window_width": 1200,
  "window_height": 800
}
EOF
```

Replace `YOUR_TOKEN_HERE` with your actual token.

## First Launch

```bash
htb-tool &
```

The app will:
1. Show a setup dialog if no token is found
2. Ask you to paste your API token
3. Save it securely to `~/.config/htb-tool/config.json`
4. Open the main window
5. Minimize to system tray (like Flameshot)

## Features Overview

### 🖥️ Machines Tab
- **Search**: Type machine name to filter
- **Filter**: Easy, Medium, Hard, Insane, Active, Retired, OS type
- **Spawn**: Click "Spawn" button to start a machine
- **Stop**: Click "Stop" to terminate running machine
- **Info**: View details and submit flags

### 🎯 Challenges Tab
- **Browse**: All HTB challenges
- **Filter**: By category (Web, Crypto, Pwn, etc.)
- **Submit**: Click challenge → Enter flag → Submit

### 🔐 VPN Tab
- **Select servers**: Choose one or multiple regions
- **Protocol**: TCP or UDP
- **Download**: Downloads .ovpn files to configured directory
- **Use**: `sudo openvpn ~/Downloads/htb-vpn/htb_vpn_123_tcp.ovpn`

### ⚙️ Settings Tab
- Update API token anytime
- Change VPN download directory
- View app info

## System Tray Usage

HTB Tool runs in the system tray (like Flameshot):

- **Left click** tray icon → Show/hide window
- **Right click** → Quick menu
  - Show
  - Refresh Machines
  - Quit

## Autostart

HTB Tool is configured to start automatically on login.

To disable:
```bash
rm ~/.config/autostart/htb-tool.desktop
```

To re-enable:
```bash
cp ~/.local/share/applications/htb-tool.desktop ~/.config/autostart/
```

## Common Workflows

### Spawn and Pwn a Machine
1. Go to **Machines** tab
2. Search for machine name
3. Click **Spawn**
4. Wait for machine to start (~2-3 minutes)
5. Click **Info** to see IP address
6. Hack the machine
7. Click **Info** → Enter flag → Submit User/Root Flag
8. Get instant feedback on correctness

### Download VPN for Multiple Regions
1. Go to **VPN** tab
2. Click **Select All** (or check specific servers)
3. Choose **TCP** or **UDP**
4. Set download directory (default: ~/Downloads/htb-vpn)
5. Click **Download Selected VPNs**
6. All configs downloaded at once

### Challenge Hunting
1. Go to **Challenges** tab
2. Filter by category (e.g., "Web")
3. Browse challenges
4. Click challenge → View info
5. Download/solve challenge
6. Return to app → Enter flag → Submit

## Troubleshooting

### "API token invalid"
- Go to Settings tab
- Update token from https://www.hackthebox.com/home/settings
- Click "Update API Token"

### "Failed to load machines"
- Check internet connection
- Verify HTB website is accessible
- Try refreshing (Ctrl+R or menu → Refresh)

### System tray icon not showing
- Make sure your DE supports system tray
- Try restarting the app
- Check if other tray apps work (nm-applet, flameshot)

### VPN download fails
- Verify you have active VIP/VIP+ subscription
- Check write permissions on download directory
- Try a different server/region

## Keyboard Shortcuts

- `Ctrl+R` - Refresh current view
- `Ctrl+F` - Focus search (in Machines/Challenges)
- `Ctrl+Q` - Quit application
- `Ctrl+,` - Open Settings

## Config File Location

```
~/.config/htb-tool/config.json
```

Contains:
- API token (encrypted storage recommended)
- VPN download directory
- Window size preferences
- Last used protocol

## Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/htb-tool

# Remove desktop entries
rm ~/.local/share/applications/htb-tool.desktop
rm ~/.config/autostart/htb-tool.desktop

# Remove config (optional - contains your API token)
rm -rf ~/.config/htb-tool

# Update desktop database
update-desktop-database ~/.local/share/applications/
```

## Tips & Tricks

1. **Pin to panel**: Right-click tray icon → "Pin to panel" for quick access
2. **Multiple VPN regions**: Download all regions, connect to closest/least loaded
3. **Flag history**: App doesn't store submitted flags (security)
4. **Quick spawn**: Use filter "Active" + "Easy" to find quick wins
5. **Search tips**: Search works on machine name AND OS
6. **System tray**: App stays running in background for quick access

## Security Notes

- API token stored in `~/.config/htb-tool/config.json` (file permissions: 0600)
- Token never displayed in UI after initial setup
- Use Settings → Update Token to change it
- No flags or sensitive data cached
- VPN configs downloaded with 0600 permissions

## Need Help?

- Check logs: Application runs in foreground, check terminal for errors
- GitHub Issues: [Create an issue](https://github.com/yourusername/htb-tool/issues)
- HTB Forums: Ask in HackTheBox forums
- Discord: HTB Discord server #tools channel

## Pro Features Coming Soon

- [ ] Auto-reset machines before expiry
- [ ] Desktop notifications for machine status
- [ ] Challenge file auto-download
- [ ] Progress tracking & statistics
- [ ] Multiple HTB account support
- [ ] VPN auto-connect integration

Enjoy hacking! 🎯
