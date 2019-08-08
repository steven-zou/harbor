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

package api

import (
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/pkg/plug/scanner/endpoint"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/q"
	"github.com/pkg/errors"
)

// ScannerAPI provides the API for managing the plugin scanners
type ScannerAPI struct {
	// The base controller to provide common utilities
	BaseController

	// Manager for scanner endpoint
	endpointMgr endpoint.Manager
}

// Prepare sth. for the subsequent actions
func (sa *ScannerAPI) Prepare() {
	// Call super prepare method
	sa.BaseController.Prepare()

	// Check access permissions
	if !sa.SecurityCtx.IsAuthenticated() {
		sa.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	if !sa.SecurityCtx.IsSysAdmin() {
		sa.SendForbiddenError(errors.New(sa.SecurityCtx.GetUsername()))
		return
	}

	// Use the default manager
	sa.endpointMgr = endpoint.DefaultManager
}

// Get the specified scanner
func (sa *ScannerAPI) Get() {
	if ept := sa.get(); ept != nil {
		// Response to the client
		sa.Data["json"] = ept
		sa.ServeJSON()
	}
}

// List all the scanners
func (sa *ScannerAPI) List() {
	p, pz, err := sa.GetPaginationParams()
	if err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: list all"))
		return
	}

	query := &q.Query{
		PageSize:   pz,
		PageNumber: p,
	}

	// Get query key words
	kws := make(map[string]string)
	kw := sa.GetString("q")
	if len(kw) > 0 {
		ws := strings.Split(kw, ",")
		for _, w := range ws {
			tw := strings.TrimSpace(w)
			kv := strings.Split(tw, "=")
			if len(kv) == 2 {
				// valid key and value for the querying keyword
				kws[kv[0]] = kv[1]
			}
		}
	}

	if len(kws) > 0 {
		query.Keywords = kws
	}

	all, err := sa.endpointMgr.List(query)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: list all"))
		return
	}

	// Response to the client
	sa.Data["json"] = all
	sa.ServeJSON()
}

// Create a new scanner
func (sa *ScannerAPI) Create() {
	e := &models.Endpoint{}
	if valid, err := sa.DecodeJSONReqAndValidate(e); !valid {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: create"))
		return
	}

	if err := e.Validate(false); err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: create"))
		return
	}

	// Explicitly check if conflict
	query := &q.Query{
		Keywords: map[string]string{
			"url": e.URL,
		},
	}
	es, err := sa.endpointMgr.List(query)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: create : check exists"))
		return
	}

	if len(es) > 0 {
		sa.SendConflictError(errors.Errorf("%s already exists", e.URL))
		return
	}

	uid, err := sa.endpointMgr.Create(e)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: create"))
		return
	}

	sa.Redirect(http.StatusCreated, uid)
}

// Update a scanner
func (sa *ScannerAPI) Update() {
	ept := sa.get()
	if ept == nil {
		// meet error
		return
	}

	// full dose updated
	e := &models.Endpoint{}
	if valid, err := sa.DecodeJSONReqAndValidate(e); !valid {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: update"))
		return
	}

	getChanges(ept, e)

	// in case missing it
	e.UUID = ept.UUID
	if err := sa.endpointMgr.Update(ept); err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: update"))
		return
	}

	sa.Redirect(http.StatusOK, e.UUID)
}

// Delete the scanner
func (sa *ScannerAPI) Delete() {
	ept := sa.get()
	if ept == nil {
		// meet error
		return
	}

	if err := sa.endpointMgr.Delete(ept.UUID); err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: delete"))
		return
	}

	sa.Data["json"] = ept
	sa.ServeJSON()
}

// get the specified scanner
func (sa *ScannerAPI) get() *models.Endpoint {
	uid := sa.GetStringFromPath(":uid")
	if len(uid) == 0 {
		sa.SendBadRequestError(errors.New("missing uid"))
		return nil
	}

	ept, err := sa.endpointMgr.Get(uid)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: get"))
		return nil
	}

	if ept == nil {
		// NOT found
		sa.SendNotFoundError(errors.Errorf("scanner: %s", uid))
		return nil
	}

	return ept
}

func getChanges(e *models.Endpoint, eChange *models.Endpoint) {
	e.URL = eChange.URL
	e.Auth = eChange.Auth
	e.AccessCredential = eChange.AccessCredential
	e.Disabled = eChange.Disabled
	e.IsDefault = eChange.IsDefault
}
