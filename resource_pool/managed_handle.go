package resource_pool

import (
	"sync/atomic"

	"github.com/dropbox/godropbox/errors"
)

// A resource handle managed by a resource pool.
type ManagedHandle interface {
	// This returns the handle's resource location.
	ResourceLocation() string

	// This returns the underlying resource handle (or error if the handle
	// is no longer active).
	Handle() (interface{}, error)

	// This returns the resource pool which owns this handle.
	Owner() ResourcePool

	// This indictes a user is done with the handle and releases the handle
	// back to the resource pool.
	Release() error

	// This indicates the handle is an invalid state, and that the
	// connection should be discarded from the connection pool.
	Discard() error
}

// A physical implementation of ManagedHandle
type ManagedHandleImpl struct {
	location string
	handle   interface{}
	pool     ResourcePool
	isActive int32 // atomic bool
	options  Options
}

// This creates a managed handle wrapper.
func NewManagedHandle(
	resourceLocation string,
	handle interface{},
	pool ResourcePool,
	options Options) ManagedHandle {

	h := &ManagedHandleImpl{
		location: resourceLocation,
		handle:   handle,
		pool:     pool,
		options:  options,
	}
	atomic.StoreInt32(&h.isActive, 1)

	return h
}

// See ManagedHandle for documentation.
func (c *ManagedHandleImpl) ResourceLocation() string {
	return c.location
}

// See ManagedHandle for documentation.
func (c *ManagedHandleImpl) Handle() (interface{}, error) {
	if atomic.LoadInt32(&c.isActive) == 0 {
		return c.handle, errors.New("Resource handle is no longer valid")
	}
	return c.handle, nil
}

// See ManagedHandle for documentation.
func (c *ManagedHandleImpl) Owner() ResourcePool {
	return c.pool
}

// See ManagedHandle for documentation.
func (c *ManagedHandleImpl) Release() error {
	if atomic.CompareAndSwapInt32(&c.isActive, 1, 0) {
		return c.pool.Release(c)
	}
	return nil
}

// See ManagedHandle for documentation.
func (c *ManagedHandleImpl) Discard() error {
	if atomic.CompareAndSwapInt32(&c.isActive, 1, 0) {
		return c.pool.Discard(c)
	}
	return nil
}
