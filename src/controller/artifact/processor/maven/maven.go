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

// Package maven provides an OCI artifact processor for Maven (Java) packages
// stored in Harbor.  When a Maven artifact is pushed through the Maven adapter
// it is wrapped in an OCI manifest whose config media type is
// MediaTypeConfig.  This processor extracts the Maven-specific metadata
// (groupId, artifactId, version, packaging) from that config blob and
// populates the artifact's ExtraAttrs field so the information is
// accessible via Harbor's regular artifact APIs.
package maven

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	ps "github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// AnnotationFilename is the OCI layer annotation key used to store the
// original filename of a Maven artifact file within the OCI manifest.
const AnnotationFilename = "org.opencontainers.image.title"

// ArtifactTypeMaven is the artifact type string used in Harbor for Maven packages.
const ArtifactTypeMaven = "MAVEN"

// MediaTypeConfig is the OCI config media type for Maven artifacts.
const MediaTypeConfig = "application/vnd.goharbor.artifact.maven.config.v1+json"

// MediaTypeLayer is the OCI layer media type for Maven artifact files.
const MediaTypeLayer = "application/vnd.goharbor.artifact.maven.file.v1"

// AdditionTypePOM is the addition type for retrieving the POM file of a Maven artifact.
const AdditionTypePOM = "POM"

func init() {
	p := &processor{}
	p.ManifestProcessor = base.NewManifestProcessor()
	if err := ps.Register(p, MediaTypeConfig); err != nil {
		log.Errorf("failed to register Maven processor for media type %s: %v", MediaTypeConfig, err)
	}
}

// MavenMetadata holds the Maven-specific metadata stored in the OCI config blob.
type MavenMetadata struct {
	GroupID    string `json:"groupId"`
	ArtifactID string `json:"artifactId"`
	Version    string `json:"version"`
	Packaging  string `json:"packaging"`
}

type processor struct {
	*base.ManifestProcessor
}

// AbstractMetadata extracts Maven metadata from the OCI manifest config blob
// and stores it in artifact.ExtraAttrs.
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

	meta := &MavenMetadata{}
	if err := json.NewDecoder(blob).Decode(meta); err != nil {
		return err
	}

	if art.ExtraAttrs == nil {
		art.ExtraAttrs = map[string]any{}
	}
	art.ExtraAttrs["groupId"] = meta.GroupID
	art.ExtraAttrs["artifactId"] = meta.ArtifactID
	art.ExtraAttrs["version"] = meta.Version
	art.ExtraAttrs["packaging"] = meta.Packaging
	return nil
}

// AbstractAddition returns the POM file content for the Maven artifact.
func (p *processor) AbstractAddition(_ context.Context, art *artifact.Artifact, addition string) (*ps.Addition, error) {
	if addition != AdditionTypePOM {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("addition %s isn't supported for %s", addition, ArtifactTypeMaven)
	}

	mani, _, err := p.RegCli.PullManifest(art.RepositoryName, art.Digest)
	if err != nil {
		return nil, err
	}
	_, payload, err := mani.Payload()
	if err != nil {
		return nil, err
	}

	manifest := &v1.Manifest{}
	if err := json.Unmarshal(payload, manifest); err != nil {
		return nil, err
	}

	for _, layer := range manifest.Layers {
		fname := layer.Annotations[AnnotationFilename]
		// A POM layer is identified by a filename ending in ".pom" or exactly "pom".
		if fname != "" && (strings.HasSuffix(fname, ".pom") || fname == "pom") {
			_, blob, err := p.RegCli.PullBlob(art.RepositoryName, layer.Digest.String())
			if err != nil {
				return nil, err
			}
			defer blob.Close()

			content, err := io.ReadAll(blob)
			if err != nil {
				return nil, err
			}
			return &ps.Addition{
				Content:     content,
				ContentType: "text/xml; charset=utf-8",
			}, nil
		}
	}
	return nil, nil
}

// GetArtifactType returns the Maven artifact type string.
func (p *processor) GetArtifactType(_ context.Context, _ *artifact.Artifact) string {
	return ArtifactTypeMaven
}

// ListAdditionTypes returns the supported addition types for Maven artifacts.
func (p *processor) ListAdditionTypes(_ context.Context, _ *artifact.Artifact) []string {
	return []string{AdditionTypePOM}
}
