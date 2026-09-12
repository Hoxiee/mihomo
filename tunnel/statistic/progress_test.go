package statistic

import "testing"

func TestTCPTrackerFirstProgressOnly(t *testing.T) {
	previous := DefaultFirstProgressNotify
	t.Cleanup(func() { DefaultFirstProgressNotify = previous })

	calls := 0
	DefaultFirstProgressNotify = func(Tracker) { calls++ }
	tracker := &tcpTracker{}

	tracker.notifyProgress(0)
	tracker.notifyProgress(2)
	tracker.notifyProgress(3)

	if calls != 1 {
		t.Fatalf("progress notifications = %d, want 1", calls)
	}
}

func TestUDPTrackerFirstProgressOnly(t *testing.T) {
	previous := DefaultFirstProgressNotify
	t.Cleanup(func() { DefaultFirstProgressNotify = previous })

	calls := 0
	DefaultFirstProgressNotify = func(Tracker) { calls++ }
	tracker := &udpTracker{}

	tracker.notifyProgress(1)
	tracker.notifyProgress(1)

	if calls != 1 {
		t.Fatalf("progress notifications = %d, want 1", calls)
	}
}
