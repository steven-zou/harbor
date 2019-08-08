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
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/plug/scanner"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"
	"github.com/pkg/errors"
)

// ScanAPI handles the scan related actions
type ScanAPI struct {
	BaseController

	// Target artifact
	artifact *models.Artifact
}

// Prepare sth. for the subsequent actions
func (sa *ScanAPI) Prepare() {
	// Call super prepare method
	sa.BaseController.Prepare()

	// Parse parameters
	repoName := sa.GetString(":splat")
	tag := sa.GetString(":tag")
	projectName, repository := utils.ParseRepository(repoName)

	pro, err := sa.ProjectMgr.Get(projectName)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: prepare: get project"))
		return
	}
	if pro == nil {
		sa.SendNotFoundError(errors.Errorf("project %s not found", projectName))
		return
	}

	// Check access permissions
	if !sa.SecurityCtx.IsAuthenticated() {
		sa.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	resource := rbac.NewProjectNamespace(projectName).Resource(rbac.ResourceRepositoryTagScanJob)
	if !sa.SecurityCtx.Can(rbac.ActionCreate, resource) {
		sa.SendForbiddenError(errors.New(sa.SecurityCtx.GetUsername()))
		return
	}

	digest, err := getDigest(repoName, tag, sa.SecurityCtx.GetUsername())
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: get digest"))
		return
	}

	// Assemble artifact object
	sa.artifact = &models.Artifact{
		NamespaceID: pro.ProjectID,
		Namespace:   projectName,
		Repository:  repository,
		Tag:         tag,
		Digest:      digest,
		Kind:        "image",
	}

	logger.Debugf("scan artifact: %#v", sa.artifact)
}

// Scan artifact
func (sa *ScanAPI) Scan() {
	if err := scanner.DefaultController.Scan(sa.artifact); err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: scan"))
		return
	}
}

// Report returns report
func (sa *ScanAPI) Report() {
	results, err := scanner.DefaultController.GetReport(sa.artifact)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scan API: get report"))
		return
	}

	vulItems := make([]*models.VulnerabilityItem, 0)
	// Keep aligned with previous
	if len(results) > 0 {
		res := results[0]
		if len(res.Report) > 0 {
			report := &models.ScanReport{}
			if err := report.FromJSON(res.Report); err != nil {
				sa.SendInternalServerError(errors.Wrap(err, "scan API"))
				return
			}

			vulItems = report.Vulnerabilities
		}
	}

	sa.Data["json"] = vulItems
	sa.ServeJSON()
}

func getDigest(repo, tag string, username string) (string, error) {
	client, err := coreutils.NewRepositoryClientForUI(username, repo)
	if err != nil {
		return "", err
	}

	digest, exists, err := client.ManifestExist(tag)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", errors.Errorf("tag %s does exist", tag)
	}

	return digest, nil
}
