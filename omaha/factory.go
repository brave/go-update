package omaha

import (
	"strings"

	"github.com/brave/go-update/omaha/protocol"
	v3impl "github.com/brave/go-update/omaha/v3"
	v4impl "github.com/brave/go-update/omaha/v4"
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
	// Check major version to determine which implementation to use
	if strings.HasPrefix(version, "3.") {
		return v3impl.NewProtocol(version)
	} else if strings.HasPrefix(version, "4.") {
		return v4impl.NewProtocol(version)
	}

	// Default to v3 for backward compatibility
	return v3impl.NewProtocol(version)
}
