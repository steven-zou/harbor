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
	"reflect"

	"github.com/pkg/errors"
)

const (
	// Severity xxx is the list of severity of image after scanning.
	_ Severity = iota
	// SevNone = none
	SevNone
	// SevUnknown = unknown
	SevUnknown
	// SevLow = low
	SevLow
	// SevMedium = medium
	SevMedium
	// SevHigh = high
	SevHigh
)

// ScanRequest is request for launching a scan action
type ScanRequest struct {
	RegistryURL   string `json:"registry_url"`
	RegistryToken string `json:"registry_token"`
	Repository    string `json:"repository"`
	Tag           string `json:"tag"`
	Digest        string `json:"digest"`
}

// Validate scan request
func (sr *ScanRequest) Validate() error {
	val := reflect.ValueOf(sr).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		if fieldType.Name == "Tag" {
			// ignore
			continue
		}

		if field.Interface() == nil {
			return errors.Errorf("missing '%s' in scan request", fieldType.Name)
		}

		if field.Type().Kind() == reflect.String {
			str, ok := field.Interface().(string)
			if !ok || len(str) == 0 {
				return errors.Errorf("malformed '%s' in scan request", fieldType.Name)
			}
		}
	}

	return nil
}

// FromJSON parses json data
func (sr *ScanRequest) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), sr)
}

// ToJSON marshals scan request to JSON data
func (sr *ScanRequest) ToJSON() (string, error) {
	data, err := json.Marshal(sr)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ScanResponse contains the related response info
type ScanResponse struct {
	DetailsKey string `json:"details_key"`
}
