package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func (p *Proxy) RouteFingerprint() string { return p.routeFingerprint }
func (p *Proxy) RouteDependency() string  { return p.routeDependency }

func proxyRouteFingerprint(mapping map[string]any) string {
	route := make(map[string]any, len(mapping))
	for key, value := range mapping {
		if key != "name" {
			route[key] = value
		}
	}
	data, err := json.Marshal(route)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return "route-v1:" + hex.EncodeToString(sum[:])
}
