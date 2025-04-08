package omaha

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

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

	// FormatWebStoreResponse formats a web store update response based on content type
	FormatWebStoreResponse(extension.Extensions, string) ([]byte, error)
}

// SupportedProtocolVersions is a list of Omaha protocol versions that we support
var SupportedProtocolVersions = map[string]bool{
	"3.0": true,
	"3.1": true,
}

// IsProtocolVersionSupported checks if the given protocol version is supported
func IsProtocolVersionSupported(version string) bool {
	return SupportedProtocolVersions[version]
}

// DetectProtocolVersion attempts to detect the protocol version from the request
func DetectProtocolVersion(data []byte, contentType string) (string, error) {
	if len(data) == 0 {
		// No data provided, default to 3.1
		return "3.1", nil
	}

	if IsJSONRequest(contentType) {
		// Parse JSON to extract protocol version
		var rawRequest map[string]interface{}
		err := json.Unmarshal(data, &rawRequest)
		if err != nil {
			return "", fmt.Errorf("error parsing JSON request: %v", err)
		}

		requestObj, ok := rawRequest["request"].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("malformed JSON request, missing 'request' object")
		}

		protocolVersion, ok := requestObj["protocol"].(string)
		if !ok {
			return "", fmt.Errorf("malformed JSON request, missing 'protocol' field")
		}

		return protocolVersion, nil
	}

	// Parse XML to extract protocol version
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("error parsing XML: %v", err)
		}

		if se, ok := token.(xml.StartElement); ok {
			if se.Name.Local == "request" {
				for _, attr := range se.Attr {
					if attr.Name.Local == "protocol" {
						return attr.Value, nil
					}
				}
				return "", fmt.Errorf("protocol attribute not found in request element")
			}
		}
	}
	return "", fmt.Errorf("request element not found in XML")
}
