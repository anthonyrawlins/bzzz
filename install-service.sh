#!/bin/bash

# Bzzz P2P Service Installation Script
# Installs Bzzz as a systemd service

set -e

echo "🐝 Installing Bzzz P2P Task Coordination Service..."

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo "❌ This script must be run as root or with sudo"
    exit 1
fi

# Define paths
BZZZ_DIR="/home/tony/AI/projects/Bzzz"
SERVICE_FILE="$BZZZ_DIR/bzzz.service"
SYSTEMD_DIR="/etc/systemd/system"

# Check if Bzzz binary exists
if [ ! -f "$BZZZ_DIR/bzzz" ]; then
    echo "❌ Bzzz binary not found at $BZZZ_DIR/bzzz"
    echo "   Please build the binary first with: go build -o bzzz"
    exit 1
fi

# Make binary executable
chmod +x "$BZZZ_DIR/bzzz"
echo "✅ Made Bzzz binary executable"

# Copy service file to systemd directory
cp "$SERVICE_FILE" "$SYSTEMD_DIR/bzzz.service"
echo "✅ Copied service file to $SYSTEMD_DIR/bzzz.service"

# Set proper permissions
chmod 644 "$SYSTEMD_DIR/bzzz.service"
echo "✅ Set service file permissions"

# Reload systemd daemon
systemctl daemon-reload
echo "✅ Reloaded systemd daemon"

# Enable service to start on boot
systemctl enable bzzz.service
echo "✅ Enabled Bzzz service for auto-start"

# Start the service
systemctl start bzzz.service
echo "✅ Started Bzzz service"

# Check service status
echo ""
echo "📊 Service Status:"
systemctl status bzzz.service --no-pager -l

echo ""
echo "🎉 Bzzz P2P Task Coordination Service installed successfully!"
echo ""
echo "Commands:"
echo "  sudo systemctl start bzzz     - Start the service"
echo "  sudo systemctl stop bzzz      - Stop the service"
echo "  sudo systemctl restart bzzz   - Restart the service"
echo "  sudo systemctl status bzzz    - Check service status"
echo "  sudo journalctl -u bzzz -f    - Follow service logs"
echo "  sudo systemctl disable bzzz   - Disable auto-start"