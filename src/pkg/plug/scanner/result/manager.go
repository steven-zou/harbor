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

package result

import (
	"time"

	"github.com/goharbor/harbor/src/pkg/plug/scanner/dao"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"
	"github.com/pkg/errors"
)

// Manager maintains the scan results of the scanners
type Manager interface {
	// Get all the reports of the given digest
	GetAll(digest string) ([]*models.Result, error)

	// Get the specified result
	Get(id int64) (*models.Result, error)

	// Get by the digest and endpoint ID
	GetBy(digest, endpointID string) (*models.Result, error)

	// Create new result record
	Create(res *models.Result) (int64, error)

	// Update the given result
	Update(res *models.Result) error

	// Update the status of the result record
	UpdateStatus(trackID int64, status string) error
}

// basicManager implements the default result manager
// Only the latest scan report copy is persisted.
type basicManager struct{}

// New basic result manager
func New() Manager {
	return &basicManager{}
}

// GetBy ...
func (bm *basicManager) GetBy(digest, endpointID string) (*models.Result, error) {
	if len(digest) == 0 || len(endpointID) == 0 {
		return nil, errors.New("bad arguments for get by: missing digest or endpoint ID")
	}

	return dao.QueryRecord(digest, endpointID)
}

// GetAll ...
func (bm *basicManager) GetAll(digest string) ([]*models.Result, error) {
	if len(digest) == 0 {
		return nil, errors.New("empty digest to get all results")
	}

	return dao.GetAllByDigest(digest)
}

// Get the specified result
func (bm *basicManager) Get(id int64) (*models.Result, error) {
	return dao.GetRecord(id)
}

// Create a scan result
func (bm *basicManager) Create(res *models.Result) (int64, error) {
	if res == nil {
		return -1, errors.New("nil result object")
	}

	// validate the object
	if len(res.Digest) == 0 || len(res.EndpointID) == 0 {
		return -1, errors.New("malformed result object")
	}

	// Check if there is existing report copy
	existingCopy, err := dao.QueryRecord(res.Digest, res.EndpointID)
	if err != nil {
		return -1, errors.Wrap(err, "check existing report copy")
	}

	// Exists
	if existingCopy != nil {
		// Limit only one scanning performed by a given provider on the specified artifact can be there
		theStatus := job.Status(existingCopy.Status)
		if theStatus.Compare(job.RunningStatus) <= 0 {
			return -1, errors.Errorf("conflict: a previous scanning is %s", existingCopy.Status)
		}
		// Otherwise it will be a completed report
		// Clear it before insert this new one
		if err := dao.DeleteRecord(existingCopy.ID); err != nil {
			return -1, errors.Wrap(err, "clear old scan report")
		}
	}

	// Fill in / override the related properties
	res.StartTime = time.Now().UTC()
	res.Status = job.PendingStatus.String()
	res.StatusCode = job.PendingStatus.Code()

	return dao.CreateRecord(res)
}

// Update updates the report
func (bm *basicManager) Update(res *models.Result) error {
	if res == nil {
		return errors.New("nil report to update")
	}

	return dao.UpdateRecord(res)
}

// UpdateStatus updates the report (scanning action) status
func (bm *basicManager) UpdateStatus(trackID int64, status string) error {
	st := job.Status(status)
	if err := st.Validate(); err != nil {
		return errors.Wrap(err, "update report status")
	}

	return dao.UpdateRecordStatus(trackID, status, st.Code())
}
