#!/bin/sh
# Removes the agent, the binary and ~/.mac-inventory. The inventory and its
# backups in the destination folder are left alone.
set -eu

label="com.mxfca.mac-inventory"
domain="gui/$(id -u)"

if launchctl print "$domain/$label" >/dev/null 2>&1; then
	launchctl bootout "$domain/$label"
fi
rm -f "$HOME/Library/LaunchAgents/$label.plist" "$HOME/.local/bin/mac-inventory" "$HOME/.mac-inventory"
echo "Removed $label. The inventory and its backups were not touched."
