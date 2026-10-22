package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func (p *Proxy) RouteFingerprint() string    { return p.routeFingerprint }
func (p *Proxy) EndpointFingerprint() string { return p.endpointFingerprint }
func (p *Proxy) RouteDependency() string     { return p.routeDependency }

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

// Providers rotate credentials (uuid, short-id, password) without moving the
// server, so the endpoint alone lets the routing engine keep a node's history
// across such a refresh.
func proxyEndpointFingerprint(mapping map[string]any) string {
	endpoint := map[string]any{
		"type":   mapping["type"],
		"server": mapping["server"],
		"port":   mapping["port"],
		"sni":    firstString(mapping, "servername", "sni", "peer"),
	}
	if host := firstString(mapping, "host"); host != "" {
		endpoint["host"] = host
	}
	for _, key := range []string{"ws-opts", "http-opts", "h2-opts", "grpc-opts", "xhttp-opts"} {
		if opts, ok := mapping[key].(map[string]any); ok {
			endpoint[key+".path"] = firstString(opts, "path", "grpc-service-name")
			if headers, ok := opts["headers"].(map[string]any); ok {
				endpoint[key+".host"] = firstString(headers, "Host", "host")
			}
		}
	}
	data, err := json.Marshal(endpoint)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return "endpoint-v1:" + hex.EncodeToString(sum[:])
}
