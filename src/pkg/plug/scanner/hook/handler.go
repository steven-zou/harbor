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

package hook

import (
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/result"
	"github.com/pkg/errors"
)

// DefaultHandler ...
var DefaultHandler = New()

// Handler handles the hook events from the job service
type Handler interface {
	// HandleJobHooks handle the hook events from the job service
	// e.g : status change of the scan job or scan result
	//
	//   Arguments:
	//     trackID int64            : ID for the result record
	//     change *job.StatusChange : change event from the job service
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	HandleJobHooks(trackID int64, change *job.StatusChange) error
}

type basicHandler struct {
	resManager result.Manager
}

// New basic handler
func New() Handler {
	return &basicHandler{
		resManager: result.New(),
	}
}

// HandleJobHooks ...
func (bh *basicHandler) HandleJobHooks(trackID int64, change *job.StatusChange) error {
	if trackID <= 0 {
		return errors.New("empty track ID")
	}

	if change == nil {
		return errors.New("nil change  object")
	}

	if len(change.CheckIn) > 0 {
		// Retrieve the record first
		res, err := bh.resManager.Get(trackID)
		if err != nil {
			return errors.Wrap(err, "handle job hook")
		}

		// Update data
		res.Report = change.CheckIn
		res.EndTime = time.Now().UTC()

		if err := bh.resManager.Update(res); err != nil {
			return errors.Wrap(err, "handle job hook")
		}

		logger.Debugf("Scan: Check in: %s", change.CheckIn)
		return nil
	}

	return bh.resManager.UpdateStatus(trackID, change.Status)
}
