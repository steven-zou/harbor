// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adapter

import (
	"fmt"
	"sync"

	"github.com/goharbor/harbor/src/lib/log"
)

var (
	mu       sync.RWMutex
	registry = map[string]Adapter{}
)

// Register adds an adapter to the global registry. It returns an error if an
// adapter with the same name has already been registered.
func Register(a Adapter) error {
	mu.Lock()
	defer mu.Unlock()

	name := a.GetName()
	if _, exists := registry[name]; exists {
		return fmt.Errorf("artifact adapter %q is already registered", name)
	}
	registry[name] = a
	log.Infof("artifact adapter %q registered", name)
	return nil
}

// Get returns the adapter registered under name, or nil if none is found.
func Get(name string) Adapter {
	mu.RLock()
	defer mu.RUnlock()
	return registry[name]
}

// List returns all registered adapters.
func List() []Adapter {
	mu.RLock()
	defer mu.RUnlock()

	adapters := make([]Adapter, 0, len(registry))
	for _, a := range registry {
		adapters = append(adapters, a)
	}
	return adapters
}
