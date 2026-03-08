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

package maven

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/docker/distribution"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
)

// mavenManifest is a pre-built OCI manifest JSON for testing.
// The layer has MediaType=MediaTypeLayer, filename annotation "pom", and a known digest.
const mavenManifest = `{"schemaVersion":2,"config":{"mediaType":"application/vnd.goharbor.artifact.maven.config.v1+json","digest":"sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","size":2},"layers":[{"mediaType":"application/vnd.goharbor.artifact.maven.file.v1","digest":"sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824","size":5,"annotations":{"org.opencontainers.image.title":"pom"}}]}`

type processorTestSuite struct {
	suite.Suite
	processor *processor
	regCli    *registry.Client
}

func (p *processorTestSuite) SetupTest() {
	p.regCli = &registry.Client{}
	p.processor = &processor{}
	p.processor.ManifestProcessor = &base.ManifestProcessor{RegCli: p.regCli}
}

func (p *processorTestSuite) TestGetArtifactType() {
	p.Equal(ArtifactTypeMaven, p.processor.GetArtifactType(nil, nil))
}

func (p *processorTestSuite) TestListAdditionTypes() {
	additions := p.processor.ListAdditionTypes(nil, nil)
	p.Equal([]string{AdditionTypePOM}, additions)
}

func (p *processorTestSuite) TestAbstractMetadata() {
	meta := MavenMetadata{
		GroupID:    "com.example",
		ArtifactID: "my-lib",
		Version:    "1.0.0",
		Packaging:  "jar",
	}
	configBytes, _ := json.Marshal(meta)

	manifest := v1.Manifest{
		Config: v1.Descriptor{
			MediaType: MediaTypeConfig,
			Size:      int64(len(configBytes)),
		},
	}
	manifestBytes, _ := json.Marshal(manifest)

	p.regCli.On("PullBlob", mock.Anything, mock.Anything).
		Return(int64(len(configBytes)), io.NopCloser(strings.NewReader(string(configBytes))), nil)

	art := &artifact.Artifact{RepositoryName: "library/test"}
	err := p.processor.AbstractMetadata(nil, art, manifestBytes)
	p.Require().NoError(err)
	p.Equal("com.example", art.ExtraAttrs["groupId"])
	p.Equal("my-lib", art.ExtraAttrs["artifactId"])
	p.Equal("1.0.0", art.ExtraAttrs["version"])
	p.Equal("jar", art.ExtraAttrs["packaging"])
}

func (p *processorTestSuite) TestAbstractAdditionUnsupported() {
	art := &artifact.Artifact{}
	_, err := p.processor.AbstractAddition(nil, art, "UNSUPPORTED")
	p.True(errors.IsErr(err, errors.BadRequestCode))
}

func (p *processorTestSuite) TestAbstractAdditionPOM() {
	pomContent := `<project><modelVersion>4.0.0</modelVersion></project>`

	mani, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(mavenManifest))
	p.Require().NoError(err)

	p.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(mani, "", nil)
	p.regCli.On("PullBlob", mock.Anything, mock.Anything).
		Return(int64(len(pomContent)), io.NopCloser(strings.NewReader(pomContent)), nil)

	art := &artifact.Artifact{RepositoryName: "library/test", Digest: "sha256:xyz"}
	addition, err := p.processor.AbstractAddition(nil, art, AdditionTypePOM)
	p.Require().NoError(err)
	p.NotNil(addition)
	p.Equal("text/xml; charset=utf-8", addition.ContentType)
	p.Equal(pomContent, string(addition.Content))
}

func TestMavenProcessorTestSuite(t *testing.T) {
	suite.Run(t, &processorTestSuite{})
}
