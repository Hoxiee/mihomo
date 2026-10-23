package adapter

import (
	"strings"
	"testing"
)

func TestProxyDiversityFingerprintSeparatesTransportAndHidesSNI(t *testing.T) {
	base := map[string]any{
		"network":    "ws",
		"tls":        true,
		"servername": "secret.example",
	}
	first := proxyDiversityFingerprint("vless", base)
	second := proxyDiversityFingerprint("vless", map[string]any{
		"network":    "grpc",
		"tls":        true,
		"servername": "secret.example",
	})

	if first == second {
		t.Fatal("different transports produced one diversity fingerprint")
	}
	if strings.Contains(first, "secret.example") {
		t.Fatal("diversity fingerprint exposed the raw SNI")
	}
	if first != proxyDiversityFingerprint("vless", base) {
		t.Fatal("diversity fingerprint is not deterministic")
	}
}

func TestProxyDiversityFingerprintTracksRealityAndObfs(t *testing.T) {
	plain := proxyDiversityFingerprint("vless", map[string]any{})
	disguised := proxyDiversityFingerprint("vless", map[string]any{
		"reality-opts": map[string]any{"server-name": "example.com"},
		"obfs":         "salamander",
	})

	if plain == disguised {
		t.Fatal("reality and obfs did not affect diversity")
	}
}
