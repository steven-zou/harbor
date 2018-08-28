package distribution

import (
	"github.com/vmware/harbor/src/distribution/provider"
	"github.com/vmware/harbor/src/distribution/storage"
)

//CompositePreheatingResults handle preheating results among multiple providers
//Key is the ID of the provider instance.
type CompositePreheatingResults map[string][]*provider.PreheatingStatus

//Controller defines related top interfaces to handle the workflow of
//the image distribution.
type Controller interface {
	//Get all the supported distribution providers
	//If succeed, an provider driver array will be returned.
	//Otherwise, a non nil error will be returned
	GetAvailableProviders() ([]provider.Driver, error)

	//Get all the setup instances of distribution providers
	//If succeed, an provider instance array will be returned.
	//Otherwise, a non nil error will be returned
	//
	//If onlyEnabled is set to true, only return the enabled ones.
	GetInstances(onlyEnabled bool) ([]*provider.Instance, error)

	//Create a new instance for the specified provider.
	//Any problems met, a non nil error will be returned.
	CreateInstance(instance *provider.Instance) error

	//Delete the specified provider instance.
	//Any problems met, a non nil error will be returned.
	DeleteInstance(ID string) error

	//Enable the specified instance if it is not enabled yet.
	//Any problems met, a non nil error will be returned.
	EnableInstance(ID string) error

	//Disable the specified instance if it is enabled.
	//Any problems met, a non nil error will be returned.
	DisableInstance(ID string) error

	//Preheat images.
	//If multiple images are provided, the status of each image will be returned respectively.
	//One preheating failure will not cause the whole process fail.
	//If meet internal problems rather than failure results returned by the providers,
	//an non nil error will be returned.
	PreheatImages(images []*provider.PreheatImage) (CompositePreheatingResults, error)

	//Load the history records on top of the query parameters.
	LoadHistoryRecords(params storage.QueryParam) ([]*storage.HistroryRecord, error)
}

//DefaultController is the default implementation of Controller interface.
type DefaultController struct{}
