package adapter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRouteFingerprintSeparatesAuthenticationAndPreservesAliases(t *testing.T) {
	a := map[string]any{"name": "first", "type": "vless", "server": "example.com", "port": 443, "uuid": "artificial-secret", "tls": true, "reality-opts": map[string]any{"public-key": "key"}}
	original, _ := json.Marshal(a)
	first := proxyRouteFingerprint(a)
	a["name"] = "alias"
	if got := proxyRouteFingerprint(a); got != first {
		t.Fatal("name changed route identity")
	}
	a["uuid"] = "other-secret"
	if got := proxyRouteFingerprint(a); got == first {
		t.Fatal("credentials pooled")
	}
	a["uuid"] = "artificial-secret"
	a["name"] = "first"
	after, _ := json.Marshal(a)
	if string(after) != string(original) {
		t.Fatal("fingerprinting changed input")
	}
	a["new-unknown-option"] = true
	if proxyRouteFingerprint(a) == first {
		t.Fatal("unknown connection option ignored")
	}
	a["invalid"] = make(chan int)
	if proxyRouteFingerprint(a) != "" {
		t.Fatal("unrepresentable route must stay unknown")
	}
}

func TestRouteMetadataStaysPrivate(t *testing.T) {
	p, err := ParseProxy(map[string]any{"type": "socks5", "name": "alias", "server": "example.com", "port": 1080, "username": "user", "password": "artificial-secret"})
	if err != nil {
		t.Fatal(err)
	}
	proxy := p.(*Proxy)
	if proxy.RouteFingerprint() == "" {
		t.Fatal("missing fingerprint")
	}
	data, err := json.Marshal(proxy)
	if err != nil {
		t.Fatal(err)
	}
	for _, private := range []string{proxy.RouteFingerprint(), "artificial-secret", "routeFingerprint"} {
		if strings.Contains(string(data), private) {
			t.Fatalf("private route data exposed: %s", private)
		}
	}
}

func TestEndpointFingerprintSurvivesCredentialRotation(t *testing.T) {
	a := map[string]any{"name": "first", "type": "vless", "server": "example.com", "port": 443, "uuid": "one", "servername": "cdn.example.com", "reality-opts": map[string]any{"short-id": "aa"}}
	first := proxyEndpointFingerprint(a)
	a["uuid"] = "two"
	a["reality-opts"] = map[string]any{"short-id": "bb"}
	a["name"] = "renamed"
	if proxyEndpointFingerprint(a) != first {
		t.Fatal("credential rotation changed the endpoint")
	}
	a["servername"] = "other.example.com"
	if proxyEndpointFingerprint(a) == first {
		t.Fatal("SNI change kept the endpoint")
	}
	a["servername"] = "cdn.example.com"
	a["port"] = 8443
	if proxyEndpointFingerprint(a) == first {
		t.Fatal("port change kept the endpoint")
	}
}
