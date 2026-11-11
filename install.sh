#!/bin/sh
# Installs (or reinstalls) mac-inventory and its weekly launchd agent.
set -eu

label="com.mxfca.mac-inventory"
config="$HOME/.mac-inventory"
bin="$HOME/.local/bin/mac-inventory"
plist="$HOME/Library/LaunchAgents/$label.plist"
domain="gui/$(id -u)"

# First install: ask for the destination folder before doing anything else.
if [ ! -e "$config" ]; then
	printf 'Destination folder for the inventory (absolute path): '
	read -r dir
	case "$dir" in
	/*) ;;
	*)
		echo "install.sh: the folder must be an absolute path; nothing was installed." >&2
		exit 1
		;;
	esac
	if [ ! -d "$dir" ]; then
		echo "install.sh: $dir is not an existing folder; nothing was installed." >&2
		exit 1
	fi
	(umask 077 && printf '%s\n' "$dir" >"$config")
fi

cd "$(dirname "$0")"
mkdir -p "$(dirname "$bin")" "$(dirname "$plist")"
go build -o "$bin.new" ./cmd
mv "$bin.new" "$bin"

# Unload a previous agent so running this again updates it.
if launchctl print "$domain/$label" >/dev/null 2>&1; then
	launchctl bootout "$domain/$label"
fi

# No StandardOutPath/StandardErrorPath: nothing is logged to a file.
cat >"$plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>$label</string>
	<key>ProgramArguments</key>
	<array>
		<string>$bin</string>
	</array>
	<key>StartCalendarInterval</key>
	<dict>
		<key>Weekday</key>
		<integer>0</integer>
		<key>Hour</key>
		<integer>12</integer>
		<key>Minute</key>
		<integer>0</integer>
	</dict>
</dict>
</plist>
PLIST

launchctl bootstrap "$domain" "$plist"
echo "Installed $bin and loaded $label (Sundays at 12:00)."
