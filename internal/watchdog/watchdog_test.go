package watchdog_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/mxfca/mac-inventory/internal/watchdog"
)

const (
	deadline = 20 * time.Millisecond
	settle   = 100 * time.Millisecond
)

func counter() (*atomic.Int32, func()) {
	var n atomic.Int32
	return &n, func() { n.Add(1) }
}

func TestStartRunsTheActionWhenTheDeadlinePasses(t *testing.T) {
	fired := make(chan struct{}, 1)

	watchdog.Start(deadline, func() { fired <- struct{}{} })

	select {
	case <-fired:
	case <-time.After(5 * time.Second):
		t.Fatal("action did not run after the deadline")
	}
}

func TestStopBeforeTheDeadlineSkipsTheAction(t *testing.T) {
	calls, action := counter()

	stop := watchdog.Start(deadline, action)
	stop()
	time.Sleep(settle)

	if got := calls.Load(); got != 0 {
		t.Errorf("action ran %d times, want 0", got)
	}
}

func TestStopTwiceIsSafe(t *testing.T) {
	calls, action := counter()

	stop := watchdog.Start(deadline, action)
	stop()
	stop()
	time.Sleep(settle)

	if got := calls.Load(); got != 0 {
		t.Errorf("action ran %d times, want 0", got)
	}
}

func TestActionRunsAtMostOnceEvenIfStoppedAfterTheDeadline(t *testing.T) {
	calls, action := counter()

	stop := watchdog.Start(deadline, action)
	time.Sleep(settle)
	stop()
	time.Sleep(settle)

	if got := calls.Load(); got != 1 {
		t.Errorf("action ran %d times, want 1", got)
	}
}
