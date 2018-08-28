package provider

const (
	//DriverStatusHealthy represents the healthy status
	DriverStatusHealthy = "healthy"

	//DriverStatusUnHealthy represents the unhealthy status
	DriverStatusUnHealthy = "Unhealthy"

	//PreheatingImageTypeImage defines the 'image' type of preheating images
	PreheatingImageTypeImage = "image"

	//PreheatingStatusPending means the preheating is waiting for starting
	PreheatingStatusPending = "PENDING"

	//PreheatingStatusRunning means the preheating is ongoing
	PreheatingStatusRunning = "RUNNING"

	//PreheatingStatusSuccess means the preheating is success
	PreheatingStatusSuccess = "SUCCESS"

	//PreheatingStatusError means the preheating is failed with error
	PreheatingStatusError = "ERROR"

	//PreheatingStatusFail means the preheating is failed
	PreheatingStatusFail = "FAIL"
)

//Driver defines the capabilities one distribution provider should have.
//Includes:
//  Health checking
//  Preheat related
type Driver interface {
	//Self returns the metadata of the driver
	Self() *Metadata
	HealthChecker
	PreheatOperator
}

//Metadata contains the basic information of the provider.
type Metadata struct {
	ID          string
	Name        string
	Icon        string `json:"icon,omitempty"`
	Maintainers []string
	Version     string
	Source      string `json:"source,omitempty"`
}

//DriverStatus keeps the health status of driver.
type DriverStatus struct {
	Status string
}

//HealthChecker define the behavior of health checking for the driver.
type HealthChecker interface {
	//Try to get the health status of the driver.
	//If succeed, a non nil status object will be returned;
	//otherwise, a non nil error will be set.
	GetHealthStatus() (*DriverStatus, error)
}

//PreheatImage contains related information which can help providers to get/pull the images.
type PreheatImage struct {
	//The image content type, only support 'image' now
	Type string

	//The accessable URL of the preheating image
	URL string

	//The headers which will be sent to the above URL of preheating image
	Headers map[string]interface{}
}

//PreheatingStatus contains the related results/status of the preheating operation
//from the provider.
type PreheatingStatus struct {
	TaskID string `json:"task_id"`
	Status string
	Error  error `json:",omitempty"`
}

//PreheatOperator is designed to handle the preheat scenario.
//Preheat means transfer the preheating image to the network of distribution provider
//in advance.
type PreheatOperator interface {
	//Preheat the specified image
	//If succeed, a non nil result object with preheating task id will be returned;
	//otherwise, a non nil error will be set.
	PreheatImage(preheatingImage *PreheatImage) (*PreheatingStatus, error)

	//Check the status of the preheating process.
	//If succeed, a non nil status object with preheating status will be returned;
	//otherwise, a non nil error will be set.
	CheckPreheatingStatus(taskID string) (*PreheatingStatus, error)
}
