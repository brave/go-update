package v4

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestUnmarshalJSON(t *testing.T) {
	// Empty data returns an error
	var req Request
	err := json.Unmarshal([]byte(""), &req)
	assert.NotNil(t, err, "UnmarshalJSON should return an error for empty content")

	// Malformed JSON returns an error
	req = Request{}
	err = json.Unmarshal([]byte("{"), &req)
	assert.NotNil(t, err, "UnmarshalJSON should return an error for malformed JSON")

	// Wrong schema returns an error
	req = Request{}
	err = json.Unmarshal([]byte(`{"foo":"hello world!"}`), &req)
	assert.NotNil(t, err, "UnmarshalJSON should return an error for wrong JSON Schema")

	// No extensions JSON with proper schema, no error with 0 extensions returned
	data := []byte(`{"request":{"protocol":"4.0","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`)
	req = Request{}
	err = json.Unmarshal(data, &req)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(req.UpdateRequest.Extensions))

	// Test v4.0 request format with single app
	v4RequestData := []byte(`{
		"request": {
			"protocol": "4.0",
			"@updater": "chromiumcrx",
			"acceptformat": "download,xz,zucc,puff,crx3,run",
			"apps": [
				{
					"appid": "test-v4-app-id",
					"version": "2.0.0",
					"cached_items": [
						{ "sha256": "test-sha256-hash" }
					],
					"updatecheck": {}
				}
			]
		}
	}`)
	req = Request{}
	err = json.Unmarshal(v4RequestData, &req)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(req.UpdateRequest.Extensions))
	assert.Equal(t, "test-v4-app-id", req.UpdateRequest.Extensions[0].ID)
	assert.Equal(t, "2.0.0", req.UpdateRequest.Extensions[0].Version)
	assert.Equal(t, "test-sha256-hash", req.UpdateRequest.Extensions[0].FP)
	assert.Equal(t, "chromiumcrx", req.UpdateRequest.UpdaterType)

	// Test v4.0 request with multiple apps
	v4MultiAppRequestData := []byte(`{
		"request": {
			"protocol": "4.0",
			"@updater": "BraveComponentUpdater",
			"acceptformat": "download,xz,zucc,puff,crx3,run",
			"apps": [
				{
					"appid": "test-v4-app-id-1",
					"version": "2.0.0",
					"cached_items": [
						{ "sha256": "test-sha256-hash-1" }
					],
					"updatecheck": {}
				},
				{
					"appid": "test-v4-app-id-2",
					"version": "3.0.0",
					"cached_items": [
						{ "sha256": "test-sha256-hash-2" }
					],
					"updatecheck": {}
				}
			]
		}
	}`)
	req = Request{}
	err = json.Unmarshal(v4MultiAppRequestData, &req)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(req.UpdateRequest.Extensions))
	assert.Equal(t, "test-v4-app-id-1", req.UpdateRequest.Extensions[0].ID)
	assert.Equal(t, "2.0.0", req.UpdateRequest.Extensions[0].Version)
	assert.Equal(t, "test-sha256-hash-1", req.UpdateRequest.Extensions[0].FP)
	assert.Equal(t, "test-v4-app-id-2", req.UpdateRequest.Extensions[1].ID)
	assert.Equal(t, "3.0.0", req.UpdateRequest.Extensions[1].Version)
	assert.Equal(t, "test-sha256-hash-2", req.UpdateRequest.Extensions[1].FP)
	assert.Equal(t, "BraveComponentUpdater", req.UpdateRequest.UpdaterType)

	// Test with empty cached_items
	v4EmptyCachedItemsData := []byte(`{
		"request": {
			"protocol": "4.0",
			"acceptformat": "download,xz,zucc,puff,crx3,run",
			"apps": [
				{
					"appid": "test-v4-app-id",
					"version": "2.0.0",
					"cached_items": [],
					"updatecheck": {}
				}
			]
		}
	}`)
	req = Request{}
	err = json.Unmarshal(v4EmptyCachedItemsData, &req)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(req.UpdateRequest.Extensions))
	assert.Equal(t, "test-v4-app-id", req.UpdateRequest.Extensions[0].ID)
	assert.Equal(t, "2.0.0", req.UpdateRequest.Extensions[0].Version)
	assert.Equal(t, "", req.UpdateRequest.Extensions[0].FP)
}
