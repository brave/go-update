package v4

import (
	"fmt"

	"github.com/brave/go-update/extension"
	"github.com/brave/go-update/omaha/protocol"
)

// SupportedV4Versions is a list of v4.x protocol versions that are supported
var SupportedV4Versions = map[string]bool{
	"4.0": true,
}

type VersionedHandler struct {
	version string
}

// NewProtocol returns a Protocol implementation for the specified version
func NewProtocol(version string) (protocol.Protocol, error) {
	if !SupportedV4Versions[version] {
		return nil, fmt.Errorf("unsupported protocol version: %s", version)
	}
	return &VersionedHandler{
		version: version,
	}, nil
}

// GetVersion returns the protocol version
func (h *VersionedHandler) GetVersion() string {
	return h.version
}

// ParseRequest parses a request in the appropriate format (JSON or XML)
func (h *VersionedHandler) ParseRequest(data []byte, contentType string) (extension.Extensions, error) {
	var request UpdateRequest
	var err error

	// Only support JSON format for v4
	if contentType != "application/json" {
		return nil, fmt.Errorf("protocol v4 only supports JSON format")
	}

	err = request.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}

	return extension.Extensions(request), nil
}

// FormatUpdateResponse formats a standard update response in the appropriate format based on content type
func (h *VersionedHandler) FormatUpdateResponse(extensions extension.Extensions, _ string) ([]byte, error) {
	response := UpdateResponse(extensions)
	return response.MarshalJSON()
}

// FormatWebStoreResponse formats a web store response in the appropriate format based on content type
func (h *VersionedHandler) FormatWebStoreResponse(_ extension.Extensions, _ string) ([]byte, error) {
	return nil, fmt.Errorf("FormatWebStoreResponse not implemented for protocol v4: WebStore responses always use protocol v3.1")
}
