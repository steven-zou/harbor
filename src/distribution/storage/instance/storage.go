package instance

import "github.com/goharbor/harbor/src/distribution/storage"

// Storage is responsible for storing the instances
type Storage interface {
	// Save the instance metadata to the backend store
	//
	// inst *Metadata : a ptr of instance
	//
	// If succeed, the uuid of the saved instance is returned;
	// otherwise, a non nil error is returned
	//
	Save(inst *Metadata) (string, error)

	// Delete the specified instance
	//
	// id string : the uuid of the instance
	//
	// If succeed, a nil error is returned;
	// otherwise, a non nil error is returned
	//
	Delete(id string) error

	// Update the specified instance
	//
	// inst *Metadata : a ptr of instance
	//
	// If succeed, a nil error is returned;
	// otherwise, a non nil error is returned
	//
	Update(inst *Metadata) error

	// Get the instance with the ID
	//
	// id string : the uuid of the instance
	//
	// If succeed, a non nil Metadata is returned;
	// otherwise, a non nil error is returned
	//
	Get(id string) (*Metadata, error)

	// Query the instacnes by the param
	//
	// param *storage.QueryParam : the query params
	//
	// If succeed, an instance metadata list is returned;
	// otherwise, a non nil error is returned
	//
	List(param *storage.QueryParam) ([]*Metadata, error)
}
