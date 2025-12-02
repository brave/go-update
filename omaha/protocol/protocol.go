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
	// It returns an UpdateRequest with extensions and metadata
	ParseRequest([]byte, string) (*extension.UpdateRequest, error)

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

// IsPingbackRequest checks if the request body is a pingback.
// For now, it only checks JSON requests for an "events" field.
func IsPingbackRequest(body []byte, contentType string) bool {
	if !IsJSONRequest(contentType) {
		return false
	}

	var pingbackCheck struct {
		Request struct {
			Apps []struct {
				Events *json.RawMessage `json:"events,omitempty"`
			} `json:"apps"`
		} `json:"request"`
	}

	if err := json.Unmarshal(body, &pingbackCheck); err == nil {
		if pingbackCheck.Request.Apps != nil {
			for _, app := range pingbackCheck.Request.Apps {
				if app.Events != nil {
					return true
				}
			}
		}
	}

	return false
}
