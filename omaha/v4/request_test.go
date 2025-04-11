package v4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestUnmarshalJSON(t *testing.T) {
	// Empty data returns an error
	request := UpdateRequest{}
	err := request.UnmarshalJSON([]byte(""))
	assert.NotNil(t, err, "UnmarshalJSON should return an error for empty content")

	// Malformed JSON returns an error
	err = request.UnmarshalJSON([]byte("{"))
	assert.NotNil(t, err, "UnmarshalJSON should return an error for malformed JSON")

	// Wrong schema returns an error
	err = request.UnmarshalJSON([]byte(`{"foo":"hello world!"}`))
	assert.NotNil(t, err, "UnmarshalJSON should return an error for wrong JSON Schema")

	// No extensions JSON with proper schema, no error with 0 extensions returned
	data := []byte(`{"request":{"protocol":"4.0","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`)
	err = request.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(request))

	// Test v4.0 request format with single app
	v4RequestData := []byte(`{
		"request": {
			"protocol": "4.0",
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
	err = request.UnmarshalJSON(v4RequestData)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(request))
	assert.Equal(t, "test-v4-app-id", request[0].ID)
	assert.Equal(t, "2.0.0", request[0].Version)
	assert.Equal(t, "test-sha256-hash", request[0].FP)

	// Test v4.0 request with multiple apps
	v4MultiAppRequestData := []byte(`{
		"request": {
			"protocol": "4.0",
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
	err = request.UnmarshalJSON(v4MultiAppRequestData)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(request))
	assert.Equal(t, "test-v4-app-id-1", request[0].ID)
	assert.Equal(t, "2.0.0", request[0].Version)
	assert.Equal(t, "test-sha256-hash-1", request[0].FP)
	assert.Equal(t, "test-v4-app-id-2", request[1].ID)
	assert.Equal(t, "3.0.0", request[1].Version)
	assert.Equal(t, "test-sha256-hash-2", request[1].FP)

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
	err = request.UnmarshalJSON(v4EmptyCachedItemsData)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(request))
	assert.Equal(t, "test-v4-app-id", request[0].ID)
	assert.Equal(t, "2.0.0", request[0].Version)
	assert.Equal(t, "", request[0].FP)
}
