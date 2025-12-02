package v3

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/brave/go-update/extension"
	"github.com/brave/go-update/omaha/protocol"
)

// SupportedV3Versions is a list of v3.x protocol versions that are supported
var SupportedV3Versions = map[string]bool{
	"3.0": true,
	"3.1": true,
}

// VersionedHandler is a unified implementation of the Protocol interface
// that handles both v3.0 and v3.1 requests
type VersionedHandler struct {
	version string
}

// NewProtocol returns a Protocol implementation for the specified version
func NewProtocol(version string) (protocol.Protocol, error) {
	if !SupportedV3Versions[version] {
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
func (h *VersionedHandler) ParseRequest(data []byte, contentType string) (*extension.UpdateRequest, error) {
	var req Request

	if contentType == "application/json" {
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return req.UpdateRequest, nil
	}

	// Set up XML decoder
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var start xml.StartElement
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		if se, ok := token.(xml.StartElement); ok {
			start = se
			break
		}
	}

	if err := req.UnmarshalXML(decoder, start); err != nil {
		return nil, err
	}

	return req.UpdateRequest, nil
}

// FormatUpdateResponse formats a standard update response in the appropriate format based on content type
func (h *VersionedHandler) FormatUpdateResponse(extensions extension.Extensions, contentType string) ([]byte, error) {
	response := UpdateResponse(extensions)

	if contentType == "application/json" {
		return response.MarshalJSON()
	}

	// XML response
	var buf strings.Builder
	encoder := xml.NewEncoder(&buf)

	err := response.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "response"}})
	if err != nil {
		return nil, err
	}

	err = encoder.Flush()
	if err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}

// FormatWebStoreResponse formats a web store response in the appropriate format based on content type
func (h *VersionedHandler) FormatWebStoreResponse(extensions extension.Extensions, contentType string) ([]byte, error) {
	response := UpdateResponse(extensions)
	webStoreResponse := WebStoreResponse(response)

	if contentType == "application/json" {
		return webStoreResponse.MarshalJSON()
	}

	// XML response
	var buf strings.Builder
	encoder := xml.NewEncoder(&buf)

	err := webStoreResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "gupdate"}})
	if err != nil {
		return nil, err
	}

	err = encoder.Flush()
	if err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}
