package provider

import (
	"testing"
	"time"
)

func TestHCTickDecision(t *testing.T) {
	const interval = 5 * time.Minute
	cases := []struct {
		name       string
		suppressed bool
		screenOff  bool
		sinceTouch time.Duration
		lazy       bool
		want       hcTickAction
	}{
		{"routing engine as sole prover defers the auto check", true, false, time.Second, false, hcDefer},
		{"suppressed wins even with the screen on", true, false, time.Second, true, hcDefer},
		{"screen off defers even when recently touched", false, true, time.Second, false, hcDefer},
		{"screen off defers a lazy idle group", false, true, 10 * time.Minute, true, hcDefer},
		{"screen on, recently touched, runs", false, false, time.Second, false, hcRun},
		{"screen on, recently touched, runs even if lazy", false, false, time.Second, true, hcRun},
		{"screen on, idle and lazy, skips", false, false, 10 * time.Minute, true, hcSkip},
		{"screen on, idle and eager, runs", false, false, 10 * time.Minute, false, hcRun},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hcTickDecision(c.suppressed, c.screenOff, c.sinceTouch, interval, c.lazy); got != c.want {
				t.Fatalf("got %d, want %d", got, c.want)
			}
		})
	}
}
