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

package pypi

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
)

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
	p.Equal(ArtifactTypePyPI, p.processor.GetArtifactType(nil, nil))
}

func (p *processorTestSuite) TestListAdditionTypes() {
	additions := p.processor.ListAdditionTypes(nil, nil)
	p.Nil(additions)
}

func (p *processorTestSuite) TestAbstractMetadata() {
	meta := PyPIMetadata{
		Name:           "requests",
		Version:        "2.31.0",
		Summary:        "HTTP library",
		RequiresPython: ">=3.7",
		PackageType:    "bdist_wheel",
		Filename:       "requests-2.31.0-py3-none-any.whl",
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
	p.Equal("requests", art.ExtraAttrs["name"])
	p.Equal("2.31.0", art.ExtraAttrs["version"])
	p.Equal("HTTP library", art.ExtraAttrs["summary"])
	p.Equal(">=3.7", art.ExtraAttrs["requiresPython"])
	p.Equal("bdist_wheel", art.ExtraAttrs["packageType"])
	p.Equal("requests-2.31.0-py3-none-any.whl", art.ExtraAttrs["filename"])
}

func (p *processorTestSuite) TestAbstractMetadataEmptyConfig() {
	manifest := v1.Manifest{
		Config: v1.Descriptor{
			MediaType: MediaTypeConfig,
			Size:      0,
		},
	}
	manifestBytes, _ := json.Marshal(manifest)

	art := &artifact.Artifact{RepositoryName: "library/test"}
	err := p.processor.AbstractMetadata(nil, art, manifestBytes)
	p.Require().NoError(err)
	p.Nil(art.ExtraAttrs)
}

func (p *processorTestSuite) TestAbstractAdditionUnsupported() {
	art := &artifact.Artifact{}
	_, err := p.processor.AbstractAddition(nil, art, "METADATA")
	p.True(errors.IsErr(err, errors.BadRequestCode))
}

func TestPyPIProcessorTestSuite(t *testing.T) {
	suite.Run(t, &processorTestSuite{})
}
