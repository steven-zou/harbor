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
	"time"

	"github.com/pkg/errors"
)

// Endpoint of the scanner adapter service
type Endpoint struct {
	UUID             string    `orm:"pk;column(uid)" json:"uid"`
	URL              string    `orm:"column(url);unique;size(1024)" json:"url"`
	Auth             string    `orm:"column(auth)" json:"auth"`
	AccessCredential string    `orm:"column(access_cred);null" json:"accessCredential,omitempty"`
	Adapter          string    `orm:"column(adapter)" json:"adapter"`
	Disabled         bool      `orm:"column(disabled);default(true)" json:"disabled"`
	IsDefault        bool      `orm:"column(is_default);default(false)" json:"isDefault"`
	CreateTime       time.Time `orm:"column(create_time);auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime       time.Time `orm:"column(update_time);auto_now;type(datetime)" json:"updateTime"`
}

// TableName for Endpoint
func (e *Endpoint) TableName() string {
	return "scanner_endpoint"
}

// FromJSON parses json data
func (e *Endpoint) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), e)
}

// ToJSON marshals endpoint to JSON data
func (e *Endpoint) ToJSON() (string, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Validate endpoint
func (e *Endpoint) Validate() error {
	if len(e.UUID) == 0 || len(e.URL) == 0 {
		return errors.New("malformed endpoint")
	}

	if len(e.Adapter) == 0 {
		return errors.Errorf("missing adapter in endpoint %s:%s", e.UUID, e.URL)
	}

	return nil
}
