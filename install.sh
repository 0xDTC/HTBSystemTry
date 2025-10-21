#!/bin/bash

set -e

echo "========================================="
echo "  HTB Tool Installer"
echo "========================================="
echo ""

# Check if running as root
if [ "$EUID" -eq 0 ]; then
   echo "❌ Please do not run as root"
   exit 1
fi

# Check for Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    echo "Install Go with: sudo apt install golang-go"
    exit 1
fi

echo "✓ Go found: $(go version)"

# Check for required libraries
echo "Checking dependencies..."
MISSING_DEPS=()

for lib in libgl1-mesa-dev libxrandr-dev libxcursor-dev libxi-dev libxinerama-dev libxxf86vm-dev; do
    if ! dpkg -l | grep -q "^ii  $lib"; then
        MISSING_DEPS+=("$lib")
    fi
done

if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
    echo "❌ Missing dependencies: ${MISSING_DEPS[*]}"
    echo ""
    echo "Install them with:"
    echo "sudo apt install ${MISSING_DEPS[*]}"
    exit 1
fi

echo "✓ All dependencies installed"

# Build the application
echo ""
echo "Building HTB Tool..."
cd "$(dirname "$0")"

if ! go mod tidy; then
    echo "❌ Failed to download Go dependencies"
    exit 1
fi

if ! go build -o htb-tool ./cmd; then
    echo "❌ Build failed"
    exit 1
fi

echo "✓ Build successful"

# Install to system
echo ""
echo "Installing to /usr/local/bin..."
if ! sudo install -m 755 htb-tool /usr/local/bin/htb-tool; then
    echo "❌ Installation failed"
    exit 1
fi

echo "✓ Installed to /usr/local/bin/htb-tool"

# Create desktop entry
echo ""
echo "Creating desktop entry..."

mkdir -p ~/.local/share/applications
mkdir -p ~/.config/autostart

cat > ~/.local/share/applications/htb-tool.desktop <<'EOF'
[Desktop Entry]
Name=HTB Tool
Comment=HackTheBox Management Tool - Manage machines, challenges, and VPN
Exec=/usr/local/bin/htb-tool
Icon=security-high
Terminal=false
Type=Application
Categories=Network;Security;System;
StartupNotify=false
X-GNOME-Autostart-enabled=true
EOF

cp ~/.local/share/applications/htb-tool.desktop ~/.config/autostart/

update-desktop-database ~/.local/share/applications/ 2>/dev/null || true

echo "✓ Desktop entry created"
echo "✓ Autostart enabled"

# Create config directory
mkdir -p ~/.config/htb-tool

echo ""
echo "========================================="
echo "  Installation Complete! ✓"
echo "========================================="
echo ""
echo "HTB Tool has been installed successfully!"
echo ""
echo "📝 Next Steps:"
echo "  1. Get your HTB API token from:"
echo "     https://www.hackthebox.com/home/settings"
echo ""
echo "  2. Launch HTB Tool:"
echo "     • Run: htb-tool"
echo "     • Or find it in your applications menu"
echo "     • Or it will auto-start on next login"
echo ""
echo "  3. Enter your API token on first launch"
echo ""
echo "📍 Config location: ~/.config/htb-tool/config.json"
echo ""
echo "🎯 Features:"
echo "  • Browse & search machines"
echo "  • Spawn/terminate machines"
echo "  • Submit flags directly"
echo "  • Download VPN configs (multiple regions)"
echo "  • System tray integration (like Flameshot)"
echo ""
echo "To start now: htb-tool &"
echo ""
