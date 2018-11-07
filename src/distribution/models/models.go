package models

// Metadata represents the basic info of one working node for the specified provider.
type Metadata struct {
	// Unique ID
	ID string

	// Based on which driver, identified by ID
	Provider string

	// The service endpoint of this instance
	Endpoint string

	// The authentication way supported
	AuthMode string `json:"auth_mode"`

	// The auth credential data if exists
	AuthData map[string]string `json:"auth_data"`

	// The health status
	Status string

	// Whether the instance is activated or not
	Enabled bool

	// The timestamp of instance setting up
	SetupTimestamp int64 `json:"setup_timestamp"`

	// Append more described data if needed
	Extensions map[string]string
}

// HistoryRecord represents one record of the image preheating process.
type HistoryRecord struct {
	Image     string
	Timestamp int64
	Status    string
	Provider  string
	Instance  string
}
