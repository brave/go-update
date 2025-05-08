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
	updateResponse := UpdateResponse{extensionWithZeroSize, extensionWithoutSize}
	jsonData, err := updateResponse.MarshalJSON()
	assert.Nil(t, err)

	// Parse the response
	var actual map[string]interface{}
	err = json.Unmarshal(jsonData, &actual)
	assert.Nil(t, err)

	// Verify Size is positive in all cases
	resp := actual["response"].(map[string]interface{})
	apps := resp["apps"].([]interface{})

	// Check both extensions using a range loop
	for i := range 2 {
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
		size, exists := out["size"]
		assert.True(t, exists, "Size field should be present in download operation")
		assert.Greater(t, size, float64(0), "Size field should be greater than 0")

		// Verify Size is exactly 1 when input was 0 or unspecified
		switch i {
		case 0:
			// First extension had Size = 0
			assert.Equal(t, float64(1), size, "Size should be normalized to 1 for zero values")
		case 1:
			// Second extension had no Size field
			assert.Equal(t, float64(1), size, "Size should default to 1 when not specified")
		}
	}
}
