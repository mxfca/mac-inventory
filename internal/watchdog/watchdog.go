// Package watchdog runs an action when work takes longer than a deadline.
package watchdog

import "time"

// Start runs onTimeout once, after timeout, unless the returned stop is called first.
// stop is safe to call more than once.
func Start(timeout time.Duration, onTimeout func()) (stop func()) {
	timer := time.AfterFunc(timeout, onTimeout)
	return func() { timer.Stop() }
}
