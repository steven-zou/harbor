package api

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/distribution/storage"

	"github.com/goharbor/harbor/src/distribution/models"

	"github.com/goharbor/harbor/src/distribution"
)

// DistributionAPI provide related API for image distribution related actions
type DistributionAPI struct {
	//The base controller to provide common utilities
	BaseController
}

// Prepare materia if needed for the follow-up actions
func (da *DistributionAPI) Prepare() {
	// Call super prepare
	da.BaseController.Prepare()

	// So far, only the system admin has permission to trigger distribution actions
	if !da.SecurityCtx.IsSysAdmin() {
		da.CustomAbort(http.StatusForbidden, fmt.Sprintf("user %s has no system admin permissions", da.SecurityCtx.GetUsername()))
	}
}

// ListProviders lists all the existing providers
// GET /api/distribution/providers
func (da *DistributionAPI) ListProviders() {
	providers, err := distribution.DefaultController.GetAvailableProviders()
	if err != nil {
		da.handleError(err)
		return
	}

	da.Data["json"] = providers
	da.ServeJSON()
}

// ListAllInstances lists all the configured instances
// GET /api/distribution/instances
func (da *DistributionAPI) ListAllInstances() {
	// TODO: parse and append query parameters
	instances, err := distribution.DefaultController.ListInstances(nil)
	if err != nil {
		da.handleError(err)
		return
	}

	da.Data["json"] = instances
	da.ServeJSON()
}

// GetInstance gets data of the specified instance
// GET /api/distribution/instances/:id
func (da *DistributionAPI) GetInstance() {
	id := da.GetStringFromPath(":id")
	meta, err := distribution.DefaultController.GetInstance(id)
	if err != nil {
		da.handleError(err)
		return
	}

	da.Data["json"] = meta
	da.ServeJSON()
}

// RemoveInstance removes the specified instance
// DELETE /api/distribution/instances/:id
func (da *DistributionAPI) RemoveInstance() {
	id := da.GetStringFromPath(":id")
	if err := distribution.DefaultController.DeleteInstance(id); err != nil {
		da.handleError(err)
		return
	}

	reply := make(map[string]string, 1)
	reply["removed"] = id

	da.Data["json"] = reply
	da.ServeJSON()
}

// CreateInstacne creates a new instances based on the specified provider
// POST /api/distribution/instances
func (da *DistributionAPI) CreateInstacne() {
	meta := &models.Metadata{}
	da.DecodeJSONReq(meta)

	id, err := distribution.DefaultController.CreateInstance(meta)
	if err != nil {
		da.handleError(err)
		return
	}

	res := make(map[string]string, 1)
	res["id"] = id

	da.Data["json"] = res
	da.ServeJSON()
}

// UpdateInstance updates the instance info
// PUT /api/distribution/instances/:id
func (da *DistributionAPI) UpdateInstance() {
	id := da.GetStringFromPath(":id")

	propertySet := make(models.PropertySet)
	da.DecodeJSONReq(&propertySet)

	if err := distribution.DefaultController.UpdateInstance(id, propertySet); err != nil {
		da.handleError(err)
		return
	}

	reply := make(map[string]string, 1)
	reply["updated"] = id

	da.Data["json"] = reply
	da.ServeJSON()
}

// ListHistories list all the distribution preheat histories
// GET /api/distribution/preheats
func (da *DistributionAPI) ListHistories() {
	histories, err := distribution.DefaultController.LoadHistoryRecords(nil)
	if err != nil {
		da.handleError(err)
		return
	}

	da.Data["json"] = histories
	da.ServeJSON()
}

// PreheatImage preheats the image to the distribution network
// POST /api/distribution/preheats
func (da *DistributionAPI) PreheatImage() {
	preheatReq := make(map[string]interface{})
	da.DecodeJSONReq(&preheatReq)

	preheatingImages, ok := preheatReq["images"]
	if !ok {
		da.HandleBadRequest("missing 'images'")
		return
	}

	imageList, ok := preheatingImages.([]interface{})
	if !ok {
		da.HandleBadRequest("'images' should be an array")
		return
	}

	if len(imageList) == 0 {
		da.HandleBadRequest("no images submitted")
		return
	}

	imageRepos := []models.ImageRepository{}
	for _, img := range imageList {
		imageRepos = append(imageRepos, models.ImageRepository(img.(string)))
	}

	res, err := distribution.DefaultController.PreheatImages(imageRepos...)
	if err != nil {
		da.handleError(err)
		return
	}

	da.Data["json"] = res
	da.ServeJSON()
}

func (da *DistributionAPI) handleError(err error) {
	if err == storage.ErrObjectNotFound {
		da.HandleNotFound(err.Error())
	} else if err == distribution.ErrorConflict {
		da.HandleConflict(err.Error())
	} else {
		da.HandleInternalServerError(err.Error())
	}
}
