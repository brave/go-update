package v4

import (
	"encoding/xml"
	"fmt"
	"strings"

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

	if contentType == "application/json" {
		err = request.UnmarshalJSON(data)
	} else {
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

		err = request.UnmarshalXML(decoder, start)
	}

	if err != nil {
		return nil, err
	}

	return extension.Extensions(request), nil
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
