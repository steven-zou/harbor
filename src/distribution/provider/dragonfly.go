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
	healthCheckEndpoint = "/_ping"
	preheatEndpoint     = "/preheats"
	preheatTaskEndpoint = "/preheats/{task_id}"
	dragonflyPending    = "WAITING"
)

type dragonflyPreheatCreateResp struct {
	ID string `json:"ID"`
}

type dragonflyPreheatInfo struct {
	ID         string `json:"ID"`
	StartTime  string `json:"startTime,omitempty"`
	FinishTime string `json:"finishTime,omitempty"`
	Status     string
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

	url := fmt.Sprintf("%s%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), healthCheckEndpoint)
	_, err := client.DefaultHTTPClient.Get(url, dd.getCred(), nil, nil)
	if err != nil {
		// Unhealthy
		return nil, err
	}

	// For Dragonfly, no error returned means healthy
	return &DriverStatus{
		Status: DriverStatusHealthy,
	}, nil
}

// Preheat implements @Driver.Preheat.
func (dd *DragonflyDriver) Preheat(preheatingImage *PreheatImage) (*PreheatingStatus, error) {
	if dd.instance == nil {
		return nil, errors.New("missing instance metadata")
	}

	if preheatingImage == nil {
		return nil, errors.New("no image specified")
	}

	url := fmt.Sprintf("%s%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), preheatEndpoint)
	bytes, err := client.DefaultHTTPClient.Post(url, dd.getCred(), preheatingImage, nil)
	if err != nil {
		return nil, err
	}

	result := &dragonflyPreheatCreateResp{}
	if err := json.Unmarshal(bytes, result); err != nil {
		return nil, err
	}

	return &PreheatingStatus{
		TaskID: result.ID,
		Status: models.PreheatingStatusPending, // default
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

	path := strings.Replace(preheatTaskEndpoint, "{task_id}", taskID, 1)
	url := fmt.Sprintf("%s%s", strings.TrimSuffix(dd.instance.Endpoint, "/"), path)
	bytes, err := client.DefaultHTTPClient.Get(url, dd.getCred(), nil, nil)
	if err != nil {
		return nil, err
	}

	status := &dragonflyPreheatInfo{}
	if err := json.Unmarshal(bytes, status); err != nil {
		return nil, err
	}

	if status.Status == dragonflyPending {
		status.Status = models.PreheatingStatusPending
	}

	res := &PreheatingStatus{
		Status: status.Status,
		TaskID: taskID,
	}
	if status.StartTime != "" {
		res.StartTime = status.StartTime
	}
	if status.FinishTime != "" {
		res.FinishTime = status.FinishTime
	}

	return res, nil
}

func (dd *DragonflyDriver) getCred() *auth.Credential {
	return &auth.Credential{
		Mode: dd.instance.AuthMode,
		Data: dd.instance.AuthData,
	}
}
