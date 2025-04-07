package v3

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/brave/go-update/extension"
	"github.com/brave/go-update/omaha/common"
)

// Protocol defines methods to support the Omaha v3 protocol
type Protocol interface {
	// GetVersion returns the protocol version
	GetVersion() string
	// ParseRequest parses a request in the appropriate format
	ParseRequest([]byte, string) (extension.UpdateRequest, error)
	// FormatResponse formats a response in the appropriate format based on content type
	FormatResponse(extension.UpdateResponse, bool, string) ([]byte, error)
}

// VersionedHandler is a unified implementation of the Protocol interface
// that handles both v3.0 and v3.1 requests
type VersionedHandler struct {
	version string
}

// NewProtocol returns a Protocol implementation for the specified version
func NewProtocol(version string) (Protocol, error) {
	if version != "3.0" && version != "3.1" {
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
func (h *VersionedHandler) ParseRequest(data []byte, contentType string) (extension.UpdateRequest, error) {
	var request Request
	var err error

	if common.IsJSONRequest(contentType) {
		err = request.UnmarshalJSON(data, h.version)
	} else {
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
		err = request.UnmarshalXML(decoder, start, h.version)
	}

	if err != nil {
		return nil, err
	}

	return extension.UpdateRequest(request), nil
}

// FormatResponse formats a response in the appropriate format based on content type
func (h *VersionedHandler) FormatResponse(response extension.UpdateResponse, isWebStore bool, contentType string) ([]byte, error) {
	if common.IsJSONRequest(contentType) {
		if isWebStore {
			webStoreResponse := WebStoreResponse(response)
			return webStoreResponse.MarshalJSON(h.version)
		}
		standardResponse := Response(response)
		return standardResponse.MarshalJSON(h.version)
	}

	// XML response
	var buf strings.Builder
	encoder := xml.NewEncoder(&buf)
	var err error

	if isWebStore {
		webStoreResponse := WebStoreResponse(response)
		err = webStoreResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "gupdate"}}, h.version)
	} else {
		standardResponse := Response(response)
		err = standardResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "response"}}, h.version)
	}

	if err != nil {
		return nil, err
	}

	err = encoder.Flush()
	if err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}
