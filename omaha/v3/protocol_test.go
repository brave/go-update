package v3

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/brave/go-update/extension"
)

func TestNewProtocol(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantErr     bool
		wantVersion string
	}{
		{
			name:        "Valid version 3.0",
			version:     "3.0",
			wantErr:     false,
			wantVersion: "3.0",
		},
		{
			name:        "Valid version 3.1",
			version:     "3.1",
			wantErr:     false,
			wantVersion: "3.1",
		},
		{
			name:    "Invalid version",
			version: "4.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocol, err := NewProtocol(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProtocol() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && protocol.GetVersion() != tt.wantVersion {
				t.Errorf("NewProtocol().GetVersion() = %v, want %v", protocol.GetVersion(), tt.wantVersion)
			}
		})
	}
}

// TestIsJSONRequest tests the content type check
func TestIsJSONRequest(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{
			name:        "JSON content type",
			contentType: "application/json",
			want:        true,
		},
		{
			name:        "XML content type",
			contentType: "application/xml",
			want:        false,
		},
		{
			name:        "Empty content type",
			contentType: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.contentType == "application/json"; got != tt.want {
				t.Errorf("Content type check = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestUnmarshalJSONV30(t *testing.T) {
	jsonStr := `{
		"request": {
		  "protocol": "3.0",
		  "app": [
			{
			  "appid": "test-app-id",
			  "version": "1.0.0",
			  "packages": {
				"package": [
				  {
					"fp": "test-fingerprint"
				  }
				]
			  }
			}
		  ]
		}
	  }`

	var request Request
	if err := request.UnmarshalJSON([]byte(jsonStr), "3.0"); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(request) != 1 {
		t.Errorf("Expected 1 extension, got %d", len(request))
	}

	if request[0].ID != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%s'", request[0].ID)
	}

	if request[0].Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", request[0].Version)
	}

	if request[0].FP != "test-fingerprint" {
		t.Errorf("Expected fingerprint 'test-fingerprint', got '%s'", request[0].FP)
	}
}

func TestRequestUnmarshalJSONV31(t *testing.T) {
	jsonStr := `{
		"request": {
		  "protocol": "3.1",
		  "app": [
			{
			  "appid": "test-app-id",
			  "version": "1.0.0",
			  "fp": "test-fingerprint"
			}
		  ]
		}
	  }`

	var request Request
	if err := request.UnmarshalJSON([]byte(jsonStr), "3.1"); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(request) != 1 {
		t.Errorf("Expected 1 extension, got %d", len(request))
	}

	if request[0].ID != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%s'", request[0].ID)
	}

	if request[0].Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", request[0].Version)
	}

	if request[0].FP != "test-fingerprint" {
		t.Errorf("Expected fingerprint 'test-fingerprint', got '%s'", request[0].FP)
	}
}

func TestRequestUnmarshalXMLV30(t *testing.T) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
	<request protocol="3.0">
	  <app appid="test-app-id" version="1.0.0">
		<updatecheck />
		<packages>
		  <package fp="test-fingerprint" />
		</packages>
	  </app>
	</request>`

	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	var start xml.StartElement
	for {
		token, err := decoder.Token()
		if err != nil {
			t.Fatalf("Failed to get XML token: %v", err)
		}
		if se, ok := token.(xml.StartElement); ok {
			start = se
			break
		}
	}

	var request Request
	if err := request.UnmarshalXML(decoder, start, "3.0"); err != nil {
		t.Fatalf("Failed to unmarshal XML: %v", err)
	}

	if len(request) != 1 {
		t.Errorf("Expected 1 extension, got %d", len(request))
	}

	if request[0].ID != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%s'", request[0].ID)
	}

	if request[0].Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", request[0].Version)
	}

	if request[0].FP != "test-fingerprint" {
		t.Errorf("Expected fingerprint 'test-fingerprint', got '%s'", request[0].FP)
	}
}

func TestRequestUnmarshalXMLV31(t *testing.T) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
	<request protocol="3.1">
	  <app appid="test-app-id" version="1.0.0" fp="test-fingerprint">
		<updatecheck />
	  </app>
	</request>`

	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	var start xml.StartElement
	for {
		token, err := decoder.Token()
		if err != nil {
			t.Fatalf("Failed to get XML token: %v", err)
		}
		if se, ok := token.(xml.StartElement); ok {
			start = se
			break
		}
	}

	var request Request
	if err := request.UnmarshalXML(decoder, start, "3.1"); err != nil {
		t.Fatalf("Failed to unmarshal XML: %v", err)
	}

	if len(request) != 1 {
		t.Errorf("Expected 1 extension, got %d", len(request))
	}

	if request[0].ID != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%s'", request[0].ID)
	}

	if request[0].Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", request[0].Version)
	}

	if request[0].FP != "test-fingerprint" {
		t.Errorf("Expected fingerprint 'test-fingerprint', got '%s'", request[0].FP)
	}
}

func TestResponseMarshalJSONV30(t *testing.T) {
	response := Response{
		{
			ID:      "test-app-id",
			Version: "1.0.0",
			SHA256:  "test-sha256",
		},
	}

	data, err := response.MarshalJSON("3.0")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON result: %v", err)
	}

	responseObj, ok := result["response"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'response' field in JSON output")
	}

	if responseObj["protocol"] != "3.0" {
		t.Errorf("Expected protocol '3.0', got '%v'", responseObj["protocol"])
	}

	apps, ok := responseObj["app"].([]interface{})
	if !ok || len(apps) != 1 {
		t.Fatalf("Expected 1 app in response")
	}

	app := apps[0].(map[string]interface{})
	if app["appid"] != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%v'", app["appid"])
	}

	updateCheck, ok := app["updatecheck"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected updatecheck in app")
	}

	if updateCheck["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", updateCheck["status"])
	}
}

func TestResponseMarshalJSONV31(t *testing.T) {
	response := Response{
		{
			ID:      "test-app-id",
			Version: "1.0.0",
			SHA256:  "test-sha256",
			PatchList: map[string]*extension.PatchInfo{
				"test-fp": {
					Hashdiff: "test-hash-diff",
					Namediff: "test-name-diff",
					Sizediff: 100,
				},
			},
			FP: "test-fp",
		},
	}

	data, err := response.MarshalJSON("3.1")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON result: %v", err)
	}

	responseObj, ok := result["response"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'response' field in JSON output")
	}

	if responseObj["protocol"] != "3.1" {
		t.Errorf("Expected protocol '3.1', got '%v'", responseObj["protocol"])
	}

	apps, ok := responseObj["app"].([]interface{})
	if !ok || len(apps) != 1 {
		t.Fatalf("Expected 1 app in response")
	}

	app := apps[0].(map[string]interface{})
	if app["appid"] != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%v'", app["appid"])
	}

	updateCheck, ok := app["updatecheck"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected updatecheck in app")
	}

	if updateCheck["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", updateCheck["status"])
	}

	// Verify that diff information is included in v3.1
	urls, ok := updateCheck["urls"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected urls in updatecheck")
	}

	urlList, ok := urls["url"].([]interface{})
	if !ok || len(urlList) < 2 {
		t.Fatalf("Expected at least 2 URLs for diff support")
	}
}

func TestWebStoreResponseMarshalJSONV30(t *testing.T) {
	response := WebStoreResponse{
		{
			ID:      "test-app-id",
			Version: "1.0.0",
			SHA256:  "test-sha256",
		},
	}

	data, err := response.MarshalJSON("3.0")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON result: %v", err)
	}

	gupdateObj, ok := result["gupdate"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'gupdate' field in JSON output")
	}

	if gupdateObj["protocol"] != "3.0" {
		t.Errorf("Expected protocol '3.0', got '%v'", gupdateObj["protocol"])
	}

	apps, ok := gupdateObj["app"].([]interface{})
	if !ok || len(apps) != 1 {
		t.Fatalf("Expected 1 app in response")
	}

	app := apps[0].(map[string]interface{})
	if app["appid"] != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%v'", app["appid"])
	}

	updateCheck, ok := app["updatecheck"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected updatecheck in app")
	}

	if updateCheck["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", updateCheck["status"])
	}
}

func TestWebStoreResponseMarshalJSONV31(t *testing.T) {
	response := WebStoreResponse{
		{
			ID:      "test-app-id",
			Version: "1.0.0",
			SHA256:  "test-sha256",
		},
	}

	data, err := response.MarshalJSON("3.1")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON result: %v", err)
	}

	gupdateObj, ok := result["gupdate"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'gupdate' field in JSON output")
	}

	if gupdateObj["protocol"] != "3.1" {
		t.Errorf("Expected protocol '3.1', got '%v'", gupdateObj["protocol"])
	}

	apps, ok := gupdateObj["app"].([]interface{})
	if !ok || len(apps) != 1 {
		t.Fatalf("Expected 1 app in response")
	}

	app := apps[0].(map[string]interface{})
	if app["appid"] != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%v'", app["appid"])
	}

	updateCheck, ok := app["updatecheck"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected updatecheck in app")
	}

	if updateCheck["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", updateCheck["status"])
	}
}

func TestProtocolHandler(t *testing.T) {
	// Test that protocol handler properly routes requests
	protocol30, _ := NewProtocol("3.0")
	jsonStr30 := `{
		"request": {
		  "protocol": "3.0",
		  "app": [
			{
			  "appid": "test-app-id",
			  "version": "1.0.0",
			  "packages": {
				"package": [
				  {
					"fp": "test-fingerprint"
				  }
				]
			  }
			}
		  ]
		}
	  }`

	updateRequest30, err := protocol30.ParseRequest([]byte(jsonStr30), "application/json")
	if err != nil {
		t.Fatalf("Failed to parse v3.0 request: %v", err)
	}

	if len(updateRequest30) != 1 {
		t.Errorf("Expected 1 extension in v3.0 request, got %d", len(updateRequest30))
	}

	// Test response formatting for v3.0
	response30 := extension.UpdateResponse{
		{
			ID:      "test-app-id",
			Version: "1.0.0",
			SHA256:  "test-sha256",
		},
	}

	jsonResponse30, err := protocol30.FormatResponse(response30, false, "application/json")
	if err != nil {
		t.Fatalf("Failed to format v3.0 response: %v", err)
	}

	// Check that response contains v3.0
	if !strings.Contains(string(jsonResponse30), `"protocol":"3.0"`) {
		t.Errorf("Expected v3.0 protocol in response")
	}

	// Test response formatting with XML content type
	xmlResponse30, err := protocol30.FormatResponse(response30, false, "application/xml")
	if err != nil {
		t.Fatalf("Failed to format v3.0 XML response: %v", err)
	}

	// Check that XML response contains v3.0
	if !strings.Contains(string(xmlResponse30), `protocol="3.0"`) {
		t.Errorf("Expected v3.0 protocol in XML response")
	}

	// Test that v3.1 protocol handler processes v3.1 requests
	protocol31, _ := NewProtocol("3.1")
	jsonStr31 := `{
		"request": {
		  "protocol": "3.1",
		  "app": [
			{
			  "appid": "test-app-id",
			  "version": "1.0.0",
			  "fp": "test-fingerprint"
			}
		  ]
		}
	  }`

	updateRequest31, err := protocol31.ParseRequest([]byte(jsonStr31), "application/json")
	if err != nil {
		t.Fatalf("Failed to parse v3.1 request: %v", err)
	}

	if len(updateRequest31) != 1 {
		t.Errorf("Expected 1 extension in v3.1 request, got %d", len(updateRequest31))
	}

	// Test response formatting for v3.1
	response31 := extension.UpdateResponse{
		{
			ID:      "test-app-id",
			Version: "1.0.0",
			SHA256:  "test-sha256",
			PatchList: map[string]*extension.PatchInfo{
				"test-fp": {
					Hashdiff: "test-hash-diff",
					Namediff: "test-name-diff",
					Sizediff: 100,
				},
			},
			FP: "test-fp",
		},
	}

	jsonResponse31, err := protocol31.FormatResponse(response31, false, "application/json")
	if err != nil {
		t.Fatalf("Failed to format v3.1 response: %v", err)
	}

	// Check that response contains v3.1 and diff information
	if !strings.Contains(string(jsonResponse31), `"protocol":"3.1"`) {
		t.Errorf("Expected v3.1 protocol in response")
	}

	if !strings.Contains(string(jsonResponse31), `"namediff"`) {
		t.Errorf("Expected diff information in v3.1 response")
	}

	// Test protocol version mismatch handling
	jsonStrWrong := `{
		"request": {
		  "protocol": "3.1",
		  "app": [
			{
			  "appid": "test-app-id",
			  "version": "1.0.0",
			  "fp": "test-fingerprint"
			}
		  ]
		}
	  }`

	_, err = protocol30.ParseRequest([]byte(jsonStrWrong), "application/json")
	if err == nil {
		t.Errorf("Expected error when using v3.1 request with v3.0 handler, but got none")
	}
}
