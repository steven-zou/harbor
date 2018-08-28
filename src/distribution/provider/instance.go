package provider

//Instance represents one working node of specified provider.
type Instance struct {
	//Based on which driver
	Provider Driver

	//The service endpoint of this instance
	Endpoint string

	//The health status
	Status string

	//Whether the instance is activated or not
	Enabled bool

	//The timestamp of instance setting up
	SetupTimestamp int64

	//For other general usage
	Extensions map[string]interface{}
}
