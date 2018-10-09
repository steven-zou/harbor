package provider

import "github.com/goharbor/harbor/src/distribution/auth"

// Instance represents one working node of specified provider.
type Instance struct {
	// Based on which driver, identified by ID
	Provider string

	// The service endpoint of this instance
	Endpoint string

	// The credential data if exists
	Credential *auth.Credential

	//The health status
	Status string

	//Whether the instance is activated or not
	Enabled bool

	//The timestamp of instance setting up
	SetupTimestamp int64

	//For other general usage
	Extensions map[string]interface{}
}
