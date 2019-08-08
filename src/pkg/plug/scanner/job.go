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
	"bytes"
	"encoding/json"
	"reflect"
	"time"

	"github.com/goharbor/harbor/src/jobservice/logger"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"
	"github.com/pkg/errors"
)

const (
	// JobParamEndpoint ...
	JobParamEndpoint = "endpoint"
	// JobParameterRequest ...
	JobParameterRequest = "scanRequest"
)

// Job for running scan in the job service with async way
type Job struct{}

// MaxFails for defining the number of retries
func (j *Job) MaxFails() uint {
	return 3
}

// ShouldRetry indicates if the job should be retried
func (j *Job) ShouldRetry() bool {
	return true
}

// Validate the parameters of this job
func (j *Job) Validate(params job.Parameters) error {
	if params == nil {
		// Params are required
		return errors.New("missing parameter of scan job")
	}

	_, err := extractEndpoint(params)
	_, err = extractScanReq(params)

	return err
}

// Run the job
func (j *Job) Run(ctx job.Context, params job.Parameters) error {
	// Get logger
	myLogger := ctx.GetLogger()

	// Ignore errors as they have been validated already
	e, _ := extractEndpoint(params)
	req, _ := extractScanReq(params)

	myLogger.Infof("Scan artifact:\n %#v\n through scanner:\n %#v", req, e)

	client, err := NewClient(e)
	if err != nil {
		return errors.Wrap(err, "run scan job")
	}
	resp, err := client.Scan(req)
	if err != nil {
		return errors.Wrap(err, "run scan job")
	}

	// Loop check if the report is ready
	tk := time.NewTicker(10 * time.Second)
	defer tk.Stop()

	var report *models.ScanReport

CHECK:
	for {
		select {
		case t := <-tk.C:
			myLogger.Debugf("check scan report: %s", t.Format("2006/01/02 15:04:05"))

			report, err = client.GetReport(resp.DetailsKey)
			if err != nil {
				return errors.Wrap(err, "check scan report")
			}

			// TODO: THERE SHOULD BE A STATUS PROPERTY TO INDICATE IF THE REPORT IS READY
			if report.Severity > 0 {
				break CHECK
			}

		case <-time.After(30 * time.Minute):
			return errors.New("check scan report timeout")
		}
	}

	// check in report
	resData, err := json.Marshal(report)
	if err != nil {
		return errors.Wrap(err, "scan job: marshal report to JSON")
	}

	if err := ctx.Checkin(string(resData)); err != nil {
		return errors.Wrap(err, "scan job: check in scan result")
	}

	printPrettyJSON(resData, myLogger)

	return nil
}

func printPrettyJSON(in []byte, logger logger.Interface) {
	var out bytes.Buffer
	if err := json.Indent(&out, in, "", "  "); err != nil {
		logger.Errorf("Print pretty JSON error: %s", err)
		return
	}

	logger.Infof("%s\n", out.String())
}

func extractScanReq(params job.Parameters) (*models.ScanRequest, error) {
	v, ok := params[JobParameterRequest]
	if !ok {
		return nil, errors.Errorf("missing job parameter '%s'", JobParameterRequest)
	}

	jsonData, ok := v.(string)
	if !ok {
		return nil, errors.Errorf(
			"malformed job parameter '%s', expecting string but got %s",
			JobParameterRequest,
			reflect.TypeOf(v).String(),
		)
	}

	req := &models.ScanRequest{}
	if err := req.FromJSON(jsonData); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}

func extractEndpoint(params job.Parameters) (*models.Endpoint, error) {
	v, ok := params[JobParamEndpoint]
	if !ok {
		return nil, errors.Errorf("missing job parameter '%s'", JobParamEndpoint)
	}

	jsonData, ok := v.(string)
	if !ok {
		return nil, errors.Errorf(
			"malformed job parameter '%s', expecting string but got %s",
			JobParamEndpoint,
			reflect.TypeOf(v).String(),
		)
	}

	e := &models.Endpoint{}
	if err := e.FromJSON(jsonData); err != nil {
		return nil, err
	}

	if err := e.Validate(true); err != nil {
		return nil, err
	}

	return e, nil
}
