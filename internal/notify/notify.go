// Package notify shows macOS notifications.
package notify

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/mxfca/mac-inventory/internal/runner"
)

var escaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

// Show displays a notification with the given title and cause.
func Show(r runner.Runner, title, cause string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, escaper.Replace(cause), escaper.Replace(title))
	_, err := r.Run("osascript", "-e", script)
	return err
}

// Cause drops the file path from a file error, which is too long for a notification.
func Cause(err error) string {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Err.Error()
	}
	return err.Error()
}
