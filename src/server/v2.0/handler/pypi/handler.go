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

// Package pypi provides HTTP handlers that implement a subset of the PyPI
// Simple Repository API (PEP 503) on top of Harbor's OCI artifact storage.
//
// URL patterns:
//
//	GET  /service/pypi/{project}/simple/
//	     – Returns an HTML index of all packages in the project.
//	GET  /service/pypi/{project}/simple/{package}/
//	     – Returns an HTML index of all files for the given package.
//	POST /service/pypi/{project}/
//	     – Upload a new Python distribution file (twine-compatible).
//	GET  /service/pypi/{project}/packages/{filename}
//	     – Download a distribution file by filename.
//
// Python packages are stored as OCI manifests:
//   - The OCI config blob contains JSON-encoded PyPIMetadata.
//   - The package file is stored as an OCI layer with media type
//     pypi.MediaTypeLayer and an annotation recording the filename.
//
// Authentication and authorization reuse Harbor's existing OCI security
// middleware stack. Project-level push/pull permissions are enforced.
package pypi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/common/rbac"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	pypiProcessor "github.com/goharbor/harbor/src/controller/artifact/processor/pypi"
	"github.com/goharbor/harbor/src/controller/project"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// Handler is the PyPI repository HTTP handler.
type Handler struct {
	regCli     registry.Client
	projectCtl project.Controller
}

// New returns a new PyPI Handler using the global registry client.
func New() *Handler {
	return &Handler{
		regCli:     registry.Cli,
		projectCtl: project.Ctl,
	}
}

// ServeHTTP dispatches PyPI repository requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Strip the common prefix /service/pypi/
	path := strings.TrimPrefix(r.URL.Path, "/service/pypi/")
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 3)
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "invalid PyPI request path", http.StatusBadRequest)
		return
	}

	projectName := parts[0]
	ctx := r.Context()

	if len(parts) == 1 {
		// POST /service/pypi/{project}/ – upload
		if r.Method == http.MethodPost {
			if err := h.requireProjectAccess(ctx, w, projectName, rbac.ActionPush); err != nil {
				return
			}
			h.handleUpload(w, r, projectName)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sub := parts[1]
	rest := ""
	if len(parts) == 3 {
		rest = parts[2]
	}

	switch sub {
	case "simple":
		if err := h.requireProjectAccess(ctx, w, projectName, rbac.ActionPull); err != nil {
			return
		}
		pkgName := strings.Trim(rest, "/")
		if pkgName == "" {
			h.handleSimpleIndex(w, r, projectName)
		} else {
			h.handlePackageIndex(w, r, projectName, normalizePackageName(pkgName))
		}

	case "packages":
		if err := h.requireProjectAccess(ctx, w, projectName, rbac.ActionPull); err != nil {
			return
		}
		h.handleDownload(w, r, projectName, rest)

	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

// handleSimpleIndex renders an HTML page listing all packages in a project.
func (h *Handler) handleSimpleIndex(w http.ResponseWriter, _ *http.Request, project string) {
	// List all repositories under the project that follow the pypi naming scheme.
	repos, err := h.regCli.Catalog()
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	prefix := project + "/pypi/"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!DOCTYPE html><html><head><title>Simple index</title></head><body>\n")
	for _, repo := range repos {
		if strings.HasPrefix(repo, prefix) {
			pkg := strings.TrimPrefix(repo, prefix)
			fmt.Fprintf(w, `<a href="%s/">%s</a><br/>`+"\n",
				html.EscapeString(pkg), html.EscapeString(pkg))
		}
	}
	fmt.Fprintf(w, "</body></html>\n")
}

// handlePackageIndex renders an HTML page with download links for all
// versions of a package, compatible with PEP 503 simple API.
func (h *Handler) handlePackageIndex(w http.ResponseWriter, _ *http.Request, project, pkgName string) {
	repository := fmt.Sprintf("%s/pypi/%s", project, pkgName)
	tags, err := h.regCli.ListTags(repository)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!DOCTYPE html><html><head><title>Links for %s</title></head><body>\n",
		html.EscapeString(pkgName))
	fmt.Fprintf(w, "<h1>Links for %s</h1>\n", html.EscapeString(pkgName))
	for _, tag := range tags {
		mani, _, err := h.regCli.PullManifest(repository, tag)
		if err != nil {
			continue
		}
		_, payload, err := mani.Payload()
		if err != nil {
			continue
		}
		manifest := &v1.Manifest{}
		if err := json.Unmarshal(payload, manifest); err != nil {
			continue
		}
		for _, layer := range manifest.Layers {
			fname := layer.Annotations[pypiProcessor.AnnotationFilename]
			if fname == "" {
				continue
			}
			// Download URL: /service/pypi/{project}/packages/{filename}
			dlURL := fmt.Sprintf("/service/pypi/%s/packages/%s", project, fname)
			dgst := layer.Digest.String()
			fmt.Fprintf(w, `<a href="%s#%s">%s</a><br/>`+"\n",
				html.EscapeString(dlURL),
				html.EscapeString(dgst),
				html.EscapeString(fname))
		}
	}
	fmt.Fprintf(w, "</body></html>\n")
}

// handleDownload streams a Python package file from the OCI registry.
// filename must include the package name and version to allow lookup, e.g.
// "requests-2.31.0-py3-none-any.whl".
func (h *Handler) handleDownload(w http.ResponseWriter, r *http.Request, project, filename string) {
	ctx := r.Context()
	logger := log.G(ctx)

	pkgName, version, err := parseFilename(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	repository := fmt.Sprintf("%s/pypi/%s", project, normalizePackageName(pkgName))

	mani, _, err := h.regCli.PullManifest(repository, version)
	if err != nil {
		logger.Errorf("pypi: failed to pull manifest %s:%s: %v", repository, version, err)
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

	for _, layer := range manifest.Layers {
		if layer.Annotations[pypiProcessor.AnnotationFilename] == filename {
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
				logger.Errorf("pypi: error streaming blob: %v", err)
			}
			return
		}
	}

	http.Error(w, fmt.Sprintf("file %s not found", filename), http.StatusNotFound)
}

// handleUpload accepts a multipart/form-data upload from pip/twine and stores
// the package as an OCI artifact in Harbor.
func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request, project string) {
	ctx := r.Context()
	logger := log.G(ctx)

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		http.Error(w, "expected multipart/form-data", http.StatusBadRequest)
		return
	}

	mr := multipart.NewReader(r.Body, params["boundary"])

	var (
		pkgName     string
		version     string
		summary     string
		reqPython   string
		pkgType     string
		filename    string
		fileContent []byte
	)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			lib_http.SendError(w, err)
			return
		}
		data, _ := io.ReadAll(part)
		switch part.FormName() {
		case "name":
			pkgName = string(data)
		case "version":
			version = string(data)
		case "summary":
			summary = string(data)
		case "requires_python":
			reqPython = string(data)
		case "filetype":
			pkgType = string(data)
		case "content":
			filename = part.FileName()
			fileContent = data
		}
	}

	if pkgName == "" || version == "" || filename == "" {
		http.Error(w, "missing required fields: name, version, content", http.StatusBadRequest)
		return
	}

	repository := fmt.Sprintf("%s/pypi/%s", project, normalizePackageName(pkgName))

	// Push the package file as a blob.
	fileDigest := digest.FromBytes(fileContent)
	if err := h.regCli.PushBlob(repository, fileDigest.String(), int64(len(fileContent)), bytes.NewReader(fileContent)); err != nil {
		logger.Errorf("pypi: failed to push blob for %s: %v", filename, err)
		lib_http.SendError(w, err)
		return
	}

	// Build and push the OCI config blob.
	meta := pypiProcessor.PyPIMetadata{
		Name:           pkgName,
		Version:        version,
		Summary:        summary,
		RequiresPython: reqPython,
		PackageType:    pkgType,
		Filename:       filename,
	}
	configBytes, err := json.Marshal(meta)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}
	configDigest := digest.FromBytes(configBytes)
	if err := h.regCli.PushBlob(repository, configDigest.String(), int64(len(configBytes)), bytes.NewReader(configBytes)); err != nil {
		logger.Errorf("pypi: failed to push config blob: %v", err)
		lib_http.SendError(w, err)
		return
	}

	// Assemble and push the OCI manifest.
	manifest := v1.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Config: v1.Descriptor{
			MediaType: pypiProcessor.MediaTypeConfig,
			Digest:    configDigest,
			Size:      int64(len(configBytes)),
		},
		Layers: []v1.Descriptor{
			{
				MediaType: pypiProcessor.MediaTypeLayer,
				Digest:    fileDigest,
				Size:      int64(len(fileContent)),
				Annotations: map[string]string{
					pypiProcessor.AnnotationFilename: filename,
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
		logger.Errorf("pypi: failed to push manifest for %s:%s: %v", repository, version, err)
		lib_http.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// normalizePackageName converts a PyPI package name to its canonical form
// (lowercased, hyphens replaced by underscores) as per PEP 503.
func normalizePackageName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "-", "_"))
}

// parseFilename attempts to derive the package name and version from a
// standard Python distribution filename.
// Wheel:  {name}-{version}(-{build})?-{python}-{abi}-{platform}.whl
// Sdist:  {name}-{version}.tar.gz  or  {name}-{version}.zip
// Returns an error when the filename cannot be parsed.
func parseFilename(filename string) (name, version string, err error) {
	// Try wheel first
	if strings.HasSuffix(filename, ".whl") {
		base := strings.TrimSuffix(filename, ".whl")
		parts := strings.SplitN(base, "-", 3)
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}
	// Try sdist
	for _, ext := range []string{".tar.gz", ".zip", ".tar.bz2"} {
		if strings.HasSuffix(filename, ext) {
			base := strings.TrimSuffix(filename, ext)
			idx := strings.LastIndex(base, "-")
			if idx > 0 {
				return base[:idx], base[idx+1:], nil
			}
		}
	}
	return "", "", fmt.Errorf("cannot parse name/version from filename: %s", filename)
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
