//go:build !no_easytier

package adapter

import "testing"

func TestEasyTierPreservesDiversityFingerprint(t *testing.T) {
	mapping := map[string]any{
		"name":         "easytier-diversity-test",
		"type":         "easytier",
		"network-name": "test-network",
		"peers":        []string{"tcp://127.0.0.1:11010"},
	}
	proxy, err := ParseProxy(mapping)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := proxy.Close(); err != nil {
			t.Error(err)
		}
	})

	want := proxyDiversityFingerprint("easytier", mapping)
	if got := proxy.ProxyInfo().DiversityFingerprint; got != want {
		t.Fatalf("diversity fingerprint = %q, want %q", got, want)
	}
}
