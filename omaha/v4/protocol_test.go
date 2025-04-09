package v4

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/brave/go-update/extension"
)

func TestRequestUnmarshalJSONV40(t *testing.T) {
	jsonStr := `{
		"request": {
		  "protocol": "4.0",
		  "acceptformat": "download,xz,zucc,puff,crx3,run",
		  "apps": [
			{
			  "appid": "test-app-id",
			  "version": "1.0.0",
			  "cached_items": [
				{ "sha256": "test-sha256-hash" }
			  ],
			  "updatecheck": {}
			}
		  ]
		}
	  }`

	var request UpdateRequest
	if err := request.UnmarshalJSON([]byte(jsonStr)); err != nil {
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

	if request[0].FP != "test-sha256-hash" {
		t.Errorf("Expected fingerprint 'test-sha256-hash', got '%s'", request[0].FP)
	}
}

func TestRequestUnmarshalXMLV40(t *testing.T) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
	<request protocol="4.0" acceptformat="download,xz,zucc,puff,crx3,run">
	  <app appid="test-app-id" version="1.0.0">
		<updatecheck />
		<cacheditems>
		  <cacheditem sha256="test-sha256-hash" />
		</cacheditems>
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

	var request UpdateRequest
	if err := request.UnmarshalXML(decoder, start); err != nil {
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

	if request[0].FP != "test-sha256-hash" {
		t.Errorf("Expected fingerprint 'test-sha256-hash', got '%s'", request[0].FP)
	}
}

func TestResponseMarshalJSONV40(t *testing.T) {
	response := UpdateResponse{
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

	data, err := response.MarshalJSON()
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

	if responseObj["protocol"] != "4.0" {
		t.Errorf("Expected protocol '4.0', got '%v'", responseObj["protocol"])
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

	// Verify that diff information is included in v4.0
	urls, ok := updateCheck["urls"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected urls in updatecheck")
	}

	urlList, ok := urls["url"].([]interface{})
	if !ok || len(urlList) < 2 {
		t.Fatalf("Expected at least 2 URLs for diff support")
	}
}

func TestWebStoreResponseMarshalJSONV40(t *testing.T) {
	response := WebStoreResponse{
		{
			ID:      "test-app-id",
			Version: "1.0.0",
			SHA256:  "test-sha256",
		},
	}

	data, err := response.MarshalJSON()
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

	if gupdateObj["protocol"] != "4.0" {
		t.Errorf("Expected protocol '4.0', got '%v'", gupdateObj["protocol"])
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

func TestNewProtocol(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "Valid version 4.0",
			version: "4.0",
			wantErr: false,
		},
		{
			name:    "Invalid version",
			version: "4.111111",
			wantErr: true,
		},
		{
			name:    "Empty version",
			version: "",
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
			if err == nil && protocol.GetVersion() != tt.version {
				t.Errorf("NewProtocol().GetVersion() = %v, want %v", protocol.GetVersion(), tt.version)
			}
		})
	}
}

func TestProtocolHandler(t *testing.T) {
	protocol40, err := NewProtocol("4.0")
	if err != nil {
		t.Fatalf("Failed to create v4.0 protocol: %v", err)
	}

	// Test v4.0 JSON request parsing
	jsonStr40 := `{
		"request": {
			"protocol": "4.0",
			"acceptformat": "download,xz,zucc,puff,crx3,run",
			"apps": [
				{
					"appid": "test-app-id",
					"version": "1.0.0",
					"cached_items": [
						{ "sha256": "test-sha256-hash" }
					],
					"updatecheck": {}
				}
			]
		}
	}`

	request40, err := protocol40.ParseRequest([]byte(jsonStr40), "application/json")
	if err != nil {
		t.Fatalf("Failed to parse v4.0 request: %v", err)
	}

	if len(request40) != 1 {
		t.Errorf("Expected 1 extension in request, got %d", len(request40))
	}

	// Test v4.0 response formatting with diff information
	response40 := UpdateResponse{
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

	jsonResponse40, err := protocol40.FormatUpdateResponse(extension.Extensions(response40), "application/json")
	if err != nil {
		t.Fatalf("Failed to format v4.0 response: %v", err)
	}

	// Check that response contains v4.0 and diff information
	if !strings.Contains(string(jsonResponse40), `"protocol":"4.0"`) {
		t.Errorf("Expected v4.0 protocol in response")
	}

	if !strings.Contains(string(jsonResponse40), `"namediff"`) {
		t.Errorf("Expected diff information in v4.0 response")
	}

	// Test web store response formatting
	webStoreResponse40, err := protocol40.FormatWebStoreResponse(extension.Extensions(response40), "application/json")
	if err != nil {
		t.Fatalf("Failed to format web store response: %v", err)
	}

	// Check that response contains gupdate
	if !strings.Contains(string(webStoreResponse40), `"gupdate"`) {
		t.Errorf("Expected gupdate in web store response")
	}
}
