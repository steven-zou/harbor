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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAdapter is a test-only Adapter implementation.
type fakeAdapter struct {
	name    string
	enabled bool
}

func (f *fakeAdapter) GetName() string       { return f.name }
func (f *fakeAdapter) IsEnabled() bool       { return f.enabled }
func (f *fakeAdapter) GetRoutes() []Route    { return nil }

func TestRegisterAndGet(t *testing.T) {
	// Use a fresh isolated registry for each test by creating a local copy.
	// Reset the global registry state for testing.
	mu.Lock()
	savedRegistry := registry
	registry = map[string]Adapter{}
	mu.Unlock()
	defer func() {
		mu.Lock()
		registry = savedRegistry
		mu.Unlock()
	}()

	a := &fakeAdapter{name: "TEST", enabled: true}
	require.NoError(t, Register(a))

	got := Get("TEST")
	assert.Equal(t, a, got)

	assert.Nil(t, Get("NONEXISTENT"))
}

func TestRegisterDuplicate(t *testing.T) {
	mu.Lock()
	savedRegistry := registry
	registry = map[string]Adapter{}
	mu.Unlock()
	defer func() {
		mu.Lock()
		registry = savedRegistry
		mu.Unlock()
	}()

	a := &fakeAdapter{name: "DUP"}
	require.NoError(t, Register(a))

	err := Register(&fakeAdapter{name: "DUP"})
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	mu.Lock()
	savedRegistry := registry
	registry = map[string]Adapter{}
	mu.Unlock()
	defer func() {
		mu.Lock()
		registry = savedRegistry
		mu.Unlock()
	}()

	require.NoError(t, Register(&fakeAdapter{name: "A"}))
	require.NoError(t, Register(&fakeAdapter{name: "B"}))

	adapters := List()
	assert.Len(t, adapters, 2)
}

func TestRoute(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	route := Route{
		Method:  http.MethodGet,
		Path:    "/service/test/*",
		Handler: h,
	}
	assert.Equal(t, http.MethodGet, route.Method)
	assert.Equal(t, "/service/test/*", route.Path)
	assert.NotNil(t, route.Handler)
}
