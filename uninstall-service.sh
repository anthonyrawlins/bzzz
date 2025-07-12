#!/bin/bash

# Bzzz P2P Service Uninstallation Script
# Removes Bzzz systemd service

set -e

echo "🐝 Uninstalling Bzzz P2P Task Coordination Service..."

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo "❌ This script must be run as root or with sudo"
    exit 1
fi

# Define paths
SYSTEMD_DIR="/etc/systemd/system"
SERVICE_FILE="$SYSTEMD_DIR/bzzz.service"

# Check if service exists
if [ ! -f "$SERVICE_FILE" ]; then
    echo "⚠️ Bzzz service not found at $SERVICE_FILE"
    echo "   Service may not be installed"
    exit 0
fi

# Stop the service if running
if systemctl is-active --quiet bzzz.service; then
    systemctl stop bzzz.service
    echo "✅ Stopped Bzzz service"
fi

# Disable the service
if systemctl is-enabled --quiet bzzz.service; then
    systemctl disable bzzz.service
    echo "✅ Disabled Bzzz service auto-start"
fi

# Remove service file
rm -f "$SERVICE_FILE"
echo "✅ Removed service file"

# Reload systemd daemon
systemctl daemon-reload
echo "✅ Reloaded systemd daemon"

# Reset failed state if any
systemctl reset-failed bzzz.service 2>/dev/null || true

echo ""
echo "🎉 Bzzz P2P Task Coordination Service uninstalled successfully!"
echo ""
echo "Note: The Bzzz binary and project files remain in /home/tony/AI/projects/Bzzz"