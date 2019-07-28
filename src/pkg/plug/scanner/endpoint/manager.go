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

package endpoint

import (
	"github.com/goharbor/harbor/src/pkg/plug/scanner/dao"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/q"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// DefaultManager ...
var DefaultManager = New()

// Manager manages the related scanner API endpoints
type Manager interface {
	// Create a new endpoint with the given data
	Create(endpoint *models.Endpoint) (string, error)
	// Get the specified endpoint
	Get(id string) (*models.Endpoint, error)
	// Update the endpoint
	Update(endpoint *models.Endpoint) error
	// Delete the specified endpoint
	Delete(id string) error
	// List all the endpoint with query
	List(query *q.Query) ([]*models.Endpoint, error)
	// Exist checks existence of the given endpoint by UUID
	Exist(UUID string) (bool, error)
}

// basicManager is the default implementation of Manager
type basicManager struct{}

// New a basic manager
func New() Manager {
	return &basicManager{}
}

// Create ...
func (bm *basicManager) Create(endpoint *models.Endpoint) (string, error) {
	if endpoint == nil {
		return "", errors.New("nil endpoint to create")
	}

	// Inject new UUID
	uid, err := uuid.NewUUID()
	if err != nil {
		return "", errors.Wrap(err, "new UUID")
	}
	endpoint.UUID = uid.String()

	if err := endpoint.Validate(); err != nil {
		return "", errors.Wrap(err, "create endpoint")
	}

	if _, err := dao.AddEndpoint(endpoint); err != nil {
		return "", errors.Wrap(err, "dao: add endpoint")
	}

	return uid.String(), nil
}

// Get ...
func (bm *basicManager) Get(id string) (*models.Endpoint, error) {
	if len(id) == 0 {
		return nil, errors.New("empty uuid of endpoint")
	}

	return dao.GetEndpoint(id)
}

// Update
func (bm *basicManager) Update(endpoint *models.Endpoint) error {
	if endpoint == nil {
		return errors.New("nil endpoint to update")
	}

	if err := endpoint.Validate(); err != nil {
		return errors.Wrap(err, "update endpoint")
	}

	return dao.UpdateEndpoint(endpoint)
}

// Delete
func (bm *basicManager) Delete(id string) error {
	if len(id) == 0 {
		return errors.New("empty id to delete")
	}

	return dao.DeleteEndpoint(id)
}

// List ...
func (bm *basicManager) List(query *q.Query) ([]*models.Endpoint, error) {
	return dao.ListEndpoints(query)
}

// Exist ...
func (bm *basicManager) Exist(UUID string) (bool, error) {
	return dao.EndpointExists(UUID)
}
