// Package di provides a dependency injection container.
package di

import (
	"fmt"
	"sync"
)

// Get returns a service from a container.
//
// If the name is empty, it is set to the type of the service.
//
// If the service is not found for the given name, it returns an error.
//
// If the service is not of the expected type, it returns an error.
//
// If the service creation fails, it returns an error.
func Get[S any](c *Container, name string) (s S, err error) {
	if name == "" {
		name = getServiceName[S]()
	}
	sw := c.get(name)
	if sw == nil {
		return s, fmt.Errorf("service %q not registered", name)
	}
	swi, ok := sw.(*serviceWrapperImpl[S])
	if !ok {
		return s, fmt.Errorf("service %q is not of type %T", name, s)
	}
	return swi.get(c)
}

// Set sets a service to a container.
//
// If the name is empty, it is set to the type of the service.
//
// If the service is already set, it panics.
func Set[S any](c *Container, name string, b Builder[S]) {
	if name == "" {
		name = getServiceName[S]()
	}
	sw := &serviceWrapperImpl[S]{
		builder: b,
	}
	c.set(name, sw)
}

// Container contains services.
type Container struct {
	mu                sync.Mutex
	services          map[string]serviceWrapper
	getServiceNames   map[string]struct{}
	getServiceOrdered []serviceWrapper
}

func (c *Container) get(name string) serviceWrapper {
	c.mu.Lock()
	defer c.mu.Unlock()
	sw, ok := c.services[name]
	if !ok {
		return nil
	}
	if c.getServiceNames == nil {
		c.getServiceNames = make(map[string]struct{})
	}
	_, ok = c.getServiceNames[name]
	if !ok {
		c.getServiceNames[name] = struct{}{}
		c.getServiceOrdered = append(c.getServiceOrdered, sw)
	}
	return sw
}

func (c *Container) set(name string, sw serviceWrapper) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.services == nil {
		c.services = make(map[string]serviceWrapper)
	}
	_, ok := c.services[name]
	if ok {
		panic(fmt.Sprintf("service %q already set", name))
	}
	c.services[name] = sw
}

// Close closes the container.
//
// It closes all services in reverse dependency order.
// The created services must not be used after this call.
//
// The container can be reused after this call.
func (c *Container) Close(onErr func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.getServiceOrdered) - 1; i >= 0; i-- {
		sw := c.getServiceOrdered[i]
		err := sw.close()
		if err != nil {
			onErr(err)
		}
	}
}

// Builder builds a service.
//
// The Close function allows to close the service.
// It can be nil if the service does not need to be closed.
// After it is called, the service instance must not be used anymore.
type Builder[S any] func(c *Container) (S, Close, error)

// Close closes a service.
type Close = func() error

type serviceWrapper interface {
	close() error
}

type serviceWrapperImpl[S any] struct {
	mu          sync.Mutex
	builder     Builder[S]
	initialized bool
	service     S
	cl          Close
}

func (sw *serviceWrapperImpl[S]) get(c *Container) (S, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	if sw.initialized {
		return sw.service, nil
	}
	s, cl, err := sw.builder(c)
	if err != nil {
		return s, err
	}
	sw.initialized = true
	sw.service = s
	sw.cl = cl
	return sw.service, nil
}

func (sw *serviceWrapperImpl[S]) close() error {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	if !sw.initialized || sw.cl == nil {
		return nil
	}
	err := sw.cl()
	sw.initialized = false
	sw.service = zero[S]()
	sw.cl = nil
	return err
}

// Must is a helper to call a function and panics if it returns an error.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func getServiceName[S any]() string {
	var s S
	if interface{}(s) != nil {
		return fmt.Sprintf("%T", s)
	}
	// For interface types, we need to get a pointer.
	v := fmt.Sprintf("%T", &s)
	v = v[1:] // Remove the leading "*".
	return v
}

func zero[S any]() S {
	var s S
	return s
}
