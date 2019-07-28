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

package v1

import (
	"fmt"
)

const (
	apiPrefix = "/api/v1"
)

// Spec for V1 REST API
type Spec struct {
	root string
}

// New v1 spec
func New(base string) *Spec {
	root := fmt.Sprintf("%s%s", "http://localhost", apiPrefix)
	if len(base) > 0 {
		root = fmt.Sprintf("%s%s", base, apiPrefix)
	}

	return &Spec{
		root: root,
	}
}

// Health is the URI of checking health
func (s *Spec) Health() string {
	return s.root
}

// Metadata is the URI of getting metadata
func (s *Spec) Metadata() string {
	return fmt.Sprintf("%s/%s", s.root, "metadata")
}

// Scan is the URI of doing scan
func (s *Spec) Scan() string {
	return fmt.Sprintf("%s/%s", s.root, "scan")
}

// Report is the URI of getting report for the given param
func (s *Spec) Report(param string) string {
	return fmt.Sprintf("%s/%s/%s", s.root, "scan", param)
}
