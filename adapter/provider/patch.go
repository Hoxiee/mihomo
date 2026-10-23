package provider

import "sync"

func (pp *proxySetProvider) GetSubscriptionInfo() *SubscriptionInfo {
	return pp.subscriptionInfo
}

var (
	screenMu     sync.Mutex
	screenIsOff  bool
	hcSuppressed map[string]struct{}
	screenWakeCh = make(chan struct{})
)

// Periodic probes are the tunnel's main idle-time radio wake-up, and a group
// nothing routed through while the screen was off has no reason to be probed.
// The tick is deferred rather than dropped: it runs when the screen comes back,
// so the delays a user sees are never the ones measured in their pocket.
func SetScreenOff(off bool) {
	screenMu.Lock()
	defer screenMu.Unlock()
	if screenIsOff == off {
		return
	}
	screenIsOff = off
	if off {
		return
	}
	close(screenWakeCh)
	screenWakeCh = make(chan struct{})
}

func screenOff() bool {
	screenMu.Lock()
	defer screenMu.Unlock()
	return screenIsOff
}

// SetAutoHealthCheckSuppressed pauses the auto health check of the named
// providers while the routing engine is their single prover, so their nodes are
// not probed twice; other providers keep checking and the engine harvests their
// results. A provider leaving the set flushes its deferred tick at once.
func SetAutoHealthCheckSuppressed(providers []string) {
	next := make(map[string]struct{}, len(providers))
	for _, name := range providers {
		next[name] = struct{}{}
	}
	screenMu.Lock()
	defer screenMu.Unlock()
	released := false
	for name := range hcSuppressed {
		if _, kept := next[name]; !kept {
			released = true
		}
	}
	hcSuppressed = next
	if !released {
		return
	}
	close(screenWakeCh)
	screenWakeCh = make(chan struct{})
}

func autoHealthCheckSuppressed(provider string) bool {
	screenMu.Lock()
	defer screenMu.Unlock()
	_, suppressed := hcSuppressed[provider]
	return suppressed
}

func screenWake() <-chan struct{} {
	screenMu.Lock()
	defer screenMu.Unlock()
	return screenWakeCh
}
