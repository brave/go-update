package protocol

import (
	"github.com/brave/go-update/extension"
)

// Protocol defines the interface for different Omaha protocol versions
type Protocol interface {
	// GetVersion returns the protocol version string
	GetVersion() string

	// ParseRequest parses an update request according to this protocol version
	// It returns a slice of Extension objects and any error encountered
	ParseRequest([]byte, string) (extension.Extensions, error)

	// FormatUpdateResponse formats a standard update response based on content type
	FormatUpdateResponse(extension.Extensions, string) ([]byte, error)

	// FormatWebStoreResponse formats a web store response based on content type
	FormatWebStoreResponse(extension.Extensions, string) ([]byte, error)
}
