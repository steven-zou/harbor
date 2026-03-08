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

// Package adapter provides a framework for serving various kinds of software
// artifacts (e.g., Java/Maven, Python/PyPI) using Harbor's OCI artifact registry
// as the underlying storage backend. Each adapter translates the native protocol
// of a specific artifact ecosystem into OCI operations and registers the
// corresponding HTTP routes with the Harbor API server.
//
// Enabling an adapter adds its specific API endpoints to Harbor, allowing
// clients to push and pull artifacts using their native tooling while
// Harbor transparently stores them as OCI artifacts. Authentication and
// authorization reuse Harbor's existing OCI artifact security workflow.
package adapter

import (
	"net/http"
)

// Route describes a single HTTP route exposed by an artifact adapter.
type Route struct {
	// Method is the HTTP method (GET, PUT, POST, DELETE, HEAD, …).
	// An empty string means the route matches any method.
	Method string
	// Path is the URL path pattern (beego-style, e.g. "/service/maven/*").
	Path string
	// Handler is the http.Handler that processes requests matching this route.
	Handler http.Handler
}

// Adapter defines the interface that every artifact-ecosystem adapter must
// implement. An adapter bridges the gap between a native package-manager
// protocol and Harbor's OCI artifact storage.
type Adapter interface {
	// GetName returns the unique, human-readable name of the artifact type
	// served by this adapter (e.g., "MAVEN", "PYPI").
	GetName() string

	// IsEnabled reports whether this adapter is currently enabled. When
	// false the adapter's routes are not registered with the HTTP server.
	IsEnabled() bool

	// GetRoutes returns the list of HTTP routes that should be registered
	// when this adapter is enabled.
	GetRoutes() []Route
}
