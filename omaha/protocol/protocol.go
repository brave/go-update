package protocol

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"encoding/xml"
	"fmt"
	"strings"

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

	if IsJSONContentType(contentType) {
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

// Example: "application/json; charset=utf-8"
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Type
func IsJSONContentType(contentType string) bool {
	mediaType := strings.TrimSpace(contentType)
	if idx := strings.Index(mediaType, ";"); idx != -1 {
		mediaType = strings.TrimSpace(mediaType[:idx])
	}
	return mediaType == "application/json"
}

// Example: "text/html, application/json;q=0.9, */*;q=0.8"
//
// Noteworthy information:
// - quality values are ignored (simply checks for "application/json" presence)
// - wildcards (*/*) are not treated as JSON (intentionally defaults to XML for backward compatibility)
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Accept
func AcceptsJSON(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		mediaType := strings.TrimSpace(part)
		if idx := strings.Index(mediaType, ";"); idx != -1 {
			mediaType = strings.TrimSpace(mediaType[:idx])
		}
		if mediaType == "application/json" {
			return true
		}
	}
	return false
}

// IsPingbackRequest checks if the request body is a pingback.
// For now, it only checks JSON requests for an "events" field.
// Uses streaming token parsing for performance - avoids full unmarshal.
func IsPingbackRequest(body []byte, contentType string) bool {
	if !IsJSONContentType(contentType) {
		return false
	}

	dec := jsontext.NewDecoder(bytes.NewReader(body))

	for {
		tok, err := dec.ReadToken()
		if err != nil {
			return false
		}

		if tok.Kind() == '"' && tok.String() == "events" {
			val, err := dec.ReadValue()
			if err != nil {
				return false
			}
			// Check if events array has content (length > 2 means more than just "[]")
			if len(val) > 2 {
				return true
			}
		}
	}
}
