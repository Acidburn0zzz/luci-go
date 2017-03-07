// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package streamclient

import (
	"fmt"
	"strings"
	"sync"
)

// ClientFactory is a generator function that is invoked by the Registry when a
// new Client is requested for its protocol.
type ClientFactory func(string) (Client, error)

// Registry maps protocol prefix strings to their Client generator functions.
//
// This allows multiple Butler stream protocols (e.g., "unix:", "net.pipe:",
// etc.) to be parsed from string.
type Registry struct {
	// lock protects the fields in Registry.
	lock sync.Mutex

	// protocols is the set of registered protocols. Each client should register
	// via registerProtocol in its init() method.
	protocols map[string]ClientFactory
}

// Register registers a new protocol and its ClientFactory.
//
// This can be invoked by calling NewClient with a path spec referencing that
// protocol.
func (r *Registry) Register(name string, f ClientFactory) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.protocols[name]; ok {
		panic(fmt.Errorf("streamclient: protocol already registered for [%s]", name))
	}
	if r.protocols == nil {
		r.protocols = make(map[string]ClientFactory)
	}
	r.protocols[name] = f
}

// NewClient invokes the protocol ClientFactory generator for the
// supplied protocol/address string, returning the generated Client.
func (r *Registry) NewClient(path string) (Client, error) {
	parts := strings.SplitN(path, ":", 2)
	value := ""
	if len(parts) == 2 {
		value = parts[1]
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	if f, ok := r.protocols[parts[0]]; ok {
		return f(value)
	}
	return nil, fmt.Errorf("streamclient: no protocol registered for [%s]", parts[0])
}

// defaultRegistry is the default protocol registry.
var (
	defaultRegistry         *Registry
	defaultRegistryInitOnce sync.Once
)

// GetDefaultRegistry returns the default client registry instance.
//
// Initializes the registry on first invocation.
func GetDefaultRegistry() *Registry {
	defaultRegistryInitOnce.Do(func() {
		defaultRegistry = &Registry{}
		RegisterDefaultProtocols(defaultRegistry)
	})
	return defaultRegistry
}

// RegisterDefaultProtocols registers the default set of protocols with a
// Registry.
func RegisterDefaultProtocols(r *Registry) {
	registerPlatformProtocols(r)

	// Register common protocols.
	r.Register("tcp4", tcpProtocolClientFactory("tcp4"))
	r.Register("tcp6", tcpProtocolClientFactory("tcp6"))
}
