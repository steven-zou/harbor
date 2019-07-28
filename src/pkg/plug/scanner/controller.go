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

package scanner

import (
	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"
)

// MixReports contains scan reports of scanners
// A simple definition now
// Key is the adapter name
type MixReports map[string]*models.ScanReport

// Controller defines operations for scan controlling
type Controller interface {
	// Scan the given artifact
	//
	//   Arguments:
	//     artifact *res.Artifact : artifact to be scanned
	//
	//   Returns:
	//     string : a UUID for tracking the reports
	//     error  : non nil error if any errors occurred
	Scan(artifact *models.Artifact) (string, error)

	// GetReport gets the reports for the given artifact identified by the digest
	//
	//   Arguments:
	//     digest string : digest of the artifact
	//
	//   Returns:
	//     MixReports : scan reports by adapter w/ endpoint indexed
	//     error  : non nil error if any errors occurred
	GetReport(digest string) (MixReports, error)
}

// basicController is default implementation of Controller
type basicController struct{}

// NewController news a basic controller
func NewController() Controller {
	return &basicController{}
}

// Scan artifact
func (bc *basicController) Scan(artifact *models.Artifact) (string, error) {
	return "", nil
}

// GetReport gets the mixed reports
func (bc *basicController) GetReport(digest string) (MixReports, error) {
	return nil, nil
}
