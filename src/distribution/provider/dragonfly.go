package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/distribution/auth"
	"github.com/goharbor/harbor/src/distribution/client"
	"github.com/goharbor/harbor/src/distribution/models"
)

const (
	healthCheckEndpoint = "/api/check"
	preheatEndpoint     = "/api/preheat"
	preheatTaskEndpoint = "/api/preheat/{taskId}"
)

type dragonflyResponse struct {
	Code int
	Msg  string                 `json:"msg,omitempty"`
	Data *dragonflyResponseData `json:"data,omitempty"`
}

type dragonflyResponseData struct {
	TaskID string `json:"taskId"`
	Status string `json:"status,omitempty"`
}

// DragonflyDriver implements the provider driver interface for Alibaba dragonfly.
// More details, please refer to https://github.com/alibaba/Dragonfly
type DragonflyDriver struct {
	instance *models.Metadata
}

// Self implements @Driver.Self.
func (dd *DragonflyDriver) Self() *Metadata {
	return &Metadata{
		ID:          "dragonfly",
		Name:        "Dragonfly",
		Icon:        "https://raw.githubusercontent.com/alibaba/Dragonfly/master/docs/images/logo.png",
		Version:     "0.10.1",
		Source:      "https://github.com/alibaba/Dragonfly",
		Maintainers: []string{"Steven Z/szou@vmware.com"},
		AuthMode:    auth.AuthModeNone,
	}
}

// GetHealth implements @Driver.GetHealth.
func (dd *DragonflyDriver) GetHealth() (*DriverStatus, error) {
	if dd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), healthCheckEndpoint)
	bytes, err := client.DefaultHTTPClient.Get(url, dd.getCred(), nil, nil)
	if err != nil {
		return nil, err
	}
	status := &dragonflyResponse{}
	if err := json.Unmarshal(bytes, status); err != nil {
		return nil, err
	}

	health := &DriverStatus{
		Status: DriverStatusHealthy,
	}

	if status.Code != 200 {
		health.Status = DriverStatusUnHealthy
	}

	return health, nil
}

// Preheat implements @Driver.Preheat.
func (dd *DragonflyDriver) Preheat(preheatingImage *PreheatImage) (*PreheatingStatus, error) {
	if dd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	if preheatingImage == nil {
		return nil, errors.New("no image specified")
	}

	body, err := json.Marshal(preheatingImage)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), preheatEndpoint)
	bytes, err := client.DefaultHTTPClient.Post(url, dd.getCred(), body, nil)
	if err != nil {
		return nil, err
	}

	result := &dragonflyResponse{}
	if err := json.Unmarshal(bytes, result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return &PreheatingStatus{
			Status: PreheatingStatusError,
			Error:  errors.New(result.Msg),
		}, nil
	}

	return &PreheatingStatus{
		TaskID: result.Data.TaskID,
		Status: PreheatingStatusPending, // default
	}, nil
}

// CheckProgress implements @Driver.CheckProgress.
func (dd *DragonflyDriver) CheckProgress(taskID string) (*PreheatingStatus, error) {
	if dd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	if len(taskID) == 0 {
		return nil, errors.New("no task ID")
	}

	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), preheatTaskEndpoint)
	bytes, err := client.DefaultHTTPClient.Get(url, dd.getCred(), nil, nil)
	if err != nil {
		return nil, err
	}

	status := &dragonflyResponse{}
	if err := json.Unmarshal(bytes, status); err != nil {
		return nil, err
	}

	if status.Code != 200 {
		return &PreheatingStatus{
			Status: PreheatingStatusError,
			Error:  errors.New(status.Msg),
		}, nil
	}

	return &PreheatingStatus{
		Status: status.Data.Status,
		TaskID: status.Data.TaskID,
	}, nil
}

func (dd *DragonflyDriver) getCred() *auth.Credential {
	return &auth.Credential{
		dd.instance.AuthMode,
		dd.instance.AuthData,
	}
}
