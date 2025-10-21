# HTB Tool

A powerful desktop application for managing your HackTheBox machines, challenges, and VPN connections.

## Features

### 🖥️ Machine Management
- **Browse all machines** (Active & Retired)
- **Search & filter** by difficulty, OS, status
- **Spawn/Terminate** machines with one click
- **Submit flags** directly from the app (User & Root)
- View detailed machine information (OS, difficulty, rating, owns, IP)
- Color-coded difficulty indicators
- Real-time status updates

### 🎯 Challenge Browser
- Browse all HTB challenges
- Filter by category (Web, Crypto, Pwn, Forensics, etc.)
- Search by name
- Submit flags directly
- View points, solves, and ratings
- Track completed challenges

### 🔐 VPN Management
- **List all VPN servers** with current load
- **Download multiple VPN configs** at once
- Choose between **TCP/UDP** protocols
- Select from any region/server
- Batch download with progress tracking

### ⚙️ Additional Features
- **System tray integration** - Minimize to tray like Flameshot
- **Quick access menu** from tray icon
- Persistent configuration
- Clean, modern UI with Fyne toolkit
- Cross-platform support (Linux, Windows, macOS)

## Installation

### Prerequisites
```bash
# Install Go (if not already installed)
# On Kali Linux:
sudo apt update
sudo apt install golang-go

# Install Fyne dependencies
sudo apt install libgl1-mesa-dev xorg-dev
```

### Build & Install
```bash
cd /media/kali/3315E7784CD31C71/Scripts/htb-tool

# Download dependencies
go mod tidy

# Build the application
go build -o htb-tool ./cmd

# Make it executable
chmod +x htb-tool

# Optional: Move to PATH
sudo mv htb-tool /usr/local/bin/
```

### Create Desktop Entry (for system tray icon)
```bash
cat > ~/.local/share/applications/htb-tool.desktop <<'EOF'
[Desktop Entry]
Name=HTB Tool
Comment=HackTheBox Management Tool
Exec=/usr/local/bin/htb-tool
Icon=network-vpn
Terminal=false
Type=Application
Categories=Network;Security;
StartupNotify=true
EOF

# Update desktop database
update-desktop-database ~/.local/share/applications/
```

## Configuration

### Getting Your HTB API Token
1. Go to https://www.hackthebox.com/home/settings
2. Scroll to "Create App Token"
3. Generate a new token
4. Copy and paste it into HTB Tool on first launch

### Configuration File
Location: `~/.config/htb-tool/config.json`

```json
{
  "api_token": "your_htb_api_token_here",
  "vpn_directory": "/home/user/Downloads/htb-vpn",
  "last_protocol": "tcp",
  "window_width": 1200,
  "window_height": 800
}
```

## Usage

### Launch the Application
```bash
htb-tool
```

### First Run Setup
1. Enter your HTB API token
2. Click "Save & Continue"
3. Start managing your HTB instances!

### Machine Management
1. Go to **Machines** tab
2. Use the search bar to find machines
3. Filter by difficulty, OS, or status
4. Click **Spawn** to start a machine
5. Click **Info** to see details and submit flags
6. Click **Stop** to terminate

### Submit Flags
1. Click **Info** on any spawned machine
2. Enter your flag in the text field
3. Click **Submit User Flag** or **Submit Root Flag**
4. Get instant feedback on correctness

### Download VPN Configs
1. Go to **VPN** tab
2. Select servers (use Select All for all regions)
3. Choose protocol (TCP/UDP)
4. Click **Download Selected VPNs**
5. Files saved to configured directory

### System Tray
- Click the tray icon to show/hide window
- Right-click for quick menu
- "Refresh Machines" to update list
- "Quit" to exit application

## Screenshots

```
┌─────────────────────────────────────────────────────┐
│ HTB Tool                                      [_][□][X]│
├─────────────────────────────────────────────────────┤
│ Machines | Challenges | VPN | Settings             │
├─────────────────────────────────────────────────────┤
│ Search: [_____________]  Filter: [All ▼]  [Refresh] │
├─────────────────────────────────────────────────────┤
│ 🟢 🟢 🐧 Keeper - Easy          [Info] [Spawn]      │
│ 🟢 🟡 🪟 Manager - Medium       [Info] [Stop]       │
│ 🔴 🟠 🐧 Neonify - Hard         [Info] [Spawn]      │
│ 🔴 🔴 🪟 Sightless - Insane     [Info] [Spawn]      │
├─────────────────────────────────────────────────────┤
│ Status: Loaded 500 machines                         │
└─────────────────────────────────────────────────────┘
```

## Keyboard Shortcuts

- `Ctrl+R` - Refresh current view
- `Ctrl+F` - Focus search bar
- `Ctrl+Q` - Quit application
- `Ctrl+,` - Open settings

## Troubleshooting

### "API token invalid" error
- Check your token at https://www.hackthebox.com/home/settings
- Make sure you copied the full token
- Generate a new token if needed

### "Failed to load machines" error
- Check your internet connection
- Verify HTB API is accessible
- Check if you have an active HTB subscription

### VPN download fails
- Ensure you have an active VIP/VIP+ subscription
- Check write permissions on download directory
- Verify server is online

### System tray icon not showing
- Make sure your desktop environment supports system tray
- Try restarting the application
- Check if other tray apps work (like Flameshot)

## Development

### Project Structure
```
htb-tool/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   └── client.go        # HTB API client
│   ├── config/
│   │   └── config.go        # Configuration management
│   └── ui/
│       ├── app.go           # Main application
│       ├── machines.go      # Machine browser
│       ├── challenges.go    # Challenge browser
│       ├── vpn.go           # VPN downloader
│       └── settings.go      # Settings page
├── go.mod
└── README.md
```

### Building from Source
```bash
# Clone and build
git clone <repo>
cd htb-tool
go mod tidy
go build -o htb-tool ./cmd
```

### Running in Development
```bash
go run ./cmd/main.go
```

## Dependencies

- [Fyne](https://fyne.io/) - Modern GUI toolkit
- HTB API v4

## License

MIT License

## Contributing

Pull requests are welcome! For major changes, please open an issue first.

## Roadmap

- [ ] Machine auto-reset before expiry
- [ ] Notification for new machines/challenges
- [ ] Challenge file downloads
- [ ] HTB Battlegrounds support
- [ ] Track time spent on machines
- [ ] Export statistics/progress

## Author

Created for HackTheBox enthusiasts who want a native desktop experience.

## Disclaimer

This tool is for educational purposes. Use responsibly and in accordance with HackTheBox's Terms of Service.
