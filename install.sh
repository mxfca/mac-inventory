#!/bin/sh
# Installs (or reinstalls) mac-inventory and its weekly launchd agent.
set -eu

label="com.mxfca.mac-inventory"
config="$HOME/.mac-inventory"
bin="$HOME/.local/bin/mac-inventory"
plist="$HOME/Library/LaunchAgents/$label.plist"
domain="gui/$(id -u)"

# Fails and sets $reason unless $1 is an absolute path to an existing folder
# (the same rule as config.Write).
check_dir() {
	case "$1" in
	"") reason="the folder is empty" ;;
	/*)
		[ -d "$1" ] && return 0
		if [ -e "$1" ]; then reason="$1 is not a folder"; else reason="$1 does not exist"; fi
		;;
	*) reason="$1 is a relative path, not an absolute one" ;;
	esac
	return 1
}

# Keep the saved folder if it is valid; otherwise ask for it before doing anything else.
# shortcut: only the first line is read, so a file with a blank first line counts as empty, when that matters read the whole file.
dir=""
if [ -e "$config" ]; then read -r dir <"$config" || true; fi
if ! check_dir "$dir"; then
	[ ! -e "$config" ] || echo "install.sh: $config is invalid ($reason); asking for the folder again."
	printf 'Destination folder for the inventory (absolute path): '
	read -r dir || true
	if ! check_dir "$dir"; then
		echo "install.sh: $reason; nothing was installed." >&2
		exit 1
	fi
	(umask 077 && printf '%s\n' "$dir" >"$config")
	chmod 600 "$config" # umask does not tighten a file that already exists
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
