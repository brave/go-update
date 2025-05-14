package protocol

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

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

// DetectProtocolVersion attempts to detect the protocol version from the request
// Supported protocol versions are implemented in version-specific packages (e.g., v3)
func DetectProtocolVersion(data []byte, contentType string) (string, error) {
	if len(data) == 0 {
		// No data provided, default to 3.1
		return "3.1", nil
	}

	if IsJSONRequest(contentType) {
		var req struct {
			Request struct {
				Protocol string `json:"protocol"`
			} `json:"request"`
		}
		err := json.Unmarshal(data, &req)
		if err != nil {
			return "", fmt.Errorf("error parsing JSON request: %v", err)
		}

		if req.Request.Protocol == "" {
			return "", fmt.Errorf("malformed JSON request, missing 'protocol' field")
		}

		return req.Request.Protocol, nil
	}

	// Parse XML to extract protocol version
	var req struct {
		XMLName  xml.Name `xml:"request"`
		Protocol string   `xml:"protocol,attr"`
	}

	err := xml.Unmarshal(data, &req)
	if err != nil {
		return "", fmt.Errorf("error parsing XML: %v", err)
	}

	if req.Protocol == "" {
		return "", fmt.Errorf("protocol attribute not found in request element")
	}

	return req.Protocol, nil
}

// IsJSONRequest determines if the request is in JSON format based on content type
func IsJSONRequest(contentType string) bool {
	return contentType == "application/json"
}
