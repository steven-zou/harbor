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

package models

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// ScanReport describes the scan report data
type ScanReport struct {
	Severity        Severity             `json:"severity"`
	Overview        *ComponentsOverview  `json:"overview"`
	Vulnerabilities []*VulnerabilityItem `json:"vulnerabilities"`
}

// Severity represents the severity of a image/component in terms of vulnerability.
type Severity int64

// ComponentsOverview has the total number and a list of components number of different serverity level.
type ComponentsOverview struct {
	Total   int                        `json:"total"`
	Summary []*ComponentsOverviewEntry `json:"summary"`
}

// ComponentsOverviewEntry ...
type ComponentsOverviewEntry struct {
	Sev   int `json:"severity"`
	Count int `json:"count"`
}

// VulnerabilityItem is an item in the vulnerability result returned by vulnerability details API.
type VulnerabilityItem struct {
	ID          string   `json:"id"`
	Severity    Severity `json:"severity"`
	Pkg         string   `json:"package"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Link        string   `json:"link"`
	Fixed       string   `json:"fixedVersion,omitempty"`
}

// FromJSON parses json data
func (r *ScanReport) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), r)
}

// ToJSON marshals endpoint to JSON data
func (r *ScanReport) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
