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

// Package pypi provides the Harbor artifact adapter for Python (PyPI) packages.
// It implements the adapter.Adapter interface and registers itself in the global
// adapter registry on package initialization.  When enabled via configuration
// the PyPI adapter adds the PyPI simple repository HTTP routes to the Harbor
// API server so that pip/twine can push and pull packages using the PyPI
// Simple API while Harbor stores them as OCI artifacts.
package pypi

import (
	"context"
	"net/http"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact/adapter"
	pypiHandler "github.com/goharbor/harbor/src/server/v2.0/handler/pypi"
)

func init() {
	if err := adapter.Register(&pypiAdapter{}); err != nil {
		log.Errorf("failed to register PyPI artifact adapter: %v", err)
	}
}

// pypiAdapter implements the adapter.Adapter interface for PyPI artifacts.
type pypiAdapter struct{}

// GetName returns the name of this adapter.
func (p *pypiAdapter) GetName() string {
	return "PYPI"
}

// IsEnabled reports whether the PyPI adapter is enabled via Harbor configuration.
func (p *pypiAdapter) IsEnabled() bool {
	return config.EnablePyPIProxyArtifact(context.Background())
}

// GetRoutes returns the HTTP routes for the PyPI simple repository protocol.
// The routes are registered under the /service/pypi/ prefix.
func (p *pypiAdapter) GetRoutes() []adapter.Route {
	h := pypiHandler.New()
	return []adapter.Route{
		// Simple index: list all packages
		{
			Method:  http.MethodGet,
			Path:    "/service/pypi/*/simple/",
			Handler: h,
		},
		// Package index: list files for a specific package
		{
			Method:  http.MethodGet,
			Path:    "/service/pypi/*/simple/*",
			Handler: h,
		},
		// Upload a package
		{
			Method:  http.MethodPost,
			Path:    "/service/pypi/*/",
			Handler: h,
		},
		// Download a package file
		{
			Method:  http.MethodGet,
			Path:    "/service/pypi/*/packages/*",
			Handler: h,
		},
	}
}
