package v3

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/brave/go-update/extension/extensiontest"
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
	data := []byte(`{"request":{"protocol":"3.1","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"}}}`)
	req = Request{}
	err = json.Unmarshal(data, &req)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(req.UpdateRequest.Extensions))

	onePasswordID := "aomjjhallfgjeglblehebfpbcfeobpgk" // #nosec
	onePasswordVersion := "4.7.0.90"
	onePasswordRequest := extensiontest.ExtensionRequestFnForJSON(onePasswordID)
	data = []byte(onePasswordRequest(onePasswordVersion))
	req = Request{}
	err = json.Unmarshal(data, &req)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(req.UpdateRequest.Extensions))
	assert.Equal(t, onePasswordID, req.UpdateRequest.Extensions[0].ID)
	assert.Equal(t, onePasswordVersion, req.UpdateRequest.Extensions[0].Version)

	pdfJSID := "jdbefljfgobbmcidnmpjamcbhnbphjnb"
	pdfJSVersion := "1.0.0"
	twoExtensionRequest := extensiontest.ExtensionRequestFnForTwoJSON(onePasswordID, pdfJSID)
	data = []byte(twoExtensionRequest(onePasswordVersion, pdfJSVersion))
	req = Request{}
	err = json.Unmarshal(data, &req)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(req.UpdateRequest.Extensions))
	assert.Equal(t, onePasswordID, req.UpdateRequest.Extensions[0].ID)
	assert.Equal(t, onePasswordVersion, req.UpdateRequest.Extensions[0].Version)
	assert.Equal(t, pdfJSID, req.UpdateRequest.Extensions[1].ID)
	assert.Equal(t, pdfJSVersion, req.UpdateRequest.Extensions[1].Version)
}

func TestRequestUnmarshalXML(t *testing.T) {
	// Empty data returns an error
	decoder := xml.NewDecoder(strings.NewReader(""))
	var req Request
	err := req.UnmarshalXML(decoder, xml.StartElement{})
	assert.NotNil(t, err, "UnmarshalXML should return an error for empty content")

	// Malformed XML returns an error
	decoder = xml.NewDecoder(strings.NewReader("<"))
	req = Request{}
	err = req.UnmarshalXML(decoder, xml.StartElement{Name: xml.Name{Local: "request"}})
	assert.NotNil(t, err, "UnmarshalXML should return an error for malformed XML")

	// Test v3.0 request
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.0" updater="chromiumcrx" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
		<app appid="test-app-id" version="1.0.0">
			<updatecheck />
			<packages>
				<package fp="test-fingerprint" />
			</packages>
		</app>
		</request>`)

	decoder = xml.NewDecoder(strings.NewReader(string(data)))
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

	req = Request{}
	err = req.UnmarshalXML(decoder, start)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(req.UpdateRequest.Extensions))
	assert.Equal(t, "test-app-id", req.UpdateRequest.Extensions[0].ID)
	assert.Equal(t, "1.0.0", req.UpdateRequest.Extensions[0].Version)
	assert.Equal(t, "test-fingerprint", req.UpdateRequest.Extensions[0].FP)
	assert.Equal(t, "chromiumcrx", req.UpdateRequest.UpdaterType)

	// Test v3.1 request
	data = []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.1" updater="BraveComponentUpdater" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
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

	req = Request{}
	err = req.UnmarshalXML(decoder, start)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(req.UpdateRequest.Extensions))
	assert.Equal(t, "test-app-id", req.UpdateRequest.Extensions[0].ID)
	assert.Equal(t, "1.0.0", req.UpdateRequest.Extensions[0].Version)
	assert.Equal(t, "test-fingerprint", req.UpdateRequest.Extensions[0].FP)
	assert.Equal(t, "BraveComponentUpdater", req.UpdateRequest.UpdaterType)
}
