package auth

import (
	"fmt"
	"reflect"
)

// Manager is used to manage and index kinds of auth handlers like
// "NONE", "BASIC", "OAUTH" and "CUSTOM"
type Manager interface {
	// Register an auth handler or multiple handlers
	//
	// Any problems met, an non-nil error is returned.
	Register(handlers ...interface{}) error

	// Get the handler by mode
	//
	// If existing, the bool flag will be set to true and the handler reference will be returned.
	GetAuthHandler(mode string) (Handler, bool)
}

// BaseManager is default implementation of auth handler manager.
type BaseManager struct {
	// Keep the handlers
	handlers map[string]interface{}
}

// NewBaseManager is constructor of BaseManager.
func NewBaseManager() *BaseManager {
	return &BaseManager{
		handlers: make(map[string]interface{}),
	}
}

// Register implements @Manager.Register
func (b *BaseManager) Register(handlers ...interface{}) error {
	if len(handlers) == 0 {
		return nil
	}

	handlerList := make(map[string]interface{})
	for _, handler := range handlers {
		if _, ok := handler.(Handler); !ok {
			return fmt.Errorf("the handler should implement the 'Handler' interface: %v", handler)
		}

		registeringHandler := newAuthHandler(handler)
		if _, existing := b.handlers[registeringHandler.Mode()]; existing {
			return fmt.Errorf("handler with mode %s is already registered", registeringHandler.Mode())
		}

		handlerList[registeringHandler.Mode()] = handler
	}

	// Copy in
	for m, h := range handlerList {
		b.handlers[m] = h
	}

	return nil
}

// GetAuthHandler implements @Manager.GetAuthHandler
func (b *BaseManager) GetAuthHandler(mode string) (Handler, bool) {
	h, existing := b.handlers[mode]

	return newAuthHandler(h), existing
}

func newAuthHandler(d interface{}) Handler {
	theType := reflect.TypeOf(d)

	if theType.Kind() == reflect.Ptr {
		theType = theType.Elem()
	}

	//Crate new
	v := reflect.New(theType).Elem()
	return v.Addr().Interface().(Handler)
}
