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
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/pkg/plug/scanner/models"

	"github.com/pkg/errors"

	"github.com/goharbor/harbor/src/pkg/plug/scanner/auth"

	v1 "github.com/goharbor/harbor/src/pkg/plug/scanner/rest/v1"
)

// Client defines the methods to access the adapter services that
// implement the REST API specs
type Client interface {
	// CheckHealth checks the health of the scanner
	CheckHealth() error

	// GetMetadata get the metadata of the given scanner
	//
	//   Returns:
	//     *models.Adapter : metadata of the given scanner
	//     error           : non nil error if any errors occurred
	GetMetadata() (*models.Adapter, error)

	// Scan the specified artifact
	//
	//   Arguments:
	//     req *models.ScanRequest : request including the registry and artifact data
	//
	//   Returns:
	//     *models.ScanResponse : response with UUID for tracking the scan results
	//     error                : non nil error if any errors occurred
	Scan(req *models.ScanRequest) (*models.ScanResponse, error)

	// GetReport get the scan report for the given artifact
	//
	//   Arguments:
	//     digest string : the digest of the artifact
	//   Returns:
	//     *models.ScanReport : the scan report of the given artifact
	//     error              : non nil error if any errors occurred
	GetReport(digest string) (*models.ScanReport, error)
}

// basicClient is default implementation of the Client interface
type basicClient struct {
	httpClient *http.Client
	spec       *v1.Spec
	authorizer auth.Authorizer
}

// NewClient news a basic client
func NewClient(e *models.Endpoint) Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &basicClient{
		httpClient: &http.Client{
			Transport: transport,
		},
		spec:       v1.New(e.URL),
		authorizer: auth.New(e.AccessCredential),
	}
}

// CheckHealth ...
func (c *basicClient) CheckHealth() error {
	req, err := http.NewRequest(http.MethodGet, c.spec.Health(), nil)
	if err != nil {
		return errors.Wrap(err, "scanner client: check health:")
	}

	data, err := c.send(req, http.StatusOK)
	if err != nil {
		if len(data) > 0 {
			err = errors.Wrap(err, string(data))
		}
		return errors.Wrap(err, "scanner client: check health:")
	}

	return nil
}

// GetMetadata ...
func (c *basicClient) GetMetadata() (*models.Adapter, error) {
	return nil, errors.New("not implemented")
}

// Scan ...
func (c *basicClient) Scan(req *models.ScanRequest) (*models.ScanResponse, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "scanner client: scan:")
	}

	request, err := http.NewRequest(http.MethodPost, c.spec.Scan(), bytes.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, "scanner client: scan:")
	}

	respData, err := c.send(request, http.StatusCreated)
	if err != nil {
		if len(respData) > 0 {
			err = errors.Wrap(err, string(respData))
		}
		return nil, errors.Wrap(err, "scanner client: scan: http call:")
	}

	resp := &models.ScanResponse{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return nil, errors.Wrap(err, "scanner client: scan: unmarshal response:")
	}

	return resp, nil
}

// GetReport ...
func (c *basicClient) GetReport(digest string) (*models.ScanReport, error) {
	if len(digest) == 0 {
		return nil, errors.New("empty digest")
	}

	req, err := http.NewRequest(http.MethodGet, c.spec.Report(digest), nil)
	if err != nil {
		return nil, errors.Wrap(err, "scanner client: get report:")
	}

	respData, err := c.send(req, http.StatusOK)
	if err != nil {
		if len(respData) > 0 {
			err = errors.Wrap(err, string(respData))
		}
		return nil, errors.Wrap(err, "scanner client: get report: http call:")
	}

	resp := &models.ScanReport{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return nil, errors.Wrap(err, "scanner client: get report: unmarshal response:")
	}

	return resp, nil
}

func (c *basicClient) send(req *http.Request, expectedStatus int) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedStatus {
		return b, errors.Errorf("Unexpected status code: %d, text: %s", resp.StatusCode, string(b))
	}

	return b, nil
}
