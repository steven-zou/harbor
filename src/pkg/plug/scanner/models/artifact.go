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
	"reflect"

	"github.com/pkg/errors"
)

// Artifact to be scanned
type Artifact struct {
	NamespaceID int64  `json:"namespaceId"`
	Namespace   string `json:"namespace"`
	Repository  string `json:"repository"`
	Tag         string `json:"tag"`
	Digest      string `json:"digest"`
	Kind        string `json:"kind"`
}

// Validate artifact
func (art *Artifact) Validate() error {
	val := reflect.ValueOf(art).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		if field.Interface() == nil {
			return errors.Errorf("missing '%s' in artifact", fieldType.Name)
		}

		switch field.Type().Kind() {
		case reflect.String:
			str, ok := field.Interface().(string)
			if !ok || len(str) == 0 {
				return errors.Errorf("malformed string '%s' in artifact", fieldType.Name)
			}
		case reflect.Int64:
			num, ok := field.Interface().(int64)
			if !ok || num <= 0 {
				return errors.Errorf("malformed int64 '%s' in artifact", fieldType.Name)
			}
		}
	}

	return nil
}
