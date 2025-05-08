package v4

import (
	"encoding/json"
	"testing"

	"github.com/brave/go-update/extension"
	"github.com/stretchr/testify/assert"
)

func TestResponseMarshalJSON(t *testing.T) {
	// Set a constant elapsed days value for consistent test output
	GetElapsedDays = func() int { return 6284 }

	allExtensionsMap := extension.NewExtensionMap()
	allExtensionsMap.StoreExtensions(&extension.OfferedExtensions)

	// Empty extension list returns a blank JSON update
	updateResponse := UpdateResponse{}
	jsonData, err := updateResponse.MarshalJSON()
	assert.Nil(t, err)

	// Parse the actual response
	var actual map[string]interface{}
	err = json.Unmarshal(jsonData, &actual)
	assert.Nil(t, err)

	// Verify the empty extension case
	resp := actual["response"].(map[string]interface{})
	assert.Equal(t, "4.0", resp["protocol"])
	assert.Equal(t, float64(6284), resp["daystart"].(map[string]interface{})["elapsed_days"])
	assert.Nil(t, resp["apps"])

	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	// Single extension list returns a single JSON update
	updateResponse = []extension.Extension{darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)

	err = json.Unmarshal(jsonData, &actual)
	assert.Nil(t, err)

	// Verify the single extension case
	resp = actual["response"].(map[string]interface{})
	assert.Equal(t, "4.0", resp["protocol"])
	assert.Equal(t, float64(6284), resp["daystart"].(map[string]interface{})["elapsed_days"])

	apps := resp["apps"].([]interface{})
	assert.Equal(t, 1, len(apps))
	app := apps[0].(map[string]interface{})
	assert.Equal(t, "bfdgpgibhagkpdlnjonhkabjoijopoge", app["appid"])
	assert.Equal(t, "ok", app["status"])

	updateCheck := app["updatecheck"].(map[string]interface{})
	assert.Equal(t, "ok", updateCheck["status"])
	assert.Equal(t, "1.0.0", updateCheck["nextversion"])

	pipelines := updateCheck["pipelines"].([]interface{})
	assert.Equal(t, 1, len(pipelines))
	pipeline := pipelines[0].(map[string]interface{})
	assert.Equal(t, "direct_full", pipeline["pipeline_id"])

	// Multiple extensions returns a multiple extension JSON update
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	updateResponse = []extension.Extension{lightThemeExtension, darkThemeExtension}
	jsonData, err = updateResponse.MarshalJSON()
	assert.Nil(t, err)

	err = json.Unmarshal(jsonData, &actual)
	assert.Nil(t, err)

	// Verify the multiple extension case
	resp = actual["response"].(map[string]interface{})
	assert.Equal(t, "4.0", resp["protocol"])
	assert.Equal(t, float64(6284), resp["daystart"].(map[string]interface{})["elapsed_days"])

	apps = resp["apps"].([]interface{})
	assert.Equal(t, 2, len(apps))

	// First app should be lightThemeExtension
	app = apps[0].(map[string]interface{})
	assert.Equal(t, "ldimlcelhnjgpjjemdjokpgeeikdinbm", app["appid"])

	// Second app should be darkThemeExtension
	app = apps[1].(map[string]interface{})
	assert.Equal(t, "bfdgpgibhagkpdlnjonhkabjoijopoge", app["appid"])
}

func TestSizeValidation(t *testing.T) {
	// Set a constant elapsed days value for consistent test output
	GetElapsedDays = func() int { return 6284 }

	// Create a valid extension with normal size
	validExtension := extension.Extension{
		ID:      "test-extension-id-3",
		Version: "1.0.0",
		SHA256:  "test-sha256-3",
		Size:    100, // Normal size
	}

	extensionWithZeroSize := extension.Extension{
		ID:      "test-extension-id",
		Version: "1.0.0",
		SHA256:  "test-sha256",
		Size:    0, // Cannot be negative as Size is an unsigned int
	}

	extensionWithoutSize := extension.Extension{
		ID:      "test-extension-id-2",
		Version: "1.0.0",
		SHA256:  "test-sha256-2",
		// Size field not set, should default to 1
	}

	// Create responses with these extensions
	// With our new validation, we expect the validation to normalize the values
	updateResponse := UpdateResponse{validExtension, extensionWithZeroSize, extensionWithoutSize}
	jsonData, err := updateResponse.MarshalJSON()
	assert.Nil(t, err)

	// Parse the response
	var actual map[string]interface{}
	err = json.Unmarshal(jsonData, &actual)
	assert.Nil(t, err)

	// Verify Size is positive in all cases
	resp := actual["response"].(map[string]interface{})
	apps := resp["apps"].([]interface{})

	// We should have three extensions in the response
	assert.Equal(t, 3, len(apps))

	// Check all extensions using a range loop
	for i := range 3 {
		app := apps[i].(map[string]interface{})
		updateCheck := app["updatecheck"].(map[string]interface{})
		pipelines := updateCheck["pipelines"].([]interface{})

		// Get the direct_full pipeline (should be last or only pipeline)
		pipeline := pipelines[len(pipelines)-1].(map[string]interface{})
		assert.Equal(t, "direct_full", pipeline["pipeline_id"])

		// Check the download operation
		operations := pipeline["operations"].([]interface{})
		downloadOp := operations[0].(map[string]interface{})
		assert.Equal(t, "download", downloadOp["type"])

		// Verify Size is present and greater than 0
		out := downloadOp["out"].(map[string]interface{})
		size, exists := downloadOp["size"]
		assert.True(t, exists, "Size field should be present in download operation")
		assert.Greater(t, size, float64(0), "Size field should be greater than 0")

		// Verify SHA256 is present and not empty
		sha256, exists := out["sha256"]
		assert.True(t, exists, "SHA256 field should be present in download operation")
		assert.NotEmpty(t, sha256, "SHA256 field should not be empty")

		// Verify URLs are present and not empty
		urls, exists := downloadOp["urls"].([]interface{})
		assert.True(t, exists, "URLs field should be present in download operation")
		assert.NotEmpty(t, urls, "URLs field should not be empty")

		// Check the first URL
		url := urls[0].(map[string]interface{})
		assert.NotEmpty(t, url["url"], "URL field should not be empty")

		// Verify Size based on extension type
		switch i {
		case 0:
			// First extension had normal Size = 100
			assert.Equal(t, float64(100), size, "Size should remain 100 for normal values")
		case 1:
			// Second extension had Size = 0
			assert.Equal(t, float64(1), size, "Size should be normalized to 1 for zero values")
		case 2:
			// Third extension had no Size field
			assert.Equal(t, float64(1), size, "Size should default to 1 when not specified")
		}
	}
}

// Test validation failures for empty SHA256 and URLs
func TestValidationFailures(t *testing.T) {
	// Set a constant elapsed days value for consistent test output
	GetElapsedDays = func() int { return 6284 }

	// Extension with empty SHA256 should fail validation
	extensionWithEmptySHA256 := extension.Extension{
		ID:      "test-empty-sha256",
		Version: "1.0.0",
		SHA256:  "", // Empty SHA256 should fail validation
		Size:    100,
	}

	// Try to marshal this extension
	updateResponse := UpdateResponse{extensionWithEmptySHA256}
	_, err := updateResponse.MarshalJSON()

	// Expect validation error
	assert.NotNil(t, err, "Should fail validation with empty SHA256")
	assert.Contains(t, err.Error(), "has empty SHA256", "Error should mention empty SHA256")

	// Extension with empty FP (for Previous.SHA256) with PatchList should fail when trying to create a diff pipeline
	extensionWithEmptyFP := extension.Extension{
		ID:      "test-empty-fp",
		Version: "1.0.0",
		SHA256:  "valid-sha256",
		Size:    100,
		FP:      "", // This will be used as Previous.SHA256 and should fail validation
		PatchList: map[string]*extension.PatchInfo{
			"": { // Empty key to match empty FP
				Hashdiff: "valid-hashdiff",
			},
		},
	}

	// Try to marshal this extension - should not fail since FP validation now happens only when we have FP != ""
	updateResponse = UpdateResponse{extensionWithEmptyFP}
	_, err = updateResponse.MarshalJSON()
	assert.Nil(t, err, "Should not fail validation with empty FP since FP is empty and won't trigger diff pipeline")

	// Extension with empty Hashdiff in PatchList should fail
	extensionWithEmptyHashdiff := extension.Extension{
		ID:      "test-empty-hashdiff",
		Version: "1.0.0",
		SHA256:  "valid-sha256",
		Size:    100,
		FP:      "valid-fp",
		PatchList: map[string]*extension.PatchInfo{
			"valid-fp": { // Match FP to trigger diff pipeline
				Hashdiff: "", // Empty Hashdiff should fail
			},
		},
	}

	// Try to marshal this extension
	updateResponse = UpdateResponse{extensionWithEmptyHashdiff}
	_, err = updateResponse.MarshalJSON()

	// Expect validation error
	assert.NotNil(t, err, "Should fail validation with empty Hashdiff")
	assert.Contains(t, err.Error(), "has empty Hashdiff", "Error should mention empty Hashdiff")
}
