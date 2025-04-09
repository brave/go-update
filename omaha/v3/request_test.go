package v3

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/brave/go-update/extension/extensiontest"
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
	data := []byte(`{"request":{"protocol":"3.1","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`)
	err = request.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(request))

	onePasswordID := "aomjjhallfgjeglblehebfpbcfeobpgk" // #nosec
	onePasswordVersion := "4.7.0.90"
	onePasswordRequest := extensiontest.ExtensionRequestFnForJSON(onePasswordID)
	data = []byte(onePasswordRequest(onePasswordVersion))
	err = request.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(request))
	assert.Equal(t, onePasswordID, request[0].ID)
	assert.Equal(t, onePasswordVersion, request[0].Version)

	pdfJSID := "jdbefljfgobbmcidnmpjamcbhnbphjnb"
	pdfJSVersion := "1.0.0"
	twoExtensionRequest := extensiontest.ExtensionRequestFnForTwoJSON(onePasswordID, pdfJSID)
	data = []byte(twoExtensionRequest(onePasswordVersion, pdfJSVersion))
	err = request.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(request))
	assert.Equal(t, onePasswordID, request[0].ID)
	assert.Equal(t, onePasswordVersion, request[0].Version)
	assert.Equal(t, pdfJSID, request[1].ID)
	assert.Equal(t, pdfJSVersion, request[1].Version)
}

func TestRequestUnmarshalXML(t *testing.T) {
	// Empty data returns an error
	request := UpdateRequest{}
	err := xml.Unmarshal([]byte(""), &request)
	assert.NotNil(t, err, "UnmarshalXML should return an error for empty content")

	// Malformed XML returns an error
	err = xml.Unmarshal([]byte("<"), &request)
	assert.NotNil(t, err, "UnmarshalXML should return an error for malformed XML")

	// Test v3.0 request
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
		<app appid="test-app-id" version="1.0.0">
			<updatecheck />
			<packages>
				<package fp="test-fingerprint" />
			</packages>
		</app>
		</request>`)

	decoder := xml.NewDecoder(strings.NewReader(string(data)))
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

	err = request.UnmarshalXML(decoder, start)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(request))
	assert.Equal(t, "test-app-id", request[0].ID)
	assert.Equal(t, "1.0.0", request[0].Version)
	assert.Equal(t, "test-fingerprint", request[0].FP)

	// Test v3.1 request
	data = []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.1" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
		<app appid="test-app-id" version="1.0.0" fp="test-fingerprint">
			<updatecheck />
		</app>
		</request>`)

	decoder = xml.NewDecoder(strings.NewReader(string(data)))
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

	request = UpdateRequest{}
	err = request.UnmarshalXML(decoder, start)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(request))
	assert.Equal(t, "test-app-id", request[0].ID)
	assert.Equal(t, "1.0.0", request[0].Version)
	assert.Equal(t, "test-fingerprint", request[0].FP)
}
