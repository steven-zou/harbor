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

// Package pypi provides an OCI artifact processor for Python packages (wheels
// and source distributions) stored in Harbor.  When a Python package is
// pushed through the PyPI adapter it is wrapped in an OCI manifest whose
// config media type is MediaTypeConfig.  This processor extracts the
// Python-specific metadata (name, version, summary, requires-python) from
// that config blob and populates the artifact's ExtraAttrs field.
package pypi

import (
	"context"
	"encoding/json"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	ps "github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// ArtifactTypePyPI is the artifact type string used in Harbor for Python packages.
const ArtifactTypePyPI = "PYPI"

// MediaTypeConfig is the OCI config media type for PyPI artifacts.
const MediaTypeConfig = "application/vnd.goharbor.artifact.pypi.config.v1+json"

// MediaTypeLayer is the OCI layer media type for PyPI package files.
const MediaTypeLayer = "application/vnd.goharbor.artifact.pypi.file.v1"

// AnnotationFilename is the OCI layer annotation key used to store the
// original filename of a Python package file within the OCI manifest.
const AnnotationFilename = "org.opencontainers.image.title"

func init() {
	p := &processor{}
	p.ManifestProcessor = base.NewManifestProcessor()
	if err := ps.Register(p, MediaTypeConfig); err != nil {
		log.Errorf("failed to register PyPI processor for media type %s: %v", MediaTypeConfig, err)
	}
}

// PyPIMetadata holds the Python package metadata stored in the OCI config blob.
// The fields correspond to the standard Python package metadata fields defined
// in PEP 241/314/566.
type PyPIMetadata struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	Summary        string `json:"summary,omitempty"`
	RequiresPython string `json:"requiresPython,omitempty"`
	PackageType    string `json:"packageType,omitempty"` // "bdist_wheel", "sdist", etc.
	Filename       string `json:"filename,omitempty"`
}

type processor struct {
	*base.ManifestProcessor
}

// AbstractMetadata extracts Python package metadata from the OCI manifest
// config blob and stores it in artifact.ExtraAttrs.
func (p *processor) AbstractMetadata(_ context.Context, art *artifact.Artifact, manifest []byte) error {
	mani := &v1.Manifest{}
	if err := json.Unmarshal(manifest, mani); err != nil {
		return err
	}
	if mani.Config.Size == 0 {
		return nil
	}

	_, blob, err := p.RegCli.PullBlob(art.RepositoryName, mani.Config.Digest.String())
	if err != nil {
		return err
	}
	defer blob.Close()

	meta := &PyPIMetadata{}
	if err := json.NewDecoder(blob).Decode(meta); err != nil {
		return err
	}

	if art.ExtraAttrs == nil {
		art.ExtraAttrs = map[string]any{}
	}
	art.ExtraAttrs["name"] = meta.Name
	art.ExtraAttrs["version"] = meta.Version
	if meta.Summary != "" {
		art.ExtraAttrs["summary"] = meta.Summary
	}
	if meta.RequiresPython != "" {
		art.ExtraAttrs["requiresPython"] = meta.RequiresPython
	}
	if meta.PackageType != "" {
		art.ExtraAttrs["packageType"] = meta.PackageType
	}
	if meta.Filename != "" {
		art.ExtraAttrs["filename"] = meta.Filename
	}
	return nil
}

// AbstractAddition is not yet supported for PyPI artifacts.
func (p *processor) AbstractAddition(_ context.Context, art *artifact.Artifact, addition string) (*ps.Addition, error) {
	return nil, errors.New(nil).WithCode(errors.BadRequestCode).
		WithMessagef("addition %s isn't supported for %s", addition, ArtifactTypePyPI)
}

// GetArtifactType returns the PyPI artifact type string.
func (p *processor) GetArtifactType(_ context.Context, _ *artifact.Artifact) string {
	return ArtifactTypePyPI
}

// ListAdditionTypes returns the supported addition types for PyPI artifacts.
func (p *processor) ListAdditionTypes(_ context.Context, _ *artifact.Artifact) []string {
	return nil
}
