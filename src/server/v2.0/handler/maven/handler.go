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

// Package maven provides HTTP handlers that implement a subset of the Maven
// repository protocol on top of Harbor's OCI artifact storage.
//
// URL patterns:
//
//	GET  /service/maven/{project}/{group:.*}/{artifact}/{version}/{filename}
//	     – Download an artifact file.
//	PUT  /service/maven/{project}/{group:.*}/{artifact}/{version}/{filename}
//	     – Upload an artifact file (deploy).
//
// Maven artifacts are stored as OCI manifests:
//   - The OCI config blob contains JSON-encoded MavenMetadata.
//   - Each artifact file is a separate OCI layer with media type
//     maven.MediaTypeLayer and an annotation that records the filename.
//
// Authentication and authorization piggyback on Harbor's existing OCI
// security middleware stack (basic auth, OIDC, robot accounts, etc.).
// Project-level push/pull permissions are enforced before any registry
// operation.
package maven

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	mavenProcessor "github.com/goharbor/harbor/src/controller/artifact/processor/maven"
	"github.com/goharbor/harbor/src/controller/project"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// Handler is the Maven repository HTTP handler.
type Handler struct {
	regCli     registry.Client
	projectCtl project.Controller
}

// New returns a new Maven Handler using the global registry client.
func New() *Handler {
	return &Handler{
		regCli:     registry.Cli,
		projectCtl: project.Ctl,
	}
}

// ServeHTTP dispatches Maven repository requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Path pattern: /service/maven/{project}/{group...}/{artifact}/{version}/{filename}
	// We strip the leading prefix and parse the remaining segments.
	path := strings.TrimPrefix(r.URL.Path, "/service/maven/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "invalid Maven request path", http.StatusBadRequest)
		return
	}

	// parts[0] = project, parts[last] = filename, parts[last-1] = version,
	// parts[last-2] = artifactId, parts[1..last-3] = groupId segments
	projectName := parts[0]
	filename := parts[len(parts)-1]
	version := parts[len(parts)-2]
	artifactID := parts[len(parts)-3]
	groupParts := parts[1 : len(parts)-3]
	groupID := strings.Join(groupParts, ".")

	ctx := r.Context()

	// Derive the OCI repository name: {project}/{groupId}/{artifactId}
	repository := fmt.Sprintf("%s/%s/%s", projectName,
		strings.Join(groupParts, "/"), artifactID)

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		// Check pull permission
		if err := h.requireProjectAccess(ctx, w, projectName, rbac.ActionPull); err != nil {
			return
		}
		h.handleGet(w, r, repository, version, filename)

	case http.MethodPut:
		// Check push permission
		if err := h.requireProjectAccess(ctx, w, projectName, rbac.ActionPush); err != nil {
			return
		}
		h.handlePut(w, r, repository, groupID, artifactID, version, filename)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet downloads a Maven artifact file from the OCI registry.
func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request, repository, version, filename string) {
	ctx := r.Context()
	logger := log.G(ctx)

	// Pull manifest by tag (version)
	mani, _, err := h.regCli.PullManifest(repository, version)
	if err != nil {
		logger.Errorf("maven: failed to pull manifest for %s:%s: %v", repository, version, err)
		lib_http.SendError(w, err)
		return
	}

	_, payload, err := mani.Payload()
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	manifest := &v1.Manifest{}
	if err := json.Unmarshal(payload, manifest); err != nil {
		lib_http.SendError(w, err)
		return
	}

	// Find the layer whose annotation matches the requested filename.
	for _, layer := range manifest.Layers {
		layerFilename := layer.Annotations[mavenProcessor.AnnotationFilename]
		if layerFilename == filename {
			if r.Method == http.MethodHead {
				w.Header().Set("Content-Length", fmt.Sprintf("%d", layer.Size))
				w.WriteHeader(http.StatusOK)
				return
			}
			size, blob, err := h.regCli.PullBlob(repository, layer.Digest.String())
			if err != nil {
				lib_http.SendError(w, err)
				return
			}
			defer blob.Close()

			w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			if _, err := io.Copy(w, blob); err != nil {
				logger.Errorf("maven: error streaming blob: %v", err)
			}
			return
		}
	}

	http.Error(w, fmt.Sprintf("file %s not found in artifact %s:%s", filename, repository, version), http.StatusNotFound)
}

// handlePut uploads a Maven artifact file and stores it as an OCI artifact.
func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request, repository, groupID, artifactID, version, filename string) {
	ctx := r.Context()
	logger := log.G(ctx)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	// Push the file as a blob.
	layerDigest := digest.FromBytes(body)
	if err := h.regCli.PushBlob(repository, layerDigest.String(), int64(len(body)), bytes.NewReader(body)); err != nil {
		logger.Errorf("maven: failed to push blob for %s: %v", filename, err)
		lib_http.SendError(w, err)
		return
	}

	// Build the OCI config blob (MavenMetadata).
	packaging := inferPackaging(filename)
	meta := mavenProcessor.MavenMetadata{
		GroupID:    groupID,
		ArtifactID: artifactID,
		Version:    version,
		Packaging:  packaging,
	}
	configBytes, err := json.Marshal(meta)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}
	configDigest := digest.FromBytes(configBytes)
	if err := h.regCli.PushBlob(repository, configDigest.String(), int64(len(configBytes)), bytes.NewReader(configBytes)); err != nil {
		logger.Errorf("maven: failed to push config blob: %v", err)
		lib_http.SendError(w, err)
		return
	}

	// Assemble and push the OCI manifest.
	manifest := v1.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Config: v1.Descriptor{
			MediaType: mavenProcessor.MediaTypeConfig,
			Digest:    configDigest,
			Size:      int64(len(configBytes)),
		},
		Layers: []v1.Descriptor{
			{
				MediaType: mavenProcessor.MediaTypeLayer,
				Digest:    layerDigest,
				Size:      int64(len(body)),
				Annotations: map[string]string{
					mavenProcessor.AnnotationFilename: filename,
				},
			},
		},
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}
	if _, err := h.regCli.PushManifest(repository, version, v1.MediaTypeImageManifest, manifestBytes); err != nil {
		logger.Errorf("maven: failed to push manifest for %s:%s: %v", repository, version, err)
		lib_http.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// inferPackaging returns the packaging type based on the filename extension.
func inferPackaging(filename string) string {
	idx := strings.LastIndex(filename, ".")
	if idx == -1 {
		return "jar"
	}
	return filename[idx+1:]
}

// requireProjectAccess verifies that the requester has the specified action
// permission on the named Harbor project's artifact resource.  It writes the
// appropriate HTTP error response (401 or 403) and returns an error when
// access is denied; returns nil when access is allowed.
func (h *Handler) requireProjectAccess(ctx context.Context, w http.ResponseWriter, projectName string, action rbac.Action) error {
	sc, ok := security.FromContext(ctx)
	if !ok || !sc.IsAuthenticated() {
		w.Header().Set("WWW-Authenticate", `Basic realm="Harbor"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return fmt.Errorf("unauthenticated")
	}

	p, err := h.projectCtl.GetByName(ctx, projectName)
	if err != nil || p == nil {
		http.Error(w, fmt.Sprintf("project %s not found", projectName), http.StatusNotFound)
		return fmt.Errorf("project not found")
	}

	resource := rbac_project.NewNamespace(p.ProjectID).Resource(rbac.ResourceArtifact)
	if !sc.Can(ctx, action, resource) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return fmt.Errorf("forbidden")
	}
	return nil
}
