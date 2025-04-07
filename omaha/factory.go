package omaha

import (
	"fmt"

	v3impl "github.com/brave/go-update/omaha/v3"
)

// Factory creates protocol handlers for a specific version
type Factory interface {
	// CreateProtocol returns a Protocol implementation for the requested version
	CreateProtocol(version string) (Protocol, error)
}

// DefaultFactory is the default implementation of Factory
type DefaultFactory struct{}

// CreateProtocol returns a Protocol implementation for the requested version
func (f *DefaultFactory) CreateProtocol(version string) (Protocol, error) {
	// Support for Omaha v3.x
	if version == "3.0" || version == "3.1" {
		return v3impl.NewProtocol(version)
	}

	return nil, fmt.Errorf("unsupported protocol version: %s", version)
}
