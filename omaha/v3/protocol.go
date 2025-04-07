package v3

import (
	"encoding/xml"
	"strings"

	"github.com/brave/go-update/extension"
)

// Protocol defines methods to support the Omaha v3 protocol
type Protocol interface {
	// GetVersion returns the protocol version
	GetVersion() string
	// ParseRequest parses a request in the appropriate format
	ParseRequest([]byte, string) (extension.Extensions, error)
	// FormatResponse formats a response in the appropriate format based on content type
	FormatResponse(extension.Extensions, bool, string) ([]byte, error)
}

// VersionedHandler is a unified implementation of the Protocol interface
// that handles both v3.0 and v3.1 requests
type VersionedHandler struct {
	version string
}

// NewProtocol returns a Protocol implementation for the specified version
func NewProtocol(version string) (Protocol, error) {
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
	var request Request
	var err error

	if contentType == "application/json" {
		// Unmarshal the JSON data
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

		// Unmarshal the XML
		err = request.UnmarshalXML(decoder, start)
	}

	if err != nil {
		return nil, err
	}

	return extension.Extensions(request), nil
}

// FormatResponse formats a response in the appropriate format based on content type
func (h *VersionedHandler) FormatResponse(extensions extension.Extensions, isWebStore bool, contentType string) ([]byte, error) {
	response := Response(extensions)

	if contentType == "application/json" {
		if isWebStore {
			webStoreResponse := WebStoreResponse(response)
			return webStoreResponse.MarshalJSON()
		}
		return response.MarshalJSON()
	}

	// XML response
	var buf strings.Builder
	encoder := xml.NewEncoder(&buf)
	var err error

	if isWebStore {
		webStoreResponse := WebStoreResponse(response)
		err = webStoreResponse.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "gupdate"}})
	} else {
		err = response.MarshalXML(encoder, xml.StartElement{Name: xml.Name{Local: "response"}})
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
