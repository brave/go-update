package v4

import (
	"encoding/json"
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

	var req Request
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(req.UpdateRequest.Extensions) != 1 {
		t.Errorf("Expected 1 extension, got %d", len(req.UpdateRequest.Extensions))
	}

	if req.UpdateRequest.Extensions[0].ID != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%s'", req.UpdateRequest.Extensions[0].ID)
	}

	if req.UpdateRequest.Extensions[0].Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", req.UpdateRequest.Extensions[0].Version)
	}

	if req.UpdateRequest.Extensions[0].FP != "test-sha256-hash" {
		t.Errorf("Expected fingerprint 'test-sha256-hash', got '%s'", req.UpdateRequest.Extensions[0].FP)
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
			Size:    100,
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

	// First pipeline should be the diff pipeline (puff_diff)
	if len(pipelines) > 1 {
		diffPipeline := pipelines[0].(map[string]interface{})
		expectedPipelinePrefix := "puff_diff_"
		if pipelineID, pipelineOK := diffPipeline["pipeline_id"].(string); !pipelineOK || !strings.HasPrefix(pipelineID, expectedPipelinePrefix) {
			t.Errorf("Expected pipeline_id to start with '%s', got '%v'", expectedPipelinePrefix, diffPipeline["pipeline_id"])
		}

		diffOperations, diffOpsOK := diffPipeline["operations"].([]interface{})
		if !diffOpsOK || len(diffOperations) != 3 {
			t.Fatalf("Expected 3 operations in diff pipeline, got %d", len(diffOperations))
		}

		// Check download operation in diff pipeline
		downloadOp := diffOperations[0].(map[string]interface{})
		if downloadOp["type"] != "download" {
			t.Errorf("Expected operation type 'download', got '%v'", downloadOp["type"])
		}
		if outObj, outOK := downloadOp["out"].(map[string]interface{}); !outOK || outObj["sha256"] != "test-hash-diff" {
			t.Errorf("Expected download operation to have out.sha256 'test-hash-diff'")
		}
		if urls, urlsOK := downloadOp["urls"].([]interface{}); !urlsOK || len(urls) == 0 {
			t.Errorf("Expected download operation to have URLs")
		}

		// Check puff operation in diff pipeline
		puffOp := diffOperations[1].(map[string]interface{})
		if puffOp["type"] != "puff" {
			t.Errorf("Expected operation type 'puff', got '%v'", puffOp["type"])
		}
		if prevObj, prevOK := puffOp["previous"].(map[string]interface{}); !prevOK || prevObj["sha256"] != "test-fp" {
			t.Errorf("Expected puff operation to have previous.sha256 'test-fp'")
		}
		if outObj, outOK := puffOp["out"].(map[string]interface{}); !outOK || outObj["sha256"] != "test-sha256" {
			t.Errorf("Expected puff operation to have out.sha256 'test-sha256'")
		}

		// Check crx3 operation in diff pipeline
		crx3OpDiff := diffOperations[2].(map[string]interface{})
		if crx3OpDiff["type"] != "crx3" {
			t.Errorf("Expected operation type 'crx3', got '%v'", crx3OpDiff["type"])
		}
		if inObj, inOK := crx3OpDiff["in"].(map[string]interface{}); !inOK || inObj["sha256"] != "test-sha256" {
			t.Errorf("Expected crx3 operation to have in.sha256 'test-sha256'")
		}
	}

	// Last pipeline should be direct_full
	fullPipeline := pipelines[len(pipelines)-1].(map[string]interface{})
	if fullPipeline["pipeline_id"] != "direct_full" {
		t.Errorf("Expected last pipeline_id to be 'direct_full', got '%v'", fullPipeline["pipeline_id"])
	}

	operations, opsOK := fullPipeline["operations"].([]interface{})
	if !opsOK || len(operations) != 2 {
		t.Fatalf("Expected 2 operations in full pipeline, got %d", len(operations))
	}

	// Check download operation in full pipeline
	downloadOp := operations[0].(map[string]interface{})
	if downloadOp["type"] != "download" {
		t.Errorf("Expected operation type 'download', got '%v'", downloadOp["type"])
	}
	if outObj, outOK := downloadOp["out"].(map[string]interface{}); !outOK || outObj["sha256"] != "test-sha256" {
		t.Errorf("Expected download operation to have out.sha256 'test-sha256'")
	}
	if urls, urlsOK := downloadOp["urls"].([]interface{}); !urlsOK || len(urls) == 0 {
		t.Errorf("Expected download operation to have URLs")
	}

	// Check crx3 operation in full pipeline
	crx3Op := operations[1].(map[string]interface{})
	if crx3Op["type"] != "crx3" {
		t.Errorf("Expected operation type 'crx3', got '%v'", crx3Op["type"])
	}
	if inObj, inOK := crx3Op["in"].(map[string]interface{}); !inOK || inObj["sha256"] != "test-sha256" {
		t.Errorf("Expected crx3 operation to have in.sha256 'test-sha256'")
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

	updateRequest, err := handler.ParseRequest(jsonData, "application/json")
	if err != nil {
		t.Fatalf("Failed to parse JSON request: %v", err)
	}

	if len(updateRequest.Extensions) != 1 {
		t.Errorf("Expected 1 extension, got %d", len(updateRequest.Extensions))
	}

	if updateRequest.Extensions[0].ID != "test-app-id" {
		t.Errorf("Expected app ID 'test-app-id', got '%s'", updateRequest.Extensions[0].ID)
	}

	if updateRequest.UpdaterType != "" {
		t.Errorf("Expected empty updater type for request without @updater, got '%s'", updateRequest.UpdaterType)
	}

	// Test JSON response formatting
	extensions := extension.Extensions{
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
