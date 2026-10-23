package outbound

import "sync"

// ExternalFactory builds an adapter for a proxy type registered by the embedding application.
type ExternalFactory func(mapping map[string]any, basic BasicOption) (ProxyAdapter, error)

var (
	externalMu        sync.RWMutex
	externalFactories = map[string]ExternalFactory{}
)

// RegisterExternal adds a proxy type the built-in parser does not know. Built-in types always win.
func RegisterExternal(name string, factory ExternalFactory) {
	externalMu.Lock()
	defer externalMu.Unlock()
	externalFactories[name] = factory
}

func LookupExternal(name string) (ExternalFactory, bool) {
	externalMu.RLock()
	defer externalMu.RUnlock()
	factory, ok := externalFactories[name]
	return factory, ok
}
