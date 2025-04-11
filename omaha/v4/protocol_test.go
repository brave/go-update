package v4

import (
	"encoding/json"
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

func TestResponseMarshalJSONV40(t *testing.T) {
	// Set a constant elapsed days value for consistent test output
	GetElapsedDays = func() int { return 6284 }

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

	// Check for daystart object
	daystart, ok := responseObj["daystart"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'daystart' field in response")
	}

	// Check for the specific elapsed_days value (as float64 from JSON parsing)
	elapsedDays, ok := daystart["elapsed_days"].(float64)
	if !ok {
		t.Fatalf("Expected 'elapsed_days' as number in daystart")
	}
	if elapsedDays != 6284 {
		t.Errorf("Expected 'elapsed_days' to be 6284, got %v", elapsedDays)
	}

	apps, ok := responseObj["apps"].([]interface{})
	if !ok || len(apps) != 1 {
		t.Fatalf("Expected 1 app in response, got %v", responseObj["apps"])
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

	// Verify new pipeline format
	if updateCheck["nextversion"] != "1.0.0" {
		t.Errorf("Expected nextversion '1.0.0', got '%v'", updateCheck["nextversion"])
	}

	pipelines, ok := updateCheck["pipelines"].([]interface{})
	if !ok || len(pipelines) < 1 {
		t.Fatalf("Expected at least 1 pipeline in updatecheck")
	}

	pipeline := pipelines[0].(map[string]interface{})
	if pipeline["pipeline_id"] != "direct_full" {
		t.Errorf("Expected pipeline_id 'direct_full', got '%v'", pipeline["pipeline_id"])
	}

	operations, ok := pipeline["operations"].([]interface{})
	if !ok || len(operations) != 2 {
		t.Fatalf("Expected 2 operations in pipeline")
	}

	// Check the download operation
	downloadOp := operations[0].(map[string]interface{})
	if downloadOp["type"] != "download" {
		t.Errorf("Expected operation type 'download', got '%v'", downloadOp["type"])
	}

	// Check the crx3 operation
	crx3Op := operations[1].(map[string]interface{})
	if crx3Op["type"] != "crx3" {
		t.Errorf("Expected operation type 'crx3', got '%v'", crx3Op["type"])
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
	handler, err := NewProtocol("4.0")
	if err != nil {
		t.Fatalf("Failed to create protocol handler: %v", err)
	}

	// Test non-JSON content type rejection
	jsonData := []byte(`{
		"request": {
			"protocol": "4.0",
			"apps": [{"appid": "test-app-id"}]
		}
	}`)
	_, err = handler.ParseRequest(jsonData, "application/xml")
	if err == nil {
		t.Errorf("Expected error for non-JSON content type, got nil")
	}
	if err != nil && err.Error() != "protocol v4 only supports JSON format" {
		t.Errorf("Expected error message 'protocol v4 only supports JSON format', got '%s'", err.Error())
	}

	// Test JSON request parsing
	jsonData = []byte(`{
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
	}`)

	extensions, err := handler.ParseRequest(jsonData, "application/json")
	if err != nil {
		t.Fatalf("Failed to parse JSON request: %v", err)
	}

	if len(extensions) != 1 {
		t.Errorf("Expected 1 extension, got %d", len(extensions))
	}

	if extensions[0].ID != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%s'", extensions[0].ID)
	}

	// Test JSON response formatting
	extensions = extension.Extensions{
		{
			ID:      "test-app-id",
			Version: "1.0.0",
			SHA256:  "test-sha256",
		},
	}

	// Format update response
	response, err := handler.FormatUpdateResponse(extensions, "application/json")
	if err != nil {
		t.Fatalf("Failed to format update response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if _, ok := result["response"]; !ok {
		t.Errorf("Expected 'response' field in JSON output")
	}
}
