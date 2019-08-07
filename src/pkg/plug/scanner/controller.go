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
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/pkg/plug/scanner/result"

	tk "github.com/docker/distribution/registry/auth/token"
	cjob "github.com/goharbor/harbor/src/common/job"
	jm "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/endpoint"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"
	"github.com/pkg/errors"
)

const (
	// ProScannerMetaKey ...
	ProScannerMetaKey = "proScanner"
)

// Controller defines operations for scan controlling
type Controller interface {
	// Scan the given artifact
	//
	//   Arguments:
	//     artifact *res.Artifact : artifact to be scanned
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	Scan(artifact *models.Artifact) error

	// GetReport gets the reports for the given artifact identified by the digest
	//
	//   Arguments:
	//     artifact *res.Artifact : the scanned artifact
	//
	//   Returns:
	//     []models.Result : scan results by different scanner vendors
	//     error           : non nil error if any errors occurred
	GetReport(artifact *models.Artifact) ([]*models.Result, error)
}

// basicController is default implementation of Controller
// scan the given artifact with the scanner configured at the namespace level
type basicController struct {
	// for getting the project level configured scanner
	proMetaMgr metamgr.ProjectMetadataManager
	// endpoint manager
	endpointMgr endpoint.Manager
	// result manager
	resManager result.Manager
}

// NewController news a basic controller
func NewController() Controller {
	return &basicController{
		proMetaMgr:  metamgr.NewDefaultProjectMetadataManager(),
		endpointMgr: endpoint.New(),
		resManager:  result.New(),
	}
}

// Scan artifact
// Use digest of artifact as the track UUID
func (bc *basicController) Scan(artifact *models.Artifact) error {
	if artifact == nil {
		return errors.New("nil artifact to scan")
	}

	scanner, err := bc.getScanner(artifact)
	if err != nil {
		return errors.Wrap(err, "scan")
	}

	// Create report placeholder first
	reportPlaceholder := &models.Result{
		Digest:     artifact.Digest,
		EndpointID: scanner.UUID,
		Vendor:     scanner.Adapter,
		Status:     job.PendingStatus.String(),
		StatusCode: job.PendingStatus.Code(),
	}
	trackID, err := bc.resManager.Create(reportPlaceholder)
	if err != nil {
		return errors.Wrap(err, "scan")
	}

	// Launch the scan job
	_, err = launchScanJob(trackID, artifact, scanner)
	if err != nil {
		// Change status to error
		// Change end time to now
		reportPlaceholder.Status = job.ErrorStatus.String()
		reportPlaceholder.StatusCode = job.ErrorStatus.Code()
		reportPlaceholder.EndTime = time.Now().UTC()

		if er := bc.resManager.Update(reportPlaceholder); er != nil {
			err = errors.Wrap(er, err.Error())
		}

		return errors.Wrap(err, "scan: launch job")
	}

	return nil
}

// GetReport gets the reports for all the launched scanning.
// Use artifact digest to get the relevant report.
// As one project can only trigger one scanner, there will be only one report generated then.
func (bc *basicController) GetReport(artifact *models.Artifact) ([]*models.Result, error) {
	if artifact == nil {
		return nil, errors.New("nil artifact to retrieve the report")
	}

	scanner, err := bc.getScanner(artifact)
	if err != nil {
		return nil, errors.Wrap(err, "scan")
	}

	r, err := bc.resManager.GetBy(artifact.Digest, scanner.UUID)
	if err != nil {
		return nil, errors.Wrap(err, "scan")
	}

	reports := make([]*models.Result, 0)
	reports = append(reports, r)

	return reports, nil
}

func (bc *basicController) getScanner(artifact *models.Artifact) (scanner *models.Endpoint, err error) {
	// Get the project level configured scanner if existing
	values, err := bc.proMetaMgr.Get(artifact.NamespaceID)
	if err != nil {
		return nil, errors.Wrap(err, "get scanner")
	}

	if scannerID, exists := values[ProScannerMetaKey]; exists {
		// Get the info of the scanner by the ID
		scanner, err = bc.endpointMgr.Get(scannerID)
		if err != nil {
			return nil, errors.Wrap(err, "get scanner")
		}
	}

	// Not configured at pro level
	if scanner == nil {
		// look for the system default one
		allEndpoints, err := bc.endpointMgr.List(nil)
		if err != nil {
			return nil, errors.Wrap(err, "get scanner: get system default scanner")
		}

		// Only check the enabled ones
		for _, e := range allEndpoints {
			if !e.Disabled && e.IsDefault {
				scanner = e
				break
			}
		}
	}

	// Both are not configured
	if scanner == nil {
		err = errors.New("project level scanner and system default scanner are both not configured")
	}

	return
}

// launchScanJob launches a job to run scan
func launchScanJob(trackID int64, artifact *models.Artifact, ept *models.Endpoint) (jobID string, err error) {
	externalURL, err := config.ExtEndpoint()
	if err != nil {
		return "", errors.Wrap(err, "launch scan job")
	}

	// Generate access token
	accessToken, err := makeAccessToken(artifact.Repository, fmt.Sprintf("%s:%s", ept.Adapter, ept.UUID))
	if err != nil {
		return "", errors.Wrap(err, "launch scan job")
	}

	// New a scan request
	req := &models.ScanRequest{
		RegistryURL:   externalURL,
		RegistryToken: accessToken,
		Repository:    artifact.Repository,
		Tag:           artifact.Tag,
		Digest:        artifact.Digest,
	}

	// Create job
	params := make(map[string]interface{})

	endpointJSON, err := ept.ToJSON()
	if err != nil {
		return "", errors.Wrap(err, "launch scan job")
	}
	reqJSON, err := req.ToJSON()
	if err != nil {
		return "", errors.Wrap(err, "launch scan job")
	}

	params[JobParamEndpoint] = endpointJSON
	params[JobParameterRequest] = reqJSON
	j := &jm.JobData{
		Name: job.ImageScanJob,
		Metadata: &jm.JobMetadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
	}

	callbackURL := config.InternalCoreURL()
	hookURL := fmt.Sprintf("%s/service/notifications/jobs/scan/tasks/%d", callbackURL, trackID)
	j.StatusHook = hookURL

	return cjob.GlobalClient.SubmitJob(j)
}

func makeAccessToken(repository string, username string) (string, error) {
	access := []*tk.ResourceActions{
		{
			Type:    "repository",
			Name:    repository,
			Actions: []string{"pull"},
		},
	}

	accessToken, err := token.MakeToken(username, token.Registry, access)
	if err != nil {
		return "", errors.Wrap(err, "make access token")
	}

	return accessToken.Token, nil
}
