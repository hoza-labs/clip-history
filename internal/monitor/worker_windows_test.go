//go:build windows

package monitor

import (
	"testing"
	"time"
)

// TestStopHeartbeatWaitsForHeartbeatExit verifies shutdown cannot race ahead of an in-progress heartbeat.
func TestStopHeartbeatWaitsForHeartbeatExit(t *testing.T) {
	started := make(chan struct{})
	stopping := make(chan struct{})
	release := make(chan struct{})
	stop := startHeartbeat(func(done <-chan struct{}) {
		close(started)
		<-done
		close(stopping)
		<-release
	})
	<-started

	stopped := make(chan struct{})
	go func() {
		stop()
		close(stopped)
	}()
	<-stopping

	select {
	case <-stopped:
		t.Fatal("stopHeartbeat returned before the heartbeat exited")
	default:
	}

	close(release)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("stopHeartbeat did not return after the heartbeat exited")
	}
}
