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

// Package maven provides the Harbor artifact adapter for Maven (Java) packages.
// It implements the adapter.Adapter interface and registers itself in the global
// adapter registry on package initialization.  When enabled via configuration
// the Maven adapter adds the Maven repository HTTP routes to the Harbor API
// server so that Maven tools can push and pull artifacts using their native
// protocol while Harbor stores them as OCI artifacts.
package maven

import (
	"context"
	"net/http"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact/adapter"
	mavenHandler "github.com/goharbor/harbor/src/server/v2.0/handler/maven"
)

func init() {
	if err := adapter.Register(&mavenAdapter{}); err != nil {
		log.Errorf("failed to register Maven artifact adapter: %v", err)
	}
}

// mavenAdapter implements the adapter.Adapter interface for Maven artifacts.
type mavenAdapter struct{}

// GetName returns the name of this adapter.
func (m *mavenAdapter) GetName() string {
	return "MAVEN"
}

// IsEnabled reports whether the Maven adapter is enabled via Harbor configuration.
func (m *mavenAdapter) IsEnabled() bool {
	return config.EnableMavenProxyArtifact(context.Background())
}

// GetRoutes returns the HTTP routes for the Maven repository protocol.
// The routes are registered under the /service/maven/ prefix.
func (m *mavenAdapter) GetRoutes() []adapter.Route {
	h := mavenHandler.New()
	return []adapter.Route{
		{
			Method:  http.MethodGet,
			Path:    "/service/maven/*",
			Handler: h,
		},
		{
			Method:  http.MethodHead,
			Path:    "/service/maven/*",
			Handler: h,
		},
		{
			Method:  http.MethodPut,
			Path:    "/service/maven/*",
			Handler: h,
		},
	}
}
