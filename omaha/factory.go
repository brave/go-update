package omaha

import (
	"github.com/brave/go-update/omaha/protocol"
	v3impl "github.com/brave/go-update/omaha/v3"
)

// Factory creates protocol handlers for a specific version
type Factory interface {
	// CreateProtocol returns a Protocol implementation for the requested version
	CreateProtocol(version string) (protocol.Protocol, error)
}

// DefaultFactory is the default implementation of Factory
type DefaultFactory struct{}

// CreateProtocol returns a Protocol implementation for the requested version
func (f *DefaultFactory) CreateProtocol(version string) (protocol.Protocol, error) {
	// Each version package now handles its own validation
	return v3impl.NewProtocol(version)
}
