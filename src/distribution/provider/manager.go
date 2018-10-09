package provider

import (
	"errors"
	"fmt"
	"reflect"
)

// DefaultManager for default use.
var DefaultManager = NewBaseManager()

// Manager is used to maintain the multiple distribution providers.
type Manager interface {
	// Register a driver or multiple drivers.
	//
	// If failed, a non-nil error will be returned
	Register(drivers ...interface{}) error

	// ListProviders returns the all the registered drivers.
	ListProviders() []Driver

	// GetProvider returns the driver identified by the ID.
	//
	// If exists, bool flag will be set to be true and a non-nil reference will be returned.
	GetProvider(ID string) (Driver, bool)
}

// BaseManager is the default implementation of provider Manager interface.
type BaseManager struct {
	// The internal driver list
	drivers map[string]interface{}
}

// NewBaseManager is constructor of BaseManager.
func NewBaseManager() *BaseManager {
	return &BaseManager{
		drivers: make(map[string]interface{}),
	}
}

// Register drivers.
func (dm *BaseManager) Register(drivers ...interface{}) error {
	if len(drivers) == 0 {
		return nil // do nothing
	}

	validDrivers := make(map[string]interface{})
	for _, driver := range drivers {
		if _, ok := driver.(Driver); !ok {
			return errors.New("driver must implement provider.Driver interface")
		}

		inst := newDriver(driver)
		metaData := inst.Self()
		tmpIdentity := reflect.TypeOf(driver).String()
		if metaData == nil {
			return fmt.Errorf("missing metadata when registering driver %s", tmpIdentity)
		}

		// Check required info
		if len(metaData.ID) == 0 {
			return fmt.Errorf("missing ID in the metadata of driver %s", tmpIdentity)
		}
		if len(metaData.Name) == 0 {
			return fmt.Errorf("missing name in the metadata of driver %s", tmpIdentity)
		}
		if len(metaData.Version) == 0 {
			return fmt.Errorf("missing version in the metadata of driver %s", tmpIdentity)
		}

		// Avoid duplicate
		if _, existing := dm.drivers[metaData.ID]; existing {
			return fmt.Errorf("Driver with ID %s is already exitsing", metaData.ID)
		}

		validDrivers[metaData.ID] = driver
	}

	// Copy in
	for k, v := range validDrivers {
		dm.drivers[k] = v
	}

	return nil
}

// ListProviders lists all drivers.
func (dm *BaseManager) ListProviders() []*Metadata {
	drivers := make([]*Metadata, 0)

	for _, driver := range dm.drivers {
		inst := newDriver(driver)
		drivers = append(drivers, inst.Self())
	}

	return drivers
}

// GetProvider gets driver by ID.
func (dm *BaseManager) GetProvider(ID string) (Driver, bool) {
	if len(ID) == 0 {
		return nil, false
	}

	if intf, ok := dm.drivers[ID]; ok {
		return newDriver(intf), true
	}

	return nil, false
}

func newDriver(d interface{}) Driver {
	theType := reflect.TypeOf(d)

	if theType.Kind() == reflect.Ptr {
		theType = theType.Elem()
	}

	//Crate new
	v := reflect.New(theType).Elem()
	return v.Addr().Interface().(Driver)
}
