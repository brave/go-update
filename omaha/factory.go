package omaha

import (
	"fmt"

	v3impl "github.com/brave/go-update/omaha/v3"
	"github.com/brave/go-update/protocol"
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
	if !IsProtocolVersionSupported(version) {
		return nil, fmt.Errorf("unsupported protocol version: %s", version)
	}

	// Currently, all supported versions are Omaha v3.x
	return v3impl.NewProtocol(version)
}
