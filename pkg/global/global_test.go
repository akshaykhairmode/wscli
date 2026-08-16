package global

import (
	"testing"
	"time"
)

func TestStopAndWaitForStop(t *testing.T) {
	done := make(chan struct{})
	go func() {
		WaitForStop()
		close(done)
	}()

	Stop()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("WaitForStop() did not return after Stop() was called")
	}
}
