package extension

import (
	"encoding/json"
	"testing"

	"github.com/brave/go-update/extension/extensiontest"
	"github.com/stretchr/testify/assert"
)

func TestUpdateResponseMarshalJSON(t *testing.T) {
	allExtensionsMap := NewExtensionMap()
	allExtensionsMap.StoreExtensions(&OfferedExtensions)
	// Empty extension list returns a blank JSON update
	updateResponse := UpdateResponse{}
	jsonData, err := json.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput := `{"response":{"protocol":"3.1","server":"prod","app":null}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	// Single extension list returns a single JSON update
	updateResponse = []Extension{darkThemeExtension}
	jsonData, err = json.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput = `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","required":true}]}}}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	// Multiple extensions returns a multiple extension JSON update
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	updateResponse = []Extension{lightThemeExtension, darkThemeExtension}
	jsonData, err = json.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput = `{"response":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + GetS3ExtensionBucketHost(lightThemeExtension.ID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618","required":true}]}}}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","urls":{"url":[{"codebase":"https://` + GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx"}]},"manifest":{"version":"1.0.0","packages":{"package":[{"name":"extension_1_0_0.crx","fp":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834","required":true}]}}}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))
}

func TestUpdateRequestUnmarshalJSON(t *testing.T) {
	// Empty data returns an error
	updateRequest := UpdateRequest{}
	err := json.Unmarshal([]byte(""), &updateRequest)
	assert.NotNil(t, err, "UnmarshalJSON should return an error for empty content")

	// Malformed JSON returns an error
	err = json.Unmarshal([]byte("{"), &updateRequest)
	assert.NotNil(t, err, "UnmarshalJSON should return an error for malformed JSON")

	// Wrong schema returns an error
	err = json.Unmarshal([]byte(`{"response":"hello world!"}`), &updateRequest)
	assert.NotNil(t, err, "UnmarshalJSON should return an error for wrong JSON Schema")

	// No extensions JSON with proper schema, no error with 0 extensions returned
	data := []byte(`{"request":{"protocol":"3.1","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`)
	err = json.Unmarshal(data, &updateRequest)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(updateRequest))

	onePasswordID := "aomjjhallfgjeglblehebfpbcfeobpgk" // #nosec
	onePasswordVersion := "4.7.0.90"
	onePasswordRequest := extensiontest.ExtensionRequestFnForJSON(onePasswordID)
	data = []byte(onePasswordRequest(onePasswordVersion))
	err = json.Unmarshal(data, &updateRequest)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(updateRequest))
	assert.Equal(t, onePasswordID, updateRequest[0].ID)
	assert.Equal(t, onePasswordVersion, updateRequest[0].Version)

	pdfJSID := "jdbefljfgobbmcidnmpjamcbhnbphjnb"
	pdfJSVersion := "1.0.0"
	twoExtensionRequest := extensiontest.ExtensionRequestFnForTwoJSON(onePasswordID, pdfJSID)
	data = []byte(twoExtensionRequest(onePasswordVersion, pdfJSVersion))
	err = json.Unmarshal(data, &updateRequest)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(updateRequest))
	assert.Equal(t, onePasswordID, updateRequest[0].ID)
	assert.Equal(t, onePasswordVersion, updateRequest[0].Version)
	assert.Equal(t, pdfJSID, updateRequest[1].ID)
	assert.Equal(t, pdfJSVersion, updateRequest[1].Version)

	// Check for unsupported protocol version
	data = []byte(`{"request":{"protocol":"2","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`)
	err = json.Unmarshal(data, &updateRequest)
	assert.NotNil(t, err, "Unrecognized protocol should have an error")

	// Test omaha4 protocol format
	data = []byte(`{
		"request": {
			"@os": "win",
			"@updater": "BraveComponentUpdater",
			"acceptformat": "crx3,download,puff,run",
			"apps": [
				{
					"appid": "aomjjhallfgjeglblehebfpbcfeobpgk",
					"enabled": true,
					"installsource": "ondemand",
					"ping": { "r": -2 },
					"updatecheck": {},
					"version": "4.7.0.90"
				}
			],
			"protocol": "4.0",
			"requestid": "{f122a713-8896-473a-a79f-c4be1755c47b}"
		}
	}`)
	err = json.Unmarshal(data, &updateRequest)
	assert.Nil(t, err, "Unmarshal should succeed for valid omaha4 format")
	assert.Equal(t, 1, len(updateRequest), "Should parse one extension from omaha4 request")
	// Expand assertions to check parsed values
	if len(updateRequest) == 1 {
		assert.Equal(t, onePasswordID, updateRequest[0].ID, "Parsed ID should match input")
		assert.Equal(t, onePasswordVersion, updateRequest[0].Version, "Parsed Version should match input")
	}
}

func TestWebStoreUpdateResponseMarshalJSON(t *testing.T) {
	// No extensions returns blank update response
	updateResponse := WebStoreUpdateResponse{}
	allExtensionsMap := NewExtensionMap()
	allExtensionsMap.StoreExtensions(&OfferedExtensions)
	jsonData, err := json.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput := `{"gupdate":{"protocol":"3.1","server":"prod","app":null}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	// Single extension list returns a single JSON update
	updateResponse = WebStoreUpdateResponse{darkThemeExtension}
	jsonData, err = json.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput = `{"gupdate":{"protocol":"3.1","server":"prod","app":[{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))

	// Multiple extensions returns a multiple extension JSON webstore update
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok = allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)
	updateResponse = WebStoreUpdateResponse{lightThemeExtension, darkThemeExtension}
	jsonData, err = json.Marshal(&updateResponse)
	assert.Nil(t, err)
	expectedOutput = `{"gupdate":{"protocol":"3.1","server":"prod","app":[{"appid":"ldimlcelhnjgpjjemdjokpgeeikdinbm","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + GetS3ExtensionBucketHost(lightThemeExtension.ID) + `/release/ldimlcelhnjgpjjemdjokpgeeikdinbm/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"1c714fadd4208c63f74b707e4c12b81b3ad0153c37de1348fa810dd47cfc5618"}},{"appid":"bfdgpgibhagkpdlnjonhkabjoijopoge","status":"ok","updatecheck":{"status":"ok","codebase":"https://` + GetS3ExtensionBucketHost(darkThemeExtension.ID) + `/release/bfdgpgibhagkpdlnjonhkabjoijopoge/extension_1_0_0.crx","version":"1.0.0","hash_sha256":"ae517d6273a4fc126961cb026e02946db4f9dbb58e3d9bc29f5e1270e3ce9834"}}]}}`
	assert.Equal(t, expectedOutput, string(jsonData))
}
